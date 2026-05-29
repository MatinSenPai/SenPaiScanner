package prober

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

func TestClassifyErr(t *testing.T) {
	cases := []struct {
		err  error
		want result.ProbeError
	}{
		{nil, result.ErrNone},
		{context.DeadlineExceeded, result.ErrTimeout},
		{errors.New("dial tcp 1.2.3.4:443: connect: connection refused"), result.ErrRefused},
		{errors.New("read tcp: connection reset by peer"), result.ErrReset},
		{errors.New("tls: handshake failure"), result.ErrTLS},
		{errors.New("x509: certificate signed by unknown authority"), result.ErrTLS},
		{errors.New("some weird error"), result.ErrOther},
	}
	for _, c := range cases {
		if got := classifyErr(c.err); got != c.want {
			t.Errorf("classifyErr(%v) = %v, want %v", c.err, got, c.want)
		}
	}
}

func TestProbeRecordsTimeoutError(t *testing.T) {
	// 203.0.113.0/24 is TEST-NET-3 (RFC 5737): routable-looking but black-holed,
	// so a probe with a tiny timeout should record a timeout, not a clean fail.
	r := Probe(t.Context(), net.ParseIP("203.0.113.1"), Config{
		Mode:    ModeTCP,
		Tries:   1,
		Port:    443,
		Timeout: 50 * time.Millisecond,
	})
	if len(r.Errors) != 1 {
		t.Fatalf("Errors len = %d, want 1", len(r.Errors))
	}
	// We don't assert the exact class (a network may RST instead of dropping),
	// only that a connection failure was classified as non-success.
	if r.Latencies[0] == 0 && r.Errors[0] == result.ErrNone {
		t.Fatalf("failed try recorded ErrNone; want a failure class")
	}
}

func TestIPPortFormatsIPv6(t *testing.T) {
	got := ipPort(net.ParseIP("2606:4700::6810:84e5"), 443)
	if got != "[2606:4700::6810:84e5]:443" {
		t.Fatalf("ipPort IPv6 = %q", got)
	}
}

func TestCloudflarePortClassification(t *testing.T) {
	cases := []struct {
		port      int
		known     bool
		cleartext bool
	}{
		{443, true, false},
		{2053, true, false},
		{8443, true, false},
		{80, true, true},
		{8080, true, true},
		{2052, true, true},
		{12345, false, false}, // unknown -> not CF, defaults to TLS
	}
	for _, c := range cases {
		if got := IsCloudflarePort(c.port); got != c.known {
			t.Errorf("IsCloudflarePort(%d) = %v, want %v", c.port, got, c.known)
		}
		if got := isCleartextPort(c.port); got != c.cleartext {
			t.Errorf("isCleartextPort(%d) = %v, want %v", c.port, got, c.cleartext)
		}
	}
}

func TestThroughputBpsExcludesSetup(t *testing.T) {
	base := time.Now()
	start := base
	firstByte := base.Add(900 * time.Millisecond) // 0.9s of TCP+TLS setup
	end := base.Add(1900 * time.Millisecond)      // 1.0s of actual transfer
	const n = 1_000_000

	got := throughputBps(n, start, firstByte, end)
	// Data window is ~1s, so throughput ≈ 1,000,000 B/s. The naive (start→end)
	// calc would be ~526,000 B/s — about half — which is the bug we fixed.
	if got < 950_000 || got > 1_050_000 {
		t.Fatalf("throughputBps = %.0f B/s, want ~1,000,000 (setup excluded)", got)
	}

	// Without a first-byte timestamp, fall back to the full wall-clock window.
	fallback := throughputBps(n, start, time.Time{}, end)
	if fallback < 510_000 || fallback > 540_000 {
		t.Fatalf("fallback throughputBps = %.0f B/s, want ~526,000", fallback)
	}

	if throughputBps(0, start, firstByte, end) != 0 {
		t.Fatalf("zero bytes must yield 0 throughput")
	}
}

func TestProbeDefaultsTriesAndPort(t *testing.T) {
	r := Probe(t.Context(), net.ParseIP("127.0.0.1"), Config{Mode: ModeTCP})
	if r.Port != 443 {
		t.Fatalf("default port = %d, want 443", r.Port)
	}
	if len(r.Latencies) != 1 {
		t.Fatalf("default tries len = %d, want 1", len(r.Latencies))
	}
}
