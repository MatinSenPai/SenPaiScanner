package ui

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/matinsenpai/senpaiscanner/internal/engine"
	"github.com/matinsenpai/senpaiscanner/internal/ipsrc"
	"github.com/matinsenpai/senpaiscanner/internal/output"
	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/store"
)

// statsInterval is how often live engine counters are pushed to the TUI.
// Pushing on every probe would flood the Bubble Tea event loop on large scans;
// a steady cadence keeps the display smooth at a fraction of the cost.
const statsInterval = 150 * time.Millisecond

// activeScan tracks the cancel function of the currently running scan. It is
// written from scan goroutines and read from the TUI goroutine, so all access
// goes through the mutex.
var activeScan struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

// setActiveScan registers cancel as the active scan, cancelling any previous
// scan that is still running so two scans never compete.
func setActiveScan(cancel context.CancelFunc) {
	activeScan.mu.Lock()
	defer activeScan.mu.Unlock()
	if activeScan.cancel != nil {
		activeScan.cancel()
	}
	activeScan.cancel = cancel
}

// cancelActiveScan cancels the running scan, if any.
func cancelActiveScan() {
	activeScan.mu.Lock()
	defer activeScan.mu.Unlock()
	if activeScan.cancel != nil {
		activeScan.cancel()
		activeScan.cancel = nil
	}
}

// StartScanCmd builds a tea.Cmd that runs the scan engine in the background,
// sending ResultMsg and StatsMsg messages to the Bubble Tea program.
func StartScanCmd(cfg ScanConfig) tea.Cmd {
	return func() tea.Msg {
		go runScan(cfg)
		return nil
	}
}

// CancelScanCmd cancels the running scan.
func CancelScanCmd() tea.Cmd {
	return func() tea.Msg {
		cancelActiveScan()
		return nil
	}
}

// StartTestCmd runs the test pass against a file of IPs.
func StartTestCmd(ipFile string) tea.Cmd {
	return func() tea.Msg {
		go runTest(ipFile)
		return nil
	}
}

// StartColosCmd discovers accessible Cloudflare PoPs.
func StartColosCmd() tea.Cmd {
	return func() tea.Msg {
		go runColos()
		return nil
	}
}

// StartUpdateRangesCmd fetches Cloudflare's current published CIDRs and caches
// them on disk so subsequent scans pick them up, reporting a RangesUpdatedMsg.
func StartUpdateRangesCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		v4, v6, err := ipsrc.UpdateRanges(ctx)
		if err != nil {
			return RangesUpdatedMsg{Err: err}
		}
		if err := ipsrc.SaveRanges(v4, v6); err != nil {
			return RangesUpdatedMsg{Err: err}
		}
		return RangesUpdatedMsg{V4: len(v4), V6: len(v6)}
	}
}

// prog is set by main before launching the Bubble Tea program so the
// background goroutines can send messages back.
var prog *tea.Program

// SetProgram must be called before any scan command is started.
func SetProgram(p *tea.Program) { prog = p }

// ---------------------------------------------------------------------------
// Background runners
// ---------------------------------------------------------------------------

func runScan(cfg ScanConfig) {
	count, _ := strconv.Atoi(cfg.Count)
	concurrency, _ := strconv.Atoi(cfg.Concurrency)
	if concurrency <= 0 {
		concurrency = 50
	}
	timeout := parseTimeout(cfg.Timeout, 3*time.Second)
	tries, _ := strconv.Atoi(cfg.Tries)
	if tries <= 0 {
		tries = 4
	}
	port, _ := strconv.Atoi(cfg.Port)
	if port <= 0 {
		port = 443
	}

	mode, err := prober.ParseMode(cfg.Mode)
	if err != nil {
		mode = prober.ModeTLS
	}

	var extra []string
	for _, c := range strings.Split(cfg.CIDR, ",") {
		c = strings.TrimSpace(c)
		if c != "" {
			extra = append(extra, c)
		}
	}

	src, err := ipsrc.New(cfg.UseV4, cfg.UseV6, extra)
	if err != nil {
		finishScan()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	setActiveScan(cancel)
	defer cancel()

	leaderboard, _ := store.Open(store.DefaultPath())
	if leaderboard != nil {
		defer func() { _ = leaderboard.Close() }()
	}

	eng := engine.New(engine.Config{
		Concurrency: concurrency,
		ProbeConfig: prober.Config{
			Port:          port,
			Mode:          mode,
			Tries:         tries,
			Timeout:       timeout,
			SNI:           cfg.SNI,
			SpeedBytes:    speedSampleForMode(mode),
			SpeedStreams:  3,
			SpeedDuration: minDuration(timeout, 3*time.Second),
			Jitter:        jitterForConfig(cfg),
		},
	})

	coloSet := buildColoSet(cfg.ColoFilter)

	var writer *output.Writer
	if cfg.OutputFile != "" {
		if w, e := output.New(cfg.OutputFile, output.DetectFormat(cfg.OutputFile)); e == nil {
			writer = w
			defer func() { _ = writer.Close() }()
		}
	}

	go streamStats(ctx, eng)

	warmIPs := warmStartIPs(ctx, leaderboard, cfg, count, concurrency)
	ipStream := scanIPStream(ctx, src, count, warmIPs)
	eng.Run(ctx, ipStream, func(r *result.Result) {
		if leaderboard != nil {
			_ = leaderboard.SaveResult(ctx, r)
		}
		if !r.IsHealthy() || !passesColoFilter(r, coloSet) {
			return
		}
		// Only healthy, filter-matching IPs reach the TUI and the output file.
		// Failed probes are reflected in the counters, not streamed as rows —
		// this is what keeps a 100k-IP scan responsive.
		if writer != nil {
			_ = writer.Write(r)
		}
		sendResult(r)
	})

	pushStats(eng) // final exact counters before signalling completion
	finishScan()
}

func runTest(ipFile string) {
	ips, err := loadIPs(ipFile)
	if err != nil || len(ips) == 0 {
		finishScan()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	setActiveScan(cancel)
	defer cancel()

	eng := engine.New(engine.Config{
		Concurrency: 20,
		ProbeConfig: prober.Config{
			Port:          443,
			Mode:          prober.ModeHTTP,
			Tries:         6,
			Timeout:       10 * time.Second,
			SNI:           "speed.cloudflare.com",
			SpeedBytes:    512 * 1024,
			SpeedStreams:  3,
			SpeedDuration: 4 * time.Second,
			Jitter:        prober.StealthJitter, // deep validation: accuracy over speed
		},
	})

	go streamStats(ctx, eng)

	// The Test flow validates a small, user-supplied list, so every result
	// (including failures) is shown — the user wants to see which candidates
	// passed and which did not.
	eng.RunList(ctx, ips, func(r *result.Result) {
		sendResult(r)
	})

	pushStats(eng) // final exact counters before signalling completion
	finishScan()
}

func runColos() {
	src, err := ipsrc.New(true, false, nil)
	if err != nil {
		finishColos()
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	setActiveScan(cancel)
	defer cancel()

	eng := engine.New(engine.Config{
		Concurrency: 80,
		ProbeConfig: prober.Config{
			Port:    443,
			Mode:    prober.ModeHTTP,
			Tries:   2,
			Timeout: 5 * time.Second,
		},
	})

	go streamStats(ctx, eng)

	ipStream := src.Stream(ctx, 300)
	eng.Run(ctx, ipStream, func(r *result.Result) {
		if r.IsHealthy() && r.Colo != "" {
			sendResult(r)
		}
	})

	pushStats(eng)
	finishColos()
}

// ---------------------------------------------------------------------------
// TUI messaging helpers
// ---------------------------------------------------------------------------

func sendResult(r *result.Result) {
	if prog != nil {
		prog.Send(ResultMsg(r))
	}
}

func finishScan() {
	if prog != nil {
		prog.Send(DoneMsg{})
	}
}

func finishColos() {
	if prog != nil {
		prog.Send(ColosDoneMsg{})
	}
}

// pushStats sends one snapshot of the engine counters to the TUI.
func pushStats(eng *engine.Engine) {
	if prog == nil {
		return
	}
	s := eng.Stats()
	prog.Send(StatsMsg{
		Tested:   s.Tested.Load(),
		Healthy:  s.Healthy.Load(),
		Failed:   s.Failed.Load(),
		InFlight: s.InFlight.Load(),
	})
}

// streamStats pushes counter snapshots at a fixed cadence until ctx is done.
func streamStats(ctx context.Context, eng *engine.Engine) {
	t := time.NewTicker(statsInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			pushStats(eng)
		}
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func buildColoSet(raw string) map[string]bool {
	if raw == "" {
		return nil
	}
	set := make(map[string]bool)
	for _, c := range strings.Split(raw, ",") {
		c = strings.TrimSpace(strings.ToUpper(c))
		if c != "" {
			set[c] = true
		}
	}
	return set
}

func passesColoFilter(r *result.Result, set map[string]bool) bool {
	if set == nil {
		return true
	}
	return set[strings.ToUpper(r.Colo)]
}

func loadIPs(path string) ([]net.IP, error) {
	var f *os.File
	var err error
	if path == "" || path == "-" {
		f = os.Stdin
	} else {
		f, err = os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", path, err)
		}
		defer func() { _ = f.Close() }()
	}
	var ips []net.IP
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "ip") {
			continue
		}
		field := strings.SplitN(line, ",", 2)[0]
		if ip := net.ParseIP(strings.TrimSpace(field)); ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips, sc.Err()
}

func speedSampleForMode(mode prober.Mode) int64 {
	if mode != prober.ModeHTTP {
		return 0
	}
	// A larger sample is now safe because prober caps download time and can
	// split the sample over a few streams. This makes throughput meaningful
	// without letting speed checks dominate the scan.
	return 512 * 1024
}

func jitterForConfig(cfg ScanConfig) time.Duration {
	if cfg.Stealth {
		return prober.StealthJitter
	}
	return 0
}

func warmStartIPs(ctx context.Context, leaderboard *store.DB, cfg ScanConfig, count, concurrency int) []net.IP {
	if leaderboard == nil || strings.TrimSpace(cfg.CIDR) != "" {
		return nil
	}
	limit := max(100, concurrency*2)
	if count > 0 && limit > count {
		limit = count
	}
	ips, err := leaderboard.TopIPs(ctx, limit)
	if err != nil {
		return nil
	}
	out := ips[:0]
	for _, ip := range ips {
		if ip == nil {
			continue
		}
		if ip.To4() != nil && cfg.UseV4 {
			out = append(out, ip)
			continue
		}
		if ip.To4() == nil && cfg.UseV6 {
			out = append(out, ip)
		}
	}
	return out
}

func scanIPStream(ctx context.Context, src *ipsrc.Source, count int, warm []net.IP) <-chan net.IP {
	ch := make(chan net.IP, 256)
	go func() {
		defer close(ch)
		seen := make(map[string]struct{}, len(warm))
		sent := 0

		send := func(ip net.IP) bool {
			if ip == nil {
				return true
			}
			if count > 0 && sent >= count {
				return false
			}
			key := string(ip)
			if _, ok := seen[key]; ok {
				return true
			}
			seen[key] = struct{}{}
			select {
			case <-ctx.Done():
				return false
			case ch <- ip:
				sent++
				return true
			}
		}

		for _, ip := range warm {
			if !send(ip) {
				return
			}
		}

		for ip := range src.Stream(ctx, 0) {
			if !send(ip) {
				return
			}
		}
	}()
	return ch
}

func parseTimeout(raw string, fallback time.Duration) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	if timeout, err := time.ParseDuration(raw); err == nil {
		return timeout
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
