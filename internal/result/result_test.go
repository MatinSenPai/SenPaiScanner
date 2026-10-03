package result

import (
	"net"
	"testing"
	"time"
)

func makeResult(latencies []time.Duration) *Result {
	return &Result{
		IP:        net.ParseIP("1.1.1.1"),
		Port:      443,
		Latencies: latencies,
		TLSOk:     true,
	}
}

func TestLoss(t *testing.T) {
	r := makeResult([]time.Duration{100 * time.Millisecond, 0, 100 * time.Millisecond, 0})
	if got := r.Loss(); got != 50.0 {
		t.Errorf("Loss() = %v, want 50.0", got)
	}
}

func TestLossAllFailed(t *testing.T) {
	r := makeResult([]time.Duration{0, 0, 0})
	if r.Loss() != 100.0 {
		t.Errorf("expected 100%% loss, got %.1f", r.Loss())
	}
	if r.IsHealthy() {
		t.Error("expected unhealthy result")
	}
}

func TestAvg(t *testing.T) {
	r := makeResult([]time.Duration{100 * time.Millisecond, 200 * time.Millisecond, 0})
	avg := r.Avg()
	if avg != 150*time.Millisecond {
		t.Errorf("Avg() = %v, want 150ms", avg)
	}
}

func TestMinMax(t *testing.T) {
	r := makeResult([]time.Duration{50 * time.Millisecond, 200 * time.Millisecond, 80 * time.Millisecond})
	if r.Min() != 50*time.Millisecond {
		t.Errorf("Min() = %v, want 50ms", r.Min())
	}
	if r.Max() != 200*time.Millisecond {
		t.Errorf("Max() = %v, want 200ms", r.Max())
	}
}

func TestJitter(t *testing.T) {
	r := makeResult([]time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond})
	if r.Jitter() != 0 {
		t.Errorf("Jitter() with identical samples = %v, want 0", r.Jitter())
	}
}

func TestSort(t *testing.T) {
	results := []*Result{
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{200 * time.Millisecond}},
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{50 * time.Millisecond}},
		{IP: net.ParseIP("1.1.1.3"), Latencies: []time.Duration{100 * time.Millisecond}},
	}
	Sort(results, SortByAvg)
	if results[0].IP.String() != "1.1.1.2" {
		t.Errorf("first result after sort = %s, want 1.1.1.2", results[0].IP)
	}
}

func TestTopN(t *testing.T) {
	results := []*Result{
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{200 * time.Millisecond, 210 * time.Millisecond}, TLSOk: true},
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{50 * time.Millisecond, 55 * time.Millisecond}, TLSOk: true},
		{IP: net.ParseIP("1.1.1.3"), Latencies: []time.Duration{0, 0}}, // unhealthy
		{IP: net.ParseIP("1.1.1.4"), Latencies: []time.Duration{100 * time.Millisecond, 105 * time.Millisecond}, TLSOk: true},
	}
	top := TopN(results, 2)
	if len(top) != 2 {
		t.Errorf("TopN(2) returned %d results, want 2", len(top))
	}
	if top[0].IP.String() != "1.1.1.2" {
		t.Errorf("best result = %s, want 1.1.1.2", top[0].IP)
	}
}

func TestHTTPHealthRequiresCloudflareValidation(t *testing.T) {
	r := makeResult([]time.Duration{100 * time.Millisecond, 120 * time.Millisecond})
	r.ProbeMode = "http"
	r.HTTPStatus = 200
	r.TLSOk = true

	if r.IsHealthy() {
		t.Fatal("expected HTTP result without colo to be unhealthy")
	}

	r.Colo = "FRA"
	if !r.IsHealthy() {
		t.Fatal("expected validated HTTP result to be healthy")
	}

	// Speed tested but 0 throughput = unhealthy (DPI allows trace but blocks data)
	r.SpeedTested = true
	if r.IsHealthy() {
		t.Fatal("expected speed-tested result with 0 throughput to be unhealthy")
	}

	r.Throughput = 256 * 1024
	if !r.IsHealthy() {
		t.Fatal("expected speed-tested result with throughput to remain healthy")
	}
}

func TestHTTPHealthRequiresTLSOnNonPlainHTTPPorts(t *testing.T) {
	r := makeResult([]time.Duration{100 * time.Millisecond, 110 * time.Millisecond})
	r.ProbeMode = "http"
	r.Port = 2087
	r.TLSOk = false
	r.HTTPStatus = 200
	r.Colo = "FRA"

	if r.IsHealthy() {
		t.Fatal("expected HTTPS-style port without TLS to be unhealthy")
	}
}

func TestHTTPHealthRequiresWebSocketWhenConfigured(t *testing.T) {
	r := makeResult([]time.Duration{100 * time.Millisecond, 105 * time.Millisecond})
	r.ProbeMode = "http"
	r.HTTPStatus = 200
	r.Colo = "FRA"
	r.RequireWS = true

	if r.IsHealthy() {
		t.Fatal("expected required websocket failure to be unhealthy")
	}

	r.WSOk = true
	if !r.IsHealthy() {
		t.Fatal("expected required websocket success to be healthy")
	}
}

func TestHTTPTimeoutIsNotHealthy(t *testing.T) {
	// Simulates the bug: all tries time out (latency 0) or previously recorded 3s.
	r := &Result{
		IP:          net.ParseIP("1.1.1.1"),
		ProbeMode:   "http",
		Latencies:   []time.Duration{0, 0, 0, 0},
		SpeedTested: true,
	}
	if r.IsHealthy() {
		t.Fatal("expected all-failed HTTP probe to be unhealthy")
	}
}

func TestTLSRequiresHandshake(t *testing.T) {
	r := makeResult([]time.Duration{100 * time.Millisecond, 110 * time.Millisecond})
	r.ProbeMode = "tls"
	r.TLSOk = false
	if r.IsHealthy() {
		t.Fatal("expected TLS result without handshake to be unhealthy")
	}
	r.TLSOk = true
	if !r.IsHealthy() {
		t.Fatal("expected TLS handshake success to be healthy")
	}
}

func TestSortBySpeed(t *testing.T) {
	results := []*Result{
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{80 * time.Millisecond}, Throughput: 100 * 1024},
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{90 * time.Millisecond}, Throughput: 900 * 1024},
	}

	Sort(results, SortBySpeed)
	if results[0].IP.String() != "1.1.1.2" {
		t.Errorf("first result after speed sort = %s, want 1.1.1.2", results[0].IP)
	}
}

func TestSortByJitterPrefersHealthyResults(t *testing.T) {
	results := []*Result{
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{0, 0, 0, 0}},
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{90 * time.Millisecond, 100 * time.Millisecond, 110 * time.Millisecond}, TLSOk: true},
	}

	Sort(results, SortByJitter)
	if results[0].IP.String() != "1.1.1.2" {
		t.Fatalf("first result after jitter sort = %s, want 1.1.1.2", results[0].IP)
	}
}

func TestSortByLossPrefersPartialOverTotalFailure(t *testing.T) {
	results := []*Result{
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{0, 0, 0, 0}},
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{100 * time.Millisecond, 0, 120 * time.Millisecond, 0}, TLSOk: true},
	}

	Sort(results, SortByLoss)
	if results[0].IP.String() != "1.1.1.2" {
		t.Fatalf("first result after loss sort = %s, want 1.1.1.2", results[0].IP)
	}
}

func TestSortBySpeedKeepsHealthyResultsAheadOfUntestedFailures(t *testing.T) {
	results := []*Result{
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{0, 0, 0}, Throughput: 0},
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{80 * time.Millisecond, 85 * time.Millisecond}, Throughput: 100 * 1024, TLSOk: true},
	}

	Sort(results, SortBySpeed)
	if results[0].IP.String() != "1.1.1.2" {
		t.Fatalf("first result after speed sort = %s, want 1.1.1.2", results[0].IP)
	}
}

// A one-try download probe has a single latency, so IsHealthy is always false; DownloadOK must still pass it.
func TestDownloadOKForSingleTryProbe(t *testing.T) {
	r := &Result{Latencies: []time.Duration{120 * time.Millisecond}, ProbeMode: "http", HTTPStatus: 200, Throughput: 250_000}
	if r.IsHealthy() {
		t.Fatal("precondition: one try can never be healthy")
	}
	if !r.DownloadOK() {
		t.Error("a successful download sample must count")
	}
	r.Throughput = 0
	if r.DownloadOK() {
		t.Error("no throughput means the sample failed")
	}
	r.Throughput, r.HTTPStatus = 250_000, 503
	if r.DownloadOK() {
		t.Error("an error status means the sample failed")
	}
}
