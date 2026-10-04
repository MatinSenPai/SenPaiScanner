package mobile

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

type fakeCB struct {
	progress []string
	results  []string
	infos    []string
	errs     []string
}

func (f *fakeCB) OnProgress(tested, healthy, failed, inFlight int, isPhase2 bool) {
	f.progress = append(f.progress, strings.Join([]string{itoa(tested), itoa(healthy), itoa(failed), itoa(inFlight)}, "/"))
}
func (f *fakeCB) OnResult(ip string, port int, latencyMs int, loss float64, colo string, isHealthy bool, isPhase2 bool, phase2Type string, speed float64, status bool, upload float64) {
	f.results = append(f.results, ip+"|"+colo+"|"+phase2Type)
}
func (f *fakeCB) OnFinished()        {}
func (f *fakeCB) OnError(err string) { f.errs = append(f.errs, err) }
func (f *fakeCB) OnInfo(msg string)  { f.infos = append(f.infos, msg) }

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

func TestJobMapsTheForm(t *testing.T) {
	cfg := ScanConfig{SourceType: "Paste", Targets: "1.1.1.1 2.2.2.2", CountType: "Custom", CustomCount: "123", WorkerType: "100 workers",
		TimeoutType: "3s", SelectedPorts: []int{443, 8443}, TopNType: "ALL", Gentle: true, Phase2Only: true, NeighborScan: true}
	p, err := cfg.job()
	if err != nil {
		t.Fatal(err)
	}
	if p.Count != 123 || p.Workers != 100 || p.TimeoutMs != 3000 || p.TopN != 0 || !p.Gentle || !p.Phase2Only {
		t.Errorf("mapping wrong: %+v", p)
	}
	if len(p.Ports) != 2 || p.Targets == "" || !p.NoLiveFile || p.Phase2Workers != phase2Workers || p.Validate == nil {
		t.Errorf("android specifics missing: %+v", p)
	}

	cfg.ConfigURL = "vless://u@h.example:2053?security=tls&type=ws#x"
	p, _ = cfg.job()
	if len(p.Ports) != 0 || p.NeighborScan {
		t.Errorf("with a config the config's port is used and neighbors stay off: %+v", p)
	}

	cfg = ScanConfig{SourceType: "Random", Targets: "ignored", Phase2Only: true}
	if p, _ = cfg.job(); p.Targets != "" || p.Phase2Only {
		t.Errorf("Phase2Only needs pasted targets, got %+v", p)
	}
}

func TestEventsAdapter(t *testing.T) {
	cb := &fakeCB{}
	e := &events{cb: cb, colo: map[string]string{}}

	e.Error("bad url") // before anything started: ends the scan
	e.Phase(1, "")
	e.Stats(4, 1, 10)
	e.Results([]*result.Result{
		{IP: net.ParseIP("1.2.3.4"), Port: 443, ProbeMode: "http", TLSOk: true, HTTPStatus: 200, Colo: "FRA", Latencies: []time.Duration{time.Second, time.Second}},
		{IP: net.ParseIP("5.6.7.8"), Port: 443}, // unhealthy: not reported
	})
	e.Phase(2, "")
	e.Validate(&xraytest.ValidationResult{IP: "1.2.3.4", Port: 443, Transport: "ws", Success: true}, 1, 1)
	e.Error("late problem") // after start: a notice, the scan goes on

	if len(cb.errs) != 1 || cb.errs[0] != "bad url" {
		t.Errorf("only the early error may end the scan: %v", cb.errs)
	}
	if len(cb.infos) != 1 || !strings.Contains(cb.infos[0], "late problem") {
		t.Errorf("a late error must be a notice: %v", cb.infos)
	}
	if len(cb.results) != 2 || cb.results[0] != "1.2.3.4|FRA|" || cb.results[1] != "1.2.3.4|FRA|ws" {
		t.Errorf("results wrong (phase 2 keeps the colo from phase 1): %v", cb.results)
	}
	if cb.progress[0] != "4/1/3/6" {
		t.Errorf("progress = %v", cb.progress)
	}
}

func TestPreviewAndAntiDpiHelpers(t *testing.T) {
	var prev struct {
		Count   int      `json:"count"`
		Skipped []string `json:"skipped"`
	}
	if err := json.Unmarshal([]byte(PreviewTargets("8.8.8.0/30 nope!")), &prev); err != nil || prev.Count != 4 || len(prev.Skipped) != 1 {
		t.Errorf("PreviewTargets = %+v err=%v", prev, err)
	}
	if ValidateAntiDpi(`{"antiDpiEnabled":true}`) != "" {
		t.Error("empty fields fall back to the published values and must validate")
	}
	if ValidateAntiDpi(`{"antiDpiEnabled":true,"adFinalmask":"{nope"}`) == "" {
		t.Error("broken finalmask JSON must be reported")
	}
	if !strings.Contains(AntiDpiDefaults(), "tlshello") {
		t.Error("defaults must contain the published recipe")
	}
	if ResumeInfo() != "" {
		t.Error("no data dir, nothing to resume")
	}
}
