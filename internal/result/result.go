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
	TLSOk       bool
	WSOk        bool // WebSocket connection survived hold test
	RequireWS   bool // true when WebSocket success is part of health criteria
	HTTPStatus  int
	Colo        string
	Throughput  float64 // bytes/sec, 0 if not measured
	SpeedTested bool    // true when a payload download check was attempted
	Timestamp   time.Time
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

// DownloadOK reports whether a single download-bearing HTTP probe succeeded. IsHealthy cannot judge
// one-try probes (it needs two successful tries), which made every direct speed test fail.
func (r *Result) DownloadOK() bool {
	return r.Throughput > 0 && r.Avg() > 0 && r.HTTPStatus >= 200 && r.HTTPStatus < 400
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

// SortBy defines the available sort criteria.
type SortBy int

const (
	SortByAvg SortBy = iota
	SortByLoss
	SortByJitter
	SortByColo
	SortBySpeed
)

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

// TopN returns the n best results by Avg latency (ignoring unhealthy IPs).
func TopN(results []*Result, n int) []*Result {
	var healthy []*Result
	for _, r := range results {
		if r.IsHealthy() {
			healthy = append(healthy, r)
		}
	}
	Sort(healthy, SortByAvg)
	if n > 0 && n < len(healthy) {
		return healthy[:n]
	}
	return healthy
}
