package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matinsenpai/senpaiscanner/internal/store"
)

func TestNormalizeRequestDefaults(t *testing.T) {
	req := normalizeRequest(scanRequest{})
	if req.Count != 5000 || req.Concurrency != 100 || req.Tries != 4 || req.Port != 443 {
		t.Fatalf("unexpected numeric defaults: %+v", req)
	}
	if req.Mode != "http" || req.Timeout != "3s" || !req.UseV4 || req.UseV6 {
		t.Fatalf("unexpected mode/family defaults: %+v", req)
	}
}

func TestNormalizeRequestClampsHostileInput(t *testing.T) {
	req := normalizeRequest(scanRequest{
		Count:       1 << 30,
		Concurrency: 100_000_000,
		Tries:       1_000_000,
		Port:        99999,
	})
	if req.Count != maxCount {
		t.Fatalf("count = %d, want clamp to %d", req.Count, maxCount)
	}
	if req.Concurrency != maxConcurrency {
		t.Fatalf("concurrency = %d, want clamp to %d", req.Concurrency, maxConcurrency)
	}
	if req.Tries != maxTries {
		t.Fatalf("tries = %d, want clamp to %d", req.Tries, maxTries)
	}
	if req.Port != 443 {
		t.Fatalf("out-of-range port = %d, want reset to 443", req.Port)
	}
}

func TestVersionEndpoint(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	s := &Server{version: "test", leaderboard: db}
	req := httptest.NewRequest(http.MethodGet, "/api/version", nil)
	rr := httptest.NewRecorder()
	s.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if got := rr.Body.String(); got != "{\"version\":\"test\"}\n" {
		t.Fatalf("body = %q", got)
	}
}

func TestPortsEndpoint(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	s := &Server{leaderboard: db}
	req := httptest.NewRequest(http.MethodGet, "/api/ports", nil)
	rr := httptest.NewRecorder()
	s.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var got map[string][]int
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got["https"]) == 0 || got["https"][0] != 443 {
		t.Fatalf("https ports = %v, want 443 first", got["https"])
	}
	if len(got["http"]) == 0 || got["http"][0] != 80 {
		t.Fatalf("http ports = %v, want 80 first", got["http"])
	}
}

func TestWarmStartSkipsCIDR(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	s := &Server{leaderboard: db}
	if got := s.warmStart(context.Background(), scanRequest{CIDR: "1.1.1.0/24", UseV4: true}); len(got) != 0 {
		t.Fatalf("warmStart with CIDR = %v, want empty", got)
	}
}
