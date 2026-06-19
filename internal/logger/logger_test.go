package logger

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewCreatesLogFile(t *testing.T) {
	dir := t.TempDir()
	l := New(Config{Dir: dir, MinLevel: LevelDebug})
	defer l.Close()

	if l.LogPath() == "" {
		t.Fatal("expected non-empty log path")
	}
	if l.LogDir() != dir {
		t.Errorf("LogDir() = %q, want %q", l.LogDir(), dir)
	}

	if _, err := os.Stat(l.LogPath()); os.IsNotExist(err) {
		t.Fatal("log file was not created")
	}
}

func TestLogFormat(t *testing.T) {
	dir := t.TempDir()
	l := New(Config{Dir: dir, MinLevel: LevelDebug})
	defer l.Close()

	l.Info(Phase1, "test message %d", 42)
	l.Close()

	content, err := os.ReadFile(l.LogPath())
	if err != nil {
		t.Fatal(err)
	}
	line := strings.TrimSpace(string(content))

	// Format: [YYYY-MM-DD HH:MM:SS.mmm] [LEVEL ] [Phase] message
	if !strings.HasPrefix(line, "[") {
		t.Errorf("line missing opening bracket: %q", line)
	}
	if !strings.Contains(line, "[INFO ]") {
		t.Errorf("line missing level tag: %q", line)
	}
	if !strings.Contains(line, "[Phase1]") {
		t.Errorf("line missing phase tag: %q", line)
	}
	if !strings.HasSuffix(line, "test message 42") {
		t.Errorf("line missing message: %q", line)
	}
}

func TestLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelWarn)

	l.Debug(Phase1, "debug msg")
	l.Info(Phase1, "info msg")
	l.Warn(Phase1, "warn msg")
	l.Error(Phase1, "error msg")

	out := buf.String()
	if strings.Contains(out, "debug msg") {
		t.Error("debug message should be filtered at WARN level")
	}
	if strings.Contains(out, "info msg") {
		t.Error("info message should be filtered at WARN level")
	}
	if !strings.Contains(out, "warn msg") {
		t.Error("warn message should be present")
	}
	if !strings.Contains(out, "error msg") {
		t.Error("error message should be present")
	}
}

func TestPhaseNoneOmitsPhaseTag(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelDebug)

	l.Info(PhaseNone, "no phase")
	l.Info(Phase2, "with phase")

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	// Line 1 should NOT have a phase tag
	if strings.Contains(lines[0], "[Phase") {
		t.Errorf("PhaseNone should not have phase tag: %q", lines[0])
	}
	// Line 2 should have a phase tag
	if !strings.Contains(lines[1], "[Phase2]") {
		t.Errorf("Phase2 should have phase tag: %q", lines[1])
	}
}

func TestFailureTracking(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelDebug)

	if l.HasFailures() {
		t.Error("expected no failures initially")
	}

	l.Warn(Phase1, "something bad")
	l.Error(Phase2, "something worse")
	l.Info(Phase1, "all good")

	if !l.HasFailures() {
		t.Error("expected failures after Warn+Error")
	}
	if l.FailureCount() != 2 {
		t.Errorf("FailureCount() = %d, want 2", l.FailureCount())
	}
	if l.FailureSummary() != "2 failures" {
		t.Errorf("FailureSummary() = %q, want %q", l.FailureSummary(), "2 failures")
	}

	lines := l.FailureLines(10)
	if len(lines) != 2 {
		t.Errorf("FailureLines(10) returned %d lines, want 2", len(lines))
	}
	if !strings.Contains(lines[0], "something bad") {
		t.Errorf("first failure line missing message: %q", lines[0])
	}
	if !strings.Contains(lines[1], "something worse") {
		t.Errorf("second failure line missing message: %q", lines[1])
	}
}

func TestFailureLinesCapping(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelDebug)

	for i := 0; i < 10; i++ {
		l.Warn(Phase1, "fail %d", i)
	}

	lines := l.FailureLines(3)
	if len(lines) != 3 {
		t.Errorf("FailureLines(3) returned %d lines, want 3", len(lines))
	}
	if !strings.Contains(lines[0], "fail 7") {
		t.Errorf("first capped line should be fail 7, got: %q", lines[0])
	}
	if !strings.Contains(lines[2], "fail 9") {
		t.Errorf("last capped line should be fail 9, got: %q", lines[2])
	}
}

func TestNilLoggerIsNoop(t *testing.T) {
	var l *Logger
	// All methods should be safe on nil
	l.Debug(Phase1, "debug")
	l.Info(Phase1, "info")
	l.Warn(Phase1, "warn")
	l.Error(Phase1, "error")
	l.ScanStart(Phase1, map[string]any{"count": 100})
	l.ScanComplete(Phase1, 100, 10, 50.0)
	l.ProbeFailure(Phase1, "1.1.1.1:443", "timeout")
	l.ProbeSuccess(Phase1, "1.1.1.1:443", 50.0)
	if l.HasFailures() {
		t.Error("nil logger should not have failures")
	}
	if l.FailureCount() != 0 {
		t.Error("nil logger failure count should be 0")
	}
	if l.FailureSummary() != "" {
		t.Error("nil logger failure summary should be empty")
	}
	if l.LogPath() != "" {
		t.Error("nil logger path should be empty")
	}
	if l.LogDir() != "" {
		t.Error("nil logger dir should be empty")
	}
}

func TestScanStartFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelDebug)

	l.ScanStart(Phase1, map[string]any{"count": 5000, "workers": 50})
	l.Close()

	out := buf.String()
	if !strings.Contains(out, "Scan started") {
		t.Errorf("missing scan started: %q", out)
	}
	if !strings.Contains(out, "count=5000") {
		t.Errorf("missing count param: %q", out)
	}
	if !strings.Contains(out, "workers=50") {
		t.Errorf("missing workers param: %q", out)
	}
}

func TestScanCompleteFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelDebug)

	l.ScanComplete(Phase2, 100, 15, 123.4)
	l.Close()

	out := buf.String()
	if !strings.Contains(out, "tested=100") {
		t.Errorf("missing tested: %q", out)
	}
	if !strings.Contains(out, "healthy=15") {
		t.Errorf("missing healthy: %q", out)
	}
	if !strings.Contains(out, "avg=123.4ms") {
		t.Errorf("missing avg: %q", out)
	}
}

func TestConcurrentSafety(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelDebug)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			l.Info(Phase1, "concurrent message %d", n)
			if n%3 == 0 {
				l.Warn(Phase1, "concurrent warn %d", n)
			}
		}(i)
	}
	wg.Wait()

	if !l.HasFailures() {
		t.Error("expected failures from concurrent warns")
	}
	// 34 warns (0,3,6,...,99 = 34 values)
	if l.FailureCount() != 34 {
		t.Errorf("FailureCount() = %d, want 34", l.FailureCount())
	}
}

func TestRotationDeletesOldest(t *testing.T) {
	dir := t.TempDir()

	// Create 7 log files (exceeds maxLogFiles=5)
	for i := 0; i < 7; i++ {
		name := filepath.Join(dir, "scan-20260101-00000"+string(rune('0'+i))+".log")
		os.WriteFile(name, []byte("old log"), 0644)
		// Ensure different mtimes
		time.Sleep(10 * time.Millisecond)
	}

	rotateOldFiles(dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var logFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".log") {
			logFiles = append(logFiles, e.Name())
		}
	}
	if len(logFiles) != maxLogFiles {
		t.Errorf("after rotation: %d files, want %d", len(logFiles), maxLogFiles)
	}
}

func TestGetRecentLogLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scan-test.log")
	lines := []string{
		"[2026-01-01 00:00:00.000] [INFO ] [Phase1] line 1",
		"[2026-01-01 00:00:01.000] [WARN ] [Phase1] line 2",
		"[2026-01-01 00:00:02.000] [ERROR] [Phase2] line 3",
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0644)

	recent := GetRecentLogLines(dir, 2)
	if len(recent) != 2 {
		t.Errorf("GetRecentLogLines(2) returned %d, want 2", len(recent))
	}
	if !strings.Contains(recent[0], "line 2") {
		t.Errorf("first recent line should be line 2, got: %q", recent[0])
	}
	if !strings.Contains(recent[1], "line 3") {
		t.Errorf("second recent line should be line 3, got: %q", recent[1])
	}
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO "},
		{LevelWarn, "WARN "},
		{LevelError, "ERROR"},
		{Level(99), "?????"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestProbeFailureFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelDebug)

	l.ProbeFailure(Phase1, "104.18.5.1:443", "timeout after 5s")

	out := buf.String()
	if !strings.Contains(out, "Failed: 104.18.5.1:443") {
		t.Errorf("missing endpoint: %q", out)
	}
	if !strings.Contains(out, "timeout after 5s") {
		t.Errorf("missing reason: %q", out)
	}
}

func TestProbeSuccessFormat(t *testing.T) {
	var buf bytes.Buffer
	l := NewFromWriter(&buf, LevelDebug)

	l.ProbeSuccess(Phase1, "1.1.1.1:443", 42.5)

	out := buf.String()
	if !strings.Contains(out, "OK: 1.1.1.1:443") {
		t.Errorf("missing endpoint: %q", out)
	}
	if !strings.Contains(out, "42.5ms") {
		t.Errorf("missing latency: %q", out)
	}
}
