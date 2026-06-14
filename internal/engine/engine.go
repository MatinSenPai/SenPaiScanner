package engine

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"

	"github.com/ehsatn/MAGIK/internal/prober"
	"github.com/ehsatn/MAGIK/internal/result"
)

// Config controls engine behaviour.
type Config struct {
	Concurrency int
	RateLimit   float64 // probes per second, <=0 means unlimited
	ProbeConfig prober.Config
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
	return &Engine{cfg: cfg, limiter: lim}
}

// Stats returns a pointer to the live statistics.
func (e *Engine) Stats() *Stats {
	return &e.stats
}

// Run consumes IPs from src, probes each one, and forwards results to fn.
// Run consumes IPs from src, probes each one, and forwards results to fn.
// It blocks until src is exhausted or ctx is cancelled.
func (e *Engine) Run(ctx context.Context, src <-chan net.IP, fn ResultFunc) {
	sem := make(chan struct{}, e.cfg.Concurrency)
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case sem <- struct{}{}:
			select {
			case <-ctx.Done():
				<-sem
				wg.Wait()
				return
			case ip, ok := <-src:
				if !ok {
					<-sem
					wg.Wait()
					return
				}

				if e.limiter != nil {
					if err := e.limiter.Wait(ctx); err != nil {
						<-sem
						wg.Wait()
						return
					}
				}

				e.stats.InFlight.Add(1)
				wg.Add(1)

				go func(ip net.IP) {
					defer func() {
						<-sem
						e.stats.InFlight.Add(-1)
						wg.Done()
					}()

					r := prober.Probe(ctx, ip, e.cfg.ProbeConfig)
					e.stats.Tested.Add(1)
					if r.IsHealthy() {
						e.stats.Healthy.Add(1)
					} else {
						e.stats.Failed.Add(1)
					}
					fn(r)
				}(ip)
			}
		}
	}
}

// RunList probes a fixed slice of IPs (used in `magik test`).
func (e *Engine) RunList(ctx context.Context, ips []net.IP, fn ResultFunc) {
	ch := make(chan net.IP, len(ips))
	for _, ip := range ips {
		ch <- ip
	}
	close(ch)

	// Raise the timeout floor for the final validation round so slow IPs
	// still get a fair chance rather than being cut off too early.
	cfg := e.cfg
	cfg.ProbeConfig.Timeout = max(cfg.ProbeConfig.Timeout, 10*time.Second)
	e2 := New(cfg)
	e2.Run(ctx, ch, fn)
}
