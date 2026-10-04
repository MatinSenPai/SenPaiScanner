package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/scanjob"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

// phase2Workers keeps tunnel validation light on phones (the desktop uses 10).
const phase2Workers = 3

var dataDir string

// SetDataDir tells the engine where the app may keep files (the resumable scan state). Call it once at start-up.
func SetDataDir(dir string) {
	mu.Lock()
	dataDir = dir
	mu.Unlock()
}

func statePath() string {
	mu.Lock()
	defer mu.Unlock()
	if dataDir == "" {
		return ""
	}
	return filepath.Join(dataDir, "scan-state.json")
}

// ResumeInfo describes the interrupted scan that can be continued, as JSON; "" when there is none.
func ResumeInfo() string {
	path := statePath()
	if path == "" {
		return ""
	}
	info, ok := scanjob.InspectSnapshot(path)
	if !ok || info.Total == 0 {
		return ""
	}
	b, _ := json.Marshal(info)
	return string(b)
}

// DiscardResume forgets the interrupted scan.
func DiscardResume() {
	if path := statePath(); path != "" {
		scanjob.DiscardSnapshot(path)
	}
}

// PreviewTargets parses pasted IPs / CIDRs / ranges / domains and returns {"count","domains","skipped"} as JSON.
func PreviewTargets(text string) string {
	t := scanjob.ParseTargets(context.Background(), text)
	skipped := t.Skipped
	if skipped == nil {
		skipped = []string{}
	}
	b, _ := json.Marshal(map[string]any{"count": len(t.IPs), "domains": t.Domains, "skipped": skipped})
	return string(b)
}

func parseInt(s string, def int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && n > 0 {
		return n
	}
	return def
}

// job maps the Android form onto the shared scan runner.
func (cfg ScanConfig) job() (scanjob.Params, error) {
	count := 5000
	if cfg.CountType == "Custom" {
		count = parseInt(cfg.CustomCount, 5000)
	} else {
		count = parseInt(cfg.CountType, 5000)
	}

	workers := 50
	switch {
	case cfg.WorkerType == "Custom":
		workers = parseInt(cfg.CustomWorkers, 50)
	case strings.HasPrefix(cfg.WorkerType, "100"):
		workers = 100
	case strings.HasPrefix(cfg.WorkerType, "200"):
		workers = 200
	}

	timeoutMs := 5000
	switch {
	case cfg.TimeoutType == "Custom":
		timeoutMs = parseInt(cfg.CustomTimeout, 5000)
	case strings.HasPrefix(cfg.TimeoutType, "2s"):
		timeoutMs = 2000
	case strings.HasPrefix(cfg.TimeoutType, "3s"):
		timeoutMs = 3000
	}

	configURL := strings.TrimSpace(cfg.ConfigURL)
	topN := 50
	switch {
	case cfg.TopNType == "Custom":
		topN = parseInt(cfg.CustomTopN, 50)
	case cfg.TopNType == "ALL":
		topN = 0
	default:
		topN = parseInt(cfg.TopNType, 50)
	}

	// With a config the port comes from it (as before); without one the selected ports are used.
	var ports []int
	if configURL == "" {
		ports = cfg.SelectedPorts
		if len(ports) == 0 {
			ports = []int{443}
		}
	}

	targets := ""
	switch {
	case cfg.SourceType == "Paste":
		targets = cfg.Targets
	case cfg.SourceType == "From File" && cfg.SourceFile != "":
		b, err := os.ReadFile(cfg.SourceFile)
		if err != nil {
			return scanjob.Params{}, fmt.Errorf("failed to load IPs: %w", err)
		}
		targets = string(b)
	}

	return scanjob.Params{
		Count: count, Workers: workers, TimeoutMs: timeoutMs, Ports: ports, ConfigURL: configURL,
		RequireWS: cfg.RequireWS, TopN: topN, MinSpeed: cfg.MinSpeed, SpeedURL: cfg.SpeedURL, SpeedSize: cfg.SpeedSize,
		UploadTest: cfg.UploadTest, NeighborScan: cfg.NeighborScan && configURL == "",
		Gentle: cfg.Gentle, Targets: targets, Phase2Only: cfg.Phase2Only && strings.TrimSpace(targets) != "",
		StatePath: statePath(), Resume: cfg.Resume,
		Validate: mobileValidateConfig, Phase2Workers: phase2Workers, NoLiveFile: true,
	}, nil
}

// events adapts scanjob events to the gomobile Callback the Kotlin side implements.
type events struct {
	cb Callback

	mu      sync.Mutex
	started bool
	colo    map[string]string
	passed  int
	failed  int
}

func (e *events) Phase(phase int, _ string) {
	e.mu.Lock()
	e.started = true
	e.passed, e.failed = 0, 0
	e.mu.Unlock()
}

func (e *events) Stats(tested, healthy, total int) {
	if e.cb != nil {
		e.cb.OnProgress(tested, healthy, tested-healthy, max(total-tested, 0), false)
	}
}

func (e *events) Results(batch []*result.Result) {
	for _, r := range batch {
		if !r.IsHealthy() {
			continue
		}
		e.mu.Lock()
		e.colo[net.JoinHostPort(r.IP.String(), strconv.Itoa(r.Port))] = r.Colo
		e.mu.Unlock()
		if e.cb != nil {
			e.cb.OnResult(r.IP.String(), r.Port, int(r.Avg().Milliseconds()), r.Loss(), r.Colo, true, false, "", 0, false, 0)
		}
	}
}

func (e *events) Validate(v *xraytest.ValidationResult, done, total int) {
	e.mu.Lock()
	if v.Success {
		e.passed++
	} else {
		e.failed++
	}
	passed, failed := e.passed, e.failed
	colo := e.colo[net.JoinHostPort(v.IP, strconv.Itoa(v.Port))]
	e.mu.Unlock()
	if e.cb == nil {
		return
	}
	e.cb.OnResult(v.IP, v.Port, int(v.Latency.Milliseconds()), 0, colo, true, true, v.Transport, v.Throughput, v.Success, v.UploadThroughput)
	e.cb.OnProgress(done, passed, failed, max(total-done, 0), true)
}

func (e *events) Info(msg string) {
	if e.cb != nil {
		e.cb.OnInfo(msg)
	}
}

// Error ends the scan when it happens before anything started (bad URL, nothing to scan); later problems are notices.
func (e *events) Error(msg string) {
	e.mu.Lock()
	started := e.started
	e.mu.Unlock()
	if e.cb == nil {
		return
	}
	if started {
		e.cb.OnInfo("Problem: " + msg)
		return
	}
	e.cb.OnError(msg)
}

func runScan(configJson string, callback Callback) {
	ctx, cancel := context.WithCancel(context.Background())
	mu.Lock()
	cancelScan = cancel
	mu.Unlock()
	defer func() {
		cancel()
		mu.Lock()
		isRunning = false
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
	params, err := cfg.job()
	if err != nil {
		if callback != nil {
			callback.OnError(err.Error())
		}
		return
	}

	out := scanjob.Run(ctx, params, &events{cb: callback, colo: map[string]string{}})
	mu.Lock()
	phase1Cache = append([]*result.Result(nil), out.Phase1...)
	mu.Unlock()
}
