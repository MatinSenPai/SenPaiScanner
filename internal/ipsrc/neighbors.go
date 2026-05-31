package ipsrc

import (
	"encoding/binary"
	"net"
)

// Default neighbor-scan limits for Phase 1 random scans.
const (
	DefaultNeighborRadius   = 32
	DefaultNeighborPerHit   = 12
	DefaultNeighborMaxTotal = 400
)

// NeighborsAround returns up to limit IPv4 addresses near ip that also fall
// inside one of nets. Addresses spread outward in both directions (±1, ±2, …)
// so hits in dense Cloudflare blocks can surface nearby working IPs.
func NeighborsAround(ip net.IP, nets []*net.IPNet, radius, limit int) []net.IP {
	if limit <= 0 || radius <= 0 || len(nets) == 0 {
		return nil
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return nil
	}

	base := binary.BigEndian.Uint32(ip4)
	out := make([]net.IP, 0, limit)

	for delta := uint32(1); delta <= uint32(radius) && len(out) < limit; delta++ {
		for _, sign := range []int32{int32(delta), -int32(delta)} {
			next, ok := offsetIPv4(base, sign)
			if !ok {
				continue
			}
			candidate := uint32ToIPv4(next)
			if candidate.Equal(ip) {
				continue
			}
			if !containsAnyNet(nets, candidate) {
				continue
			}
			out = append(out, candidate)
			if len(out) >= limit {
				return out
			}
		}
	}
	return out
}

func offsetIPv4(base uint32, delta int32) (uint32, bool) {
	if delta >= 0 {
		sum := uint64(base) + uint64(delta)
		if sum > 0xFFFFFFFF {
			return 0, false
		}
		return uint32(sum), true
	}
	d := uint32(-delta)
	if d > base {
		return 0, false
	}
	return base - d, true
}

func uint32ToIPv4(v uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, v)
	return ip
}

func containsAnyNet(nets []*net.IPNet, ip net.IP) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
