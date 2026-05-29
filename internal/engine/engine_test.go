package engine

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// startTCPListener spins up a localhost TCP listener that accepts and
// immediately closes every connection until cleanup is called. It returns the
// bound port so probes can target a guaranteed-open socket.
func startTCPListener(t *testing.T) (port int, cleanup func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			go conn.Close() // keep Accept hot so the backlog never starves under burst
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, func() { ln.Close() }
}

func loopbackIPs(n int) []net.IP {
	ips := make([]net.IP, n)
	for i := range ips {
		ips[i] = net.ParseIP("127.0.0.1")
	}
	return ips
}

// fakeProber is a deterministic, network-free Prober for testing the worker
// pool in isolation from real network timing.
type fakeProber struct {
	result func(ip net.IP) *result.Result
	delay  time.Duration
}

func (f *fakeProber) Probe(ctx context.Context, ip net.IP) *result.Result {
	if f.delay > 0 {
		select {
		case <-ctx.Done():
		case <-time.After(f.delay):
		}
	}
	return f.result(ip)
}

func healthyResult(ip net.IP) *result.Result {
	return &result.Result{IP: ip, Port: 443, ProbeMode: "tcp", Latencies: []time.Duration{5 * time.Millisecond}}
}

func failedResult(ip net.IP) *result.Result {
	return &result.Result{IP: ip, Port: 443, ProbeMode: "tcp", Latencies: []time.Duration{0}}
}

// With an injected prober that always returns a healthy result, the engine must
// count every probe as healthy — deterministically, with no dependence on host
// clock resolution.
func TestEngineInjectedProberHealthy(t *testing.T) {
	eng := New(Config{Concurrency: 8, Prober: &fakeProber{result: healthyResult}})

	const n = 100
	eng.RunList(context.Background(), loopbackIPs(n), func(r *result.Result) {})

	s := eng.Stats()
	if s.Tested.Load() != n || s.Healthy.Load() != n || s.Failed.Load() != 0 || s.InFlight.Load() != 0 {
		t.Errorf("Tested=%d Healthy=%d Failed=%d InFlight=%d, want %d/%d/0/0",
			s.Tested.Load(), s.Healthy.Load(), s.Failed.Load(), s.InFlight.Load(), n, n)
	}
}

// With an injected prober that always fails, every probe must be counted as
// failed and none as healthy.
func TestEngineInjectedProberFailed(t *testing.T) {
	eng := New(Config{Concurrency: 8, Prober: &fakeProber{result: failedResult}})

	const n = 100
	eng.RunList(context.Background(), loopbackIPs(n), func(r *result.Result) {})

	s := eng.Stats()
	if s.Tested.Load() != n || s.Failed.Load() != n || s.Healthy.Load() != 0 {
		t.Errorf("Tested=%d Healthy=%d Failed=%d, want %d/0/%d",
			s.Tested.Load(), s.Healthy.Load(), s.Failed.Load(), n, n)
	}
}

// The engine's contract is accounting, not classification: every IP must be
// probed exactly once, classified exactly once (healthy or failed), the
// callback must fire once per result, and InFlight must return to zero. (How
// many loopback connects measure as healthy depends on the host clock's
// latency resolution, so that split is deliberately not asserted here.)
func TestEngineAccountsEveryProbe(t *testing.T) {
	port, cleanup := startTCPListener(t)
	defer cleanup()

	eng := New(Config{
		Concurrency: 8,
		ProbeConfig: prober.Config{
			Port:    port,
			Mode:    prober.ModeTCP,
			Tries:   1,
			Timeout: 2 * time.Second,
		},
	})

	const n = 50
	var got int64
	eng.RunList(context.Background(), loopbackIPs(n), func(r *result.Result) {
		atomic.AddInt64(&got, 1)
	})

	s := eng.Stats()
	if s.Tested.Load() != n {
		t.Errorf("Tested = %d, want %d", s.Tested.Load(), n)
	}
	if h, f := s.Healthy.Load(), s.Failed.Load(); h+f != n {
		t.Errorf("Healthy+Failed = %d+%d = %d, want %d (every probe classified exactly once)", h, f, h+f, n)
	}
	if s.InFlight.Load() != 0 {
		t.Errorf("InFlight = %d, want 0 (must balance after the run)", s.InFlight.Load())
	}
	if got != n {
		t.Errorf("callback fired %d times, want %d", got, n)
	}
}

// A TLS handshake against a listener that closes immediately always fails, so
// every probe is recorded as failed — never as a false-positive healthy IP.
func TestEngineRunListCountsFailed(t *testing.T) {
	port, cleanup := startTCPListener(t)
	defer cleanup()

	eng := New(Config{
		Concurrency: 4,
		ProbeConfig: prober.Config{
			Port:    port,
			Mode:    prober.ModeTLS,
			Tries:   1,
			Timeout: 2 * time.Second,
		},
	})

	const n = 10
	eng.RunList(context.Background(), loopbackIPs(n), func(r *result.Result) {})

	s := eng.Stats()
	if s.Tested.Load() != n {
		t.Errorf("Tested = %d, want %d", s.Tested.Load(), n)
	}
	if s.Failed.Load() != n {
		t.Errorf("Failed = %d, want %d", s.Failed.Load(), n)
	}
	if s.Healthy.Load() != 0 {
		t.Errorf("Healthy = %d, want 0", s.Healthy.Load())
	}
	if s.InFlight.Load() != 0 {
		t.Errorf("InFlight = %d, want 0", s.InFlight.Load())
	}
}

// With an already-cancelled context and a source that never delivers, every
// worker must observe ctx.Done() and exit so Run returns promptly.
func TestEngineRunRespectsCancelledContext(t *testing.T) {
	eng := New(Config{
		Concurrency: 4,
		ProbeConfig: prober.Config{Mode: prober.ModeTCP, Tries: 1, Timeout: time.Second},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	src := make(chan net.IP) // never fed, never closed
	done := make(chan struct{})
	go func() {
		eng.Run(ctx, src, func(r *result.Result) {})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after the context was cancelled")
	}
	if got := eng.Stats().Tested.Load(); got != 0 {
		t.Errorf("Tested = %d, want 0 — nothing should be probed under a cancelled context", got)
	}
}

// A callback that always panics must be contained so the worker pool survives
// and still processes every IP.
func TestEngineCallbackPanicIsContained(t *testing.T) {
	port, cleanup := startTCPListener(t)
	defer cleanup()

	eng := New(Config{
		Concurrency: 4,
		ProbeConfig: prober.Config{Port: port, Mode: prober.ModeTCP, Tries: 1, Timeout: 2 * time.Second},
	})

	const n = 10
	eng.RunList(context.Background(), loopbackIPs(n), func(r *result.Result) {
		panic("boom")
	})

	if got := eng.Stats().Tested.Load(); got != n {
		t.Errorf("Tested = %d, want %d — the pool must survive callback panics", got, n)
	}
}

// BenchmarkEnginePool measures pure worker-pool dispatch overhead using a
// network-free prober, so the number reflects scheduling/accounting cost rather
// than probe latency. Run: go test ./internal/engine -bench=Pool -run=^$
func BenchmarkEnginePool(b *testing.B) {
	eng := New(Config{Concurrency: 64, Prober: &fakeProber{result: healthyResult}})

	src := make(chan net.IP, 256)
	ip := net.ParseIP("127.0.0.1")
	go func() {
		for i := 0; i < b.N; i++ {
			src <- ip
		}
		close(src)
	}()

	b.ResetTimer()
	eng.Run(context.Background(), src, func(r *result.Result) {})
}
