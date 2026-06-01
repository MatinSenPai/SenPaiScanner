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

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestMenuOnlyShowsMainWorkflow(t *testing.T) {
	if len(menuEntries) != 3 {
		t.Fatalf("menu entries = %d, want 3", len(menuEntries))
	}
	if menuEntries[0].label != "Find Working IPs" {
		t.Fatalf("first menu item = %q, want Find Working IPs", menuEntries[0].label)
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
	m := NewApp("test")
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
	m := NewApp("test")
	m.configIPMode = 1
	opts := m.resolvePhase1Options()
	if !opts.fromFile {
		t.Fatal("fromFile = false, want true")
	}
}

func TestResolveConfigPortsMultiSelect(t *testing.T) {
	m := NewApp("test")
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
	m := NewApp("test")
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

func TestConfigOptionalTypingTrojanURLKeepsURLInputFocused(t *testing.T) {
	raw := "trojan://jkh-password@example.com:443?security=tls&type=ws&host=example.com&path=%2F#test"
	m := NewApp("test")
	m.page = PageConfigOptional
	m.configOptionalRow = 0
	m.configInput.Focus()

	for _, r := range raw {
		next, _ := m.handleConfigOptionalKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(AppModel)
	}

	if got := m.configInput.Value(); got != raw {
		t.Fatalf("config input = %q, want %q", got, raw)
	}
	if m.configOptionalRow != 0 {
		t.Fatalf("config optional row = %d, want 0", m.configOptionalRow)
	}
}

func TestConfigOptionalArrowDownStillMovesToTopN(t *testing.T) {
	m := NewApp("test")
	m.page = PageConfigOptional
	m.configOptionalRow = 0
	m.configInput.Focus()

	next, _ := m.handleConfigOptionalKey(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(AppModel)

	if m.configOptionalRow != 1 {
		t.Fatalf("config optional row = %d, want 1", m.configOptionalRow)
	}
}
