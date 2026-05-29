package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/engine"
	"github.com/matinsenpai/senpaiscanner/internal/ipsrc"
	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/store"
)

//go:embed assets/*
var assetFS embed.FS

// Serve starts the embedded Web UI and blocks until the process is interrupted
// (Ctrl-C / SIGTERM) or the server fails. It binds the listener up front so a
// port clash is reported immediately instead of being swallowed inside
// ListenAndServe, and shuts down gracefully so in-flight requests can drain.
func Serve(addr, version string) error {
	if addr == "" {
		addr = "127.0.0.1:8787"
	}
	db, err := store.Open(store.DefaultPath())
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	s := &Server{version: version, leaderboard: db}
	srv := &http.Server{
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second, // mitigate Slowloris-style header stalls
		IdleTimeout:       60 * time.Second,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("bind %s: %w", addr, err)
	}
	fmt.Printf("SenPai Scanner web UI: http://%s\n", ln.Addr())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		stop() // restore default signal handling: a second Ctrl-C force-quits
		fmt.Println("\nshutting down web UI…")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// Server owns the web scan state.
type Server struct {
	version     string
	leaderboard *store.DB

	mu      sync.Mutex
	nextID  int64
	current *scanState
}

type scanState struct {
	ID      int64        `json:"id"`
	Started time.Time    `json:"started"`
	Done    bool         `json:"done"`
	Error   string       `json:"error,omitempty"`
	Stats   webStats     `json:"stats"`
	Results []resultView `json:"results"`
	cancel  context.CancelFunc
}

type scanSnapshot struct {
	ID      int64        `json:"id"`
	Started time.Time    `json:"started"`
	Done    bool         `json:"done"`
	Error   string       `json:"error,omitempty"`
	Stats   webStats     `json:"stats"`
	Results []resultView `json:"results"`
}

type webStats struct {
	Tested   int64 `json:"tested"`
	Healthy  int64 `json:"healthy"`
	Failed   int64 `json:"failed"`
	InFlight int64 `json:"in_flight"`
}

type scanRequest struct {
	Count       int    `json:"count"`
	Concurrency int    `json:"concurrency"`
	Timeout     string `json:"timeout"`
	Tries       int    `json:"tries"`
	Port        int    `json:"port"`
	Mode        string `json:"mode"`
	CIDR        string `json:"cidr"`
	SNI         string `json:"sni"`
	UseV4       bool   `json:"use_v4"`
	UseV6       bool   `json:"use_v6"`
	Stealth     bool   `json:"stealth"`
	Adaptive    bool   `json:"adaptive"` // EWMA-driven per-probe timeout
	Fast        bool   `json:"fast"`     // two-phase pipeline + dead-prefix gate
}

type resultView struct {
	IP            string  `json:"ip"`
	Score         float64 `json:"score"`
	LossPct       float64 `json:"loss_pct"`
	AvgMs         float64 `json:"avg_ms"`
	JitterMs      float64 `json:"jitter_ms"`
	ThroughputKB  float64 `json:"throughput_kb"`
	ThroughputBPS float64 `json:"throughput_bps"`
	Colo          string  `json:"colo"`
	TLSOk         bool    `json:"tls_ok"`
	HTTPStatus    int     `json:"http_status"`
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	assets, _ := fs.Sub(assetFS, "assets")
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/version", s.handleVersion)
	mux.HandleFunc("/api/ports", s.handlePorts)
	mux.HandleFunc("/api/leaderboard", s.handleLeaderboard)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/scan", s.handleScan)
	mux.HandleFunc("/api/cancel", s.handleCancel)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFileFS(w, r, assetFS, "assets/index.html")
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"version": s.version})
}

// handlePorts returns Cloudflare's proxied ports so the UI can offer a curated
// picker instead of a free-form number field.
func (s *Server) handlePorts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string][]int{
		"https": prober.HTTPSPorts,
		"http":  prober.HTTPPorts,
	})
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	recs, err := s.leaderboard.TopRecords(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, recs)
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.current == nil {
		s.mu.Unlock()
		writeJSON(w, map[string]any{"running": false})
		return
	}
	snap := snapshotScanState(s.current)
	s.mu.Unlock()
	writeJSON(w, snap)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req scanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req = normalizeRequest(req)

	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	if s.current != nil && !s.current.Done && s.current.cancel != nil {
		s.current.cancel()
	}
	s.nextID++
	state := &scanState{ID: s.nextID, Started: time.Now(), cancel: cancel}
	s.current = state
	snap := snapshotScanState(state)
	s.mu.Unlock()

	go s.runScan(ctx, state, req)
	writeJSON(w, snap)
}

func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.current != nil && !s.current.Done && s.current.cancel != nil {
		s.current.cancel()
	}
	writeJSON(w, map[string]bool{"ok": true})
}

func (s *Server) runScan(ctx context.Context, state *scanState, req scanRequest) {
	mode, err := prober.ParseMode(req.Mode)
	if err != nil {
		mode = prober.ModeHTTP
	}
	timeout := parseDuration(req.Timeout, 3*time.Second)
	extra := splitCIDRs(req.CIDR)
	src, err := ipsrc.New(req.UseV4, req.UseV6, extra)
	if err != nil {
		s.finish(state, err)
		return
	}

	cfg := prober.Config{
		Port:          req.Port,
		Mode:          mode,
		Tries:         req.Tries,
		Timeout:       timeout,
		SNI:           req.SNI,
		SpeedBytes:    speedSampleForMode(mode),
		SpeedStreams:  3,
		SpeedDuration: minDuration(timeout, 3*time.Second),
	}
	if req.Stealth {
		cfg.Jitter = prober.StealthJitter
	}

	warm := s.warmStart(ctx, req)
	stream := scanIPStream(ctx, src, req.Count, warm)

	// onResult records a probe outcome: it persists every result to the
	// leaderboard and, for healthy IPs, keeps the live top-500 board sorted.
	// Stats are refreshed on a ticker (see streamStats) rather than here, so the
	// UI stays live even in fast mode where this fires only for phase-2 survivors.
	onResult := func(r *result.Result) {
		_ = s.leaderboard.SaveResult(ctx, r)
		if !r.IsHealthy() {
			return
		}
		s.mu.Lock()
		state.Results = append(state.Results, viewResult(r))
		sort.Slice(state.Results, func(i, j int) bool {
			return state.Results[i].Score > state.Results[j].Score
		})
		if len(state.Results) > 500 {
			state.Results = state.Results[:500]
		}
		s.mu.Unlock()
	}

	if req.Fast {
		// Phase 1 is a cheap TCP reachability ping at the target port, run at the
		// user's full worker count for maximum coverage; only IPs that connect are
		// promoted to phase 2 for the (expensive) configured validation.
		p1cfg := prober.Config{Port: req.Port, Mode: prober.ModeTCP, Tries: 1, Timeout: minDuration(timeout, time.Second)}
		phase1 := engine.New(engine.Config{Concurrency: req.Concurrency, ProbeConfig: p1cfg, AdaptiveTimeout: req.Adaptive})
		phase2 := engine.New(engine.Config{Concurrency: req.Concurrency, ProbeConfig: cfg})

		stop := s.streamStats(ctx, state, func() webStats { return combinedStats(phase1, phase2) })
		engine.RunTwoPhaseFast(ctx, stream, phase1, phase2, nil, onResult)
		stop()
		s.finishStats(state, combinedStats(phase1, phase2))
		return
	}

	eng := engine.New(engine.Config{Concurrency: req.Concurrency, ProbeConfig: cfg, AdaptiveTimeout: req.Adaptive})
	stop := s.streamStats(ctx, state, func() webStats { return snapshotStats(eng) })
	eng.Run(ctx, stream, onResult)
	stop()
	s.finishStats(state, snapshotStats(eng))
}

// streamStats periodically copies live engine counters into the scan state so
// /api/state reflects progress in near-real-time without coupling the refresh
// rate to how often results complete. The returned stop function ends the
// ticker; it is safe to call exactly once after the scan loop returns.
func (s *Server) streamStats(ctx context.Context, state *scanState, snap func() webStats) (stop func()) {
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(250 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case <-t.C:
				st := snap()
				s.mu.Lock()
				state.Stats = st
				s.mu.Unlock()
			}
		}
	}()
	return func() { close(done) }
}

// finishStats writes the final counters and marks the scan complete.
func (s *Server) finishStats(state *scanState, st webStats) {
	s.mu.Lock()
	state.Stats = st
	state.Done = true
	state.cancel = nil
	s.mu.Unlock()
}

// combinedStats projects a two-phase run onto the flat webStats the UI expects:
// Tested counts every IP that cleared phase 1 (the coverage number), Healthy
// counts phase-2 validated edges, and Failed is everything probed that did not
// validate. It is a live estimate while phase 2 drains and exact once it does.
func combinedStats(phase1, phase2 *engine.Engine) webStats {
	p1, p2 := phase1.Stats(), phase2.Stats()
	tested := p1.Tested.Load()
	healthy := p2.Healthy.Load()
	failed := tested - healthy
	if failed < 0 {
		failed = 0
	}
	return webStats{
		Tested:   tested,
		Healthy:  healthy,
		Failed:   failed,
		InFlight: p1.InFlight.Load() + p2.InFlight.Load(),
	}
}

func (s *Server) finish(state *scanState, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		state.Error = err.Error()
	}
	state.Done = true
	state.cancel = nil
}

func (s *Server) warmStart(ctx context.Context, req scanRequest) []net.IP {
	if strings.TrimSpace(req.CIDR) != "" {
		return nil
	}
	limit := max(100, req.Concurrency*2)
	if req.Count > 0 && limit > req.Count {
		limit = req.Count
	}
	ips, err := s.leaderboard.TopIPs(ctx, limit)
	if err != nil {
		return nil
	}
	out := ips[:0]
	for _, ip := range ips {
		if ip.To4() != nil && req.UseV4 {
			out = append(out, ip)
		}
		if ip.To4() == nil && req.UseV6 {
			out = append(out, ip)
		}
	}
	return out
}

// Upper bounds for client-supplied scan parameters. The web API is unauthenticated
// and can be bound to 0.0.0.0, so these clamps stop a hostile or fat-fingered
// request (e.g. concurrency=100000000) from exhausting memory by spawning a huge
// worker pool or allocating an enormous per-IP latency slice.
const (
	maxCount       = 2_000_000
	maxConcurrency = 2_000
	maxTries       = 20
)

func normalizeRequest(req scanRequest) scanRequest {
	if req.Count <= 0 {
		req.Count = 5000
	}
	if req.Count > maxCount {
		req.Count = maxCount
	}
	if req.Concurrency <= 0 {
		req.Concurrency = 100
	}
	if req.Concurrency > maxConcurrency {
		req.Concurrency = maxConcurrency
	}
	if req.Tries <= 0 {
		req.Tries = 4
	}
	if req.Tries > maxTries {
		req.Tries = maxTries
	}
	if req.Port <= 0 || req.Port > 65535 {
		req.Port = 443
	}
	if req.Timeout == "" {
		req.Timeout = "3s"
	}
	if req.Mode == "" {
		req.Mode = "http"
	}
	if !req.UseV4 && !req.UseV6 {
		req.UseV4 = true
	}
	return req
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

func snapshotScanState(state *scanState) scanSnapshot {
	results := make([]resultView, len(state.Results))
	copy(results, state.Results)
	return scanSnapshot{
		ID:      state.ID,
		Started: state.Started,
		Done:    state.Done,
		Error:   state.Error,
		Stats:   state.Stats,
		Results: results,
	}
}

func snapshotStats(eng *engine.Engine) webStats {
	st := eng.Stats()
	return webStats{
		Tested:   st.Tested.Load(),
		Healthy:  st.Healthy.Load(),
		Failed:   st.Failed.Load(),
		InFlight: st.InFlight.Load(),
	}
}

func viewResult(r *result.Result) resultView {
	return resultView{
		IP:            r.IP.String(),
		Score:         r.Score(),
		LossPct:       r.Loss(),
		AvgMs:         result.Ms(r.Avg()),
		JitterMs:      result.Ms(r.Jitter()),
		ThroughputKB:  r.Throughput / 1024,
		ThroughputBPS: r.Throughput,
		Colo:          r.Colo,
		TLSOk:         r.TLSOk,
		HTTPStatus:    r.HTTPStatus,
	}
}

func splitCIDRs(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseDuration(raw string, fallback time.Duration) time.Duration {
	if d, err := time.ParseDuration(strings.TrimSpace(raw)); err == nil {
		return d
	}
	return fallback
}

func speedSampleForMode(mode prober.Mode) int64 {
	if mode != prober.ModeHTTP {
		return 0
	}
	return 512 * 1024
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
