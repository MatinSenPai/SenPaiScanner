package engine

import (
	"context"
	"net"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// pipelineBuffer bounds the hand-off channel between phase 1 and phase 2. It is
// generous enough that a burst of phase-1 survivors never stalls the cheap
// filter while phase 2 catches up, but small enough to stay memory-bounded.
const pipelineBuffer = 512

// DefaultKeep is the survivor predicate used by RunTwoPhase when none is given:
// a candidate advances to phase 2 if it connected at all in phase 1 (i.e. it has
// at least one successful latency sample). The expensive validation in phase 2
// then makes the final healthy/unhealthy decision.
func DefaultKeep(r *result.Result) bool {
	return r != nil && r.Avg() > 0
}

// RunTwoPhase runs a coarse→fine pipeline: every IP from src is probed cheaply
// by phase1 (e.g. a short-timeout TCP/TLS ping at high concurrency); only those
// that satisfy keep are forwarded to phase2 for expensive validation (e.g. an
// HTTP trace plus a throughput sample). fn receives every phase-2 result.
//
// This is the single biggest lever for "more IPs tested per second": the vast
// majority of random Cloudflare-space addresses are dead or filtered, and
// filtering them with a cheap probe means the costly HTTP+download work is spent
// almost entirely on real candidates. phase1.Stats() and phase2.Stats() expose
// each stage's live counters; phase2 reflects the validated results.
//
// It blocks until src is exhausted (and phase 2 drains) or ctx is cancelled.
func RunTwoPhase(ctx context.Context, src <-chan net.IP, phase1, phase2 *Engine, keep func(*result.Result) bool, fn ResultFunc) {
	RunTwoPhaseGated(ctx, src, phase1, phase2, nil, keep, fn)
}

// RunTwoPhaseFast is RunTwoPhase with the default prefix gate engaged — the
// configuration intended for wide internet sweeps, where most candidates are
// dead and skipping proven-dead blocks is the dominant throughput win. It exists
// so callers can opt into the gate without depending on the gate type.
func RunTwoPhaseFast(ctx context.Context, src <-chan net.IP, phase1, phase2 *Engine, keep func(*result.Result) bool, fn ResultFunc) {
	RunTwoPhaseGated(ctx, src, phase1, phase2, newPrefixGate(0), keep, fn)
}

// RunTwoPhaseGated is RunTwoPhase with an optional prefix gate that suppresses
// phase-1 probes to address blocks that have proven dead (see prefixGate). The
// gate both filters src before phase 1 and learns from every phase-1 outcome, so
// the probe budget drifts toward blocks that still might hold a usable edge. A
// nil gate makes this behave exactly like the ungated pipeline.
func RunTwoPhaseGated(ctx context.Context, src <-chan net.IP, phase1, phase2 *Engine, gate *prefixGate, keep func(*result.Result) bool, fn ResultFunc) {
	if keep == nil {
		keep = DefaultKeep
	}

	in := src
	if gate != nil {
		in = gateFilter(ctx, src, gate)
		keep = gate.keepAndObserve(keep)
	}

	mid := make(chan net.IP, pipelineBuffer)
	go func() {
		defer close(mid)
		phase1.Run(ctx, in, func(r *result.Result) {
			if !keep(r) {
				return
			}
			select {
			case <-ctx.Done():
			case mid <- r.IP:
			}
		})
	}()

	phase2.Run(ctx, mid, fn)
}

// gateFilter forwards only the IPs the gate still allows, dropping (and counting)
// those whose prefix has been poisoned. It honours ctx on both the receive and
// the send so it never blocks past cancellation.
func gateFilter(ctx context.Context, src <-chan net.IP, gate *prefixGate) <-chan net.IP {
	out := make(chan net.IP, pipelineBuffer)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case ip, ok := <-src:
				if !ok {
					return
				}
				if !gate.Allow(ip) {
					gate.markSkipped()
					continue
				}
				select {
				case <-ctx.Done():
					return
				case out <- ip:
				}
			}
		}
	}()
	return out
}
