package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/matinsenpai/senpaiscanner/internal/banner"
	"github.com/matinsenpai/senpaiscanner/internal/config"
	"github.com/matinsenpai/senpaiscanner/internal/logger"
	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

// ---------------------------------------------------------------------------
// Message types
// ---------------------------------------------------------------------------

// ResultMsg carries a completed probe result from the engine.
type ResultMsg struct {
	ScanID int64
	Result *result.Result
}

// StatsMsg carries live engine counters.
type StatsMsg struct {
	ScanID                            int64
	Tested, Healthy, Failed, InFlight int64
}

// DoneMsg signals the scan has finished.
type DoneMsg struct{ ScanID int64 }

// ErrorMsg carries a user-visible background task error.
type ErrorMsg struct {
	ScanID int64
	Text   string
}

// tickMsg drives banner animation and stat refresh.
type tickMsg time.Time

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F6821F"))

	styleSelected = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFE066")).
			Background(lipgloss.Color("#1A1A2E"))

	styleNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555"))

	styleGood = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#27AE60")).Bold(true)

	styleWarn = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F39C12"))

	styleBad = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E74C3C"))

	styleAccent = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F6821F")).Bold(true)

	styleHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444466")).Italic(true)

	styleHeader = lipgloss.NewStyle().
			Bold(true).Foreground(lipgloss.Color("#888888"))

	styleSep = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333"))
)

// ---------------------------------------------------------------------------
// ScanConfig holds form state.
// ---------------------------------------------------------------------------

type ScanConfig struct {
	Count       string
	Concurrency string
	Timeout     string
	Tries       string
	Port        string
	Mode        string // tcp|tls|http
	CIDR        string
	OutputFile  string
	ColoFilter  string
	SNI         string
	UseV4       bool
	UseV6       bool
	RequireWS   bool
}

func defaultScanConfig() ScanConfig {
	return ScanConfig{
		Count:       strconv.Itoa(config.ScanDefaults.Count),
		Concurrency: strconv.Itoa(config.ScanDefaults.Concurrency),
		Timeout:     config.ScanDefaults.Timeout.String(),
		Tries:       strconv.Itoa(config.ScanDefaults.Tries),
		Port:        strconv.Itoa(config.ScanDefaults.Port),
		Mode:        config.ScanDefaults.Mode,
		UseV4:       config.ScanDefaults.UseV4,
		UseV6:       config.ScanDefaults.UseV6,
		RequireWS:   true,
	}
}

// ---------------------------------------------------------------------------
// Quick Scan setup rows
// ---------------------------------------------------------------------------

type quickPreset struct {
	label string
	value string // empty = custom
}

var quickCountPresets = []quickPreset{
	{"5,000", "5000"},
	{"20,000", "20000"},
	{"100,000", "100000"},
	{"Custom", ""},
}

var quickWorkersPresets = []quickPreset{
	{"50  — default (restricted net)", "50"},
	{"100 — balanced", "100"},
	{"200 — fast (good connections)", "200"},
	{"Custom", ""},
}

var quickTimeoutPresets = []quickPreset{
	{"2s  — aggressive (fast net)", "2s"},
	{"3s  — balanced", "3s"},
	{"5s  — default (restricted net)", "5s"},
	{"Custom", ""},
}

// ---------------------------------------------------------------------------
// AppModel — root Bubble Tea model
// ---------------------------------------------------------------------------

type AppModel struct {
	page   Page
	width  int
	height int

	// animation
	bannerFrame int
	spinner     spinner.Model

	// home menu
	menuIdx int

	// scan config form
	scanCfg    ScanConfig
	modeIdx    int

	// live scan state
	activeScanID int64
	scanResults  []*result.Result
	sortBy       result.SortBy
	sortIdx      int
	scanStats    StatsMsg
	scanDone     bool
	scanStarted  time.Time
	scanTotal    int

	// scan with config
	configInput    textinput.Model
	configResults  []*xraytest.ValidationResult
	configScanning bool
	configDone     bool
	configTotal    int
	// config setup options
	configURL      string
	configCountIdx int // index into configCountValues
	configTopNIdx  int // index into configTopNValues
	configSetupRow int // 0=source, 1=count, 2=workers, 3=timeout, 4=ports
	// quick-scan-style pickers for Phase 1
	configWorkersIdx    int
	configTimeoutIdx    int
	configIPMode        int // 0=random Cloudflare IPs, 1=from ips.txt
	configCustomInput   textinput.Model
	configCustomMode    bool
	configCustomRow     int    // 1=count, 2=workers, 3=timeout, 5=topN custom, 6=min download custom, 7=speed size custom, 8=min upload custom
	configCountCustom   string // value when Custom count is selected
	configWorkersCustom string // value when Custom workers is selected
	configTimeoutCustom string // value when Custom timeout is selected
	configTopNCustom    string // value when Custom top N is selected
	configOptionalRow   int    // 0=config URL, 1=top N, 2=min download, 3=min upload, 4=speed size, 5=speed URL, 6=start scan
	configPortFocus     int
	configSelectedPorts map[int]bool
	// phase 1 state
	configPhase1Results []*result.Result
	configPhase1Done    bool
	configPhase1Only    bool // true when scan stops after Phase 1 (no config URL)
	configPhase1Stats   StatsMsg
	configPhase1Total   int // intended IP count for Phase 1 progress display
	liveResultPath      string

	configMinDownloadIdx     int
	configMinDownloadCustom  string
	configMinUploadIdx    int
	configMinUploadCustom string
	configSpeedURLInput   textinput.Model
	configSpeedSizeIdx    int
	configSpeedSizeCustom string
	configPhase2WorkersIdx    int
	configPhase2WorkersCustom string
	configUploadTest      bool
	ispInfo               string

	// log viewer
	logViewerLines []string

	// shared
	statusMsg string
	version   string
}

type menuEntry struct {
	label string
	desc  string
}

var menuEntries = []menuEntry{
	{"Find Working IPs", "scan Cloudflare IPs — config optional"},
	{"Retry Last Scan", "retry last scan with previous configuration"},
	{"About", ""},
	{"Quit", ""},
}

const menuLabelWidth = 16

const (
	menuFindWorking = 0
	menuRetryLast   = 1
	menuAbout       = 2
	menuQuit        = 3
)

var modes = []string{"tls", "tcp", "http"}

// SavedConfig represents the scan settings that can be persisted.
type SavedConfig struct {
	IPMode          int    `json:"ip_mode"`
	CountIdx        int    `json:"count_idx"`
	CountCustom     string `json:"count_custom"`
	WorkersIdx      int    `json:"workers_idx"`
	WorkersCustom   string `json:"workers_custom"`
	TimeoutIdx      int    `json:"timeout_idx"`
	TimeoutCustom   string `json:"timeout_custom"`
	Ports           []int  `json:"ports"`
	ConfigURL       string `json:"config_url"`
	TopNIdx         int    `json:"top_n_idx"`
	TopNCustom      string `json:"top_n_custom"`
	MinDownloadIdx     int    `json:"min_download_idx"`
	MinDownloadCustom  string `json:"min_download_custom"`
	MinUploadIdx    int    `json:"min_upload_idx"`
	MinUploadCustom string `json:"min_upload_custom"`
	SpeedURL        string `json:"speed_url"`
	SpeedSizeIdx    int    `json:"speed_size_idx"`
	SpeedSizeCustom string `json:"speed_size_custom"`
	DownloadTest    bool   `json:"download_test"`
	UploadTest      bool   `json:"upload_test"`
	RequireWS       bool   `json:"require_ws"`
	Phase2WorkersIdx    int    `json:"phase2_workers_idx"`
	Phase2WorkersCustom string `json:"phase2_workers_custom"`
}

// AppConfig wraps SavedConfig to allow for future settings.
type AppConfig struct {
	LastConfig SavedConfig `json:"last_config"`
}

var configPathOverride string

func getConfigFilePath() string {
	if configPathOverride != "" {
		return configPathOverride
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "senpaiscanner-config.json"
	}
	appDir := filepath.Join(dir, "senpaiscanner")
	_ = os.MkdirAll(appDir, 0755)
	return filepath.Join(appDir, "config.json")
}

func loadAppConfig() AppConfig {
	path := getConfigFilePath()
	b, err := os.ReadFile(path)
	if err != nil {
		return defaultAppConfig()
	}
	cfg := defaultAppConfig()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return defaultAppConfig()
	}
	return cfg
}

func saveAppConfig(cfg AppConfig) error {
	path := getConfigFilePath()
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func defaultAppConfig() AppConfig {
	return AppConfig{
		LastConfig: SavedConfig{
			IPMode:          0, // Random
			CountIdx:        1, // 5,000
			CountCustom:     "",
			WorkersIdx:      0, // 50
			WorkersCustom:   "",
			TimeoutIdx:      2, // 5s
			TimeoutCustom:   "",
			Ports:           nil,
			ConfigURL:       "",
			TopNIdx:         2, // 50
			TopNCustom:      "",
			MinDownloadIdx:  1, // None (index 1)
			MinDownloadCustom: "",
			MinUploadIdx:    0, // OFF
			MinUploadCustom: "",
			SpeedURL:        "",
			SpeedSizeIdx:    1, // 512 KB
			SpeedSizeCustom: "",
			RequireWS:       true,
			Phase2WorkersIdx: 1, // 10 (index 1)
			Phase2WorkersCustom: "",
		},
	}
}

func (m *AppModel) applySavedConfig(cfg SavedConfig) {
	m.configIPMode = cfg.IPMode
	m.configCountIdx = cfg.CountIdx
	m.configCountCustom = cfg.CountCustom
	m.configWorkersIdx = cfg.WorkersIdx
	m.configWorkersCustom = cfg.WorkersCustom
	m.configTimeoutIdx = cfg.TimeoutIdx
	m.configTimeoutCustom = cfg.TimeoutCustom
	m.configTopNIdx = cfg.TopNIdx
	m.configTopNCustom = cfg.TopNCustom
	m.configMinDownloadIdx = cfg.MinDownloadIdx
	m.configMinDownloadCustom = cfg.MinDownloadCustom
	m.configMinUploadIdx = cfg.MinUploadIdx
	m.configMinUploadCustom = cfg.MinUploadCustom
	m.configSpeedURLInput.SetValue(cfg.SpeedURL)
	m.configSpeedSizeIdx = cfg.SpeedSizeIdx
	m.configSpeedSizeCustom = cfg.SpeedSizeCustom
	m.configPhase2WorkersIdx = cfg.Phase2WorkersIdx
	m.configPhase2WorkersCustom = cfg.Phase2WorkersCustom
	m.configUploadTest = cfg.UploadTest
	m.configSelectedPorts = make(map[int]bool)
	for _, port := range cfg.Ports {
		m.configSelectedPorts[port] = true
	}
	if len(m.configSelectedPorts) == 0 {
		m.configSelectedPorts[0] = true
	}
	m.configInput.SetValue(cfg.ConfigURL)
	m.scanCfg.RequireWS = cfg.RequireWS
}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func NewApp(version string) AppModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#F6821F"))

	speedURLInput := textinput.New()
	speedURLInput.Placeholder = "empty = default (speed.cloudflare.com)"
	speedURLInput.CharLimit = 500
	speedURLInput.Width = 60

	m := AppModel{
		page:                PageHome,
		spinner:             sp,
		scanCfg:             defaultScanConfig(),
		version:             version,
		width:               120,
		height:              40,
		scanStarted:         time.Now(),
		configSpeedURLInput: speedURLInput,
	}

	// Config input for "Scan with Config"
	cfgInput := textinput.New()
	cfgInput.Placeholder = "vless://, trojan://, or vmess:// share URL"
	cfgInput.CharLimit = 2000
	cfgInput.Width = 60
	m.configInput = cfgInput

	cfgCustom := textinput.New()
	cfgCustom.Placeholder = "enter value"
	cfgCustom.CharLimit = 10
	cfgCustom.Width = 12
	m.configCustomInput = cfgCustom

	// Load configuration from file
	appCfg := loadAppConfig()
	m.applySavedConfig(appCfg.LastConfig)

	m.modeIdx = modeIndex(m.scanCfg.Mode)
	return m
}

func modeIndex(mode string) int {
	for i, candidate := range modes {
		if candidate == mode {
			return i
		}
	}
	return 0
}

// ---------------------------------------------------------------------------
// tea.Model interface
// ---------------------------------------------------------------------------

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		m.spinner.Tick,
		textinput.Blink,
		tea.EnableBracketedPaste,
		FetchMetaCmd(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case MetaMsg:
		if msg.ASOrganization != "" {
			if msg.Colo != "" {
				m.ispInfo = fmt.Sprintf("ISP: %s (%s)", msg.ASOrganization, msg.Colo)
			} else {
				m.ispInfo = fmt.Sprintf("ISP: %s", msg.ASOrganization)
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		m.bannerFrame++
		return m, tick()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ResultMsg:
		if msg.ScanID != m.activeScanID || msg.Result == nil {
			return m, nil
		}
		m.scanResults = append(m.scanResults, msg.Result)
		result.Sort(m.scanResults, m.sortBy)
		return m, nil

	case StatsMsg:
		if msg.ScanID == m.activeScanID {
			m.scanStats = msg
		}
		return m, nil

	case ErrorMsg:
		if msg.ScanID == m.activeScanID {
			m.statusMsg = msg.Text
		}
		return m, nil

	case DoneMsg:
		if msg.ScanID == m.activeScanID {
			m.scanDone = true
		}
		return m, nil

	case ConfigProgressMsg:
		m.configResults = append(m.configResults, msg.Result)
		m.configTotal = msg.Total
		return m, nil

	case ConfigDoneMsg:
		m.configScanning = false
		m.configDone = true
		sort.Slice(m.configResults, func(i, j int) bool {
			if m.configResults[i].Success != m.configResults[j].Success {
				return m.configResults[i].Success
			}
			if m.configResults[i].Throughput != m.configResults[j].Throughput {
				return m.configResults[i].Throughput > m.configResults[j].Throughput
			}
			return m.configResults[i].Latency < m.configResults[j].Latency
		})
		return m, nil

	case ConfigPhase1ResultMsg:
		m.configPhase1Results = append(m.configPhase1Results, msg.Result)
		return m, nil

	case ConfigPhase1ErrMsg:
		m.configScanning = false
		clearLiveResultWriter()
		m.page = PageScanWithConfig
		m.statusMsg = msg.Err
		return m, nil

	case ConfigPhase1DoneMsg:
		m.configPhase1Done = true
		if strings.TrimSpace(m.configURL) == "" {
			m.configPhase1Only = true
			if liveResultWriter != nil {
				liveResultWriter.FinishPhase1Only()
			}
			return m, nil
		}
		topN := m.resolveTopN()
		var topIPs []*result.Result
		if topN == 0 {
			topIPs = result.TopN(m.configPhase1Results, 0)
		} else {
			topIPs = result.TopN(m.configPhase1Results, topN)
		}
		m.configTotal = len(topIPs)
		// If Phase 1 found no healthy IPs, stay on the Phase 1 page and show
		// a clear "no results" message (Phase 2 would have nothing to do).
		if len(topIPs) == 0 {
			m.configPhase1Done = true
			m.page = PageConfigPhase2
			m.configScanning = false
			m.configDone = true
			return m, nil
		}
		// Start Phase 2 with top N IPs
		if liveResultWriter != nil {
			liveResultWriter.BeginPhase2()
		}
		m.page = PageConfigPhase2
		m.configScanning = true
		m.configDone = false
		m.configResults = nil
		return m, m.startConfigPhase2(topIPs)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m.updateFormInputs(msg)
}

// ---------------------------------------------------------------------------
// Key handling (dispatched by page)
// ---------------------------------------------------------------------------

func (m AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.page {
	case PageHome:
		return m.handleHomeKey(msg)
	case PageLiveScan:
		return m.handleLiveScanKey(msg)
	case PageResults:
		return m.handleResultsKey(msg)
	case PageAbout:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "enter" {
			m.page = PageHome
		}
		return m, nil
	case PageScanWithConfig:
		return m.handleScanSetupKey(msg)
	case PageConfigOptional:
		return m.handleConfigOptionalKey(msg)
	case PageConfigPhase1:
		return m.handleConfigPhase1Key(msg)
	case PageConfigPhase2:
		return m.handlePhase2Key(msg)
	case PageLogViewer:
		return m.handleLogViewerKey(msg)
	}
	return m, nil
}

func (m AppModel) handleHomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.menuIdx > 0 {
			m.menuIdx--
		}
	case "down", "j":
		if m.menuIdx < len(menuEntries)-1 {
			m.menuIdx++
		}
	case "enter":
		return m.selectMenuItem()
	}
	return m, nil
}

func (m AppModel) selectMenuItem() (tea.Model, tea.Cmd) {
	switch m.menuIdx {
	case menuFindWorking:
		m.page = PageScanWithConfig
		appCfg := loadAppConfig()
		m.applySavedConfig(appCfg.LastConfig)
		m.configInput.Blur()
		m.configResults = nil
		m.configScanning = false
		m.configDone = false
		m.configSetupRow = 0
		m.configOptionalRow = 0
		m.configPortFocus = 0
		m.configCustomMode = false
		m.configCustomInput.SetValue("")
		m.configCustomInput.Blur()
		m.configURL = ""
		m.configPhase1Only = false
		m.liveResultPath = ""
		clearLiveResultWriter()
		m.statusMsg = ""
		return m, nil
	case menuRetryLast:
		appCfg := loadAppConfig()
		m.applySavedConfig(appCfg.LastConfig)
		return m.launchPhase1FromOptional()
	case menuAbout:
		m.page = PageAbout
	case menuQuit:
		return m, tea.Quit
	}
	return m, nil
}

func (m AppModel) handleLiveScanKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q", "esc":
		if m.scanDone {
			m.page = PageResults
		} else {
			m.page = PageHome
			return m, CancelScanCmd()
		}
	case "s":
		m.sortIdx = (m.sortIdx + 1) % 5
		m.sortBy = result.SortBy(m.sortIdx)
		result.Sort(m.scanResults, m.sortBy)
	case "enter":
		if m.scanDone {
			m.page = PageResults
		}
	case "c":
		m.statusMsg = "use Find Working IPs to copy config-tested IPs"
	}
	return m, nil
}

func (m AppModel) handleResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q", "esc", "enter":
		m.page = PageHome
	case "s":
		m.sortIdx = (m.sortIdx + 1) % 5
		m.sortBy = result.SortBy(m.sortIdx)
		result.Sort(m.scanResults, m.sortBy)
	case "c":
		m.statusMsg = m.copyHealthyIPsToClipboard()
	}
	return m, nil
}

var clipboardWriteAll = clipboard.WriteAll
var osExecutable = os.Executable

// copyHealthyIPsToClipboard writes one IP per line to the system clipboard
// and returns a short status message to display to the user.
func (m AppModel) copyHealthyIPsToClipboard() string {
	top := result.TopN(m.scanResults, 0) // all healthy IPs, sorted by avg
	if len(top) == 0 {
		return "no healthy IPs to copy"
	}
	var sb strings.Builder
	for _, r := range top {
		ipStr := r.IP.String()
		if m.ispInfo != "" {
			sb.WriteString(fmt.Sprintf("%s # %s\n", ipStr, m.ispInfo))
		} else {
			sb.WriteString(ipStr + "\n")
		}
	}
	if err := clipboard.WriteAll(sb.String()); err != nil {
		return fmt.Sprintf("clipboard error: %v", err)
	}
	return fmt.Sprintf("✓ copied %d IPs to clipboard", len(top))
}

func (m AppModel) copyWorkingIPs() string {
	endpoints := workingEndpoints(m.configResults)
	if len(endpoints) == 0 {
		return "no working endpoints to copy"
	}
	return copyAndSaveIPs(endpoints, m.ispInfo)
}

func workingEndpoints(results []*xraytest.ValidationResult) []string {
	var endpoints []string
	seen := make(map[string]struct{})
	for _, r := range results {
		if r == nil || !r.Success || r.IP == "" {
			continue
		}
		endpoint := formatEndpoint(r.IP, r.Port)
		if _, ok := seen[endpoint]; ok {
			continue
		}
		seen[endpoint] = struct{}{}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

func formatEndpoint(ip string, port int) string {
	if port <= 0 {
		return ip
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

func formatValidationSpeed(throughput float64) string {
	if throughput <= 0 {
		return "n/a"
	}
	mbps := throughput * 8 / 1_000_000
	if mbps >= 100 {
		return fmt.Sprintf("%.0f Mbps", mbps)
	}
	return fmt.Sprintf("%.1f Mbps", mbps)
}

func formatValidationLatency(latency time.Duration) string {
	if latency <= 0 {
		return "—"
	}
	ms := latency.Milliseconds()
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", latency.Seconds())
}

func copyAndSaveIPs(ips []string, isp string) string {
	suffixed := make([]string, len(ips))
	for i, ep := range ips {
		if isp != "" {
			suffixed[i] = fmt.Sprintf("%s # %s", ep, isp)
		} else {
			suffixed[i] = ep
		}
	}
	text := strings.Join(suffixed, "\n") + "\n"
	clipErr := clipboardWriteAll(text)
	path, fileErr := writeIPsBesideExecutable(suffixed)

	switch {
	case clipErr == nil && fileErr == nil:
		return fmt.Sprintf("copied %d working endpoints; saved to %s", len(ips), path)
	case clipErr != nil && fileErr == nil:
		return fmt.Sprintf("clipboard failed; saved %d working endpoints to %s", len(ips), path)
	case clipErr == nil && fileErr != nil:
		return fmt.Sprintf("copied %d working endpoints; save failed: %v", len(ips), fileErr)
	default:
		return fmt.Sprintf("copy failed: %v; save failed: %v", clipErr, fileErr)
	}
}

func writeIPsBesideExecutable(ips []string) (string, error) {
	exe, err := osExecutable()
	dir := ""
	if err == nil {
		dir = filepath.Dir(exe)
	}
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			dir = "."
		}
	}
	path := filepath.Join(dir, "working_ips.txt")
	if err := writeIPsFile(path, ips); err == nil {
		return path, nil
	}

	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return path, wdErr
	}
	fallback := filepath.Join(wd, "working_ips.txt")
	if fallback == path {
		err := writeIPsFile(fallback, ips)
		return fallback, err
	}
	err = writeIPsFile(fallback, ips)
	return fallback, err
}

func writeIPsFile(path string, ips []string) error {
	text := strings.Join(ips, "\n")
	if text != "" {
		text += "\n"
	}
	return os.WriteFile(path, []byte(text), 0644)
}

// updateFormInputs forwards non-key messages (e.g. paste events, resize) to
// the currently focused text input only.
func (m AppModel) updateFormInputs(msg tea.Msg) (AppModel, tea.Cmd) {
	var cmds []tea.Cmd

	if m.page == PageScanWithConfig && !m.configScanning && !m.configDone {
		if m.configCustomMode {
			var cmd tea.Cmd
			m.configCustomInput, cmd = m.configCustomInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.page == PageConfigOptional {
		if m.configCustomMode {
			var cmd tea.Cmd
			m.configCustomInput, cmd = m.configCustomInput.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.configOptionalRow == 0 {
			var cmd tea.Cmd
			m.configInput, cmd = m.configInput.Update(msg)
			cmds = append(cmds, cmd)
		} else if m.configOptionalRow == 3 {
			var cmd tea.Cmd
			m.configSpeedURLInput, cmd = m.configSpeedURLInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (m AppModel) View() string {
	switch m.page {
	case PageHome:
		return m.viewHome()
	case PageLiveScan:
		return m.viewLiveScan()
	case PageResults:
		return m.viewResults()
	case PageAbout:
		return m.viewAbout()
	case PageScanWithConfig:
		return m.viewScanWithConfig()
	case PageConfigOptional:
		return m.viewConfigOptional()
	case PageConfigPhase1:
		return m.viewConfigPhase1()
	case PageConfigPhase2:
		return m.viewScanWithConfig()
	case PageLogViewer:
		return m.viewLogViewer()
	}
	return ""
}

// ---------------------------------------------------------------------------
// Home page
// ---------------------------------------------------------------------------

func (m AppModel) viewHome() string {
	var sb strings.Builder

	// Animated banner (shrink if terminal too narrow)
	art := banner.Render(m.bannerFrame / 2)
	sb.WriteString(art)
	sb.WriteRune('\n')

	ver := m.version
	if !strings.HasPrefix(strings.ToLower(ver), "v") {
		ver = "v" + ver
	}
	sb.WriteString(styleDim.Render(fmt.Sprintf("  %s", ver)))
	sb.WriteRune('\n')
	sb.WriteString(styleDim.Render(fmt.Sprintf("  config: %s", getConfigFilePath())))
	sb.WriteString("\n\n")

	// ISP info — prominent display
	if m.ispInfo != "" {
		sb.WriteString(fmt.Sprintf("  %s  %s\n\n",
			styleAccent.Render("🌐"),
			styleAccent.Render(m.ispInfo),
		))
	}

	// Menu
	for i, item := range menuEntries {
		cursor := "  "
		labelStyle := styleNormal
		if i == m.menuIdx {
			cursor = styleAccent.Render("▶ ")
			labelStyle = styleSelected
		}

		line := "  " + cursor + labelStyle.Render(fmt.Sprintf("%-*s", menuLabelWidth, item.label))
		if item.desc != "" {
			line += "  " + styleDim.Render(item.desc)
		}
		sb.WriteString(line)
		sb.WriteRune('\n')
	}

	sb.WriteRune('\n')
	sb.WriteString(styleHint.Render("  ↑/↓ navigate   enter select   q quit"))
	sb.WriteRune('\n')

	return sb.String()
}

// ---------------------------------------------------------------------------
// Live Scan page
// ---------------------------------------------------------------------------

func (m AppModel) viewLiveScan() string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("\n  ⚡  Live Scan\n"))
	sb.WriteString(fmt.Sprintf("%s\n\n", styleSep.Render("  "+strings.Repeat("─", minInt(m.width-4, 70)))))

	// Stats row
	elapsed := time.Since(m.scanStarted).Round(time.Second)
	rateStr := "—"
	if elapsed.Seconds() > 0 && m.scanStats.Tested > 0 {
		rateStr = fmt.Sprintf("%.0f/s", float64(m.scanStats.Tested)/elapsed.Seconds())
	}

	icon := m.spinner.View()
	if m.scanDone {
		icon = styleGood.Render("✓")
	}

	progBar := ""
	if m.scanTotal > 0 {
		pct := float64(m.scanStats.Tested) / float64(m.scanTotal) * 100
		bw := 22
		filled := int(pct / 100 * float64(bw))
		progBar = "  [" + styleAccent.Render(strings.Repeat("█", filled)) +
			styleDim.Render(strings.Repeat("░", bw-filled)) + "]" +
			fmt.Sprintf(" %.0f%%", pct)
	}

	sb.WriteString(fmt.Sprintf("  %s  tested: %s  healthy: %s  failed: %s  flying: %s  rate: %s  %s%s\n\n",
		icon,
		styleAccent.Render(fmt.Sprintf("%d", m.scanStats.Tested)),
		styleGood.Render(fmt.Sprintf("%d", m.scanStats.Healthy)),
		styleBad.Render(fmt.Sprintf("%d", m.scanStats.Failed)),
		styleDim.Render(fmt.Sprintf("%d", m.scanStats.InFlight)),
		styleDim.Render(rateStr),
		styleDim.Render(elapsed.String()),
		progBar,
	))

	// Table header
	hdr := fmt.Sprintf("  %-18s  %7s  %9s  %8s  %9s  %5s  %-6s",
		"IP", "LOSS", "AVG(ms)", "JTR(ms)", "DL(KB/s)", "TLS", "COLO")
	sb.WriteString(fmt.Sprintf("%s\n%s\n", styleHeader.Render(hdr), styleSep.Render("  "+strings.Repeat("─", 72))))

	maxRows := m.height - 14
	if maxRows < 3 {
		maxRows = 3
	}
	rows := m.scanResults
	if len(rows) > maxRows {
		rows = rows[:maxRows]
	}

	for _, r := range rows {
		tlsIcon := styleBad.Render("✗")
		if r.TLSOk {
			tlsIcon = styleGood.Render("✓")
		}
		colo := r.Colo
		if colo == "" {
			colo = "—"
		}
		line := fmt.Sprintf("  %-18s  %6.1f%%  %9.2f  %8.2f  %9.1f  %5s  %-6s",
			r.IP.String(), r.Loss(),
			float64(r.Avg().Milliseconds()),
			float64(r.Jitter().Milliseconds()),
			r.Throughput/1024,
			tlsIcon, colo)

		switch {
		case r.IsHealthy() && r.Loss() == 0 && r.Avg().Milliseconds() < 200:
			sb.WriteString(fmt.Sprintf("%s\n", styleGood.Render(line)))
		case !r.IsHealthy():
			sb.WriteString(fmt.Sprintf("%s\n", styleBad.Render(line)))
		default:
			sb.WriteString(fmt.Sprintf("%s\n", styleWarn.Render(line)))
		}
	}

	sb.WriteRune('\n')
	sortNames := []string{"avg", "loss", "jitter", "colo", "speed"}
	hint := fmt.Sprintf("  s sort: %s   c copy IPs   q/esc cancel", sortNames[m.sortIdx%5])
	if m.scanDone {
		hint = fmt.Sprintf("  s sort: %s   c copy IPs   enter/q view results", sortNames[m.sortIdx%5])
	}
	if m.statusMsg != "" {
		sb.WriteString(styleGood.Render("  "+m.statusMsg) + "\n")
	}
	sb.WriteString(styleHint.Render(hint))
	return sb.String()
}

// ---------------------------------------------------------------------------
// Results page
// ---------------------------------------------------------------------------

func (m AppModel) viewResults() string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("\n  ✅  Scan Results\n"))
	sb.WriteString(fmt.Sprintf("%s\n\n", styleSep.Render("  "+strings.Repeat("─", 60))))

	top := result.TopN(m.scanResults, 20)
	if len(top) == 0 {
		sb.WriteString(styleWarn.Render("  No healthy IPs found. Try raising timeout, lowering workers, or using a different SNI.\n"))
	} else {
		hdr := fmt.Sprintf("  %-18s  %7s  %9s  %8s  %9s  %5s  %-6s",
			"IP", "LOSS", "AVG(ms)", "JTR(ms)", "DL(KB/s)", "TLS", "COLO")
		sb.WriteString(fmt.Sprintf("%s\n%s\n", styleHeader.Render(hdr), styleSep.Render("  "+strings.Repeat("─", 72))))

		for i, r := range top {
			tlsIcon := "✗"
			if r.TLSOk {
				tlsIcon = "✓"
			}
			colo := r.Colo
			if colo == "" {
				colo = "—"
			}
			rank := styleAccent.Render(fmt.Sprintf(" %2d. ", i+1))
			line := fmt.Sprintf("%-18s  %6.1f%%  %9.2f  %8.2f  %9.1f  %5s  %-6s",
				r.IP.String(), r.Loss(),
				float64(r.Avg().Milliseconds()),
				float64(r.Jitter().Milliseconds()),
				r.Throughput/1024,
				tlsIcon, colo)
			sb.WriteString(fmt.Sprintf("%s%s\n", rank, styleGood.Render(line)))
		}
	}

	total := len(m.scanResults)
	healthy := 0
	for _, r := range m.scanResults {
		if r.IsHealthy() {
			healthy++
		}
	}
	sb.WriteString("\n")
	sb.WriteString(styleDim.Render(fmt.Sprintf("  Total probed: %d   healthy: %d   unhealthy: %d\n", total, healthy, total-healthy)))
	if m.scanCfg.OutputFile != "" {
		sb.WriteString(styleDim.Render(fmt.Sprintf("  Saved → %s\n", m.scanCfg.OutputFile)))
	}
	sb.WriteString("\n")
	if m.statusMsg != "" {
		sb.WriteString(styleGood.Render("  "+m.statusMsg) + "\n")
	}
	sb.WriteString(styleHint.Render("  s sort   c copy IPs   enter/q home menu"))
	return sb.String()
}

// ---------------------------------------------------------------------------
// Live Colos page
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// About page
// ---------------------------------------------------------------------------

func (m AppModel) viewAbout() string {
	var sb strings.Builder
	sb.WriteString(banner.Render(m.bannerFrame / 2))
	sb.WriteRune('\n')
	sb.WriteString(styleTitle.Render("  SenPai Scanner\n"))
	sb.WriteString(styleDim.Render(fmt.Sprintf("  version %s", m.version)))
	sb.WriteString("\n\n")
	sb.WriteString(styleNormal.Render("  A Cloudflare IP scanner built for high-latency, restricted networks."))
	sb.WriteRune('\n')

	sb.WriteString(styleNormal.Render("  Probes Cloudflare's edge nodes via TCP/TLS/HTTP, measures loss,"))
	sb.WriteRune('\n')

	sb.WriteString(styleNormal.Render("  jitter, and identifies the colo (PoP) behind each IP."))
	sb.WriteString("\n\n")

	sb.WriteString(styleDim.Render("  github.com/matinsenpai/senpaiscanner"))
	sb.WriteString("\n\n")
	sb.WriteString(styleHint.Render("  enter/q back"))
	return sb.String()
}

// ---------------------------------------------------------------------------
// Exported helpers for non-TUI callers
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Command factories (implemented in cmds.go)
// ---------------------------------------------------------------------------

// StartScanCmd, CancelScanCmd, StartTestCmd, StartColosCmd are defined in cmds.go.

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func scanPulse(frame int) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return frames[frame%len(frames)]
}

func scanWave(frame, width int) string {
	if width < 8 {
		width = 8
	}
	pos := frame % width
	var sb strings.Builder
	for i := 0; i < width; i++ {
		switch {
		case i == pos:
			sb.WriteString(styleAccent.Render("●"))
		case (i+1)%4 == 0:
			sb.WriteString(styleDim.Render("·"))
		default:
			sb.WriteString(styleDim.Render("─"))
		}
	}
	return sb.String()
}

func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "config"
	}
	parts := make([]string, len(ports))
	for i, port := range ports {
		parts[i] = strconv.Itoa(port)
	}
	return strings.Join(parts, ",")
}

func (m AppModel) selectedPortSet() map[int]bool {
	if len(m.configSelectedPorts) == 0 {
		return map[int]bool{0: true}
	}
	out := make(map[int]bool, len(m.configSelectedPorts))
	for port, on := range m.configSelectedPorts {
		if on {
			out[port] = true
		}
	}
	if len(out) == 0 {
		out[0] = true
	}
	return out
}

func (m *AppModel) toggleFocusedConfigPort() {
	if m.configPortFocus < 0 || m.configPortFocus >= len(configPortChoices) {
		return
	}
	port := configPortChoices[m.configPortFocus].port
	if m.configSelectedPorts == nil {
		m.configSelectedPorts = map[int]bool{0: true}
	}
	if port == 0 {
		m.configSelectedPorts = map[int]bool{0: true}
		return
	}
	delete(m.configSelectedPorts, 0)
	m.configSelectedPorts[port] = !m.configSelectedPorts[port]
	if !m.configSelectedPorts[port] {
		delete(m.configSelectedPorts, port)
	}
	if len(m.configSelectedPorts) == 0 {
		m.configSelectedPorts[0] = true
	}
}

// ---------------------------------------------------------------------------
// Scan with Config page
// ---------------------------------------------------------------------------

func (m AppModel) viewScanWithConfig() string {
	var sb strings.Builder

	title := "\n  ⚡  Find Working IPs"
	if m.ispInfo != "" {
		title += "  " + styleAccent.Render(fmt.Sprintf("[%s]", m.ispInfo))
	}
	sb.WriteString(title + "\n")
	sb.WriteString(fmt.Sprintf("%s\n\n", styleSep.Render("  "+strings.Repeat("─", minInt(m.width-4, 70)))))

	if !m.configScanning && !m.configDone {
		// helper: render a preset pill row
		renderPills := func(labels []string, selected int) {
			for i, label := range labels {
				if i == selected {
					sb.WriteString(styleSelected.Render(" " + label + " "))
				} else {
					sb.WriteString(styleNormal.Render("  " + label + "  "))
				}
				if i < len(labels)-1 {
					sb.WriteString(styleDim.Render("│"))
				}
			}
		}

		rowLabel := func(row int, text string) {
			if m.configSetupRow == row {
				sb.WriteString(styleAccent.Render(text))
			} else {
				sb.WriteString(styleDim.Render(text))
			}
		}
		renderMultiPorts := func() {
			enabled := m.selectedPortSet()
			for i, choice := range configPortChoices {
				label := choice.label
				if enabled[choice.port] {
					label = "✓ " + label
				} else {
					label = "  " + label
				}
				if i == m.configPortFocus && m.configSetupRow == 4 {
					sb.WriteString(styleSelected.Render(" " + label + " "))
				} else if enabled[choice.port] {
					sb.WriteString(styleGood.Render(" " + label + " "))
				} else {
					sb.WriteString(styleNormal.Render(" " + label + " "))
				}
				if i < len(configPortChoices)-1 {
					sb.WriteString(styleDim.Render("│"))
				}
			}
		}

		// Row 0: Source
		rowLabel(0, "  Source ")
		sb.WriteString(" ")
		renderPills(configIPModeLabels, m.configIPMode)
		sb.WriteString("\n")
		if m.configIPMode == 0 {
			sb.WriteString(styleDim.Render("            random Cloudflare IPv4 IPs") + "\n\n")
		} else {
			sb.WriteString(styleDim.Render("            read custom IPs from ips.txt next to the app or working directory") + "\n\n")
		}

		// Row 1: Count
		rowLabel(1, "  Count  ")
		sb.WriteString(" ")
		renderPills(configCountLabels, m.configCountIdx)
		sb.WriteString("\n")
		if m.configCustomMode && m.configCustomRow == 1 {
			sb.WriteString(styleAccent.Render("            custom count: ") + m.configCustomInput.View() + "\n\n")
		} else if configCountValues[m.configCountIdx] == 0 && m.configCountCustom != "" {
			sb.WriteString(styleDim.Render(fmt.Sprintf("            caps the number of random Cloudflare IPs generated and scanned  (custom: %s)", m.configCountCustom)) + "\n\n")
		} else if m.configIPMode == 1 {
			sb.WriteString(styleDim.Render("            caps the number of custom IPs loaded from ips.txt") + "\n\n")
		} else {
			sb.WriteString(styleDim.Render("            caps the number of random Cloudflare IPs generated and scanned") + "\n\n")
		}

		// Row 2: Workers
		rowLabel(2, "  Workers")
		sb.WriteString(" ")
		renderPills(quickWorkersLabels(), m.configWorkersIdx)
		sb.WriteString("\n")
		if m.configCustomMode && m.configCustomRow == 2 {
			sb.WriteString(styleAccent.Render("            custom workers: ") + m.configCustomInput.View() + "\n\n")
		} else if quickWorkersPresets[m.configWorkersIdx].value == "" && m.configWorkersCustom != "" {
			sb.WriteString(styleDim.Render(fmt.Sprintf("            concurrent probes  (custom: %s)", m.configWorkersCustom)) + "\n\n")
		} else {
			sb.WriteString(styleDim.Render("            concurrent probes") + "\n\n")
		}

		// Row 3: Timeout
		rowLabel(3, "  Timeout")
		sb.WriteString(" ")
		renderPills(quickTimeoutLabels(), m.configTimeoutIdx)
		sb.WriteString("\n")
		if m.configCustomMode && m.configCustomRow == 3 {
			sb.WriteString(styleAccent.Render("            custom timeout: ") + m.configCustomInput.View() + "\n\n")
		} else if quickTimeoutPresets[m.configTimeoutIdx].value == "" && m.configTimeoutCustom != "" {
			sb.WriteString(styleDim.Render(fmt.Sprintf("            per-probe deadline  (custom: %s)", m.configTimeoutCustom)) + "\n\n")
		} else {
			sb.WriteString(styleDim.Render("            per-probe deadline") + "\n\n")
		}

		// Row 4: Ports
		rowLabel(4, "  Ports  ")
		sb.WriteString(" ")
		renderMultiPorts()
		sb.WriteString("\n")
		sb.WriteString(styleDim.Render("            space toggles a port; enter proceeds to next step") + "\n\n")

		// Row 5: WebSocket
		rowLabel(5, "  WebSocket")
		sb.WriteString(" ")
		wss := styleGood.Render("ON")
		if !m.scanCfg.RequireWS {
			wss = styleBad.Render("OFF")
		}
		sb.WriteString(wss + "\n")
		sb.WriteString(styleDim.Render("            require a successful WebSocket check in Phase 1 (left/right toggle)") + "\n\n")

		// Row 6: Next Step
		sb.WriteString("  ")
		if m.configSetupRow == 6 {
			sb.WriteString(styleSelected.Render(" [ Next Step ] ") + "\n\n")
		} else {
			sb.WriteString(styleNormal.Render("   [ Next Step ] ") + "\n\n")
		}

		hint := "  ↑/↓ row   ←/→ option   enter next/confirm   esc back"
		if m.configSetupRow == 4 {
			hint = "  ↑/↓ row   ←/→ focus   space toggle port   enter next   esc back"
		} else if m.configSetupRow == 6 {
			hint = "  ↑/↓ row   enter next step   esc back"
		}
		if m.configCustomMode {
			hint = "  type value   enter confirm   esc cancel"
		}
		sb.WriteString(styleHint.Render(hint) + "\n")
		if m.statusMsg != "" {
			sb.WriteString(styleWarn.Render("  ⚠  "+m.statusMsg) + "\n")
		}
		return sb.String()
	}

	// Stats row — if Phase 1 found no candidates, show a clear message instead
	// of a fake "0/10" counter.
	if m.configTotal == 0 && m.configDone {
		sb.WriteString(fmt.Sprintf("  %s  %s\n\n",
			styleGood.Render("✓"),
			styleBad.Render("No working candidates found"),
		))
		sb.WriteString(styleHint.Render("  esc back") + "\n")
		return sb.String()
	}

	done := len(m.configResults)
	total := m.configTotal
	success := m.configSuccessCount()
	failed := m.configFailCount()

	icon := m.spinner.View()
	if m.configDone {
		icon = styleGood.Render("✓")
	}

	// Progress bar
	pct := 0.0
	if total > 0 {
		pct = float64(done) / float64(total) * 100
	}
	bw := 22
	filled := int(pct / 100 * float64(bw))
	progBar := "[" + styleAccent.Render(strings.Repeat("█", filled)) +
		styleDim.Render(strings.Repeat("░", bw-filled)) + "]" +
		fmt.Sprintf(" %.0f%%", pct)

	sb.WriteString(fmt.Sprintf("  %s  tested: %s  working: %s  failed: %s  %s\n\n",
		icon,
		styleAccent.Render(fmt.Sprintf("%d/%d", done, total)),
		styleGood.Render(fmt.Sprintf("%d", success)),
		styleBad.Render(fmt.Sprintf("%d", failed)),
		progBar,
	))
	if !m.configDone {
		sb.WriteString(fmt.Sprintf("  %s  xray validating candidates (%d workers)  %s\n\n",
			styleAccent.Render(scanPulse(m.bannerFrame)),
			m.resolvePhase2Workers(),
			scanWave(m.bannerFrame+5, 32),
		))
	}

	// Table header
	uploadCol := m.configUploadTest
	hdr := fmt.Sprintf("  %-22s  %-8s  %8s", "ENDPOINT", "TYPE", "SPEED")
	if uploadCol {
		hdr += fmt.Sprintf("  %8s", "UPLOAD")
	}
	hdr += fmt.Sprintf("  %8s  %6s", "LATENCY", "STATUS")
	sepLen := 64
	if uploadCol {
		sepLen += 10
	}
	sb.WriteString(fmt.Sprintf("%s\n%s\n", styleHeader.Render(hdr), styleSep.Render("  "+strings.Repeat("─", sepLen))))

	// Results
	maxRows := m.height - 12
	if maxRows < 3 {
		maxRows = 3
	}
	rows := m.configResults
	if len(rows) > maxRows {
		rows = rows[len(rows)-maxRows:]
	}

	for i := len(rows) - 1; i >= 0; i-- {
		r := rows[i]
		if r.Success {
			line := fmt.Sprintf("  %-22s  %-8s  %8s", formatEndpoint(r.IP, r.Port), r.Transport, formatValidationSpeed(r.Throughput))
			if uploadCol {
				if r.UploadThroughput > 0 {
					line += fmt.Sprintf("  %8s", formatValidationSpeed(r.UploadThroughput))
				} else {
					line += fmt.Sprintf("  %8s", "—")
				}
			}
			line += fmt.Sprintf("  %8s  %6s", formatValidationLatency(r.Latency), "✓")
			sb.WriteString(styleGood.Render(line) + "\n")
		} else {
			errMsg := r.Error
			if len(errMsg) > 20 {
				errMsg = errMsg[:20] + "…"
			}
			line := fmt.Sprintf("  %-22s  %-8s  %9s", formatEndpoint(r.IP, r.Port), r.Transport, "—")
			if uploadCol {
				line += fmt.Sprintf("  %8s", "—")
			}
			line += fmt.Sprintf("  %8s  %6s", "—", "✗")
			sb.WriteString(styleBad.Render(line) + "\n")
		}
	}

	sb.WriteRune('\n')
	if m.statusMsg != "" {
		sb.WriteString(styleGood.Render("  "+m.statusMsg) + "\n")
	}
	if m.configDone {
		hint := "  c copy endpoints   e export configs   enter/q home menu"
		if scanLog != nil && scanLog.HasFailures() {
			hint += fmt.Sprintf("   l view %s", scanLog.FailureSummary())
		}
		if m.liveResultPath != "" {
			hint += "\n" + styleDim.Render("  live results → "+m.liveResultPath)
		}
		sb.WriteString(styleHint.Render(hint) + "\n")
	} else {
		if m.liveResultPath != "" {
			sb.WriteString(styleDim.Render("  live results → "+m.liveResultPath) + "\n")
		}
	}

	return sb.String()
}

func (m AppModel) configSuccessCount() int {
	count := 0
	for _, r := range m.configResults {
		if r.Success {
			count++
		}
	}
	return count
}

func (m AppModel) configFailCount() int {
	count := 0
	for _, r := range m.configResults {
		if !r.Success {
			count++
		}
	}
	return count
}

// handleScanSetupKey handles keys on the PageScanWithConfig setup screen
// (Source, Count, Workers, Timeout, Ports, WebSocket rows).
func (m AppModel) handleScanSetupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// --- Custom input mode: route all keys to the active text input ---
	if m.configCustomMode {
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(m.configCustomInput.Value())
			switch m.configCustomRow {
			case 1:
				m.configCountCustom = val
			case 2:
				m.configWorkersCustom = val
			case 3:
				m.configTimeoutCustom = val
			}
			m.configCustomMode = false
			m.configCustomInput.Blur()
			if m.configSetupRow < 6 {
				m.configSetupRow++
			}
			return m, nil
		case "esc":
			m.configCustomMode = false
			m.configCustomInput.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.configCustomInput, cmd = m.configCustomInput.Update(msg)
		return m, cmd
	}

	// --- Global keys ---
	switch msg.String() {
	case "esc":
		m.page = PageHome
		return m, nil
	}

	// --- Setup navigation (Source → Count → Workers → Timeout → Ports → WebSocket) ---
	const maxRow = 6

	configNavLeft := func() {
		switch m.configSetupRow {
		case 0:
			if m.configIPMode > 0 {
				m.configIPMode--
			}
		case 1:
			if m.configCountIdx > 0 {
				m.configCountIdx--
			}
		case 2:
			if m.configWorkersIdx > 0 {
				m.configWorkersIdx--
			}
		case 3:
			if m.configTimeoutIdx > 0 {
				m.configTimeoutIdx--
			}
		case 4:
			if m.configPortFocus > 0 {
				m.configPortFocus--
			}
		case 5:
			m.scanCfg.RequireWS = !m.scanCfg.RequireWS
		}
	}
	configNavRight := func() {
		switch m.configSetupRow {
		case 0:
			if m.configIPMode < len(configIPModeLabels)-1 {
				m.configIPMode++
			}
		case 1:
			if m.configCountIdx < len(configCountValues)-1 {
				m.configCountIdx++
			}
		case 2:
			if m.configWorkersIdx < len(quickWorkersPresets)-1 {
				m.configWorkersIdx++
			}
		case 3:
			if m.configTimeoutIdx < len(quickTimeoutPresets)-1 {
				m.configTimeoutIdx++
			}
		case 4:
			if m.configPortFocus < len(configPortChoices)-1 {
				m.configPortFocus++
			}
		case 5:
			m.scanCfg.RequireWS = !m.scanCfg.RequireWS
		}
	}

	switch msg.String() {
	case "up", "k":
		if m.configSetupRow > 0 {
			m.configSetupRow--
		}
		return m, nil
	case "down", "j":
		if m.configSetupRow < maxRow {
			m.configSetupRow++
		}
		return m, nil
	case "left", "h", "right", "l":
		if msg.String() == "left" || msg.String() == "h" {
			configNavLeft()
		} else {
			configNavRight()
		}
		return m, nil
	case " ":
		if m.configSetupRow == 4 {
			m.toggleFocusedConfigPort()
			return m, nil
		}
	case "enter":
		if m.configSetupRow == 6 {
			m.statusMsg = ""
			m.page = PageConfigOptional
			m.configOptionalRow = 0
			m.configInput.Focus()
			return m, textinput.Blink
		}
		if m.configSetupRow == 1 && configCountValues[m.configCountIdx] == 0 {
			m.configCustomMode = true
			m.configCustomRow = 1
			m.configCustomInput.SetValue(m.configCountCustom)
			m.configCustomInput.Placeholder = "e.g. 50000"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		if m.configSetupRow == 2 && quickWorkersPresets[m.configWorkersIdx].value == "" {
			m.configCustomMode = true
			m.configCustomRow = 2
			m.configCustomInput.SetValue(m.configWorkersCustom)
			m.configCustomInput.Placeholder = "e.g. 150"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		if m.configSetupRow == 3 && quickTimeoutPresets[m.configTimeoutIdx].value == "" {
			m.configCustomMode = true
			m.configCustomRow = 3
			m.configCustomInput.SetValue(m.configTimeoutCustom)
			m.configCustomInput.Placeholder = "e.g. 7s"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		m.configSetupRow++
		return m, nil
	}
	return m, nil
}

// handlePhase2Key handles keys on the PageConfigPhase2 screen (xray validation
// scanning in progress or results shown).
func (m AppModel) handlePhase2Key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if scanCancel != nil {
			scanCancel()
		}
		clearLiveResultWriter()
		m.page = PageHome
		m.configScanning = false
		m.configDone = false
		return m, nil
	case "q", "enter":
		if m.configDone {
			m.page = PageHome
			m.configDone = false
			return m, nil
		}
	case "c":
		if m.configDone {
			m.statusMsg = m.copyWorkingIPs()
			return m, nil
		}
	case "e":
		if m.configDone {
			m.statusMsg = m.exportAllConfigs()
			return m, nil
		}
	case "l":
		if m.configDone && scanLog != nil && scanLog.HasFailures() {
			m.openLogViewer()
			return m, nil
		}
	}
	return m, nil
}

func (m AppModel) viewConfigOptional() string {
	var sb strings.Builder
	sb.WriteString(styleTitle.Render("\n  ⚡  Find Working IPs — optional config\n"))
	sb.WriteString(fmt.Sprintf("%s\n\n", styleSep.Render("  "+strings.Repeat("─", minInt(m.width-4, 70)))))

	rowLabel := func(row int, text string) {
		if m.configOptionalRow == row {
			sb.WriteString(styleAccent.Render(text))
		} else {
			sb.WriteString(styleDim.Render(text))
		}
	}

	renderPills := func(labels []string, selected int) {
		for i, label := range labels {
			if i == selected {
				sb.WriteString(styleSelected.Render(" " + label + " "))
			} else {
				sb.WriteString(styleNormal.Render("  " + label + "  "))
			}
			if i < len(labels)-1 {
				sb.WriteString(styleDim.Render("│"))
			}
		}
	}

	// Row 0: Config URL
	rowLabel(0, "  Config      ")
	sb.WriteString(" " + m.configInput.View() + "\n")
	sb.WriteString(styleDim.Render("              optional — leave empty for Phase 1 only") + "\n\n")

	// Row 1: Top N
	rowLabel(1, "  Top N       ")
	sb.WriteString(" ")
	renderPills(configTopNLabels, m.configTopNIdx)
	sb.WriteString("\n")
	if m.configCustomMode && m.configCustomRow == 5 {
		sb.WriteString(styleAccent.Render("              custom top N: ") + m.configCustomInput.View() + "\n\n")
	} else if m.isTopNCustomSelected() && m.configTopNCustom != "" {
		sb.WriteString(styleDim.Render(fmt.Sprintf("              candidates to validate in Phase 2  (custom: %s)", m.configTopNCustom)) + "\n\n")
	} else {
		sb.WriteString(styleDim.Render("              candidates to validate in Phase 2") + "\n\n")
	}

	// Row 2: P2 Workers
	rowLabel(2, "  P2 Workers  ")
	sb.WriteString(" ")
	renderPills(configPhase2WorkersLabels, m.configPhase2WorkersIdx)
	sb.WriteString("\n")
	if m.configCustomMode && m.configCustomRow == 9 {
		sb.WriteString(styleAccent.Render("              custom workers: ") + m.configCustomInput.View() + "\n\n")
	} else if m.configPhase2WorkersIdx == len(configPhase2WorkersLabels)-1 && m.configPhase2WorkersCustom != "" {
		sb.WriteString(styleDim.Render(fmt.Sprintf("              concurrent xray validation workers in Phase 2  (custom: %s)", m.configPhase2WorkersCustom)) + "\n\n")
	} else {
		sb.WriteString(styleDim.Render("              concurrent xray validation workers in Phase 2") + "\n\n")
	}

	// Row 3: Min Download
	rowLabel(3, "  Min Download")
	sb.WriteString(" ")
	renderPills(configMinSpeedLabels, m.configMinDownloadIdx)
	sb.WriteString("\n")
	if m.configCustomMode && m.configCustomRow == 6 {
		sb.WriteString(styleAccent.Render("              custom min download: ") + m.configCustomInput.View() + " Mbps\n\n")
	} else if m.configMinDownloadIdx == len(configMinSpeedLabels)-1 && m.configMinDownloadCustom != "" {
		sb.WriteString(styleDim.Render(fmt.Sprintf("              minimum download speed limit to filter candidates  (custom: %s Mbps)", m.configMinDownloadCustom)) + "\n\n")
	} else if m.configMinDownloadIdx == 0 {
		sb.WriteString(styleDim.Render("              disable download speed testing") + "\n\n")
	} else if m.configMinDownloadIdx == 1 {
		sb.WriteString(styleDim.Render("              run download speed test without enforcing a minimum limit") + "\n\n")
	} else {
		sb.WriteString(styleDim.Render("              minimum download speed limit to filter candidates") + "\n\n")
	}

	// Row 4: Min Upload
	rowLabel(4, "  Min Upload  ")
	sb.WriteString(" ")
	renderPills(configMinUploadLabels, m.configMinUploadIdx)
	sb.WriteString("\n")
	if m.configCustomMode && m.configCustomRow == 8 {
		sb.WriteString(styleAccent.Render("              custom min upload: ") + m.configCustomInput.View() + " Mbps\n\n")
	} else if m.configMinUploadIdx == len(configMinUploadLabels)-1 && m.configMinUploadCustom != "" {
		sb.WriteString(styleDim.Render(fmt.Sprintf("              minimum upload speed limit to filter candidates  (custom: %s Mbps)", m.configMinUploadCustom)) + "\n\n")
	} else if m.configMinUploadIdx == 0 {
		sb.WriteString(styleDim.Render("              disable upload speed testing") + "\n\n")
	} else if m.configMinUploadIdx == 1 {
		sb.WriteString(styleDim.Render("              run upload speed test without enforcing a minimum limit") + "\n\n")
	} else {
		sb.WriteString(styleDim.Render("              minimum upload speed limit to filter candidates") + "\n\n")
	}

	// Row 5: Speed Size
	rowLabel(5, "  Speed Size  ")
	sb.WriteString(" ")
	renderPills(configSpeedSizeLabels, m.configSpeedSizeIdx)
	sb.WriteString("\n")
	if m.configCustomMode && m.configCustomRow == 7 {
		sb.WriteString(styleAccent.Render("              custom speed size: ") + m.configCustomInput.View() + " MB\n\n")
	} else if m.configSpeedSizeIdx == len(configSpeedSizeLabels)-1 && m.configSpeedSizeCustom != "" {
		sb.WriteString(styleDim.Render(fmt.Sprintf("              speed test sample size for download and upload  (custom: %s MB)", m.configSpeedSizeCustom)) + "\n\n")
	} else {
		sb.WriteString(styleDim.Render("              speed test sample size for download and upload") + "\n\n")
	}

	// Row 6: Speed URL
	rowLabel(6, "  Speed URL   ")
	sb.WriteString(" " + m.configSpeedURLInput.View() + "\n")
	sb.WriteString(styleDim.Render("              custom speed test URL (optional — leave empty for default)") + "\n\n")

	// Row 7: Start Scan
	sb.WriteString("  ")
	if m.configOptionalRow == 7 {
		sb.WriteString(styleSelected.Render(" [ Start Scan ] ") + "\n\n")
	} else {
		sb.WriteString(styleNormal.Render("   [ Start Scan ] ") + "\n\n")
	}

	hint := "  ↑/↓ row   ←/→ option   enter next/confirm   esc back"
	if m.configOptionalRow == 0 {
		hint = "  paste URL   ctrl+x clear   enter next/confirm   esc back"
	} else if m.configOptionalRow == 6 {
		hint = "  type URL   ctrl+x clear   enter next/confirm   esc back"
	} else if m.configOptionalRow == 7 {
		hint = "  enter start scan   esc back"
	}
	if m.configCustomMode {
		hint = "  type value   enter confirm   esc cancel"
	}
	sb.WriteString(styleHint.Render(hint) + "\n")
	if m.liveResultPath != "" {
		sb.WriteString(styleDim.Render("  live file: "+m.liveResultPath) + "\n")
	}
	if m.statusMsg != "" {
		sb.WriteString(styleWarn.Render("  ⚠  "+m.statusMsg) + "\n")
	}
	return sb.String()
}

func (m AppModel) handleConfigOptionalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.configCustomMode {
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(m.configCustomInput.Value())
			if m.configCustomRow == 5 {
				m.configTopNCustom = val
			} else if m.configCustomRow == 6 {
				m.configMinDownloadCustom = val
			} else if m.configCustomRow == 7 {
				m.configSpeedSizeCustom = val
			} else if m.configCustomRow == 8 {
				m.configMinUploadCustom = val
			} else if m.configCustomRow == 9 {
				m.configPhase2WorkersCustom = val
			}
			m.configCustomMode = false
			m.configCustomInput.Blur()
			if m.configOptionalRow < 7 {
				m.configOptionalRow++
				if m.configOptionalRow == 6 {
					m.configSpeedURLInput.Focus()
					return m, textinput.Blink
				}
			}
			return m, nil
		case "esc":
			m.configCustomMode = false
			m.configCustomInput.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.configCustomInput, cmd = m.configCustomInput.Update(msg)
		return m, cmd
	}

	// --- When a text input is focused (row 0 or row 6) ---
	if m.configOptionalRow == 0 || m.configOptionalRow == 6 {
		switch msg.String() {
		case "esc":
			m.page = PageScanWithConfig
			m.configInput.Blur()
			m.configSpeedURLInput.Blur()
			return m, nil

		case "up":
			if m.configOptionalRow > 0 {
				m.configOptionalRow--
				m.configInput.Blur()
				m.configSpeedURLInput.Blur()
				if m.configOptionalRow == 0 {
					m.configInput.Focus()
					return m, textinput.Blink
				}
			}
			return m, nil

		case "down":
			if m.configOptionalRow < 7 {
				m.configOptionalRow++
				m.configInput.Blur()
				m.configSpeedURLInput.Blur()
				if m.configOptionalRow == 6 {
					m.configSpeedURLInput.Focus()
					return m, textinput.Blink
				}
			}
			return m, nil

		case "enter":
			if m.configOptionalRow == 0 {
				rawURL := strings.TrimSpace(m.configInput.Value())
				if rawURL == "" {
					m.configOptionalRow = 7
					m.configInput.Blur()
					return m, nil
				}
				m.configOptionalRow = 1
				m.configInput.Blur()
				return m, nil
			}
			if m.configOptionalRow == 6 {
				m.configOptionalRow = 7
				m.configSpeedURLInput.Blur()
				return m, nil
			}

		case "ctrl+x":
			if m.configOptionalRow == 0 {
				m.configInput.SetValue("")
				return m, nil
			} else if m.configOptionalRow == 6 {
				m.configSpeedURLInput.SetValue("")
				return m, nil
			}
		}

		// All other keys (runes, left/right arrow keys, backspace, etc.) go to the active text input
		if m.configOptionalRow == 0 {
			var cmd tea.Cmd
			m.configInput, cmd = m.configInput.Update(msg)
			return m, cmd
		} else {
			var cmd tea.Cmd
			m.configSpeedURLInput, cmd = m.configSpeedURLInput.Update(msg)
			return m, cmd
		}
	}

	// --- When navigation rows (1, 2, 3, 4, 5, 7) are focused ---
	switch msg.String() {
	case "esc":
		m.page = PageScanWithConfig
		m.configInput.Blur()
		m.configSpeedURLInput.Blur()
		return m, nil

	case "up", "k":
		if m.configOptionalRow > 0 {
			m.configOptionalRow--
			m.configInput.Blur()
			m.configSpeedURLInput.Blur()
			if m.configOptionalRow == 0 {
				m.configInput.Focus()
				return m, textinput.Blink
			} else if m.configOptionalRow == 6 {
				m.configSpeedURLInput.Focus()
				return m, textinput.Blink
			}
		}
		return m, nil

	case "down", "j":
		if m.configOptionalRow < 7 {
			m.configOptionalRow++
			m.configInput.Blur()
			m.configSpeedURLInput.Blur()
			if m.configOptionalRow == 6 {
				m.configSpeedURLInput.Focus()
				return m, textinput.Blink
			}
		}
		return m, nil

	case "left", "h":
		switch m.configOptionalRow {
		case 1:
			if m.configTopNIdx > 0 {
				m.configTopNIdx--
			}
		case 2:
			if m.configPhase2WorkersIdx > 0 {
				m.configPhase2WorkersIdx--
			}
		case 3:
			if m.configMinDownloadIdx > 0 {
				m.configMinDownloadIdx--
			}
		case 4:
			if m.configMinUploadIdx > 0 {
				m.configMinUploadIdx--
			}
		case 5:
			if m.configSpeedSizeIdx > 0 {
				m.configSpeedSizeIdx--
			}
		}
		return m, nil

	case "right", "l":
		switch m.configOptionalRow {
		case 1:
			if m.configTopNIdx < len(configTopNLabels)-1 {
				m.configTopNIdx++
			}
		case 2:
			if m.configPhase2WorkersIdx < len(configPhase2WorkersLabels)-1 {
				m.configPhase2WorkersIdx++
			}
		case 3:
			if m.configMinDownloadIdx < len(configMinSpeedLabels)-1 {
				m.configMinDownloadIdx++
			}
		case 4:
			if m.configMinUploadIdx < len(configMinUploadLabels)-1 {
				m.configMinUploadIdx++
			}
		case 5:
			if m.configSpeedSizeIdx < len(configSpeedSizeLabels)-1 {
				m.configSpeedSizeIdx++
			}
		}
		return m, nil

	case "enter":
		if m.configOptionalRow == 1 && m.isTopNCustomSelected() {
			m.configCustomMode = true
			m.configCustomRow = 5
			m.configCustomInput.SetValue(m.configTopNCustom)
			m.configCustomInput.Placeholder = "e.g. 75"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		if m.configOptionalRow == 2 && m.configPhase2WorkersIdx == len(configPhase2WorkersLabels)-1 {
			m.configCustomMode = true
			m.configCustomRow = 9
			m.configCustomInput.SetValue(m.configPhase2WorkersCustom)
			m.configCustomInput.Placeholder = "e.g. 15"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		if m.configOptionalRow == 3 && m.configMinDownloadIdx == len(configMinSpeedLabels)-1 {
			m.configCustomMode = true
			m.configCustomRow = 6
			m.configCustomInput.SetValue(m.configMinDownloadCustom)
			m.configCustomInput.Placeholder = "e.g. 3.5"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		if m.configOptionalRow == 4 && m.configMinUploadIdx == len(configMinUploadLabels)-1 {
			m.configCustomMode = true
			m.configCustomRow = 8
			m.configCustomInput.SetValue(m.configMinUploadCustom)
			m.configCustomInput.Placeholder = "e.g. 1.5"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		if m.configOptionalRow == 5 && m.configSpeedSizeIdx == len(configSpeedSizeLabels)-1 {
			m.configCustomMode = true
			m.configCustomRow = 7
			m.configCustomInput.SetValue(m.configSpeedSizeCustom)
			m.configCustomInput.Placeholder = "e.g. 10 (MB)"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		if m.configOptionalRow == 7 {
			return m.launchPhase1FromOptional()
		}
		m.configOptionalRow++
		if m.configOptionalRow == 6 {
			m.configSpeedURLInput.Focus()
			return m, textinput.Blink
		}
		return m, nil
	}

	return m, nil
}

func (m AppModel) isTopNCustomSelected() bool {
	return m.configTopNIdx == len(configTopNLabels)-1
}

func (m AppModel) launchPhase1FromOptional() (AppModel, tea.Cmd) {
	rawURL := strings.TrimSpace(m.configInput.Value())
	withConfig := rawURL != ""
	if withConfig {
		if _, err := xraytest.ParseProxyURL(rawURL); err != nil {
			m.statusMsg = fmt.Sprintf("invalid URL: %v", err)
			m.configOptionalRow = 0
			m.configInput.Focus()
			return m, textinput.Blink
		}
		m.configURL = rawURL
	} else {
		m.configURL = ""
	}

	writer, path, err := newLiveResultWriter(withConfig)
	if err != nil {
		m.statusMsg = fmt.Sprintf("could not create results file: %v", err)
		return m, nil
	}
	setLiveResultWriter(writer)
	m.liveResultPath = path
	m.statusMsg = ""
	m.configPhase1Only = !withConfig
	m.configPhase1Results = nil
	m.configPhase1Done = false
	m.configPhase1Stats = StatsMsg{}
	m.page = PageConfigPhase1

	count := configCountValues[m.configCountIdx]
	if count == 0 {
		count, _ = strconv.Atoi(m.configCountCustom)
		if count <= 0 {
			count = 1000
		}
	}
	m.configPhase1Total = m.phase1TargetTotal(count)

	// Save the current config to disk
	savedCfg := SavedConfig{
		IPMode:          m.configIPMode,
		CountIdx:        m.configCountIdx,
		CountCustom:     m.configCountCustom,
		WorkersIdx:      m.configWorkersIdx,
		WorkersCustom:   m.configWorkersCustom,
		TimeoutIdx:      m.configTimeoutIdx,
		TimeoutCustom:   m.configTimeoutCustom,
		ConfigURL:       rawURL,
		TopNIdx:         m.configTopNIdx,
		TopNCustom:      m.configTopNCustom,
		MinDownloadIdx:     m.configMinDownloadIdx,
		MinDownloadCustom:  m.configMinDownloadCustom,
		MinUploadIdx:    m.configMinUploadIdx,
		MinUploadCustom: m.configMinUploadCustom,
		SpeedURL:        strings.TrimSpace(m.configSpeedURLInput.Value()),
		SpeedSizeIdx:    m.configSpeedSizeIdx,
		SpeedSizeCustom: m.configSpeedSizeCustom,
		DownloadTest:    m.isDownloadTestEnabled(),
		UploadTest:      m.isUploadTestEnabled(),
		RequireWS:       m.scanCfg.RequireWS,
		Phase2WorkersIdx:    m.configPhase2WorkersIdx,
		Phase2WorkersCustom: m.configPhase2WorkersCustom,
	}
	for port, on := range m.configSelectedPorts {
		if on {
			savedCfg.Ports = append(savedCfg.Ports, port)
		}
	}
	_ = saveAppConfig(AppConfig{LastConfig: savedCfg})

	return m, m.startConfigPhase1()
}

func (m AppModel) resolveTopN() int {
	if m.isTopNCustomSelected() {
		n, _ := strconv.Atoi(strings.TrimSpace(m.configTopNCustom))
		if n <= 0 {
			return 50
		}
		return n
	}
	if m.configTopNIdx < 0 || m.configTopNIdx >= len(configTopNValues) {
		return 50
	}
	return configTopNValues[m.configTopNIdx]
}

func (m AppModel) resolvePhase2Workers() int {
	if m.configPhase2WorkersIdx == len(configPhase2WorkersLabels)-1 {
		n, _ := strconv.Atoi(strings.TrimSpace(m.configPhase2WorkersCustom))
		if n <= 0 {
			return 10
		}
		return n
	}
	if m.configPhase2WorkersIdx < 0 || m.configPhase2WorkersIdx >= len(configPhase2WorkersValues) {
		return 10
	}
	val := configPhase2WorkersValues[m.configPhase2WorkersIdx]
	if val <= 0 {
		return 10
	}
	return val
}

// ConfigDoneMsg signals all config validations are complete.
type ConfigDoneMsg struct{}

// ---------------------------------------------------------------------------
// Config Phase 1 — fast connectivity scan
// ---------------------------------------------------------------------------
type ConfigProgressMsg struct {
	Result *xraytest.ValidationResult
	Done   int
	Total  int
}

// ---------------------------------------------------------------------------
// Config Setup presets
// ---------------------------------------------------------------------------

var configCountValues = []int{1000, 5000, 20000, 0} // 0 = custom
var configCountLabels = []string{"1,000", "5,000", "20,000", "Custom"}
var configTopNValues = []int{10, 25, 50, 100, 0} // 0 = all
var configTopNLabels = []string{"10", "25", "50", "100", "All", "Custom"}
var configPhase2WorkersValues = []int{5, 10, 20, 50, 0}
var configPhase2WorkersLabels = []string{"5", "10", "20", "50", "Custom"}
var configIPModeLabels = []string{"Random", "From File"}
var configPortChoices = []struct {
	label string
	port  int
}{
	{"Config", 0},
	{"443", 443},
	{"8443", 8443},
	{"2053", 2053},
	{"2083", 2083},
	{"2087", 2087},
	{"2096", 2096},
}


// quickWorkersLabels and quickTimeoutLabels return the display labels for the
// shared quick-scan preset slices so viewScanWithConfig can use them.
func quickWorkersLabels() []string {
	out := make([]string, len(quickWorkersPresets))
	for i, p := range quickWorkersPresets {
		out[i] = p.label
	}
	return out
}

func quickTimeoutLabels() []string {
	out := make([]string, len(quickTimeoutPresets))
	for i, p := range quickTimeoutPresets {
		out[i] = p.label
	}
	return out
}

// ---------------------------------------------------------------------------
// Config Phase 1 — fast connectivity scan
// ---------------------------------------------------------------------------

type ConfigPhase1ResultMsg struct {
	Result *result.Result
}

// ConfigPhase1ErrMsg is sent when Phase 1 cannot proceed (e.g. ips.txt missing).
type ConfigPhase1ErrMsg struct{ Err string }

type ConfigPhase1DoneMsg struct{}

func (m AppModel) viewConfigPhase1() string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("\n  ⚡  Phase 1 — Finding reachable IPs\n"))
	sb.WriteString(fmt.Sprintf("%s\n\n", styleSep.Render("  "+strings.Repeat("─", minInt(m.width-4, 70)))))

	icon := m.spinner.View()
	if m.configPhase1Done {
		icon = styleGood.Render("✓")
	}

	healthy := 0
	for _, r := range m.configPhase1Results {
		if r.IsHealthy() {
			healthy++
		}
	}

	targetStr := fmt.Sprintf("%d", m.configPhase1Total)
	sb.WriteString(fmt.Sprintf("  %s  tested: %s  candidates: %s  target: %s\n\n",
		icon,
		styleAccent.Render(fmt.Sprintf("%d", len(m.configPhase1Results))),
		styleGood.Render(fmt.Sprintf("%d", healthy)),
		styleDim.Render(targetStr),
	))
	if !m.configPhase1Done {
		sb.WriteString(fmt.Sprintf("  %s  %s  ports: %s\n\n",
			styleAccent.Render(scanPulse(m.bannerFrame)),
			scanWave(m.bannerFrame, 28),
			styleDim.Render(formatPorts(m.resolveConfigPorts())),
		))
	}

	if m.configPhase1Done {
		if m.configPhase1Only {
			sb.WriteString(styleGood.Render(fmt.Sprintf("  Done — %d healthy endpoints found.", healthy)))
			sb.WriteString("\n\n")
		} else {
			topN := m.resolveTopN()
			label := fmt.Sprintf("%d", topN)
			if topN == 0 {
				label = "all"
			}
			sb.WriteString(styleGood.Render(fmt.Sprintf("  Found %d candidates. Testing top %s with xray...", healthy, label)))
			sb.WriteString("\n\n")
		}
	} else if m.configIPMode == 1 {
		sb.WriteString(styleNormal.Render("  Probing IPs from ips.txt on the selected ports..."))
		sb.WriteString("\n\n")
	} else if strings.TrimSpace(m.configURL) == "" {
		sb.WriteString(styleNormal.Render("  Scanning random Cloudflare IPv4 IPs (standard HTTP probe)..."))
		sb.WriteString("\n")
		sb.WriteString(styleDim.Render("  healthy hits also explore nearby addresses in the same Cloudflare block"))
		sb.WriteString("\n\n")
	} else {
		sb.WriteString(styleNormal.Render("  Scanning Cloudflare IPs using your config probe settings..."))
		sb.WriteString("\n")
		sb.WriteString(styleDim.Render("  healthy hits also explore nearby addresses in the same Cloudflare block"))
		sb.WriteString("\n\n")
	}

	if m.liveResultPath != "" {
		sb.WriteString(styleDim.Render("  live results → " + m.liveResultPath))
		sb.WriteString("\n\n")
	}

	if len(m.configPhase1Results) > 0 {
		hdr := fmt.Sprintf("  %-22s  %7s  %9s  %-8s  %6s",
			"ENDPOINT", "LOSS", "AVG(ms)", "COLO", "STATUS")
		sb.WriteString(fmt.Sprintf("%s\n%s\n", styleHeader.Render(hdr), styleSep.Render("  "+strings.Repeat("─", 64))))

		top := result.TopN(m.configPhase1Results, 20)
		for _, r := range top {
			colo := r.Colo
			if colo == "" {
				colo = "—"
			}
			status := "✓"
			lineStyle := styleGood
			if !r.IsHealthy() {
				status = "✗"
				lineStyle = styleBad
			}
			line := fmt.Sprintf("  %-22s  %6.1f%%  %9.2f  %-8s  %6s",
				formatEndpoint(r.IP.String(), r.Port), r.Loss(),
				float64(r.Avg().Milliseconds()), colo, status)
			sb.WriteString(lineStyle.Render(line) + "\n")
		}
		sb.WriteRune('\n')
	}

	if m.configPhase1Done && m.configPhase1Only {
		hint := "  c/enter copy endpoints   q/esc back"
		if scanLog != nil && scanLog.HasFailures() {
			hint += fmt.Sprintf("   l view %s", scanLog.FailureSummary())
		}
		sb.WriteString(styleHint.Render(hint) + "\n")
	} else if m.configPhase1Done && !m.configPhase1Only {
		hint := "  enter start Phase 2   q/esc cancel"
		if scanLog != nil && scanLog.HasFailures() {
			hint += fmt.Sprintf("   l view %s", scanLog.FailureSummary())
		}
		sb.WriteString(styleHint.Render(hint) + "\n")
	} else {
		sb.WriteString(styleHint.Render("  q/esc cancel") + "\n")
	}
	return sb.String()
}

func (m AppModel) handleConfigPhase1Key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.configPhase1Done && m.configPhase1Only {
			m.statusMsg = m.copyPhase1HealthyEndpoints()
			return m, nil
		}
		if m.configPhase1Done && !m.configPhase1Only && len(m.configPhase1Results) > 0 {
			// Manually start Phase 2 from Phase 1 results
			topN := m.resolveTopN()
			var topIPs []*result.Result
			if topN == 0 {
				topIPs = result.TopN(m.configPhase1Results, 0)
			} else {
				topIPs = result.TopN(m.configPhase1Results, topN)
			}
			if len(topIPs) == 0 {
				m.statusMsg = "no healthy candidates to validate"
				return m, nil
			}
			if liveResultWriter != nil {
				liveResultWriter.BeginPhase2()
			}
			m.page = PageConfigPhase2
			m.configScanning = true
			m.configDone = false
			m.configResults = nil
			return m, m.startConfigPhase2(topIPs)
		}
	case "c":
		if m.configPhase1Done && m.configPhase1Only {
			m.statusMsg = m.copyPhase1HealthyEndpoints()
			return m, nil
		}
	case "l":
		if m.configPhase1Done && scanLog != nil && scanLog.HasFailures() {
			m.openLogViewer()
			return m, nil
		}
	case "esc", "q":
		if scanCancel != nil {
			scanCancel()
		}
		clearLiveResultWriter()
		m.page = PageHome
		return m, nil
	}
	return m, nil
}

func (m AppModel) copyPhase1HealthyEndpoints() string {
	top := result.TopN(m.configPhase1Results, 0)
	if len(top) == 0 {
		return "no healthy endpoints to copy"
	}
	endpoints := make([]string, 0, len(top))
	for _, r := range top {
		endpoints = append(endpoints, formatEndpoint(r.IP.String(), r.Port))
	}
	return copyAndSaveIPs(endpoints, m.ispInfo)
}

// configPhase1Options holds the resolved settings for a Phase 1 engine run.
type configPhase1Options struct {
	count       int
	concurrency int
	timeout     time.Duration
	rawURL      string
	ports       []int
	fromFile    bool
	requireWS   bool
}

func (m AppModel) startConfigPhase1() tea.Cmd {
	opts := m.resolvePhase1Options()
	return func() tea.Msg {
		go runConfigPhase1(opts)
		return nil
	}
}

func (m AppModel) resolveTimeout() time.Duration {
	var timeout time.Duration
	if m.configTimeoutIdx < len(quickTimeoutPresets) {
		tp := quickTimeoutPresets[m.configTimeoutIdx]
		if tp.value != "" {
			timeout, _ = time.ParseDuration(tp.value)
		} else {
			timeout, _ = time.ParseDuration(m.configTimeoutCustom)
		}
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return timeout
}

// resolvePhase1Options reads the current picker state and returns concrete values
// for the Phase 1 engine run.
func (m AppModel) resolvePhase1Options() configPhase1Options {
	// Count
	count := configCountValues[m.configCountIdx]
	if count == 0 {
		count, _ = strconv.Atoi(m.configCountCustom)
		if count <= 0 {
			count = 1000
		}
	}

	concurrency := 0
	if m.configWorkersIdx < len(quickWorkersPresets) {
		wp := quickWorkersPresets[m.configWorkersIdx]
		if wp.value != "" {
			concurrency, _ = strconv.Atoi(wp.value)
		} else {
			concurrency, _ = strconv.Atoi(m.configWorkersCustom)
		}
	}
	if concurrency <= 0 {
		concurrency = 50
	}

	return configPhase1Options{
		count:       count,
		concurrency: concurrency,
		timeout:     m.resolveTimeout(),
		rawURL:      m.configURL,
		ports:       m.resolveConfigPorts(),
		fromFile:    m.configIPMode == 1,
		requireWS:   m.scanCfg.RequireWS,
	}
}

func (m AppModel) phase1TargetTotal(count int) int {
	ports := len(m.resolveConfigPorts())
	if ports <= 0 {
		ports = 1
	}
	if m.configIPMode == 1 {
		if ips, err := loadDefaultIPsFile(); err == nil {
			numIPs := len(ips)
			if numIPs > count {
				numIPs = count
			}
			return numIPs * ports
		}
		return 0
	}
	return count * ports
}

func (m AppModel) resolveConfigPorts() []int {
	selected := m.selectedPortSet()
	var ports []int
	for _, choice := range configPortChoices {
		if choice.port > 0 && selected[choice.port] {
			ports = append(ports, choice.port)
		}
	}
	if len(ports) > 0 {
		return ports
	}
	cfg, err := xraytest.ParseProxyURL(m.configURL)
	if err != nil || cfg.Port <= 0 {
		return []int{443}
	}
	return []int{cfg.Port}
}

// runConfigPhase1 is defined in cmds.go and accepts a configPhase1Options struct.

// ---------------------------------------------------------------------------
// Config Phase 2 — xray validation of top IPs
// ---------------------------------------------------------------------------

func (m AppModel) startConfigPhase2(topIPs []*result.Result) tea.Cmd {
	url := m.configURL
	minSpeed := m.resolveMinSpeed()
	minUploadSpeed := m.resolveMinUploadSpeed()
	speedURL := strings.TrimSpace(m.configSpeedURLInput.Value())
	speedSize := m.resolveSpeedSize()
	timeout := m.resolveTimeout()
	uploadTest := m.isUploadTestEnabled()

	downloadTest := m.isDownloadTestEnabled()
	workers := m.resolvePhase2Workers()

	// Xray validation has startup and SOCKS proxy overheads.
	// We enforce a minimum floor of 20s and scale with the user timeout.
	xrayTimeout := timeout * 2
	if xrayTimeout < 20*time.Second {
		xrayTimeout = 20 * time.Second
	}

	// Add dynamic budget for speed testing if a speed check is requested/configured.
	// We calculate how long a download is expected to take at the user's minSpeed
	// threshold (in Mbps) or a default of 1 Mbps.
	speedBits := speedSize * 8
	effectiveMinSpeed := minSpeed
	if effectiveMinSpeed <= 0 {
		effectiveMinSpeed = 1.0 // 1 Mbps default for estimation
	}
	expectedSpeedSec := float64(speedBits) / (effectiveMinSpeed * 1_000_000)
	// Give a 3x buffer to allow slow/flaky but working connections to finish the test,
	// with a minimum floor of 5s and ceiling of 30s.
	speedLimit := time.Duration(expectedSpeedSec * 3 * float64(time.Second))
	if speedLimit < 5*time.Second {
		speedLimit = 5 * time.Second
	}
	if speedLimit > 30*time.Second {
		speedLimit = 30 * time.Second
	}
	if downloadTest {
		xrayTimeout += speedLimit
	}

	return func() tea.Msg {
		go runConfigPhase2(url, topIPs, minSpeed, minUploadSpeed, speedURL, speedSize, xrayTimeout, downloadTest, uploadTest, workers)
		return nil
	}
}

func runConfigPhase2(rawURL string, topIPs []*result.Result, minSpeed float64, minUploadSpeed float64, speedURL string, speedSize int64, timeout time.Duration, downloadTest bool, uploadTest bool, workers int) {
	cfg, err := xraytest.ParseProxyURL(rawURL)
	if err != nil {
		scanLog.Error(logger.Phase2, "invalid URL: %v", err)
		if prog != nil {
			prog.Send(ConfigDoneMsg{})
		}
		return
	}

	ctx := context.Background()
	total := len(topIPs)
	if total == 0 {
		scanLog.Info(logger.Phase2, "no candidates to validate")
		if prog != nil {
			prog.Send(ConfigDoneMsg{})
		}
		return
	}

	if errMsg := cfg.Phase2SanityError(); errMsg != "" {
		scanLog.Error(logger.Phase2, "config sanity check failed: %s", errMsg)
		for i, r := range topIPs {
			vr := &xraytest.ValidationResult{
				IP:        r.IP.String(),
				Port:      r.Port,
				Transport: cfg.Network,
				Error:     errMsg,
			}
			if liveResultWriter != nil {
				liveResultWriter.AddPhase2(vr)
			}
			if prog != nil {
				prog.Send(ConfigProgressMsg{Result: vr, Done: i + 1, Total: total})
			}
		}
		if prog != nil {
			prog.Send(ConfigDoneMsg{})
		}
		return
	}

	scanLog.ScanStart(logger.Phase2, map[string]any{
		"candidates": total, "timeout": timeout, "minSpeed": minSpeed,
		"network": cfg.Network, "workers": workers,
	})

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var done atomic.Int32
	var working int32

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
			swapped.DownloadTest = downloadTest
			swapped.UploadTest = uploadTest
			vr := xraytest.ValidateConfig(ctx, swapped, timeout)
			if vr.Success && minSpeed > 0 && downloadTest {
				mbps := vr.Throughput * 8 / 1_000_000
				if mbps < minSpeed {
					vr.Success = false
					vr.Error = fmt.Sprintf("speed below threshold (%.1f < %.1f Mbps)", mbps, minSpeed)
				}
			}
			if vr.Success && minUploadSpeed > 0 && uploadTest {
				upMbps := vr.UploadThroughput * 8 / 1_000_000
				if upMbps < minUploadSpeed {
					vr.Success = false
					vr.Error = fmt.Sprintf("upload speed below threshold (%.1f < %.1f Mbps)", upMbps, minUploadSpeed)
				}
			}
			if vr.Success {
				working++
				scanLog.ProbeSuccess(logger.Phase2, formatEndpoint(vr.IP, vr.Port), float64(vr.Latency.Milliseconds()))
			} else {
				scanLog.ProbeFailure(logger.Phase2, formatEndpoint(vr.IP, vr.Port), vr.Error)
			}
			if liveResultWriter != nil {
				liveResultWriter.AddPhase2(vr)
			}
			if prog != nil {
				prog.Send(ConfigProgressMsg{
					Result: vr,
					Done:   int(done.Add(1)),
					Total:  total,
				})
			}
		}()
	}

	wg.Wait()

	scanLog.ScanComplete(logger.Phase2, total, int(working), 0)
	if prog != nil {
		prog.Send(ConfigDoneMsg{})
	}
}

// ---------------------------------------------------------------------------
// Config Resolvers and Exporters
// ---------------------------------------------------------------------------

var configMinSpeedValues = []float64{-1.0, 0.0, 1.0, 2.0, 5.0, -2.0} // -1.0 = OFF, 0.0 = No Limit, -2.0 = Custom
var configMinSpeedLabels = []string{"OFF", "No Limit", "1 Mbps", "2 Mbps", "5 Mbps", "Custom"}

func (m AppModel) isDownloadTestEnabled() bool {
	return m.configMinDownloadIdx != 0
}

var configMinUploadValues = []float64{-1.0, 0.0, 0.5, 1.0, 2.0, -2.0} // -1.0 = OFF, 0.0 = No Limit, -2.0 = Custom
var configMinUploadLabels = []string{"OFF", "No Limit", "0.5 Mbps", "1.0 Mbps", "2.0 Mbps", "Custom"}

func (m AppModel) isUploadTestEnabled() bool {
	return m.configMinUploadIdx != 0
}

func (m AppModel) resolveMinUploadSpeed() float64 {
	if m.configMinUploadIdx == len(configMinUploadLabels)-1 {
		val, _ := strconv.ParseFloat(strings.TrimSpace(m.configMinUploadCustom), 64)
		if val <= 0 {
			return 0
		}
		return val
	}
	if m.configMinUploadIdx <= 1 {
		return 0
	}
	if m.configMinUploadIdx < 0 || m.configMinUploadIdx >= len(configMinUploadValues) {
		return 0
	}
	return configMinUploadValues[m.configMinUploadIdx]
}

var configSpeedSizeValues = []int64{128 * 1024, 512 * 1024, 1024 * 1024, 5 * 1024 * 1024, 0} // 0 = custom
var configSpeedSizeLabels = []string{"128 KB", "512 KB (default)", "1 MB", "5 MB", "Custom"}

func (m AppModel) resolveMinSpeed() float64 {
	if m.configMinDownloadIdx == len(configMinSpeedLabels)-1 {
		n, _ := strconv.ParseFloat(strings.TrimSpace(m.configMinDownloadCustom), 64)
		if n <= 0 {
			return 0
		}
		return n
	}
	if m.configMinDownloadIdx <= 1 {
		return 0
	}
	if m.configMinDownloadIdx < 0 || m.configMinDownloadIdx >= len(configMinSpeedValues) {
		return 0
	}
	return configMinSpeedValues[m.configMinDownloadIdx]
}

func (m AppModel) resolveSpeedSize() int64 {
	if m.configSpeedSizeIdx == len(configSpeedSizeLabels)-1 {
		n, _ := strconv.ParseInt(strings.TrimSpace(m.configSpeedSizeCustom), 10, 64)
		if n <= 0 {
			return 512 * 1024
		}
		return n * 1024 * 1024 // Custom input is in MB
	}
	if m.configSpeedSizeIdx < 0 || m.configSpeedSizeIdx >= len(configSpeedSizeValues) {
		return 512 * 1024
	}
	return configSpeedSizeValues[m.configSpeedSizeIdx]
}

func (m AppModel) exportAllConfigs() string {
	endpoints := workingEndpoints(m.configResults)
	if len(endpoints) == 0 {
		return "no working endpoints to export"
	}

	cfg, err := xraytest.ParseProxyURL(m.configURL)
	if err != nil {
		return fmt.Sprintf("invalid config URL: %v", err)
	}

	exe, err := osExecutable()
	dir := ""
	if err == nil {
		dir = filepath.Dir(exe)
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	if dir == "" {
		dir = "."
	}

	// 1. Export Subscription URLs
	var subUrls []string
	for i, ep := range endpoints {
		parts := strings.Split(ep, ":")
		port := cfg.Port
		if len(parts) > 1 {
			port, _ = strconv.Atoi(parts[1])
		}
		swapped := cfg.WithEndpoint(parts[0], port)
		
		// Suffix sub link remark with endpoint number and user ISP if known
		var remarkSuffix string
		if m.ispInfo != "" {
			remarkSuffix = fmt.Sprintf("CF-Endpoint-%d-%s", i+1, m.ispInfo)
		} else {
			remarkSuffix = fmt.Sprintf("CF-Endpoint-%d", i+1)
		}
		if cfg.Remark != "" {
			swapped.Remark = cfg.Remark + "-" + remarkSuffix
		} else {
			swapped.Remark = remarkSuffix
		}
		
		subUrls = append(subUrls, swapped.ToShareURL())
	}
	subPath := filepath.Join(dir, "senpaiscanner-sub.txt")
	_ = os.WriteFile(subPath, []byte(strings.Join(subUrls, "\n")+"\n"), 0644)

	// 2. Export Sing-Box JSON
	type SBOutbound struct {
		Type       string                 `json:"type"`
		Tag        string                 `json:"tag"`
		Server     string                 `json:"server"`
		ServerPort int                    `json:"server_port"`
		UUID       string                 `json:"uuid,omitempty"`
		Password   string                 `json:"password,omitempty"`
		TLS        map[string]interface{} `json:"tls,omitempty"`
		Transport  map[string]interface{} `json:"transport,omitempty"`
	}

	var sbOutbounds []interface{}
	for i, ep := range endpoints {
		parts := strings.Split(ep, ":")
		ip := parts[0]
		port := cfg.Port
		if len(parts) > 1 {
			port, _ = strconv.Atoi(parts[1])
		}
		tag := fmt.Sprintf("CF-Endpoint-%d", i+1)
		if m.ispInfo != "" {
			tag = fmt.Sprintf("CF-Endpoint-%d-%s", i+1, m.ispInfo)
		}

		o := SBOutbound{
			Type:       cfg.Protocol,
			Tag:        tag,
			Server:     ip,
			ServerPort: port,
		}

		if cfg.Protocol == "trojan" {
			o.Password = cfg.Password
		} else { // vless or vmess
			o.UUID = cfg.UUID
		}

		// TLS
		if cfg.Security == "tls" {
			tlsConf := map[string]interface{}{
				"enabled":     true,
				"server_name": cfg.SNI,
			}
			if cfg.Fingerprint != "" {
				tlsConf["utls"] = map[string]interface{}{
					"enabled":     true,
					"fingerprint": cfg.Fingerprint,
				}
			}
			if len(cfg.ALPN) > 0 {
				tlsConf["alpn"] = cfg.ALPN
			}
			o.TLS = tlsConf
		}

		// Transport
		if cfg.Network == "ws" {
			wsConf := map[string]interface{}{
				"type": "ws",
				"path": cfg.Path,
			}
			if cfg.Host != "" {
				wsConf["headers"] = map[string]interface{}{
					"Host": cfg.Host,
				}
			}
			o.Transport = wsConf
		} else if cfg.Network == "grpc" {
			grpcConf := map[string]interface{}{
				"type":         "grpc",
				"service_name": cfg.ServiceName,
			}
			o.Transport = grpcConf
		}

		sbOutbounds = append(sbOutbounds, o)
	}

	// Minimal Sing-Box client config
	sbConfig := map[string]interface{}{
		"outbounds": sbOutbounds,
	}
	sbJSON, _ := json.MarshalIndent(sbConfig, "", "  ")
	sbPath := filepath.Join(dir, "senpaiscanner-singbox.json")
	_ = os.WriteFile(sbPath, sbJSON, 0644)

	// 3. Export Clash YAML
	var clashLines []string
	clashLines = append(clashLines, "proxies:")
	for i, ep := range endpoints {
		parts := strings.Split(ep, ":")
		ip := parts[0]
		port := cfg.Port
		if len(parts) > 1 {
			port, _ = strconv.Atoi(parts[1])
		}
		name := fmt.Sprintf("CF-Endpoint-%d", i+1)
		if m.ispInfo != "" {
			name = fmt.Sprintf("CF-Endpoint-%d-%s", i+1, m.ispInfo)
		}

		clashLines = append(clashLines, fmt.Sprintf("  - name: \"%s\"", name))
		clashLines = append(clashLines, fmt.Sprintf("    type: %s", cfg.Protocol))
		clashLines = append(clashLines, fmt.Sprintf("    server: %s", ip))
		clashLines = append(clashLines, fmt.Sprintf("    port: %d", port))

		if cfg.Protocol == "trojan" {
			clashLines = append(clashLines, fmt.Sprintf("    password: %s", cfg.Password))
			clashLines = append(clashLines, "    udp: true")
		} else { // vless or vmess
			clashLines = append(clashLines, fmt.Sprintf("    uuid: %s", cfg.UUID))
			if cfg.Protocol == "vmess" {
				clashLines = append(clashLines, "    alterId: 0")
				clashLines = append(clashLines, "    cipher: auto")
			}
		}

		if cfg.Security == "tls" {
			clashLines = append(clashLines, "    tls: true")
			if cfg.SNI != "" {
				clashLines = append(clashLines, fmt.Sprintf("    servername: %s", cfg.SNI))
			}
			if cfg.Fingerprint != "" {
				clashLines = append(clashLines, fmt.Sprintf("    client-fingerprint: %s", cfg.Fingerprint))
			}
		}

		if cfg.Network == "ws" {
			clashLines = append(clashLines, "    network: ws")
			clashLines = append(clashLines, "    ws-opts:")
			clashLines = append(clashLines, fmt.Sprintf("      path: %s", cfg.Path))
			if cfg.Host != "" {
				clashLines = append(clashLines, "      headers:")
				clashLines = append(clashLines, fmt.Sprintf("        Host: %s", cfg.Host))
			}
		} else if cfg.Network == "grpc" {
			clashLines = append(clashLines, "    network: grpc")
			clashLines = append(clashLines, "    grpc-opts:")
			clashLines = append(clashLines, fmt.Sprintf("      grpc-service-name: %s", cfg.ServiceName))
		}
	}
	clashPath := filepath.Join(dir, "senpaiscanner-clash.yaml")
	_ = os.WriteFile(clashPath, []byte(strings.Join(clashLines, "\n")+"\n"), 0644)

	return fmt.Sprintf("✓ configs exported to Clash, Sing-Box, and Sub files in %s", dir)
}

// ---------------------------------------------------------------------------
// Log Viewer
// ---------------------------------------------------------------------------

func (m AppModel) viewLogViewer() string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("\n  📋  Scan Log\n"))
	sb.WriteString(fmt.Sprintf("%s\n\n", styleSep.Render("  "+strings.Repeat("─", minInt(m.width-4, 70)))))

	if len(m.logViewerLines) == 0 {
		sb.WriteString(styleDim.Render("  No log entries found.\n\n"))
	} else {
		maxRows := m.height - 8
		if maxRows < 5 {
			maxRows = 5
		}
		lines := m.logViewerLines
		if len(lines) > maxRows {
			lines = lines[len(lines)-maxRows:]
		}
		for _, line := range lines {
			lineStyle := styleNormal
			switch {
			case strings.Contains(line, "[ERROR]"):
				lineStyle = styleBad
			case strings.Contains(line, "[WARN ]"):
				lineStyle = styleWarn
			case strings.Contains(line, "[INFO ]"):
				lineStyle = styleNormal
			case strings.Contains(line, "[DEBUG]"):
				lineStyle = styleDim
			}
			sb.WriteString(lineStyle.Render("  " + line) + "\n")
		}
	}

	sb.WriteRune('\n')
	sb.WriteString(styleHint.Render("  c copy logs   q/esc back"))
	sb.WriteRune('\n')
	if m.statusMsg != "" {
		sb.WriteString(styleGood.Render("  " + m.statusMsg) + "\n")
	}
	return sb.String()
}

func (m AppModel) handleLogViewerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		if len(m.logViewerLines) > 0 {
			_ = clipboard.WriteAll(strings.Join(m.logViewerLines, "\n"))
			m.statusMsg = "✓ logs copied to clipboard"
		} else {
			m.statusMsg = "no logs to copy"
		}
		return m, nil
	case "q", "esc", "enter":
		m.statusMsg = ""
		m.page = PageConfigPhase1
		if m.configPhase1Done {
			if m.configPhase1Only {
				m.page = PageConfigPhase1
			} else {
				m.page = PageConfigPhase2
			}
		}
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *AppModel) openLogViewer() {
	if scanLog == nil {
		return
	}
	dir := scanLog.LogDir()
	if dir == "" {
		return
	}
	m.statusMsg = ""
	m.logViewerLines = logger.GetRecentLogLines(dir, 200)
	m.page = PageLogViewer
}
