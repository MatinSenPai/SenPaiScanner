package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/ipsrc"
	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

// This file exports the CLI scan primitives for non-tea frontends (the Wails
// desktop GUI). Every function wraps the exact code path the TUI uses; the
// unexported originals remain untouched.

// Phase1ProbeConfig builds the Phase 1 probe config exactly like the CLI:
// defaultPhase1ProbeConfig when rawURL is empty, otherwise the URL-derived
// probe (configProbeFromURL), with RequireWebSocket applied.
func Phase1ProbeConfig(rawURL string, timeout time.Duration, requireWS bool) (prober.Config, error) {
	var cfg prober.Config
	var err error
	if trimURL(rawURL) == "" {
		cfg = defaultPhase1ProbeConfig(timeout)
	} else {
		cfg, err = configProbeFromURL(rawURL, timeout)
		if err != nil {
			return prober.Config{}, err
		}
	}
	cfg.RequireWebSocket = requireWS
	return cfg, nil
}

func trimURL(raw string) string {
	start := 0
	for start < len(raw) && (raw[start] == ' ' || raw[start] == '\t' || raw[start] == '\n' || raw[start] == '\r') {
		start++
	}
	end := len(raw)
	for end > start && (raw[end-1] == ' ' || raw[end-1] == '\t' || raw[end-1] == '\n' || raw[end-1] == '\r') {
		end--
	}
	return raw[start:end]
}

// NeighborScanOpts is the exported mirror of the CLI's neighborScanOpts.
type NeighborScanOpts struct {
	Enabled  bool
	Nets     []*net.IPNet
	Radius   int
	PerHit   int
	MaxTotal int
}

// DefaultNeighborOpts builds the Random-source neighbor scan settings the CLI
// uses (disabled when nets is empty, which only happens for file-mode scans).
func DefaultNeighborOpts(nets []*net.IPNet) NeighborScanOpts {
	if len(nets) == 0 {
		return NeighborScanOpts{}
	}
	return NeighborScanOpts{
		Enabled:  true,
		Nets:     nets,
		Radius:   ipsrc.DefaultNeighborRadius,
		PerHit:   ipsrc.DefaultNeighborPerHit,
		MaxTotal: ipsrc.DefaultNeighborMaxTotal,
	}
}

// RunPortProbes fans probes over IPs × ports with neighbor scanning, exactly
// like the CLI's runConfigPortProbes. The callback is invoked once per result.
func RunPortProbes(ctx context.Context, ips <-chan net.IP, ports []int, concurrency int, base prober.Config, callback func(*result.Result), neighbor NeighborScanOpts) {
	runConfigPortProbes(ctx, ips, ports, concurrency, base, callback, neighborScanOpts{
		enabled:  neighbor.Enabled,
		nets:     neighbor.Nets,
		radius:   neighbor.Radius,
		perHit:   neighbor.PerHit,
		maxTotal: neighbor.MaxTotal,
	})
}

// LoadIPsFile wraps the CLI's ips.txt discovery (next to the app or cwd).
func LoadIPsFile() ([]net.IP, error) {
	return loadDefaultIPsFile()
}

// LoadIPsFromPath loads IPs from an explicit file path chosen by the user.
func LoadIPsFromPath(path string) ([]net.IP, error) {
	if strings.TrimSpace(path) == "" {
		return loadDefaultIPsFile()
	}
	return loadIPs(path)
}

// NewLiveResultWriter creates a live results writer and returns it with the
// path it will write to.
func NewLiveResultWriter(withConfig bool) (*LiveResultWriter, string, error) {
	w, path, err := newLiveResultWriter(withConfig)
	return w, path, err
}

// LoadAppConfig and SaveAppConfig expose the shared config.json persistence
// used by the CLI's Retry Last Scan.
func LoadAppConfig() AppConfig          { return loadAppConfig() }
func SaveAppConfig(cfg AppConfig) error { return saveAppConfig(cfg) }

// LoadScanConfigFile reads a saved scan config from an arbitrary path. The file
// may be either a bare SavedConfig or an AppConfig wrapper ({"last_config":…});
// both shapes are accepted so users can reload the app's own config.json or a
// file exported via SaveScanConfigFile.
func LoadScanConfigFile(path string) (SavedConfig, error) {
	return loadScanConfigFile(path)
}

// SaveScanConfigFile writes a SavedConfig to an arbitrary path as an AppConfig
// wrapper, matching the shared config.json layout.
func SaveScanConfigFile(path string, cfg SavedConfig) error {
	return saveScanConfigFile(path, cfg)
}

// FetchMeta fetches connection metadata (Cloudflare meta with ip-api.com
// fallback) without any tea dependency.
func FetchMeta() MetaMsg { return fetchMeta() }

// Phase2Timeout reproduces the CLI's xray validation timeout budget.
func Phase2Timeout(timeout time.Duration, minSpeed float64, speedSize int64) time.Duration {
	return phase2TimeoutBudget(timeout, minSpeed, speedSize)
}

// phase2TimeoutBudget is the CLI formula extracted so both frontends share it.
func phase2TimeoutBudget(timeout time.Duration, minSpeed float64, speedSize int64) time.Duration {
	xrayTimeout := timeout * 2
	if xrayTimeout < 10*time.Second {
		xrayTimeout = 10 * time.Second
	}
	speedBits := speedSize * 8
	effectiveMinSpeed := minSpeed
	if effectiveMinSpeed <= 0 {
		effectiveMinSpeed = 1.0
	}
	expectedSpeedSec := float64(speedBits) / (effectiveMinSpeed * 1_000_000)
	speedLimit := time.Duration(expectedSpeedSec * 3 * float64(time.Second))
	if speedLimit < 5*time.Second {
		speedLimit = 5 * time.Second
	}
	if speedLimit > 30*time.Second {
		speedLimit = 30 * time.Second
	}
	return xrayTimeout + speedLimit
}

// RunPhase2 validates topIPs against rawURL through xray with the same
// semantics as the CLI's runConfigPhase2: 10 workers, min-speed filter,
// speed URL/size/upload settings applied per candidate. onResult is called
// once per validated endpoint.
func RunPhase2(ctx context.Context, rawURL string, topIPs []*result.Result, minSpeed float64, speedURL string, speedSize int64, timeout time.Duration, uploadTest bool, uploadSize int64, uploadURL string, speedMode string, onResult func(*xraytest.ValidationResult, int, int)) error {
	cfg, err := xraytest.ParseProxyURL(rawURL)
	if err != nil {
		return fmt.Errorf("invalid config URL: %w", err)
	}
	total := len(topIPs)
	if total == 0 || onResult == nil {
		return nil
	}

	if errMsg := cfg.Phase2SanityError(); errMsg != "" {
		for i, r := range topIPs {
			onResult(&xraytest.ValidationResult{
				IP:        r.IP.String(),
				Port:      r.Port,
				Transport: cfg.Network,
				Error:     errMsg,
			}, i+1, total)
		}
		return nil
	}

	sem := make(chan struct{}, phase2WorkersCount)
	var wg sync.WaitGroup
	var done atomicInt32

	for _, r := range topIPs {
		r := r
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			swapped := cfg.WithEndpoint(r.IP.String(), r.Port)
			swapped.SpeedURL = speedURL
			swapped.SpeedSize = speedSize
			swapped.UploadTest = uploadTest
			swapped.UploadSize = uploadSize
			swapped.UploadURL = uploadURL
			swapped.SpeedMode = speedMode
			vr := xraytest.ValidateConfig(ctx, swapped, timeout)
			if vr.Success && minSpeed > 0 {
				mbps := vr.Throughput * 8 / 1_000_000
				if speedMode == "upload" {
					mbps = vr.UploadThroughput * 8 / 1_000_000
				}
				if mbps < minSpeed {
					vr.Success = false
					vr.Error = fmt.Sprintf("speed below threshold (%.1f < %.1f Mbps)", mbps, minSpeed)
				}
			}
			onResult(vr, int(done.add()), total)
		}()
	}

	wg.Wait()
	return nil
}

// ---------------------------------------------------------------------------
// Preset accessors — one source of truth for CLI/GUI index mapping
// ---------------------------------------------------------------------------

// PresetList pairs display labels with raw preset values.
type PresetList struct {
	Labels []string
	Values []string
}

func presetLabelsValues(presets []quickPreset) PresetList {
	list := PresetList{Labels: make([]string, len(presets)), Values: make([]string, len(presets))}
	for i, p := range presets {
		list.Labels[i] = p.label
		list.Values[i] = p.value
	}
	return list
}

func ConfigCountPresets() PresetList {
	return PresetList{Labels: configCountLabels, Values: intsToStrings(configCountValues)}
}
func ConfigTopNPresets() PresetList {
	return PresetList{Labels: configTopNLabels, Values: intsToStrings(configTopNValues)}
}
func ConfigMinSpeedPresets() PresetList {
	list := PresetList{Labels: configMinSpeedLabels, Values: make([]string, len(configMinSpeedLabels))}
	for i, v := range configMinSpeedValues {
		list.Values[i] = fmt.Sprintf("%v", v)
	}
	return list
}
func ConfigSpeedSizePresets() PresetList {
	list := PresetList{Labels: configSpeedSizeLabels, Values: make([]string, len(configSpeedSizeLabels))}
	for i, v := range configSpeedSizeValues {
		list.Values[i] = fmt.Sprintf("%v", v)
	}
	return list
}
func ConfigWorkerPresets() PresetList  { return presetLabelsValues(quickWorkersPresets) }
func ConfigTimeoutPresets() PresetList { return presetLabelsValues(quickTimeoutPresets) }

// ConfigPorts returns the selectable ports; 0 means "Config" (URL-derived).
func ConfigPorts() []int {
	ports := make([]int, len(configPortChoices))
	for i, c := range configPortChoices {
		ports[i] = c.port
	}
	return ports
}

func intsToStrings(values []int) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = fmt.Sprintf("%d", v)
	}
	return out
}

// WorkingEndpoints returns "ip:port" strings for successful validations.
func WorkingEndpoints(results []*xraytest.ValidationResult) []string {
	return workingEndpoints(results)
}

// WorkingIPs returns one IP per line for successful validations.
func WorkingIPs(results []*xraytest.ValidationResult) []string {
	return workingIPs(results)
}

// fetchMeta is the tea-free body of FetchMetaCmd. Cloudflare is the primary
// source because it also reports the active edge colo. If it is incomplete or
// blocked, two HTTPS providers are queried concurrently and merged field by
// field. Team Cymru DNS is the final ASN/name fallback when only the public IP
// could be discovered.
func fetchMeta() MetaMsg {
	client := &http.Client{Timeout: 3500 * time.Millisecond}
	best, _ := fetchCloudflareMeta(client)
	if metaComplete(best) {
		return best
	}

	type metaResult struct {
		meta MetaMsg
		err  error
	}
	results := make(chan metaResult, 2)
	go func() {
		m, err := fetchIPWhoMeta(client, best.IP)
		results <- metaResult{m, err}
	}()
	go func() {
		m, err := fetchIPInfoMeta(client, best.IP)
		results <- metaResult{m, err}
	}()
	for i := 0; i < 2; i++ {
		r := <-results
		if r.err == nil {
			best = mergeMeta(best, r.meta)
		}
	}

	if best.IP != "" && (best.ASN == 0 || best.ASOrganization == "") {
		if m, err := lookupCymruMeta(best.IP); err == nil {
			best = mergeMeta(best, m)
		}
	}
	if best.IP != "" && cleanASOrganization(best.ASOrganization) == "" {
		if isp, ok := LookupIranISP(best.IP); ok {
			best.ASOrganization = isp
			if best.Source == "" {
				best.Source = "Embedded Iran ISP ranges"
			} else {
				best.Source += " + Embedded Iran ISP ranges"
			}
		}
	}
	best.ASOrganization = cleanASOrganization(best.ASOrganization)
	if best.ASOrganization == "" && best.ASN > 0 {
		best.ASOrganization = fmt.Sprintf("AS%d", best.ASN)
	}
	if best.ASOrganization == "" {
		best.ASOrganization = "Unknown ISP"
	}
	return best
}

func fetchCloudflareMeta(client *http.Client) (MetaMsg, error) {
	var m MetaMsg
	if err := getJSON(client, "https://speed.cloudflare.com/meta", &m); err != nil {
		return MetaMsg{}, err
	}
	m.Source = "Cloudflare"
	m.ASOrganization = cleanASOrganization(m.ASOrganization)
	return m, nil
}

func fetchIPWhoMeta(client *http.Client, ip string) (MetaMsg, error) {
	url := "https://ipwho.is/"
	if parsed := net.ParseIP(strings.TrimSpace(ip)); parsed != nil {
		url += parsed.String()
	}
	var raw struct {
		Success     bool   `json:"success"`
		IP          string `json:"ip"`
		CountryCode string `json:"country_code"`
		Connection  struct {
			ASN int    `json:"asn"`
			Org string `json:"org"`
			ISP string `json:"isp"`
		} `json:"connection"`
	}
	if err := getJSON(client, url, &raw); err != nil {
		return MetaMsg{}, err
	}
	if !raw.Success {
		return MetaMsg{}, fmt.Errorf("ipwho lookup failed")
	}
	org := raw.Connection.ISP
	if org == "" {
		org = raw.Connection.Org
	}
	return MetaMsg{ASN: raw.Connection.ASN, ASOrganization: cleanASOrganization(org), Country: raw.CountryCode, IP: raw.IP, Source: "IPWhois"}, nil
}

func fetchIPInfoMeta(client *http.Client, ip string) (MetaMsg, error) {
	url := "https://ipinfo.io/json"
	if parsed := net.ParseIP(strings.TrimSpace(ip)); parsed != nil {
		url = "https://ipinfo.io/" + parsed.String() + "/json"
	}
	var raw struct {
		IP      string `json:"ip"`
		Org     string `json:"org"`
		Country string `json:"country"`
	}
	if err := getJSON(client, url, &raw); err != nil {
		return MetaMsg{}, err
	}
	return MetaMsg{ASN: parseASN(raw.Org), ASOrganization: cleanASOrganization(raw.Org), Country: raw.Country, IP: raw.IP, Source: "IPinfo"}, nil
}

func getJSON(client *http.Client, url string, dst any) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SenPaiScanner/0.8")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("metadata endpoint returned %s", resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

func metaComplete(m MetaMsg) bool {
	return m.IP != "" && m.ASN > 0 && cleanASOrganization(m.ASOrganization) != ""
}

func mergeMeta(primary, fallback MetaMsg) MetaMsg {
	if primary.ASN == 0 {
		primary.ASN = fallback.ASN
	}
	if cleanASOrganization(primary.ASOrganization) == "" {
		primary.ASOrganization = fallback.ASOrganization
	}
	if primary.Colo == "" {
		primary.Colo = fallback.Colo
	}
	if primary.Country == "" {
		primary.Country = fallback.Country
	}
	if primary.IP == "" {
		primary.IP = fallback.IP
	}
	if primary.Source == "" {
		primary.Source = fallback.Source
	} else if fallback.Source != "" && !strings.Contains(primary.Source, fallback.Source) {
		primary.Source += " + " + fallback.Source
	}
	return primary
}

func parseASN(raw string) int {
	upper := strings.ToUpper(strings.TrimSpace(raw))
	if !strings.HasPrefix(upper, "AS") {
		return 0
	}
	end := 2
	for end < len(upper) && upper[end] >= '0' && upper[end] <= '9' {
		end++
	}
	n, _ := strconv.Atoi(upper[2:end])
	return n
}

func cleanASOrganization(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" || strings.EqualFold(value, "unknown isp") {
		return ""
	}
	upper := strings.ToUpper(value)
	if strings.HasPrefix(upper, "AS") {
		i := 2
		for i < len(value) && value[i] >= '0' && value[i] <= '9' {
			i++
		}
		value = strings.TrimSpace(value[i:])
	}
	return strings.Trim(value, " |,-")
}

func lookupCymruMeta(rawIP string) (MetaMsg, error) {
	ip := net.ParseIP(strings.TrimSpace(rawIP))
	if ip == nil {
		return MetaMsg{}, fmt.Errorf("invalid public IP")
	}
	query := cymruOriginName(ip)
	if query == "" {
		return MetaMsg{}, fmt.Errorf("unsupported IP")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	txt, err := net.DefaultResolver.LookupTXT(ctx, query)
	if err != nil || len(txt) == 0 {
		return MetaMsg{}, fmt.Errorf("Cymru origin lookup failed: %w", err)
	}
	parts := strings.Split(txt[0], "|")
	if len(parts) == 0 {
		return MetaMsg{}, fmt.Errorf("invalid Cymru origin response")
	}
	asn, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	m := MetaMsg{ASN: asn, IP: ip.String(), Source: "Team Cymru DNS"}
	if len(parts) > 2 {
		m.Country = strings.TrimSpace(parts[2])
	}
	if asn <= 0 {
		return m, nil
	}
	nameTXT, err := net.DefaultResolver.LookupTXT(ctx, fmt.Sprintf("AS%d.asn.cymru.com", asn))
	if err == nil && len(nameTXT) > 0 {
		nameParts := strings.Split(nameTXT[0], "|")
		if len(nameParts) >= 5 {
			m.ASOrganization = cleanASOrganization(nameParts[4])
		}
	}
	return m, nil
}

func cymruOriginName(ip net.IP) string {
	if ip4 := ip.To4(); ip4 != nil {
		return fmt.Sprintf("%d.%d.%d.%d.origin.asn.cymru.com", ip4[3], ip4[2], ip4[1], ip4[0])
	}
	ip16 := ip.To16()
	if ip16 == nil {
		return ""
	}
	var b strings.Builder
	for i := len(ip16) - 1; i >= 0; i-- {
		fmt.Fprintf(&b, "%x.%x.", ip16[i]&0x0f, ip16[i]>>4)
	}
	b.WriteString("origin6.asn.cymru.com")
	return b.String()
}

// atomicInt32 is a tiny counter used by RunPhase2.
type atomicInt32 struct{ v atomic.Int32 }

func (a *atomicInt32) add() int32 { return a.v.Add(1) }
