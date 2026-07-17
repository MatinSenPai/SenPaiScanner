package ipsrc

import (
	"context"
	"net"
	"testing"
)

func TestNewV4Only(t *testing.T) {
	s, err := New(true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.v4Nets) == 0 {
		t.Error("expected v4 nets to be loaded")
	}
	if len(s.v6Nets) != 0 {
		t.Error("expected no v6 nets when useV6=false")
	}
}

func TestNewV6Only(t *testing.T) {
	s, err := New(false, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.v6Nets) == 0 {
		t.Error("expected v6 nets to be loaded")
	}
}

func TestNewExtraCIDR(t *testing.T) {
	s, err := New(false, false, []string{"1.1.1.0/24"})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.v4Nets) == 0 {
		t.Error("extra v4 CIDR not loaded")
	}
}

func TestNewNoRanges(t *testing.T) {
	_, err := New(false, false, nil)
	if err == nil {
		t.Error("expected error with no ranges")
	}
}

func TestRandom(t *testing.T) {
	s, err := New(true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		ip := s.Random()
		if ip == nil {
			t.Fatal("Random() returned nil")
		}
		if ip.To4() == nil {
			t.Errorf("expected IPv4, got %s", ip)
		}
	}
}

func TestRandomIsInCFRange(t *testing.T) {
	s, err := New(true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		ip := s.Random()
		inRange := false
		for _, n := range s.v4Nets {
			if n.Contains(ip) {
				inRange = true
				break
			}
		}
		if !inRange {
			t.Errorf("random IP %s not in any Cloudflare range", ip)
		}
	}
}

func TestStream(t *testing.T) {
	s, err := New(true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	ch := s.Stream(ctx, 10)
	count := 0
	for range ch {
		count++
	}
	if count != 10 {
		t.Errorf("Stream(10) emitted %d IPs, want 10", count)
	}
}

func TestStreamCancel(t *testing.T) {
	s, err := New(true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ch := s.Stream(ctx, 0)
	cancel()
	count := 0
	for range ch {
		count++
	}
	// Some IPs may have been buffered before cancel; just ensure it terminates
}

func TestFromCIDR(t *testing.T) {
	ctx := context.Background()
	ch, err := FromCIDR(ctx, "192.0.2.0/30")
	if err != nil {
		t.Fatal(err)
	}
	var ips []net.IP
	for ip := range ch {
		ips = append(ips, ip)
	}
	if len(ips) != 4 {
		t.Errorf("expected 4 IPs from /30, got %d", len(ips))
	}
}

func TestInvalidCIDR(t *testing.T) {
	_, err := New(false, false, []string{"not-a-cidr"})
	if err == nil {
		t.Error("expected error for invalid CIDR")
	}
}

func TestNewWithOptionsCIDROnly(t *testing.T) {
	s, err := NewWithOptions(true, true, []string{"192.0.2.0/30"}, Options{UseBuiltin: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.v4Nets) != 1 {
		t.Fatalf("expected exactly one v4 CIDR, got %d", len(s.v4Nets))
	}
	if got := s.v4Nets[0].String(); got != "192.0.2.0/30" {
		t.Fatalf("expected custom CIDR only, got %s", got)
	}
	if len(s.v6Nets) != 0 {
		t.Fatalf("expected no v6 CIDRs, got %d", len(s.v6Nets))
	}
}

func TestWeightedRandomIPv4Selection(t *testing.T) {
	// Create a source with two CIDR ranges of vastly different sizes:
	// A: 192.0.2.0/24 (256 IPs)
	// B: 198.51.100.0/30 (4 IPs)
	s, err := NewWithOptions(true, false, []string{"192.0.2.0/24", "198.51.100.0/30"}, Options{UseBuiltin: false})
	if err != nil {
		t.Fatal(err)
	}

	// Total IPv4 size is 256 + 4 = 260.
	// Subnet A (192.0.2.0/24) has size 256. Probability is 256/260 = 98.46%.
	// Subnet B (198.51.100.0/30) has size 4. Probability is 4/260 = 1.54%.
	// We run 1000 random selections and check that Subnet A is chosen significantly more than Subnet B.
	countA := 0
	countB := 0
	for i := 0; i < 1000; i++ {
		ip := s.Random()
		if ip.To4() != nil && ip.To4()[0] == 192 {
			countA++
		} else if ip.To4() != nil && ip.To4()[0] == 198 {
			countB++
		}
	}

	if countA < 900 {
		t.Errorf("expected Subnet A to be chosen around 98%% of the time, got A=%d, B=%d", countA, countB)
	}
}

func TestRandomV6OnlyOnV4OnlySourceDoesNotPanic(t *testing.T) {
	// Regression: when Source has v4 CIDRs but v6 enabled with no v6 CIDRs
	// (common for CIDR-only scans), v6CumSizes is empty and Random() must
	// not panic on the index access s.v6CumSizes[len(...)-1].
	s, err := NewWithOptions(true, true, []string{"192.0.2.0/30"}, Options{UseBuiltin: false})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		if ip := s.Random(); ip == nil {
			// nil is acceptable; panic is not.
		}
	}
}