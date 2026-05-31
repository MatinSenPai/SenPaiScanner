package ipsrc

import (
	"net"
	"testing"
)

func TestNeighborsAroundSpreadsOutward(t *testing.T) {
	s, err := New(true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	hit := net.ParseIP("104.16.72.100")
	neighbors := NeighborsAround(hit, s.v4Nets, 8, 4)
	if len(neighbors) != 4 {
		t.Fatalf("neighbors = %d, want 4", len(neighbors))
	}
	want := []string{"104.16.72.101", "104.16.72.99", "104.16.72.102", "104.16.72.98"}
	for i, ip := range neighbors {
		if ip.String() != want[i] {
			t.Fatalf("neighbor[%d] = %s, want %s", i, ip, want[i])
		}
	}
}

func TestNeighborsAroundStaysInsideCloudflareRanges(t *testing.T) {
	s, err := New(true, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	hit := net.ParseIP("104.16.72.162")
	for _, ip := range NeighborsAround(hit, s.v4Nets, DefaultNeighborRadius, DefaultNeighborPerHit) {
		if !containsAnyNet(s.v4Nets, ip) {
			t.Fatalf("neighbor %s outside Cloudflare ranges", ip)
		}
	}
}

func TestNeighborsAroundSkipsOutsideRange(t *testing.T) {
	nets := []*net.IPNet{mustParseCIDR(t, "192.0.2.0/29")}
	hit := net.ParseIP("192.0.2.3")
	got := NeighborsAround(hit, nets, 32, 10)
	if len(got) != 7 {
		t.Fatalf("neighbors = %d, want 7 inside /29", len(got))
	}
}

func mustParseCIDR(t *testing.T, cidr string) *net.IPNet {
	t.Helper()
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatal(err)
	}
	return n
}
