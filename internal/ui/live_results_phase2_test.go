package ui

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

// newScanWriter builds a writer in a temp dir and registers it as the live
// global, exactly as launchPhase1FromOptional does for a config-carrying scan.
// Cleanup restores the global so tests cannot leak into each other.
func newScanWriter(t *testing.T, withConfig bool) (*LiveResultWriter, string) {
	t.Helper()
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	w, path, err := newLiveResultWriter(withConfig)
	if err != nil {
		t.Fatal(err)
	}
	setLiveResultWriter(w)
	t.Cleanup(clearLiveResultWriter)
	return w, path
}

func healthyRow(ip string) *result.Result {
	return &result.Result{
		IP:         net.ParseIP(ip),
		Port:       443,
		Latencies:  []time.Duration{100 * time.Millisecond, 110 * time.Millisecond},
		ProbeMode:  "http",
		TLSOk:      true,
		HTTPStatus: 200,
		Colo:       "FRA",
	}
}

// TestLiveFileKeepsPhase2AfterWriterGlobalCleared reproduces the reported bug.
//
// Phase 1 was cancelled with q on the Phase 1 page. That handler called
// clearLiveResultWriter, nilling the global, but cancelling Phase 1 still
// delivers ConfigPhase1DoneMsg, which starts Phase 2. Every AddPhase2 then hit
// its nil-receiver guard and returned silently, so the live file ended after
// Phase 1 with no Phase 2 section, even though Phase 2 completed and wrote all
// its endpoints to working_ips.txt.
func TestLiveFileKeepsPhase2AfterWriterGlobalCleared(t *testing.T) {
	w, path := newScanWriter(t, true)

	w.AddPhase1(healthyRow("104.18.1.1"))
	w.BeginPhase2()

	// The bug: navigating away from the Phase 1 page nils the global.
	clearLiveResultWriter()

	// Phase 2 runs to completion after that, with its captured writer.
	w.AddPhase2(&xraytest.ValidationResult{
		IP:         "104.18.1.1",
		Port:       443,
		Transport:  "ws",
		Success:    true,
		Throughput: 4_800_000,
	})

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)

	if !strings.Contains(text, "Phase 1") {
		t.Fatalf("live file lost its Phase 1 section:\n%s", text)
	}
	if !strings.Contains(text, "Phase 2") {
		t.Fatalf("live file lost its Phase 2 section after the writer global was cleared:\n%s", text)
	}
	if !strings.Contains(text, "104.18.1.1") {
		t.Fatalf("live file lost the Phase 2 endpoint:\n%s", text)
	}
}

// TestLiveFileHasBothPhasesAtEnd is the requirement behind the report: once
// everything has run, the single text file carries both phases.
func TestLiveFileHasBothPhasesAtEnd(t *testing.T) {
	w, path := newScanWriter(t, true)

	w.AddPhase1(healthyRow("104.18.1.1"))
	w.AddPhase1(healthyRow("172.67.1.1"))
	w.BeginPhase2()
	for _, v := range []struct{ ip, fail string }{
		{"104.18.1.1", ""},
		{"172.67.1.1", ""},
		// Failed validation: no speed, and the reason must reach the file.
		{"198.41.1.1", "speed below threshold"},
	} {
		w.AddPhase2(&xraytest.ValidationResult{
			IP:        v.ip,
			Port:      443,
			Transport: "ws",
			Success:   v.fail == "",
			Error:     v.fail,
		})
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)

	p1 := strings.Index(text, "Phase 1")
	p2 := strings.Index(text, "Phase 2")
	if p1 < 0 || p2 < 0 {
		t.Fatalf("file must contain both phases:\n%s", text)
	}
	if p1 > p2 {
		t.Fatalf("Phase 2 must be rendered after Phase 1:\n%s", text)
	}
	if !strings.Contains(text, "Phase 2 — xray validation (3 tested)") {
		t.Fatalf("Phase 2 header should count all three validated endpoints:\n%s", text)
	}
	// A failed validation must still be listed, with its reason.
	if !strings.Contains(text, "speed below threshold") {
		t.Fatalf("live file dropped the failed validation row:\n%s", text)
	}
}

// TestPhase1OnlyScanHasNoPhase2Section guards the opposite direction: a scan
// without a config URL must not grow an empty Phase 2 section.
func TestPhase1OnlyScanHasNoPhase2Section(t *testing.T) {
	w, path := newScanWriter(t, false)

	w.AddPhase1(healthyRow("104.18.1.1"))
	w.FinishPhase1Only()

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if strings.Contains(text, "Phase 2") {
		t.Fatalf("a Phase-1-only scan must not render a Phase 2 section:\n%s", text)
	}
	if !strings.Contains(text, "Phase 1 connectivity only") {
		t.Fatalf("plan line should say connectivity only:\n%s", text)
	}
}

// TestAddPhase2OnNilWriterIsSafe documents the guard that made the bug silent:
// a nil writer must not panic, but callers are responsible for not relying on it
// once the global has been cleared. runConfigPhase2 now passes the writer in for
// exactly this reason.
func TestAddPhase2OnNilWriterIsSafe(t *testing.T) {
	var w *LiveResultWriter
	w.AddPhase2(&xraytest.ValidationResult{IP: "1.2.3.4", Port: 443})
	w.BeginPhase2()
	w.AddPhase1(healthyRow("104.18.1.1"))
	w.FinishPhase1Only()
}