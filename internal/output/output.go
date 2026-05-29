package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// Format identifies the output format.
type Format int

const (
	FormatCSV Format = iota
	FormatJSON
	FormatTXT
)

// DetectFormat infers the output format from the file extension.
func DetectFormat(path string) Format {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return FormatJSON
	case ".txt":
		return FormatTXT
	default:
		return FormatCSV
	}
}

// Writer writes results to a file in a thread-safe manner.
//
// JSON output uses newline-delimited JSON (JSONL / JSON Lines): one JSON object
// per line, with no enclosing array. Each line is a self-contained, valid JSON
// value, so the file remains fully readable even if the process is interrupted
// mid-scan. Readers can parse it with standard JSON streaming libraries or a
// simple line-by-line loop.
type Writer struct {
	mu     sync.Mutex
	f      *os.File
	format Format
	csv    *csv.Writer
}

// New creates (or truncates) the output file and returns a ready Writer.
func New(path string, format Format) (*Writer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("opening output file %q: %w", path, err)
	}

	w := &Writer{f: f, format: format}

	if format == FormatCSV {
		w.csv = csv.NewWriter(f)
		_ = w.csv.Write([]string{
			"ip", "score", "loss_pct", "avg_ms", "min_ms", "max_ms",
			"jitter_ms", "download_kbps", "speed_tested", "colo", "tls_ok", "http_status",
		})
		w.csv.Flush()
	}

	return w, nil
}

// Write appends a result to the output file.
func (w *Writer) Write(r *result.Result) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	switch w.format {
	case FormatJSON:
		return w.writeJSON(r)
	case FormatTXT:
		return w.writeTXT(r)
	default:
		return w.writeCSV(r)
	}
}

// Close flushes and closes the underlying file.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.csv != nil {
		w.csv.Flush()
	}
	return w.f.Close()
}

func (w *Writer) writeCSV(r *result.Result) error {
	row := []string{
		r.IP.String(),
		fmt.Sprintf("%.1f", r.Score()),
		fmt.Sprintf("%.1f", r.Loss()),
		fmt.Sprintf("%.2f", result.Ms(r.Avg())),
		fmt.Sprintf("%.2f", result.Ms(r.Min())),
		fmt.Sprintf("%.2f", result.Ms(r.Max())),
		fmt.Sprintf("%.2f", result.Ms(r.Jitter())),
		fmt.Sprintf("%.1f", r.Throughput/1024),
		strconv.FormatBool(r.SpeedTested),
		r.Colo,
		strconv.FormatBool(r.TLSOk),
		strconv.Itoa(r.HTTPStatus),
	}
	_ = w.csv.Write(row)
	w.csv.Flush()
	return w.csv.Error()
}

// writeJSON appends one JSONL record (no trailing comma, no enclosing array).
func (w *Writer) writeJSON(r *result.Result) error {
	type jsonResult struct {
		IP          string  `json:"ip"`
		Score       float64 `json:"score"`
		LossPct     float64 `json:"loss_pct"`
		AvgMs       float64 `json:"avg_ms"`
		MinMs       float64 `json:"min_ms"`
		MaxMs       float64 `json:"max_ms"`
		JitterMs    float64 `json:"jitter_ms"`
		DownloadKB  float64 `json:"download_kbps,omitempty"`
		SpeedTested bool    `json:"speed_tested,omitempty"`
		Colo        string  `json:"colo,omitempty"`
		TLSOk       bool    `json:"tls_ok"`
		HTTPStatus  int     `json:"http_status,omitempty"`
	}
	obj := jsonResult{
		IP:          r.IP.String(),
		Score:       r.Score(),
		LossPct:     r.Loss(),
		AvgMs:       result.Ms(r.Avg()),
		MinMs:       result.Ms(r.Min()),
		MaxMs:       result.Ms(r.Max()),
		JitterMs:    result.Ms(r.Jitter()),
		DownloadKB:  r.Throughput / 1024,
		SpeedTested: r.SpeedTested,
		Colo:        r.Colo,
		TLSOk:       r.TLSOk,
		HTTPStatus:  r.HTTPStatus,
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = w.f.Write(b)
	return err
}

func (w *Writer) writeTXT(r *result.Result) error {
	// Plain IP-per-line format so the file can be pasted directly into
	// proxy / VPN tools (Xray, Sing-Box, etc.) without editing.
	_, err := w.f.WriteString(r.IP.String() + "\n")
	return err
}
