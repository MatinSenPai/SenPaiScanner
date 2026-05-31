package ui

import (
	"testing"
	"time"
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
