package main

import (
	"context"
	"fmt"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/matinsenpai/senpaiscanner/internal/engine"
	"github.com/matinsenpai/senpaiscanner/internal/export"
	"github.com/matinsenpai/senpaiscanner/internal/ipsrc"
	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
	"github.com/matinsenpai/senpaiscanner/pkg/version"
)

// App is the Wails application backend.
type App struct {
	ctx        context.Context
	scanCancel context.CancelFunc

	mu         sync.Mutex
	scanning   bool
	phase2Done atomic.Bool
}

// NewApp creates the backend.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetVersion returns the build version string.
func (a *App) GetVersion() string {
	return version.String()
}

// ScanParams mirrors the CLI scan options.
type ScanParams struct {
	Count       int    `json:"count"`
	Workers     int    `json:"workers"`
	TimeoutMs   int    `json:"timeoutMs"`
	Tries       int    `json:"tries"`
	Port        int    `json:"port"`
	Mode        string `json:"mode"` // tcp | tls | http
	SNI         string `json:"sni"`
	SpeedTest   bool   `json:"speedTest"`
	RequireWS   bool   `json:"requireWS"`
	ColoFilter  string `json:"coloFilter"`
	OutputFile  string `json:"outputFile"`
}

// ScanStats is emitted periodically during a scan.
type ScanStats struct {
	Tested   int64 `json:"tested"`
	Healthy  int64 `json:"healthy"`
	Failed   int64 `json:"failed"`
	InFlight int64 `json:"inFlight"`
	Total    int   `json:"total"`
}

// ScanResult is one finished probe, sent to the UI.
type ScanResult struct {
	IP         string  `json:"ip"`
	Port       int     `json:"port"`
	Colo       string  `json:"colo"`
	AvgMs      float64 `json:"avgMs"`
	MinMs      float64 `json:"minMs"`
	Loss       float64 `json:"loss"`
	JitterMs   float64 `json:"jitterMs"`
	Throughput float64 `json:"throughput"` // bytes/sec
	Healthy    bool    `json:"healthy"`
	Mode       string  `json:"mode"`
}

func toScanResult(r *result.Result) ScanResult {
	return ScanResult{
		IP:         r.IP.String(),
		Port:       r.Port,
		Colo:       r.Colo,
		AvgMs:      float64(r.Avg()) / float64(time.Millisecond),
		MinMs:      float64(r.Min()) / float64(time.Millisecond),
		Loss:       r.Loss(),
		JitterMs:   float64(r.Jitter()) / float64(time.Millisecond),
		Throughput: r.Throughput,
		Healthy:    r.IsHealthy(),
		Mode:       r.ProbeMode,
	}
}

// StartScan launches a Phase 1 scan with the given parameters. Progress and
// results are streamed to the frontend as events:
//   - "scan:stats"  (ScanStats)
//   - "scan:result" (ScanResult)
//   - "scan:done"   (ScanStats)
func (a *App) StartScan(params ScanParams) error {
	a.mu.Lock()
	if a.scanning {
		a.mu.Unlock()
		return fmt.Errorf("a scan is already running")
	}
	a.scanning = true
	a.phase2Done.Store(false)
	a.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	a.scanCancel = cancel

	go func() {
		defer func() {
			a.mu.Lock()
			a.scanning = false
			a.mu.Unlock()
		}()
		a.runScan(ctx, params)
	}()
	return nil
}

func (a *App) runScan(ctx context.Context, params ScanParams) {
	if params.Count <= 0 {
		params.Count = 5000
	}
	if params.Workers <= 0 {
		params.Workers = 50
	}
	timeout := time.Duration(params.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	tries := params.Tries
	if tries <= 0 {
		tries = 4
	}
	port := params.Port
	if port <= 0 {
		port = 443
	}
	mode, err := prober.ParseMode(params.Mode)
	if err != nil {
		mode = prober.ModeHTTP
	}
	sni := params.SNI
	if sni == "" {
		sni = "speed.cloudflare.com"
	}
	speedBytes := int64(0)
	if params.SpeedTest && mode == prober.ModeHTTP {
		speedBytes = 128 * 1024
	}

	src, err := ipsrc.New(true, false, nil)
	if err != nil {
		a.emitErr("scan", err.Error())
		return
	}

	engCfg := engine.Config{
		Concurrency: params.Workers,
		ProbeConfig: prober.Config{
			Port:             port,
			Mode:             mode,
			Tries:            tries,
			Timeout:          timeout,
			SNI:              sni,
			SpeedBytes:       speedBytes,
			RequireWebSocket: params.RequireWS && mode == prober.ModeHTTP,
		},
	}
	eng := engine.New(engCfg)

	coloSet := buildColoSet(params.ColoFilter)
	ipStream := src.MahsaNGV4Stream(ctx, params.Count)

	stats := ScanStats{Total: params.Count}
	eng.Run(ctx, ipStream, func(r *result.Result) {
		s := eng.Stats()
		stats.Tested = s.Tested.Load()
		stats.Healthy = s.Healthy.Load()
		stats.Failed = s.Failed.Load()
		stats.InFlight = s.InFlight.Load()
		runtime.EventsEmit(a.ctx, "scan:stats", stats)
		if !passesColoFilter(r, coloSet) {
			return
		}
		runtime.EventsEmit(a.ctx, "scan:result", toScanResult(r))
	})
	runtime.EventsEmit(a.ctx, "scan:done", stats)
}

// StopScan cancels the running scan.
func (a *App) StopScan() {
	if a.scanCancel != nil {
		a.scanCancel()
	}
}

// ValidationParams drives Phase 2 (xray validation) of the top IPs.
type ValidationParams struct {
	ConfigURL string `json:"configUrl"`
	TopN      int    `json:"topN"`
	TimeoutMs int    `json:"timeoutMs"`
}

// ValidationOutcome is the result of validating one IP through xray.
type ValidationOutcome struct {
	IP              string  `json:"ip"`
	Port            int     `json:"port"`
	Success         bool    `json:"success"`
	LatencyMs       float64 `json:"latencyMs"`
	Throughput      float64 `json:"throughput"`
	UploadThroughput float64 `json:"uploadThroughput"`
	Error           string  `json:"error"`
}

// ValidateTopIPs validates the best candidates through xray and emits
// "validate:result" per IP plus "validate:done" at the end.
func (a *App) ValidateTopIPs(params ValidationParams, candidates []ScanResult) error {
	cfg, err := xraytest.ParseProxyURL(params.ConfigURL)
	if err != nil {
		return fmt.Errorf("invalid config URL: %v", err)
	}

	timeout := time.Duration(params.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	topN := params.TopN
	if topN <= 0 {
		topN = 10
	}

	healthy := make([]result.Result, 0, len(candidates))
	for _, c := range candidates {
		if c.Healthy {
			healthy = append(healthy, result.Result{
				IP:        net.ParseIP(c.IP),
				Port:      c.Port,
				Latencies: []time.Duration{time.Duration(c.AvgMs) * time.Millisecond},
			})
		}
	}
	sort.Slice(healthy, func(i, j int) bool {
		return healthy[i].Avg() < healthy[j].Avg()
	})
	if len(healthy) > topN {
		healthy = healthy[:topN]
	}

	go func() {
		for i := range healthy {
			if ctx := a.ctx; ctx != nil && ctx.Err() != nil {
				break
			}
			probeCfg := cfg.WithEndpoint(healthy[i].IP.String(), healthy[i].Port)
			res := xraytest.ValidateConfig(context.Background(), probeCfg, timeout)
			runtime.EventsEmit(a.ctx, "validate:result", ValidationOutcome{
				IP:               res.IP,
				Port:             res.Port,
				Success:          res.Success,
				LatencyMs:        float64(res.Latency) / float64(time.Millisecond),
				Throughput:       res.Throughput,
				UploadThroughput: res.UploadThroughput,
				Error:            res.Error,
			})
		}
		runtime.EventsEmit(a.ctx, "validate:done", map[string]interface{}{"count": len(healthy)})
	}()
	return nil
}

// ExportBundle is returned to the frontend for saving/copying.
type ExportBundle struct {
	Subscription string   `json:"subscription"`
	ShareURLs    []string `json:"shareUrls"`
	SingBox      string   `json:"singBox"`
	Clash        string   `json:"clash"`
	Count        int      `json:"count"`
}

// GenerateConfigs builds export content from one template config URL and a
// list of working endpoints ("ip:port" strings).
func (a *App) GenerateConfigs(configURL string, endpoints []string) (*ExportBundle, error) {
	cfg, err := xraytest.ParseProxyURL(configURL)
	if err != nil {
		return nil, fmt.Errorf("invalid config URL: %v", err)
	}
	eps := export.ParseEndpoints(endpoints)
	bundle, err := export.Generate(cfg, eps)
	if err != nil {
		return nil, err
	}
	return &ExportBundle{
		Subscription: bundle.Subscription,
		ShareURLs:    bundle.ShareURLs,
		SingBox:      bundle.SingBox,
		Clash:        bundle.Clash,
		Count:        len(eps),
	}, nil
}

// SaveText writes text to a user-chosen location via the native file dialog.
func (a *App) SaveText(defaultName, content string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultName,
		Title:           "Save exported configs",
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := writeFile(path, content); err != nil {
		return "", err
	}
	return path, nil
}

// CopyText copies text to the system clipboard.
func (a *App) CopyText(text string) error {
	return copyToClipboard(text)
}

func (a *App) emitErr(kind, text string) {
	runtime.EventsEmit(a.ctx, kind+":error", text)
}

// --- helpers ----------------------------------------------------------------

func buildColoSet(raw string) map[string]bool {
	if raw == "" {
		return nil
	}
	set := make(map[string]bool)
	for _, c := range splitTrim(raw, ",") {
		set[normalizeColo(c)] = true
	}
	return set
}

func passesColoFilter(r *result.Result, set map[string]bool) bool {
	if set == nil {
		return true
	}
	return set[normalizeColo(r.Colo)]
}

func normalizeColo(s string) string {
	if len(s) < 3 {
		return ""
	}
	return s[:3]
}

func splitTrim(s, sep string) []string {
	var out []string
	start := 0
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			if v := trimSpace(s[start:i]); v != "" {
				out = append(out, v)
			}
			start = i + len(sep)
		}
	}
	if v := trimSpace(s[start:]); v != "" {
		out = append(out, v)
	}
	return out
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

var _ = fmt.Sprintf
