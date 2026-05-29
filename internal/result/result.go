package result

import (
	"math"
	"net"
	"sort"
	"time"
)

// Result holds all measured statistics for a single Cloudflare IP.
type Result struct {
	IP          net.IP
	Port        int
	ProbeMode   string          // tcp | tls | http
	Latencies   []time.Duration // per-try latencies; 0 = failed try
	Errors      []ProbeError    // per-try failure class; len matches Latencies
	TLSOk       bool
	HTTPStatus  int
	Colo        string
	Throughput  float64 // bytes/sec, 0 if not measured
	SpeedTested bool    // true when a payload download check was attempted
	Timestamp   time.Time
}

// ProbeError classifies why a single probe try failed. Distinguishing failure
// modes turns a flat "loss" percentage into a diagnosis: a refused connection
// (origin reachable, port closed) is very different from a timeout (filtered or
// blackholed by DPI), a reset (mid-stream RST), or a TLS error (SNI mismatch /
// interception). This lets the scanner — and the user — tell "this IP is dead"
// apart from "this network is throttling me".
type ProbeError uint8

const (
	ErrNone    ProbeError = iota // the try succeeded (connection established)
	ErrTimeout                   // deadline exceeded — often DPI blackholing
	ErrRefused                   // connection refused — port closed
	ErrReset                     // connection reset by peer
	ErrTLS                       // TLS handshake / certificate failure
	ErrHTTP                      // connected, but not a valid Cloudflare edge
	ErrOther                     // anything else
)

// String returns a short, stable label for a probe error class.
func (e ProbeError) String() string {
	switch e {
	case ErrNone:
		return "ok"
	case ErrTimeout:
		return "timeout"
	case ErrRefused:
		return "refused"
	case ErrReset:
		return "reset"
	case ErrTLS:
		return "tls"
	case ErrHTTP:
		return "http"
	default:
		return "other"
	}
}

// DominantError returns the most common non-success failure class across the
// tries, or ErrNone when at least one try connected. Ties break toward the more
// severe (lower-numbered) class so the worst symptom surfaces. It returns
// ErrNone when no error information was recorded.
func (r *Result) DominantError() ProbeError {
	var counts [ErrOther + 1]int
	connected := false
	for _, e := range r.Errors {
		if e == ErrNone {
			connected = true
			continue
		}
		counts[e]++
	}
	if connected {
		return ErrNone
	}
	best := ErrNone
	bestN := 0
	for class := ErrTimeout; class <= ErrOther; class++ {
		if counts[class] > bestN {
			bestN = counts[class]
			best = class
		}
	}
	return best
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

// Median returns the median of successful latency measurements.
//
// It is more robust than Avg for ranking edges: the first try to an IP often
// pays a one-off cost (cold route cache, ARP, a slow TLS handshake) that drags
// the mean up and unfairly penalises an otherwise-fast IP. The median ignores
// that single outlier, so Score reflects the latency the user will actually
// experience on a warm connection.
func (r *Result) Median() time.Duration {
	succ := make([]time.Duration, 0, len(r.Latencies))
	for _, l := range r.Latencies {
		if l > 0 {
			succ = append(succ, l)
		}
	}
	if len(succ) == 0 {
		return 0
	}
	sort.Slice(succ, func(i, j int) bool { return succ[i] < succ[j] })
	n := len(succ)
	if n%2 == 1 {
		return succ[n/2]
	}
	return (succ[n/2-1] + succ[n/2]) / 2
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
	if r.Loss() >= 50 || r.Avg() <= 0 {
		return false
	}

	switch r.ProbeMode {
	case "http":
		// TLS is only expected on port 443; plain HTTP (port 80) has no TLS.
		if r.Port == 443 && !r.TLSOk {
			return false
		}
		if r.HTTPStatus < 200 || r.HTTPStatus >= 400 || r.Colo == "" {
			return false
		}
		if r.SpeedTested && r.Throughput <= 0 {
			return false
		}
		return true
	case "tls":
		return r.TLSOk
	default: // tcp
		return true
	}
}

// Score returns a single 0–100 quality score (higher is better) that combines
// the metrics that matter when picking a Cloudflare edge IP: packet loss,
// average latency, jitter, and — when it was measured — download throughput.
// It lets results be ranked by overall quality instead of one axis at a time.
//
// An unhealthy result scores 0. Colo is intentionally excluded: a preferred
// colo is a user choice applied as a filter, not an intrinsic property of the
// IP, so folding it into the score would bias every user's ranking the same way.
func (r *Result) Score() float64 {
	if !r.IsHealthy() {
		return 0
	}
	clamp01 := func(f float64) float64 {
		if f < 0 {
			return 0
		}
		if f > 1 {
			return 1
		}
		return f
	}
	lossScore := clamp01(1 - r.Loss()/100)
	latScore := clamp01(1 - Ms(r.Median())/500) // median: robust to a slow first try; 0ms → 1, ≥500ms → 0
	jitScore := clamp01(1 - Ms(r.Jitter())/100) // 0ms → 1, ≥100ms → 0

	// When a real download was measured, give throughput a fifth of the weight;
	// otherwise redistribute that weight onto loss and latency so latency-only
	// scans still rank sensibly.
	if r.SpeedTested && r.Throughput > 0 {
		thrScore := clamp01(r.Throughput / (10 * 1024 * 1024)) // 0–10 MB/s → 0–1
		return 100 * (0.35*lossScore + 0.30*latScore + 0.15*jitScore + 0.20*thrScore)
	}
	return 100 * (0.45*lossScore + 0.40*latScore + 0.15*jitScore)
}

// SortBy defines the available sort criteria.
type SortBy int

const (
	SortByAvg SortBy = iota
	SortByLoss
	SortByJitter
	SortByColo
	SortBySpeed
	SortByScore
)

// Sort reorders results in-place according to the given criterion (ascending).
//
// The per-result metrics (avg, jitter, loss) are derived by scanning the
// latency slice, so they are computed once up front rather than recomputed on
// every comparison — turning an O(n·log n · tries) sort into O(n·log n).
func Sort(results []*Result, by SortBy) {
	type sortKey struct {
		r          *Result
		avg, jit   time.Duration
		loss, thru float64
		score      float64
		colo       string
	}
	keys := make([]sortKey, len(results))
	for i, r := range results {
		keys[i] = sortKey{
			r:     r,
			avg:   r.Avg(),
			jit:   r.Jitter(),
			loss:  r.Loss(),
			thru:  r.Throughput,
			score: r.Score(),
			colo:  r.Colo,
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		switch by {
		case SortByLoss:
			if a.loss != b.loss {
				return a.loss < b.loss
			}
			return a.avg < b.avg
		case SortByJitter:
			if a.jit != b.jit {
				return a.jit < b.jit
			}
			return a.avg < b.avg
		case SortByColo:
			if a.colo != b.colo {
				return a.colo < b.colo
			}
			return a.avg < b.avg
		case SortBySpeed:
			if a.thru != b.thru {
				return a.thru > b.thru
			}
			return a.avg < b.avg
		case SortByScore:
			if a.score != b.score {
				return a.score > b.score // higher score ranks first
			}
			return a.avg < b.avg
		default: // SortByAvg
			if a.avg != b.avg {
				if a.avg == 0 {
					return false
				}
				if b.avg == 0 {
					return true
				}
				return a.avg < b.avg
			}
			return a.loss < b.loss
		}
	})

	for i, k := range keys {
		results[i] = k.r
	}
}

// Ms converts a duration to fractional milliseconds, preserving sub-millisecond
// precision. Duration.Milliseconds truncates to whole milliseconds, which loses
// meaningful detail for sub-millisecond and edge-latency measurements.
func Ms(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

// TopN returns the n best results by composite score (ignoring unhealthy IPs).
func TopN(results []*Result, n int) []*Result {
	var healthy []*Result
	for _, r := range results {
		if r.IsHealthy() {
			healthy = append(healthy, r)
		}
	}
	Sort(healthy, SortByScore)
	if n > 0 && n < len(healthy) {
		return healthy[:n]
	}
	return healthy
}
