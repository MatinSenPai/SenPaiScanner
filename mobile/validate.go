package mobile

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/xraytest"

	xcore "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf/serial"
)

// ---------------------------------------------------------------------------
// Mobile-safe xray validation
//
// The shared xraytest.ValidateConfig in runner.go has three issues on Android:
//
//  1. It globally redirects os.Stdout/os.Stderr to /dev/null to suppress xray
//     output. On Android the gomobile/JNI bridge uses these FDs for cross-
//     language communication — swapping them causes deadlocks.
//
//  2. If xcore.New() fails (line 134-138 in runner.go), os.Stdout/os.Stderr
//     are NEVER restored and the devNull file descriptor leaks. After several
//     failed IPs the process runs out of FDs.
//
//  3. It writes a temp file and re-reads it, which can fail on Android
//     depending on the temp directory setup.
//
// This file reimplements the validation loop specifically for mobile, avoiding
// all three issues while keeping the same test logic (connectivity check via
// cloudflare trace + best-effort speed measurement).
// ---------------------------------------------------------------------------

var (
	mobilePortCounter atomic.Int32
	// mobileInstanceMu serializes xray instance creation/teardown so that
	// we never have two instances competing for ports and FDs on Android.
	mobileInstanceMu sync.Mutex
)

func init() {
	mobilePortCounter.Store(30000) // different range than desktop to avoid collisions
}

func mobileNextPort() int {
	return int(mobilePortCounter.Add(1))
}

var mobileTraceURLs = []string{
	"https://cp.cloudflare.com/cdn-cgi/trace",
	"https://cloudflare.com/cdn-cgi/trace",
}

// mobileValidateConfig is the Android-safe replacement for xraytest.ValidateConfig.
// It starts an xray instance, sends test traffic through it, and returns
// the result. Retries once on failure.
func mobileValidateConfig(ctx context.Context, cfg *xraytest.VLESSConfig, timeout time.Duration) *xraytest.ValidationResult {
	res := mobileValidateOnce(ctx, cfg, timeout)
	if !res.Success {
		time.Sleep(500 * time.Millisecond)
		res2 := mobileValidateOnce(ctx, cfg, timeout)
		res2.Retries = 1
		if res2.Success {
			return res2
		}
		res.Retries = 1
	}
	return res
}

func mobileValidateOnce(ctx context.Context, cfg *xraytest.VLESSConfig, timeout time.Duration) *xraytest.ValidationResult {
	res := &xraytest.ValidationResult{
		IP:        cfg.Address,
		Port:      cfg.Port,
		Transport: cfg.Network,
	}

	socksPort := mobileNextPort()

	configJSON, err := xraytest.BuildXrayConfig(cfg, socksPort)
	if err != nil {
		res.Error = fmt.Sprintf("build config: %v", err)
		return res
	}

	// Parse JSON config directly from bytes — no temp file needed.
	jsonConfig, err := serial.DecodeJSONConfig(bytes.NewReader(configJSON))
	if err != nil {
		res.Error = fmt.Sprintf("decode json config: %v", err)
		return res
	}

	pbConfig, err := jsonConfig.Build()
	if err != nil {
		res.Error = fmt.Sprintf("build pb config: %v", err)
		return res
	}

	// Serialize instance lifecycle to prevent resource contention on Android.
	mobileInstanceMu.Lock()

	instance, err := xcore.New(pbConfig)
	if err != nil {
		mobileInstanceMu.Unlock()
		res.Error = fmt.Sprintf("create instance: %v", err)
		return res
	}

	if err := instance.Start(); err != nil {
		instance.Close()
		mobileInstanceMu.Unlock()
		time.Sleep(100 * time.Millisecond)
		res.Error = fmt.Sprintf("start xray: %v", err)
		return res
	}

	// Instance is running and SOCKS port is bound — release the lock.
	mobileInstanceMu.Unlock()

	// NOTE: We intentionally do NOT redirect os.Stdout/os.Stderr.
	// The xray config already sets loglevel to "none" (see builder.go).
	// On Android, touching these global FDs deadlocks the JNI bridge.

	// Ensure cleanup + short delay for the OS to release the port.
	defer func() {
		instance.Close()
		time.Sleep(150 * time.Millisecond)
	}()

	if !mobileWaitForPort(socksPort, 5*time.Second) {
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

	// Step 1: connectivity check via cloudflare trace.
	traceOk, latency, traceErr := mobileConnectivityCheck(testCtx, proxyURL, cfg)
	res.Latency = latency
	if !traceOk {
		res.Error = fmt.Sprintf("connectivity: %v", traceErr)
		return res
	}

	// Step 2: best-effort speed measurement (does not affect Success).
	speedCtx, speedCancel := context.WithTimeout(ctx, mobileSpeedBudget(timeout, latency))
	defer speedCancel()
	bytesRecv, throughput := mobileSpeedTest(speedCtx, proxyURL, cfg)
	res.BytesRecv = bytesRecv
	res.Throughput = throughput
	res.Success = true

	// Step 3: optional upload speed test.
	if cfg.UploadTest {
		uploadCtx, uploadCancel := context.WithTimeout(ctx, mobileSpeedBudget(timeout, latency))
		defer uploadCancel()
		uploadSent, uploadThroughput := mobileUploadSpeedTest(uploadCtx, proxyURL, cfg)
		res.UploadBytesSent = uploadSent
		res.UploadThroughput = uploadThroughput
	}

	return res
}

// ---------------------------------------------------------------------------
// Helper functions (re-implemented to avoid depending on unexported xraytest)
// ---------------------------------------------------------------------------

func mobileWaitForPort(port int, timeout time.Duration) bool {
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

func mobileProxyTransport(proxyAddr string) *http.Transport {
	return mobileProxyTransportForTarget(proxyAddr, "", "")
}

// mobileProxyTransportForTarget builds a SOCKS transport. When the probe URL
// dials a literal IP but the HTTP authority is a domain (typical CF IP scans),
// TLS must use the domain as ServerName — req.Host alone does not fix the
// ClientHello.
func mobileProxyTransportForTarget(proxyAddr, targetURL, authority string) *http.Transport {
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

func mobileClientTimeout(ctx context.Context, fallback time.Duration) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return fallback
	}
	if remaining := time.Until(deadline); remaining > 0 {
		return remaining
	}
	return fallback
}

// ---------------------------------------------------------------------------
// Connectivity check
// ---------------------------------------------------------------------------

type mobileTraceTarget struct {
	url  string
	host string // HTTP Host / authority when url dials the CF IP directly
}

// mobileTraceTargetsForConfig builds connectivity probe URLs. The IP-based
// target matches Phase 1 (no DNS lookup) and is tried first — critical on
// cellular where UDP DNS to 1.1.1.1 is often blocked but the VLESS tunnel works.
func mobileTraceTargetsForConfig(cfg *xraytest.VLESSConfig) []mobileTraceTarget {
	var targets []mobileTraceTarget
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
		targets = append(targets, mobileTraceTarget{url: url, host: host})
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
			add(fmt.Sprintf("%s://%s:%d/cdn-cgi/trace", scheme, cfg.Address, port), host)
			add(fmt.Sprintf("%s://%s:%d/cdn-cgi/trace", scheme, host, port), "")
		}
	}

	for _, u := range mobileTraceURLs {
		add(u, "")
	}
	return targets
}

func mobileTunnelPathTargets(cfg *xraytest.VLESSConfig) []mobileTraceTarget {
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
	return []mobileTraceTarget{{
		url:  fmt.Sprintf("%s://%s:%d%s", scheme, cfg.Address, port, path),
		host: host,
	}}
}

func mobileConnectivityCheck(ctx context.Context, proxyAddr string, cfg *xraytest.VLESSConfig) (bool, time.Duration, error) {
	// Domain-first: SOCKS traffic uses natural TLS SNI and the worker forwards it.
	if cfg != nil {
		host := cfg.Host
		if host == "" {
			host = cfg.SNI
		}
		if host != "" {
			domainTrace := "https://" + host + "/cdn-cgi/trace"
			if ok, latency, err := mobileConnectivityCheckTarget(ctx, proxyAddr, domainTrace, ""); ok {
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
				if ok, latency, err := mobileProxyRelaxedEndpointCheck(ctx, proxyAddr, domainPath, "", 1); ok {
					return true, latency, nil
				} else if err != nil {
					_ = err
				}
			}
		}
	}

	// Then hit the WS path through the CF IP (TLS SNI overridden in transport).
	if cfg != nil {
		if ok, latency, err := mobileProxyTunnelPathCheck(ctx, proxyAddr, cfg); ok {
			return true, latency, nil
		} else if err != nil {
			_ = err
		}
	}

	targets := mobileTraceTargetsForConfig(cfg)
	ok, latency, err := mobileConnectivityCheckTargets(ctx, proxyAddr, targets)
	if ok {
		return true, latency, nil
	}

	// Fallback: a small download through the config host/path proves the tunnel
	// carries data even when trace endpoints are filtered on cellular links.
	if cfg != nil {
		if ok, dlLatency, dlErr := mobileProxyDataPathCheck(ctx, proxyAddr, cfg); ok {
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

func mobileConnectivityCheckTargets(ctx context.Context, proxyAddr string, targets []mobileTraceTarget) (bool, time.Duration, error) {
	if len(targets) == 0 {
		return false, 0, fmt.Errorf("no trace probe targets configured")
	}

	var failures []string
	var lastLatency time.Duration
	for _, target := range targets {
		ok, latency, err := mobileConnectivityCheckTarget(ctx, proxyAddr, target.url, target.host)
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

	return false, lastLatency, fmt.Errorf("trace probe failed: %s", strings.Join(failures, "; "))
}

func mobileConnectivityCheckTarget(ctx context.Context, proxyAddr, target, authority string) (bool, time.Duration, error) {
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
		Transport: mobileProxyTransportForTarget(proxyAddr, target, authority),
		Timeout:   mobileClientTimeout(ctx, 20*time.Second),
	}

	req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, target, nil)
	if err != nil {
		return false, 0, err
	}
	req.Header.Set("User-Agent", "senpaiscanner/1.0")
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
	if !gotFirst {
		latency = time.Since(start)
	}
	return true, latency, nil
}

func mobileProxyTunnelPathCheck(ctx context.Context, proxyAddr string, cfg *xraytest.VLESSConfig) (bool, time.Duration, error) {
	for _, target := range mobileTunnelPathTargets(cfg) {
		ok, latency, err := mobileProxyRelaxedEndpointCheck(ctx, proxyAddr, target.url, target.host, 1)
		if ok {
			return true, latency, nil
		}
		if err != nil {
			return false, latency, err
		}
	}
	return false, 0, fmt.Errorf("tunnel path unreachable")
}

func mobileProxyRelaxedEndpointCheck(ctx context.Context, proxyAddr, targetURL, authority string, minBytes int64) (bool, time.Duration, error) {
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
		Transport: mobileProxyTransportForTarget(proxyAddr, targetURL, authority),
		Timeout:   mobileClientTimeout(ctx, 20*time.Second),
	}
	req, err := http.NewRequestWithContext(traceCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		return false, 0, err
	}
	req.Header.Set("User-Agent", "senpaiscanner/1.0")
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
	if !gotFirst {
		latency = time.Since(start)
	}
	return true, latency, nil
}

func mobileProxyDataPathCheck(ctx context.Context, proxyAddr string, cfg *xraytest.VLESSConfig) (bool, time.Duration, error) {
	const sample int64 = 16 * 1024
	for _, target := range mobileTunnelPathTargets(cfg) {
		ok, latency, _ := mobileProxyRelaxedEndpointCheck(ctx, proxyAddr, target.url, target.host, 8192)
		if ok {
			return true, latency, nil
		}
	}
	for _, target := range mobileSpeedTestTargets(cfg, sample) {
		ok, latency, _ := mobileProxyRelaxedEndpointCheck(ctx, proxyAddr, target.url, target.host, target.minBytes)
		if ok {
			return true, latency, nil
		}
	}
	return false, 0, fmt.Errorf("no data-path response")
}

// ---------------------------------------------------------------------------
// Speed measurement
// ---------------------------------------------------------------------------

type mobileSpeedTestTarget struct {
	url      string
	host     string // HTTP Host when url dials a CF IP directly
	relaxed  bool
	minBytes int64
}

func mobileSpeedBudget(total, spent time.Duration) time.Duration {
	budget := 4 * time.Second
	remaining := total - spent
	if remaining < budget {
		budget = remaining
	}
	if budget < time.Second {
		return time.Second
	}
	return budget
}

func mobileDownload(ctx context.Context, proxyAddr, dlURL string, maxBytes int64, relaxed bool, authority string) (int64, float64, error) {
	if maxBytes <= 0 {
		return 0, 0, fmt.Errorf("invalid maxBytes %d", maxBytes)
	}
	client := &http.Client{
		Transport: mobileProxyTransportForTarget(proxyAddr, dlURL, authority),
		Timeout:   mobileClientTimeout(ctx, 30*time.Second),
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dlURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", "senpaiscanner/1.0")
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

func mobileSpeedTestTargets(cfg *xraytest.VLESSConfig, sampleBytes int64) []mobileSpeedTestTarget {
	minBytes := int64(8 * 1024)
	if sampleBytes < minBytes {
		minBytes = sampleBytes / 2
	}
	if minBytes < 4096 {
		minBytes = 4096
	}

	var targets []mobileSpeedTestTarget
	add := func(url, host string, relaxed bool) {
		if url == "" {
			return
		}
		targets = append(targets, mobileSpeedTestTarget{
			url:      url,
			host:     host,
			relaxed:  relaxed,
			minBytes: minBytes,
		})
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
					ipURL := fmt.Sprintf("%s://%s:%d%s", scheme, cfg.Address, port, path)
					if _, ok := seen[ipURL]; !ok {
						seen[ipURL] = struct{}{}
						targets = append(targets, mobileSpeedTestTarget{
							url: ipURL, host: host, relaxed: true, minBytes: minBytes,
						})
					}
				}
				u := "https://" + host + path
				if _, ok := seen[u]; ok {
					continue
				}
				seen[u] = struct{}{}
				add(u, "", true)
			}
		}
	}

	if cfg != nil && cfg.SpeedURL != "" {
		add(cfg.SpeedURL, "", true)
	}

	add(fmt.Sprintf("https://speed.cloudflare.com/__down?bytes=%d", sampleBytes), "", false)
	add("https://www.cloudflare.com/", "", true)
	return targets
}

func mobileSpeedTest(ctx context.Context, proxyAddr string, cfg *xraytest.VLESSConfig) (int64, float64) {
	sampleBytes := int64(128 * 1024)
	if cfg != nil && cfg.SpeedSize > 0 {
		sampleBytes = cfg.SpeedSize
	}
	minBytes := int64(8 * 1024)
	if sampleBytes < minBytes {
		minBytes = sampleBytes / 2
	}
	if minBytes < 4096 {
		minBytes = 4096
	}

	for _, target := range mobileSpeedTestTargets(cfg, sampleBytes) {
		bytesRecv, throughput, err := mobileDownload(ctx, proxyAddr, target.url, sampleBytes, target.relaxed, target.host)
		if err == nil && bytesRecv >= minBytes && throughput > 0 {
			return bytesRecv, throughput
		}
	}

	// Last fallback: burst trace requests.
	return mobileBurstThroughput(ctx, proxyAddr, mobileTraceURLs[0], sampleBytes)
}

func mobileBurstThroughput(ctx context.Context, proxyAddr, targetURL string, targetBytes int64) (int64, float64) {
	if targetBytes <= 0 {
		return 0, 0
	}
	start := time.Now()
	var total int64
	const workers = 4 // fewer workers on mobile to conserve resources

	for total < targetBytes && ctx.Err() == nil {
		var wg sync.WaitGroup
		var batch int64
		var bmu sync.Mutex

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				n, _, err := mobileDownload(ctx, proxyAddr, targetURL, 16384, true, "")
				if err != nil || n <= 0 {
					return
				}
				bmu.Lock()
				batch += n
				bmu.Unlock()
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

func mobileUploadSpeedTest(ctx context.Context, proxyAddr string, cfg *xraytest.VLESSConfig) (int64, float64) {
	sampleBytes := int64(128 * 1024) // speedSampleBytesFast equivalent
	if cfg != nil && cfg.SpeedSize > 0 {
		sampleBytes = cfg.SpeedSize
	}
	const speedMinBytes = int64(8 * 1024)
	if sampleBytes < speedMinBytes {
		sampleBytes = speedMinBytes
	}

	var targets []mobileSpeedTestTarget
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
			targets = append(targets, mobileSpeedTestTarget{
				url:      fmt.Sprintf("%s://%s/cdn-cgi/trace", scheme, net.JoinHostPort(cfg.Address, strconv.Itoa(port))),
				host:     host,
				relaxed:  true,
				minBytes: 1,
			})
		}
	}
	targets = append(targets, mobileSpeedTestTarget{
		url:      "https://cp.cloudflare.com/cdn-cgi/trace",
		relaxed:  true,
		minBytes: 1,
	})

	for _, target := range targets {
		sent, throughput, err := mobileUpload(ctx, proxyAddr, target.url, sampleBytes, target.relaxed, target.host)
		if err == nil && sent > 0 && throughput > 0 {
			return sent, throughput
		}
	}

	return 0, 0
}

func mobileUpload(ctx context.Context, proxyAddr, uploadURL string, maxBytes int64, relaxed bool, authority string) (int64, float64, error) {
	if maxBytes <= 0 {
		return 0, 0, fmt.Errorf("invalid maxBytes %d", maxBytes)
	}

	client := &http.Client{
		Transport: mobileProxyTransportForTarget(proxyAddr, uploadURL, authority),
		Timeout:   mobileClientTimeout(ctx, 30*time.Second),
	}

	body := io.LimitReader(mobileRandReader(), maxBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, body)
	if err != nil {
		return 0, 0, err
	}
	req.ContentLength = maxBytes
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("User-Agent", "senpaiscanner/1.0")
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

var mobileRandReader = func() io.Reader {
	return &mobileZeroReader{}
}

type mobileZeroReader struct{}

func (z *mobileZeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
