package prober

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
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
	Port               int
	Mode               Mode
	Tries              int
	Timeout            time.Duration
	SNI                string // empty = rotate automatically
	SpeedBytes         int64  // optional HTTP download sample size; 0 disables it
	InsecureSkipVerify bool   // skip TLS cert verification (use for Phase 1 where Phase 2 validates properly)
	WebSocketHost      string // empty = SNI
	WebSocketPath      string // empty = /
	RequireWebSocket   bool   // require a successful WebSocket probe for HTTP health
}

// WithPort returns a copy of Config targeting another remote port.
func (c Config) WithPort(port int) Config {
	c.Port = port
	return c
}

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
	r := &result.Result{
		IP:        ip,
		Port:      cfg.Port,
		ProbeMode: cfg.Mode.String(),
		Timestamp: time.Now(),
		Latencies: make([]time.Duration, cfg.Tries),
		RequireWS: cfg.RequireWebSocket,
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
			sni = sniHostnames[rand.Intn(len(sniHostnames))]
		}

		var lat time.Duration
		var tlsOk bool
		var httpStatus int
		var colo string
		var throughput float64

		switch cfg.Mode {
		case ModeTCP:
			lat = probeTCP(ctx, ip, cfg.Port, cfg.Timeout)
		case ModeTLS:
			lat, tlsOk = probeTLS(ctx, ip, cfg.Port, sni, cfg.Timeout, cfg.InsecureSkipVerify)
		case ModeHTTP:
			var wsOk bool
			lat, tlsOk, httpStatus, colo, throughput, wsOk = probeHTTP(ctx, ip, cfg.Port, sni, cfg.Timeout, cfg.SpeedBytes, cfg.InsecureSkipVerify, cfg.WebSocketHost, cfg.WebSocketPath, cfg.RequireWebSocket)
			if wsOk {
				r.WSOk = true
			}
		}

		r.Latencies[i] = lat
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

		// Small jitter between tries to avoid looking like a scanner
		if i < cfg.Tries-1 {
			jitter := time.Duration(rand.Intn(50)+10) * time.Millisecond
			select {
			case <-ctx.Done():
			case <-time.After(jitter):
			}
		}
	}

	return r
}

// probeTCP measures a raw TCP connect time.
func probeTCP(ctx context.Context, ip net.IP, port int, timeout time.Duration) time.Duration {
	addr := fmt.Sprintf("%s:%d", ip.String(), port)
	dl := time.Now().Add(timeout)
	dialCtx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()

	d := net.Dialer{}
	start := time.Now()
	conn, err := d.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return 0
	}
	lat := time.Since(start)
	conn.Close()
	return lat
}

// probeTLS measures a TLS handshake time.
func probeTLS(ctx context.Context, ip net.IP, port int, sni string, timeout time.Duration, insecure bool) (time.Duration, bool) {
	addr := fmt.Sprintf("%s:%d", ip.String(), port)
	dl := time.Now().Add(timeout)
	dialCtx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()

	d := tls.Dialer{
		NetDialer: &net.Dialer{},
		Config: &tls.Config{
			ServerName:         sni,
			InsecureSkipVerify: insecure,
			MinVersion:         tls.VersionTLS12,
		},
	}

	start := time.Now()
	conn, err := d.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		return 0, false
	}
	lat := time.Since(start)
	conn.Close()
	return lat, true
}

// probeHTTP fetches /cdn-cgi/trace to confirm the IP is a real Cloudflare edge
// and to determine the colo identifier.
func probeHTTP(ctx context.Context, ip net.IP, port int, sni string, timeout time.Duration, speedBytes int64, insecure bool, wsHost, wsPath string, requireWS bool) (
	lat time.Duration, tlsOk bool, httpStatus int, colo string, throughput float64, wsOk bool,
) {
	addr := fmt.Sprintf("%s:%d", ip.String(), port)

	// Budget split: TCP gets ¼, TLS gets ½, leaving ¼ guaranteed for the HTTP
	// GET+response. Without this, on DPI-throttled networks the TLS handshake
	// can silently consume the entire http.Client.Timeout, making the HTTP
	// phase impossible and producing false-positive packet loss.
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: timeout / 4}).DialContext(ctx, network, addr)
		},
		TLSClientConfig: &tls.Config{
			ServerName:         sni,
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecure,
		},
		DisableKeepAlives:   true,
		TLSHandshakeTimeout: timeout / 2,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	scheme := "https"
	if port == 80 {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://%s/cdn-cgi/trace", scheme, sni)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "senpaiscanner/1.0")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, false, 0, "", 0, false
	}
	lat = time.Since(start)
	defer resp.Body.Close()

	tlsOk = resp.TLS != nil
	httpStatus = resp.StatusCode
	colo = parseColoRay(resp.Header.Get("CF-Ray"))

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if traceColo := parseColoCDN(string(body)); traceColo != "" {
		colo = traceColo
	}

	if httpStatus >= 200 && httpStatus < 400 && colo != "" {
		if speedBytes > 0 {
			throughput = probeDownload(ctx, ip, port, timeout, speedBytes)
		}
		if speedBytes > 0 || requireWS {
			wsOk = probeWebSocket(ctx, ip, port, sni, wsHost, wsPath, timeout)
		}
	}

	return
}

// probeWebSocket tests whether WebSocket-grade TLS connections reach the
// Cloudflare edge without being killed by DPI. It does two things:
//
//  1. Holds the TLS connection idle for 2 s before sending any data.
//     Some DPI systems RST connections that look like long-lived TLS tunnels
//     without early data — if the connection dies during the idle hold, WSOk
//     is false.
//
//  2. Sends a WebSocket upgrade request and checks that any HTTP response
//     arrives (even 400/404). If DPI drops the connection before CF can
//     respond, WSOk is false.
//
// TLS cert verification is skipped here because the main probeHTTP call
// already verified the certificate for this IP.
func probeWebSocket(ctx context.Context, ip net.IP, port int, sni, host, path string, timeout time.Duration) bool {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	wsCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	deadline, _ := wsCtx.Deadline()

	addr := fmt.Sprintf("%s:%d", ip.String(), port)
	if host == "" {
		host = sni
	}
	path = normalizeWSPath(path)

	dialer := &net.Dialer{Timeout: timeout / 3}
	conn, err := dialer.DialContext(wsCtx, "tcp", addr)
	if err != nil {
		return false
	}
	defer conn.Close()

	tlsConn := tls.Client(conn, &tls.Config{
		ServerName:         sni,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, // cert already verified in probeHTTP
	})
	_ = tlsConn.SetDeadline(deadline)
	if err := tlsConn.HandshakeContext(wsCtx); err != nil {
		return false
	}

	// Phase 1: idle hold — detect DPI that RSTs long-lived TLS connections
	// before any application data is exchanged.
	idleHold := 2 * time.Second
	if remaining := time.Until(deadline); remaining < 2*idleHold {
		idleHold = remaining / 2
	}
	idleDeadline, ok := boundedDeadline(deadline, idleHold)
	if !ok {
		return false
	}
	_ = tlsConn.SetReadDeadline(idleDeadline)
	oneByte := make([]byte, 1)
	if _, err := tlsConn.Read(oneByte); err != nil {
		// A timeout here is EXPECTED (server speaks first only after WS upgrade).
		// Any other error (RST, EOF) means the connection was killed while idle.
		if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
			return false
		}
	}

	// Phase 2: send WebSocket upgrade and verify CF responds.
	// If DPI RSTs connections containing WS upgrade headers, we won't get a
	// response — returning false signals that WS traffic is DPI-blocked.
	wsReq := fmt.Sprintf(
		"GET %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Key: c2VucGFpc2Nhbm5lcg==\r\n"+
			"Sec-WebSocket-Version: 13\r\n"+
			"\r\n", path, host)

	writeDeadline, ok := boundedDeadline(deadline, timeout/2)
	if !ok {
		return false
	}
	_ = tlsConn.SetWriteDeadline(writeDeadline)
	if _, err := tlsConn.Write([]byte(wsReq)); err != nil {
		return false
	}

	// Read the first chunk of the response. CF will answer with at least an
	// HTTP status line (e.g. "HTTP/1.1 400 Bad Request"). If we see "HTTP/",
	// the WS upgrade reached CF — the connection is DPI-permissive.
	respBuf := make([]byte, 1024)
	readDeadline, ok := boundedDeadline(deadline, timeout/3)
	if !ok {
		return false
	}
	_ = tlsConn.SetReadDeadline(readDeadline)
	n, err := tlsConn.Read(respBuf)
	if err != nil || n == 0 {
		return false
	}

	return strings.Contains(string(respBuf[:n]), "HTTP/")
}

func boundedDeadline(global time.Time, maxWait time.Duration) (time.Time, bool) {
	if maxWait <= 0 {
		maxWait = time.Millisecond
	}
	now := time.Now()
	if !global.IsZero() && !global.After(now) {
		return time.Time{}, false
	}
	local := now.Add(maxWait)
	if !global.IsZero() && global.Before(local) {
		return global, true
	}
	return local, true
}

func normalizeWSPath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// probeDownload fetches a small sample from speed.cloudflare.com while forcing
// the TCP connection to the candidate IP. This is still not a full Xray/V2Ray
// test, but it catches many IPs that handshake cleanly and then stall on data.
func probeDownload(ctx context.Context, ip net.IP, port int, timeout time.Duration, bytes int64) float64 {
	if bytes <= 0 {
		return 0
	}

	addr := fmt.Sprintf("%s:%d", ip.String(), port)
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
	client := &http.Client{Timeout: timeout, Transport: transport}

	scheme := "https"
	if port == 80 {
		scheme = "http"
	}
	url := fmt.Sprintf("%s://speed.cloudflare.com/__down?bytes=%d", scheme, bytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0
	}
	req.Header.Set("User-Agent", "senpaiscanner/1.0")

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return 0
	}

	n, err := io.Copy(io.Discard, io.LimitReader(resp.Body, bytes))
	if err != nil || n <= 0 {
		return 0
	}
	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(n) / elapsed
}

// parseColoCDN extracts the "colo" field from /cdn-cgi/trace responses.
func parseColoCDN(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "colo=") {
			return strings.TrimSpace(strings.TrimPrefix(line, "colo="))
		}
	}
	return ""
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
