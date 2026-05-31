package ipsrc

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
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
	v4Nets []*net.IPNet
	v6Nets []*net.IPNet
	rng    *rand.Rand
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

	return s, nil
}

// IPv4Nets returns the loaded IPv4 CIDR blocks (read-only slice header).
func (s *Source) IPv4Nets() []*net.IPNet {
	return s.v4Nets
}

// Random returns a single random IP from the configured ranges.
func (s *Source) Random() net.IP {
	all := append(s.v4Nets, s.v6Nets...)
	target := all[s.rng.Intn(len(all))]
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
