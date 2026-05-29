package store

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/result"
)

func testResult(ip string, scoreSpeed float64) *result.Result {
	return &result.Result{
		IP:          net.ParseIP(ip),
		Port:        443,
		ProbeMode:   "http",
		Latencies:   []time.Duration{40 * time.Millisecond, 45 * time.Millisecond},
		TLSOk:       true,
		HTTPStatus:  200,
		Colo:        "FRA",
		Throughput:  scoreSpeed,
		SpeedTested: true,
		Timestamp:   time.Now(),
	}
}

func TestLeaderboardSaveAndTopIPs(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := db.SaveResult(ctx, testResult("1.1.1.1", 512*1024)); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveResult(ctx, testResult("1.1.1.2", 8*1024*1024)); err != nil {
		t.Fatal(err)
	}

	ips, err := db.TopIPs(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(ips) != 1 || ips[0].String() != "1.1.1.2" {
		t.Fatalf("TopIPs = %v, want [1.1.1.2]", ips)
	}
}

func TestLeaderboardUpdatesLatestScore(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := db.SaveResult(ctx, testResult("1.1.1.1", 4*1024*1024)); err != nil {
		t.Fatal(err)
	}
	dead := &result.Result{
		IP:        net.ParseIP("1.1.1.1"),
		Port:      443,
		ProbeMode: "http",
		Latencies: []time.Duration{0, 0},
		Timestamp: time.Now(),
	}
	if err := db.SaveResult(ctx, dead); err != nil {
		t.Fatal(err)
	}

	ips, err := db.TopIPs(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(ips) != 0 {
		t.Fatalf("dead latest result should drop from warm-start list, got %v", ips)
	}

	recs, err := db.TopRecords(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 0 {
		t.Fatalf("dead latest result should drop from leaderboard, got %v", recs)
	}
}

func TestCheckpoints(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if _, ok, err := db.LoadCheckpoint(ctx, "sweep"); err != nil || ok {
		t.Fatalf("initial checkpoint ok=%v err=%v, want missing nil", ok, err)
	}
	if err := db.SaveCheckpoint(ctx, "sweep", 1234); err != nil {
		t.Fatal(err)
	}
	got, ok, err := db.LoadCheckpoint(ctx, "sweep")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || got != 1234 {
		t.Fatalf("checkpoint = %d ok=%v, want 1234 true", got, ok)
	}
}
