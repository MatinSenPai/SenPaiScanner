package ui

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

func healthyTCP(ip string) *result.Result {
	return &result.Result{
		IP:        net.ParseIP(ip),
		Port:      443,
		ProbeMode: "tcp",
		Latencies: []time.Duration{time.Millisecond},
		Timestamp: time.Now(),
	}
}

func TestSaveResultsToFileWritesHealthy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.csv")
	m := AppModel{scanResults: []*result.Result{healthyTCP("1.1.1.1")}}

	status := m.saveResultsToFile(path)
	if !strings.HasPrefix(status, "✓") {
		t.Fatalf("status = %q, want a success message", status)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "1.1.1.1") {
		t.Errorf("saved file missing IP, got:\n%s", data)
	}
}

func TestSaveResultsToFileEmptyName(t *testing.T) {
	m := AppModel{scanResults: []*result.Result{healthyTCP("1.1.1.1")}}
	if status := m.saveResultsToFile("   "); !strings.Contains(status, "empty filename") {
		t.Errorf("status = %q, want an empty-filename message", status)
	}
}

func TestSaveResultsToFileNoHealthy(t *testing.T) {
	// An unhealthy result (all tries failed) must not be written.
	dead := &result.Result{
		IP:        net.ParseIP("2.2.2.2"),
		Port:      443,
		ProbeMode: "tcp",
		Latencies: []time.Duration{0, 0},
		Timestamp: time.Now(),
	}
	m := AppModel{scanResults: []*result.Result{dead}}

	path := filepath.Join(t.TempDir(), "out.csv")
	status := m.saveResultsToFile(path)
	if !strings.Contains(status, "no healthy IPs") {
		t.Errorf("status = %q, want a no-healthy-IPs message", status)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected no file to be created, stat err = %v", err)
	}
}
