package scanjob

import (
	"context"
	"net"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

func ipset(t Targets) map[string]bool {
	m := map[string]bool{}
	for _, ip := range t.IPs {
		m[ip.String()] = true
	}
	return m
}

func TestParseTargets(t *testing.T) {
	got := ParseTargets(context.Background(), `
# my list
104.18.1.1, 104.18.1.2:443  [2606:4700::1]:2053
1.2.3.4-1.2.3.6 ; 5.5.5.5-7
8.8.8.0/30
104.18.1.1
not a thing!
10.0.0.0/8x
`)
	have := ipset(got)
	for _, want := range []string{"104.18.1.1", "104.18.1.2", "2606:4700::1", "1.2.3.4", "1.2.3.5", "1.2.3.6", "5.5.5.5", "5.5.5.6", "5.5.5.7", "8.8.8.0", "8.8.8.3"} {
		if !have[want] {
			t.Errorf("missing %s in %v", want, have)
		}
	}
	if len(got.IPs) != 13 {
		t.Errorf("want 13 unique IPs (104.18.1.1 appears twice), got %d", len(got.IPs))
	}
	if len(got.Skipped) < 2 {
		t.Errorf("garbage entries must be reported, got %v", got.Skipped)
	}
}

func TestParseTargetsBigCIDRIsSampled(t *testing.T) {
	got := ParseTargets(context.Background(), "104.16.0.0/13")
	if len(got.IPs) != cidrSample {
		t.Errorf("a /13 must be sampled to %d addresses, got %d", cidrSample, len(got.IPs))
	}
	for _, ip := range got.IPs {
		if _, n, _ := net.ParseCIDR("104.16.0.0/13"); !n.Contains(ip) {
			t.Fatalf("%s is outside the CIDR", ip)
		}
	}
}

func TestParseTargetsRangeLimitAndDomains(t *testing.T) {
	lookupIP = func(_ context.Context, host string) ([]net.IPAddr, error) {
		if host == "edge.example.com" {
			return []net.IPAddr{{IP: net.ParseIP("104.20.1.1")}, {IP: net.ParseIP("104.20.1.2")}}, nil
		}
		return nil, &net.DNSError{Err: "no such host", Name: host}
	}
	t.Cleanup(func() { lookupIP = net.DefaultResolver.LookupIPAddr })

	got := ParseTargets(context.Background(), "1.0.0.0-1.0.255.255 Edge.Example.com:2053 gone.example.org noDots")
	if len(got.Skipped) != 3 || !strings.Contains(got.Skipped[0], "more than") {
		t.Errorf("a huge range, an unresolvable and a dotless name must be reported, got %v", got.Skipped)
	}
	if have := ipset(got); got.Domains != 1 || !have["104.20.1.1"] || !have["104.20.1.2"] {
		t.Errorf("the domain must resolve to its addresses, got %v domains=%d", got.IPs, got.Domains)
	}
}

func TestSnapshotRoundTripAndResumeSkipsDoneWork(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	healthy := &result.Result{IP: net.ParseIP("104.18.1.1"), Port: 443, ProbeMode: "http", TLSOk: true, HTTPStatus: 200, Colo: "FRA",
		Latencies: []time.Duration{100 * time.Millisecond, 120 * time.Millisecond}}
	r := &run{p: Params{StatePath: path, Workers: 7}, ports: []int{443}, phase: 1, done: map[string]bool{},
		pool: []net.IP{net.ParseIP("104.18.1.1"), net.ParseIP("104.18.1.2"), net.ParseIP("104.18.1.3")}}
	r.done["104.18.1.1:443"], r.done["104.18.1.2:443"] = true, true
	r.probed, r.healthy = 2, []*result.Result{healthy}
	r.save()

	info, ok := InspectSnapshot(path)
	if !ok || info.Tested != 2 || info.Total != 3 || info.Healthy != 1 || info.Phase != 1 {
		t.Fatalf("InspectSnapshot = %+v ok=%v", info, ok)
	}
	s, err := LoadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Healthy[0].result(); !got.IsHealthy() || got.Colo != "FRA" || got.Avg() != 110*time.Millisecond {
		t.Fatalf("healthy row did not survive the round trip: %+v", got)
	}
	if s.Params.Workers != 7 {
		t.Errorf("settings were not saved: %+v", s.Params)
	}

	// the resumed run only probes the endpoint that is still missing
	r2 := &run{ports: []int{443}, done: map[string]bool{}}
	for _, p := range s.Pool {
		r2.pool = append(r2.pool, net.ParseIP(p))
	}
	for _, k := range s.Done {
		r2.done[k] = true
	}
	var left []string
	for ip := range r2.remaining(context.Background()) {
		left = append(left, ip.String())
	}
	if len(left) != 1 || left[0] != "104.18.1.3" {
		t.Errorf("resume must only probe the untested IP, got %v", left)
	}
}

func TestFinishKeepsSnapshotOnlyWhenInterrupted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	r := &run{p: Params{StatePath: path}, ports: []int{443}, phase: 1, done: map[string]bool{}, pool: []net.IP{net.ParseIP("1.1.1.1")}}
	r.finish(true)
	if _, ok := InspectSnapshot(path); !ok {
		t.Fatal("an interrupted scan must leave a snapshot to resume")
	}
	r.finish(false)
	if _, ok := InspectSnapshot(path); ok {
		t.Fatal("a completed scan must remove its snapshot")
	}
}

func TestGentleCapsWorkersAndTimeout(t *testing.T) {
	w, to, rate := Params{Workers: 200, TimeoutMs: 2000, Gentle: true}.timing()
	if w != GentleWorkers || to != GentleTimeout || rate != GentleRate {
		t.Errorf("gentle timing = %d %v %v", w, to, rate)
	}
	w, to, rate = Params{Workers: 200, TimeoutMs: 2000}.timing()
	if w != 200 || to != 2*time.Second || rate != 0 {
		t.Errorf("normal timing must be untouched, got %d %v %v", w, to, rate)
	}
	_, to, _ = Params{TimeoutMs: 9000, Gentle: true}.timing()
	if to != 9*time.Second {
		t.Errorf("gentle must never shorten a longer timeout, got %v", to)
	}
}

type recEvents struct {
	mu      sync.Mutex
	results []*result.Result
	infos   []string
}

func (r *recEvents) Phase(int, string)                             {}
func (r *recEvents) Stats(int, int, int)                           {}
func (r *recEvents) Validate(*xraytest.ValidationResult, int, int) {}
func (r *recEvents) Error(string)                                  {}
func (r *recEvents) Info(m string)                                 { r.mu.Lock(); r.infos = append(r.infos, m); r.mu.Unlock() }
func (r *recEvents) Results(b []*result.Result) {
	r.mu.Lock()
	r.results = append(r.results, b...)
	r.mu.Unlock()
}

// Resuming must show the healthy rows the earlier session found, not only count them.
func TestResumeReplaysEarlierHealthyRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	healthy := &result.Result{IP: net.ParseIP("104.18.1.1"), Port: 443, ProbeMode: "http", TLSOk: true, HTTPStatus: 200, Colo: "FRA",
		Latencies: []time.Duration{100 * time.Millisecond, 120 * time.Millisecond}}
	r := &run{p: Params{StatePath: path, NoLiveFile: true}, ports: []int{443}, phase: 1, done: map[string]bool{"104.18.1.1:443": true},
		pool: []net.IP{net.ParseIP("104.18.1.1")}, probed: 1, healthy: []*result.Result{healthy}}
	r.save()

	ev := &recEvents{}
	out := Run(context.Background(), Params{StatePath: path, Resume: true, NoLiveFile: true}, ev)
	if len(ev.results) != 1 || ev.results[0].IP.String() != "104.18.1.1" || ev.results[0].Colo != "FRA" {
		t.Fatalf("earlier healthy row was not replayed: %+v", ev.results)
	}
	if out.Healthy != 1 || out.Cancelled {
		t.Errorf("outcome = %+v", out)
	}
	if _, still := InspectSnapshot(path); still {
		t.Error("a finished resume must remove the snapshot")
	}
}
