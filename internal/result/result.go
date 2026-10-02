package result

import (
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"time"
)

// Result holds all measured statistics for a single Cloudflare IP.
type Result struct {
	IP          net.IP
	Port        int
	ProbeMode   string          // tcp | tls | http
	Latencies   []time.Duration // per-try latencies; 0 = failed try
	TLSOk       bool
	WSOk        bool // WebSocket connection survived hold test
	RequireWS   bool // true when WebSocket success is part of health criteria
	HTTPStatus  int
	Colo        string
	Throughput  float64 // bytes/sec, 0 if not measured
	SpeedTested bool    // true when a payload download check was attempted
	Timestamp   time.Time
}

// MaxLossPct is the highest packet loss an IP may show and still be reported.
// Anything strictly above it is dropped before it reaches any result list.
// A proxy carrying 25%+ loss is unusable in practice: TCP retransmits mask
// it for large transfers, but interactive traffic stutters and every stall
// gets re-tested, so such an IP only wastes probe time and Phase 2 slots.
const MaxLossPct = 25.0

// WithinLossLimit reports whether the result is inside the packet-loss cutoff.
// The comparison is inclusive: exactly MaxLossPct is kept.
func (r *Result) WithinLossLimit() bool {
	return r.Loss() <= MaxLossPct
}

// Loss returns packet loss percentage (0–100).
func (r *Result) Loss() float64 {
	if len(r.Latencies) == 0 {
		return 100
	}
	failed := 0
	for _, l := range r.Latencies {
		if l == 0 {
			failed++
		}
	}
	return float64(failed) / float64(len(r.Latencies)) * 100
}

// TrimmedAvg returns a mean that discards the worst successful sample before
// averaging, which is the convention ping tools use for a headline latency.
//
// Unlike a percentile this is not a rank lookup, so it stays a genuine average
// and stays distinct from Max whenever at least three samples succeeded. With
// the sample counts this scanner uses, P90 and P95 collapse onto Max by
// construction: nearest-rank at N successful samples below ten selects rank
// ceil(0.9*N) == N, so the "tail" column was just a second copy of Max carrying
// a percentile name it did not deserve. The trim fraction scales the same way —
// a quarter of the samples at the slow end are dropped, with at least one and
// never more than a third, so small and large counts both behave.
//
// The result sits below Avg, never above it: dropping the slowest samples can
// only lower a mean. Read it as "typical latency once the worst outlier is
// discounted", which is lower and steadier than Avg and never collapses onto
// Max the way a percentile does at these sample sizes.
//
// Returns 0 when fewer than two samples succeeded, and falls back to Avg when
// trimming would leave nothing to average.
func (r *Result) TrimmedAvg() time.Duration {
	samples := make([]time.Duration, 0, len(r.Latencies))
	for _, l := range r.Latencies {
		if l > 0 {
			samples = append(samples, l)
		}
	}
	if len(samples) < 2 {
		return 0
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })

	discard := len(samples) / 4
	if discard < 1 {
		discard = 1
	}
	if max := len(samples) / 3; discard > max {
		discard = max
	}
	kept := samples[:len(samples)-discard]
	if len(kept) == 0 {
		return r.Avg()
	}
	var sum time.Duration
	for _, l := range kept {
		sum += l
	}
	return sum / time.Duration(len(kept))
}

// Avg returns the mean of successful latency measurements.
func (r *Result) Avg() time.Duration {
	var sum time.Duration
	var count int
	for _, l := range r.Latencies {
		if l > 0 {
			sum += l
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / time.Duration(count)
}

// Min returns the best successful latency.
func (r *Result) Min() time.Duration {
	var m time.Duration
	for _, l := range r.Latencies {
		if l > 0 && (m == 0 || l < m) {
			m = l
		}
	}
	return m
}

// Max returns the worst successful latency.
func (r *Result) Max() time.Duration {
	var m time.Duration
	for _, l := range r.Latencies {
		if l > m {
			m = l
		}
	}
	return m
}

// Jitter returns the standard deviation of successful latencies.
func (r *Result) Jitter() time.Duration {
	var count int
	for _, l := range r.Latencies {
		if l > 0 {
			count++
		}
	}
	if count < 2 {
		return 0
	}
	avg := float64(r.Avg())
	var variance float64
	for _, l := range r.Latencies {
		if l > 0 {
			diff := float64(l) - avg
			variance += diff * diff
		}
	}
	variance /= float64(count)
	return time.Duration(math.Sqrt(variance))
}

// IsHealthy returns true only when the probe mode's success criteria are met.
// A failed try must record latency 0; timeouts must never count as success.
func (r *Result) IsHealthy() bool {
	if len(r.Latencies) < 2 {
		return false
	}

	successCount := 0
	for _, l := range r.Latencies {
		if l > 0 {
			successCount++
		}
	}
	// Two successful tries are the floor. Note that the Loss() >= 50 check
	// below already rejects anything worse than half, so no separate
	// "majority of tries" test is needed here.
	if successCount < 2 {
		return false
	}

	if r.Loss() >= 50 || r.Avg() <= 0 {
		return false
	}

	switch r.ProbeMode {
	case "http":
		// Plain HTTP (port 80) has no TLS; every other HTTP-mode port is HTTPS.
		if r.Port != 80 && !r.TLSOk {
			return false
		}
		if r.HTTPStatus < 200 || r.HTTPStatus >= 400 || r.Colo == "" {
			return false
		}
		// Require successful download when a speed test was attempted.
		// On Iranian ISPs, /cdn-cgi/trace succeeds but actual data transfer
		// fails — this catches IPs that are reachable but can't carry traffic.
		if r.SpeedTested && r.Throughput <= 0 {
			return false
		}
		if r.RequireWS && !r.WSOk {
			return false
		}
		return true
	case "tls":
		return r.TLSOk
	default: // tcp
		return true
	}
}

// P95 returns the 95th percentile latency, computed with the nearest-rank
// method over successful samples. It exposes tail behaviour that Avg and
// Jitter hide: on a DPI-throttled link the average can look acceptable while
// the worst requests are several times slower. Returns 0 when fewer than two
// samples succeeded.
func (r *Result) P95() time.Duration {
	return r.percentile(95)
}

// P90 returns the 90th percentile latency. Returns 0 when fewer than two
// samples succeeded.
func (r *Result) P90() time.Duration {
	return r.percentile(90)
}

// percentile computes a percentile over successful latencies, discarding the
// slowest tail. It sorts a copy so the caller's slice order is preserved.
//
// The number of discarded samples is floor((100-p)/100 * N), with a floor of
// one. This deviates from standard nearest-rank (rank = ceil(p/100 * N)) on
// purpose: nearest-rank discards nothing below twenty tries and reports
// exactly Max, so at this scanner's sample sizes the column would be a second
// copy of Max.
//
// That deviation is why the percentile is no longer surfaced in the UI. The
// live file and the result tables report TrimmedAvg instead, which is a real
// average and stays distinct from Max. Percentile and PercentileLabel remain
// for the CSV export, which records raw statistics for later analysis, and the
// label keeps reporting the true percentile so the exported figure is not
// misread as a higher one.
func (r *Result) percentile(p int) time.Duration {
	samples := make([]time.Duration, 0, len(r.Latencies))
	for _, l := range r.Latencies {
		if l > 0 {
			samples = append(samples, l)
		}
	}
	if len(samples) < 2 {
		return 0
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })

	discard := ((100 - p) * len(samples)) / 100
	if discard < 1 {
		discard = 1
	}
	idx := len(samples) - discard - 1
	if idx < 0 {
		idx = 0
	}
	return samples[idx]
}

// PercentileLabel returns the percentile P95 actually reports for this
// result's successful sample count, e.g. "P87.5" for 8 tries. It returns an
// empty string when there are too few samples for the value to mean anything.
func (r *Result) PercentileLabel() string {
	n := 0
	for _, l := range r.Latencies {
		if l > 0 {
			n++
		}
	}
	if n < 2 {
		return ""
	}
	discard := ((100 - 95) * n) / 100
	if discard < 1 {
		discard = 1
	}
	effective := float64(n-discard) / float64(n) * 100
	if effective == math.Trunc(effective) {
		return fmt.Sprintf("P%d", int(effective))
	}
	return fmt.Sprintf("P%.1f", effective)
}

// SortBy defines the available sort criteria.
type SortBy int

const (
	SortByAvg SortBy = iota
	SortByLoss
	SortByJitter
	SortByColo
	SortBySpeed
	SortByP95      // tail latency; prefers IPs that stay fast on their worst try
	SortByReliable // packet loss first, then tail latency -- the honest order
)

func (s SortBy) String() string {
	switch s {
	case SortByLoss:
		return "loss"
	case SortByJitter:
		return "jitter"
	case SortByColo:
		return "colo"
	case SortBySpeed:
		return "speed"
	case SortByP95:
		return "p95"
	case SortByReliable:
		return "reliable"
	default:
		return "avg"
	}
}

// ParseSortBy resolves a sort-criterion name. Unknown names fall back to
// SortByAvg so a stale preference file can never break startup.
func ParseSortBy(s string) SortBy {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "loss":
		return SortByLoss
	case "jitter":
		return SortByJitter
	case "colo":
		return SortByColo
	case "speed":
		return SortBySpeed
	case "p95":
		return SortByP95
	default:
		return SortByAvg
	}
}

// sortNames is the display order of sort criteria. Its length is the number of
// criteria the UI cycles through, so keep it in step with the SortBy constants.
var sortNames = []string{"avg", "loss", "jitter", "colo", "speed", "p95", "reliable"}

// SortCount returns how many sort criteria exist. Callers cycling through
// criteria should use this rather than a hardcoded number, so adding a
// criterion cannot silently make the cycle skip or repeat one.
func SortCount() int { return len(sortNames) }

// SortNames returns the display names of every sort criterion, in cycle order.
// The returned slice is a copy; callers may modify it freely.
func SortNames() []string {
	out := make([]string, len(sortNames))
	copy(out, sortNames)
	return out
}

func sortRank(r *Result) int {
	if r.IsHealthy() {
		return 0
	}
	if r.Avg() > 0 || r.Loss() < 100 {
		return 1
	}
	return 2
}

func cmpBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return -1
	default:
		return 1
	}
}

func cmpDuration(a, b time.Duration) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func cmpFloatAsc(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func cmpFloatDesc(a, b float64) int {
	return -cmpFloatAsc(a, b)
}

func cmpString(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareResults(a, b *Result, by SortBy) int {
	if rankCmp := sortRank(a) - sortRank(b); rankCmp != 0 {
		return rankCmp
	}

	switch by {
	case SortByLoss:
		if cmp := cmpFloatAsc(a.Loss(), b.Loss()); cmp != 0 {
			return cmp
		}
		if cmp := cmpDuration(a.Avg(), b.Avg()); cmp != 0 {
			return cmp
		}
		if cmp := cmpDuration(a.Jitter(), b.Jitter()); cmp != 0 {
			return cmp
		}
	case SortByJitter:
		if cmp := cmpDuration(a.Jitter(), b.Jitter()); cmp != 0 {
			return cmp
		}
		if cmp := cmpFloatAsc(a.Loss(), b.Loss()); cmp != 0 {
			return cmp
		}
		if cmp := cmpDuration(a.Avg(), b.Avg()); cmp != 0 {
			return cmp
		}
	case SortByColo:
		if cmp := cmpString(a.Colo, b.Colo); cmp != 0 {
			return cmp
		}
		if cmp := cmpDuration(a.Avg(), b.Avg()); cmp != 0 {
			return cmp
		}
		if cmp := cmpFloatAsc(a.Loss(), b.Loss()); cmp != 0 {
			return cmp
		}
	case SortBySpeed:
		if cmp := cmpFloatDesc(a.Throughput, b.Throughput); cmp != 0 {
			return cmp
		}
		if cmp := cmpDuration(a.Avg(), b.Avg()); cmp != 0 {
			return cmp
		}
		if cmp := cmpFloatAsc(a.Loss(), b.Loss()); cmp != 0 {
			return cmp
		}
	case SortByP95:
		// Ascending: the IP whose worst sample is smallest wins. Loss is the
		// tiebreaker because two IPs can share a p95 yet differ sharply in how
		// often they dropped the request entirely.
		if cmp := cmpDuration(a.P95(), b.P95()); cmp != 0 {
			return cmp
		}
		if cmp := cmpFloatAsc(a.Loss(), b.Loss()); cmp != 0 {
			return cmp
		}
		if cmp := cmpDuration(a.Avg(), b.Avg()); cmp != 0 {
			return cmp
		}
	case SortByReliable:
		// Loss dominates: an IP that drops a quarter of its requests is worse
		// than a slower IP that never drops one, so it cannot rank above one
		// however good its average looks.
		//
		// This then sorts on Avg, the column the table actually prints. It used
		// to fall through to P95, which no table ever displayed: within a group
		// of equal-loss rows the visible order was decided by a number the user
		// could not see, and TopN applied the same hidden key when choosing which
		// IPs reached Phase 2. Sorting on the printed column keeps the ranking
		// and the display telling the same story.
		if cmp := cmpFloatAsc(a.Loss(), b.Loss()); cmp != 0 {
			return cmp
		}
		if cmp := cmpDuration(a.Avg(), b.Avg()); cmp != 0 {
			return cmp
		}
	default:
		if cmp := cmpDuration(a.Avg(), b.Avg()); cmp != 0 {
			return cmp
		}
		if cmp := cmpFloatAsc(a.Loss(), b.Loss()); cmp != 0 {
			return cmp
		}
		if cmp := cmpDuration(a.Jitter(), b.Jitter()); cmp != 0 {
			return cmp
		}
	}

	if cmp := cmpBool(a.TLSOk, b.TLSOk); cmp != 0 {
		return cmp
	}
	if cmp := cmpBool(a.WSOk, b.WSOk); cmp != 0 {
		return cmp
	}
	if cmp := cmpString(a.IP.String(), b.IP.String()); cmp != 0 {
		return cmp
	}
	return 0
}

// Sort reorders results in-place according to the given criterion (ascending).
func Sort(results []*Result, by SortBy) {
	sort.SliceStable(results, func(i, j int) bool {
		return compareResults(results[i], results[j], by) < 0
	})
}

// TopN returns the n best usable results (unhealthy and over-loss IPs are
// dropped), ranked by packet loss then tail latency. This is the candidate
// list Phase 2 validates, so the ordering decides which IPs actually get
// tested: sorting by average alone used to promote 40%-loss IPs to the top.
func TopN(results []*Result, n int) []*Result {
	var healthy []*Result
	for _, r := range results {
		if r.IsHealthy() && r.WithinLossLimit() {
			healthy = append(healthy, r)
		}
	}
	Sort(healthy, SortByReliable)
	if n > 0 && n < len(healthy) {
		return healthy[:n]
	}
	return healthy
}
