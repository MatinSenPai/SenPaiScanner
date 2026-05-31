package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/matinsenpai/senpaiscanner/internal/banner"
	"github.com/matinsenpai/senpaiscanner/internal/config"
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

// ColosDoneMsg signals the colo discovery finished.
type ColosDoneMsg struct{ ScanID int64 }

// tickMsg drives banner animation and stat refresh.
type tickMsg time.Time

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F6821F"))

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

// quickSetupRow identifies which row is focused on the Quick Scan setup page.
type quickSetupRow int

const (
	qRowCount   quickSetupRow = 0
	qRowWorkers quickSetupRow = 1
	qRowTimeout quickSetupRow = 2
)

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

	// quick scan setup (3-row picker)
	quickRow         quickSetupRow
	quickCountIdx    int
	quickWorkersIdx  int
	quickTimeoutIdx  int
	quickCustomInput textinput.Model
	quickCustomRow   quickSetupRow // which row triggered custom input
	quickCustomMode  bool

	// scan config form
	scanCfg    ScanConfig
	formInputs []textinput.Model
	formFocus  int
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

	// colos
	colosResults []*result.Result
	colosDone    bool

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
	configCustomRow     int    // 1=count, 2=workers, 3=timeout, 5=topN custom
	configCountCustom   string // value when Custom count is selected
	configWorkersCustom string // value when Custom workers is selected
	configTimeoutCustom string // value when Custom timeout is selected
	configTopNCustom string // value when Custom top N is selected
	configOptionalRow int   // 0=config URL, 1=validate top N
	configPortFocus     int
	configSelectedPorts map[int]bool
	// phase 1 state
	configPhase1Results []*result.Result
	configPhase1Done    bool
	configPhase1Only    bool // true when scan stops after Phase 1 (no config URL)
	configPhase1Stats   StatsMsg
	configPhase1Total   int // intended IP count for Phase 1 progress display
	liveResultPath      string

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
	{"About", ""},
	{"Quit", ""},
}

const menuLabelWidth = 16

const (
	menuFindWorking = 0
	menuAbout       = 1
	menuQuit        = 2
)

var modes = []string{"tls", "tcp", "http"}

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func NewApp(version string) AppModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#F6821F"))

	customInput := textinput.New()
	customInput.Placeholder = "e.g. 50000"
	customInput.CharLimit = 10
	customInput.Width = 14

	m := AppModel{
		page:             PageHome,
		spinner:          sp,
		scanCfg:          defaultScanConfig(),
		version:          version,
		width:            120,
		height:           40,
		scanStarted:      time.Now(),
		quickCustomInput: customInput,
		configWorkersIdx: 0,
		configTimeoutIdx: 2,
	}

	// Config input for "Scan with Config"
	cfgInput := textinput.New()
	cfgInput.Placeholder = "vless:// or trojan:// share URL"
	cfgInput.CharLimit = 2000
	cfgInput.Width = 0 // 0 = no fixed width, grows with content
	m.configInput = cfgInput

	cfgCustom := textinput.New()
	cfgCustom.Placeholder = "enter value"
	cfgCustom.CharLimit = 10
	cfgCustom.Width = 12
	m.configCustomInput = cfgCustom

	m.modeIdx = modeIndex(m.scanCfg.Mode)
	m.buildFormInputs()
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

func (m *AppModel) buildFormInputs() {
	fields := []struct{ placeholder, value string }{
		{"count (default 500)", m.scanCfg.Count},
		{"concurrency (default 50)", m.scanCfg.Concurrency},
		{"timeout (default 5s)", m.scanCfg.Timeout},
		{"tries per IP (default 4)", m.scanCfg.Tries},
		{"port (default 443)", m.scanCfg.Port},
		{"CIDR filter (e.g. 104.16.0.0/13, empty = all CF)", m.scanCfg.CIDR},
		{"output file (.csv/.json/.txt, empty = none)", m.scanCfg.OutputFile},
		{"colo filter (e.g. FRA,AMS, empty = all)", m.scanCfg.ColoFilter},
		{"SNI override (empty = auto-rotate)", m.scanCfg.SNI},
	}

	inputs := make([]textinput.Model, len(fields))
	for i, f := range fields {
		ti := textinput.New()
		ti.Placeholder = f.placeholder
		ti.SetValue(f.value)
		ti.CharLimit = 80
		ti.Width = 50
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	m.formInputs = inputs
	m.formFocus = 0
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
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

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
		if m.page == PageLiveColos {
			m.colosResults = append(m.colosResults, msg.Result)
		} else {
			m.scanResults = append(m.scanResults, msg.Result)
			result.Sort(m.scanResults, m.sortBy)
		}
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

	case ColosDoneMsg:
		if msg.ScanID == m.activeScanID {
			m.colosDone = true
		}
		return m, nil

	case ConfigProgressMsg:
		m.configResults = append(m.configResults, msg.Result)
		m.configTotal = msg.Total
		return m, nil

	case ConfigDoneMsg:
		m.configScanning = false
		m.configDone = true
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
	case PageQuickScanCount:
		return m.handleQuickCountKey(msg)
	case PageScanConfig:
		return m.handleConfigKey(msg)
	case PageLiveScan:
		return m.handleLiveScanKey(msg)
	case PageResults:
		return m.handleResultsKey(msg)
	case PageColos, PageLiveColos:
		return m.handleColosKey(msg)
	case PageAbout:
		if msg.String() == "q" || msg.String() == "esc" || msg.String() == "enter" {
			m.page = PageHome
		}
		return m, nil
	case PageScanWithConfig:
		return m.handleScanWithConfigKey(msg)
	case PageConfigOptional:
		return m.handleConfigOptionalKey(msg)
	case PageConfigPhase1:
		return m.handleConfigPhase1Key(msg)
	case PageConfigPhase2:
		return m.handleScanWithConfigKey(msg)
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
	case "enter", " ":
		return m.selectMenuItem()
	}
	return m, nil
}

func (m AppModel) selectMenuItem() (tea.Model, tea.Cmd) {
	switch m.menuIdx {
	case menuFindWorking:
		m.page = PageScanWithConfig
		m.configInput.SetValue("")
		m.configInput.Blur()
		m.configResults = nil
		m.configScanning = false
		m.configDone = false
		m.configSetupRow = 0
		m.configOptionalRow = 0
		m.configCountIdx = 1   // default: 5,000
		m.configTopNIdx = 2    // default: 50 for Phase 2
		m.configWorkersIdx = 0 // default: 50 workers (restricted-net safe)
		m.configTimeoutIdx = 2 // default: 5s
		m.configIPMode = 0     // default: random Cloudflare IPs
		m.configPortFocus = 0
		m.configSelectedPorts = nil
		m.configCustomMode = false
		m.configCountCustom = ""
		m.configWorkersCustom = ""
		m.configTimeoutCustom = ""
		m.configTopNCustom = ""
		m.configCustomInput.SetValue("")
		m.configCustomInput.Blur()
		m.configURL = ""
		m.configPhase1Only = false
		m.liveResultPath = ""
		clearLiveResultWriter()
		m.statusMsg = ""
		return m, nil
	case menuAbout:
		m.page = PageAbout
	case menuQuit:
		return m, tea.Quit
	}
	return m, nil
}

func (m AppModel) handleQuickCountKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If the user is typing a custom value, route keys there first.
	if m.quickCustomMode {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.quickCustomMode = false
			m.quickCustomInput.Blur()
			return m, nil
		case "enter":
			val := strings.TrimSpace(m.quickCustomInput.Value())
			return m.applyCustomValue(val)
		}
		return m.updateFormInputs(msg)
	}

	presets := m.presetsForRow(m.quickRow)
	idx := m.idxForRow(m.quickRow)

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.page = PageHome
	case "up", "k":
		if m.quickRow > 0 {
			m.quickRow--
		}
	case "down", "j":
		if int(m.quickRow) < 2 {
			m.quickRow++
		}
	case "left", "h":
		if idx > 0 {
			m.setIdxForRow(m.quickRow, idx-1)
		}
	case "right", "l":
		if idx < len(presets)-1 {
			m.setIdxForRow(m.quickRow, idx+1)
		}
	case "enter", " ":
		p := presets[idx]
		if p.value == "" {
			// Activate custom input for this row
			m.quickCustomRow = m.quickRow
			m.quickCustomMode = true
			m.quickCustomInput.SetValue("")
			m.quickCustomInput.Placeholder = m.customPlaceholderForRow(m.quickRow)
			m.quickCustomInput.CharLimit = m.customCharLimitForRow(m.quickRow)
			m.quickCustomInput.Focus()
			return m, textinput.Blink
		}
		// If all non-custom rows have a selection, launch
		return m.launchQuickScan()
	}
	return m, nil
}

func (m AppModel) presetsForRow(row quickSetupRow) []quickPreset {
	switch row {
	case qRowWorkers:
		return quickWorkersPresets
	case qRowTimeout:
		return quickTimeoutPresets
	default:
		return quickCountPresets
	}
}

func (m AppModel) idxForRow(row quickSetupRow) int {
	switch row {
	case qRowWorkers:
		return m.quickWorkersIdx
	case qRowTimeout:
		return m.quickTimeoutIdx
	default:
		return m.quickCountIdx
	}
}

func (m *AppModel) setIdxForRow(row quickSetupRow, idx int) {
	switch row {
	case qRowWorkers:
		m.quickWorkersIdx = idx
	case qRowTimeout:
		m.quickTimeoutIdx = idx
	default:
		m.quickCountIdx = idx
	}
}

func (m AppModel) customPlaceholderForRow(row quickSetupRow) string {
	switch row {
	case qRowWorkers:
		return "e.g. 150"
	case qRowTimeout:
		return "e.g. 4s"
	default:
		return "e.g. 50000"
	}
}

func (m AppModel) customCharLimitForRow(row quickSetupRow) int {
	switch row {
	case qRowTimeout:
		return 8
	default:
		return 10
	}
}

// applyCustomValue stores the typed value back into the right row index and
// advances to the next row or launches if on the last row.
func (m AppModel) applyCustomValue(val string) (tea.Model, tea.Cmd) {
	if val == "" {
		// restore placeholder default
		switch m.quickCustomRow {
		case qRowCount:
			val = "5000"
		case qRowWorkers:
			val = "100"
		case qRowTimeout:
			val = "3s"
		}
	}
	// Store in a dedicated custom-value slot by overwriting the last preset's value.
	// We use a simpler approach: just store in scanCfg directly and flag "custom used".
	switch m.quickCustomRow {
	case qRowCount:
		m.scanCfg.Count = val
		m.quickCountIdx = len(quickCountPresets) - 1 // keep "Custom" highlighted
	case qRowWorkers:
		m.scanCfg.Concurrency = val
		m.quickWorkersIdx = len(quickWorkersPresets) - 1
	case qRowTimeout:
		m.scanCfg.Timeout = val
		m.quickTimeoutIdx = len(quickTimeoutPresets) - 1
	}
	m.quickCustomMode = false
	m.quickCustomInput.Blur()
	if m.quickCustomRow < qRowTimeout {
		m.quickRow = m.quickCustomRow + 1
		return m, nil
	}
	m.quickRow = qRowTimeout
	return m.launchQuickScan()
}

func (m AppModel) customValueForRow(row quickSetupRow) string {
	switch row {
	case qRowWorkers:
		return m.scanCfg.Concurrency
	case qRowTimeout:
		return m.scanCfg.Timeout
	default:
		return m.scanCfg.Count
	}
}

func (m AppModel) customLabelForRow(row quickSetupRow) string {
	value := strings.TrimSpace(m.customValueForRow(row))
	if value == "" {
		return "Custom"
	}
	return "Custom: " + value
}

func (m AppModel) customHelpForRow(row quickSetupRow) string {
	switch row {
	case qRowWorkers:
		return "type an integer worker count, e.g. 75 or 150"
	case qRowTimeout:
		return "type a Go duration, e.g. 4s, 1500ms, 8s"
	default:
		return "type an integer IP count, e.g. 50000"
	}
}

func (m AppModel) launchQuickScan() (tea.Model, tea.Cmd) {
	cfg := defaultScanConfig()

	// Count
	cp := quickCountPresets[m.quickCountIdx]
	if cp.value != "" {
		cfg.Count = cp.value
	} else if m.scanCfg.Count != "" {
		cfg.Count = m.scanCfg.Count
	}

	// Workers
	wp := quickWorkersPresets[m.quickWorkersIdx]
	if wp.value != "" {
		cfg.Concurrency = wp.value
	} else if m.scanCfg.Concurrency != "" {
		cfg.Concurrency = m.scanCfg.Concurrency
	}

	// Timeout
	tp := quickTimeoutPresets[m.quickTimeoutIdx]
	if tp.value != "" {
		cfg.Timeout = tp.value
	} else if m.scanCfg.Timeout != "" {
		cfg.Timeout = m.scanCfg.Timeout
	}

	m.scanCfg = cfg
	m.activeScanID = nextScanID()
	m.statusMsg = ""
	m.scanResults = nil
	m.scanDone = false
	m.scanStats = StatsMsg{ScanID: m.activeScanID}
	m.scanStarted = time.Now()
	n, _ := fmt.Sscanf(cfg.Count, "%d", &m.scanTotal)
	if n == 0 {
		m.scanTotal = 0
	}
	m.page = PageLiveScan
	return m, StartScanCmd(cfg, m.activeScanID)
}

func (m AppModel) handleConfigKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.page = PageHome
		return m, nil
	case "tab", "down":
		m.formFocus = (m.formFocus + 1) % len(m.formInputs)
		for i := range m.formInputs {
			if i == m.formFocus {
				m.formInputs[i].Focus()
			} else {
				m.formInputs[i].Blur()
			}
		}
	case "shift+tab", "up":
		m.formFocus = (m.formFocus - 1 + len(m.formInputs)) % len(m.formInputs)
		for i := range m.formInputs {
			if i == m.formFocus {
				m.formInputs[i].Focus()
			} else {
				m.formInputs[i].Blur()
			}
		}
	case "ctrl+left", "ctrl+right":
		if msg.String() == "ctrl+right" {
			m.modeIdx = (m.modeIdx + 1) % len(modes)
		} else {
			m.modeIdx = (m.modeIdx - 1 + len(modes)) % len(modes)
		}
		m.scanCfg.Mode = modes[m.modeIdx]
	case "f2":
		m.scanCfg.UseV4 = !m.scanCfg.UseV4
	case "f3":
		m.scanCfg.UseV6 = !m.scanCfg.UseV6
	case "enter":
		m.saveScanConfig()
		m.activeScanID = nextScanID()
		m.statusMsg = ""
		m.scanResults = nil
		m.scanDone = false
		m.scanStats = StatsMsg{ScanID: m.activeScanID}
		m.scanStarted = time.Now()
		n, _ := fmt.Sscanf(m.scanCfg.Count, "%d", &m.scanTotal)
		if n == 0 {
			m.scanTotal = 0
		}
		m.page = PageLiveScan
		return m, StartScanCmd(m.scanCfg, m.activeScanID)
	}
	return m.updateFormInputs(msg)
}

func (m *AppModel) saveScanConfig() {
	if len(m.formInputs) >= 9 {
		m.scanCfg.Count = m.formInputs[0].Value()
		m.scanCfg.Concurrency = m.formInputs[1].Value()
		m.scanCfg.Timeout = m.formInputs[2].Value()
		m.scanCfg.Tries = m.formInputs[3].Value()
		m.scanCfg.Port = m.formInputs[4].Value()
		m.scanCfg.CIDR = m.formInputs[5].Value()
		m.scanCfg.OutputFile = m.formInputs[6].Value()
		m.scanCfg.ColoFilter = m.formInputs[7].Value()
		m.scanCfg.SNI = m.formInputs[8].Value()
		m.scanCfg.Mode = modes[m.modeIdx]
	}
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
		m.statusMsg = "use Find Working IPs to copy config-tested IPs"
	}
	return m, nil
}

var clipboardWriteAll = clipboard.WriteAll

// copyHealthyIPsToClipboard writes one IP per line to the system clipboard
// and returns a short status message to display to the user.
func (m AppModel) copyHealthyIPsToClipboard() string {
	top := result.TopN(m.scanResults, 0) // all healthy IPs, sorted by avg
	if len(top) == 0 {
		return "no healthy IPs to copy"
	}
	var sb strings.Builder
	for _, r := range top {
		sb.WriteString(r.IP.String())
		sb.WriteRune('\n')
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
	return copyAndSaveIPs(endpoints)
}

func workingIPs(results []*xraytest.ValidationResult) []string {
	return workingEndpoints(results)
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

func copyAndSaveIPs(ips []string) string {
	text := strings.Join(ips, "\n") + "\n"
	clipErr := clipboardWriteAll(text)
	path, fileErr := writeIPsBesideExecutable(ips)

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
	exe, err := os.Executable()
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
	path := filepath.Join(dir, "ips.txt")
	if err := writeIPsFile(path, ips); err == nil {
		return path, nil
	}

	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return path, wdErr
	}
	fallback := filepath.Join(wd, "ips.txt")
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

func (m AppModel) handleColosKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q", "esc", "enter":
		if m.colosDone || m.page == PageColos {
			m.page = PageHome
		}
	}
	return m, nil
}

// updateFormInputs forwards non-key messages (e.g. paste events, resize) to
// every focused text input so they can handle them independently.
func (m AppModel) updateFormInputs(msg tea.Msg) (AppModel, tea.Cmd) {
	var cmds []tea.Cmd

	if m.page == PageQuickScanCount && m.quickCustomMode {
		var cmd tea.Cmd
		m.quickCustomInput, cmd = m.quickCustomInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.page == PageScanConfig && len(m.formInputs) > 0 {
		for i := range m.formInputs {
			var cmd tea.Cmd
			m.formInputs[i], cmd = m.formInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

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
		} else {
			var cmd tea.Cmd
			m.configInput, cmd = m.configInput.Update(msg)
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
	case PageQuickScanCount:
		return m.viewQuickScanCount()
	case PageScanConfig:
		return m.viewScanConfig()
	case PageLiveScan:
		return m.viewLiveScan()
	case PageResults:
		return m.viewResults()
	case PageLiveColos:
		return m.viewLiveColos()
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

	// Version — keep newlines outside styled output; lipgloss pads blank lines with spaces.
	sb.WriteString(styleDim.Render(fmt.Sprintf("  v%s", m.version)))
	sb.WriteString("\n\n")

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
// Quick Scan setup page (3-row picker: Count / Workers / Timeout)
// ---------------------------------------------------------------------------

func (m AppModel) viewQuickScanCount() string {
	var sb strings.Builder

	separator := fmt.Sprintf("  %v\n\n", strings.Repeat("─", 64))

	sb.WriteString(banner.Render(m.bannerFrame / 2))
	sb.WriteRune('\n')
	sb.WriteString(styleTitle.Render("  ⚡  Quick Scan Setup\n"))
	sb.WriteString(separator)

	type rowDef struct {
		label   string
		presets []quickPreset
		selIdx  int
		row     quickSetupRow
		hint    string
	}

	rows := []rowDef{
		{
			label:   "  Count   ",
			presets: quickCountPresets,
			selIdx:  m.quickCountIdx,
			row:     qRowCount,
			hint:    "number of Cloudflare IPs to probe",
		},
		{
			label:   "  Workers ",
			presets: quickWorkersPresets,
			selIdx:  m.quickWorkersIdx,
			row:     qRowWorkers,
			hint:    "parallel goroutines — higher is faster but harder on slow networks",
		},
		{
			label:   "  Timeout ",
			presets: quickTimeoutPresets,
			selIdx:  m.quickTimeoutIdx,
			row:     qRowTimeout,
			hint:    "per-probe deadline — raise this if you see lots of timeouts",
		},
	}

	for _, r := range rows {
		focused := m.quickRow == r.row
		labelStyle := styleHeader
		if focused {
			labelStyle = styleAccent
		}
		sb.WriteString(labelStyle.Render(r.label))

		// Render preset pills
		for i, p := range r.presets {
			label := strings.SplitN(p.label, " —", 2)[0] // short label only
			// trim trailing spaces
			label = strings.TrimRight(label, " ")
			if p.value == "" {
				label = m.customLabelForRow(r.row)
			}

			if i == r.selIdx {
				if p.value == "" && m.quickCustomMode && m.quickCustomRow == r.row {
					// Active custom input
					sb.WriteString(fmt.Sprintf("%s%s%s",
						styleAccent.Render("["),
						m.quickCustomInput.View(),
						styleAccent.Render("]"),
					))
				} else {
					sb.WriteString(styleSelected.Render(fmt.Sprintf(" %s ", label)))
				}
			} else {
				sb.WriteString(styleDim.Render(fmt.Sprintf(" %s ", label)))
			}
			if i < len(r.presets)-1 {
				sb.WriteString(styleSep.Render(" │ "))
			}
		}
		sb.WriteRune('\n')

		// Show hint only for the focused row
		if focused {
			sb.WriteString(styleHint.Render("    " + r.hint + "\n"))
		}
		sb.WriteRune('\n')
	}

	if m.quickCustomMode {
		sb.WriteString(styleHint.Render("  " + m.customHelpForRow(m.quickCustomRow) + "   enter confirm   esc cancel"))
	} else {
		sb.WriteString(styleHint.Render("  ↑/↓ row   ←/→ option   enter select/start   esc back"))
	}
	sb.WriteRune('\n')
	return sb.String()
}

// ---------------------------------------------------------------------------
// Scan Config page
// ---------------------------------------------------------------------------

func (m AppModel) viewScanConfig() string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("\n  ⚙  Custom Scan Configuration\n"))
	sb.WriteString(fmt.Sprintf("%s\n\n",
		styleSep.Render("  "+strings.Repeat("─", 56)),
	))

	labels := []string{
		"Count      ", "Workers    ", "Timeout    ", "Tries      ", "Port       ",
		"CIDR       ", "Output     ", "Colo Filter", "SNI        ",
	}

	for i, inp := range m.formInputs {
		prefix := "  "
		label := styleHeader.Render(labels[i] + "  ")
		if i == m.formFocus {
			prefix = styleAccent.Render("  ▶ ")
			label = styleAccent.Render(labels[i] + "  ")
		}
		sb.WriteString(fmt.Sprintf("%s%s%s\n", prefix, label, inp.View()))
	}

	// Mode toggle
	sb.WriteRune('\n')
	sb.WriteString(styleHeader.Render("  Mode        "))
	for i, mode := range modes {
		if i == m.modeIdx {
			sb.WriteString(styleSelected.Render(fmt.Sprintf(" %s ", strings.ToUpper(mode))))
		} else {
			sb.WriteString(styleDim.Render(fmt.Sprintf(" %s ", strings.ToUpper(mode))))
		}
		sb.WriteString("  ")
	}
	sb.WriteString(fmt.Sprintf("%s\n", styleDim.Render("  ←/→ to cycle")))

	// IPv4/v6 toggles
	v4s := styleGood.Render("ON")
	if !m.scanCfg.UseV4 {
		v4s = styleBad.Render("OFF")
	}
	v6s := styleGood.Render("ON")
	if !m.scanCfg.UseV6 {
		v6s = styleBad.Render("OFF")
	}
	sb.WriteString(fmt.Sprintf("%s%s%s\n", styleHeader.Render("  IPv4         "), v4s, styleDim.Render("  F2 toggle")))
	sb.WriteString(fmt.Sprintf("%s%s%s\n", styleHeader.Render("  IPv6         "), v6s, styleDim.Render("  F3 toggle")))

	sb.WriteRune('\n')
	sb.WriteString(styleHint.Render("  tab/↑↓ navigate   enter start scan   esc back"))
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
	hint := fmt.Sprintf("  s sort(→%s)   c copy IPs   q/esc back", sortNames[m.sortIdx%5])
	if m.scanDone {
		hint = fmt.Sprintf("  s sort(→%s)   c copy IPs   enter/q → results", sortNames[m.sortIdx%5])
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
	sb.WriteString(styleHint.Render("  s sort   c copy IPs   enter/q → home menu"))
	return sb.String()
}

// ---------------------------------------------------------------------------
// Live Colos page
// ---------------------------------------------------------------------------

func (m AppModel) viewLiveColos() string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("\n  🌍  Discovering Cloudflare PoPs\n"))
	sb.WriteString(fmt.Sprintf("%s\n\n", styleSep.Render("  "+strings.Repeat("─", 56))))

	if !m.colosDone {
		sb.WriteString(fmt.Sprintf("  %s probing IPs via /cdn-cgi/trace…\n\n", m.spinner.View()))
	} else {
		sb.WriteString(styleGood.Render("  ✓ Discovery complete\n\n"))
	}

	PrintColoTableBuf(&sb, m.colosResults)

	sb.WriteRune('\n')
	sb.WriteString(styleHint.Render("  q/esc → home menu"))
	return sb.String()
}

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
	sb.WriteString(styleHint.Render("  enter/q → back"))
	return sb.String()
}

// ---------------------------------------------------------------------------
// Exported helpers for non-TUI callers
// ---------------------------------------------------------------------------

// PrintTable prints a sorted results table to stdout.
func PrintTable(results []*result.Result, top int) {
	sorted := make([]*result.Result, len(results))
	copy(sorted, results)
	result.Sort(sorted, result.SortByAvg)
	if top > 0 && top < len(sorted) {
		sorted = sorted[:top]
	}

	hdr := fmt.Sprintf("  %-18s  %7s  %9s  %8s  %9s  %4s  %-5s",
		"IP", "LOSS", "AVG(ms)", "JTR(ms)", "DL(KB/s)", "TLS", "COLO")
	fmt.Println(hdr)
	fmt.Println("  " + strings.Repeat("─", 72))
	for _, r := range sorted {
		tls := "✗"
		if r.TLSOk {
			tls = "✓"
		}
		colo := r.Colo
		if colo == "" {
			colo = "—"
		}
		fmt.Printf("  %-18s  %6.1f%%  %9.2f  %8.2f  %9.1f  %4s  %-5s\n",
			r.IP.String(), r.Loss(),
			float64(r.Avg().Milliseconds()),
			float64(r.Jitter().Milliseconds()),
			r.Throughput/1024,
			tls, colo)
	}
	fmt.Println()
}

// SimpleProgress prints a one-liner progress update.
func SimpleProgress(tested, healthy, total int64) {
	if total > 0 {
		fmt.Printf("\r  tested: %d/%d (%.0f%%)  healthy: %d",
			tested, total, float64(tested)/float64(total)*100, healthy)
	} else {
		fmt.Printf("\r  tested: %d  healthy: %d", tested, healthy)
	}
}

// PrintColoTableBuf writes a colo summary into sb.
func PrintColoTableBuf(sb *strings.Builder, results []*result.Result) {
	type cs struct {
		count  int
		avgSum int64
		bestMS int64
		bestIP string
	}
	byC := map[string]*cs{}
	for _, r := range results {
		if !r.IsHealthy() || r.Colo == "" {
			continue
		}
		s, ok := byC[r.Colo]
		if !ok {
			s = &cs{bestMS: r.Avg().Milliseconds(), bestIP: r.IP.String()}
			byC[r.Colo] = s
		}
		s.count++
		s.avgSum += r.Avg().Milliseconds()
		if r.Avg().Milliseconds() < s.bestMS {
			s.bestMS = r.Avg().Milliseconds()
			s.bestIP = r.IP.String()
		}
	}
	if len(byC) == 0 {
		sb.WriteString(styleDim.Render("  No colos discovered yet…\n"))
		return
	}
	type row struct {
		colo   string
		count  int
		avgMs  float64
		bestMs int64
		bestIP string
	}
	var rows []row
	for colo, s := range byC {
		rows = append(rows, row{colo, s.count, float64(s.avgSum) / float64(s.count), s.bestMS, s.bestIP})
	}
	// sort by bestMs
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].bestMs < rows[j-1].bestMs; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
	sb.WriteString(styleHeader.Render(fmt.Sprintf("  %-6s  %6s  %9s  %9s  %s\n",
		"COLO", "COUNT", "AVG(ms)", "BEST(ms)", "BEST IP")))
	sb.WriteString(styleSep.Render("  " + strings.Repeat("─", 52) + "\n"))
	for _, r := range rows {
		line := fmt.Sprintf("  %-6s  %6d  %9.2f  %9d  %s\n",
			r.colo, r.count, r.avgMs, r.bestMs, r.bestIP)
		sb.WriteString(styleGood.Render(line))
	}
}

// ColoTable prints colo summary to stdout.
func ColoTable(results []*result.Result) {
	var sb strings.Builder
	PrintColoTableBuf(&sb, results)
	fmt.Print(sb.String())
}

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

	sb.WriteString(styleTitle.Render("\n  ⚡  Find Working IPs\n"))
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
			sb.WriteString(styleDim.Render(fmt.Sprintf("            IPs to probe in Phase 1  (custom: %s)", m.configCountCustom)) + "\n\n")
		} else if m.configIPMode == 1 {
			sb.WriteString(styleDim.Render("            ignored when Source is From File — all IPs in ips.txt are used") + "\n\n")
		} else {
			sb.WriteString(styleDim.Render("            IPs to probe in Phase 1") + "\n\n")
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
		sb.WriteString(styleDim.Render("            space toggles a port; selecting multiple ports multiplies work") + "\n\n")

		hint := "  ↑/↓ row   ←/→ option   enter continue   esc back"
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
		sb.WriteString(fmt.Sprintf("  %s  xray validating candidates  %s\n\n",
			styleAccent.Render(scanPulse(m.bannerFrame)),
			scanWave(m.bannerFrame+5, 32),
		))
	}

	// Table header
	hdr := fmt.Sprintf("  %-22s  %-8s  %8s  %8s  %6s",
		"ENDPOINT", "TYPE", "SPEED", "LATENCY", "STATUS")
	sb.WriteString(fmt.Sprintf("%s\n%s\n", styleHeader.Render(hdr), styleSep.Render("  "+strings.Repeat("─", 64))))

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
			line := fmt.Sprintf("  %-22s  %-8s  %8s  %8s  %6s",
				formatEndpoint(r.IP, r.Port), r.Transport, formatValidationSpeed(r.Throughput), formatValidationLatency(r.Latency), "✓")
			sb.WriteString(styleGood.Render(line) + "\n")
		} else {
			errMsg := r.Error
			if len(errMsg) > 20 {
				errMsg = errMsg[:20] + "…"
			}
			line := fmt.Sprintf("  %-22s  %-8s  %9s  %8s  %6s",
				formatEndpoint(r.IP, r.Port), r.Transport, "—", "—", "✗")
			sb.WriteString(styleBad.Render(line) + "\n")
		}
	}

	sb.WriteRune('\n')
	if m.statusMsg != "" {
		sb.WriteString(styleGood.Render("  "+m.statusMsg) + "\n")
	}
	if m.configDone {
		hint := "  c copy working endpoints   q/esc back to menu"
		if m.liveResultPath != "" {
			hint += "\n" + styleDim.Render("  live results → "+m.liveResultPath)
		}
		sb.WriteString(styleHint.Render(hint) + "\n")
	} else if m.liveResultPath != "" {
		sb.WriteString(styleDim.Render("  live results → "+m.liveResultPath) + "\n")
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

func (m AppModel) handleScanWithConfigKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		if m.configScanning || m.configDone {
			clearLiveResultWriter()
			m.page = PageHome
			m.configScanning = false
			m.configDone = false
			return m, nil
		}
		m.page = PageHome
		return m, nil
	case "q":
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
	}

	if m.configScanning || m.configDone {
		return m, nil
	}

	// --- Setup navigation (Source → Count → Workers → Timeout → Ports) ---
	const maxRow = 4

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
		if m.configSetupRow == 4 {
			m.toggleFocusedConfigPort()
			return m, nil
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

		m.statusMsg = ""
		m.page = PageConfigOptional
		m.configOptionalRow = 0
		m.configInput.Focus()
		return m, textinput.Blink
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

	rowLabel(0, "  Config ")
	sb.WriteString(" " + m.configInput.View() + "\n")
	sb.WriteString(styleDim.Render("            optional — leave empty for Phase 1 only") + "\n\n")

	rowLabel(1, "  Top N  ")
	sb.WriteString(" ")
	renderPills(configTopNLabels, m.configTopNIdx)
	sb.WriteString("\n")
	if m.configCustomMode && m.configCustomRow == 5 {
		sb.WriteString(styleAccent.Render("            custom top N: ") + m.configCustomInput.View() + "\n\n")
	} else if m.isTopNCustomSelected() && m.configTopNCustom != "" {
		sb.WriteString(styleDim.Render(fmt.Sprintf("            Phase 2 candidates to validate  (custom: %s)", m.configTopNCustom)) + "\n\n")
	} else {
		sb.WriteString(styleDim.Render("            Phase 2 picks — used only when a config URL is entered") + "\n\n")
	}

	hint := "  ↑/↓ row   ←/→ option   enter start scan   esc back"
	if m.configOptionalRow == 0 {
		hint = "  paste URL or leave empty   enter start   ↓ navigate   esc back"
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
			if m.configCustomRow == 5 {
				m.configTopNCustom = strings.TrimSpace(m.configCustomInput.Value())
			}
			m.configCustomMode = false
			m.configCustomInput.Blur()
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

	switch msg.String() {
	case "esc":
		m.page = PageScanWithConfig
		m.configInput.Blur()
		return m, nil
	case "up", "k":
		if m.configOptionalRow > 0 {
			m.configOptionalRow--
			if m.configOptionalRow == 0 {
				m.configInput.Focus()
				return m, textinput.Blink
			}
			m.configInput.Blur()
		}
		return m, nil
	case "down", "j":
		if m.configOptionalRow == 0 {
			m.configOptionalRow = 1
			m.configInput.Blur()
			return m, nil
		}
		return m, nil
	case "left", "h":
		if m.configOptionalRow == 1 {
			if m.configTopNIdx > 0 {
				m.configTopNIdx--
			}
			return m, nil
		}
	case "right", "l":
		if m.configOptionalRow == 1 {
			if m.configTopNIdx < len(configTopNLabels)-1 {
				m.configTopNIdx++
			}
			return m, nil
		}
	case "enter":
		if m.configOptionalRow == 0 {
			if strings.TrimSpace(m.configInput.Value()) == "" {
				return m.launchPhase1FromOptional()
			}
			m.configOptionalRow = 1
			m.configInput.Blur()
			return m, nil
		}
		if m.isTopNCustomSelected() {
			m.configCustomMode = true
			m.configCustomRow = 5
			m.configCustomInput.SetValue(m.configTopNCustom)
			m.configCustomInput.Placeholder = "e.g. 75"
			m.configCustomInput.Focus()
			return m, textinput.Blink
		}
		return m.launchPhase1FromOptional()
	}

	if m.configOptionalRow == 0 {
		var cmd tea.Cmd
		m.configInput, cmd = m.configInput.Update(msg)
		return m, cmd
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

// ConfigDoneMsg signals all config validations are complete.
type ConfigDoneMsg struct{}

// ConfigBatchResultMsg is no longer used — results come one by one via ConfigProgressMsg.
type ConfigBatchResultMsg struct {
	Results []*xraytest.ValidationResult
}

// ConfigProgressMsg carries a single result during scanning.
type ConfigProgressMsg struct {
	Result *xraytest.ValidationResult
	Done   int
	Total  int
}

func (m AppModel) startConfigScan(rawURL string) tea.Cmd {
	return func() tea.Msg {
		go runConfigScan(rawURL)
		return nil
	}
}

func runConfigScan(rawURL string) {
	cfg, err := xraytest.ParseProxyURL(rawURL)
	if err != nil {
		if prog != nil {
			prog.Send(ConfigDoneMsg{})
		}
		return
	}

	// Top CF IPs to test
	testIPs := []string{
		"104.18.5.1", "104.17.0.1", "172.66.40.1",
		"172.67.186.127", "104.21.19.146", "104.16.0.1",
		"104.19.229.21", "104.18.10.1", "104.17.100.1",
		"104.16.200.1",
	}

	ctx := context.Background()
	total := len(testIPs)

	for i, ip := range testIPs {
		swapped := cfg.WithAddress(ip)
		r := xraytest.ValidateConfig(ctx, swapped, 20*time.Second)
		if prog != nil {
			prog.Send(ConfigProgressMsg{
				Result: r,
				Done:   i + 1,
				Total:  total,
			})
		}
	}

	if prog != nil {
		prog.Send(ConfigDoneMsg{})
	}
}

// ---------------------------------------------------------------------------
// Config Setup presets
// ---------------------------------------------------------------------------

var configCountValues = []int{1000, 5000, 20000, 0} // 0 = custom
var configCountLabels = []string{"1,000", "5,000", "20,000", "Custom"}
var configTopNValues = []int{10, 25, 50, 100, 0} // 0 = all
var configTopNLabels = []string{"10", "25", "50", "100", "All", "Custom"}
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

func configPortLabels() []string {
	labels := make([]string, len(configPortChoices))
	for i, p := range configPortChoices {
		labels[i] = p.label
	}
	return labels
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
// Config Setup page
// ---------------------------------------------------------------------------

func (m AppModel) viewConfigSetup() string {
	var sb strings.Builder

	sb.WriteString(styleTitle.Render("\n  ⚡  Scan with Config — Setup\n"))
	sb.WriteString(fmt.Sprintf("%s\n\n", styleSep.Render("  "+strings.Repeat("─", minInt(m.width-4, 70)))))

	sb.WriteString(styleNormal.Render("  Phase 1: Fast connectivity scan to find reachable IPs") + "\n")
	sb.WriteString(styleNormal.Render("  Phase 2: Test top IPs with your actual xray config") + "\n\n")

	// Count row
	countLabel := "  Count   "
	for i, label := range configCountLabels {
		if i == m.configCountIdx && m.configSetupRow == 0 {
			sb.WriteString(styleSelected.Render(" " + label + " "))
		} else {
			sb.WriteString(styleNormal.Render("  " + label + "  "))
		}
		if i < len(configCountLabels)-1 {
			sb.WriteString(styleDim.Render("│"))
		}
	}
	sb.WriteString("\n")
	if m.configSetupRow == 0 {
		sb.WriteString(styleAccent.Render(countLabel) + styleDim.Render("IPs to probe in Phase 1") + "\n\n")
	} else {
		sb.WriteString(styleDim.Render(countLabel+"IPs to probe in Phase 1") + "\n\n")
	}

	// Top N row
	topLabel := "  Top N   "
	for i, label := range configTopNLabels {
		if i == m.configTopNIdx && m.configSetupRow == 1 {
			sb.WriteString(styleSelected.Render(" " + label + " "))
		} else {
			sb.WriteString(styleNormal.Render("  " + label + "  "))
		}
		if i < len(configTopNLabels)-1 {
			sb.WriteString(styleDim.Render("│"))
		}
	}
	sb.WriteString("\n")
	if m.configSetupRow == 1 {
		sb.WriteString(styleAccent.Render(topLabel) + styleDim.Render("best IPs to validate with xray") + "\n\n")
	} else {
		sb.WriteString(styleDim.Render(topLabel+"best IPs to validate with xray") + "\n\n")
	}

	sb.WriteString(styleHint.Render("  ↑/↓ row   ←/→ option   enter start   esc back") + "\n")

	return sb.String()
}

func (m AppModel) handleConfigSetupKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.page = PageScanWithConfig
		return m, nil
	case "up", "k":
		if m.configSetupRow > 0 {
			m.configSetupRow--
		}
	case "down", "j":
		if m.configSetupRow < 1 {
			m.configSetupRow++
		}
	case "left", "h":
		if m.configSetupRow == 0 && m.configCountIdx > 0 {
			m.configCountIdx--
		} else if m.configSetupRow == 1 && m.configTopNIdx > 0 {
			m.configTopNIdx--
		}
	case "right", "l":
		if m.configSetupRow == 0 && m.configCountIdx < len(configCountLabels)-1 {
			m.configCountIdx++
		} else if m.configSetupRow == 1 && m.configTopNIdx < len(configTopNLabels)-1 {
			m.configTopNIdx++
		}
	case "enter":
		// Start Phase 1
		m.page = PageConfigPhase1
		m.configPhase1Results = nil
		m.configPhase1Done = false
		m.configPhase1Stats = StatsMsg{}
		count := configCountValues[m.configCountIdx]
		if count == 0 {
			count, _ = strconv.Atoi(m.configCountCustom)
			if count <= 0 {
				count = 1000
			}
		}
		m.configPhase1Total = count
		return m, m.startConfigPhase1()
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Config Phase 1 — fast connectivity scan
// ---------------------------------------------------------------------------

type ConfigPhase1ResultMsg struct {
	Result *result.Result
}

// ConfigPhase1ErrMsg is sent when Phase 1 cannot proceed (e.g. ips.txt missing).
type ConfigPhase1ErrMsg struct{ Err string }

type ConfigPhase1StatsMsg = StatsMsg

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
			sb.WriteString(styleGood.Render(fmt.Sprintf("  Done — %d healthy endpoints found.\n\n", healthy)))
		} else {
			topN := m.resolveTopN()
			label := fmt.Sprintf("%d", topN)
			if topN == 0 {
				label = "all"
			}
			sb.WriteString(styleGood.Render(fmt.Sprintf("  Found %d candidates. Testing top %s with xray...\n\n", healthy, label)))
		}
	} else if m.configIPMode == 1 {
		sb.WriteString(styleNormal.Render("  Probing IPs from ips.txt on the selected ports...\n\n"))
	} else if strings.TrimSpace(m.configURL) == "" {
		sb.WriteString(styleNormal.Render("  Scanning random Cloudflare IPv4 IPs (standard HTTP probe)...\n"))
		sb.WriteString(styleDim.Render("  healthy hits also explore nearby addresses in the same Cloudflare block\n\n"))
	} else {
		sb.WriteString(styleNormal.Render("  Scanning Cloudflare IPs using your config probe settings...\n"))
		sb.WriteString(styleDim.Render("  healthy hits also explore nearby addresses in the same Cloudflare block\n\n"))
	}

	if m.liveResultPath != "" {
		sb.WriteString(styleDim.Render("  live results → "+m.liveResultPath+"\n\n"))
	}

	if len(m.configPhase1Results) > 0 {
		hdr := fmt.Sprintf("  %-22s  %7s  %9s  %8s  %6s",
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
		sb.WriteString(styleHint.Render("  c copy healthy endpoints   q/esc back") + "\n")
	} else {
		sb.WriteString(styleHint.Render("  q/esc cancel") + "\n")
	}
	return sb.String()
}

func (m AppModel) handleConfigPhase1Key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		if m.configPhase1Done && m.configPhase1Only {
			m.statusMsg = m.copyPhase1HealthyEndpoints()
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
	return copyAndSaveIPs(endpoints)
}

// configPhase1Options holds the resolved settings for a Phase 1 engine run.
type configPhase1Options struct {
	count       int
	concurrency int
	timeout     time.Duration
	rawURL      string
	ports       []int
	fromFile    bool
}

func (m AppModel) startConfigPhase1() tea.Cmd {
	opts := m.resolvePhase1Options()
	return func() tea.Msg {
		go runConfigPhase1(opts)
		return nil
	}
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

	return configPhase1Options{
		count:       count,
		concurrency: concurrency,
		timeout:     timeout,
		rawURL:      m.configURL,
		ports:       m.resolveConfigPorts(),
		fromFile:    m.configIPMode == 1,
	}
}

func (m AppModel) phase1TargetTotal(count int) int {
	ports := len(m.resolveConfigPorts())
	if ports <= 0 {
		ports = 1
	}
	if m.configIPMode == 1 {
		if ips, err := loadDefaultIPsFile(); err == nil {
			return len(ips) * ports
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
	return func() tea.Msg {
		go runConfigPhase2(url, topIPs)
		return nil
	}
}

func runConfigPhase2(rawURL string, topIPs []*result.Result) {
	cfg, err := xraytest.ParseProxyURL(rawURL)
	if err != nil {
		if prog != nil {
			prog.Send(ConfigDoneMsg{})
		}
		return
	}

	ctx := context.Background()
	total := len(topIPs)

	for i, r := range topIPs {
		ip := r.IP.String()
		swapped := cfg.WithEndpoint(ip, r.Port)
		vr := xraytest.ValidateConfig(ctx, swapped, 20*time.Second)
		if liveResultWriter != nil {
			liveResultWriter.AddPhase2(vr)
		}
		if prog != nil {
			prog.Send(ConfigProgressMsg{
				Result: vr,
				Done:   i + 1,
				Total:  total,
			})
		}
	}

	if prog != nil {
		prog.Send(ConfigDoneMsg{})
	}
}
