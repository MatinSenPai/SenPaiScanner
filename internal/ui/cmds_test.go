package ui

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matinsenpai/senpaiscanner/internal/prober"
	"github.com/matinsenpai/senpaiscanner/internal/result"
)

func TestConfigProbeFromURLUsesConfigPortSNIAndWebSocket(t *testing.T) {
	raw := "vless://3441b906-471f-4160-8f2c-a981793e6155@104.17.89.5:2087?encryption=none&security=tls&sni=winter-thunder-0638.matinsenpaivideo2.workers.dev&fp=chrome&insecure=0&allowInsecure=0&type=ws&host=winter-thunder-0638.matinsenpaivideo2.workers.dev&path=%2F#CF"

	cfg, err := configProbeFromURL(raw, 7*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Port != 2087 {
		t.Fatalf("port = %d, want 2087", cfg.Port)
	}
	if cfg.SNI != "winter-thunder-0638.matinsenpaivideo2.workers.dev" {
		t.Fatalf("SNI = %q", cfg.SNI)
	}
	if cfg.WebSocketHost != "winter-thunder-0638.matinsenpaivideo2.workers.dev" {
		t.Fatalf("WebSocketHost = %q", cfg.WebSocketHost)
	}
	if cfg.WebSocketPath != "/" {
		t.Fatalf("WebSocketPath = %q, want /", cfg.WebSocketPath)
	}
	if !cfg.RequireWebSocket {
		t.Fatal("RequireWebSocket = false, want true")
	}
}

func TestRunConfigPortProbesCompletesWhenNeighborsFillQueue(t *testing.T) {
	_, ipNet, err := net.ParseCIDR("192.0.2.0/24")
	if err != nil {
		t.Fatal(err)
	}

	ips := make(chan net.IP, 1)
	ips <- net.ParseIP("192.0.2.32")
	close(ips)

	var callbacks atomic.Int64
	done := make(chan struct{})
	go func() {
		defer close(done)
		runConfigPortProbesWithProbe(
			context.Background(),
			ips,
			[]int{443},
			2,
			prober.Config{Port: 443, Mode: prober.ModeTCP},
			func(*result.Result) {
				callbacks.Add(1)
			},
			neighborScanOpts{
				enabled:  true,
				nets:     []*net.IPNet{ipNet},
				radius:   64,
				perHit:   64,
				maxTotal: 64,
			},
			func(_ context.Context, ip net.IP, cfg prober.Config) *result.Result {
				return &result.Result{
					IP:        ip,
					Port:      cfg.Port,
					ProbeMode: cfg.Mode.String(),
					Latencies: []time.Duration{time.Millisecond},
					Timestamp: time.Now(),
				}
			},
		)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("runConfigPortProbes did not finish after queuing neighbor probes")
	}

	if got := callbacks.Load(); got != 65 {
		t.Fatalf("callbacks = %d, want 65 (1 seed + 64 neighbors)", got)
	}
}
