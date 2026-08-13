package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/matinsenpai/senpaiscanner/internal/export"
	"github.com/matinsenpai/senpaiscanner/internal/ipsrc"
	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/ui"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
	"github.com/matinsenpai/senpaiscanner/pkg/version"
)

// App is the Wails application backend. It mirrors the CLI two-phase flow:
// Phase 1 probes IPs (config URL optional), Phase 2 validates the top IPs
// through xray when a config URL is given.
type App struct {
	ctx context.Context

	mu            sync.Mutex
	scanning      bool
	lastParams    ScanParams
	phase1Results []*result.Result

	scanID     atomic.Int64
	scanCancel context.CancelFunc
}

// NewApp creates the backend.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go func() {
		runtime.EventsEmit(a.ctx, "app:meta", ui.FetchMeta())
	}()
}

// GetVersion returns the build version string.
func (a *App) GetVersion() string {
	return version.String()
}

// ---------------------------------------------------------------------------
// Presets
// ---------------------------------------------------------------------------

// PresetData carries every preset the frontend needs to render chips. The
// labels/values come straight from the CLI so index mapping can never drift.
type PresetData struct {
	CountLabels     []string `json:"countLabels"`
	CountValues     []string `json:"countValues"`
	WorkerLabels    []string `json:"workerLabels"`
	WorkerValues    []string `json:"workerValues"`
	TimeoutLabels   []string `json:"timeoutLabels"`
	TimeoutValues   []string `json:"timeoutValues"`
	TopNLabels      []string `json:"topNLabels"`
	TopNValues      []string `json:"topNValues"`
	MinSpeedLabels  []string `json:"minSpeedLabels"`
	MinSpeedValues  []string `json:"minSpeedValues"`
	SpeedSizeLabels []string `json:"speedSizeLabels"`
	SpeedSizeValues []string `json:"speedSizeValues"`
	Ports           []int    `json:"ports"`
}

// Presets returns all CLI presets plus the port chip list.
func (a *App) Presets() PresetData {
	count := ui.ConfigCountPresets()
	workers := ui.ConfigWorkerPresets()
	timeout := ui.ConfigTimeoutPresets()
	topN := ui.ConfigTopNPresets()
	minSpeed := ui.ConfigMinSpeedPresets()
	speedSize := ui.ConfigSpeedSizePresets()
	return PresetData{
		CountLabels:     count.Labels,
		CountValues:     count.Values,
		WorkerLabels:    workers.Labels,
		WorkerValues:    workers.Values,
		TimeoutLabels:   timeout.Labels,
		TimeoutValues:   timeout.Values,
		TopNLabels:      topN.Labels,
		TopNValues:      topN.Values,
		MinSpeedLabels:  minSpeed.Labels,
		MinSpeedValues:  minSpeed.Values,
		SpeedSizeLabels: speedSize.Labels,
		SpeedSizeValues: speedSize.Values,
		Ports:           ui.ConfigPorts(),
	}
}

// ---------------------------------------------------------------------------
// Scan parameters
// ---------------------------------------------------------------------------

// ScanParams carries resolved values (used to run the scan) plus the raw
// index/custom fields that round-trip into the shared CLI config file.
type ScanParams struct {
	// Resolved values
	IPMode       int     `json:"ipMode"` // 0 = random, 1 = ips.txt
	IPFile       string  `json:"ipFile"` // explicit ips file path (IPMode == 1); empty = auto-discover
	Count        int     `json:"count"`
	Workers      int     `json:"workers"`
	TimeoutMs    int     `json:"timeoutMs"`
	Ports        []int   `json:"ports"` // 0 = config port
	ConfigURL    string  `json:"configUrl"`
	RequireWS    bool    `json:"requireWS"`
	TopN         int     `json:"topN"` // 0 = all
	MinSpeed     float64 `json:"minSpeed"`
	SpeedURL     string  `json:"speedUrl"`
	SpeedSize    int64   `json:"speedSize"` // bytes
	UploadTest   bool    `json:"uploadTest"`
	UploadSize   int64   `json:"uploadSize"` // bytes; 0 = fall back to SpeedSize
	UploadURL    string  `json:"uploadUrl"`
	SpeedMode    string  `json:"speedMode"` // "download" | "upload" | "both"
	NeighborScan bool    `json:"neighborScan"`

	// Round-trip fields for the shared config file
	CountIdx        int    `json:"countIdx"`
	CountCustom     string `json:"countCustom"`
	WorkersIdx      int    `json:"workersIdx"`
	WorkersCustom   string `json:"workersCustom"`
	TimeoutIdx      int    `json:"timeoutIdx"`
	TimeoutCustom   string `json:"timeoutCustom"`
	TopNIdx         int    `json:"topNIdx"`
	TopNCustom      string `json:"topNCustom"`
	MinSpeedIdx     int    `json:"minSpeedIdx"`
	MinSpeedCustom  string `json:"minSpeedCustom"`
	SpeedSizeIdx    int    `json:"speedSizeIdx"`
	SpeedSizeCustom string `json:"speedSizeCustom"`
	UploadSizeIdx    int    `json:"uploadSizeIdx"`
	UploadSizeCustom string `json:"uploadSizeCustom"`
}

func (p ScanParams) toSavedConfig() ui.SavedConfig {
	ports := append([]int(nil), p.Ports...)
	if len(ports) == 0 {
		ports = []int{0}
	}
	return ui.SavedConfig{
		IPMode:          p.IPMode,
		IPFile:          strings.TrimSpace(p.IPFile),
		CountIdx:        p.CountIdx,
		CountCustom:     p.CountCustom,
		WorkersIdx:      p.WorkersIdx,
		WorkersCustom:   p.WorkersCustom,
		TimeoutIdx:      p.TimeoutIdx,
		TimeoutCustom:   p.TimeoutCustom,
		Ports:           ports,
		ConfigURL:       strings.TrimSpace(p.ConfigURL),
		TopNIdx:         p.TopNIdx,
		TopNCustom:      p.TopNCustom,
		MinSpeedIdx:     p.MinSpeedIdx,
		MinSpeedCustom:  p.MinSpeedCustom,
		SpeedURL:        strings.TrimSpace(p.SpeedURL),
		SpeedSizeIdx:    p.SpeedSizeIdx,
		SpeedSizeCustom: p.SpeedSizeCustom,
		UploadURL:        strings.TrimSpace(p.UploadURL),
		UploadSizeIdx:    p.UploadSizeIdx,
		UploadSizeCustom: p.UploadSizeCustom,
		UploadTest:      p.UploadTest,
		SpeedMode:       p.SpeedMode,
		RequireWS:       p.RequireWS,
		NeighborScan:    p.NeighborScan,
	}
}

// ---------------------------------------------------------------------------
// Events
// ---------------------------------------------------------------------------

// ScanResult is one finished probe, sent to the UI in batches.
type ScanResult struct {
	IP         string  `json:"ip"`
	Port       int     `json:"port"`
	Colo       string  `json:"colo"`
	AvgMs      float64 `json:"avgMs"`
	Loss       float64 `json:"loss"`
	JitterMs   float64 `json:"jitterMs"`
	Throughput float64 `json:"throughput"`
	Healthy    bool    `json:"healthy"`
}

func toScanResult(r *result.Result) ScanResult {
	return ScanResult{
		IP:         r.IP.String(),
		Port:       r.Port,
		Colo:       r.Colo,
		AvgMs:      float64(r.Avg()) / float64(time.Millisecond),
		Loss:       r.Loss(),
		JitterMs:   float64(r.Jitter()) / float64(time.Millisecond),
		Throughput: r.Throughput,
		Healthy:    r.IsHealthy(),
	}
}

// ValidationOutcome is one Phase 2 result.
type ValidationOutcome struct {
	IP               string  `json:"ip"`
	Port             int     `json:"port"`
	Transport        string  `json:"transport"`
	Success          bool    `json:"success"`
	LatencyMs        float64 `json:"latencyMs"`
	Throughput       float64 `json:"throughput"`
	UploadThroughput float64 `json:"uploadThroughput"`
	Error            string  `json:"error"`
	Done             int     `json:"done"`
	Total            int     `json:"total"`
}

// emit emits an event tagged with scanID so events from a stopped scan never
// leak into a newer run.
func (a *App) emit(scanID int64, name string, payload any) {
	if scanID != a.scanID.Load() {
		return
	}
	runtime.EventsEmit(a.ctx, name, payload)
}

// ---------------------------------------------------------------------------
// Scan lifecycle
// ---------------------------------------------------------------------------

// StartScan persists the settings (shared with CLI Retry Last Scan), then
// runs Phase 1 and — when a config URL is present — Phase 2.
func (a *App) StartScan(params ScanParams) error {
	a.mu.Lock()
	if a.scanning {
		a.mu.Unlock()
		return fmt.Errorf("a scan is already running")
	}
	a.scanning = true
	a.lastParams = params
	a.phase1Results = nil
	a.mu.Unlock()

	// Persist settings before running so "Retry Last Scan" (both apps) sees
	// them even if the scan is stopped immediately.
	_ = ui.SaveAppConfig(ui.AppConfig{LastConfig: params.toSavedConfig()})

	scanID := a.scanID.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	a.mu.Lock()
	a.scanCancel = cancel
	a.mu.Unlock()

	go func() {
		defer func() {
			cancel()
			a.mu.Lock()
			a.scanCancel = nil
			a.scanning = false
			a.mu.Unlock()
		}()
		a.runScan(ctx, scanID, params)
	}()
	return nil
}

// StopScan cancels the running scan (both phases).
func (a *App) StopScan() {
	a.mu.Lock()
	cancel := a.scanCancel
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (a *App) runScan(ctx context.Context, scanID int64, params ScanParams) {
	configURL := strings.TrimSpace(params.ConfigURL)
	timeout := time.Duration(params.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	// Probe config: exact CLI derivation.
	probeCfg, err := ui.Phase1ProbeConfig(configURL, timeout, params.RequireWS)
	if err != nil {
		a.emit(scanID, "scan:error", fmt.Sprintf("invalid URL: %v", err))
		a.emit(scanID, "scan:done", map[string]any{"cancelled": false})
		return
	}

	// Port resolution: 0 = config/URL port.
	ports := resolvePorts(params.Ports, configURL, probeCfg.Port)

	var ipStream <-chan net.IP
	neighbor := ui.NeighborScanOpts{}
	count := params.Count
	if count <= 0 {
		count = 1000
	}
	totalTarget := count * len(ports)
	if params.IPMode == 1 {
		ips, err := ui.LoadIPsFromPath(params.IPFile)
		if err != nil {
			a.emit(scanID, "scan:error", err.Error())
			a.emit(scanID, "scan:done", map[string]any{"cancelled": false})
			return
		}
		if len(ips) == 0 {
			a.emit(scanID, "scan:error", "ips.txt is empty — add one IP per line")
			a.emit(scanID, "scan:done", map[string]any{"cancelled": false})
			return
		}
		if len(ips) > count {
			ips = ips[:count]
		}
		totalTarget = len(ips) * len(ports)
		ch := make(chan net.IP, len(ips))
		for _, ip := range ips {
			ch <- ip
		}
		close(ch)
		ipStream = ch
	} else {
		src, err := ipsrc.New(true, false, nil)
		if err != nil {
			a.emit(scanID, "scan:error", err.Error())
			a.emit(scanID, "scan:done", map[string]any{"cancelled": false})
			return
		}
		ipStream = src.MahsaNGV4Stream(ctx, count)
		if params.NeighborScan {
			neighbor = ui.DefaultNeighborOpts(src.IPv4Nets())
		}
	}

	writer, livePath, _ := ui.NewLiveResultWriter(configURL != "")
	a.emit(scanID, "scan:phase", map[string]any{"phase": 1, "livePath": livePath})

	// Batching emitter: collect results + counters, flush every250ms.
	var batchMu sync.Mutex
	var pending []ScanResult
	var allResults []*result.Result
	var tested, healthy atomic.Int64
	var phase1Collected = &allResults

	flush := func() {
		batchMu.Lock()
		batch := pending
		pending = nil
		t := int(tested.Load())
		h := int(healthy.Load())
		batchMu.Unlock()

		a.emit(scanID, "scan:stats", map[string]any{
			"tested": t, "healthy": h, "failed": t - h, "total": totalTarget,
		})
		if len(batch) > 0 {
			a.emit(scanID, "scan:results", batch)
		}
	}

	callback := func(r *result.Result) {
		if writer != nil {
			writer.AddPhase1(r)
		}
		if r.IsHealthy() {
			healthy.Add(1)
		}
		tested.Add(1)
		batchMu.Lock()
		*phase1Collected = append(*phase1Collected, r)
		pending = append(pending, toScanResult(r))
		batchMu.Unlock()
	}

	ticker := time.NewTicker(250 * time.Millisecond)
	doneTick := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				flush()
			case <-doneTick:
				ticker.Stop()
				return
			}
		}
	}()

	ui.RunPortProbes(ctx, ipStream, ports, params.Workers, probeCfg, callback, neighbor)

	close(doneTick)
	flush()

	cancelled := ctx.Err() != nil
	batchMu.Lock()
	collected := *phase1Collected
	batchMu.Unlock()
	a.mu.Lock()
	a.phase1Results = append([]*result.Result(nil), collected...)
	a.lastParams = params
	a.mu.Unlock()

	if configURL == "" {
		if writer != nil {
			writer.FinishPhase1Only()
		}
		a.emit(scanID, "scan:done", map[string]any{
			"cancelled":        cancelled,
			"healthy":          int(healthy.Load()),
			"workingEndpoints": []string{},
		})
		return
	}

	// Phase 2: validate top IPs through xray.
	topIPs := result.TopN(collected, params.TopN)
	if len(topIPs) == 0 || cancelled {
		a.emit(scanID, "scan:done", map[string]any{
			"cancelled":        cancelled,
			"healthy":          int(healthy.Load()),
			"workingEndpoints": []string{},
		})
		return
	}

	if writer != nil {
		writer.BeginPhase2()
	}
	a.emit(scanID, "scan:phase", map[string]any{"phase": 2, "livePath": livePath})

	xrayTimeout := ui.Phase2Timeout(timeout, params.MinSpeed, params.SpeedSize)
	var validationResults []*xraytest.ValidationResult
	var valMu sync.Mutex

	err = ui.RunPhase2(ctx, configURL, topIPs, params.MinSpeed, params.SpeedURL,
		params.SpeedSize, xrayTimeout, params.UploadTest, params.UploadSize, params.UploadURL, params.SpeedMode,
		func(vr *xraytest.ValidationResult, done, total int) {
			if writer != nil {
				writer.AddPhase2(vr)
			}
			valMu.Lock()
			validationResults = append(validationResults, vr)
			valMu.Unlock()
			a.emit(scanID, "validate:result", ValidationOutcome{
				IP:               vr.IP,
				Port:             vr.Port,
				Transport:        vr.Transport,
				Success:          vr.Success,
				LatencyMs:        float64(vr.Latency) / float64(time.Millisecond),
				Throughput:       vr.Throughput,
				UploadThroughput: vr.UploadThroughput,
				Error:            vr.Error,
				Done:             done,
				Total:            total,
			})
		})
	if err != nil {
		a.emit(scanID, "scan:error", err.Error())
	}

	valMu.Lock()
	endpoints := ui.WorkingEndpoints(validationResults)
	valMu.Unlock()

	a.emit(scanID, "scan:done", map[string]any{
		"cancelled":        ctx.Err() != nil,
		"healthy":          int(healthy.Load()),
		"workingEndpoints": endpoints,
	})
}

// StartSpeedTest tests every currently healthy Phase 1 result. It is intended
// for the explicit post-stop action in the GUI: users can stop a long scan as
// soon as they have enough green rows, then measure those rows immediately.
// With a config URL it validates through xray; without one it runs a direct
// Cloudflare download sample so the action remains useful for Phase 1 scans.
func (a *App) StartSpeedTest(params ScanParams) error {
	a.mu.Lock()
	if a.scanning {
		a.mu.Unlock()
		return fmt.Errorf("stop the current scan before starting a speed test")
	}
	captured := append([]*result.Result(nil), a.phase1Results...)
	a.scanning = true
	a.lastParams = params
	a.mu.Unlock()

	candidates := result.TopN(captured, 0)
	if len(candidates) == 0 {
		a.mu.Lock()
		a.scanning = false
		a.mu.Unlock()
		return fmt.Errorf("no healthy results are available for a speed test")
	}

	scanID := a.scanID.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	a.mu.Lock()
	a.scanCancel = cancel
	a.mu.Unlock()

	go func() {
		defer func() {
			cancel()
			a.mu.Lock()
			a.scanCancel = nil
			a.scanning = false
			a.mu.Unlock()
		}()
		a.runSpeedTest(ctx, scanID, params, candidates)
	}()
	return nil
}

func (a *App) runSpeedTest(ctx context.Context, scanID int64, params ScanParams, candidates []*result.Result) {
	configURL := strings.TrimSpace(params.ConfigURL)
	a.emit(scanID, "scan:phase", map[string]any{
		"phase": 2, "manual": true, "mode": map[bool]string{true: "tunnel", false: "direct"}[configURL != ""],
	})

	if configURL != "" {
		timeout := time.Duration(params.TimeoutMs) * time.Millisecond
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		xrayTimeout := ui.Phase2Timeout(timeout, params.MinSpeed, params.SpeedSize)
		var outcomesMu sync.Mutex
		var outcomes []*xraytest.ValidationResult
		err := ui.RunPhase2(ctx, configURL, candidates, params.MinSpeed, params.SpeedURL,
			params.SpeedSize, xrayTimeout, params.UploadTest, params.UploadSize, params.UploadURL, params.SpeedMode,
			func(vr *xraytest.ValidationResult, done, total int) {
				outcomesMu.Lock()
				outcomes = append(outcomes, vr)
				outcomesMu.Unlock()
				a.emit(scanID, "validate:result", validationOutcome(vr, done, total))
			})
		if err != nil {
			a.emit(scanID, "scan:error", err.Error())
		}
		outcomesMu.Lock()
		endpoints := ui.WorkingEndpoints(outcomes)
		outcomesMu.Unlock()
		a.emit(scanID, "scan:done", map[string]any{
			"cancelled": ctx.Err() != nil, "manualSpeed": true,
			"healthy": len(candidates), "workingEndpoints": endpoints,
		})
		return
	}

	// Direct mode uses one download-bearing HTTP probe per endpoint. The
	// original scan results are never overwritten, so copy actions remain live.
	timeout := time.Duration(params.TimeoutMs) * time.Millisecond
	if timeout < 10*time.Second {
		timeout = 10 * time.Second
	}
	sampleBytes := params.SpeedSize
	if sampleBytes <= 0 {
		sampleBytes = 512 * 1024
	}
	base := prober.Config{
		Mode: prober.ModeHTTP, Tries: 1, Timeout: timeout,
		SNI: "speed.cloudflare.com", SpeedBytes: sampleBytes,
		InsecureSkipVerify: true,
	}
	workers := params.Workers
	if workers <= 0 {
		workers = 10
	}
	if workers > 20 {
		workers = 20
	}

	jobs := make(chan *result.Result)
	var wg sync.WaitGroup
	var done atomic.Int64
	var endpointMu sync.Mutex
	var endpoints []string
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for candidate := range jobs {
				if ctx.Err() != nil {
					return
				}
				measured := prober.Probe(ctx, candidate.IP, base.WithPort(candidate.Port))
				n := int(done.Add(1))
				success := measured.IsHealthy() && measured.Throughput > 0
				errText := ""
				if !success {
					errText = "download sample failed"
				} else {
					endpointMu.Lock()
					endpoints = append(endpoints, net.JoinHostPort(candidate.IP.String(), strconv.Itoa(candidate.Port)))
					endpointMu.Unlock()
				}
				a.emit(scanID, "validate:result", ValidationOutcome{
					IP: candidate.IP.String(), Port: candidate.Port, Transport: "direct",
					Success: success, LatencyMs: float64(measured.Avg()) / float64(time.Millisecond),
					Throughput: measured.Throughput, Error: errText, Done: n, Total: len(candidates),
				})
			}
		}()
	}
sendCandidates:
	for _, candidate := range candidates {
		select {
		case jobs <- candidate:
		case <-ctx.Done():
			break sendCandidates
		}
	}
	close(jobs)
	wg.Wait()
	endpointMu.Lock()
	completed := append([]string(nil), endpoints...)
	endpointMu.Unlock()
	a.emit(scanID, "scan:done", map[string]any{
		"cancelled": ctx.Err() != nil, "manualSpeed": true,
		"healthy": len(candidates), "workingEndpoints": completed,
	})
}

func validationOutcome(vr *xraytest.ValidationResult, done, total int) ValidationOutcome {
	return ValidationOutcome{
		IP: vr.IP, Port: vr.Port, Transport: vr.Transport, Success: vr.Success,
		LatencyMs:  float64(vr.Latency) / float64(time.Millisecond),
		Throughput: vr.Throughput, UploadThroughput: vr.UploadThroughput,
		Error: vr.Error, Done: done, Total: total,
	}
}

// resolvePorts mirrors the CLI's resolveConfigPorts: explicit ports win;
// otherwise fall back to the URL-derived port (probeCfg.Port).
func resolvePorts(selected []int, configURL string, probePort int) []int {
	var ports []int
	for _, p := range selected {
		if p > 0 {
			ports = append(ports, p)
		}
	}
	if len(ports) > 0 {
		return ports
	}
	if strings.TrimSpace(configURL) != "" {
		if probePort > 0 {
			return []int{probePort}
		}
		return []int{443}
	}
	return []int{probePort}
}

// ---------------------------------------------------------------------------
// Retry last scan
// ---------------------------------------------------------------------------

// RetryLastScan loads the shared CLI config file, resolves indices back to
// values, and starts a new scan. It returns the resolved params so the
// frontend can re-sync its form.
func (a *App) RetryLastScan() (ScanParams, error) {
	cfg := ui.LoadAppConfig().LastConfig

	count, _ := presetValueInt(ui.ConfigCountPresets(), cfg.CountIdx, cfg.CountCustom, 1000)
	workers, _ := presetValueInt(ui.ConfigWorkerPresets(), cfg.WorkersIdx, cfg.WorkersCustom, 50)
	timeoutStr := presetValueStr(ui.ConfigTimeoutPresets(), cfg.TimeoutIdx, cfg.TimeoutCustom, "5s")
	timeoutMs := 5000
	if d, err := time.ParseDuration(timeoutStr); err == nil && d > 0 {
		timeoutMs = int(d.Milliseconds())
	}
	topN, _ := presetValueInt(ui.ConfigTopNPresets(), cfg.TopNIdx, cfg.TopNCustom, 50)
	minSpeed := presetValueFloat(ui.ConfigMinSpeedPresets(), cfg.MinSpeedIdx, cfg.MinSpeedCustom, 0)
	speedSize := presetValueInt64(ui.ConfigSpeedSizePresets(), cfg.SpeedSizeIdx, cfg.SpeedSizeCustom, 512*1024)
	uploadSize := presetValueInt64(ui.ConfigSpeedSizePresets(), cfg.UploadSizeIdx, cfg.UploadSizeCustom, 512*1024)

	return ScanParams{
		IPMode:       cfg.IPMode,
		IPFile:       cfg.IPFile,
		Count:        count,
		Workers:      workers,
		TimeoutMs:    timeoutMs,
		Ports:        cfg.Ports,
		ConfigURL:    cfg.ConfigURL,
		RequireWS:    cfg.RequireWS,
		TopN:         topN,
		MinSpeed:     minSpeed,
		SpeedURL:     cfg.SpeedURL,
		SpeedSize:    speedSize,
		UploadTest:   cfg.UploadTest,
		UploadSize:   uploadSize,
		UploadURL:    cfg.UploadURL,
		SpeedMode:    cfg.SpeedMode,
		NeighborScan: cfg.NeighborScan,

		CountIdx:         cfg.CountIdx,
		CountCustom:      cfg.CountCustom,
		WorkersIdx:       cfg.WorkersIdx,
		WorkersCustom:    cfg.WorkersCustom,
		TimeoutIdx:       cfg.TimeoutIdx,
		TimeoutCustom:    cfg.TimeoutCustom,
		TopNIdx:          cfg.TopNIdx,
		TopNCustom:       cfg.TopNCustom,
		MinSpeedIdx:      cfg.MinSpeedIdx,
		MinSpeedCustom:   cfg.MinSpeedCustom,
		SpeedSizeIdx:     cfg.SpeedSizeIdx,
		SpeedSizeCustom:  cfg.SpeedSizeCustom,
		UploadSizeIdx:    cfg.UploadSizeIdx,
		UploadSizeCustom: cfg.UploadSizeCustom,
	}, nil
}

// presetValueInt resolves an index+custom preset to an int value.
func presetValueInt(list ui.PresetList, idx int, custom string, fallback int) (int, bool) {
	if idx == len(list.Labels)-1 {
		v, err := strconv.Atoi(strings.TrimSpace(custom))
		if err != nil || v <= 0 {
			return fallback, false
		}
		return v, true
	}
	if idx >= 0 && idx < len(list.Values) {
		if v, err := strconv.Atoi(list.Values[idx]); err == nil {
			return v, true
		}
	}
	return fallback, false
}

func presetValueInt64(list ui.PresetList, idx int, custom string, fallback int64) int64 {
	if idx == len(list.Labels)-1 {
		v, err := strconv.ParseInt(strings.TrimSpace(custom), 10, 64)
		if err != nil || v <= 0 {
			return fallback
		}
		return v * 1024 * 1024 // custom speed size is in MB
	}
	if idx >= 0 && idx < len(list.Values) {
		if v, err := strconv.ParseInt(list.Values[idx], 10, 64); err == nil && v > 0 {
			return v
		}
	}
	return fallback
}

func presetValueFloat(list ui.PresetList, idx int, custom string, fallback float64) float64 {
	if idx == len(list.Labels)-1 {
		v, err := strconv.ParseFloat(strings.TrimSpace(custom), 64)
		if err != nil || v <= 0 {
			return fallback
		}
		return v
	}
	if idx >= 0 && idx < len(list.Values) {
		if v, err := strconv.ParseFloat(list.Values[idx], 64); err == nil && v >= 0 {
			return v
		}
	}
	return fallback
}

func presetValueStr(list ui.PresetList, idx int, custom, fallback string) string {
	if idx == len(list.Labels)-1 {
		if v := strings.TrimSpace(custom); v != "" {
			return v
		}
		return fallback
	}
	if idx >= 0 && idx < len(list.Values) && list.Values[idx] != "" {
		return list.Values[idx]
	}
	return fallback
}

// ---------------------------------------------------------------------------
// Export
// ---------------------------------------------------------------------------

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

// ExportAllToDisk mirrors the CLI's 'e' action: writes subscription txt,
// sing-box JSON, and Clash YAML next to the app binary.
func (a *App) ExportAllToDisk(configURL string, endpoints []string) (string, error) {
	cfg, err := xraytest.ParseProxyURL(configURL)
	if err != nil {
		return "", fmt.Errorf("invalid config URL: %v", err)
	}
	bundle, err := export.Generate(cfg, export.ParseEndpoints(endpoints))
	if err != nil {
		return "", fmt.Errorf("export failed: %v", err)
	}

	dir := ""
	if exe, err := os.Executable(); err == nil {
		dir = filepath.Dir(exe)
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	if dir == "" {
		dir = "."
	}

	write := func(name, content string) error {
		return os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	}
	if err := write("senpaiscanner-sub.txt", bundle.Subscription); err != nil {
		return "", err
	}
	if err := write("senpaiscanner-singbox.json", bundle.SingBox); err != nil {
		return "", err
	}
	if err := write("senpaiscanner-clash.yaml", bundle.Clash); err != nil {
		return "", err
	}
	return dir, nil
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

// CopyText copies text to the system clipboard.
func (a *App) CopyText(text string) error {
	return runtime.ClipboardSetText(a.ctx, text)
}

// SelectIPsFile opens a native open-file dialog so the user can pick the IP
// list to scan. Returns the chosen path (empty if the dialog was cancelled).
func (a *App) SelectIPsFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select IP list",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text / CSV (*.txt;*.csv)", Pattern: "*.txt;*.csv"},
			{DisplayName: "All files (*.*)", Pattern: "*.*"},
		},
	})
}

// ---------------------------------------------------------------------------
// Single-endpoint re-test (right-click actions)
// ---------------------------------------------------------------------------

func (a *App) busy() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.scanning
}

// RetestReachability re-probes one IP:port with the Phase 1 reachability probe
// derived from the current form settings, returning a fresh scan result.
func (a *App) RetestReachability(params ScanParams, ip string, port int) (ScanResult, error) {
	if a.busy() {
		return ScanResult{}, fmt.Errorf("stop the current scan before re-testing")
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return ScanResult{}, fmt.Errorf("invalid IP: %s", ip)
	}
	timeout := time.Duration(params.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	configURL := strings.TrimSpace(params.ConfigURL)
	probeCfg, err := ui.Phase1ProbeConfig(configURL, timeout, params.RequireWS)
	if err != nil {
		return ScanResult{}, fmt.Errorf("invalid URL: %v", err)
	}
	if port <= 0 {
		port = probeCfg.Port
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Duration(max(1, probeCfg.Tries))+2*time.Second)
	defer cancel()
	r := prober.Probe(ctx, parsed, probeCfg.WithPort(port))
	return toScanResult(r), nil
}

// RetestEndpoint re-validates one IP:port. With a config URL it routes through
// xray (tunnel validation); without one it runs a direct Cloudflare download
// sample, mirroring the batch speed test.
func (a *App) RetestEndpoint(params ScanParams, ip string, port int) (ValidationOutcome, error) {
	if a.busy() {
		return ValidationOutcome{}, fmt.Errorf("stop the current scan before re-testing")
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return ValidationOutcome{}, fmt.Errorf("invalid IP: %s", ip)
	}
	if port <= 0 {
		port = 443
	}
	configURL := strings.TrimSpace(params.ConfigURL)
	candidate := &result.Result{IP: parsed, Port: port}

	if configURL != "" {
		timeout := time.Duration(params.TimeoutMs) * time.Millisecond
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		xrayTimeout := ui.Phase2Timeout(timeout, params.MinSpeed, params.SpeedSize)
		ctx, cancel := context.WithTimeout(context.Background(), xrayTimeout+5*time.Second)
		defer cancel()
		var outcome ValidationOutcome
		err := ui.RunPhase2(ctx, configURL, []*result.Result{candidate}, params.MinSpeed,
			params.SpeedURL, params.SpeedSize, xrayTimeout, params.UploadTest, params.UploadSize, params.UploadURL, params.SpeedMode,
			func(vr *xraytest.ValidationResult, done, total int) {
				outcome = validationOutcome(vr, 1, 1)
			})
		if err != nil {
			return ValidationOutcome{}, err
		}
		return outcome, nil
	}

	// Direct download sample (no config URL).
	timeout := time.Duration(params.TimeoutMs) * time.Millisecond
	if timeout < 10*time.Second {
		timeout = 10 * time.Second
	}
	sampleBytes := params.SpeedSize
	if sampleBytes <= 0 {
		sampleBytes = 512 * 1024
	}
	base := prober.Config{
		Mode: prober.ModeHTTP, Tries: 1, Timeout: timeout,
		SNI: "speed.cloudflare.com", SpeedBytes: sampleBytes,
		InsecureSkipVerify: true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout+2*time.Second)
	defer cancel()
	measured := prober.Probe(ctx, parsed, base.WithPort(port))
	success := measured.IsHealthy() && measured.Throughput > 0
	errText := ""
	if !success {
		errText = "download sample failed"
	}
	return ValidationOutcome{
		IP: parsed.String(), Port: port, Transport: "direct", Success: success,
		LatencyMs: float64(measured.Avg()) / float64(time.Millisecond),
		Throughput: measured.Throughput, Error: errText, Done: 1, Total: 1,
	}, nil
}
