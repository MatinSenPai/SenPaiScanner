package ui

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ehsatn/MAGIK/internal/engine"
	"github.com/ehsatn/MAGIK/internal/ipsrc"
	"github.com/ehsatn/MAGIK/internal/output"
	"github.com/ehsatn/MAGIK/internal/prober"
	"github.com/ehsatn/MAGIK/internal/result"
	"github.com/ehsatn/MAGIK/internal/xraytest"
)

// scanCancel holds the cancel function for the active scan so the TUI can
// abort it when the user presses esc/q.
var scanCancel context.CancelFunc
var scanIDCounter atomic.Int64

func nextScanID() int64 { return scanIDCounter.Add(1) }

// StartScanCmd builds a tea.Cmd that runs the scan engine in the background,
// sending ResultMsg and StatsMsg messages to the Bubble Tea program.
func StartScanCmd(cfg ScanConfig, scanID int64) tea.Cmd {
	return func() tea.Msg {
		go runScan(cfg, scanID)
		return nil
	}
}

// CancelScanCmd cancels the running scan.
func CancelScanCmd() tea.Cmd {
	return func() tea.Msg {
		if scanCancel != nil {
			scanCancel()
		}
		return nil
	}
}

// StartTestCmd runs the test pass against a file of IPs.
func StartTestCmd(ipFile string, scanID int64) tea.Cmd {
	return func() tea.Msg {
		go runTest(ipFile, scanID)
		return nil
	}
}

// StartColosCmd discovers accessible Cloudflare PoPs.
func StartColosCmd(scanID int64) tea.Cmd {
	return func() tea.Msg {
		go runColos(scanID)
		return nil
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

func runScan(cfg ScanConfig, scanID int64) {
	count, _ := strconv.Atoi(cfg.Count)
	concurrency, _ := strconv.Atoi(cfg.Concurrency)
	if concurrency <= 0 {
		concurrency = 50
	}
	timeout := parseTimeout(cfg.Timeout, 5*time.Second)
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
		mode = prober.ModeHTTP
	}

	var extra []string
	for _, c := range strings.Split(cfg.CIDR, ",") {
		c = strings.TrimSpace(c)
		if c != "" {
			extra = append(extra, c)
		}
	}

	useBuiltin := len(extra) == 0
	src, err := ipsrc.NewWithOptions(cfg.UseV4, cfg.UseV6, extra, ipsrc.Options{UseBuiltin: useBuiltin})
	if err != nil {
		sendError(scanID, fmt.Sprintf("Scan setup failed: %v", err))
		sendDone(scanID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	scanCancel = cancel
	defer cancel()

	engCfg := engine.Config{
		Concurrency: concurrency,
		ProbeConfig: prober.Config{
			Port:             port,
			Mode:             mode,
			Tries:            tries,
			Timeout:          timeout,
			SNI:              cfg.SNI,
			SpeedBytes:       speedSampleForMode(mode),
			RequireWebSocket: mode == prober.ModeHTTP && cfg.RequireWS,
		},
	}
	eng := engine.New(engCfg)

	coloSet := buildColoSet(cfg.ColoFilter)

	var writer *output.Writer
	if cfg.OutputFile != "" {
		fmt2 := output.DetectFormat(cfg.OutputFile)
		if w, e := output.New(cfg.OutputFile, fmt2); e == nil {
			writer = w
			defer writer.Close()
		} else {
			sendError(scanID, fmt.Sprintf("Output disabled: %v", e))
		}
	}

	ipStream := src.Stream(ctx, count)
	eng.Run(ctx, ipStream, func(r *result.Result) {
		if prog != nil {
			s := eng.Stats()
			prog.Send(StatsMsg{ScanID: scanID, Tested: s.Tested.Load(), Healthy: s.Healthy.Load(), Failed: s.Failed.Load(), InFlight: s.InFlight.Load()})
		}
		if !passesColoFilter(r, coloSet) {
			return
		}
		// Only healthy IPs go to the output file; writing every scanned IP
		// would flood the file with thousands of failed probes.
		if writer != nil && r.IsHealthy() {
			if err := writer.Write(r); err != nil {
				sendError(scanID, fmt.Sprintf("Output write failed: %v", err))
			}
		}
		if prog != nil {
			prog.Send(ResultMsg{ScanID: scanID, Result: r})
		}
	})

	sendDone(scanID)
}

func runTest(ipFile string, scanID int64) {
	ips, err := loadIPs(ipFile)
	if err != nil {
		sendError(scanID, fmt.Sprintf("Test IPs failed: %v", err))
		sendDone(scanID)
		return
	}
	if len(ips) == 0 {
		sendError(scanID, fmt.Sprintf("Test IPs failed: no valid IPs found in %s", ipFile))
		sendDone(scanID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	scanCancel = cancel
	defer cancel()

	engCfg := engine.Config{
		Concurrency: 20,
		ProbeConfig: prober.Config{
			Port:             443,
			Mode:             prober.ModeHTTP,
			Tries:            6,
			Timeout:          10 * time.Second,
			SNI:              "speed.cloudflare.com",
			SpeedBytes:       512 * 1024,
			RequireWebSocket: true,
		},
	}
	eng := engine.New(engCfg)

	eng.RunList(ctx, ips, func(r *result.Result) {
		if prog != nil {
			s := eng.Stats()
			prog.Send(ResultMsg{ScanID: scanID, Result: r})
			prog.Send(StatsMsg{ScanID: scanID, Tested: s.Tested.Load(), Healthy: s.Healthy.Load(), Failed: s.Failed.Load(), InFlight: s.InFlight.Load()})
		}
	})

	sendDone(scanID)
}

func runColos(scanID int64) {
	src, err := ipsrc.New(true, false, nil)
	if err != nil {
		sendError(scanID, fmt.Sprintf("Colo discovery failed: %v", err))
		sendColosDone(scanID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	scanCancel = cancel
	defer cancel()

	engCfg := engine.Config{
		Concurrency: 80,
		ProbeConfig: prober.Config{
			Port:       443,
			Mode:       prober.ModeHTTP,
			Tries:      2,
			Timeout:    5 * time.Second,
			SpeedBytes: 0,
		},
	}
	eng := engine.New(engCfg)
	ipStream := src.Stream(ctx, 300)

	eng.Run(ctx, ipStream, func(r *result.Result) {
		if prog != nil {
			s := eng.Stats()
			prog.Send(StatsMsg{ScanID: scanID, Tested: s.Tested.Load(), Healthy: s.Healthy.Load(), Failed: s.Failed.Load(), InFlight: s.InFlight.Load()})
		}
		if !r.IsHealthy() || r.Colo == "" {
			return
		}
		if prog != nil {
			prog.Send(ResultMsg{ScanID: scanID, Result: r})
		}
	})

	sendColosDone(scanID)
}

func sendError(scanID int64, text string) {
	if prog != nil {
		prog.Send(ErrorMsg{ScanID: scanID, Text: text})
	}
}

func sendDone(scanID int64) {
	if prog != nil {
		prog.Send(DoneMsg{ScanID: scanID})
	}
}

func sendColosDone(scanID int64) {
	if prog != nil {
		prog.Send(ColosDoneMsg{ScanID: scanID})
	}
}

// runConfigPhase1 runs Phase 1 of "Scan with Config": a fast connectivity scan
// that finds healthy Cloudflare IPs (or validates IPs from a file), then signals
// the UI to start Phase 2 (xray validation) with the best candidates.
func runConfigPhase1(opts configPhase1Options) {
	var probeCfg prober.Config
	var err error
	if strings.TrimSpace(opts.rawURL) == "" {
		probeCfg = defaultPhase1ProbeConfig(opts.timeout)
	} else {
		probeCfg, err = configProbeFromURL(opts.rawURL, opts.timeout)
		if err != nil {
			if prog != nil {
				prog.Send(ConfigPhase1ErrMsg{Err: fmt.Sprintf("invalid URL: %v", err)})
			}
			return
		}
	}
	probeCfg.RequireWebSocket = opts.requireWS
	ports := opts.ports
	if len(ports) == 0 {
		ports = []int{probeCfg.Port}
	}

	ctx, cancel := context.WithCancel(context.Background())
	scanCancel = cancel
	defer cancel()

	callback := func(r *result.Result) {
		if liveResultWriter != nil {
			liveResultWriter.AddPhase1(r)
		}
		if prog != nil {
			prog.Send(ConfigPhase1ResultMsg{Result: r})
		}
	}

	var ipStream <-chan net.IP
	neighbor := neighborScanOpts{}
	if opts.fromFile {
		ips, err := loadDefaultIPsFile()
		if err != nil {
			if prog != nil {
				prog.Send(ConfigPhase1ErrMsg{Err: err.Error()})
			}
			return
		}
		if len(ips) == 0 {
			if prog != nil {
				prog.Send(ConfigPhase1ErrMsg{Err: "ips.txt is empty — add one IP per line"})
			}
			return
		}
		if len(ips) > opts.count {
			ips = ips[:opts.count]
		}
		ch := make(chan net.IP, len(ips))
		for _, ip := range ips {
			ch <- ip
		}
		close(ch)
		ipStream = ch
	} else {
		src, err := ipsrc.New(true, false, nil)
		if err != nil {
			if prog != nil {
				prog.Send(ConfigPhase1DoneMsg{})
			}
			return
		}
		ipStream = src.Stream(ctx, opts.count)
		neighbor = neighborScanOpts{
			enabled:  true,
			nets:     src.IPv4Nets(),
			radius:   ipsrc.DefaultNeighborRadius,
			perHit:   ipsrc.DefaultNeighborPerHit,
			maxTotal: ipsrc.DefaultNeighborMaxTotal,
		}
	}
	runConfigPortProbes(ctx, ipStream, ports, opts.concurrency, probeCfg, callback, neighbor)

	if prog != nil {
		prog.Send(ConfigPhase1DoneMsg{})
	}
}

type configProbeJob struct {
	ip   net.IP
	port int
}

type neighborScanOpts struct {
	enabled  bool
	nets     []*net.IPNet
	radius   int
	perHit   int
	maxTotal int
}

type probeFunc func(context.Context, net.IP, prober.Config) *result.Result

func runConfigPortProbes(ctx context.Context, ips <-chan net.IP, ports []int, concurrency int, base prober.Config, callback func(*result.Result), neighbor neighborScanOpts) {
	runConfigPortProbesWithProbe(ctx, ips, ports, concurrency, base, callback, neighbor, prober.Probe)
}

func runConfigPortProbesWithProbe(ctx context.Context, ips <-chan net.IP, ports []int, concurrency int, base prober.Config, callback func(*result.Result), neighbor neighborScanOpts, probe probeFunc) {
	if concurrency <= 0 {
		concurrency = 50
	}
	if neighbor.enabled {
		if neighbor.radius <= 0 {
			neighbor.radius = ipsrc.DefaultNeighborRadius
		}
		if neighbor.perHit <= 0 {
			neighbor.perHit = ipsrc.DefaultNeighborPerHit
		}
		if neighbor.maxTotal <= 0 {
			neighbor.maxTotal = ipsrc.DefaultNeighborMaxTotal
		}
	}

	jobs := make(chan configProbeJob)
	results := make(chan *result.Result, concurrency)
	seen := make(map[string]struct{})
	var queue []configProbeJob
	var pending int
	neighborsQueued := 0

	jobKey := func(ip net.IP, port int) string {
		return net.JoinHostPort(ip.String(), strconv.Itoa(port))
	}

	submit := func(ip net.IP, port int) bool {
		key := jobKey(ip, port)
		if _, ok := seen[key]; ok {
			return false
		}
		seen[key] = struct{}{}
		queue = append(queue, configProbeJob{ip: ip, port: port})
		pending++
		return true
	}

	enqueueIP := func(ip net.IP) {
		for _, port := range ports {
			submit(ip, port)
		}
	}

	maybeEnqueueNeighbors := func(r *result.Result) {
		if !neighbor.enabled || !r.IsHealthy() || len(neighbor.nets) == 0 {
			return
		}

		remaining := neighbor.maxTotal - neighborsQueued
		if remaining <= 0 {
			return
		}
		limit := neighbor.perHit
		if limit > remaining {
			limit = remaining
		}

		for _, nip := range ipsrc.NeighborsAround(r.IP, neighbor.nets, neighbor.radius, limit) {
			if neighborsQueued >= neighbor.maxTotal {
				break
			}
			added := 0
			for _, port := range ports {
				if submit(nip, port) {
					added++
				}
			}
			if added > 0 {
				neighborsQueued++
			}
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if ctx.Err() != nil {
					continue
				}
				r := probe(ctx, job.ip, base.WithPort(job.port))
				select {
				case results <- r:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	input := ips
	for input != nil || pending > 0 || len(queue) > 0 {
		var send chan<- configProbeJob
		var next configProbeJob
		if len(queue) > 0 {
			send = jobs
			next = queue[0]
		}

		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return
		case ip, ok := <-input:
			if !ok {
				input = nil
				continue
			}
			enqueueIP(ip)
		case send <- next:
			queue[0] = configProbeJob{}
			queue = queue[1:]
		case r := <-results:
			pending--
			if r == nil {
				continue
			}
			callback(r)
			maybeEnqueueNeighbors(r)
		}
	}
	close(jobs)
	wg.Wait()
}

func defaultPhase1ProbeConfig(timeout time.Duration) prober.Config {
	return prober.Config{
		Port:               443,
		Mode:               prober.ModeHTTP,
		Tries:              4,
		Timeout:            timeout,
		SNI:                "speed.cloudflare.com",
		InsecureSkipVerify: true,
	}
}

func configProbeFromURL(rawURL string, timeout time.Duration) (prober.Config, error) {
	cfg, err := xraytest.ParseProxyURL(rawURL)
	if err != nil {
		return prober.Config{}, err
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
		// Phase 2 validates WS+VLESS through xray. A naked WS upgrade probe
		// false-fails on cellular/DPI even when the real tunnel works.
	}
	return probeCfg, nil
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

func ipsFileSearchPaths() []string {
	seen := make(map[string]struct{})
	add := func(paths *[]string, path string) {
		if path == "" {
			return
		}
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		*paths = append(*paths, path)
	}

	var paths []string
	if wd, err := os.Getwd(); err == nil {
		add(&paths, filepath.Join(wd, "ips.txt"))
	}
	if exe, err := os.Executable(); err == nil {
		add(&paths, filepath.Join(filepath.Dir(exe), "ips.txt"))
	}
	return paths
}

func loadDefaultIPsFile() ([]net.IP, error) {
	for _, path := range ipsFileSearchPaths() {
		ips, err := loadIPs(path)
		if err == nil {
			return ips, nil
		}
	}
	return nil, fmt.Errorf("ips.txt not found — place it next to the binary or run folder")
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
		defer f.Close()
	}
	var ips []net.IP
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "ip") {
			continue
		}
		field := strings.SplitN(line, ",", 2)[0]
		field = strings.TrimSpace(field)
		if ip := net.ParseIP(field); ip != nil {
			if ip.To4() != nil {
				ips = append(ips, ip)
			}
		} else if strings.Contains(field, "/") {
			_, ipNet, err := net.ParseCIDR(field)
			if err == nil {
				if ipNet.IP.To4() != nil {
					ips = append(ips, sampleFromSubnet(ipNet, 256)...)
				}
			} else {
				return nil, fmt.Errorf("invalid CIDR %q: %w", field, err)
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	// Shuffle loaded IPs to ensure random order and prevent sequential scanning bias
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(ips), func(i, j int) {
		ips[i], ips[j] = ips[j], ips[i]
	})

	return ips, nil
}

func sampleFromSubnet(ipNet *net.IPNet, count int) []net.IP {
	ip4 := ipNet.IP.To4()
	if ip4 == nil {
		return nil
	}

	ones, bits := ipNet.Mask.Size()
	hostBits := bits - ones

	// If the subnet contains <= count IPs, expand it fully.
	// For IPv4, hostBits <= 8 (e.g. /24 and smaller) means <= 256 IPs.
	if hostBits <= 8 {
		var ips []net.IP
		for ip := cloneIP(ipNet.IP); ipNet.Contains(ip); incrementIP(ip) {
			if ip.To4() != nil {
				ips = append(ips, cloneIP(ip))
			}
		}
		return ips
	}

	// Otherwise, randomly sample unique IPs
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	base := binary.BigEndian.Uint32(ip4)
	mask := binary.BigEndian.Uint32([]byte(ipNet.Mask))
	size := ^mask

	seen := make(map[uint32]struct{})
	var ips []net.IP
	// Try up to count * 3 times to avoid infinite loop
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

func speedSampleForMode(mode prober.Mode) int64 {
	if mode != prober.ModeHTTP {
		return 0
	}
	// 128 KB gives DPI more time to act on the connection. On Iranian ISPs,
	// smaller samples sometimes complete before DPI identifies and blocks
	// proxy traffic patterns, producing false-positive "healthy" results.
	return 128 * 1024
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

// MetaMsg holds network metadata fetched from speed.cloudflare.com/meta
type MetaMsg struct {
	ASN            int    `json:"asn"`
	ASOrganization string `json:"asOrganization"`
	Colo           string `json:"colo"`
	Country        string `json:"country"`
	IP             string `json:"ip"`
}

// FetchMetaCmd fetches connection metadata from cloudflare meta server,
// falling back to ip-api.com if cloudflare is unreachable.
func FetchMetaCmd() tea.Cmd {
	return func() tea.Msg {
		client := http.Client{Timeout: 5 * time.Second}

		// Try Cloudflare first (provides colo data).
		resp, err := client.Get("https://speed.cloudflare.com/meta")
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				var m MetaMsg
				if err := json.NewDecoder(resp.Body).Decode(&m); err == nil && m.ASOrganization != "" {
					return m
				}
			}
		}

		// Fallback: ip-api.com (HTTP, free tier).
		resp2, err2 := client.Get("http://ip-api.com/json")
		if err2 == nil {
			defer resp2.Body.Close()
			if resp2.StatusCode >= 200 && resp2.StatusCode < 300 {
				var raw struct {
					ISP string `json:"isp"`
					AS  string `json:"as"`
					IP  string `json:"query"`
				}
				if err := json.NewDecoder(resp2.Body).Decode(&raw); err == nil && raw.ISP != "" {
					return MetaMsg{
						ASOrganization: raw.ISP,
						IP:             raw.IP,
					}
				}
			}
		}

		// Fallback 2: ip.dnslab.link (Iranian IP provider).
		resp3, err3 := client.Get("http://ip.dnslab.link")
		if err3 == nil {
			defer resp3.Body.Close()
			if resp3.StatusCode >= 200 && resp3.StatusCode < 300 {
				bodyBytes, err := io.ReadAll(io.LimitReader(resp3.Body, 64))
				if err == nil {
					ipStr := strings.TrimSpace(string(bodyBytes))
					if net.ParseIP(ipStr) != nil {
						ispName, found := LookupIranISP(ipStr)
						if !found {
							ispName = "Iranian ISP"
						}
						return MetaMsg{
							ASOrganization: ispName,
							IP:             ipStr,
						}
					}
				}
			}
		}

		// Fallback 3: ipify.ir (Iranian IP provider).
		resp4, err4 := client.Get("http://ipify.ir")
		if err4 == nil {
			defer resp4.Body.Close()
			if resp4.StatusCode >= 200 && resp4.StatusCode < 300 {
				bodyBytes, err := io.ReadAll(io.LimitReader(resp4.Body, 64))
				if err == nil {
					ipStr := strings.TrimSpace(string(bodyBytes))
					if net.ParseIP(ipStr) != nil {
						ispName, found := LookupIranISP(ipStr)
						if !found {
							ispName = "Iranian ISP"
						}
						return MetaMsg{
							ASOrganization: ispName,
							IP:             ipStr,
						}
					}
				}
			}
		}

		return MetaMsg{ASOrganization: "Unknown ISP"}
	}
}
