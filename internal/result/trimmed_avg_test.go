package result

import (
	"testing"
	"time"
)

// lat builds a Result from successful latencies in milliseconds, ignoring
// zeros exactly the way the scanner treats a failed probe.
func lat(values ...int) *Result {
	out := make([]time.Duration, len(values))
	for i, v := range values {
		out[i] = ms(v)
	}
	return &Result{Latencies: out}
}

// TestTrimmedAvgDropsWorstSample pins the core promise of the column that
// replaced the percentile: the slowest samples are discarded before averaging,
// so a spike cannot dominate the reported figure.
//
// Ten samples, one 5000ms spike. The trim is a quarter of the samples, so two
// are dropped: the 5000ms spike and the 200ms one. The remaining eight are
// [100 120 130 140 150 160 170 180], summing to 1150ms.
//
// The expectation is exact nanoseconds, not a whole millisecond: dividing one
// time.Duration by another keeps the fractional part, so 1150ms/8 is 143.75ms
// and ms(143) would be wrong by 750 microseconds.
func TestTrimmedAvgDropsWorstSample(t *testing.T) {
	r := lat(100, 120, 130, 140, 150, 160, 170, 180, 200, 5000)
	if got, want := r.TrimmedAvg(), 143_750_000*time.Nanosecond; got != want {
		t.Errorf("TrimmedAvg() = %v, want %v (1150ms/8 = 143.75ms)", got, want)
	}
}

// TestTrimmedAvgBelowMaxAndBelowAvg brackets the statistic: it must sit below
// the full average, because dropping only the slowest samples can never raise a
// mean, and strictly below the maximum, because it is still an average rather
// than a rank lookup that collapses onto Max at these sample sizes.
func TestTrimmedAvgBelowMaxAndBelowAvg(t *testing.T) {
	for _, n := range []int{3, 4, 8, 10, 20} {
		r := &Result{Latencies: samples(n, 900)}
		got := r.TrimmedAvg()
		switch {
		case got >= r.Avg():
			t.Errorf("N=%d: TrimmedAvg() = %v, want below Avg %v", n, got, r.Avg())
		case got >= r.Max():
			t.Errorf("N=%d: TrimmedAvg() = %v, want strictly below Max %v", n, got, r.Max())
		}
	}
}

// TestTrimmedAvgSkipsFailedProbes pins that zeros from failed probes are not
// treated as the slowest samples. A row with 40% loss must trim on its
// successful probes only, otherwise the trim would silently do nothing.
func TestTrimmedAvgSkipsFailedProbes(t *testing.T) {
	// Eight successes with one spike, plus two failures.
	withFailures := &Result{Latencies: samples(8, 900)}
	withFailures.Latencies = append(withFailures.Latencies, 0, 0)
	clean := &Result{Latencies: samples(8, 900)}

	if got, want := withFailures.TrimmedAvg(), clean.TrimmedAvg(); got != want {
		t.Errorf("TrimmedAvg() = %v with two failed probes, want %v from eight successes alone", got, want)
	}
	if got := withFailures.Avg(); got != clean.Avg() {
		t.Errorf("Avg() = %v with two failed probes, want %v — zeros must not count", got, clean.Avg())
	}
}

// TestTrimmedAvgSmallSamples covers the degenerate counts so the trim can
// never index out of range or return a meaningless zero mid-scan.
func TestTrimmedAvgSmallSamples(t *testing.T) {
	if got := lat().TrimmedAvg(); got != 0 {
		t.Errorf("TrimmedAvg() with no samples = %v, want 0", got)
	}
	if got := lat(700).TrimmedAvg(); got != 0 {
		t.Errorf("TrimmedAvg() with one sample = %v, want 0", got)
	}
	// Two samples: trim is capped at len/3 = 0, so both survive and this equals
	// the plain average rather than dropping half the evidence.
	if got, want := lat(700, 900).TrimmedAvg(), ms(800); got != want {
		t.Errorf("TrimmedAvg() with two samples = %v, want %v", got, want)
	}
	// Three samples: at most one is dropped, leaving two.
	if got, want := lat(100, 200, 3000).TrimmedAvg(), ms(150); got != want {
		t.Errorf("TrimmedAvg() with three samples = %v, want %v", got, want)
	}
}

// TestTrimmedAvgMatchesObservedRow reproduces a row from a real ten-try run
// that made the percentile look wrong. The row was 104.19.50.70:443 with 10%
// loss, so nine successes out of ten probes, reporting AVG 698 and MAX 1385 —
// while the percentile column showed 1343, which is simply the second-worst of
// the nine samples wearing a percentile name it had not earned.
//
// The latencies below are reconstructed, not recorded: nine values summing to
// 6282 so the mean is 698, with 1385 as the outlier. The point of the test is
// how the statistic behaves on that shape, not the exact wire timings.
func TestTrimmedAvgMatchesObservedRow(t *testing.T) {
	r := &Result{Latencies: []time.Duration{
		0, ms(300), ms(420), ms(550), ms(600), ms(650), ms(700), ms(780), ms(897), ms(1385),
	}}

	successes := 0
	for _, l := range r.Latencies {
		if l > 0 {
			successes++
		}
	}
	if successes != 9 {
		t.Fatalf("successful probes = %d, want 9", successes)
	}
	if got, want := r.Avg(), ms(698); got != want {
		t.Fatalf("Avg() = %v, want %v — the fixture must reproduce the reported row", got, want)
	}

	trimmed := r.TrimmedAvg()
	if trimmed >= r.Max() {
		t.Errorf("TrimmedAvg() = %v, want strictly below Max %v", trimmed, r.Max())
	}
	if trimmed >= r.Avg() {
		t.Errorf("TrimmedAvg() = %v, want below Avg %v (the 1385ms outlier lifted the full average)", trimmed, r.Avg())
	}
	// Nine successes, so the quarter-trim drops the two slowest: 1385 and 897.
	// The remaining seven sum to 4000ms. Duration division keeps the fractional
	// part and truncates at nanosecond resolution: 4000ms/7 = 571.428571...ms.
	if got, want := trimmed, 571_428_571*time.Nanosecond; got != want {
		t.Errorf("TrimmedAvg() = %v, want %v for this row (4000ms/7)", got, want)
	}
	// The old percentile reported 1343 here. This statistic must not reproduce a
	// value that close to Max for a row whose body sits near 600ms.
	if trimmed > ms(700) {
		t.Errorf("TrimmedAvg() = %v, want at or below 700ms so the 1385ms outlier cannot dominate", trimmed)
	}
}