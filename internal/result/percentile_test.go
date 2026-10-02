package result

import (
	"net"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/config"
)

// ms is a small readability helper for building latency samples.
func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

// samples builds n successful latencies in 10ms steps, then replaces the
// slowest with spike so the tail is distinguishable from the body.
func samples(n int, spike int) []time.Duration {
	out := make([]time.Duration, n)
	for i := range out {
		out[i] = ms((i + 1) * 10)
	}
	out[n-1] = ms(spike)
	return out
}

// TestPercentileDiscardsWorstSample pins the rule that replaces nearest-rank:
// the percentile always discards at least one sample, so the tail column
// reports something other than the maximum at every usable sample count.
// Nearest-rank discarded floor(0.05*N) samples, which is zero below 20 tries
// and made P95 an exact copy of Max at the then-default of 10.
func TestPercentileDiscardsWorstSample(t *testing.T) {
	for _, n := range []int{4, 8, 10, 15, 20} {
		r := &Result{Latencies: samples(n, 900)}
		if got, max := r.P95(), r.Max(); got >= max {
			t.Errorf("N=%d: P95() = %v, want strictly below Max %v", n, got, max)
		}
	}
}

// TestPercentileAtTwentyTriesUnchanged pins the high-sample behaviour: at 20
// tries the discarded count is exactly one and P95 is the 19th of 20, which
// is the same rank nearest-rank reported. The fix must not shift the
// well-defined case, only rescue the small ones.
func TestPercentileAtTwentyTriesUnchanged(t *testing.T) {
	r := &Result{Latencies: samples(20, 900)}
	if got, want := r.P95(), ms(190); got != want {
		t.Errorf("P95() = %v, want the 19th of 20 (%v)", got, want)
	}
}

// TestPercentileSurvivesPacketLoss covers the case that matters most: an IP
// with heavy loss has few successful samples, and that must not push the
// tail back onto the maximum.
func TestPercentileSurvivesPacketLoss(t *testing.T) {
	// 20 tries, 8 of which failed: 12 usable samples with one spike.
	lat := append(samples(12, 900), make([]time.Duration, 8)...)
	r := &Result{Latencies: lat}

	if got, want := r.Loss(), 40.0; got != want {
		t.Errorf("Loss() = %v, want 40", got)
	}
	if got, max := r.P95(), r.Max(); got >= max {
		t.Errorf("P95() = %v, want strictly below Max %v under 40%% loss", got, max)
	}
}

// TestP90MatchesP95UntilTheyDiffer pins a property of the discard rule: both
// percentiles discard the same single sample until 0.10*N and 0.05*N round to
// different counts, so P90 and P95 are identical below 20 tries. This is
// expected, not a bug -- it is why only one tail column is shown in the UI.
func TestP90MatchesP95UntilTheyDiffer(t *testing.T) {
	for _, n := range []int{4, 8, 10, 15} {
		r := &Result{Latencies: samples(n, 900)}
		if r.P90() != r.P95() {
			t.Errorf("N=%d: P90 (%v) and P95 (%v) should be identical below 20 tries",
				n, r.P90(), r.P95())
		}
	}
	r := &Result{Latencies: samples(20, 900)}
	if r.P90() >= r.P95() {
		t.Errorf("at N=20 P90 (%v) should rank below P95 (%v)", r.P90(), r.P95())
	}
}

// TestPercentileLabel pins the honesty requirement: the UI prints whatever
// label this returns, so it must not claim P95 for a value that is really the
// 87.5th percentile.
func TestPercentileLabel(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{n: 8, want: "P87.5"},
		{n: 10, want: "P90"},
		{n: 20, want: "P95"},
		{n: 1, want: ""},
	}
	for _, tc := range tests {
		lat := make([]time.Duration, tc.n)
		for i := range lat {
			lat[i] = ms((i + 1) * 10)
		}
		r := &Result{Latencies: lat}
		if got := r.PercentileLabel(); got != tc.want {
			t.Errorf("N=%d: PercentileLabel() = %q, want %q", tc.n, got, tc.want)
		}
	}
}

// TestDefaultTriesIsEight pins the default sample count and the reason for
// it: 8 is the smallest count whose tail percentile still differs from the
// maximum while resolving loss to 12.5% steps.
func TestDefaultTriesIsEight(t *testing.T) {
	if config.ScanDefaults.Tries != 8 {
		t.Fatalf("ScanDefaults.Tries = %d, want 8", config.ScanDefaults.Tries)
	}
	r := &Result{Latencies: samples(config.ScanDefaults.Tries, 900)}
	if r.P95() >= r.Max() {
		t.Errorf("at the default %d tries, P95 (%v) must differ from Max (%v)",
			config.ScanDefaults.Tries, r.P95(), r.Max())
	}
}

// TestWithinLossLimit pins the packet-loss cutoff. Exactly 25% is kept; above
// it the IP is dropped. With 8 tries the reported steps are 12.5%, so 25% is
// reachable and must be inclusive.
func TestWithinLossLimit(t *testing.T) {
	tests := []struct {
		name      string
		latencies []time.Duration
		wantKeep  bool
	}{
		{
			name:      "no loss",
			latencies: samples(8, 300),
			wantKeep:  true,
		},
		{
			name:      "one of eight lost is 12.5 percent",
			latencies: []time.Duration{0, ms(10), ms(20), ms(30), ms(40), ms(50), ms(60), ms(70)},
			wantKeep:  true,
		},
		{
			name:      "exactly 25 percent is kept",
			latencies: []time.Duration{0, 0, ms(10), ms(20), ms(30), ms(40), ms(50), ms(60)},
			wantKeep:  true,
		},
		{
			name:      "37.5 percent is dropped",
			latencies: []time.Duration{0, 0, 0, ms(10), ms(20), ms(30), ms(40), ms(50)},
			wantKeep:  false,
		},
		{
			name:      "50 percent is dropped",
			latencies: []time.Duration{0, 0, 0, 0, ms(10), ms(20), ms(30), ms(40)},
			wantKeep:  false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &Result{Latencies: tc.latencies}
			if got := r.WithinLossLimit(); got != tc.wantKeep {
				t.Errorf("loss %.1f%%: WithinLossLimit() = %v, want %v",
					r.Loss(), got, tc.wantKeep)
			}
		})
	}
}

// TestTopNRejectsHighLossAndRanksByReliability uses a miniature version of a
// real scan: a fast IP that drops 40% of requests, and a slower IP that drops
// none. Sorting by average alone promoted the lossy one to the top.
func TestTopNRejectsHighLossAndRanksByReliability(t *testing.T) {
	mk := func(ip string, lat []time.Duration, colo string) *Result {
		return &Result{IP: net.ParseIP(ip), Port: 443, ProbeMode: "http", Latencies: lat,
			TLSOk: true, Colo: colo, HTTPStatus: 200}
	}
	results := []*Result{
		// fast but 40% loss
		mk("1.1.1.1", []time.Duration{0, 0, 0, ms(30), ms(31), ms(32), ms(33), ms(34)}, "AMS"),
		// slower but lossless
		mk("1.1.1.2", samples(8, 120), "AMS"),
		// mid loss, still kept
		mk("1.1.1.3", []time.Duration{0, ms(40), ms(41), ms(42), ms(43), ms(44), ms(45), ms(46)}, "FRA"),
	}

	top := TopN(results, 0)
	if len(top) != 2 {
		t.Fatalf("TopN(0) returned %d results, want 2 (the 40%%-loss IP must be dropped)", len(top))
	}
	if top[0].IP.String() != "1.1.1.2" {
		t.Errorf("first = %s, want 1.1.1.2 (lossless beats faster-but-lossy)", top[0].IP)
	}
	if top[1].IP.String() != "1.1.1.3" {
		t.Errorf("second = %s, want 1.1.1.3 (12.5%% loss is inside the limit)", top[1].IP)
	}
}

// TestSortByReliableOrdersByLossFirst pins the new sort criterion directly.
func TestSortByReliableOrdersByLossFirst(t *testing.T) {
	mk := func(ip string, lat []time.Duration) *Result {
		return &Result{IP: net.ParseIP(ip), Port: 443, ProbeMode: "http", Latencies: lat,
			TLSOk: true, Colo: "AMS", HTTPStatus: 200}
	}
	rows := []*Result{
		mk("1.1.1.3", []time.Duration{0, 0, 0, ms(30), ms(31), ms(32), ms(33), ms(34)}),
		mk("1.1.1.1", samples(8, 100)),
		mk("1.1.1.2", []time.Duration{0, ms(50), ms(51), ms(52), ms(53), ms(54), ms(55), ms(56)}),
	}
	Sort(rows, SortByReliable)

	if rows[0].IP.String() != "1.1.1.1" {
		t.Errorf("first = %s, want 1.1.1.1 (0%% loss)", rows[0].IP)
	}
	if rows[1].IP.String() != "1.1.1.2" {
		t.Errorf("second = %s, want 1.1.1.2 (12.5%% loss)", rows[1].IP)
	}
	if rows[2].IP.String() != "1.1.1.3" {
		t.Errorf("third = %s, want 1.1.1.3 (37.5%% loss sorts last)", rows[2].IP)
	}
}
