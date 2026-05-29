package engine

import (
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/prober"
)

func TestAdaptiveTimeoutStartsAtBase(t *testing.T) {
	a := newAdaptiveProber(prober.Config{Timeout: 3 * time.Second})
	if got := a.currentTimeout(); got != 3*time.Second {
		t.Fatalf("initial timeout = %v, want 3s (base)", got)
	}
}

func TestAdaptiveTimeoutClampsToFloor(t *testing.T) {
	a := newAdaptiveProber(prober.Config{Timeout: 3 * time.Second})
	// Many fast samples drive the EWMA toward ~20ms; k*ewma (60ms) is below the
	// floor (300ms = ceil/10), so the timeout must clamp up to the floor.
	for i := 0; i < 50; i++ {
		a.observe(20 * time.Millisecond)
	}
	if got := a.currentTimeout(); got != a.floor {
		t.Fatalf("timeout = %v, want floor %v", got, a.floor)
	}
}

func TestAdaptiveTimeoutClampsToCeil(t *testing.T) {
	a := newAdaptiveProber(prober.Config{Timeout: 3 * time.Second})
	// Slow samples (2s) want k*ewma = 6s, which must clamp down to the 3s ceil.
	for i := 0; i < 50; i++ {
		a.observe(2 * time.Second)
	}
	if got := a.currentTimeout(); got != a.ceil {
		t.Fatalf("timeout = %v, want ceil %v", got, a.ceil)
	}
}

func TestAdaptiveTimeoutTracksMidRange(t *testing.T) {
	a := newAdaptiveProber(prober.Config{Timeout: 3 * time.Second})
	for i := 0; i < 50; i++ {
		a.observe(200 * time.Millisecond)
	}
	// k=3 => ~600ms, inside [floor=300ms, ceil=3s].
	got := a.currentTimeout()
	if got < 500*time.Millisecond || got > 700*time.Millisecond {
		t.Fatalf("timeout = %v, want ~600ms", got)
	}
}

func TestEngineAdaptiveWiring(t *testing.T) {
	// With AdaptiveTimeout set and no injected Prober, New must build an
	// adaptiveProber rather than the fixed defaultProber.
	e := New(Config{AdaptiveTimeout: true, ProbeConfig: prober.Config{Timeout: time.Second}})
	if _, ok := e.prober.(*adaptiveProber); !ok {
		t.Fatalf("prober = %T, want *adaptiveProber", e.prober)
	}
}
