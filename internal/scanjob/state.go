package scanjob

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

const snapshotVersion = 1

// Snapshot is everything needed to continue an interrupted scan in another session: the settings,
// the target pool, which endpoints were already probed, the healthy results, and (once Phase 2 has
// started) the validations that already finished.
type Snapshot struct {
	Version   int       `json:"version"`
	Saved     time.Time `json:"saved"`
	Params    Params    `json:"params"`
	Pool      []string  `json:"pool"`
	Done      []string  `json:"done"` // "ip:port" already probed in Phase 1
	Probed    int       `json:"probed"`
	Healthy   []Row     `json:"healthy"`
	Phase     int       `json:"phase"` // 1 or 2
	Validated []ValRow  `json:"validated,omitempty"`
}

// Row is a Phase 1 result in a form that survives JSON.
type Row struct {
	IP          string    `json:"ip"`
	Port        int       `json:"port"`
	Mode        string    `json:"mode"`
	LatenciesNs []int64   `json:"latenciesNs"`
	TLSOk       bool      `json:"tlsOk"`
	WSOk        bool      `json:"wsOk"`
	RequireWS   bool      `json:"requireWs"`
	HTTPStatus  int       `json:"httpStatus"`
	Colo        string    `json:"colo"`
	Throughput  float64   `json:"throughput"`
	SpeedTested bool      `json:"speedTested"`
	Timestamp   time.Time `json:"timestamp"`
}

// ValRow is a finished Phase 2 validation.
type ValRow struct {
	IP         string  `json:"ip"`
	Port       int     `json:"port"`
	Transport  string  `json:"transport"`
	Success    bool    `json:"success"`
	LatencyNs  int64   `json:"latencyNs"`
	Throughput float64 `json:"throughput"`
	Upload     float64 `json:"upload"`
	Error      string  `json:"error"`
}

func rowOf(r *result.Result) Row {
	ns := make([]int64, len(r.Latencies))
	for i, l := range r.Latencies {
		ns[i] = int64(l)
	}
	return Row{IP: r.IP.String(), Port: r.Port, Mode: r.ProbeMode, LatenciesNs: ns, TLSOk: r.TLSOk, WSOk: r.WSOk,
		RequireWS: r.RequireWS, HTTPStatus: r.HTTPStatus, Colo: r.Colo, Throughput: r.Throughput,
		SpeedTested: r.SpeedTested, Timestamp: r.Timestamp}
}

func (w Row) result() *result.Result {
	lat := make([]time.Duration, len(w.LatenciesNs))
	for i, n := range w.LatenciesNs {
		lat[i] = time.Duration(n)
	}
	return &result.Result{IP: net.ParseIP(w.IP), Port: w.Port, ProbeMode: w.Mode, Latencies: lat, TLSOk: w.TLSOk, WSOk: w.WSOk,
		RequireWS: w.RequireWS, HTTPStatus: w.HTTPStatus, Colo: w.Colo, Throughput: w.Throughput,
		SpeedTested: w.SpeedTested, Timestamp: w.Timestamp}
}

func valRowOf(v *xraytest.ValidationResult) ValRow {
	return ValRow{IP: v.IP, Port: v.Port, Transport: v.Transport, Success: v.Success, LatencyNs: int64(v.Latency),
		Throughput: v.Throughput, Upload: v.UploadThroughput, Error: v.Error}
}

func (v ValRow) result() *xraytest.ValidationResult {
	return &xraytest.ValidationResult{IP: v.IP, Port: v.Port, Transport: v.Transport, Success: v.Success,
		Latency: time.Duration(v.LatencyNs), Throughput: v.Throughput, UploadThroughput: v.Upload, Error: v.Error}
}

// DefaultStatePath is where the GUI and the CLI keep the scan to resume.
func DefaultStatePath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return "senpaiscanner-scan-state.json"
	}
	return filepath.Join(dir, "senpaiscanner", "scan-state.json")
}

// SaveSnapshot writes s atomically (temp file + rename) so a crash or power cut never leaves a half-written file.
func SaveSnapshot(path string, s *Snapshot) error {
	s.Version, s.Saved = snapshotVersion, time.Now()
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// LoadSnapshot reads a saved scan.
func LoadSnapshot(path string) (*Snapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("saved scan %s is damaged: %w", path, err)
	}
	if s.Version != snapshotVersion {
		return nil, fmt.Errorf("saved scan %s was written by another version", path)
	}
	return &s, nil
}

// ResumeInfo summarises a saved scan for the "Resume" button; ok is false when there is nothing to resume.
type ResumeInfo struct {
	Saved   time.Time `json:"saved"`
	Tested  int       `json:"tested"`
	Total   int       `json:"total"`
	Healthy int       `json:"healthy"`
	Phase   int       `json:"phase"`
}

func InspectSnapshot(path string) (ResumeInfo, bool) {
	s, err := LoadSnapshot(path)
	if err != nil {
		return ResumeInfo{}, false
	}
	return ResumeInfo{Saved: s.Saved, Tested: len(s.Done), Total: len(s.Pool) * max(1, len(s.Params.Ports)), Healthy: len(s.Healthy), Phase: s.Phase}, true
}

func removeFile(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DiscardSnapshot deletes a saved scan.
func DiscardSnapshot(path string) { _ = removeFile(path) }
