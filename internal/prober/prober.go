package prober

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

// sniHostnames is a list of well-known Cloudflare hostnames used as SNI values.
// Rotating SNI reduces the chance of deep-packet inspection blackholing.
var sniHostnames = []string{
	"speed.cloudflare.com",
	"www.cloudflare.com",
	"cloudflare.com",
	"1.1.1.1.cdn.cloudflare.net",
	"blog.cloudflare.com",
}

// Config holds parameters for a single probe session.
type Config struct {
	Port       int
	Mode       Mode
	Tries      int
	Timeout    time.Duration
	SNI        string // empty = rotate automatically
	SpeedBytes int64  // optional HTTP download sample size; 0 disables it

	// SpeedStreams splits the optional download sample over parallel HTTP
	// requests. Values <=0 default to 1. A small value (2-4) better represents
	// proxy throughput on throttled networks without turning a scan into a full
	// speed test.
	SpeedStreams int

	// SpeedDuration caps the optional download sample. Zero uses Timeout.
	SpeedDuration time.Duration

	// Jitter is the upper bound of a random pause inserted between tries to
	// look less like a sequential scanner. Zero (the default) disables it —
	// "turbo" mode, which is dramatically faster on large scans because the
	// per-try sleep dominates wall-clock time for fast IPs. Set it to a small
	// value (tens of ms) for a "stealth" pass on DPI-sensitive networks.
	Jitter time.Duration
}

// StealthJitter is the recommended inter-try pause for a low-profile scan.
const StealthJitter = 60 * time.Millisecond

// Mode selects the probe type.
type Mode int

const (
	ModeTCP  Mode = iota // bare TCP connect
	ModeTLS              // TLS handshake (no HTTP)
	ModeHTTP             // full HTTPS GET /cdn-cgi/trace
)

func (m Mode) String() string {
	switch m {
	case ModeTLS:
		return "tls"
	case ModeHTTP:
		return "http"
	default:
		return "tcp"
	}
}

// ParseMode parses a mode string.
func ParseMode(s string) (Mode, error) {
	switch strings.ToLower(s) {
	case "tcp":
		return ModeTCP, nil
	case "tls":
		return ModeTLS, nil
	case "http", "https":
		return ModeHTTP, nil
	default:
		return ModeTCP, fmt.Errorf("unknown mode %q (want tcp|tls|http)", s)
	}
}

// Probe runs a full measurement session against ip and returns a Result.
func Probe(ctx context.Context, ip net.IP, cfg Config) *result.Result {
	if cfg.Tries <= 0 {
		cfg.Tries = 1
	}
	if cfg.Port <= 0 {
		cfg.Port = 443
	}
	r := &result.Result{
		IP:        ip,
		Port:      cfg.Port,
		ProbeMode: cfg.Mode.String(),
		Timestamp: time.Now(),
		Latencies: make([]time.Duration, cfg.Tries),
		Errors:    make([]result.ProbeError, cfg.Tries),
	}
	if cfg.Mode == ModeHTTP && cfg.SpeedBytes > 0 {
		r.SpeedTested = true
	}

	for i := 0; i < cfg.Tries; i++ {
		if ctx.Err() != nil {
			break
		}
		sni := cfg.SNI
		if sni == "" && cfg.Mode == ModeHTTP {
			sni = "speed.cloudflare.com"
		} else if sni == "" {
			sni = sniHostnames[rand.IntN(len(sniHostnames))]
		}

		var lat time.Duration
		var tlsOk bool
		var httpStatus int
		var colo string
		var throughput float64
		var perr result.ProbeError

		switch cfg.Mode {
		case ModeTCP:
			lat, perr = probeTCP(ctx, ip, cfg.Port, cfg.Timeout)
		case ModeTLS:
			lat, tlsOk, perr = probeTLS(ctx, ip, cfg.Port, sni, cfg.Timeout)
		case ModeHTTP:
			lat, tlsOk, httpStatus, colo, throughput, perr = probeHTTP(ctx, ip, cfg.Port, sni, cfg.Timeout, cfg.SpeedBytes, cfg.SpeedStreams, cfg.SpeedDuration)
		}

		r.Latencies[i] = lat
		r.Errors[i] = perr
		if tlsOk {
			r.TLSOk = true
		}
		if httpStatus != 0 {
			r.HTTPStatus = httpStatus
		}
		if colo != "" {
			r.Colo = colo
		}
		if throughput > 0 {
			r.Throughput = throughput
		}

		// Optional jitter between tries to avoid looking like a scanner. Skipped
		// entirely in turbo mode (cfg.Jitter == 0), which is the default.
		if i < cfg.Tries-1 && cfg.Jitter > 0 {
			pause := time.Duration(rand.Int64N(int64(cfg.Jitter)))
			select {
			case <-ctx.Done():
			case <-time.After(pause):
			}
		}
	}

	return r
}

// probeTCP measures a raw TCP connect time.
func probeTCP(ctx context.Context, ip net.IP, port int, timeout time.Duration) (time.Duration, result.ProbeError) {
	addr := ipPort(ip, port)
	dl := time.Now().Add(timeout)
	dialCtx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()

	d := net.Dialer{}
	start := time.Now()
	conn, err := d.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return 0, classifyErr(err)
	}
	lat := time.Since(start)
	_ = conn.Close()
	return lat, result.ErrNone
}

// probeTLS measures a TLS handshake time.
func probeTLS(ctx context.Context, ip net.IP, port int, sni string, timeout time.Duration) (time.Duration, bool, result.ProbeError) {
	addr := ipPort(ip, port)
	dl := time.Now().Add(timeout)
	dialCtx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()

	d := tls.Dialer{
		NetDialer: &net.Dialer{},
		Config: &tls.Config{
			ServerName:         sni,
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}

	start := time.Now()
	conn, err := d.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return 0, false, classifyErr(err)
	}
	lat := time.Since(start)
	_ = conn.Close()
	return lat, true, result.ErrNone
}

// probeHTTP fetches /cdn-cgi/trace to confirm the IP is a real Cloudflare edge
// and to determine the colo identifier.
func probeHTTP(ctx context.Context, ip net.IP, port int, sni string, timeout time.Duration, speedBytes int64, speedStreams int, speedDuration time.Duration) (
	lat time.Duration, tlsOk bool, httpStatus int, colo string, throughput float64, perr result.ProbeError,
) {
	addr := ipPort(ip, port)

	// Budget split: TCP gets ¼, TLS gets ½, leaving ¼ guaranteed for the HTTP
	// GET+response. Without this, on DPI-throttled networks the TLS handshake
	// can silently consume the entire http.Client.Timeout, making the HTTP
	// phase impossible and producing false-positive packet loss.
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: timeout / 4}).DialContext(ctx, network, addr)
		},
		TLSClientConfig: &tls.Config{
			ServerName: sni,
			MinVersion: tls.VersionTLS12,
		},
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: timeout / 2,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	scheme := "https"
	if isCleartextPort(port) {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s/cdn-cgi/trace", scheme, sni)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		perr = result.ErrOther
		return
	}
	req.Header.Set("User-Agent", "senpaiscanner/1.0")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, 0, "", 0, classifyErr(err)
	}
	lat = time.Since(start)
	defer func() { _ = resp.Body.Close() }()

	tlsOk = resp.TLS != nil
	httpStatus = resp.StatusCode
	colo = parseColoRay(resp.Header.Get("CF-Ray"))

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if traceColo := parseColoCDN(string(body)); traceColo != "" {
		colo = traceColo
	}
	if speedBytes > 0 && httpStatus >= 200 && httpStatus < 400 && colo != "" {
		throughput = probeDownload(ctx, ip, port, timeout, speedBytes, speedStreams, speedDuration)
	}
	// The connection succeeded but, if the response isn't a healthy Cloudflare
	// edge trace, record it as an HTTP-class failure rather than a clean success.
	if httpStatus < 200 || httpStatus >= 400 || colo == "" {
		perr = result.ErrHTTP
	}
	return
}

// probeDownload fetches a small sample from speed.cloudflare.com while forcing
// the TCP connection to the candidate IP. This is still not a full Xray/V2Ray
// test, but it catches many IPs that handshake cleanly and then stall on data.
func probeDownload(ctx context.Context, ip net.IP, port int, timeout time.Duration, bytes int64, streams int, duration time.Duration) float64 {
	if bytes <= 0 {
		return 0
	}
	if streams <= 0 {
		streams = 1
	}
	if streams > 8 {
		streams = 8
	}
	if duration <= 0 || duration > timeout {
		duration = timeout
	}

	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	addr := ipPort(ip, port)
	scheme := "https"
	if isCleartextPort(port) {
		scheme = "http"
	}

	start := time.Now()
	var total atomic.Int64

	// firstByte holds the unix-nanos timestamp of the earliest first response
	// byte across all streams (0 = none yet). Measuring throughput from this
	// point — rather than from before the TCP+TLS handshake — excludes connection
	// setup, which otherwise makes a fast link look slow on small samples.
	var firstByte atomic.Int64
	markFirstByte := func() {
		now := time.Now().UnixNano()
		for {
			cur := firstByte.Load()
			if cur != 0 && cur <= now {
				return
			}
			if firstByte.CompareAndSwap(cur, now) {
				return
			}
		}
	}

	var wg sync.WaitGroup
	perStream := bytes / int64(streams)
	if perStream <= 0 {
		perStream = bytes
		streams = 1
	}

	for i := 0; i < streams; i++ {
		want := perStream
		if i == streams-1 {
			want += bytes - perStream*int64(streams)
		}
		wg.Add(1)
		go func(want int64) {
			defer wg.Done()
			transport := &http.Transport{
				DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
					return (&net.Dialer{Timeout: timeout / 4}).DialContext(ctx, network, addr)
				},
				TLSClientConfig: &tls.Config{
					ServerName: "speed.cloudflare.com",
					MinVersion: tls.VersionTLS12,
				},
				DisableKeepAlives:   true,
				TLSHandshakeTimeout: timeout / 2,
			}
			defer transport.CloseIdleConnections()

			client := &http.Client{Timeout: duration, Transport: transport}
			url := fmt.Sprintf("%s://speed.cloudflare.com/__down?bytes=%d", scheme, want)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return
			}
			req.Header.Set("User-Agent", "senpaiscanner/1.0")
			trace := &httptrace.ClientTrace{GotFirstResponseByte: markFirstByte}
			req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode < 200 || resp.StatusCode >= 400 {
				return
			}

			n, err := io.Copy(io.Discard, io.LimitReader(resp.Body, want))
			if err == nil && n > 0 {
				total.Add(n)
			}
		}(want)
	}
	wg.Wait()

	end := time.Now()
	var fbTime time.Time
	if fb := firstByte.Load(); fb != 0 {
		fbTime = time.Unix(0, fb)
	}
	return throughputBps(total.Load(), start, fbTime, end)
}

// throughputBps computes bytes/sec over the data-transfer window. It prefers the
// span from the first received byte to the end (excluding TCP+TLS setup) and
// falls back to start→end when no first byte was observed. Extracted as a pure
// function so the timing logic can be unit-tested without the network.
func throughputBps(n int64, start, firstByte, end time.Time) float64 {
	if n <= 0 {
		return 0
	}
	elapsed := end.Sub(start).Seconds()
	if !firstByte.IsZero() {
		if dl := end.Sub(firstByte).Seconds(); dl > 0 {
			elapsed = dl
		}
	}
	if elapsed <= 0 {
		return 0
	}
	return float64(n) / elapsed
}

// classifyErr maps a network/transport error to a result.ProbeError class.
//
// It deliberately leans on net.Error.Timeout and substring matching rather than
// platform-specific syscall constants (ECONNREFUSED is WSAECONNREFUSED on
// Windows, etc.), so the classification stays correct across OSes without build
// tags.
func classifyErr(err error) result.ProbeError {
	if err == nil {
		return result.ErrNone
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return result.ErrTimeout
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return result.ErrTimeout
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "refused"):
		return result.ErrRefused
	case strings.Contains(msg, "reset"):
		return result.ErrReset
	case strings.Contains(msg, "tls"),
		strings.Contains(msg, "certificate"),
		strings.Contains(msg, "handshake"):
		return result.ErrTLS
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return result.ErrTimeout
	default:
		return result.ErrOther
	}
}

// parseColoCDN extracts the "colo" field from /cdn-cgi/trace responses.
func parseColoCDN(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "colo=") {
			return strings.TrimPrefix(line, "colo=")
		}
	}
	return ""
}

func ipPort(ip net.IP, port int) string {
	return net.JoinHostPort(ip.String(), strconv.Itoa(port))
}

func parseColoRay(ray string) string {
	parts := strings.Split(ray, "-")
	if len(parts) < 2 {
		return ""
	}
	colo := strings.TrimSpace(parts[len(parts)-1])
	if len(colo) < 3 {
		return ""
	}
	return strings.ToUpper(colo[:3])
}
