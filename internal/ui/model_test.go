package ui

import (
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ehsatn/MAGIK/internal/result"
	"github.com/ehsatn/MAGIK/internal/xraytest"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func newTestApp(t *testing.T) AppModel {
	configPathOverride = filepath.Join(t.TempDir(), "config.json")
	t.Cleanup(func() {
		configPathOverride = ""
	})
	return NewApp("test")
}

func TestMenuOnlyShowsMainWorkflow(t *testing.T) {
	if len(menuEntries) != 4 {
		t.Fatalf("menu entries = %d, want 4", len(menuEntries))
	}
	if menuEntries[0].label != "Find Working IPs" {
		t.Fatalf("first menu item = %q, want Find Working IPs", menuEntries[0].label)
	}
	if menuEntries[1].label != "Retry Last Scan" {
		t.Fatalf("second menu item = %q, want Retry Last Scan", menuEntries[1].label)
	}
	for _, entry := range menuEntries {
		for _, removed := range []string{"Quick Scan", "Custom Scan", "Test IPs", "Discover Colos"} {
			if entry.label == removed {
				t.Fatalf("removed menu item %q is still visible", removed)
			}
		}
	}
}

func TestResolvePhase1OptionsUsesRandomCloudflareDefaults(t *testing.T) {
	m := newTestApp(t)
	m.configURL = "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=ws&host=example.com&path=%2F#test"
	m.configCountIdx = 1

	opts := m.resolvePhase1Options()
	if opts.count != 5000 {
		t.Fatalf("count = %d, want 5000", opts.count)
	}
	if opts.concurrency != 50 {
		t.Fatalf("concurrency = %d, want 50", opts.concurrency)
	}
	if opts.timeout.String() != "5s" {
		t.Fatalf("timeout = %s, want 5s", opts.timeout)
	}
	if opts.rawURL != m.configURL {
		t.Fatal("rawURL was not preserved")
	}
	if opts.fromFile {
		t.Fatal("fromFile = true, want random Cloudflare IPs")
	}
}

func TestResolvePhase1OptionsFromFile(t *testing.T) {
	m := newTestApp(t)
	m.configIPMode = 1
	opts := m.resolvePhase1Options()
	if !opts.fromFile {
		t.Fatal("fromFile = false, want true")
	}
}

func TestResolveConfigPortsMultiSelect(t *testing.T) {
	m := newTestApp(t)
	m.configURL = "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=ws&host=example.com&path=%2F#test"
	m.configSelectedPorts = map[int]bool{443: true, 8443: true}

	got := m.resolveConfigPorts()
	want := []string{"443", "8443"}
	parts := make([]string, len(got))
	for i, port := range got {
		parts[i] = strconv.Itoa(port)
	}
	if strings.Join(parts, ",") != strings.Join(want, ",") {
		t.Fatalf("ports = %v, want %v", got, want)
	}
}

func TestConfigPhase1TableColumnsStayAligned(t *testing.T) {
	m := newTestApp(t)
	m.page = PageConfigPhase1
	m.width = 120
	m.configPhase1Total = 1000
	m.configPhase1Results = []*result.Result{
		{
			IP:          net.ParseIP("172.67.145.191"),
			Port:        8443,
			ProbeMode:   "http",
			Latencies:   []time.Duration{223 * time.Millisecond, 223 * time.Millisecond, 223 * time.Millisecond},
			TLSOk:       true,
			HTTPStatus:  200,
			Colo:        "DME",
			Throughput:  1024,
			SpeedTested: true,
			Timestamp:   time.Now(),
		},
	}

	lines := strings.Split(ansiRE.ReplaceAllString(m.viewConfigPhase1(), ""), "\n")
	var headerLine, rowLine string
	for _, line := range lines {
		if strings.Contains(line, "ENDPOINT") && strings.Contains(line, "AVG(ms)") {
			headerLine = line
		}
		if strings.Contains(line, "172.67.145.191:8443") {
			rowLine = line
		}
	}
	if headerLine == "" {
		t.Fatal("missing Phase 1 table header")
	}
	if rowLine == "" {
		t.Fatal("missing Phase 1 table row")
	}
	for _, col := range []string{"LOSS", "AVG(ms)", "STATUS"} {
		if !strings.Contains(headerLine, col) {
			t.Fatalf("header missing %s: %q", col, headerLine)
		}
	}
	for _, tc := range []struct {
		header string
		value  string
	}{
		{header: "ENDPOINT", value: "172.67.145.191:8443"},
		{header: "COLO", value: "DME"},
	} {
		headerStart := strings.Index(headerLine, tc.header)
		valueStart := strings.Index(rowLine, tc.value)
		if headerStart < 0 || valueStart < 0 || headerStart != valueStart {
			t.Fatalf("%s column misaligned\nheader: %q\nrow:    %q", tc.header, headerLine, rowLine)
		}
	}
}

func TestLoadDefaultIPsFileFindsWorkingDirectoryFile(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	if err := writeIPsFile(filepath.Join(dir, "ips.txt"), []string{"104.18.1.1", "104.18.1.2"}); err != nil {
		t.Fatal(err)
	}

	ips, err := loadDefaultIPsFile()
	if err != nil {
		t.Fatal(err)
	}
	if len(ips) != 2 {
		t.Fatalf("loaded %d IPs, want 2", len(ips))
	}
}

func TestWorkingIPsOnlyIncludesSuccessfulValidationResults(t *testing.T) {
	got := workingIPs([]*xraytest.ValidationResult{
		{IP: "104.18.1.1", Success: true},
		{IP: "104.18.1.2", Success: false},
		{IP: "104.18.1.1", Success: true},
		nil,
		{IP: "", Success: true},
	})
	want := []string{"104.18.1.1"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("working IPs = %v, want %v", got, want)
	}
}

func TestWriteIPsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ips.txt")
	if err := writeIPsFile(path, []string{"104.18.1.1", "104.18.1.2"}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), "104.18.1.1\n104.18.1.2\n"; got != want {
		t.Fatalf("file contents = %q, want %q", got, want)
	}
}

func TestCopyWorkingIPsNoSuccesses(t *testing.T) {
	m := AppModel{
		configResults: []*xraytest.ValidationResult{
			{IP: "104.18.1.2", Success: false},
		},
	}
	if got := m.copyWorkingIPs(); got != "no working endpoints to copy" {
		t.Fatalf("message = %q", got)
	}
}

func TestFormatValidationSpeed(t *testing.T) {
	if got := formatValidationSpeed(0); got != "n/a" {
		t.Fatalf("zero throughput = %q, want n/a", got)
	}
	// 1.25 MiB/s ~= 10.5 Mbps
	if got := formatValidationSpeed(1.25 * 1024 * 1024); got != "10.5 Mbps" {
		t.Fatalf("throughput formatting = %q, want 10.5 Mbps", got)
	}
}

func TestFormatValidationLatency(t *testing.T) {
	if got := formatValidationLatency(250 * time.Millisecond); got != "250ms" {
		t.Fatalf("latency = %q, want 250ms", got)
	}
	if got := formatValidationLatency(1500 * time.Millisecond); got != "1.5s" {
		t.Fatalf("latency = %q, want 1.5s", got)
	}
}

func TestWorkingEndpointsIncludePorts(t *testing.T) {
	got := workingEndpoints([]*xraytest.ValidationResult{
		{IP: "104.18.1.1", Port: 443, Success: true},
		{IP: "104.18.1.1", Port: 8443, Success: true},
		{IP: "104.18.1.2", Port: 443, Success: false},
	})
	want := []string{"104.18.1.1:443", "104.18.1.1:8443"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("working endpoints = %v, want %v", got, want)
	}
}

func TestGenericScanCopyDoesNotExportHealthyIPs(t *testing.T) {
	m := AppModel{page: PageResults}
	next, _ := m.handleResultsKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	got := next.(AppModel).statusMsg
	if !strings.Contains(got, "Find Working IPs") {
		t.Fatalf("generic copy message = %q", got)
	}
}

func TestLoadIPsSubnets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ips.txt")

	// Test inputs
	lines := []string{
		"# Comment line",
		"10.0.0.1",       // IPv4 plain
		"2001:db8::1",    // IPv6 plain (should be ignored)
		"10.0.0.8/30",    // IPv4 small subnet (4 IPs: 10.0.0.8, 10.0.0.9, 10.0.0.10, 10.0.0.11)
		"2001:db8::/120", // IPv6 subnet (should be ignored)
		"192.168.0.0/16", // IPv4 large subnet (> 256 IPs, should sample exactly 256 IPs)
		"invalid-line",   // Should be ignored
		"10.0.0.0/99",    // Invalid CIDR (should return error)
	}

	// First, test with valid lines (excluding the invalid CIDR)
	if err := writeIPsFile(path, lines[:len(lines)-1]); err != nil {
		t.Fatal(err)
	}

	ips, err := loadIPs(path)
	if err != nil {
		t.Fatalf("loadIPs failed: %v", err)
	}

	// Verify plain IPv4 was loaded
	foundPlain := false
	for _, ip := range ips {
		if ip.String() == "10.0.0.1" {
			foundPlain = true
			break
		}
	}
	if !foundPlain {
		t.Error("expected to find 10.0.0.1 in loaded IPs")
	}

	// Verify IPv6 is completely ignored
	for _, ip := range ips {
		if ip.To4() == nil {
			t.Errorf("found IPv6 address %s, but IPv6 is not supported", ip.String())
		}
	}

	// Verify 10.0.0.8/30 subnet was fully expanded (4 IPs)
	subnetIPs := map[string]bool{
		"10.0.0.8":  false,
		"10.0.0.9":  false,
		"10.0.0.10": false,
		"10.0.0.11": false,
	}
	for _, ip := range ips {
		if _, ok := subnetIPs[ip.String()]; ok {
			subnetIPs[ip.String()] = true
		}
	}
	for ip, found := range subnetIPs {
		if !found {
			t.Errorf("expected to find subnet IP %s", ip)
		}
	}

	// Verify 192.168.0.0/16 large subnet was sampled (exactly 256 unique IPs within range)
	sampledCount := 0
	_, sampledNet, err := net.ParseCIDR("192.168.0.0/16")
	if err != nil {
		t.Fatal(err)
	}
	sampledIPs := make(map[string]bool)
	for _, ip := range ips {
		if sampledNet.Contains(ip) {
			sampledCount++
			sampledIPs[ip.String()] = true
		}
	}
	if sampledCount != 256 {
		t.Errorf("expected exactly 256 sampled IPs from 192.168.0.0/16, got %d", sampledCount)
	}
	if len(sampledIPs) != 256 {
		t.Errorf("expected 256 unique sampled IPs, got %d", len(sampledIPs))
	}

	// Verify invalid CIDR block returns an error
	if err := writeIPsFile(path, lines); err != nil {
		t.Fatal(err)
	}
	_, err = loadIPs(path)
	if err == nil {
		t.Error("expected error when loading invalid CIDR, but got nil")
	}
}

func TestAppConfigPersistence(t *testing.T) {
	tempDir := t.TempDir()

	oldAppData := os.Getenv("APPDATA")
	oldHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("APPDATA", oldAppData)
		os.Setenv("HOME", oldHome)
	})

	os.Setenv("APPDATA", tempDir)
	os.Setenv("HOME", tempDir)

	// Ensure config file doesn't exist initially
	path := getConfigFilePath()
	_ = os.Remove(path)

	// Loading should fall back to default config
	cfg := loadAppConfig()
	if cfg.LastConfig.CountIdx != 1 {
		t.Errorf("expected default CountIdx to be 1, got %d", cfg.LastConfig.CountIdx)
	}

	// Save custom config
	custom := AppConfig{
		LastConfig: SavedConfig{
			IPMode:        1,
			CountIdx:      3,
			CountCustom:   "9999",
			WorkersIdx:    2,
			WorkersCustom: "111",
			TimeoutIdx:    1,
			TimeoutCustom: "4s",
			Ports:         []int{8443, 2053},
			ConfigURL:     "vless://test-url",
			TopNIdx:       1,
			TopNCustom:    "5",
		},
	}

	if err := saveAppConfig(custom); err != nil {
		t.Fatalf("saveAppConfig failed: %v", err)
	}

	// Verify loaded config matches custom config
	loaded := loadAppConfig()
	if loaded.LastConfig.IPMode != 1 || loaded.LastConfig.CountCustom != "9999" || loaded.LastConfig.ConfigURL != "vless://test-url" {
		t.Errorf("loaded config does not match custom config: %+v", loaded.LastConfig)
	}

	// Verify applying saved config to model
	m := newTestApp(t)
	m.applySavedConfig(loaded.LastConfig)

	if m.configIPMode != 1 || m.configCountCustom != "9999" || m.configInput.Value() != "vless://test-url" {
		t.Errorf("model fields do not match applied saved config")
	}

	if !m.configSelectedPorts[8443] || !m.configSelectedPorts[2053] {
		t.Errorf("model configSelectedPorts does not match applied saved config: %v", m.configSelectedPorts)
	}
}

func TestVersionFormatting(t *testing.T) {
	m1 := AppModel{version: "v0.5.0-dirty"}
	view1 := m1.viewHome()
	if strings.Contains(view1, "vv0.5.0") {
		t.Error("home view contains double 'v' version prefix")
	}
	if !strings.Contains(view1, "  v0.5.0-dirty") {
		t.Errorf("home view does not contain expected version, view: %s", view1)
	}

	m2 := AppModel{version: "0.5.0"}
	view2 := m2.viewHome()
	if !strings.Contains(view2, "  v0.5.0") {
		t.Errorf("home view does not prepended version with 'v', view: %s", view2)
	}
}

func TestPhase1TargetTotalEnforcesCountLimit(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5", "10.0.0.6", "10.0.0.7", "10.0.0.8", "10.0.0.9", "10.0.0.10"}
	if err := writeIPsFile(filepath.Join(dir, "ips.txt"), ips); err != nil {
		t.Fatal(err)
	}

	m := newTestApp(t)
	m.configIPMode = 1 // From File
	m.configSelectedPorts = map[int]bool{443: true}

	total := m.phase1TargetTotal(5)
	if total != 5 {
		t.Errorf("expected target total to be capped at 5, got %d", total)
	}

	total = m.phase1TargetTotal(20)
	if total != 10 {
		t.Errorf("expected target total to be 10, got %d", total)
	}
}

func TestConfigOptionalCtrlXClearsInput(t *testing.T) {
	m := newTestApp(t)
	m.configInput.SetValue("vless://test-url-to-clear")
	m.page = PageConfigOptional
	m.configOptionalRow = 0

	// Simulate pressing ctrl+x
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	m2 := res.(AppModel)
	if m2.configInput.Value() != "" {
		t.Fatalf("expected configInput to be cleared, got %q", m2.configInput.Value())
	}

	// Also test speed URL input clear
	m2.configSpeedURLInput.SetValue("http://test-speed-url")
	m2.configOptionalRow = 3

	res2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyCtrlX})
	m3 := res2.(AppModel)
	if m3.configSpeedURLInput.Value() != "" {
		t.Fatalf("expected configSpeedURLInput to be cleared, got %q", m3.configSpeedURLInput.Value())
	}
}

func TestScanConfigF4TogglesRequireWS(t *testing.T) {
	m := newTestApp(t)
	m.page = PageScanConfig
	m.scanCfg.RequireWS = true

	// Simulate pressing f4
	res, _ := m.Update(tea.KeyMsg{Type: tea.KeyF4})
	m2 := res.(AppModel)
	if m2.scanCfg.RequireWS {
		t.Error("expected RequireWS to be false after pressing f4")
	}

	// Simulate pressing f4 again
	res2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyF4})
	m3 := res2.(AppModel)
	if !m3.scanCfg.RequireWS {
		t.Error("expected RequireWS to be true after pressing f4 again")
	}
}
