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

func TestLiveResultFileNameFormat(t *testing.T) {
	path, err := liveResultFilePath()
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "SenPaiScannerResult-") {
		t.Fatalf("basename = %q", base)
	}
	if !strings.HasSuffix(base, ".txt") {
		t.Fatalf("basename = %q, want .txt suffix", base)
	}
}

func TestLiveResultWriterRewritesHealthyPhase1Rows(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	w, path, err := newLiveResultWriter(false)
	if err != nil {
		t.Fatal(err)
	}
	w.AddPhase1(&result.Result{
		IP:         net.ParseIP("104.18.1.1"),
		Port:       443,
		Latencies:  []time.Duration{100 * time.Millisecond},
		ProbeMode:  "http",
		TLSOk:      true,
		HTTPStatus: 200,
		Colo:       "FRA",
	})
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "104.18.1.1:443") {
		t.Fatalf("file missing endpoint:\n%s", text)
	}
	if !strings.Contains(text, "Phase 1") {
		t.Fatalf("file missing phase header:\n%s", text)
	}
}

func TestResolveTopNCustom(t *testing.T) {
	m := NewApp("test")
	m.configTopNIdx = len(configTopNLabels) - 1
	m.configTopNCustom = "75"
	if got := m.resolveTopN(); got != 75 {
		t.Fatalf("topN = %d, want 75", got)
	}
}

func TestResolveTopNPreset(t *testing.T) {
	m := NewApp("test")
	m.configTopNIdx = 2
	if got := m.resolveTopN(); got != 50 {
		t.Fatalf("topN = %d, want 50", got)
	}
}
