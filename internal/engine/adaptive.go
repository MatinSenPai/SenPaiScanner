package engine

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// adaptiveProber adjusts each probe's timeout from an exponentially-weighted
// moving average (EWMA) of recent successful latencies.
//
// A fixed timeout is a poor fit for an internet-wide scan: set it high and every
// dead IP wastes the full budget; set it low and good-but-distant edges are cut
// off and mis-counted as packet loss. By learning the latency the network is
// actually delivering and waiting a small multiple of that, the scanner spends
// its time where IPs are likely to answer and bails quickly on the rest — which
// directly raises throughput (more IPs tested per second) without inflating loss.
//
// It implements the Prober interface and is safe for concurrent use.
type adaptiveProber struct {
	cfg    prober.Config
	ewmaNs atomic.Int64 // EWMA of successful latency (ns); 0 = not yet observed
	floor  time.Duration
	ceil   time.Duration
	k      float64 // timeout = k * ewma
	alpha  float64 // EWMA smoothing factor in (0,1]
}

// newAdaptiveProber builds an adaptive prober around a base config. The base
// Timeout becomes the ceiling (we never wait longer than the user asked); the
// floor keeps a fast network from clipping legitimately slightly-slower edges.
func newAdaptiveProber(cfg prober.Config) *adaptiveProber {
	ceil := cfg.Timeout
	if ceil <= 0 {
		ceil = 3 * time.Second
	}
	floor := ceil / 10
	if floor < 200*time.Millisecond {
		floor = 200 * time.Millisecond
	}
	if floor > ceil {
		floor = ceil
	}
	return &adaptiveProber{cfg: cfg, floor: floor, ceil: ceil, k: 3, alpha: 0.2}
}

// currentTimeout returns the timeout to use for the next probe given what has
// been observed so far. Before any successful probe it returns the base timeout.
func (a *adaptiveProber) currentTimeout() time.Duration {
	e := a.ewmaNs.Load()
	if e <= 0 {
		return a.cfg.Timeout
	}
	t := time.Duration(float64(e) * a.k)
	if t < a.floor {
		t = a.floor
	}
	if t > a.ceil {
		t = a.ceil
	}
	return t
}

// observe folds a successful latency sample into the EWMA.
func (a *adaptiveProber) observe(lat time.Duration) {
	for {
		cur := a.ewmaNs.Load()
		var next int64
		if cur == 0 {
			next = int64(lat)
		} else {
			next = int64(a.alpha*float64(lat) + (1-a.alpha)*float64(cur))
		}
		if a.ewmaNs.CompareAndSwap(cur, next) {
			return
		}
	}
}

// Probe runs a single probe with an adapted timeout and updates the EWMA from
// the (robust median) latency of the result.
func (a *adaptiveProber) Probe(ctx context.Context, ip net.IP) *result.Result {
	cfg := a.cfg
	cfg.Timeout = a.currentTimeout()

	r := prober.Probe(ctx, ip, cfg)
	if lat := r.Median(); lat > 0 {
		a.observe(lat)
	}
	return r
}
