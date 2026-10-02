package result

import (
	"net"
	"testing"
	"time"
)

// TestSortByReliableBreaksLossTiesOnAvg pins the second sort key.
//
// SortByReliable used to order equal-loss rows by P95, a statistic no result
// table displays. Within a group of IPs that all lost the same fraction of
// probes, the visible order was therefore decided by a number the user could not
// see, and TopN applied the same hidden key when choosing which IPs went on to
// Phase 2 validation.
//
// Avg is the column the tables actually print, so it is now the tiebreak.
func TestSortByReliableBreaksLossTiesOnAvg(t *testing.T) {
	// Same loss on all three, so the tiebreak alone decides the order.
	mk := func(ip string, lats []time.Duration) *Result {
		return &Result{IP: net.ParseIP(ip), Port: 443, ProbeMode: "http", Latencies: lats,
			TLSOk: true, Colo: "AMS", HTTPStatus: 200}
	}
	// tail is far worse on the middle IP, so a P95 tiebreak would put it last.
	rows := []*Result{
		mk("1.1.1.3", []time.Duration{ms(900), ms(900), ms(900), ms(900)}),
		mk("1.1.1.2", []time.Duration{ms(200), ms(200), ms(200), ms(9000)}),
		mk("1.1.1.1", []time.Duration{ms(300), ms(300), ms(300), ms(300)}),
	}
	for _, r := range rows {
		if got := r.Loss(); got != 0 {
			t.Fatalf("%s loss = %v, want 0 for this fixture", r.IP, got)
		}
	}

	Sort(rows, SortByReliable)

	want := []string{"1.1.1.1", "1.1.1.3", "1.1.1.2"}
	for i, w := range want {
		if got := rows[i].IP.String(); got != w {
			names := []string{}
			for _, r := range rows {
				names = append(names, r.IP.String())
			}
			t.Fatalf("order = %v, want %v (equal loss must fall back to Avg, not tail latency)", names, want)
		}
	}
}

// TestTopNSelectsByLossThenAvg guards the Phase 2 candidate selection, which is
// the user-visible consequence of the tiebreak: TopN feeds the same comparator,
// so the IPs that reach xray validation are the ones the Phase 1 table ranks
// first.
func TestTopNSelectsByLossThenAvg(t *testing.T) {
	mk := func(ip string, lats []time.Duration) *Result {
		return &Result{IP: net.ParseIP(ip), Port: 443, ProbeMode: "http", Latencies: lats,
			TLSOk: true, Colo: "AMS", HTTPStatus: 200}
	}
	results := []*Result{
		// Lowest average, but drops a quarter of its probes.
		mk("1.1.1.3", []time.Duration{ms(50), ms(50), ms(50), 0}),
		// No loss, slow.
		mk("1.1.1.2", []time.Duration{ms(800), ms(800), ms(800), ms(800)}),
		// No loss, quick.
		mk("1.1.1.1", []time.Duration{ms(200), ms(200), ms(200), ms(200)}),
	}

	top := TopN(results, 2)
	if len(top) != 2 {
		t.Fatalf("TopN(2) returned %d results, want 2", len(top))
	}
	if got := top[0].IP.String(); got != "1.1.1.1" {
		t.Errorf("first = %s, want 1.1.1.1 (zero loss, fastest average)", got)
	}
	if got := top[1].IP.String(); got != "1.1.1.2" {
		t.Errorf("second = %s, want 1.1.1.2 (zero loss; 1.1.1.3 drops a quarter of probes)", got)
	}
}