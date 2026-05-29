package ipsrc

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
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

	// On-disk overrides. When present and non-empty these take precedence over
	// the embedded lists, so UpdateRanges can refresh Cloudflare's published
	// CIDRs without rebuilding the binary.
	cacheV4Path = "cf_ranges_v4.txt"
	cacheV6Path = "cf_ranges_v6.txt"

	// streamBuffer is the channel depth between the IP generator and the worker
	// pool. It is sized so a few hundred high-concurrency workers never starve
	// waiting on the single generator goroutine, while staying bounded memory.
	streamBuffer = 256
)

// Source holds the CIDR ranges used for IP generation along with the
// cumulative address-space weights used to sample them uniformly.
type Source struct {
	v4Nets []*net.IPNet
	v6Nets []*net.IPNet

	// cumulative weights, one entry per net, used for weighted selection so
	// that a larger CIDR is chosen proportionally more often than a smaller
	// one (i.e. the result is uniform across the family's address space).
	v4Cum []float64
	v6Cum []float64
}

// New builds a Source from the embedded Cloudflare ranges plus any extra CIDRs.
func New(useV4, useV6 bool, extra []string) (*Source, error) {
	s := &Source{}

	if useV4 {
		nets, err := parseLines(rangesText(cacheV4Path, builtinV4))
		if err != nil {
			return nil, err
		}
		s.v4Nets = nets
	}

	if useV6 {
		nets, err := parseLines(rangesText(cacheV6Path, builtinV6))
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

	s.v4Cum = cumulativeWeights(s.v4Nets)
	s.v6Cum = cumulativeWeights(s.v6Nets)
	return s, nil
}

// Random returns a single random IP from the configured ranges.
//
// Selection happens in two steps. The address family is chosen in proportion
// to how many ranges each family contributes (so enabling IPv6 — whose address
// space dwarfs IPv4 by ~2^64 — does not starve IPv4 sampling). Within a family,
// a range is chosen weighted by its address-space size, which makes the host
// pick uniform across that family's space.
func (s *Source) Random() net.IP {
	v4, v6 := len(s.v4Nets), len(s.v6Nets)
	switch {
	case v4 > 0 && v6 > 0:
		if rand.IntN(v4+v6) < v4 {
			return randomFromNet(pickWeighted(s.v4Nets, s.v4Cum))
		}
		return randomFromNet(pickWeighted(s.v6Nets, s.v6Cum))
	case v6 > 0:
		return randomFromNet(pickWeighted(s.v6Nets, s.v6Cum))
	default:
		return randomFromNet(pickWeighted(s.v4Nets, s.v4Cum))
	}
}

// Stream emits random IPs on the returned channel until ctx is cancelled or
// count IPs have been sent (count <= 0 means unlimited).
//
// When count > 0 each IP is emitted at most once; duplicates are skipped via a
// seen-set bounded by count. For unlimited scans deduplication is disabled —
// tracking every emitted IP forever would leak memory, and collisions across
// Cloudflare's address space are rare enough to be harmless.
func (s *Source) Stream(ctx context.Context, count int) <-chan net.IP {
	ch := make(chan net.IP, streamBuffer)
	go func() {
		defer close(ch)

		dedupe := count > 0
		var seen map[string]struct{}
		if dedupe {
			seen = make(map[string]struct{}, count)
		}

		sent := 0
		for {
			if count > 0 && sent >= count {
				return
			}
			ip := s.Random()
			if dedupe {
				// Key on the raw address bytes rather than the dotted/colon
				// string: it is both cheaper to produce and smaller to store,
				// which matters when count reaches the millions.
				key := string(ip)
				if _, dup := seen[key]; dup {
					continue
				}
				seen[key] = struct{}{}
			}
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

// Sweep emits every address in the configured ranges sequentially (v4 first,
// then v6), skipping the first skip addresses so an interrupted scan can resume
// from a checkpoint. Unlike Stream it is exhaustive and deterministic — each
// address is emitted exactly once — which suits a full-coverage scan.
//
// Sweep is intended for IPv4 and other bounded ranges. IPv6 ranges are so large
// that a sequential walk never completes in practice; use Stream for v6.
func (s *Source) Sweep(ctx context.Context, skip int) <-chan net.IP {
	ch := make(chan net.IP, streamBuffer)
	go func() {
		defer close(ch)

		emitted := 0
		walk := func(nets []*net.IPNet) bool {
			for _, n := range nets {
				for ip := cloneIP(n.IP); n.Contains(ip); incrementIP(ip) {
					if emitted < skip {
						emitted++
						continue
					}
					select {
					case <-ctx.Done():
						return false
					case ch <- cloneIP(ip):
						emitted++
					}
				}
			}
			return true
		}
		if !walk(s.v4Nets) {
			return
		}
		walk(s.v6Nets)
	}()
	return ch
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
	client := &http.Client{Timeout: 30 * time.Second}

	fetch := func(url string) ([]string, error) {
		req, e := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if e != nil {
			return nil, e
		}
		resp, e := client.Do(req)
		if e != nil {
			return nil, fmt.Errorf("fetch %s: %w", url, e)
		}
		defer func() { _ = resp.Body.Close() }()
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
	seen := make(map[string]struct{})
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
		// Skip duplicates: a CIDR listed twice would be double-weighted in
		// selection, biasing sampling toward that range.
		key := ipNet.String()
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		nets = append(nets, ipNet)
	}
	return nets, sc.Err()
}

// rangesText returns the on-disk override at path when it exists and is
// non-empty, otherwise the embedded fallback.
func rangesText(path, fallback string) string {
	if b, err := os.ReadFile(path); err == nil && len(strings.TrimSpace(string(b))) > 0 {
		return string(b)
	}
	return fallback
}

// SaveRanges writes validated v4/v6 CIDR lists to the on-disk override files so
// they take precedence on the next scan. Invalid entries are dropped; a list
// with no valid CIDRs leaves its file untouched rather than wiping it.
func SaveRanges(v4, v6 []string) error {
	if valid := validCIDRs(v4); len(valid) > 0 {
		if err := os.WriteFile(cacheV4Path, []byte(strings.Join(valid, "\n")+"\n"), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", cacheV4Path, err)
		}
	}
	if valid := validCIDRs(v6); len(valid) > 0 {
		if err := os.WriteFile(cacheV6Path, []byte(strings.Join(valid, "\n")+"\n"), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", cacheV6Path, err)
		}
	}
	return nil
}

// validCIDRs returns the subset of in that parse as CIDRs (trimmed).
func validCIDRs(in []string) []string {
	out := make([]string, 0, len(in))
	for _, c := range in {
		c = strings.TrimSpace(c)
		if _, _, err := net.ParseCIDR(c); err == nil {
			out = append(out, c)
		}
	}
	return out
}

// cumulativeWeights builds a prefix-sum of each net's address-space size so a
// range can be selected in proportion to how many addresses it holds.
func cumulativeWeights(nets []*net.IPNet) []float64 {
	if len(nets) == 0 {
		return nil
	}
	cum := make([]float64, len(nets))
	var running float64
	for i, n := range nets {
		ones, bits := n.Mask.Size()
		running += math.Ldexp(1, bits-ones) // 2^(host bits)
		cum[i] = running
	}
	return cum
}

// pickWeighted returns a net chosen with probability proportional to its
// address-space weight. cum must be the prefix-sum produced by cumulativeWeights.
func pickWeighted(nets []*net.IPNet, cum []float64) *net.IPNet {
	total := cum[len(cum)-1]
	target := rand.Float64() * total
	i := sort.SearchFloat64s(cum, target)
	if i >= len(nets) {
		i = len(nets) - 1
	}
	return nets[i]
}

func randomFromNet(n *net.IPNet) net.IP {
	if ip4 := n.IP.To4(); ip4 != nil {
		base := binary.BigEndian.Uint32(ip4)
		mask := binary.BigEndian.Uint32(n.Mask)
		offset := rand.Uint32() &^ mask
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, base|offset)
		return ip
	}
	// IPv6: randomise the host portion byte-by-byte.
	ip := make(net.IP, len(n.IP))
	for i := range n.IP {
		host := byte(rand.IntN(256)) &^ n.Mask[i]
		ip[i] = n.IP[i] | host
	}
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
