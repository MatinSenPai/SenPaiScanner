package mobile

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ehsatn/MAGIK/internal/ipsrc"
	"github.com/ehsatn/MAGIK/internal/prober"
	"github.com/ehsatn/MAGIK/internal/result"
	"github.com/ehsatn/MAGIK/internal/xraytest"
)

type Callback interface {
	OnProgress(tested int, healthy int, failed int, inFlight int, isPhase2 bool)
	OnResult(ip string, port int, latencyMs int, loss float64, colo string, isHealthy bool, isPhase2 bool, phase2Type string, phase2Speed float64, phase2Status bool)
	OnFinished()
	OnError(err string)
}

var (
	mu         sync.Mutex
	cancelScan context.CancelFunc
	isRunning  bool
)

type ScanConfig struct {
	SourceType    string `json:"sourceType"`
	SourceFile    string `json:"sourceFile"`
	CountType     string `json:"countType"`
	CustomCount   string `json:"customCount"`
	WorkerType    string `json:"workerType"`
	CustomWorkers string `json:"customWorkers"`
	TimeoutType   string `json:"timeoutType"`
	CustomTimeout string `json:"customTimeout"`
	PortType      string `json:"portType"`
	SelectedPorts []int  `json:"selectedPorts"`
	ConfigURL     string `json:"configUrl"`
	TopNType      string `json:"topNType"`
	CustomTopN    string `json:"customTopN"`
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

func loadIPs(path string) ([]net.IP, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ips []net.IP
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		field := strings.SplitN(line, ",", 2)[0]
		if ip := net.ParseIP(strings.TrimSpace(field)); ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips, sc.Err()
}

func runScan(configJson string, callback Callback) {
	defer func() {
		mu.Lock()
		isRunning = false
		mu.Unlock()
		if callback != nil {
			callback.OnFinished()
		}
	}()

	var cfg ScanConfig
	err := json.Unmarshal([]byte(configJson), &cfg)
	if err != nil {
		if callback != nil {
			callback.OnError(fmt.Sprintf("Failed to parse config: %v", err))
		}
		return
	}

	count := 5000
	if cfg.CountType == "Custom" {
		count, _ = strconv.Atoi(cfg.CustomCount)
	} else if c, err := strconv.Atoi(cfg.CountType); err == nil {
		count = c
	}
	if count <= 0 {
		count = 5000
	}

	concurrency := 50
	if cfg.WorkerType == "Custom" {
		concurrency, _ = strconv.Atoi(cfg.CustomWorkers)
	} else if strings.HasPrefix(cfg.WorkerType, "50") {
		concurrency = 50
	} else if strings.HasPrefix(cfg.WorkerType, "100") {
		concurrency = 100
	} else if strings.HasPrefix(cfg.WorkerType, "200") {
		concurrency = 200
	}
	if concurrency <= 0 {
		concurrency = 50
	}

	timeout := 5 * time.Second
	if cfg.TimeoutType == "Custom" {
		t, _ := strconv.Atoi(cfg.CustomTimeout)
		if t > 0 {
			timeout = time.Duration(t) * time.Millisecond
		}
	} else if strings.HasPrefix(cfg.TimeoutType, "2s") {
		timeout = 2 * time.Second
	} else if strings.HasPrefix(cfg.TimeoutType, "3s") {
		timeout = 3 * time.Second
	}

	var ports []int
	var probeCfg prober.Config
	isConfigMode := strings.TrimSpace(cfg.ConfigURL) != ""

	if isConfigMode {
		xCfg, err := xraytest.ParseProxyURL(cfg.ConfigURL)
		if err != nil {
			if callback != nil {
				callback.OnError(fmt.Sprintf("Invalid Config URL: %v", err))
			}
			return
		}
		sni := xCfg.SNI
		if sni == "" {
			sni = xCfg.Host
		}
		probeCfg = prober.Config{
			Port:               xCfg.Port,
			Mode:               prober.ModeHTTP,
			Tries:              3,
			Timeout:            timeout,
			SNI:                sni,
			InsecureSkipVerify: true,
		}
		if xCfg.Network == "ws" {
			probeCfg.WebSocketHost = xCfg.Host
			probeCfg.WebSocketPath = xCfg.Path
		}
		ports = []int{xCfg.Port}
	} else {
		ports = cfg.SelectedPorts
		if len(ports) == 0 {
			ports = []int{443}
		}
		probeCfg = prober.Config{
			Mode:               prober.ModeHTTP,
			Tries:              3,
			Timeout:            timeout,
			SNI:                "speed.cloudflare.com",
			InsecureSkipVerify: true,
		}
	}

	var ipStream <-chan net.IP
	var neighborNets []*net.IPNet
	if cfg.SourceType == "From File" && cfg.SourceFile != "" {
		ips, err := loadIPs(cfg.SourceFile)
		if err != nil {
			if callback != nil {
				callback.OnError(fmt.Sprintf("Failed to load IPs: %v", err))
			}
			return
		}
		if len(ips) == 0 {
			if callback != nil {
				callback.OnError("File is empty or contains invalid IPs")
			}
			return
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
			if callback != nil {
				callback.OnError(fmt.Sprintf("Failed to initialize IP source: %v", err))
			}
			return
		}
		srcCtx, srcCancel := context.WithCancel(context.Background())
		defer srcCancel()
		ipStream = src.Stream(srcCtx, count)
		neighborNets = src.IPv4Nets()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mu.Lock()
	cancelScan = cancel
	mu.Unlock()

	var statsTested, statsHealthy, statsFailed, statsInFlight int32
	var isPhase2 int32 // 0 = false, 1 = true

	var phase1Results []*result.Result
	var resMu sync.Mutex

	// Update stats routine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
				// Use atomic loads to read stats without blocking
				tested := atomic.LoadInt32(&statsTested)
				healthy := atomic.LoadInt32(&statsHealthy)
				failed := atomic.LoadInt32(&statsFailed)
				inFlight := atomic.LoadInt32(&statsInFlight)
				phase2 := atomic.LoadInt32(&isPhase2) == 1
				if callback != nil {
					callback.OnProgress(int(tested), int(healthy), int(failed), int(inFlight), phase2)
				}
			}
		}
	}()

	// Phase 1 Engine with neighbor scanning
	type probeJob struct {
		ip   net.IP
		port int
	}
	jobs := make(chan probeJob)
	results := make(chan *result.Result, concurrency)

	seen := make(map[string]struct{})
	var queue []probeJob
	var pending int
	neighborsQueued := 0

	neighborEnabled := len(neighborNets) > 0 && !isConfigMode
	neighborRadius := ipsrc.DefaultNeighborRadius
	neighborPerHit := ipsrc.DefaultNeighborPerHit
	neighborMaxTotal := ipsrc.DefaultNeighborMaxTotal

	jobKey := func(ip net.IP, port int) string {
		return ip.String() + ":" + strconv.Itoa(port)
	}

	submit := func(ip net.IP, port int) bool {
		key := jobKey(ip, port)
		if _, ok := seen[key]; ok {
			return false
		}
		seen[key] = struct{}{}
		queue = append(queue, probeJob{ip: ip, port: port})
		pending++
		return true
	}

	enqueueIP := func(ip net.IP) {
		for _, port := range ports {
			submit(ip, port)
		}
	}

	maybeEnqueueNeighbors := func(r *result.Result) {
		if !neighborEnabled || !r.IsHealthy() || len(neighborNets) == 0 {
			return
		}
		remaining := neighborMaxTotal - neighborsQueued
		if remaining <= 0 {
			return
		}
		limit := neighborPerHit
		if limit > remaining {
			limit = remaining
		}
		for _, nip := range ipsrc.NeighborsAround(r.IP, neighborNets, neighborRadius, limit) {
			if neighborsQueued >= neighborMaxTotal {
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
				atomic.AddInt32(&statsTested, 1)
				atomic.AddInt32(&statsInFlight, 1)

				r := prober.Probe(ctx, job.ip, probeCfg.WithPort(job.port))
				select {
				case results <- r:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	input := ipStream
	for input != nil || pending > 0 || len(queue) > 0 {
		var send chan<- probeJob
		var next probeJob
		if len(queue) > 0 {
			send = jobs
			next = queue[0]
		}

		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			close(results)
			goto phase1Done
		case ip, ok := <-input:
			if !ok {
				input = nil
				continue
			}
			enqueueIP(ip)
		case send <- next:
			queue[0] = probeJob{}
			queue = queue[1:]
		case r := <-results:
			pending--
			if r == nil {
				continue
			}
			atomic.AddInt32(&statsInFlight, -1)
			if r.IsHealthy() {
				atomic.AddInt32(&statsHealthy, 1)
			} else {
				atomic.AddInt32(&statsFailed, 1)
			}

			if r.IsHealthy() {
				resMu.Lock()
				phase1Results = append(phase1Results, r)
				resMu.Unlock()
				if callback != nil {
					callback.OnResult(r.IP.String(), r.Port, int(r.Avg().Milliseconds()), r.Loss(), r.Colo, true, false, "", 0.0, false)
				}
			}
			maybeEnqueueNeighbors(r)
		}
	}
	close(jobs)
	wg.Wait()
	close(results)

	// Drain remaining results
	for r := range results {
		pending--
		if r == nil {
			continue
		}
		atomic.AddInt32(&statsInFlight, -1)
		if r.IsHealthy() {
			atomic.AddInt32(&statsHealthy, 1)
		} else {
			atomic.AddInt32(&statsFailed, 1)
		}

		if r.IsHealthy() {
			resMu.Lock()
			phase1Results = append(phase1Results, r)
			resMu.Unlock()
			if callback != nil {
				callback.OnResult(r.IP.String(), r.Port, int(r.Avg().Milliseconds()), r.Loss(), r.Colo, true, false, "", 0.0, false)
			}
		}
	}

phase1Done:

	if ctx.Err() != nil {
		return
	}

	// Send final Phase 1 progress update
	if callback != nil {
		tested := int(atomic.LoadInt32(&statsTested))
		healthy := int(atomic.LoadInt32(&statsHealthy))
		failed := int(atomic.LoadInt32(&statsFailed))
		inFlight := int(atomic.LoadInt32(&statsInFlight))
		callback.OnProgress(tested, healthy, failed, inFlight, false)
	}

	// Phase 2
	if isConfigMode && len(phase1Results) > 0 {
		topN := 50
		if cfg.TopNType == "Custom" {
			topN, _ = strconv.Atoi(cfg.CustomTopN)
		} else if cfg.TopNType == "ALL" {
			topN = len(phase1Results)
		} else if n, err := strconv.Atoi(cfg.TopNType); err == nil {
			topN = n
		}
		if topN <= 0 {
			topN = 50
		}

		sort.Slice(phase1Results, func(i, j int) bool {
			return phase1Results[i].Avg() < phase1Results[j].Avg()
		})

		if len(phase1Results) > topN {
			phase1Results = phase1Results[:topN]
		}

		xCfg, err := xraytest.ParseProxyURL(cfg.ConfigURL)
		if err != nil {
			if callback != nil {
				callback.OnError(fmt.Sprintf("Phase 2 failed to parse config URL: %v", err))
			}
			return
		}

		atomic.StoreInt32(&statsTested, 0)
		atomic.StoreInt32(&statsHealthy, 0)
		atomic.StoreInt32(&statsFailed, 0)
		atomic.StoreInt32(&statsInFlight, int32(len(phase1Results)))
		atomic.StoreInt32(&isPhase2, 1)

		// Create buffered channel for Phase 2 results
		type phase2ResultMsg struct {
			ip         string
			port       int
			latencyMs  int
			colo       string
			transport  string
			throughput float64
			success    bool
		}
		phase2ResultChan := make(chan phase2ResultMsg, len(phase1Results))

		// Create WaitGroup to ensure callback goroutine completes before OnFinished
		var callbackWg sync.WaitGroup
		callbackWg.Add(1)

		// Launch callback goroutine to decouple callback invocation from validation loop
		go func() {
			defer callbackWg.Done()
			for {
				select {
				case msg, ok := <-phase2ResultChan:
					if !ok {
						// Channel closed, all results processed
						return
					}
					if callback != nil {
						callback.OnResult(msg.ip, msg.port, msg.latencyMs, 0.0, msg.colo, true, true, msg.transport, msg.throughput, msg.success)
					}
				case <-ctx.Done():
					// Context cancelled, drain remaining messages and exit
					for msg := range phase2ResultChan {
						if callback != nil {
							callback.OnResult(msg.ip, msg.port, msg.latencyMs, 0.0, msg.colo, true, true, msg.transport, msg.throughput, msg.success)
						}
					}
					return
				}
			}
		}()

		// Phase 2 validation loop
		for _, r := range phase1Results {
			if ctx.Err() != nil {
				break
			}
			swapped := xCfg.WithEndpoint(r.IP.String(), r.Port)
			vr := mobileValidateConfig(ctx, swapped, 22*time.Second)

			atomic.AddInt32(&statsTested, 1)
			atomic.AddInt32(&statsInFlight, -1)
			if vr.Success {
				atomic.AddInt32(&statsHealthy, 1)
			} else {
				atomic.AddInt32(&statsFailed, 1)
			}

			// Send result to callback goroutine instead of calling callback directly
			select {
			case phase2ResultChan <- phase2ResultMsg{
				ip:         r.IP.String(),
				port:       r.Port,
				latencyMs:  int(vr.Latency.Milliseconds()),
				colo:       r.Colo,
				transport:  vr.Transport,
				throughput: vr.Throughput,
				success:    vr.Success,
			}:
			case <-ctx.Done():
				break
			}
		}

		// Send final progress update so the UI shows 100% (n/n) before finishing
		if callback != nil {
			tested := int(atomic.LoadInt32(&statsTested))
			healthy := int(atomic.LoadInt32(&statsHealthy))
			failed := int(atomic.LoadInt32(&statsFailed))
			inFlight := int(atomic.LoadInt32(&statsInFlight))
			callback.OnProgress(tested, healthy, failed, inFlight, true)
		}

		// Close result channel after Phase 2 loop completes
		close(phase2ResultChan)

		// Wait for callback goroutine to finish processing all results before OnFinished
		callbackWg.Wait()
	}
}
