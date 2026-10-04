package scanjob

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cidrSample   = 256  // addresses taken from one CIDR (all of it when it is a /24 or smaller)
	rangeMax     = 4096 // addresses accepted from one "a-b" range
	resolveLimit = 8    // parallel DNS lookups
)

// Targets is the parsed form of pasted text or an ips.txt file.
type Targets struct {
	IPs     []net.IP
	Skipped []string // entries that could not be used, with the reason
	Domains int      // how many entries were domains that got resolved
}

// ParseTargets reads IPs, CIDRs, "a.b.c.d-e.f.g.h" / "a.b.c.d-N" ranges and domain names from free text.
// Entries may be separated by whitespace, commas or semicolons; "#" starts a comment. A ":port" suffix is
// ignored. Domains are resolved to their A/AAAA records (so a Cloudflare-fronted hostname works as a target).
// The result is de-duplicated and shuffled so a partial scan is not biased towards the start of a range.
func ParseTargets(ctx context.Context, text string) Targets {
	var t Targets
	seen := map[string]bool{}
	add := func(ip net.IP) {
		if k := ip.String(); !seen[k] {
			seen[k] = true
			t.IPs = append(t.IPs, ip)
		}
	}
	var domains []string
	for _, tok := range tokens(text) {
		host := stripPort(tok)
		switch {
		case net.ParseIP(host) != nil:
			add(net.ParseIP(host))
		case strings.Contains(host, "/"):
			ips, err := fromCIDR(host)
			if err != nil {
				t.Skipped = append(t.Skipped, fmt.Sprintf("%s (%v)", tok, err))
				continue
			}
			for _, ip := range ips {
				add(ip)
			}
		case strings.Contains(host, "-") && net.ParseIP(strings.SplitN(host, "-", 2)[0]) != nil:
			ips, err := fromRange(host)
			if err != nil {
				t.Skipped = append(t.Skipped, fmt.Sprintf("%s (%v)", tok, err))
				continue
			}
			for _, ip := range ips {
				add(ip)
			}
		case looksLikeDomain(host):
			domains = append(domains, strings.ToLower(host))
		default:
			t.Skipped = append(t.Skipped, tok+" (not an IP, CIDR, range or domain)")
		}
	}
	resolved, failed := resolveAll(ctx, domains)
	t.Domains = len(uniq(domains)) - len(failed)
	for _, ip := range resolved {
		add(ip)
	}
	for _, d := range failed {
		t.Skipped = append(t.Skipped, d+" (could not be resolved)")
	}
	rand.Shuffle(len(t.IPs), func(i, j int) { t.IPs[i], t.IPs[j] = t.IPs[j], t.IPs[i] })
	return t
}

func tokens(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		out = append(out, strings.FieldsFunc(line, func(r rune) bool {
			return r == ',' || r == ';' || r == ' ' || r == '\t' || r == '\r'
		})...)
	}
	return out
}

// stripPort turns "1.2.3.4:443", "[::1]:443" and "example.com:2053" into the host part; a bare IPv6 is kept.
func stripPort(s string) string {
	if h, p, err := net.SplitHostPort(s); err == nil {
		if _, perr := strconv.Atoi(p); perr == nil {
			return h
		}
	}
	return strings.Trim(s, "[]")
}

func looksLikeDomain(s string) bool {
	if len(s) > 253 || !strings.Contains(s, ".") || strings.HasPrefix(s, ".") || strings.HasSuffix(s, ".") {
		return false
	}
	for _, r := range s {
		ok := r == '.' || r == '-' || r == '_' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r > 127
		if !ok {
			return false
		}
	}
	return true
}

func fromCIDR(s string) ([]net.IP, error) {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	base := n.IP.To4()
	if base == nil {
		return sampleIPv6CIDR(n), nil
	}
	ones, bits := n.Mask.Size()
	size := uint64(1) << uint(bits-ones)
	start := binary.BigEndian.Uint32(base)
	pick := func(i uint64) net.IP {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, start+uint32(i))
		return ip
	}
	var out []net.IP
	if size <= cidrSample {
		for i := uint64(0); i < size; i++ {
			out = append(out, pick(i))
		}
		return out, nil
	}
	for _, i := range rand.Perm(int(min(size, 1<<20)))[:cidrSample] { // ponytail: samples the first 1M addresses of a huge CIDR
		out = append(out, pick(uint64(i)))
	}
	return out, nil
}

// sampleIPv6CIDR expands small IPv6 subnets and samples larger ones. IPv6
// pools are too large to enumerate in practice, so cap them at cidrSample
// candidates just like large IPv4 CIDRs.
func sampleIPv6CIDR(n *net.IPNet) []net.IP {
	ones, bits := n.Mask.Size()
	hostBits := bits - ones
	if hostBits <= 8 {
		var out []net.IP
		for ip := cloneIP(n.IP); n.Contains(ip); incrementIP(ip) {
			out = append(out, cloneIP(ip))
		}
		return out
	}

	out := make([]net.IP, 0, cidrSample)
	seen := make(map[string]struct{}, cidrSample)
	for attempts := 0; len(out) < cidrSample && attempts < cidrSample*40; attempts++ {
		ip := cloneIP(n.IP)
		for i := range ip {
			hostMask := ^n.Mask[i]
			ip[i] |= byte(rand.Intn(256)) & hostMask
		}
		key := ip.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, ip)
	}
	return out
}

// fromRange handles "1.2.3.4-1.2.3.40" and the short form "1.2.3.4-40" (last octet).
func fromRange(s string) ([]net.IP, error) {
	a, b, _ := strings.Cut(s, "-")
	lo := net.ParseIP(a).To4()
	if lo == nil {
		return nil, fmt.Errorf("only IPv4 ranges are supported")
	}
	var hi net.IP
	if n, err := strconv.Atoi(b); err == nil {
		if n < int(lo[3]) || n > 255 {
			return nil, fmt.Errorf("bad last octet")
		}
		hi = net.IPv4(lo[0], lo[1], lo[2], byte(n)).To4()
	} else if hi = net.ParseIP(b).To4(); hi == nil {
		return nil, fmt.Errorf("bad range end")
	}
	from, to := binary.BigEndian.Uint32(lo), binary.BigEndian.Uint32(hi)
	if to < from {
		return nil, fmt.Errorf("range end is before its start")
	}
	if to-from+1 > rangeMax {
		return nil, fmt.Errorf("range has more than %d addresses", rangeMax)
	}
	var out []net.IP
	for v := from; ; v++ {
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, v)
		out = append(out, ip)
		if v == to {
			break
		}
	}
	return out, nil
}

// lookupIP is a variable so tests can resolve without the network.
var lookupIP = net.DefaultResolver.LookupIPAddr

func resolveAll(ctx context.Context, domains []string) (ips []net.IP, failed []string) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, resolveLimit)
	for _, d := range uniq(domains) {
		wg.Add(1)
		sem <- struct{}{}
		go func(d string) {
			defer wg.Done()
			defer func() { <-sem }()
			c, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			addrs, err := lookupIP(c, d)
			mu.Lock()
			defer mu.Unlock()
			if err != nil || len(addrs) == 0 {
				failed = append(failed, d)
				return
			}
			for _, a := range addrs {
				ips = append(ips, a.IP)
			}
		}(d)
	}
	wg.Wait()
	return
}

func uniq(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
