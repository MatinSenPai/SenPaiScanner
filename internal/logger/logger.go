package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level represents the severity of a log message.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO "
	case LevelWarn:
		return "WARN "
	case LevelError:
		return "ERROR"
	default:
		return "?????"
	}
}

// Phase identifies the scan phase a log entry belongs to.
type Phase string

const (
	PhaseNone    Phase = ""
	Phase1       Phase = "Phase1"
	Phase2       Phase = "Phase2"
	PhaseQuick   Phase = "Quick"
	PhaseColos   Phase = "Colos"
	PhaseStartup Phase = "Startup"
)

const (
	maxLogFiles  = 5
	maxLogSizeMB = 10
)

// Logger writes structured, formatted log entries to a file.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	minLevel Level
	dir      string

	failureCount int
	failureLines []string

	customWriter io.Writer
}

// Config holds logger initialization options.
type Config struct {
	Dir      string // directory for log files; empty = auto-detect
	MinLevel Level  // minimum severity to record
}

// New creates and initializes a Logger. It returns a no-op logger on failure.
func New(cfg Config) *Logger {
	dir := cfg.Dir
	if dir == "" {
		dir = defaultLogDir()
	}

	l := &Logger{
		dir:      dir,
		minLevel: cfg.MinLevel,
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return l
	}

	rotateOldFiles(dir)

	path := filepath.Join(dir, fmt.Sprintf("scan-%s.log", time.Now().Format("20060102-150405")))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return l
	}
	l.file = f
	return l
}

// NewFromWriter creates a Logger that writes to the given writer (for testing).
func NewFromWriter(w io.Writer, minLevel Level) *Logger {
	return &Logger{
		minLevel:    minLevel,
		customWriter: w,
	}
}

// Close flushes and closes the underlying file.
func (l *Logger) Close() {
	if l == nil || l.file == nil {
		return
	}
	l.file.Sync()
	l.file.Close()
}

// LogDir returns the directory where log files are stored.
func (l *Logger) LogDir() string {
	if l == nil {
		return ""
	}
	return l.dir
}

// LogPath returns the full path of the active log file.
func (l *Logger) LogPath() string {
	if l == nil || l.file == nil {
		return ""
	}
	return l.file.Name()
}

// HasFailures returns true if any WARN or ERROR entries were logged.
func (l *Logger) HasFailures() bool {
	if l == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.failureCount > 0
}

// FailureCount returns the number of WARN/ERROR entries logged.
func (l *Logger) FailureCount() int {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.failureCount
}

// FailureSummary returns a short string summarizing failures (e.g. "3 failures").
func (l *Logger) FailureSummary() string {
	if l == nil {
		return ""
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.failureCount == 0 {
		return ""
	}
	return fmt.Sprintf("%d failures", l.failureCount)
}

// FailureLines returns up to maxLines of recent failure log entries.
func (l *Logger) FailureLines(maxLines int) []string {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	n := len(l.failureLines)
	if n == 0 {
		return nil
	}
	start := n - maxLines
	if start < 0 {
		start = 0
	}
	out := make([]string, n-start)
	copy(out, l.failureLines[start:])
	return out
}

func (l *Logger) log(level Level, phase Phase, msg string) {
	if l == nil || level < l.minLevel {
		return
	}

	ts := time.Now().Format("2006-01-02 15:04:05.000")
	phaseTag := ""
	if phase != PhaseNone {
		phaseTag = fmt.Sprintf(" [%s]", phase)
	}
	line := fmt.Sprintf("[%s] [%s]%s %s\n", ts, level, phaseTag, msg)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.WriteString(line)
	} else if l.customWriter != nil {
		l.customWriter.Write([]byte(line))
	}

	if level >= LevelWarn {
		l.failureCount++
		l.failureLines = append(l.failureLines, strings.TrimSpace(line))
	}
}

// Debug logs a debug message.
func (l *Logger) Debug(phase Phase, format string, args ...any) {
	l.log(LevelDebug, phase, fmt.Sprintf(format, args...))
}

// Info logs an informational message.
func (l *Logger) Info(phase Phase, format string, args ...any) {
	l.log(LevelInfo, phase, fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func (l *Logger) Warn(phase Phase, format string, args ...any) {
	l.log(LevelWarn, phase, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func (l *Logger) Error(phase Phase, format string, args ...any) {
	l.log(LevelError, phase, fmt.Sprintf(format, args...))
}

// ScanStart logs the beginning of a scan with its parameters.
func (l *Logger) ScanStart(phase Phase, params map[string]any) {
	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	sort.Strings(parts)
	l.Info(phase, "Scan started — %s", strings.Join(parts, " "))
}

// ScanComplete logs the end of a scan with summary statistics.
func (l *Logger) ScanComplete(phase Phase, tested, healthy int, avgMs float64) {
	l.Info(phase, "Scan complete — tested=%d healthy=%d avg=%.1fms", tested, healthy, avgMs)
}

// ProbeFailure logs a single probe failure with a reason.
func (l *Logger) ProbeFailure(phase Phase, endpoint, reason string) {
	l.Warn(phase, "Failed: %s — %s", endpoint, reason)
}

// ProbeSuccess logs a successful probe (at debug level to avoid noise).
func (l *Logger) ProbeSuccess(phase Phase, endpoint string, latencyMs float64) {
	l.Debug(phase, "OK: %s (%.1fms)", endpoint, latencyMs)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func defaultLogDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "senpaiscanner", "logs")
	}
	return "logs"
}

// rotateOldFiles removes oldest log files when count exceeds maxLogFiles,
// and removes oversized files exceeding maxLogSizeMB.
func rotateOldFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var logFiles []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			logFiles = append(logFiles, e)
		}
	}

	// Remove oversized files
	for _, f := range logFiles {
		info, err := f.Info()
		if err != nil {
			continue
		}
		if info.Size() > int64(maxLogSizeMB)*1024*1024 {
			os.Remove(filepath.Join(dir, f.Name()))
		}
	}

	// Re-read after removing oversized
	entries, err = os.ReadDir(dir)
	if err != nil {
		return
	}
	logFiles = nil
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			logFiles = append(logFiles, e)
		}
	}

	if len(logFiles) <= maxLogFiles {
		return
	}

	// Sort by name (timestamp-based names sort chronologically)
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Name() < logFiles[j].Name()
	})

	toRemove := logFiles[:len(logFiles)-maxLogFiles]
	for _, f := range toRemove {
		os.Remove(filepath.Join(dir, f.Name()))
	}
}

// GetFailureLog returns the full content of the most recent log file, or empty string.
func GetFailureLog(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var latest string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			if latest == "" || e.Name() > latest {
				latest = e.Name()
			}
		}
	}
	if latest == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(dir, latest))
	if err != nil {
		return ""
	}
	return string(data)
}

// GetRecentLogLines returns the last N lines from the most recent log file.
func GetRecentLogLines(dir string, n int) []string {
	content := GetFailureLog(dir)
	if content == "" {
		return nil
	}
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) <= n {
		return lines
	}
	return lines[len(lines)-n:]
}
