package engine

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// Prober probes a single IP and returns a Result. Implementations must be safe
// for concurrent use by multiple worker goroutines. Defining the interface here
// (rather than depending on the concrete prober) lets the engine be driven by a
// fake in tests, so worker-pool behaviour can be verified without any network.
type Prober interface {
	Probe(ctx context.Context, ip net.IP) *result.Result
}

// Config controls engine behaviour.
type Config struct {
	Concurrency int
	RateLimit   float64 // probes per second, <=0 means unlimited
	ProbeConfig prober.Config

	// AdaptiveTimeout, when true, replaces the fixed per-probe timeout with one
	// derived from an EWMA of recent successful latencies (bounded by
	// ProbeConfig.Timeout). Ignored when Prober is set.
	AdaptiveTimeout bool

	// Prober, when set, overrides the default prober built from ProbeConfig.
	// Production code leaves this nil; tests inject a deterministic fake.
	Prober Prober
}

// defaultProber adapts the package-level prober.Probe to the Prober interface.
type defaultProber struct{ cfg prober.Config }

func (d defaultProber) Probe(ctx context.Context, ip net.IP) *result.Result {
	return prober.Probe(ctx, ip, d.cfg)
}

// Stats exposes real-time counters.
type Stats struct {
	Tested   atomic.Int64
	Healthy  atomic.Int64
	Failed   atomic.Int64
	InFlight atomic.Int64
}

// ResultFunc is called for every completed probe result. It is invoked from
// worker goroutines, so implementations must be goroutine-safe.
type ResultFunc func(*result.Result)

// Engine orchestrates a pool of prober goroutines.
type Engine struct {
	cfg     Config
	stats   Stats
	limiter *rate.Limiter
	prober  Prober
}

// New creates a new Engine.
func New(cfg Config) *Engine {
	var lim *rate.Limiter
	if cfg.RateLimit > 0 {
		lim = rate.NewLimiter(rate.Limit(cfg.RateLimit), int(cfg.RateLimit)+1)
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 100
	}
	p := cfg.Prober
	if p == nil {
		if cfg.AdaptiveTimeout {
			p = newAdaptiveProber(cfg.ProbeConfig)
		} else {
			p = defaultProber{cfg.ProbeConfig}
		}
	}
	return &Engine{cfg: cfg, limiter: lim, prober: p}
}

// Stats returns a pointer to the live statistics.
func (e *Engine) Stats() *Stats {
	return &e.stats
}

// Run consumes IPs from src, probes each one, and forwards results to fn.
// It blocks until src is exhausted or ctx is cancelled.
//
// Concurrency is bounded by a fixed pool of long-lived worker goroutines rather
// than a goroutine-per-IP gated by a semaphore. For a 100k-IP scan that turns
// 100k goroutine spawns into a constant Concurrency, eliminating scheduler and
// allocation churn while giving the same parallelism.
func (e *Engine) Run(ctx context.Context, src <-chan net.IP, fn ResultFunc) {
	var wg sync.WaitGroup
	wg.Add(e.cfg.Concurrency)
	for i := 0; i < e.cfg.Concurrency; i++ {
		go e.worker(ctx, src, fn, &wg)
	}
	wg.Wait()
}

// worker pulls IPs off src until the channel closes or ctx is cancelled.
func (e *Engine) worker(ctx context.Context, src <-chan net.IP, fn ResultFunc, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ip, ok := <-src:
			if !ok {
				return
			}
			e.probeOne(ctx, ip, fn)
		}
	}
}

// probeOne runs a single probe, updates counters, and invokes fn. A panic in a
// probe or in fn is contained here so one bad IP can never tear down the pool,
// and InFlight is balanced even on that path.
func (e *Engine) probeOne(ctx context.Context, ip net.IP, fn ResultFunc) {
	if e.limiter != nil {
		if err := e.limiter.Wait(ctx); err != nil {
			return
		}
	}

	e.stats.InFlight.Add(1)
	defer e.stats.InFlight.Add(-1)

	r := e.safeProbe(ctx, ip)
	e.stats.Tested.Add(1)
	if r.IsHealthy() {
		e.stats.Healthy.Add(1)
	} else {
		e.stats.Failed.Add(1)
	}
	safeCallback(fn, r)
}

// safeProbe wraps prober.Probe so a panic yields an (unhealthy) result instead
// of crashing the worker.
func (e *Engine) safeProbe(ctx context.Context, ip net.IP) (r *result.Result) {
	defer func() {
		if recover() != nil {
			r = &result.Result{
				IP:        ip,
				Port:      e.cfg.ProbeConfig.Port,
				ProbeMode: e.cfg.ProbeConfig.Mode.String(),
			}
		}
	}()
	return e.prober.Probe(ctx, ip)
}

// safeCallback isolates a panic in the user-supplied result handler.
func safeCallback(fn ResultFunc, r *result.Result) {
	defer func() { _ = recover() }()
	fn(r)
}

// RunList probes a fixed slice of IPs (used by the "Test IPs" flow).
func (e *Engine) RunList(ctx context.Context, ips []net.IP, fn ResultFunc) {
	ch := make(chan net.IP, len(ips))
	for _, ip := range ips {
		ch <- ip
	}
	close(ch)

	// Raise the timeout floor for the final validation round so slow IPs still
	// get a fair chance rather than being cut off too early. Run on this engine
	// (rather than a fresh one) so callers observe progress through e.Stats().
	e.cfg.ProbeConfig.Timeout = max(e.cfg.ProbeConfig.Timeout, 10*time.Second)
	e.Run(ctx, ch, fn)
}
