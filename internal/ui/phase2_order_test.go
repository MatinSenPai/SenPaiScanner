package ui

import (
	"net"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

// p1Row builds a Phase 1 row with an explicit loss count and average, so a test
// can state the ranking it expects without restating the latency maths.
//
// z is the number of failed probes out of n; the remaining probes are all avg.
func p1Row(ip string, port, n, z int, avg time.Duration) *result.Result {
	lats := make([]time.Duration, 0, n)
	for i := 0; i < z; i++ {
		lats = append(lats, 0)
	}
	for i := z; i < n; i++ {
		lats = append(lats, avg)
	}
	return &result.Result{
		IP:         net.ParseIP(ip),
		Port:       port,
		Latencies:  lats,
		ProbeMode:  "http",
		TLSOk:      true,
		HTTPStatus: 200,
	}
}

func vRow(ip string, port int, ok bool) *xraytest.ValidationResult {
	return &xraytest.ValidationResult{IP: ip, Port: port, Transport: "ws", Success: ok}
}

// TestPhase2FollowsPhase1Order is the core guarantee behind showing Phase 1
// metrics beside Phase 2 results: the two tables have to be comparable row for
// row. Phase 2 results arrive in worker-completion order, so without this the
// second table is shuffled relative to the first and the columns beside it are
// only useful if you re-derive the mapping yourself.
func TestPhase2FollowsPhase1Order(t *testing.T) {
	// Phase 1 order by loss then avg: a (0% loss, 300ms) ranks before
	// b (0% loss, 700ms), which ranks before c (25% loss, 100ms) because loss
	// dominates. Passing them in shuffled Phase 2 arrival order must not matter.
	p1 := []*result.Result{
		p1Row("1.1.1.3", 443, 4, 1, 100*time.Millisecond), // 25% loss
		p1Row("1.1.1.2", 443, 4, 0, 700*time.Millisecond), // 0%, slow
		p1Row("1.1.1.1", 443, 4, 0, 300*time.Millisecond), // 0%, fast
	}
	p2 := []*xraytest.ValidationResult{
		vRow("1.1.1.3", 443, false),
		vRow("1.1.1.1", 443, true),
		vRow("1.1.1.2", 443, true),
	}

	got := phase2InPhase1Order(p1, p2)
	want := []string{"1.1.1.1", "1.1.1.2", "1.1.1.3"}
	for i, w := range want {
		if got[i].IP != w {
			t.Errorf("row %d = %s, want %s", i, got[i].IP, w)
		}
	}
}

// TestPhase2OrderKeepsUnknownEndpointsAtTheEnd covers the case where a Phase 2
// result exists for an endpoint Phase 1 never recorded. It must not be dropped,
// and it must not displace a ranked row.
func TestPhase2OrderKeepsUnknownEndpointsAtTheEnd(t *testing.T) {
	p1 := []*result.Result{
		p1Row("1.1.1.2", 443, 4, 0, 700*time.Millisecond),
		p1Row("1.1.1.1", 443, 4, 0, 300*time.Millisecond),
	}
	p2 := []*xraytest.ValidationResult{
		vRow("9.9.9.9", 443, false), // never seen in Phase 1
		vRow("1.1.1.1", 443, true),
		vRow("1.1.1.2", 443, true),
	}

	got := phase2InPhase1Order(p1, p2)
	if len(got) != 3 {
		t.Fatalf("got %d rows, want 3 (an unknown endpoint must not be dropped)", len(got))
	}
	if got[0].IP != "1.1.1.1" || got[1].IP != "1.1.1.2" {
		t.Errorf("ranked rows = %s,%s; want 1.1.1.1,1.1.1.2 first", got[0].IP, got[1].IP)
	}
	if got[2].IP != "9.9.9.9" {
		t.Errorf("last row = %s, want 9.9.9.9 (unknown endpoint sorted last)", got[2].IP)
	}
}

// TestPhase2OrderDoesNotMutateInput guards the ordering helper against sorting
// m.configResults in place. The model appends to that slice as results arrive,
// so an in-place sort would reshuffle live state on every rendered frame.
func TestPhase2OrderDoesNotMutateInput(t *testing.T) {
	p1 := []*result.Result{
		p1Row("1.1.1.2", 443, 4, 0, 700*time.Millisecond),
		p1Row("1.1.1.1", 443, 4, 0, 300*time.Millisecond),
	}
	p2 := []*xraytest.ValidationResult{
		vRow("1.1.1.2", 443, true),
		vRow("1.1.1.1", 443, true),
	}
	arrival := []string{"1.1.1.2", "1.1.1.1"}

	_ = phase2InPhase1Order(p1, p2)

	for i, want := range arrival {
		if p2[i].IP != want {
			t.Errorf("input p2[%d] = %s, want %s (input slice was sorted in place)", i, p2[i].IP, want)
		}
	}
}

// TestPhase2OrderHandlesEmptyAndNil pins the two degenerate inputs, because the
// caller renders before any Phase 2 result exists and after Phase 1 was skipped.
func TestPhase2OrderHandlesEmptyAndNil(t *testing.T) {
	if got := phase2InPhase1Order(nil, nil); got != nil {
		t.Errorf("nil/nil = %v, want nil", got)
	}
	p1 := []*result.Result{p1Row("1.1.1.1", 443, 4, 0, 300*time.Millisecond)}
	if got := phase2InPhase1Order(p1, nil); got != nil {
		t.Errorf("p1 with no p2 rows = %v, want nil", got)
	}
}

// TestPhase1IndexLookupFindsMetrics covers the O(1) lookup used to fill the
// P1-LOSS / P1-AVG columns, including the miss that must render as a dash.
func TestPhase1IndexLookupFindsMetrics(t *testing.T) {
	p1 := []*result.Result{
		p1Row("1.1.1.1", 443, 4, 0, 300*time.Millisecond),
		p1Row("1.1.1.2", 8443, 4, 1, 100*time.Millisecond),
	}
	idx := newPhase1Index(p1)

	if got := idx.get("1.1.1.1", 443); got == nil || got.Avg().Milliseconds() != 300 {
		t.Errorf("get(1.1.1.1:443) = %v, want avg 300ms", got)
	}
	if got := idx.get("1.1.1.2", 8443); got == nil || got.Loss() != 25 {
		t.Errorf("get(1.1.1.2:8443) = %v, want loss 25%%", got)
	}
	// A nil-receiver lookup must be safe: the Phase 2 page can render before any
	// Phase 1 row has arrived.
	var nilIdx *phase1Index
	if got := nilIdx.get("1.1.1.1", 443); got != nil {
		t.Errorf("nil index get = %v, want nil", got)
	}
}

// TestPhase1IndexDistinguishesPorts guards the key format. The key is built with
// Sprintf("%s:%d"), so an IP and port pair must never collide with a bare IP.
func TestPhase1IndexDistinguishesPorts(t *testing.T) {
	p1 := []*result.Result{
		p1Row("1.1.1.1", 443, 4, 0, 300*time.Millisecond),
		p1Row("1.1.1.1", 8443, 4, 0, 900*time.Millisecond),
	}
	idx := newPhase1Index(p1)

	if got := idx.get("1.1.1.1", 443); got == nil || got.Avg().Milliseconds() != 300 {
		t.Errorf("get(:443) = %v, want avg 300ms", got)
	}
	if got := idx.get("1.1.1.1", 8443); got == nil || got.Avg().Milliseconds() != 900 {
		t.Errorf("get(:8443) = %v, want avg 900ms", got)
	}
}