// Package scanjob runs a complete scan (Phase 1 reachability, Phase 2 tunnel or direct speed test) for any
// front end: the plain-text CLI and the desktop app both drive it and only differ in how they show Events.
package scanjob

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/ipsrc"
	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/ui"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

// The "Gentle" profile lives in ui so the terminal UI and every other front end share one definition.
const (
	GentleWorkers = ui.GentleWorkers
	GentleTimeout = ui.GentleTimeout
	GentleRate    = ui.GentleRate
)

// Params describes one scan. Zero values mean the usual defaults.
type Params struct {
	Count        int     `json:"count"`
	Workers      int     `json:"workers"`
	TimeoutMs    int     `json:"timeoutMs"`
	Ports        []int   `json:"ports"` // 0 = the config's port
	ConfigURL    string  `json:"configUrl"`
	RequireWS    bool    `json:"requireWS"`
	TopN         int     `json:"topN"`
	MinSpeed     float64 `json:"minSpeed"`
	SpeedURL     string  `json:"speedUrl"`
	SpeedSize    int64   `json:"speedSize"`
	UploadTest   bool    `json:"uploadTest"`
	NeighborScan bool    `json:"neighborScan"`

	Gentle     bool   `json:"gentle"`
	Targets    string `json:"targets"`    // pasted IPs / CIDRs / ranges / domains; empty = random Cloudflare pool
	Phase2Only bool   `json:"phase2Only"` // skip reachability: test the targets directly
	StatePath  string `json:"-"`          // where snapshots are written; empty disables them
	Resume     bool   `json:"-"`          // continue the scan saved at StatePath
}

// Events receives what happens during a run. All methods may be called from different goroutines.
type Events interface {
	Phase(phase int, livePath string)
	Stats(tested, healthy, total int)
	Results(batch []*result.Result)
	Validate(v *xraytest.ValidationResult, done, total int)
	Info(msg string)
	Error(msg string)
}

// Outcome is the final state of a run.
type Outcome struct {
	Cancelled bool
	Healthy   int
	Working   []string
	Phase1    []*result.Result
}

func (p Params) timing() (workers int, timeout time.Duration, ratePerSec float64) {
	workers = p.Workers
	if workers <= 0 {
		workers = 50
	}
	timeout = time.Duration(p.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if p.Gentle {
		workers = min(workers, GentleWorkers)
		timeout = max(timeout, GentleTimeout)
		ratePerSec = GentleRate
	}
	return
}

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
	if strings.TrimSpace(configURL) != "" && probePort <= 0 {
		return []int{443}
	}
	if probePort <= 0 {
		return []int{443}
	}
	return []int{probePort}
}

func key(ip net.IP, port int) string { return net.JoinHostPort(ip.String(), strconv.Itoa(port)) }

type run struct {
	p       Params
	ev      Events
	ports   []int
	pool    []net.IP
	mu      sync.Mutex
	done    map[string]bool
	probed  int
	healthy []*result.Result
	valid   []*xraytest.ValidationResult
	phase   int
}

// Run executes the scan described by p. It returns when the scan is finished or ctx is cancelled.
func Run(ctx context.Context, p Params, ev Events) Outcome {
	r := &run{p: p, ev: ev, done: map[string]bool{}, phase: 1}
	configURL := strings.TrimSpace(p.ConfigURL)

	if p.Resume {
		snap, err := LoadSnapshot(p.StatePath)
		if err != nil {
			ev.Error(err.Error())
			return Outcome{}
		}
		sp := snap.Params
		sp.StatePath, sp.Resume = p.StatePath, true
		r.p, p = sp, sp
		configURL = strings.TrimSpace(p.ConfigURL)
		for _, s := range snap.Pool {
			r.pool = append(r.pool, net.ParseIP(s))
		}
		for _, k := range snap.Done {
			r.done[k] = true
		}
		r.probed = snap.Probed
		for _, h := range snap.Healthy {
			r.healthy = append(r.healthy, h.result())
		}
		for _, v := range snap.Validated {
			r.valid = append(r.valid, v.result())
		}
		r.phase = snap.Phase
		ev.Info(fmt.Sprintf("Resuming a scan saved %s: %d endpoints already probed, %d healthy.", snap.Saved.Format("2006-01-02 15:04"), len(snap.Done), len(snap.Healthy)))
	}

	workers, timeout, ratePerSec := p.timing()
	probeCfg, err := ui.Phase1ProbeConfig(configURL, timeout, p.RequireWS)
	if err != nil {
		ev.Error(fmt.Sprintf("invalid URL: %v", err))
		return Outcome{}
	}
	r.ports = resolvePorts(p.Ports, configURL, probeCfg.Port)

	var neighbor ui.NeighborScanOpts
	if !p.Resume {
		var ok bool
		if neighbor, ok = r.buildPool(ctx); !ok {
			return Outcome{}
		}
	} else if p.NeighborScan && strings.TrimSpace(p.Targets) == "" {
		if src, err := ipsrc.New(true, false, nil); err == nil {
			neighbor = ui.DefaultNeighborOpts(src.IPv4Nets())
		}
	}

	writer, livePath, _ := ui.NewLiveResultWriter(configURL != "")
	stopSave := r.autosave()
	defer stopSave()

	if p.Phase2Only {
		return r.phase2(ctx, writer, livePath, configURL, timeout, workers, r.directRows())
	}

	ev.Phase(1, livePath)
	total := len(r.pool) * len(r.ports)
	var batchMu sync.Mutex
	var pending []*result.Result
	var tested, healthy atomic.Int64
	tested.Store(int64(r.probed))
	healthy.Store(int64(len(r.healthy)))

	flush := func() {
		batchMu.Lock()
		batch := pending
		pending = nil
		batchMu.Unlock()
		t, h := int(tested.Load()), int(healthy.Load())
		ev.Stats(t, h, total)
		if len(batch) > 0 {
			ev.Results(batch)
		}
	}
	ticker := time.NewTicker(250 * time.Millisecond)
	stopTick := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				flush()
			case <-stopTick:
				ticker.Stop()
				return
			}
		}
	}()

	callback := func(res *result.Result) {
		if writer != nil {
			writer.AddPhase1(res)
		}
		r.mu.Lock()
		r.done[key(res.IP, res.Port)] = true
		r.probed++
		if res.IsHealthy() {
			r.healthy = append(r.healthy, res)
		}
		r.mu.Unlock()
		if res.IsHealthy() {
			healthy.Add(1)
		}
		tested.Add(1)
		batchMu.Lock()
		pending = append(pending, res)
		batchMu.Unlock()
	}

	if r.phase == 1 {
		ui.RunPortProbesLimited(ctx, r.remaining(ctx), r.ports, workers, probeCfg, callback, neighbor, ratePerSec)
	}
	close(stopTick)
	flush()

	r.mu.Lock()
	collected := append([]*result.Result(nil), r.healthy...)
	r.mu.Unlock()
	out := Outcome{Cancelled: ctx.Err() != nil, Healthy: len(collected), Phase1: collected, Working: []string{}}

	if configURL == "" || out.Cancelled {
		if writer != nil && configURL == "" {
			writer.FinishPhase1Only()
		}
		r.finish(out.Cancelled)
		return out
	}
	top := result.TopN(collected, p.TopN)
	if len(top) == 0 {
		r.finish(false)
		return out
	}
	phase2 := r.phase2(ctx, writer, livePath, configURL, timeout, workers, top)
	phase2.Phase1, phase2.Healthy = collected, len(collected)
	return phase2
}

// buildPool fills r.pool from the pasted targets or a random Cloudflare sample.
func (r *run) buildPool(ctx context.Context) (ui.NeighborScanOpts, bool) {
	var neighbor ui.NeighborScanOpts
	if strings.TrimSpace(r.p.Targets) != "" {
		t := ParseTargets(ctx, r.p.Targets)
		for _, s := range t.Skipped {
			r.ev.Info("Skipped: " + s)
		}
		if t.Domains > 0 {
			r.ev.Info(fmt.Sprintf("Resolved %d domain(s) to addresses.", t.Domains))
		}
		if len(t.IPs) == 0 {
			r.ev.Error("No usable IPs, CIDRs, ranges or domains were found in the targets.")
			return neighbor, false
		}
		r.pool = t.IPs
		return neighbor, true
	}
	count := r.p.Count
	if count <= 0 {
		count = 1000
	}
	src, err := ipsrc.New(true, false, nil)
	if err != nil {
		r.ev.Error(err.Error())
		return neighbor, false
	}
	r.pool = src.MahsaNGV4Sample(count)
	if r.p.NeighborScan {
		neighbor = ui.DefaultNeighborOpts(src.IPv4Nets())
	}
	return neighbor, true
}

// remaining streams the pool IPs that still have at least one port to probe.
func (r *run) remaining(ctx context.Context) <-chan net.IP {
	ch := make(chan net.IP)
	go func() {
		defer close(ch)
		for _, ip := range r.pool {
			r.mu.Lock()
			all := true
			for _, port := range r.ports {
				all = all && r.done[key(ip, port)]
			}
			r.mu.Unlock()
			if all {
				continue
			}
			select {
			case ch <- ip:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

// directRows turns the pool into Phase 2 candidates without probing them first.
func (r *run) directRows() []*result.Result {
	var rows []*result.Result
	for _, ip := range r.pool {
		for _, port := range r.ports {
			rows = append(rows, &result.Result{IP: ip, Port: port})
		}
	}
	return rows
}

// phase2 validates candidates through xray (config URL) or with a direct download sample (no config).
func (r *run) phase2(ctx context.Context, w *ui.LiveResultWriter, livePath, configURL string, timeout time.Duration, workers int, rows []*result.Result) Outcome {
	r.mu.Lock()
	r.phase = 2
	have := map[string]bool{}
	for _, v := range r.valid {
		have[net.JoinHostPort(v.IP, strconv.Itoa(v.Port))] = true
	}
	r.mu.Unlock()
	var todo []*result.Result
	for _, row := range rows {
		if !have[key(row.IP, row.Port)] {
			todo = append(todo, row)
		}
	}
	if w != nil {
		w.BeginPhase2()
	}
	r.ev.Phase(2, livePath)
	total := len(rows)
	var doneN atomic.Int64
	doneN.Store(int64(total - len(todo)))
	record := func(v *xraytest.ValidationResult) {
		if w != nil {
			w.AddPhase2(v)
		}
		r.mu.Lock()
		r.valid = append(r.valid, v)
		r.mu.Unlock()
		r.ev.Validate(v, int(doneN.Add(1)), total)
	}
	for _, v := range r.snapshotValidated() { // replay what the earlier session finished so the UI shows it
		r.ev.Validate(v, int(doneN.Load()), total)
	}

	if configURL != "" {
		xrayTimeout := ui.Phase2Timeout(timeout, r.p.MinSpeed, r.p.SpeedSize)
		err := ui.RunPhase2(ctx, configURL, todo, r.p.MinSpeed, r.p.SpeedURL, r.p.SpeedSize, xrayTimeout, r.p.UploadTest,
			func(v *xraytest.ValidationResult, _, _ int) { record(v) })
		if err != nil {
			r.ev.Error(err.Error())
		}
	} else {
		r.direct(ctx, todo, timeout, workers, record)
	}
	r.mu.Lock()
	working := ui.WorkingEndpoints(r.valid)
	r.mu.Unlock()
	cancelled := ctx.Err() != nil
	r.finish(cancelled)
	return Outcome{Cancelled: cancelled, Healthy: len(rows), Working: working}
}

func (r *run) snapshotValidated() []*xraytest.ValidationResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]*xraytest.ValidationResult(nil), r.valid...)
}

// direct measures one download sample per candidate (the "just test my list" path).
func (r *run) direct(ctx context.Context, rows []*result.Result, timeout time.Duration, workers int, record func(*xraytest.ValidationResult)) {
	timeout = max(timeout, 10*time.Second)
	sample := r.p.SpeedSize
	if sample <= 0 {
		sample = 512 * 1024
	}
	base := prober.Config{Mode: prober.ModeHTTP, Tries: 1, Timeout: timeout, SNI: "speed.cloudflare.com", SpeedBytes: sample,
		InsecureSkipVerify: true, AntiDPI: ui.CurrentAntiDPI()}
	workers = min(max(workers, 1), 20)
	jobs := make(chan *result.Result)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range jobs {
				if ctx.Err() != nil {
					return
				}
				m := prober.Probe(ctx, c.IP, base.WithPort(c.Port))
				ok := m.DownloadOK()
				v := &xraytest.ValidationResult{IP: c.IP.String(), Port: c.Port, Transport: "direct", Success: ok,
					Latency: m.Avg(), Throughput: m.Throughput}
				if !ok {
					v.Error = "download sample failed"
				} else if r.p.MinSpeed > 0 && m.Throughput*8/1e6 < r.p.MinSpeed {
					v.Success, v.Error = false, fmt.Sprintf("speed below threshold (%.1f < %.1f Mbps)", m.Throughput*8/1e6, r.p.MinSpeed)
				}
				record(v)
			}
		}()
	}
loop:
	for _, c := range rows {
		select {
		case jobs <- c:
		case <-ctx.Done():
			break loop
		}
	}
	close(jobs)
	wg.Wait()
}

// ---- snapshots -------------------------------------------------------------------------------

func (r *run) snapshot() *Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	sp := r.p
	sp.Ports = r.ports
	sp.Targets = ""
	if strings.TrimSpace(r.p.Targets) != "" {
		sp.Targets = "resumed" // the pool is stored in the snapshot; this only keeps "pasted targets" mode on resume
	}
	s := &Snapshot{Params: sp, Probed: r.probed, Phase: r.phase}
	for _, ip := range r.pool {
		s.Pool = append(s.Pool, ip.String())
	}
	for k := range r.done {
		s.Done = append(s.Done, k)
	}
	for _, h := range r.healthy {
		s.Healthy = append(s.Healthy, rowOf(h))
	}
	for _, v := range r.valid {
		s.Validated = append(s.Validated, valRowOf(v))
	}
	return s
}

func (r *run) save() {
	if r.p.StatePath == "" || len(r.pool) == 0 {
		return
	}
	if err := SaveSnapshot(r.p.StatePath, r.snapshot()); err != nil {
		r.ev.Error("could not save the scan state: " + err.Error())
	}
}

// autosave writes a snapshot every 20 seconds and returns a stop function that writes the last one.
func (r *run) autosave() func() {
	if r.p.StatePath == "" {
		return func() {}
	}
	stop := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				r.save()
			case <-stop:
				return
			}
		}
	}()
	return func() { close(stop); <-finished }
}

// finish keeps the snapshot when the scan was interrupted and removes it when the scan completed.
func (r *run) finish(cancelled bool) {
	if r.p.StatePath == "" {
		return
	}
	if cancelled {
		r.save()
		return
	}
	_ = removeFile(r.p.StatePath)
}
