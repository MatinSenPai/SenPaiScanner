package xraytest

import (
	"encoding/json"
	"net"
	"strconv"
	"testing"

	"github.com/matinsenpai/senpaiscanner/internal/antidpi"
)

func outboundOf(t *testing.T, cfg *VLESSConfig) (addr string, port float64, tls map[string]any) {
	t.Helper()
	b, err := BuildXrayConfig(cfg, 10808)
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Outbounds []struct {
			Settings struct {
				Vnext []struct {
					Address string
					Port    float64
				}
			}
			StreamSettings struct {
				TLSSettings map[string]any `json:"tlsSettings"`
			} `json:"streamSettings"`
		}
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	o := out.Outbounds[0]
	return o.Settings.Vnext[0].Address, o.Settings.Vnext[0].Port, o.StreamSettings.TLSSettings
}

func TestStartAntiDPIRedirectsToLoopbackAndAppliesTLS(t *testing.T) {
	cfg := &VLESSConfig{Protocol: "vless", UUID: "u", Encryption: "none", Address: "104.18.1.1", Port: 443,
		Network: "ws", Path: "/", Security: "tls", SNI: "a.workers.dev", Host: "a.workers.dev", Fingerprint: "chrome",
		AntiDPI: antidpi.DefaultProfile()}

	got, stop, err := StartAntiDPI(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer stop()
	addr, port, tls := outboundOf(t, got)
	if addr != "127.0.0.1" || int(port) == 443 {
		t.Errorf("xray must dial the loopback fragmenter, got %s:%v", addr, port)
	}
	if _, err := net.Dial("tcp", net.JoinHostPort(addr, strconv.Itoa(int(port)))); err != nil {
		t.Errorf("fragmenter is not listening: %v", err)
	}
	if tls["fingerprint"] != "unsafe" || tls["serverName"] != "a.workers.dev" {
		t.Errorf("tls settings wrong: %v", tls)
	}
	if cfg.Address != "104.18.1.1" || cfg.dialHost != "" {
		t.Error("the caller's config must not be modified")
	}
}

func TestStartAntiDPIOffKeepsConfig(t *testing.T) {
	cfg := &VLESSConfig{Protocol: "vless", Address: "104.18.1.1", Port: 443, Network: "tcp", Security: "none"}
	got, stop, err := StartAntiDPI(cfg)
	if err != nil || got != cfg {
		t.Fatalf("off must return cfg itself (err=%v)", err)
	}
	stop()
	if addr, port, _ := outboundOf(t, got); addr != "104.18.1.1" || port != 443 {
		t.Errorf("direct address lost: %s:%v", addr, port)
	}
}
