package engine

import (
	"net"
	"sync"
	"testing"
)

func TestPrefixKeyAggregation(t *testing.T) {
	// Two addresses in the same /24 share a key; a third in a neighbouring /24
	// does not.
	a := prefixKey(net.ParseIP("104.16.0.1"))
	b := prefixKey(net.ParseIP("104.16.0.250"))
	c := prefixKey(net.ParseIP("104.16.1.1"))
	if a != b {
		t.Errorf("104.16.0.1 and 104.16.0.250 should share a /24 key")
	}
	if a == c {
		t.Errorf("104.16.0.0/24 and 104.16.1.0/24 must have different keys")
	}

	// IPv6 aggregates on the /48.
	v6a := prefixKey(net.ParseIP("2606:4700:0::1"))
	v6b := prefixKey(net.ParseIP("2606:4700:0:ffff::2"))
	v6c := prefixKey(net.ParseIP("2606:4700:1::1"))
	if v6a != v6b {
		t.Errorf("addresses in the same /48 should share a key")
	}
	if v6a == v6c {
		t.Errorf("different /48s must have different keys")
	}

	// IPv4 and IPv6 keys never collide.
	if a == v6a {
		t.Errorf("IPv4 and IPv6 keys must not collide")
	}
}

func TestPrefixGatePoisonsAfterThreshold(t *testing.T) {
	g := newPrefixGate(4)
	ip := net.ParseIP("104.16.0.1")

	for i := 0; i < 3; i++ {
		g.Observe(net.IPv4(104, 16, 0, byte(i)), false)
		if !g.Allow(ip) {
			t.Fatalf("prefix poisoned after %d dead, want >=4", i+1)
		}
	}
	g.Observe(net.IPv4(104, 16, 0, 99), false) // 4th dead
	if g.Allow(ip) {
		t.Fatalf("prefix should be poisoned after 4 dead probes")
	}
}

func TestPrefixGateLiveHitRedeems(t *testing.T) {
	g := newPrefixGate(4)
	base := func(b byte) net.IP { return net.IPv4(104, 16, 0, b) }

	// Three dead, then a live hit: the block is redeemed and stays open even as
	// further dead probes arrive (a confirmed-good /24 is never poisoned).
	for i := 0; i < 3; i++ {
		g.Observe(base(byte(i)), false)
	}
	g.Observe(base(50), true)
	for i := 0; i < 100; i++ {
		g.Observe(base(byte(i)), false)
	}
	if !g.Allow(base(200)) {
		t.Fatalf("a /24 with a live hit must never be poisoned")
	}
}

func TestPrefixGateLiveHitUnpoisons(t *testing.T) {
	g := newPrefixGate(4)
	base := func(b byte) net.IP { return net.IPv4(104, 16, 0, b) }

	for i := 0; i < 4; i++ {
		g.Observe(base(byte(i)), false)
	}
	if g.Allow(base(10)) {
		t.Fatalf("precondition: prefix should be poisoned")
	}
	g.Observe(base(10), true) // a late live hit
	if !g.Allow(base(11)) {
		t.Fatalf("a live hit must redeem a previously poisoned prefix")
	}
}

func TestPrefixGateIsolatesPrefixes(t *testing.T) {
	g := newPrefixGate(4)
	// Poison 104.16.0.0/24; a different /24 must remain open.
	for i := 0; i < 4; i++ {
		g.Observe(net.IPv4(104, 16, 0, byte(i)), false)
	}
	if g.Allow(net.IPv4(104, 16, 0, 5)) {
		t.Fatalf("104.16.0.0/24 should be poisoned")
	}
	if !g.Allow(net.IPv4(104, 16, 1, 5)) {
		t.Fatalf("104.16.1.0/24 must be unaffected by a neighbour's deaths")
	}
}

func TestPrefixGateZeroThresholdFallsBack(t *testing.T) {
	g := newPrefixGate(0)
	if g.k != defaultDeadThreshold {
		t.Fatalf("k = %d, want default %d", g.k, defaultDeadThreshold)
	}
}

// The gate must be safe under the concurrent access the phase-1 worker pool
// subjects it to. (Run with -race in CI; locally this still exercises the locks.)
func TestPrefixGateConcurrent(t *testing.T) {
	g := newPrefixGate(8)
	var wg sync.WaitGroup
	for w := 0; w < 16; w++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				ip := net.IPv4(104, byte(seed), byte(i>>8), byte(i))
				g.Allow(ip)
				g.Observe(ip, i%50 == 0)
			}
		}(w)
	}
	wg.Wait()
}
