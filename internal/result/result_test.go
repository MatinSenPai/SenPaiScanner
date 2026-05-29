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

func TestMedian(t *testing.T) {
	// Odd count.
	r := makeResult([]time.Duration{30 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond})
	if got := r.Median(); got != 20*time.Millisecond {
		t.Fatalf("Median odd = %v, want 20ms", got)
	}
	// Even count averages the middle two.
	r = makeResult([]time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 30 * time.Millisecond, 40 * time.Millisecond})
	if got := r.Median(); got != 25*time.Millisecond {
		t.Fatalf("Median even = %v, want 25ms", got)
	}
	// Failed tries (0) are ignored.
	r = makeResult([]time.Duration{0, 10 * time.Millisecond, 0, 30 * time.Millisecond})
	if got := r.Median(); got != 20*time.Millisecond {
		t.Fatalf("Median with failures = %v, want 20ms", got)
	}
}

func TestScoreUsesRobustLatency(t *testing.T) {
	// Three fast tries plus one slow outlier. A mean-based score would be
	// dragged down by the 500ms try; a median-based score should stay high.
	r := makeResult([]time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond, 500 * time.Millisecond})
	r.ProbeMode = "tcp"
	if r.Median() >= r.Avg() {
		t.Fatalf("expected median (%v) < avg (%v) for this sample", r.Median(), r.Avg())
	}
	// Median ~10ms => latScore ~0.98; the score must reflect the fast typical
	// latency rather than the outlier-inflated mean.
	if got := r.Score(); got < 80 {
		t.Fatalf("Score = %.1f, want >=80 (median-based, outlier ignored)", got)
	}
}

func TestDominantError(t *testing.T) {
	// Mixed failures, no success -> most common (timeout) wins.
	r := &Result{Errors: []ProbeError{ErrTimeout, ErrTimeout, ErrRefused}}
	if got := r.DominantError(); got != ErrTimeout {
		t.Fatalf("DominantError = %v, want timeout", got)
	}
	// Any success -> ErrNone, regardless of other failed tries.
	r = &Result{Errors: []ProbeError{ErrTimeout, ErrNone, ErrTimeout}}
	if got := r.DominantError(); got != ErrNone {
		t.Fatalf("DominantError with a success = %v, want ok", got)
	}
	// Tie breaks toward the more severe (lower-numbered) class.
	r = &Result{Errors: []ProbeError{ErrRefused, ErrTLS}}
	if got := r.DominantError(); got != ErrRefused {
		t.Fatalf("DominantError tie = %v, want refused", got)
	}
	// No data -> ErrNone.
	if got := (&Result{}).DominantError(); got != ErrNone {
		t.Fatalf("DominantError empty = %v, want ok", got)
	}
}

func TestProbeErrorString(t *testing.T) {
	if ErrTimeout.String() != "timeout" || ErrNone.String() != "ok" {
		t.Fatalf("unexpected ProbeError labels")
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
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{200 * time.Millisecond}, TLSOk: true},
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{50 * time.Millisecond}, TLSOk: true},
		{IP: net.ParseIP("1.1.1.3"), Latencies: []time.Duration{0}}, // unhealthy
		{IP: net.ParseIP("1.1.1.4"), Latencies: []time.Duration{100 * time.Millisecond}, TLSOk: true},
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

	r.SpeedTested = true
	if r.IsHealthy() {
		t.Fatal("expected speed-tested result with zero throughput to be unhealthy")
	}

	r.Throughput = 256 * 1024
	if !r.IsHealthy() {
		t.Fatal("expected speed-tested result with throughput to be healthy")
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
	r := makeResult([]time.Duration{100 * time.Millisecond})
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

// A fully-failed result has Avg()==0; it must sort to the end, never the front,
// so dead IPs don't masquerade as the fastest.
func TestSortByAvgZeroLast(t *testing.T) {
	results := []*Result{
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{0, 0}}, // dead
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{120 * time.Millisecond}},
		{IP: net.ParseIP("1.1.1.3"), Latencies: []time.Duration{40 * time.Millisecond}},
	}
	Sort(results, SortByAvg)
	if results[0].IP.String() != "1.1.1.3" {
		t.Errorf("first = %s, want 1.1.1.3", results[0].IP)
	}
	if results[len(results)-1].IP.String() != "1.1.1.1" {
		t.Errorf("last = %s, want dead IP 1.1.1.1", results[len(results)-1].IP)
	}
}

func TestScoreUnhealthyIsZero(t *testing.T) {
	r := makeResult([]time.Duration{0, 0}) // 100% loss → unhealthy
	if s := r.Score(); s != 0 {
		t.Errorf("Score() of unhealthy result = %v, want 0", s)
	}
}

func TestScoreRewardsLowerLatency(t *testing.T) {
	fast := makeResult([]time.Duration{20 * time.Millisecond, 20 * time.Millisecond})
	slow := makeResult([]time.Duration{300 * time.Millisecond, 300 * time.Millisecond})
	if fast.Score() <= slow.Score() {
		t.Errorf("fast score %.1f should exceed slow score %.1f", fast.Score(), slow.Score())
	}
}

func TestScoreRewardsThroughput(t *testing.T) {
	a := makeResult([]time.Duration{50 * time.Millisecond, 50 * time.Millisecond})
	a.ProbeMode, a.HTTPStatus, a.Colo, a.SpeedTested, a.Throughput = "http", 200, "FRA", true, 8*1024*1024
	b := makeResult([]time.Duration{50 * time.Millisecond, 50 * time.Millisecond})
	b.ProbeMode, b.HTTPStatus, b.Colo, b.SpeedTested, b.Throughput = "http", 200, "FRA", true, 256*1024
	if a.Score() <= b.Score() {
		t.Errorf("higher-throughput score %.1f should exceed lower %.1f", a.Score(), b.Score())
	}
}

func TestSortByScore(t *testing.T) {
	results := []*Result{
		{IP: net.ParseIP("1.1.1.1"), Latencies: []time.Duration{300 * time.Millisecond}, TLSOk: true},
		{IP: net.ParseIP("1.1.1.2"), Latencies: []time.Duration{20 * time.Millisecond}, TLSOk: true},
		{IP: net.ParseIP("1.1.1.3"), Latencies: []time.Duration{0, 0}}, // unhealthy → score 0
	}
	Sort(results, SortByScore)
	if results[0].IP.String() != "1.1.1.2" {
		t.Errorf("best by score = %s, want 1.1.1.2", results[0].IP)
	}
	if results[len(results)-1].IP.String() != "1.1.1.3" {
		t.Errorf("worst by score = %s, want unhealthy 1.1.1.3", results[len(results)-1].IP)
	}
}

func BenchmarkSort(b *testing.B) {
	base := make([]*Result, 1000)
	for i := range base {
		base[i] = &Result{
			IP:    net.ParseIP("1.1.1.1"),
			TLSOk: true,
			Latencies: []time.Duration{
				time.Duration(i%200) * time.Millisecond,
				time.Duration((i*7)%200) * time.Millisecond,
				time.Duration((i*13)%200) * time.Millisecond,
			},
		}
	}
	work := make([]*Result, len(base))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(work, base)
		Sort(work, SortByScore)
	}
}

func TestMs(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want float64
	}{
		{0, 0},
		{time.Millisecond, 1},
		{1500 * time.Microsecond, 1.5}, // sub-ms precision Duration.Milliseconds() would lose
		{time.Second, 1000},
	}
	for _, c := range cases {
		if got := Ms(c.in); got != c.want {
			t.Errorf("Ms(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
