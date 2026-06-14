package xraytest

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	xcore "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
	_ "github.com/xtls/xray-core/main/distro/all" // register all xray features
)

var portCounter atomic.Int32
var stdioMu sync.Mutex

const (
	speedSampleBytes     = 512 * 1024
	speedSampleBytesFast = 128 * 1024
	speedMinBytes        = 8 * 1024
	traceProbeURL        = "https://cp.cloudflare.com/cdn-cgi/trace"
)

var traceProbeURLs = []string{
	traceProbeURL,
	"https://cloudflare.com/cdn-cgi/trace",
}

func init() {
	portCounter.Store(20000)
}

// nextPort returns the next available port for testing.
func nextPort() int {
	return int(portCounter.Add(1))
}

// ValidationResult holds the outcome of testing a VLESS config through xray.
type ValidationResult struct {
	IP               string
	Port             int
	Success          bool
	Latency          time.Duration // time to first byte
	Throughput       float64       // bytes/sec for download test
	BytesRecv        int64
	UploadThroughput float64 // bytes/sec for upload test (0 if not tested)
	UploadBytesSent  int64
	Error            string
	Transport        string // ws, grpc, xhttp
	Retries          int    // how many attempts were needed
}

// ValidateConfig starts an xray instance with the given config, sends test
// traffic through it, and returns the result. Retries once on failure.
func ValidateConfig(ctx context.Context, cfg *VLESSConfig, timeout time.Duration) *ValidationResult {
	res := validateOnce(ctx, cfg, timeout)
	if !res.Success {
		// Retry once — DPI is flaky
		time.Sleep(500 * time.Millisecond)
		res2 := validateOnce(ctx, cfg, timeout)
		res2.Retries = 1
		if res2.Success {
			return res2
		}
		res.Retries = 1
	}
	return res
}

func validateOnce(ctx context.Context, cfg *VLESSConfig, timeout time.Duration) *ValidationResult {
	res := &ValidationResult{
		IP:        cfg.Address,
		Port:      cfg.Port,
		Transport: cfg.Network,
	}

	socksPort := nextPort()

	configJSON, err := BuildXrayConfig(cfg, socksPort)
	if err != nil {
		res.Error = fmt.Sprintf("build config: %v", err)
		return res
	}

	tmpFile, err := os.CreateTemp("", "xray-test-*.json")
	if err != nil {
		res.Error = fmt.Sprintf("create temp file: %v", err)
		return res
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(configJSON); err != nil {
		tmpFile.Close()
		res.Error = fmt.Sprintf("write config: %v", err)
		return res
	}
	tmpFile.Close()

	var instance *xcore.Instance
	err = withSuppressedXrayOutput(func() error {
		tmpFile2, err := os.Open(tmpFile.Name())
		if err != nil {
			return fmt.Errorf("reopen config: %w", err)
		}
		defer tmpFile2.Close()

		jsonConfig, err := serial.DecodeJSONConfig(tmpFile2)
		if err != nil {
			return fmt.Errorf("decode json config: %w", err)
		}

		pbConfig, err := jsonConfig.Build()
		if err != nil {
			return fmt.Errorf("build config: %w", err)
		}

		instance, err = xcore.New(pbConfig)
		if err != nil {
			return fmt.Errorf("create instance: %w", err)
		}

		if err := instance.Start(); err != nil {
			instance.Close()
			instance = nil
			return fmt.Errorf("start xray: %w", err)
		}
		return nil
	})
	if err != nil {
		res.Error = err.Error()
		return res
	}
	defer func() {
		_ = withSuppressedXrayOutput(func() error {
			if instance != nil {
				instance.Close()
			}
			return nil
		})
	}()

	if !waitForPort(socksPort, 5*time.Second) {
		res.Error = "socks port not ready after 5s"
		return res
	}

	// socks5h resolves hostnames through the proxy — required on cellular where
	// local DNS is blocked but the VLESS tunnel still works.
	proxyURL := fmt.Sprintf("socks5h://127.0.0.1:%d", socksPort)

	connectTimeout := timeout
	if connectTimeout > 18*time.Second {
		connectTimeout = 18 * time.Second
	}
	testCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	// Step 1: connectivity check and true TTFB latency.
	traceOk, latency, traceErr := proxyConnectivityCheck(testCtx, proxyURL, cfg)
	res.Latency = latency
	if !traceOk {
		res.Error = fmt.Sprintf("connectivity: %v", traceErr)
		return res
	}

	// Step 2: best-effort download speed (does not affect Success).
	speedCtx, speedCancel := context.WithTimeout(ctx, speedBudget(timeout, latency))
	defer speedCancel()
	bytesRecv, throughput := measureProxySpeed(speedCtx, proxyURL, cfg)
	res.BytesRecv = bytesRecv
	res.Throughput = throughput
	res.Success = true

	// Step 3: optional upload speed test.
	if cfg.UploadTest {
		uploadCtx, uploadCancel := context.WithTimeout(ctx, speedBudget(timeout, latency))
		defer uploadCancel()
		uploadSent, uploadThroughput := measureProxyUploadSpeed(uploadCtx, proxyURL, cfg)
		res.UploadBytesSent = uploadSent
		res.UploadThroughput = uploadThroughput
	}

	return res
}

type traceTarget struct {
	url  string
	host string // HTTP Host / authority when url dials the CF IP directly
}

// traceTargetsForConfig builds connectivity probe URLs. The IP-based target
// matches Phase 1 (no DNS lookup) and is tried first — critical on cellular
// where UDP DNS to 1.1.1.1 is often blocked but the VLESS tunnel works.
func traceTargetsForConfig(cfg *VLESSConfig) []traceTarget {
	var targets []traceTarget
	seen := make(map[string]struct{})
	add := func(url, host string) {
		if url == "" {
			return
		}
		key := url + "|" + host
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		targets = append(targets, traceTarget{url: url, host: host})
	}

	if cfg != nil && cfg.Address != "" {
		host := cfg.Host
		if host == "" {
			host = cfg.SNI
		}
		if host != "" {
			port := cfg.Port
			if port <= 0 {
				port = 443
			}
			scheme := "https"
			if port == 80 {
				scheme = "http"
			}
			add(fmt.Sprintf("%s://%s/cdn-cgi/trace", scheme, net.JoinHostPort(cfg.Address, strconv.Itoa(port))), host)
			add(fmt.Sprintf("%s://%s:%d/cdn-cgi/trace", scheme, host, port), "")
		}
	}

	for _, u := range traceProbeURLs {
		add(u, "")
	}
	return targets
}

func tunnelPathTargets(cfg *VLESSConfig) []traceTarget {
	if cfg == nil || cfg.Address == "" {
		return nil
	}
	host := cfg.Host
	if host == "" {
		host = cfg.SNI
	}
	if host == "" {
		return nil
	}
	path := cfg.Path
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	port := cfg.Port
	if port <= 0 {
		port = 443
	}
	scheme := "https"
	if port == 80 {
		scheme = "http"
	}
	return []traceTarget{{
		url:  fmt.Sprintf("%s://%s%s", scheme, net.JoinHostPort(cfg.Address, strconv.Itoa(port)), path),
		host: host,
	}}
}

func proxyTunnelPathCheck(ctx context.Context, proxyAddr string, cfg *VLESSConfig) (bool, time.Duration, error) {
	for _, target := range tunnelPathTargets(cfg) {
		ok, latency, err := proxyRelaxedEndpointCheck(ctx, proxyAddr, target.url, target.host, 1)
		if ok {
			return true, latency, nil
		}
		if err != nil {
			return false, latency, err
		}
	}
	return false, 0, fmt.Errorf("tunnel path unreachable")
}

func proxyRelaxedEndpointCheck(ctx context.Context, proxyAddr, targetURL, authority string, minBytes int64) (bool, time.Duration, error) {
	start := time.Now()
	var latency time.Duration
	gotFirst := false
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			if !gotFirst {
				latency = time.Since(start)
				gotFirst = true
			}
		},
	}
	traceCtx := httptrace.WithClientTrace(ctx, trace)

	client := &http.Client{
		Transport: proxyTransportForTarget(proxyAddr, targetURL, authority),
		Timeout:   clientTimeoutForContext(ctx, 20*time.Second),
	}
	req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return false, 0, err
	}
	req.Header.Set("User-Agent", "magik/1.0")
	if authority != "" {
		req.Host = authority
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, latency, err
	}
	n, _ := io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	status := resp.StatusCode
	resp.Body.Close()
	if status >= 500 {
		return false, latency, fmt.Errorf("HTTP %d", status)
	}
	if n < minBytes {
		return false, latency, fmt.Errorf("short response (%d bytes)", n)
	}
	latency = ensurePositiveLatency(start, latency, gotFirst)
	return true, latency, nil
}

// ProxyConnectivityCheck performs a lightweight connectivity test through the
// SOCKS5 proxy. Exported for the Android gomobile bridge.
func ProxyConnectivityCheck(ctx context.Context, proxyAddr string, cfg *VLESSConfig) (bool, time.Duration, error) {
	return proxyConnectivityCheck(ctx, proxyAddr, cfg)
}

// proxyConnectivityCheck performs a lightweight GET /cdn-cgi/trace through the
// SOCKS5 proxy. It returns true when the response body contains "colo=",
// proving that real Cloudflare traffic flowed through the proxy.
func proxyConnectivityCheck(ctx context.Context, proxyAddr string, cfg *VLESSConfig) (bool, time.Duration, error) {
	// Domain-first: SOCKS traffic uses natural TLS SNI and the worker forwards it.
	if cfg != nil {
		host := cfg.Host
		if host == "" {
			host = cfg.SNI
		}
		if host != "" {
			domainTrace := "https://" + host + "/cdn-cgi/trace"
			if ok, latency, err := proxyConnectivityCheckTarget(ctx, proxyAddr, domainTrace, ""); ok {
				return true, latency, nil
			} else if err != nil {
				_ = err
			}
			if cfg.Path != "" {
				path := cfg.Path
				if !strings.HasPrefix(path, "/") {
					path = "/" + path
				}
				domainPath := "https://" + host + path
				if ok, latency, err := proxyRelaxedEndpointCheck(ctx, proxyAddr, domainPath, "", 1); ok {
					return true, latency, nil
				} else if err != nil {
					_ = err
				}
			}
		}
	}

	// Then hit the WS path through the CF IP (TLS SNI overridden in transport).
	if cfg != nil {
		if ok, latency, err := proxyTunnelPathCheck(ctx, proxyAddr, cfg); ok {
			return true, latency, nil
		} else if err != nil {
			_ = err
		}
	}

	targets := traceTargetsForConfig(cfg)
	ok, latency, err := proxyConnectivityCheckTargets(ctx, proxyAddr, targets)
	if ok {
		return true, latency, nil
	}

	// Fallback: a small download through the config host/path proves the tunnel
	// carries data even when trace endpoints are filtered on cellular links.
	if cfg != nil {
		if ok, dlLatency, dlErr := proxyDataPathCheck(ctx, proxyAddr, cfg); ok {
			if dlLatency > 0 {
				return true, dlLatency, nil
			}
			return true, latency, nil
		} else if dlErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; data-path: %v", err, dlErr)
			} else {
				err = dlErr
			}
		}
	}

	return false, latency, err
}

func proxyConnectivityCheckTargets(ctx context.Context, proxyAddr string, targets []traceTarget) (bool, time.Duration, error) {
	if len(targets) == 0 {
		return false, 0, fmt.Errorf("no trace probe targets configured")
	}

	var failures []string
	var lastLatency time.Duration
	for _, target := range targets {
		ok, latency, err := proxyConnectivityCheckTarget(ctx, proxyAddr, target.url, target.host)
		if ok {
			return true, latency, nil
		}
		if latency > 0 {
			lastLatency = latency
		}
		if err != nil {
			label := target.url
			if target.host != "" {
				label = fmt.Sprintf("%s (host=%s)", target.url, target.host)
			}
			failures = append(failures, fmt.Sprintf("%s: %v", label, err))
		}
		if ctx.Err() != nil {
			return false, lastLatency, ctx.Err()
		}
	}

	errMsg := fmt.Errorf("trace probe failed: %s", strings.Join(failures, "; "))
	return false, lastLatency, errMsg
}

func proxyConnectivityCheckTarget(ctx context.Context, proxyAddr, target, authority string) (bool, time.Duration, error) {
	start := time.Now()
	var latency time.Duration
	gotFirst := false
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			if !gotFirst {
				latency = time.Since(start)
				gotFirst = true
			}
		},
	}
	traceCtx := httptrace.WithClientTrace(ctx, trace)

	client := &http.Client{
		Transport: proxyTransportForTarget(proxyAddr, target, authority),
		Timeout:   clientTimeoutForContext(ctx, 20*time.Second),
	}

	req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, target, nil)
	if err != nil {
		return false, 0, err
	}
	req.Header.Set("User-Agent", "magik/1.0")
	if authority != "" {
		req.Host = authority
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, latency, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return false, latency, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if !strings.Contains(string(body), "colo=") {
		return false, latency, fmt.Errorf("no colo in trace response")
	}
	latency = ensurePositiveLatency(start, latency, gotFirst)
	return true, latency, nil
}

func ensurePositiveLatency(start time.Time, latency time.Duration, gotFirst bool) time.Duration {
	if !gotFirst || latency <= 0 {
		latency = time.Since(start)
	}
	if latency <= 0 {
		return time.Nanosecond
	}
	return latency
}

func proxyDataPathCheck(ctx context.Context, proxyAddr string, cfg *VLESSConfig) (bool, time.Duration, error) {
	const sample = speedMinBytes * 2
	for _, target := range tunnelPathTargets(cfg) {
		ok, latency, _ := proxyRelaxedEndpointCheck(ctx, proxyAddr, target.url, target.host, speedMinBytes)
		if ok {
			return true, latency, nil
		}
	}
	for _, target := range speedTestTargets(cfg, sample) {
		ok, latency, _ := proxyRelaxedEndpointCheck(ctx, proxyAddr, target.url, target.host, target.minBytes)
		if ok {
			return true, latency, nil
		}
	}
	return false, 0, fmt.Errorf("no data-path response")
}

type speedTarget struct {
	url      string
	host     string // HTTP Host when url dials a CF IP directly
	relaxed  bool
	minBytes int64
}

func speedBudget(total, spent time.Duration) time.Duration {
	remaining := total - spent
	if remaining < time.Second {
		return time.Second
	}
	return remaining
}

func measureProxySpeed(ctx context.Context, proxyAddr string, cfg *VLESSConfig) (int64, float64) {
	samples := []int64{int64(speedSampleBytes), int64(speedSampleBytesFast)}
	if cfg != nil && cfg.SpeedSize > 0 {
		samples = []int64{cfg.SpeedSize}
	}
	for _, sample := range samples {
		for _, target := range speedTestTargets(cfg, sample) {
			bytesRecv, throughput, err := downloadThroughProxy(ctx, proxyAddr, target.url, sample, target.relaxed, target.host)
			if err == nil && bytesRecv >= target.minBytes && throughput > 0 {
				return bytesRecv, throughput
			}
		}
	}

	burstBytes := int64(speedSampleBytesFast)
	if cfg != nil && cfg.SpeedSize > 0 {
		burstBytes = cfg.SpeedSize
	}
	// WS/xhttp tunnels often block speed.cloudflare.com but still carry trace traffic.
	// Estimate throughput by saturating the known-good trace endpoint in parallel.
	return burstProxyThroughput(ctx, proxyAddr, traceProbeURL, burstBytes)
}

// measureProxyUploadSpeed measures upload throughput through a SOCKS5 proxy by
// sending a POST request with a synthetic body to an upload-capable endpoint.
func measureProxyUploadSpeed(ctx context.Context, proxyAddr string, cfg *VLESSConfig) (int64, float64) {
	sampleBytes := int64(speedSampleBytesFast)
	if cfg != nil && cfg.SpeedSize > 0 {
		sampleBytes = cfg.SpeedSize
	}
	if sampleBytes < speedMinBytes {
		sampleBytes = speedMinBytes
	}

	// Upload targets — POST to cloudflare trace (always accepts body) or the
	// config host. The upload goes through the xray SOCKS proxy so it measures
	// real upstream capacity.
	var targets []speedTarget
	if cfg != nil {
		host := cfg.Host
		if host == "" {
			host = cfg.SNI
		}
		port := cfg.Port
		if port <= 0 {
			port = 443
		}
		scheme := "https"
		if port == 80 {
			scheme = "http"
		}
		if host != "" && cfg.Address != "" {
			targets = append(targets, speedTarget{
				url:      fmt.Sprintf("%s://%s/cdn-cgi/trace", scheme, net.JoinHostPort(cfg.Address, strconv.Itoa(port))),
				host:     host,
				relaxed:  true,
				minBytes: 1,
			})
		}
	}
	targets = append(targets, speedTarget{
		url:      traceProbeURL,
		relaxed:  true,
		minBytes: 1,
	})

	for _, target := range targets {
		sent, throughput, err := uploadThroughProxy(ctx, proxyAddr, target.url, sampleBytes, target.relaxed, target.host)
		if err == nil && sent > 0 && throughput > 0 {
			return sent, throughput
		}
	}

	return 0, 0
}

// uploadThroughProxy sends a POST request with a synthetic body through a SOCKS5
// proxy and returns bytes sent plus throughput in bytes/sec.
func uploadThroughProxy(ctx context.Context, proxyAddr, uploadURL string, maxBytes int64, relaxed bool, authority string) (int64, float64, error) {
	if maxBytes <= 0 {
		return 0, 0, fmt.Errorf("invalid maxBytes %d", maxBytes)
	}

	client := &http.Client{
		Transport: proxyTransportForTarget(proxyAddr, uploadURL, authority),
		Timeout:   clientTimeoutForContext(ctx, 30*time.Second),
	}

	body := io.LimitReader(randReader(), maxBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return 0, 0, err
	}
	req.ContentLength = maxBytes
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("User-Agent", "magik/1.0")
	if authority != "" {
		req.Host = authority
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	if !relaxed && (resp.StatusCode < 200 || resp.StatusCode >= 400) {
		return 0, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if relaxed && resp.StatusCode >= 500 {
		return 0, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	elapsed := time.Since(start).Seconds()
	if elapsed <= 0 {
		return maxBytes, 0, nil
	}
	return maxBytes, float64(maxBytes) / elapsed, nil
}

// randReader returns an io.Reader that yields random bytes.
var randReader = func() io.Reader {
	return &zeroReader{}
}

type zeroReader struct{}

func (z *zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func speedTestTargets(cfg *VLESSConfig, sampleBytes int64) []speedTarget {
	minBytes := int64(speedMinBytes)
	if sampleBytes < minBytes {
		minBytes = sampleBytes / 2
	}
	if minBytes < 4096 {
		minBytes = 4096
	}

	var targets []speedTarget
	add := func(rawURL string, relaxed bool) {
		if rawURL == "" {
			return
		}
		targets = append(targets, speedTarget{
			url:      rawURL,
			relaxed:  relaxed,
			minBytes: minBytes,
		})
	}

	if cfg != nil && cfg.SpeedURL != "" {
		add(cfg.SpeedURL, true)
	}

	if cfg != nil {
		host := cfg.Host
		if host == "" {
			host = cfg.SNI
		}
		port := cfg.Port
		if port <= 0 {
			port = 443
		}
		scheme := "https"
		if port == 80 {
			scheme = "http"
		}
		if host != "" {
			paths := []string{"/"}
			if cfg.Path != "" {
				paths = append([]string{cfg.Path}, paths...)
			}
			seen := make(map[string]struct{})
			for _, path := range paths {
				if !strings.HasPrefix(path, "/") {
					path = "/" + path
				}
				if cfg.Address != "" {
					ipURL := fmt.Sprintf("%s://%s%s", scheme, net.JoinHostPort(cfg.Address, strconv.Itoa(port)), path)
					if _, ok := seen[ipURL]; !ok {
						seen[ipURL] = struct{}{}
						targets = append(targets, speedTarget{
							url: ipURL, host: host, relaxed: true, minBytes: minBytes,
						})
					}
				}
				u := "https://" + host + path
				if _, ok := seen[u]; ok {
					continue
				}
				seen[u] = struct{}{}
				add(u, true)
			}
		}
	}

	add(fmt.Sprintf("https://speed.cloudflare.com/__down?bytes=%d", sampleBytes), false)
	add("https://www.cloudflare.com/", true)
	return targets
}

func burstProxyThroughput(ctx context.Context, proxyAddr, url string, targetBytes int64) (int64, float64) {
	if targetBytes <= 0 {
		return 0, 0
	}

	start := time.Now()
	var total int64
	const workers = 8

	for total < targetBytes && ctx.Err() == nil {
		var wg sync.WaitGroup
		var batch int64
		var mu sync.Mutex

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				n, _, err := downloadThroughProxy(ctx, proxyAddr, url, 16384, true, "")
				if err != nil || n <= 0 {
					return
				}
				mu.Lock()
				batch += n
				mu.Unlock()
			}()
		}
		wg.Wait()
		if batch == 0 {
			break
		}
		total += batch
	}

	elapsed := time.Since(start).Seconds()
	if total < 4096 || elapsed <= 0 {
		return total, 0
	}
	return total, float64(total) / elapsed
}

func proxyTransport(proxyAddr string) *http.Transport {
	return proxyTransportForTarget(proxyAddr, "", "")
}

// proxyTransportForTarget builds a SOCKS transport. When the probe URL dials a
// literal IP but the HTTP authority is a domain (typical CF IP scans), TLS must
// use the domain as ServerName — req.Host alone does not fix the ClientHello.
func proxyTransportForTarget(proxyAddr, targetURL, authority string) *http.Transport {
	t := &http.Transport{
		DialContext:         (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
	}
	if proxyAddr != "" {
		t.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxyAddr)
		}
	}
	if authority == "" || targetURL == "" {
		return t
	}
	u, err := url.Parse(targetURL)
	if err != nil || u.Scheme != "https" {
		return t
	}
	if net.ParseIP(u.Hostname()) == nil {
		return t
	}
	serverName := authority
	if h, _, err := net.SplitHostPort(authority); err == nil {
		serverName = h
	}
	t.TLSClientConfig = &tls.Config{ServerName: serverName}
	return t
}

func clientTimeoutForContext(ctx context.Context, fallback time.Duration) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fallback
	}
	if remaining := time.Until(deadline); remaining > 0 {
		return remaining
	}
	return fallback
}

// downloadThroughProxy fetches a URL through a SOCKS5 proxy and returns bytes
// received plus throughput in bytes/sec. When relaxed is true, any HTTP response
// with a readable body counts (needed for WS endpoints that answer 400/404).
func downloadThroughProxy(ctx context.Context, proxyAddr, dlURL string, maxBytes int64, relaxed bool, authority string) (int64, float64, error) {
	if maxBytes <= 0 {
		return 0, 0, fmt.Errorf("invalid maxBytes %d", maxBytes)
	}

	client := &http.Client{
		Transport: proxyTransportForTarget(proxyAddr, dlURL, authority),
		Timeout:   clientTimeoutForContext(ctx, 30*time.Second),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dlURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", "magik/1.0")
	if authority != "" {
		req.Host = authority
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if !relaxed && (resp.StatusCode < 200 || resp.StatusCode >= 400) {
		return 0, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if relaxed && resp.StatusCode >= 500 {
		return 0, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	n, err := io.Copy(io.Discard, io.LimitReader(resp.Body, maxBytes))
	elapsed := time.Since(start).Seconds()
	if err != nil || n <= 0 || elapsed <= 0 {
		return n, 0, err
	}
	return n, float64(n) / elapsed, nil
}

// waitForPort waits until a TCP port is accepting connections.
func waitForPort(port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func withSuppressedXrayOutput(fn func() error) error {
	restore := suppressXrayOutput()
	defer restore()
	return fn()
}

func suppressXrayOutput() func() {
	stdioMu.Lock()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		stdioMu.Unlock()
		return func() {}
	}

	os.Stdout = devNull
	os.Stderr = devNull

	return func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		devNull.Close()
		stdioMu.Unlock()
	}
}
