package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	configexport "github.com/matinsenpai/senpaiscanner/internal/export"
	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/ui"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

type Callback interface {
	OnProgress(tested int, healthy int, failed int, inFlight int, isPhase2 bool)
	OnResult(ip string, port int, latencyMs int, loss float64, colo string, isHealthy bool, isPhase2 bool, phase2Type string, phase2Speed float64, phase2Status bool, phase2UploadSpeed float64)
	OnFinished()
	OnError(err string)
	// OnInfo carries notices that do not end the scan (resume message, skipped targets, save problems).
	OnInfo(msg string)
}

var (
	mu          sync.Mutex
	cancelScan  context.CancelFunc
	isRunning   bool
	phase1Cache []*result.Result
)

type ScanConfig struct {
	SourceType    string  `json:"sourceType"`
	SourceFile    string  `json:"sourceFile"`
	CountType     string  `json:"countType"`
	CustomCount   string  `json:"customCount"`
	WorkerType    string  `json:"workerType"`
	CustomWorkers string  `json:"customWorkers"`
	TimeoutType   string  `json:"timeoutType"`
	CustomTimeout string  `json:"customTimeout"`
	PortType      string  `json:"portType"`
	SelectedPorts []int   `json:"selectedPorts"`
	ConfigURL     string  `json:"configUrl"`
	TopNType      string  `json:"topNType"`
	CustomTopN    string  `json:"customTopN"`
	NeighborScan  bool    `json:"neighborScan"`
	RequireWS     bool    `json:"requireWebSocket"`
	MinSpeed      float64 `json:"minSpeed"`
	SpeedURL      string  `json:"speedUrl"`
	SpeedSize     int64   `json:"speedSize"`
	UploadTest    bool    `json:"uploadTest"`

	// Gentle caps workers/rate; Targets (SourceType "Paste") replaces the random pool; Phase2Only skips
	// reachability; Resume continues the scan saved by the previous session.
	Gentle     bool   `json:"gentle"`
	Targets    string `json:"targets"`
	Phase2Only bool   `json:"phase2Only"`
	Resume     bool   `json:"resume"`

	// Anti-DPI recipe (see internal/antidpi); empty fields with AntiDpiEnabled use the published values.
	AntiDpiEnabled bool   `json:"antiDpiEnabled"`
	AdFinalmask    string `json:"adFinalmask"`
	AdFingerprint  string `json:"adFingerprint"`
	AdAlpn         string `json:"adAlpn"`
	AdCiphers      string `json:"adCiphers"`
}

func StartScan(configJson string, callback Callback) {
	mu.Lock()
	if isRunning {
		mu.Unlock()
		if callback != nil {
			callback.OnError("Scan is already running")
		}
		return
	}
	isRunning = true
	phase1Cache = nil
	mu.Unlock()

	go runScan(configJson, callback)
}

func StopScan() {
	mu.Lock()
	defer mu.Unlock()
	if cancelScan != nil {
		cancelScan()
		cancelScan = nil
	}
}

func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return isRunning
}

// StartSpeedTest measures every healthy Phase 1 endpoint retained from the
// current session. It is intentionally separate from StartScan so Android can
// stop a long scan and immediately test the green rows, matching the desktop UI.
func StartSpeedTest(configJson string, callback Callback) {
	mu.Lock()
	if isRunning {
		mu.Unlock()
		if callback != nil {
			callback.OnError("Stop the current scan before starting a speed test")
		}
		return
	}
	candidates := append([]*result.Result(nil), phase1Cache...)
	if len(candidates) == 0 {
		mu.Unlock()
		if callback != nil {
			callback.OnError("No healthy results are available for a speed test")
		}
		return
	}
	isRunning = true
	mu.Unlock()

	go runSpeedTest(configJson, candidates, callback)
}

func runSpeedTest(configJson string, candidates []*result.Result, callback Callback) {
	defer func() {
		mu.Lock()
		isRunning = false
		cancelScan = nil
		mu.Unlock()
		if callback != nil {
			callback.OnFinished()
		}
	}()

	var cfg ScanConfig
	if err := json.Unmarshal([]byte(configJson), &cfg); err != nil {
		if callback != nil {
			callback.OnError(fmt.Sprintf("Failed to parse config: %v", err))
		}
		return
	}
	if err := applyAntiDpi(cfg); err != nil {
		if callback != nil {
			callback.OnError(err.Error())
		}
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mu.Lock()
	cancelScan = cancel
	mu.Unlock()

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Avg() < candidates[j].Avg() })
	total := len(candidates)
	if callback != nil {
		callback.OnProgress(0, 0, 0, total, true)
	}

	configURL := strings.TrimSpace(cfg.ConfigURL)
	if configURL != "" {
		template, err := xraytest.ParseProxyURL(configURL)
		if err != nil {
			if callback != nil {
				callback.OnError(fmt.Sprintf("Invalid Config URL: %v", err))
			}
			return
		}
		passed := 0
		for index, candidate := range candidates {
			if ctx.Err() != nil {
				return
			}
			probe := template.WithEndpoint(candidate.IP.String(), candidate.Port)
			probe.AntiDPI = ui.CurrentAntiDPI()
			probe.SpeedURL = cfg.SpeedURL
			probe.SpeedSize = cfg.SpeedSize
			probe.UploadTest = cfg.UploadTest
			vr := mobileValidateConfig(ctx, probe, 22*time.Second)
			if vr.Success && cfg.MinSpeed > 0 && vr.Throughput*8/1_000_000 < cfg.MinSpeed {
				vr.Success = false
				vr.Error = fmt.Sprintf("speed below threshold (%.1f < %.1f Mbps)", vr.Throughput*8/1_000_000, cfg.MinSpeed)
			}
			done := index + 1
			if vr.Success {
				passed++
			}
			if callback != nil {
				callback.OnResult(candidate.IP.String(), candidate.Port, int(vr.Latency.Milliseconds()), 0, candidate.Colo, true, true, vr.Transport, vr.Throughput, vr.Success, vr.UploadThroughput)
				callback.OnProgress(done, passed, done-passed, total-done, true)
			}
		}
		return
	}

	timeout := 10 * time.Second
	if cfg.TimeoutType == "Custom" {
		if ms, err := strconv.Atoi(cfg.CustomTimeout); err == nil && ms > 10_000 {
			timeout = time.Duration(ms) * time.Millisecond
		}
	}
	sampleBytes := cfg.SpeedSize
	if sampleBytes <= 0 {
		sampleBytes = 512 * 1024
	}
	base := prober.Config{
		Mode: prober.ModeHTTP, Tries: 1, Timeout: timeout,
		SNI: "speed.cloudflare.com", SpeedBytes: sampleBytes,
		InsecureSkipVerify: true,
		AntiDPI:            ui.CurrentAntiDPI(),
	}
	passed := 0
	for index, candidate := range candidates {
		if ctx.Err() != nil {
			return
		}
		measured := prober.Probe(ctx, candidate.IP, base.WithPort(candidate.Port))
		success := measured.DownloadOK()
		if success {
			passed++
		}
		done := index + 1
		if callback != nil {
			callback.OnResult(candidate.IP.String(), candidate.Port, int(measured.Avg().Milliseconds()), measured.Loss(), candidate.Colo, true, true, "direct", measured.Throughput, success, 0.0)
			callback.OnProgress(done, passed, done-passed, total-done, true)
		}
	}
}

// GenerateConfigs returns a JSON export bundle that gomobile can bridge to
// Kotlin without exposing Go slices or structs in the public API.
func GenerateConfigs(configURL, endpointsText string) (string, error) {
	template, err := xraytest.ParseProxyURL(configURL)
	if err != nil {
		return "", fmt.Errorf("invalid config URL: %w", err)
	}
	rawEndpoints := strings.Fields(endpointsText)
	bundle, err := configexport.Generate(template, configexport.ParseEndpoints(rawEndpoints))
	if err != nil {
		return "", err
	}
	payload := struct {
		Subscription string   `json:"subscription"`
		ShareURLs    []string `json:"shareUrls"`
		SingBox      string   `json:"singBox"`
		Clash        string   `json:"clash"`
		Count        int      `json:"count"`
	}{bundle.Subscription, bundle.ShareURLs, bundle.SingBox, bundle.Clash, len(bundle.ShareURLs)}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// FetchMeta returns the shared multi-source ISP metadata as JSON for gomobile.
// The shared resolver combines Cloudflare, IPWhois and IPinfo and falls back to
// Team Cymru DNS, so Android and the desktop interfaces report the same result.
func FetchMeta() string {
	meta := ui.FetchMeta()
	encoded, err := json.Marshal(meta)
	if err != nil {
		return ""
	}
	return string(encoded)
}
