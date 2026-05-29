package ipsrc

import (
	"context"
	"net"
	"os"
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

func TestParseLinesDedupe(t *testing.T) {
	raw := "1.2.3.0/24\n1.2.3.0/24\n4.5.6.0/24\n# comment\n\n"
	nets, err := parseLines(raw)
	if err != nil {
		t.Fatal(err)
	}
	// A CIDR listed twice would be double-weighted in selection; dedupe must
	// collapse it so each range is counted once.
	if len(nets) != 2 {
		t.Errorf("parseLines deduped len = %d, want 2", len(nets))
	}
}

func TestValidCIDRs(t *testing.T) {
	in := []string{"1.2.3.0/24", "garbage", "  4.5.6.0/24  ", "999.1.1.1/24"}
	out := validCIDRs(in)
	if len(out) != 2 {
		t.Fatalf("validCIDRs = %v, want 2 valid entries", out)
	}
	if out[0] != "1.2.3.0/24" || out[1] != "4.5.6.0/24" {
		t.Errorf("validCIDRs = %v, want trimmed valid entries", out)
	}
}

func TestSaveRangesAndOverride(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	if err := SaveRanges([]string{"203.0.113.0/24"}, nil); err != nil {
		t.Fatalf("SaveRanges: %v", err)
	}

	// The on-disk override must take precedence over the embedded list.
	s, err := New(true, false, nil)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	got := s.V4Ranges()
	if len(got) != 1 || got[0] != "203.0.113.0/24" {
		t.Errorf("override not applied: V4Ranges = %v, want [203.0.113.0/24]", got)
	}
}

func TestSaveRangesIgnoresEmptyAndInvalid(t *testing.T) {
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	// A list with no valid CIDRs must leave the file untouched rather than
	// wiping it, so a later New falls back to the embedded ranges.
	if err := SaveRanges([]string{"garbage", ""}, nil); err != nil {
		t.Fatalf("SaveRanges: %v", err)
	}
	if _, err := os.Stat(cacheV4Path); !os.IsNotExist(err) {
		t.Errorf("expected no override file to be written, stat err = %v", err)
	}
}

func TestSweepExhaustive(t *testing.T) {
	s, err := New(false, false, []string{"192.0.2.0/30"})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	count := 0
	for ip := range s.Sweep(context.Background(), 0) {
		seen[ip.String()] = true
		count++
	}
	if count != 4 {
		t.Errorf("Sweep emitted %d, want 4", count)
	}
	if len(seen) != 4 {
		t.Errorf("Sweep emitted duplicates: %d unique of %d", len(seen), count)
	}
}

func TestSweepResumeSkip(t *testing.T) {
	s, err := New(false, false, []string{"192.0.2.0/30"})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for ip := range s.Sweep(context.Background(), 2) {
		got = append(got, ip.String())
	}
	if len(got) != 2 || got[0] != "192.0.2.2" || got[1] != "192.0.2.3" {
		t.Errorf("Sweep(skip=2) = %v, want [192.0.2.2 192.0.2.3]", got)
	}
}

func TestSweepCancel(t *testing.T) {
	s, err := New(true, false, nil) // full v4 space — far too large to drain fully
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for range s.Sweep(ctx, 0) {
		// Drain whatever was buffered before cancellation; must terminate.
	}
}
