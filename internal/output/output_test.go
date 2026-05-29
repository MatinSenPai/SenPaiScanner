package output

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

func sampleResult() *result.Result {
	return &result.Result{
		IP:         net.ParseIP("1.1.1.1"),
		Port:       443,
		ProbeMode:  "http",
		Latencies:  []time.Duration{1500 * time.Microsecond, 2500 * time.Microsecond},
		TLSOk:      true,
		HTTPStatus: 200,
		Colo:       "FRA",
	}
}

func TestDetectFormat(t *testing.T) {
	cases := map[string]Format{
		"out.json": FormatJSON,
		"out.txt":  FormatTXT,
		"out.csv":  FormatCSV,
		"out":      FormatCSV, // default
	}
	for path, want := range cases {
		if got := DetectFormat(path); got != want {
			t.Errorf("DetectFormat(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestWriteCSVPreservesSubMs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	w, err := New(path, FormatCSV)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := w.Write(sampleResult()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header + 1 row, got %d lines", len(lines))
	}
	if !strings.HasPrefix(lines[0], "ip,score,loss_pct,avg_ms") {
		t.Errorf("unexpected header: %q", lines[0])
	}
	// Avg of 1.5ms and 2.5ms is 2.0ms; truncating to whole ms would have lost it.
	if !strings.Contains(lines[1], "2.00") {
		t.Errorf("row missing fractional avg_ms: %q", lines[1])
	}
}

func TestWriteJSONLIsValidPerLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	w, err := New(path, FormatJSON)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := w.Write(sampleResult()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var obj struct {
		IP    string  `json:"ip"`
		AvgMs float64 `json:"avg_ms"`
		Colo  string  `json:"colo"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(data))), &obj); err != nil {
		t.Fatalf("line is not valid JSON: %v", err)
	}
	if obj.IP != "1.1.1.1" || obj.Colo != "FRA" {
		t.Errorf("unexpected record: %+v", obj)
	}
	if obj.AvgMs != 2.0 {
		t.Errorf("avg_ms = %v, want 2.0", obj.AvgMs)
	}
}

func TestWriteTXTIsPlainIP(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.txt")
	w, err := New(path, FormatTXT)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := w.Write(sampleResult()); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.TrimSpace(string(data)) != "1.1.1.1" {
		t.Errorf("txt output = %q, want bare IP", string(data))
	}
}
