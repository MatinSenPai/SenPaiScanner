package main

import (
	"strings"
	"testing"
)

func TestResolvePortsExplicitWins(t *testing.T) {
	got := resolvePorts([]int{0, 443, 8443}, "vless://x", 2053)
	want := []int{443, 8443}
	if len(got) != len(want) {
		t.Fatalf("resolvePorts = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("resolvePorts = %v, want %v", got, want)
		}
	}
}

func TestResolvePortsConfigFallsBackToProbePort(t *testing.T) {
	// Only the Config(0) chip selected with a config URL: use the URL port.
	got := resolvePorts([]int{0}, "vless://x", 2053)
	if len(got) != 1 || got[0] != 2053 {
		t.Fatalf("resolvePorts = %v, want [2053]", got)
	}
}

func TestResolvePortsEmptyDefaults443(t *testing.T) {
	got := resolvePorts(nil, "", 443)
	if len(got) != 1 || got[0] != 443 {
		t.Fatalf("resolvePorts = %v, want [443]", got)
	}
}

func TestScanParamsPersistsNeighborChoice(t *testing.T) {
	saved := (ScanParams{NeighborScan: true}).toSavedConfig()
	if !saved.NeighborScan {
		t.Fatal("NeighborScan was not persisted")
	}
}

func TestStartSpeedTestRequiresGreenResults(t *testing.T) {
	app := NewApp()
	err := app.StartSpeedTest(ScanParams{})
	if err == nil || !strings.Contains(err.Error(), "no healthy results") {
		t.Fatalf("StartSpeedTest error = %v, want no healthy results", err)
	}
	app.mu.Lock()
	running := app.scanning
	app.mu.Unlock()
	if running {
		t.Fatal("app remained in scanning state after rejected speed test")
	}
}
