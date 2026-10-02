package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

var liveResultWriter *LiveResultWriter

func setLiveResultWriter(w *LiveResultWriter) { liveResultWriter = w }

func clearLiveResultWriter() { liveResultWriter = nil }

// LiveResultWriter appends scan results to a text file the user can open while
// the scan runs. The file is rewritten on each update so external viewers refresh.
type LiveResultWriter struct {
	mu sync.Mutex

	path         string
	started      time.Time
	withConfig   bool
	phase        int
	phase1Only   bool
	phase1Done   bool
	phase1Rows   []*result.Result
	phase2Rows   []*xraytest.ValidationResult
	phase1Probed int
}

func newLiveResultWriter(withConfig bool) (*LiveResultWriter, string, error) {
	path, err := liveResultFilePath()
	if err != nil {
		return nil, "", err
	}
	w := &LiveResultWriter{
		path:       path,
		started:    time.Now(),
		withConfig: withConfig,
		phase:      1,
	}
	return w, path, nil
}

func liveResultFilePath() (string, error) {
	name := fmt.Sprintf("SenPaiScannerResult-%s.txt", time.Now().Format("20060102-150405"))
	for _, dir := range resultFileDirs() {
		if dir == "" {
			continue
		}
		return filepath.Join(dir, name), nil
	}
	return name, nil
}

func resultFileDirs() []string {
	seen := make(map[string]struct{})
	var dirs []string
	add := func(dir string) {
		if dir == "" {
			return
		}
		if _, ok := seen[dir]; ok {
			return
		}
		seen[dir] = struct{}{}
		dirs = append(dirs, dir)
	}
	if wd, err := os.Getwd(); err == nil {
		add(wd)
	}
	if exe, err := os.Executable(); err == nil {
		add(filepath.Dir(exe))
	}
	return dirs
}

func (w *LiveResultWriter) AddPhase1(r *result.Result) {
	if w == nil || r == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.phase1Probed++
	// WithinLossLimit is checked here, at the door, so an over-loss IP never
	// reaches the file. Filtering later leaves a window where a reader tailing
	// the live file sees an IP the final ranking will drop.
	if r.IsHealthy() && r.WithinLossLimit() {
		w.phase1Rows = append(w.phase1Rows, r)
	}
	_ = w.writeLocked()
}

func (w *LiveResultWriter) BeginPhase2() {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.phase = 2
	w.phase1Done = true
	w.phase2Rows = nil
	_ = w.writeLocked()
}

func (w *LiveResultWriter) FinishPhase1Only() {
	if w == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.phase1Only = true
	w.phase1Done = true
	_ = w.writeLocked()
}

func (w *LiveResultWriter) AddPhase2(v *xraytest.ValidationResult) {
	if w == nil || v == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.phase2Rows = append(w.phase2Rows, v)
	_ = w.writeLocked()
}

// endpointKey is the lookup key shared by the Phase 1 and Phase 2 tables.
func endpointKey(ip string, port int) string {
	return fmt.Sprintf("%s:%d", ip, port)
}

// phase1Index indexes Phase 1 rows by endpoint for O(1) metric lookup, and
// carries the Phase 1 display order for ordering Phase 2 rows.
//
// A lookup map rather than a scan: the Phase 2 table asks for the Phase 1
// metrics of every visible row on every rendered frame, and a linear scan
// building a key per comparison turned that into thousands of Sprintf calls
// per frame.
type phase1Index struct {
	byEndpoint map[string]*result.Result
	rank       map[string]int
}

func newPhase1Index(p1Rows []*result.Result) *phase1Index {
	idx := &phase1Index{
		byEndpoint: make(map[string]*result.Result, len(p1Rows)),
		rank:       make(map[string]int, len(p1Rows)),
	}
	sorted := make([]*result.Result, len(p1Rows))
	copy(sorted, p1Rows)
	result.Sort(sorted, result.SortByReliable)
	for i, r := range sorted {
		if r == nil {
			continue
		}
		key := endpointKey(r.IP.String(), r.Port)
		// First occurrence wins so a duplicate endpoint takes the better rank,
		// matching the order the Phase 1 table prints.
		if _, seen := idx.byEndpoint[key]; !seen {
			idx.byEndpoint[key] = r
			idx.rank[key] = i
		}
	}
	return idx
}

// get returns the Phase 1 row for ip:port, or nil when the endpoint was
// validated without appearing in Phase 1.
func (idx *phase1Index) get(ip string, port int) *result.Result {
	if idx == nil {
		return nil
	}
	return idx.byEndpoint[endpointKey(ip, port)]
}

// phase2InPhase1Order returns the Phase 2 rows ordered to match the Phase 1
// table's display order, so the two tables in the live file and the on-screen
// view line up row for row and a reader can compare an endpoint's Phase 1
// ranking against its validation result without cross-referencing.
//
// Phase 2 results arrive in worker-completion order — effectively random with
// respect to the Phase 1 ranking — so without this the second table gives no
// hint of why each IP was picked. Rows whose endpoint is absent from Phase 1
// keep their relative arrival order and are appended at the end.
func phase2InPhase1Order(p1Rows []*result.Result, p2Rows []*xraytest.ValidationResult) []*xraytest.ValidationResult {
	return newPhase1Index(p1Rows).order(p2Rows)
}

// order sorts a copy of p2Rows into Phase 1 rank order.
func (idx *phase1Index) order(p2Rows []*xraytest.ValidationResult) []*xraytest.ValidationResult {
	if len(p2Rows) == 0 {
		return nil
	}
	out := make([]*xraytest.ValidationResult, len(p2Rows))
	copy(out, p2Rows)
	// An endpoint missing from Phase 1 sorts after every ranked row but keeps its
	// own relative order, because sort.SliceStable preserves ties.
	unranked := len(idx.rank) + 1
	rankOf := func(v *xraytest.ValidationResult) int {
		if v == nil {
			return unranked
		}
		if i, ok := idx.rank[endpointKey(v.IP, v.Port)]; ok {
			return i
		}
		return unranked
	}
	sort.SliceStable(out, func(i, j int) bool { return rankOf(out[i]) < rankOf(out[j]) })
	return out
}

func (w *LiveResultWriter) writeLocked() error {

	if len(w.phase1Rows) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("SenPai Scanner — live results\n")
	fmt.Fprintf(&sb, "Started: %s\n", w.started.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&sb, "Updated: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	if w.withConfig {
		sb.WriteString("Plan: Phase 1 connectivity, then Phase 2 xray validation\n")
	} else {
		sb.WriteString("Plan: Phase 1 connectivity only\n")
	}
	sb.WriteString("\n")

	healthy := len(w.phase1Rows)
	fmt.Fprintf(&sb, "=== Phase 1 — connectivity (%d healthy / %d probed) ===\n\n", healthy, w.phase1Probed)
	rows := append([]*result.Result(nil), w.phase1Rows...)
	result.Sort(rows, result.SortByReliable)

	// The tail column used to carry a percentile name derived from the first
	// row's sample count, so a single header described rows that each had a
	// different number of successful probes. TrimmedAvg is one statistic with
	// one name, so the label is now a constant.
	fmt.Fprintf(&sb, "  %-22s  %7s  %9s  %8s  %8s  %8s  %6s\n",
		"ENDPOINT", "LOSS", "AVG(ms)", "TRM(ms)", "MAX(ms)", "COLO", "STATUS")
	sb.WriteString("  " + strings.Repeat("─", 82) + "\n")

	if len(rows) == 0 {
		sb.WriteString("  (no healthy results yet)\n")
	} else {
		for _, r := range rows {
			colo := r.Colo
			if colo == "" {
				colo = "—"
			}
			fmt.Fprintf(&sb, "  %-22s  %6.1f%%  %9.2f  %8.2f  %8.2f  %-8s  %s\n",
				formatEndpoint(r.IP.String(), r.Port),
				r.Loss(),
				float64(r.Avg().Milliseconds()),
				float64(r.TrimmedAvg().Milliseconds()),
				float64(r.Max().Milliseconds()),
				colo,
				"healthy",
			)
		}
	}

	if w.phase >= 2 && !w.phase1Only {
		sb.WriteString("\n")
		fmt.Fprintf(&sb, "=== Phase 2 — xray validation (%d tested) ===\n\n", len(w.phase2Rows))
		// P1 columns carry the Phase 1 ranking metrics for the same endpoint, so a
		// reader can see why an IP was tested and how it ranked before validation.
		fmt.Fprintf(&sb, "  %-22s  %-8s  %8s  %8s  %8s  %8s  %8s  %8s  %6s  %s\n",
			"ENDPOINT", "TYPE", "SPEED", "UPLOAD", "LATENCY", "P1-TRM", "P1-AVG", "P1-LOSS", "STATUS", "ERROR")
		sb.WriteString("  " + strings.Repeat("─", 124) + "\n")
		if len(w.phase2Rows) == 0 {
			sb.WriteString("  (no validation results yet)\n")
		} else {
			// Phase 2 rows arrive in worker-completion order, which is unrelated to
			// the Phase 1 ranking. Emit them in Phase 1 display order so the second
			// table reads as a continuation of the first.
			p1 := newPhase1Index(w.phase1Rows)
			for _, r := range p1.order(w.phase2Rows) {
				status := "fail"
				speed := "—"
				upload := "—"
				latency := "—"
				errText := "—"
				if r.Success {
					status = "ok"
					speed = formatValidationSpeed(r.Throughput)
					if r.UploadThroughput > 0 {
						upload = formatValidationSpeed(r.UploadThroughput)
					}
					latency = formatValidationLatency(r.Latency)
				} else if r.Error != "" {
					errText = r.Error
				}
				// Phase 1 metrics for this endpoint. An endpoint that reached Phase 2
				// normally has a row; the "—" keeps the table aligned if not.
				p1TRM, p1AVG, p1Loss := "—", "—", "—"
				if p := p1.get(r.IP, r.Port); p != nil {
					p1TRM = fmt.Sprintf("%.0f", float64(p.TrimmedAvg().Milliseconds()))
					p1AVG = fmt.Sprintf("%.0f", float64(p.Avg().Milliseconds()))
					p1Loss = fmt.Sprintf("%.1f%%", p.Loss())
				}
				fmt.Fprintf(&sb, "  %-22s  %-8s  %8s  %8s  %8s  %8s  %8s  %8s  %6s  %s\n",
					formatEndpoint(r.IP, r.Port),
					r.Transport,
					speed,
					upload,
					latency,
					p1TRM,
					p1AVG,
					p1Loss,
					status,
					errText,
				)
			}
		}
	}

	return os.WriteFile(w.path, []byte(sb.String()), 0644)
}
