package engine

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// distinctIPs returns n distinct 10.0.0.0/8 addresses so a fake prober can make
// a per-IP healthy/failed decision (loopbackIPs hands back the same IP n times,
// which is no use when the test needs to tell survivors from drops).
func distinctIPs(n int) []net.IP {
	ips := make([]net.IP, n)
	for i := range ips {
		ips[i] = net.IPv4(10, byte(i>>16), byte(i>>8), byte(i))
	}
	return ips
}

func feed(ips []net.IP) <-chan net.IP {
	ch := make(chan net.IP, len(ips))
	for _, ip := range ips {
		ch <- ip
	}
	close(ch)
	return ch
}

// RunTwoPhase must forward exactly the phase-1 survivors to phase 2 — no dead IP
// may leak through, and every live one must be validated exactly once.
func TestRunTwoPhaseForwardsOnlySurvivors(t *testing.T) {
	const n = 200
	ips := distinctIPs(n)

	// Phase 1: an IP is "alive" iff its low octet is even. DefaultKeep forwards
	// only those (healthyResult has a non-zero latency, so Avg() > 0).
	alive := func(ip net.IP) bool { return ip[len(ip)-1]%2 == 0 }
	phase1 := New(Config{Concurrency: 16, Prober: &fakeProber{result: func(ip net.IP) *result.Result {
		if alive(ip) {
			return healthyResult(ip)
		}
		return failedResult(ip)
	}}})

	// Phase 2 validates whatever it receives; record each IP it sees.
	var mu sync.Mutex
	seen := map[string]int{}
	phase2 := New(Config{Concurrency: 16, Prober: &fakeProber{result: healthyResult}})

	var callbacks int64
	RunTwoPhase(context.Background(), feed(ips), phase1, phase2, nil, func(r *result.Result) {
		mu.Lock()
		seen[r.IP.String()]++
		callbacks++
		mu.Unlock()
	})

	wantSurvivors := 0
	for _, ip := range ips {
		if alive(ip) {
			wantSurvivors++
		}
	}

	if got := phase1.Stats().Tested.Load(); got != n {
		t.Errorf("phase1 Tested = %d, want %d (every IP probed cheaply)", got, n)
	}
	if got := phase2.Stats().Tested.Load(); got != int64(wantSurvivors) {
		t.Errorf("phase2 Tested = %d, want %d (only survivors validated)", got, wantSurvivors)
	}
	if callbacks != int64(wantSurvivors) {
		t.Errorf("callback fired %d times, want %d", callbacks, wantSurvivors)
	}

	// Every IP phase 2 saw must have been a phase-1 survivor, and each exactly once.
	if len(seen) != wantSurvivors {
		t.Errorf("phase 2 saw %d distinct IPs, want %d", len(seen), wantSurvivors)
	}
	for _, ip := range ips {
		got := seen[ip.String()]
		if alive(ip) && got != 1 {
			t.Errorf("survivor %s validated %d times, want 1", ip, got)
		}
		if !alive(ip) && got != 0 {
			t.Errorf("dead IP %s leaked to phase 2 (%d times)", ip, got)
		}
	}
}

// A nil keep predicate must fall back to DefaultKeep.
func TestDefaultKeep(t *testing.T) {
	if DefaultKeep(nil) {
		t.Error("DefaultKeep(nil) = true, want false")
	}
	if got := DefaultKeep(failedResult(net.IPv4(10, 0, 0, 1))); got {
		t.Error("DefaultKeep(failed) = true, want false (no successful sample)")
	}
	if got := DefaultKeep(healthyResult(net.IPv4(10, 0, 0, 1))); !got {
		t.Error("DefaultKeep(healthy) = false, want true")
	}
}

// A gate poisoned before the run must skip every IP in that prefix: phase 1
// never probes them and the gate counts them all as saved. distinctIPs(n) for
// n<=256 stays inside one /24, so the whole feed shares the poisoned prefix.
func TestRunTwoPhaseGatedSkipsPoisonedPrefix(t *testing.T) {
	const n = 100
	ips := distinctIPs(n)

	gate := newPrefixGate(4)
	for i := 0; i < 4; i++ { // pre-poison 10.0.0.0/24
		gate.Observe(ips[i], false)
	}

	phase1 := New(Config{Concurrency: 8, Prober: &fakeProber{result: failedResult}})
	phase2 := New(Config{Concurrency: 8, Prober: &fakeProber{result: healthyResult}})

	var callbacks int64
	RunTwoPhaseGated(context.Background(), feed(ips), phase1, phase2, gate, nil,
		func(*result.Result) { callbacks++ })

	if got := phase1.Stats().Tested.Load(); got != 0 {
		t.Errorf("phase1 Tested = %d, want 0 (whole /24 was poisoned)", got)
	}
	if got := gate.Skipped(); got != n {
		t.Errorf("gate Skipped = %d, want %d", got, n)
	}
	if got := phase2.Stats().Tested.Load(); got != 0 {
		t.Errorf("phase2 Tested = %d, want 0", got)
	}
	if callbacks != 0 {
		t.Errorf("callback fired %d times, want 0", callbacks)
	}
}

// Whatever the scheduling, every fed IP is either probed by phase 1 or skipped
// by the gate — never both, never neither. This conservation invariant holds
// regardless of how many the gate happens to poison mid-run.
func TestRunTwoPhaseGatedConservation(t *testing.T) {
	const n = 100
	ips := distinctIPs(n) // all dead, all one /24 -> the gate will poison it

	gate := newPrefixGate(8)
	phase1 := New(Config{Concurrency: 8, Prober: &fakeProber{result: failedResult}})
	phase2 := New(Config{Concurrency: 8, Prober: &fakeProber{result: healthyResult}})

	RunTwoPhaseGated(context.Background(), feed(ips), phase1, phase2, gate, nil,
		func(*result.Result) {})

	probed := phase1.Stats().Tested.Load()
	skipped := gate.Skipped()
	if probed+skipped != n {
		t.Errorf("probed(%d) + skipped(%d) = %d, want %d (every IP accounted once)",
			probed, skipped, probed+skipped, n)
	}
	if got := phase2.Stats().Tested.Load(); got != 0 {
		t.Errorf("phase2 Tested = %d, want 0 (all dead)", got)
	}
}

// A custom keep predicate must override DefaultKeep: here it drops everything,
// so phase 2 must never run.
func TestRunTwoPhaseCustomKeepDropsAll(t *testing.T) {
	ips := distinctIPs(50)
	phase1 := New(Config{Concurrency: 8, Prober: &fakeProber{result: healthyResult}})
	phase2 := New(Config{Concurrency: 8, Prober: &fakeProber{result: healthyResult}})

	var callbacks int64
	RunTwoPhase(context.Background(), feed(ips), phase1, phase2,
		func(*result.Result) bool { return false },
		func(*result.Result) { callbacks++ })

	if got := phase1.Stats().Tested.Load(); got != int64(len(ips)) {
		t.Errorf("phase1 Tested = %d, want %d", got, len(ips))
	}
	if got := phase2.Stats().Tested.Load(); got != 0 {
		t.Errorf("phase2 Tested = %d, want 0 (keep dropped everything)", got)
	}
	if callbacks != 0 {
		t.Errorf("callback fired %d times, want 0", callbacks)
	}
}
