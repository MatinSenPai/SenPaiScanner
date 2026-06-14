package webgui

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ehsatn/MAGIK/internal/ipsrc"
	"github.com/ehsatn/MAGIK/internal/prober"
	"github.com/ehsatn/MAGIK/internal/result"
	"github.com/ehsatn/MAGIK/internal/xraytest"
)

const maxResultRows = 600

// StartRequest is the JSON payload used by the browser UI to start a scan.
type StartRequest struct {
	Source        string  `json:"source"`
	Count         int     `json:"count"`
	Workers       int     `json:"workers"`
	Timeout       string  `json:"timeout"`
	Ports         []int   `json:"ports"`
	UseConfigPort bool    `json:"use_config_port"`
	ConfigURL     string  `json:"config_url"`
	TopN          int     `json:"top_n"`
	MinSpeedMbps  float64 `json:"min_speed_mbps"`
	SpeedSizeKB   int64   `json:"speed_size_kb"`
	UploadTest    bool    `json:"upload_test"`
	RequireWS     bool    `json:"require_ws"`
}

type PhaseStats struct {
	Done     int `json:"done"`
	Total    int `json:"total"`
	Healthy  int `json:"healthy"`
	Failed   int `json:"failed"`
	InFlight int `json:"in_flight"`
}

type ResultView struct {
	Phase      string  `json:"phase"`
	Endpoint   string  `json:"endpoint"`
	Colo       string  `json:"colo,omitempty"`
	LossPct    float64 `json:"loss_pct,omitempty"`
	AvgMS      float64 `json:"avg_ms,omitempty"`
	SpeedMbps  float64 `json:"speed_mbps,omitempty"`
	LatencyMS  float64 `json:"latency_ms,omitempty"`
	Transport  string  `json:"transport,omitempty"`
	Status     string  `json:"status"`
	Error      string  `json:"error,omitempty"`
	IsWorking  bool    `json:"is_working"`
	HTTPStatus int     `json:"http_status,omitempty"`
}

type EndpointView struct {
	Endpoint  string  `json:"endpoint"`
	IP        string  `json:"ip"`
	Port      int     `json:"port"`
	Phase     string  `json:"phase"`
	Colo      string  `json:"colo,omitempty"`
	AvgMS     float64 `json:"avg_ms,omitempty"`
	SpeedMbps float64 `json:"speed_mbps,omitempty"`
}

type ScanState struct {
	Running        bool           `json:"running"`
	Done           bool           `json:"done"`
	Phase          string         `json:"phase"`
	Status         string         `json:"status"`
	Error          string         `json:"error,omitempty"`
	StartedAt      string         `json:"started_at,omitempty"`
	FinishedAt     string         `json:"finished_at,omitempty"`
	ElapsedSeconds float64        `json:"elapsed_seconds"`
	Settings       StartRequest   `json:"settings"`
	Phase1         PhaseStats     `json:"phase1"`
	Phase2         PhaseStats     `json:"phase2"`
	Results        []ResultView   `json:"results"`
	Working        []EndpointView `json:"working"`
	SavePath       string         `json:"save_path,omitempty"`
}

type ScanManager struct {
	mu          sync.Mutex
	cancel      context.CancelFunc
	state       ScanState
	started     time.Time
	finished    time.Time
	phase1Raw   []*result.Result
	workingSeen map[string]struct{}
}

func NewScanManager() *ScanManager {
	return &ScanManager{
		state: ScanState{
			Phase:   "idle",
			Status:  "Ready",
			Results: []ResultView{},
			Working: []EndpointView{},
		},
		workingSeen: make(map[string]struct{}),
	}
}

func (m *ScanManager) Start(req StartRequest) error {
	req = normalizeStartRequest(req)

	m.mu.Lock()
	if m.state.Running {
		m.mu.Unlock()
		return fmt.Errorf("a scan is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.started = time.Now()
	m.finished = time.Time{}
	m.phase1Raw = nil
	m.workingSeen = make(map[string]struct{})
	m.state = ScanState{
		Running:   true,
		Done:      false,
		Phase:     "phase1",
		Status:    "Preparing Phase 1",
		StartedAt: m.started.Format(time.RFC3339),
		Settings:  req,
		Results:   []ResultView{},
		Working:   []EndpointView{},
	}
	m.mu.Unlock()

	go m.run(ctx, req)
	return nil
}

func (m *ScanManager) Stop() {
	m.mu.Lock()
	cancel := m.cancel
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (m *ScanManager) State() ScanState {
	m.mu.Lock()
	defer m.mu.Unlock()

	st := m.state
	if !m.started.IsZero() {
		end := time.Now()
		if !m.finished.IsZero() {
			end = m.finished
		}
		st.ElapsedSeconds = end.Sub(m.started).Seconds()
	}
	st.Results = append([]ResultView(nil), m.state.Results...)
	st.Working = append([]EndpointView(nil), m.state.Working...)
	if st.Results == nil {
		st.Results = []ResultView{}
	}
	if st.Working == nil {
		st.Working = []EndpointView{}
	}
	return st
}

func (m *ScanManager) SaveWorking() (string, int, error) {
	m.mu.Lock()
	endpoints := make([]string, 0, len(m.state.Working))
	for _, ep := range m.state.Working {
		endpoints = append(endpoints, ep.Endpoint)
	}
	m.mu.Unlock()

	if len(endpoints) == 0 {
		return "", 0, fmt.Errorf("no working endpoints to save")
	}
	path, err := writeWorkingEndpoints(endpoints)
	if err != nil {
		return path, len(endpoints), err
	}

	m.mu.Lock()
	m.state.SavePath = path
	m.state.Status = fmt.Sprintf("Saved %d endpoints", len(endpoints))
	m.mu.Unlock()
	return path, len(endpoints), nil
}

func normalizeStartRequest(req StartRequest) StartRequest {
	req.Source = strings.ToLower(strings.TrimSpace(req.Source))
	if req.Source != "file" {
		req.Source = "random"
	}
	if req.Count <= 0 {
		req.Count = 1000
	}
	if req.Count > 500000 {
		req.Count = 500000
	}
	if req.Workers <= 0 {
		req.Workers = 50
	}
	if req.Workers > 1000 {
		req.Workers = 1000
	}
	req.Timeout = strings.TrimSpace(req.Timeout)
	if req.Timeout == "" {
		req.Timeout = "5s"
	}
	req.ConfigURL = strings.TrimSpace(req.ConfigURL)
	if req.TopN <= 0 {
		req.TopN = 50
	}
	if req.SpeedSizeKB <= 0 {
		req.SpeedSizeKB = 512
	}
	return req
}

func (m *ScanManager) run(ctx context.Context, req StartRequest) {
	timeout := parseTimeout(req.Timeout, 5*time.Second)
	probeCfg, proxyCfg, err := buildProbeConfig(req.ConfigURL, timeout)
	if err != nil {
		m.fail(err)
		return
	}
	probeCfg.RequireWebSocket = req.RequireWS

	ports := resolvePorts(req, proxyCfg)
	ipStream, ipTotal, err := sourceIPs(ctx, req)
	if err != nil {
		m.fail(err)
		return
	}

	m.mu.Lock()
	m.state.Phase = "phase1"
	m.state.Status = "Scanning Cloudflare endpoints"
	m.state.Phase1.Total = ipTotal * len(ports)
	m.mu.Unlock()

	m.runPhase1(ctx, req, probeCfg, ports, ipStream)
	if ctx.Err() != nil {
		m.stopped()
		return
	}

	top := m.topPhase1(req.TopN)
	if proxyCfg == nil {
		m.done("Phase 1 complete")
		return
	}
	if len(top) == 0 {
		m.done("No healthy Phase 1 candidates found")
		return
	}

	m.mu.Lock()
	m.state.Phase = "phase2"
	m.state.Status = "Validating candidates with Xray"
	m.state.Phase2.Total = len(top)
	m.mu.Unlock()

	m.runPhase2(ctx, req, proxyCfg, top, timeout)
	if ctx.Err() != nil {
		m.stopped()
		return
	}
	m.done("Scan complete")
}

func (m *ScanManager) runPhase1(ctx context.Context, req StartRequest, base prober.Config, ports []int, ips <-chan net.IP) {
	type job struct {
		ip   net.IP
		port int
	}

	jobs := make(chan job, req.Workers*2)
	var wg sync.WaitGroup
	for i := 0; i < req.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if ctx.Err() != nil {
					continue
				}
				m.addPhase1InFlight(1)
				r := prober.Probe(ctx, j.ip, base.WithPort(j.port))
				m.addPhase1InFlight(-1)
				m.recordPhase1(r)
			}
		}()
	}

produce:
	for ip := range ips {
		for _, port := range ports {
			select {
			case jobs <- job{ip: cloneIP(ip), port: port}:
			case <-ctx.Done():
				break produce
			}
		}
	}
	close(jobs)
	wg.Wait()
}

func (m *ScanManager) runPhase2(ctx context.Context, req StartRequest, cfg *xraytest.VLESSConfig, candidates []*result.Result, phaseTimeout time.Duration) {
	workers := req.Workers
	if workers > 10 {
		workers = 10
	}
	if workers <= 0 {
		workers = 10
	}

	timeout := phase2Timeout(phaseTimeout, req.SpeedSizeKB*1024, req.MinSpeedMbps)
	jobs := make(chan *result.Result, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := range jobs {
				if ctx.Err() != nil {
					continue
				}
				m.addPhase2InFlight(1)
				swapped := cfg.WithEndpoint(r.IP.String(), r.Port)
				swapped.SpeedSize = req.SpeedSizeKB * 1024
				swapped.UploadTest = req.UploadTest
				vr := xraytest.ValidateConfig(ctx, swapped, timeout)
				if vr.Success && req.MinSpeedMbps > 0 {
					mbps := vr.Throughput * 8 / 1_000_000
					if mbps < req.MinSpeedMbps {
						vr.Success = false
						vr.Error = fmt.Sprintf("speed below threshold %.1f Mbps", req.MinSpeedMbps)
					}
				}
				m.addPhase2InFlight(-1)
				m.recordPhase2(vr)
			}
		}()
	}

produce:
	for _, r := range candidates {
		select {
		case jobs <- r:
		case <-ctx.Done():
			break produce
		}
	}
	close(jobs)
	wg.Wait()
}

func (m *ScanManager) addPhase1InFlight(delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Phase1.InFlight += delta
	if m.state.Phase1.InFlight < 0 {
		m.state.Phase1.InFlight = 0
	}
}

func (m *ScanManager) addPhase2InFlight(delta int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Phase2.InFlight += delta
	if m.state.Phase2.InFlight < 0 {
		m.state.Phase2.InFlight = 0
	}
}

func (m *ScanManager) recordPhase1(r *result.Result) {
	if r == nil {
		return
	}
	view := phase1ResultView(r)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.phase1Raw = append(m.phase1Raw, r)
	m.state.Phase1.Done++
	if r.IsHealthy() {
		m.state.Phase1.Healthy++
		if strings.TrimSpace(m.state.Settings.ConfigURL) == "" {
			m.addWorkingLocked(endpointViewFromPhase1(r))
		}
	} else {
		m.state.Phase1.Failed++
	}
	m.appendResultLocked(view)
}

func (m *ScanManager) recordPhase2(r *xraytest.ValidationResult) {
	if r == nil {
		return
	}
	view := phase2ResultView(r)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Phase2.Done++
	if r.Success {
		m.state.Phase2.Healthy++
		m.addWorkingLocked(endpointViewFromPhase2(r))
	} else {
		m.state.Phase2.Failed++
	}
	m.appendResultLocked(view)
}

func (m *ScanManager) appendResultLocked(view ResultView) {
	m.state.Results = append(m.state.Results, view)
	if len(m.state.Results) > maxResultRows {
		m.state.Results = m.state.Results[len(m.state.Results)-maxResultRows:]
	}
}

func (m *ScanManager) addWorkingLocked(ep EndpointView) {
	if ep.Endpoint == "" {
		return
	}
	if _, ok := m.workingSeen[ep.Endpoint]; ok {
		return
	}
	m.workingSeen[ep.Endpoint] = struct{}{}
	m.state.Working = append(m.state.Working, ep)
	sort.SliceStable(m.state.Working, func(i, j int) bool {
		if m.state.Working[i].SpeedMbps != m.state.Working[j].SpeedMbps {
			return m.state.Working[i].SpeedMbps > m.state.Working[j].SpeedMbps
		}
		return m.state.Working[i].AvgMS < m.state.Working[j].AvgMS
	})
}

func (m *ScanManager) topPhase1(n int) []*result.Result {
	m.mu.Lock()
	rows := append([]*result.Result(nil), m.phase1Raw...)
	m.mu.Unlock()
	return result.TopN(rows, n)
}

func (m *ScanManager) fail(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.finished = time.Now()
	m.cancel = nil
	m.state.Running = false
	m.state.Done = true
	m.state.Phase = "error"
	m.state.Status = "Scan failed"
	m.state.Error = err.Error()
	m.state.FinishedAt = m.finished.Format(time.RFC3339)
}

func (m *ScanManager) stopped() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.finished = time.Now()
	m.cancel = nil
	m.state.Running = false
	m.state.Done = true
	m.state.Phase = "stopped"
	m.state.Status = "Scan stopped"
	m.state.FinishedAt = m.finished.Format(time.RFC3339)
}

func (m *ScanManager) done(status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.finished = time.Now()
	m.cancel = nil
	m.state.Running = false
	m.state.Done = true
	m.state.Phase = "done"
	m.state.Status = status
	m.state.FinishedAt = m.finished.Format(time.RFC3339)
}

func buildProbeConfig(rawURL string, timeout time.Duration) (prober.Config, *xraytest.VLESSConfig, error) {
	if strings.TrimSpace(rawURL) == "" {
		return prober.Config{
			Port:               443,
			Mode:               prober.ModeHTTP,
			Tries:              4,
			Timeout:            timeout,
			SNI:                "speed.cloudflare.com",
			InsecureSkipVerify: true,
		}, nil, nil
	}

	cfg, err := xraytest.ParseProxyURL(rawURL)
	if err != nil {
		return prober.Config{}, nil, err
	}
	sni := cfg.SNI
	if sni == "" {
		sni = cfg.Host
	}
	probeCfg := prober.Config{
		Port:               cfg.Port,
		Mode:               prober.ModeHTTP,
		Tries:              4,
		Timeout:            timeout,
		SNI:                sni,
		InsecureSkipVerify: true,
	}
	if cfg.Network == "ws" {
		probeCfg.WebSocketHost = cfg.Host
		probeCfg.WebSocketPath = cfg.Path
	}
	return probeCfg, cfg, nil
}

func resolvePorts(req StartRequest, cfg *xraytest.VLESSConfig) []int {
	seen := make(map[int]struct{})
	var ports []int
	add := func(port int) {
		if port == 0 {
			if cfg != nil && cfg.Port > 0 {
				port = cfg.Port
			} else {
				port = 443
			}
		}
		if port <= 0 || port > 65535 {
			return
		}
		if _, ok := seen[port]; ok {
			return
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
	}

	if req.UseConfigPort {
		add(0)
	}
	for _, port := range req.Ports {
		add(port)
	}
	if len(ports) == 0 {
		add(0)
	}
	return ports
}

func sourceIPs(ctx context.Context, req StartRequest) (<-chan net.IP, int, error) {
	if req.Source == "file" {
		ips, err := loadDefaultIPsFile()
		if err != nil {
			return nil, 0, err
		}
		if len(ips) > req.Count {
			ips = ips[:req.Count]
		}
		ch := make(chan net.IP, len(ips))
		for _, ip := range ips {
			ch <- cloneIP(ip)
		}
		close(ch)
		return ch, len(ips), nil
	}

	src, err := ipsrc.New(true, false, nil)
	if err != nil {
		return nil, 0, err
	}
	return src.Stream(ctx, req.Count), req.Count, nil
}

func phase1ResultView(r *result.Result) ResultView {
	status := "failed"
	if r.IsHealthy() {
		status = "healthy"
	}
	return ResultView{
		Phase:      "phase1",
		Endpoint:   formatEndpoint(r.IP.String(), r.Port),
		Colo:       r.Colo,
		LossPct:    round1(r.Loss()),
		AvgMS:      round1(ms(r.Avg())),
		SpeedMbps:  round1(mbps(r.Throughput)),
		Status:     status,
		IsWorking:  r.IsHealthy(),
		HTTPStatus: r.HTTPStatus,
	}
}

func phase2ResultView(r *xraytest.ValidationResult) ResultView {
	status := "failed"
	if r.Success {
		status = "working"
	}
	return ResultView{
		Phase:     "phase2",
		Endpoint:  formatEndpoint(r.IP, r.Port),
		SpeedMbps: round1(mbps(r.Throughput)),
		LatencyMS: round1(ms(r.Latency)),
		Transport: r.Transport,
		Status:    status,
		Error:     r.Error,
		IsWorking: r.Success,
	}
}

func endpointViewFromPhase1(r *result.Result) EndpointView {
	return EndpointView{
		Endpoint:  formatEndpoint(r.IP.String(), r.Port),
		IP:        r.IP.String(),
		Port:      r.Port,
		Phase:     "phase1",
		Colo:      r.Colo,
		AvgMS:     round1(ms(r.Avg())),
		SpeedMbps: round1(mbps(r.Throughput)),
	}
}

func endpointViewFromPhase2(r *xraytest.ValidationResult) EndpointView {
	return EndpointView{
		Endpoint:  formatEndpoint(r.IP, r.Port),
		IP:        r.IP,
		Port:      r.Port,
		Phase:     "phase2",
		AvgMS:     round1(ms(r.Latency)),
		SpeedMbps: round1(mbps(r.Throughput)),
	}
}

func parseTimeout(raw string, fallback time.Duration) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	if seconds, err := strconv.Atoi(raw); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func phase2Timeout(base time.Duration, speedBytes int64, minSpeed float64) time.Duration {
	timeout := base * 2
	if timeout < 10*time.Second {
		timeout = 10 * time.Second
	}
	if speedBytes <= 0 {
		speedBytes = 512 * 1024
	}
	if minSpeed <= 0 {
		minSpeed = 1
	}
	expectedSeconds := float64(speedBytes*8) / (minSpeed * 1_000_000)
	speedBudget := time.Duration(expectedSeconds * 3 * float64(time.Second))
	if speedBudget < 5*time.Second {
		speedBudget = 5 * time.Second
	}
	if speedBudget > 30*time.Second {
		speedBudget = 30 * time.Second
	}
	return timeout + speedBudget
}

func loadDefaultIPsFile() ([]net.IP, error) {
	for _, path := range ipsFileSearchPaths() {
		ips, err := loadIPs(path)
		if err == nil {
			return ips, nil
		}
	}
	return nil, fmt.Errorf("ips.txt not found next to the app or in the working directory")
}

func ipsFileSearchPaths() []string {
	seen := make(map[string]struct{})
	var paths []string
	add := func(path string) {
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	if wd, err := os.Getwd(); err == nil {
		add(filepath.Join(wd, "ips.txt"))
	}
	if exe, err := os.Executable(); err == nil {
		add(filepath.Join(filepath.Dir(exe), "ips.txt"))
	}
	return paths
}

func loadIPs(path string) ([]net.IP, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var ips []net.IP
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(strings.ToLower(line), "ip") {
			continue
		}
		field := strings.TrimSpace(strings.SplitN(line, ",", 2)[0])
		if ip := parseIPField(field); ip != nil {
			if ip.To4() != nil {
				ips = append(ips, ip)
			}
			continue
		}
		if strings.Contains(field, "/") {
			_, ipNet, err := net.ParseCIDR(field)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR %q: %w", field, err)
			}
			if ipNet.IP.To4() != nil {
				ips = append(ips, sampleFromSubnet(ipNet, 256)...)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("%s does not contain any IPv4 addresses", path)
	}
	return ips, nil
}

func parseIPField(field string) net.IP {
	if ip := net.ParseIP(field); ip != nil {
		return ip
	}
	if host, _, err := net.SplitHostPort(field); err == nil {
		return net.ParseIP(strings.Trim(host, "[]"))
	}
	if strings.Count(field, ":") == 1 && strings.Contains(field, ".") {
		host, _, _ := strings.Cut(field, ":")
		return net.ParseIP(host)
	}
	return nil
}

func sampleFromSubnet(ipNet *net.IPNet, count int) []net.IP {
	ip4 := ipNet.IP.To4()
	if ip4 == nil {
		return nil
	}
	ones, bits := ipNet.Mask.Size()
	hostBits := bits - ones
	if hostBits <= 8 {
		var ips []net.IP
		for ip := cloneIP(ipNet.IP); ipNet.Contains(ip); incrementIP(ip) {
			if ip.To4() != nil {
				ips = append(ips, cloneIP(ip))
			}
		}
		return ips
	}

	base := binary.BigEndian.Uint32(ip4)
	mask := binary.BigEndian.Uint32([]byte(ipNet.Mask))
	size := ^mask
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	seen := make(map[uint32]struct{})
	var ips []net.IP
	for i := 0; i < count*3 && len(ips) < count; i++ {
		offset := rng.Uint32() & size
		if _, ok := seen[offset]; ok {
			continue
		}
		seen[offset] = struct{}{}
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, base|offset)
		ips = append(ips, ip)
	}
	return ips
}

func writeWorkingEndpoints(endpoints []string) (string, error) {
	text := strings.Join(endpoints, "\n") + "\n"
	if exe, err := os.Executable(); err == nil {
		path := filepath.Join(filepath.Dir(exe), "working_ips.txt")
		if err := os.WriteFile(path, []byte(text), 0644); err == nil {
			return path, nil
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	path := filepath.Join(wd, "working_ips.txt")
	return path, os.WriteFile(path, []byte(text), 0644)
}

func formatEndpoint(ip string, port int) string {
	if port <= 0 {
		return ip
	}
	if strings.Contains(ip, ":") {
		return net.JoinHostPort(ip, strconv.Itoa(port))
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

func cloneIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func incrementIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func ms(d time.Duration) float64 {
	return float64(d.Microseconds()) / 1000
}

func mbps(bytesPerSecond float64) float64 {
	if bytesPerSecond <= 0 {
		return 0
	}
	return bytesPerSecond * 8 / 1_000_000
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
