package engine

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// defaultDeadThreshold is how many dead probes a prefix may accumulate (with no
// live hit) before the gate stops spending probes on it. It is deliberately
// generous: Cloudflare announces contiguous /24s where a real edge usually
// answers within the first handful of addresses, so 16 misses is strong evidence
// the whole block is filtered or dark — while still being small relative to the
// 256 addresses a /24 holds, so a single live IP is very unlikely to be skipped.
const defaultDeadThreshold = 16

// prefixGate suppresses probes to address blocks that have proven dead. Random
// scanning of Cloudflare space wastes most of its budget on prefixes that never
// answer; once a block shows K dead probes and zero live ones, every further
// address in it is almost certainly dead too. Skipping the rest concentrates the
// probe budget on blocks that might actually contain a usable edge — a direct
// "more good IPs per second" win — and a single live hit instantly redeems the
// block so a real edge in an otherwise-quiet /24 is never lost.
//
// The aggregation unit is the /24 for IPv4 and the /48 for IPv6, matching the
// granularity at which CDNs announce and route address space. It is safe for
// concurrent use by the phase-1 worker pool.
type prefixGate struct {
	k       int
	mu      sync.Mutex
	dead    map[string]int  // prefix -> consecutive dead probes (no live hit yet)
	alive   map[string]bool // prefix -> a live IP has been seen (permanent redemption)
	skipped atomic.Int64    // probes avoided by the gate
}

// newPrefixGate builds a gate that poisons a prefix after k dead probes with no
// live hit. A non-positive k falls back to defaultDeadThreshold.
func newPrefixGate(k int) *prefixGate {
	if k <= 0 {
		k = defaultDeadThreshold
	}
	return &prefixGate{
		k:     k,
		dead:  make(map[string]int),
		alive: make(map[string]bool),
	}
}

// prefixKey maps an IP to the routing block the gate aggregates on: the /24 for
// IPv4, the /48 for IPv6. The raw bytes are used as the map key to avoid any
// per-lookup formatting cost.
func prefixKey(ip net.IP) string {
	if v4 := ip.To4(); v4 != nil {
		return string(v4[:3])
	}
	if v6 := ip.To16(); v6 != nil {
		return string(v6[:6])
	}
	return string(ip)
}

// Allow reports whether the prefix containing ip is still worth probing. It is
// true until the block is poisoned and again the moment a live IP redeems it.
func (g *prefixGate) Allow(ip net.IP) bool {
	key := prefixKey(ip)
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.alive[key] {
		return true
	}
	return g.dead[key] < g.k
}

// Observe folds a phase-1 outcome for ip into the gate. A live result redeems the
// block permanently (and forgets its dead tally); a dead result advances the
// tally only while the block has never produced a live hit.
func (g *prefixGate) Observe(ip net.IP, alive bool) {
	key := prefixKey(ip)
	g.mu.Lock()
	defer g.mu.Unlock()
	if alive {
		g.alive[key] = true
		delete(g.dead, key)
		return
	}
	if g.alive[key] {
		return
	}
	g.dead[key]++
}

// markSkipped records that the gate avoided one probe.
func (g *prefixGate) markSkipped() { g.skipped.Add(1) }

// Skipped returns how many probes the gate has avoided so far.
func (g *prefixGate) Skipped() int64 { return g.skipped.Load() }

// keepAndObserve adapts a survivor predicate into one that also feeds the gate:
// it reports the same alive/dead verdict keep would, after recording it. Used by
// the gated pipeline so phase-1 outcomes both drive forwarding and poison prefixes.
func (g *prefixGate) keepAndObserve(keep func(*result.Result) bool) func(*result.Result) bool {
	return func(r *result.Result) bool {
		alive := keep(r)
		if r != nil {
			g.Observe(r.IP, alive)
		}
		return alive
	}
}
