package ipsrc

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	_ "embed"
)

//go:embed ranges_v4.txt
var builtinV4 string

//go:embed ranges_v6.txt
var builtinV6 string

const (
	cfIPsV4URL = "https://www.cloudflare.com/ips-v4/"
	cfIPsV6URL = "https://www.cloudflare.com/ips-v6/"

)

// Source holds the CIDR ranges used for IP generation.
type Source struct {
	v4Nets     []*net.IPNet
	v6Nets     []*net.IPNet
	v4CumSizes []int64
	v6CumSizes []float64
	rng        *rand.Rand
}

// Options controls how a Source is built.
type Options struct {
	// UseBuiltin controls whether embedded Cloudflare ranges are loaded before
	// any extra CIDRs are added. Set it to false when user-provided CIDRs should
	// be treated as an exact scan scope rather than as additions to Cloudflare's
	// full published ranges.
	UseBuiltin bool
}

// New builds a Source from the embedded Cloudflare ranges plus optional extra
// CIDRs.
func New(useV4, useV6 bool, extra []string) (*Source, error) {
	return NewWithOptions(useV4, useV6, extra, Options{UseBuiltin: true})
}

// NewWithOptions builds a Source with explicit control over built-in ranges.
func NewWithOptions(useV4, useV6 bool, extra []string, opts Options) (*Source, error) {
	s := &Source{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if opts.UseBuiltin && useV4 {
		nets, err := parseLines(builtinV4)
		if err != nil {
			return nil, err
		}
		s.v4Nets = nets
	}

	if opts.UseBuiltin && useV6 {
		nets, err := parseLines(builtinV6)
		if err != nil {
			return nil, err
		}
		s.v6Nets = nets
	}

	for _, cidr := range extra {
		_, ipNet, err := net.ParseCIDR(strings.TrimSpace(cidr))
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
		}
		if ipNet.IP.To4() != nil {
			s.v4Nets = append(s.v4Nets, ipNet)
		} else {
			s.v6Nets = append(s.v6Nets, ipNet)
		}
	}

	if len(s.v4Nets)+len(s.v6Nets) == 0 {
		return nil, fmt.Errorf("no IP ranges available (enable --v4 and/or --v6)")
	}

	// Calculate cumulative sizes for weighted random selection
	if len(s.v4Nets) > 0 {
		s.v4CumSizes = make([]int64, len(s.v4Nets))
		var sum int64
		for i, n := range s.v4Nets {
			ones, bits := n.Mask.Size()
			size := int64(1) << (bits - ones)
			sum += size
			s.v4CumSizes[i] = sum
		}
	}

	if len(s.v6Nets) > 0 {
		s.v6CumSizes = make([]float64, len(s.v6Nets))
		var sum float64
		for i, n := range s.v6Nets {
			ones, bits := n.Mask.Size()
			size := math.Exp2(float64(bits - ones))
			sum += size
			s.v6CumSizes[i] = sum
		}
	}

	return s, nil
}

// IPv4Nets returns the loaded IPv4 CIDR blocks (read-only slice header).
func (s *Source) IPv4Nets() []*net.IPNet {
	return s.v4Nets
}

// Random returns a single random IP from the configured ranges.
func (s *Source) Random() net.IP {
	var target *net.IPNet
	useV6 := false

	if len(s.v4Nets) > 0 && len(s.v6Nets) > 0 {
		// Both v4 and v6 are active. Pick between them.
		// To match the original behavior's balance, we choose proportional to the number of subnets.
		totalSubnets := len(s.v4Nets) + len(s.v6Nets)
		if s.rng.Intn(totalSubnets) >= len(s.v4Nets) {
			useV6 = true
		}
	} else if len(s.v6Nets) > 0 {
		useV6 = true
	}

	if useV6 {
		totalWeight := s.v6CumSizes[len(s.v6CumSizes)-1]
		r := s.rng.Float64() * totalWeight
		idx := sort.Search(len(s.v6CumSizes), func(i int) bool {
			return s.v6CumSizes[i] >= r
		})
		if idx >= len(s.v6CumSizes) {
			idx = len(s.v6CumSizes) - 1
		}
		target = s.v6Nets[idx]
	} else {
		totalSize := s.v4CumSizes[len(s.v4CumSizes)-1]
		r := s.rng.Int63n(totalSize)
		idx := sort.Search(len(s.v4CumSizes), func(i int) bool {
			return s.v4CumSizes[i] > r
		})
		if idx >= len(s.v4CumSizes) {
			idx = len(s.v4CumSizes) - 1
		}
		target = s.v4Nets[idx]
	}

	return randomFromNet(target, s.rng)
}

// Stream emits random IPs on the returned channel until ctx is cancelled or
// count IPs have been sent (count <= 0 means unlimited).
//
// Each IP is emitted at most once per call: duplicates are silently skipped.
// For very large counts relative to the available address space the loop may
// spin for longer, but Cloudflare's published ranges are large enough that
// this is not a practical concern for the scan sizes the TUI exposes.
func (s *Source) Stream(ctx context.Context, count int) <-chan net.IP {
	ch := make(chan net.IP, 64)
	go func() {
		defer close(ch)
		seen := make(map[string]struct{})
		sent := 0
		for {
			if count > 0 && sent >= count {
				return
			}
			ip := s.Random()
			key := ip.String()
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			select {
			case <-ctx.Done():
				return
			case ch <- ip:
				sent++
			}
		}
	}()
	return ch
}

// MahsaNGV4Stream emits a MahsaNG-style Cloudflare candidate batch: unique
// IPv4 addresses, selected with each individual address in the configured
// CIDRs having equal probability. Small pools are enumerated once and
// shuffled instead of being repeatedly sampled. The caller controls the batch
// size; no MahsaNG-specific cap is imposed.
func (s *Source) MahsaNGV4Stream(ctx context.Context, count int) <-chan net.IP {
	if count <= 0 {
		return s.Stream(ctx, count)
	}
	items := s.MahsaNGV4Sample(count)
	ch := make(chan net.IP, min(len(items), 64))
	go func() {
		defer close(ch)
		for _, ip := range items {
			select {
			case <-ctx.Done():
				return
			case ch <- ip:
			}
		}
	}()
	return ch
}

// MahsaNGV4Sample materialises a single MahsaNG-style candidate batch. It
// intentionally includes every address in a small CIDR, including its network
// and broadcast addresses, matching the reference scanner's range handling.
func (s *Source) MahsaNGV4Sample(count int) []net.IP {
	if len(s.v4Nets) == 0 || count <= 0 {
		return nil
	}
	limit := count

	total := s.v4CumSizes[len(s.v4CumSizes)-1]
	if total <= int64(limit) {
		items := make([]net.IP, 0, total)
		for _, n := range s.v4Nets {
			base, size, ok := ipv4NetworkAndSize(n)
			if !ok {
				continue
			}
			for offset := int64(0); offset < size; offset++ {
				items = append(items, ipv4FromUint32(base+uint32(offset)))
			}
		}
		s.rng.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
		return items
	}

	selected := make(map[uint32]struct{}, limit)
	items := make([]net.IP, 0, limit)
	attemptLimit := int64(limit * 40)
	for attempts := int64(0); len(items) < limit && attempts < attemptLimit; attempts++ {
		index := s.rng.Int63n(total)
		before := int64(0)
		for _, n := range s.v4Nets {
			base, size, ok := ipv4NetworkAndSize(n)
			if !ok {
				continue
			}
			after := before + size
			if index < after {
				value := base + uint32(index-before)
				if _, exists := selected[value]; !exists {
					selected[value] = struct{}{}
					items = append(items, ipv4FromUint32(value))
				}
				break
			}
			before = after
		}
	}
	return items
}

// FromCIDR expands a single CIDR string into a channel of all its IPs.
// For large ranges use caution — prefer Stream for /16 and above.
func FromCIDR(ctx context.Context, cidr string) (<-chan net.IP, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}

	ch := make(chan net.IP, 128)
	go func() {
		defer close(ch)
		for ip := cloneIP(ipNet.IP); ipNet.Contains(ip); incrementIP(ip) {
			select {
			case <-ctx.Done():
				return
			case ch <- cloneIP(ip):
			}
		}
	}()
	return ch, nil
}

// UpdateRanges fetches the latest Cloudflare IP ranges from cloudflare.com.
// Returns the raw CIDRs for v4 and v6.
func UpdateRanges(ctx context.Context) (v4, v6 []string, err error) {
	fetch := func(url string) ([]string, error) {
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if e != nil {
			return nil, e
		}
		resp, e := http.DefaultClient.Do(req)
		if e != nil {
			return nil, fmt.Errorf("fetch %s: %w", url, e)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
		}
		var lines []string
		sc := bufio.NewScanner(resp.Body)
		for sc.Scan() {
			l := strings.TrimSpace(sc.Text())
			if l != "" {
				lines = append(lines, l)
			}
		}
		return lines, sc.Err()
	}

	v4, err = fetch(cfIPsV4URL)
	if err != nil {
		return
	}
	v6, err = fetch(cfIPsV6URL)
	return
}

// V4Ranges returns the currently loaded v4 nets as CIDR strings.
func (s *Source) V4Ranges() []string {
	return netsToStrings(s.v4Nets)
}

// V6Ranges returns the currently loaded v6 nets as CIDR strings.
func (s *Source) V6Ranges() []string {
	return netsToStrings(s.v6Nets)
}

// MarshalJSON allows serialising the current source ranges.
func (s *Source) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string][]string{
		"v4": s.V4Ranges(),
		"v6": s.V6Ranges(),
	})
}

// --- helpers ----------------------------------------------------------------

func parseLines(raw string) ([]*net.IPNet, error) {
	var nets []*net.IPNet
	sc := bufio.NewScanner(strings.NewReader(raw))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		_, ipNet, err := net.ParseCIDR(line)
		if err != nil {
			return nil, fmt.Errorf("parsing %q: %w", line, err)
		}
		nets = append(nets, ipNet)
	}
	return nets, sc.Err()
}

func randomFromNet(n *net.IPNet, rng *rand.Rand) net.IP {
	ip4 := n.IP.To4()
	if ip4 != nil {
		base := binary.BigEndian.Uint32(ip4)
		mask := binary.BigEndian.Uint32([]byte(n.Mask))
		size := ^mask
		offset := rng.Uint32() & size
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, base|offset)
		return ip
	}
	// IPv6: randomise the host portion byte-by-byte
	ip := make(net.IP, len(n.IP))
	copy(ip, n.IP)
	for i, b := range n.Mask {
		host := byte(rng.Intn(256)) &^ b
		ip[i] = n.IP[i] | host
	}
	return ip
}

func ipv4NetworkAndSize(n *net.IPNet) (uint32, int64, bool) {
	ip := n.IP.To4()
	if ip == nil {
		return 0, 0, false
	}
	ones, bits := n.Mask.Size()
	if bits != 32 || ones < 0 {
		return 0, 0, false
	}
	return binary.BigEndian.Uint32(ip), int64(1) << (bits - ones), true
}

func ipv4FromUint32(value uint32) net.IP {
	ip := make(net.IP, net.IPv4len)
	binary.BigEndian.PutUint32(ip, value)
	return ip
}

func cloneIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func netsToStrings(nets []*net.IPNet) []string {
	s := make([]string, len(nets))
	for i, n := range nets {
		s[i] = n.String()
	}
	return s
}
