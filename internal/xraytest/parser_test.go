package xraytest

import (
	"fmt"
	"testing"
)

func TestParseVLESS_WS(t *testing.T) {
	raw := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&sni=example.com&fp=chrome&alpn=h2%2Chttp%2F1.1&insecure=1&allowInsecure=1&type=ws&host=example.com&path=%2Fdownload#CF-WS-079xe1rr"

	cfg, err := ParseVLESS(raw)
	if err != nil {
		t.Fatalf("ParseVLESS failed: %v", err)
	}

	assertEqual(t, "UUID", cfg.UUID, "12345678-1234-1234-1234-123456789abc")
	assertEqual(t, "Address", cfg.Address, "example.com")
	assertEqual(t, "Port", itoa(cfg.Port), "443")
	assertEqual(t, "Network", cfg.Network, "ws")
	assertEqual(t, "Security", cfg.Security, "tls")
	assertEqual(t, "SNI", cfg.SNI, "example.com")
	assertEqual(t, "Fingerprint", cfg.Fingerprint, "chrome")
	assertEqual(t, "Path", cfg.Path, "/download")
	assertEqual(t, "Host", cfg.Host, "example.com")
	assertEqual(t, "Remark", cfg.Remark, "CF-WS-079xe1rr")

	if !cfg.Insecure {
		t.Error("expected Insecure=true")
	}
	if len(cfg.ALPN) != 2 || cfg.ALPN[0] != "h2" || cfg.ALPN[1] != "http/1.1" {
		t.Errorf("unexpected ALPN: %v", cfg.ALPN)
	}
}

func TestParseVLESS_GRPC(t *testing.T) {
	raw := "vless://87654321-4321-4321-4321-cba987654321@example.com:8443?encryption=none&security=tls&sni=example.com&fp=chrome&alpn=h2&insecure=1&allowInsecure=1&type=grpc&authority=example.com&serviceName=download&mode=multi#CF-GRPC-f8k8s2jp"

	cfg, err := ParseVLESS(raw)
	if err != nil {
		t.Fatalf("ParseVLESS failed: %v", err)
	}

	assertEqual(t, "UUID", cfg.UUID, "87654321-4321-4321-4321-cba987654321")
	assertEqual(t, "Port", itoa(cfg.Port), "8443")
	assertEqual(t, "Network", cfg.Network, "grpc")
	assertEqual(t, "ServiceName", cfg.ServiceName, "download")
	assertEqual(t, "Authority", cfg.Authority, "example.com")
	assertEqual(t, "Mode", cfg.Mode, "multi")
}

func TestParseVLESS_XHTTP(t *testing.T) {
	raw := "vless://abcdef12-3456-7890-abcd-ef1234567890@test.example.org:2053?encryption=none&security=tls&sni=test.example.org&fp=chrome&alpn=h2%2Chttp%2F1.1&insecure=1&allowInsecure=1&type=xhttp&host=test.example.org&path=%2Fdownload&mode=auto#CF-XHTTP-o9xk21gf"

	cfg, err := ParseVLESS(raw)
	if err != nil {
		t.Fatalf("ParseVLESS failed: %v", err)
	}

	assertEqual(t, "UUID", cfg.UUID, "abcdef12-3456-7890-abcd-ef1234567890")
	assertEqual(t, "Port", itoa(cfg.Port), "2053")
	assertEqual(t, "Network", cfg.Network, "xhttp")
	assertEqual(t, "Path", cfg.Path, "/download")
	assertEqual(t, "Host", cfg.Host, "test.example.org")
	assertEqual(t, "Mode", cfg.Mode, "auto")
}

func TestWithAddress(t *testing.T) {
	raw := "vless://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&sni=example.com&type=ws&path=%2Fdownload&host=example.com#test"

	cfg, err := ParseVLESS(raw)
	if err != nil {
		t.Fatalf("ParseVLESS failed: %v", err)
	}

	swapped := cfg.WithAddress("172.66.40.1")

	assertEqual(t, "original address", cfg.Address, "example.com")
	assertEqual(t, "swapped address", swapped.Address, "172.66.40.1")
	assertEqual(t, "port preserved", itoa(swapped.Port), "443")
	assertEqual(t, "SNI preserved", swapped.SNI, "example.com")
	assertEqual(t, "Host preserved", swapped.Host, "example.com")
}

func TestParseVLESS_Invalid(t *testing.T) {
	cases := []string{
		"",
		"vmess://something",
		"vless://no-at-sign",
		"vless://uuid@host-no-port",
		"vless://12345678-1234-1234-1234-123456789abc@example.com:0?encryption=none",
		"vless://12345678-1234-1234-1234-123456789abc@example.com:65536?encryption=none",
	}
	for _, c := range cases {
		_, err := ParseVLESS(c)
		if err == nil {
			t.Errorf("expected error for %q, got nil", c)
		}
	}
}

func TestParseProxyURLAcceptsCaseInsensitiveScheme(t *testing.T) {
	raw := "VLESS://12345678-1234-1234-1234-123456789abc@example.com:443?encryption=none&security=tls&type=ws#test"

	cfg, err := ParseProxyURL(raw)
	if err != nil {
		t.Fatalf("ParseProxyURL failed: %v", err)
	}
	if cfg.Protocol != "vless" {
		t.Fatalf("Protocol = %q, want vless", cfg.Protocol)
	}
	if cfg.Port != 443 {
		t.Fatalf("Port = %d, want 443", cfg.Port)
	}
}

func TestParseTrojanRejectsInvalidPortRange(t *testing.T) {
	for _, raw := range []string{
		"trojan://password@example.com:0?security=tls",
		"trojan://password@example.com:65536?security=tls",
	} {
		if _, err := ParseTrojan(raw); err == nil {
			t.Fatalf("expected error for %q, got nil", raw)
		}
	}
}

func assertEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func TestParseShadowsocksSIP002(t *testing.T) {
	// SIP002 format: ss://BASE64(method:password)@host:port#remark
	// aes-256-gcm:hunter2 → base64 = YWVzLTI1Ni1nY206aHVudGVyMg==
	raw := "ss://YWVzLTI1Ni1nY206aHVudGVyMg==@192.0.2.1:8388#My%20Server"
	cfg, err := ParseShadowsocks(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Protocol", cfg.Protocol, "shadowsocks")
	assertEqual(t, "Method", cfg.Method, "aes-256-gcm")
	assertEqual(t, "Password", cfg.Password, "hunter2")
	assertEqual(t, "Address", cfg.Address, "192.0.2.1")
	assertEqual(t, "Port", itoa(cfg.Port), "8388")
	assertEqual(t, "Remark", cfg.Remark, "My Server")
	assertEqual(t, "Network", cfg.Network, "tcp")
}

func TestParseShadowsocksLegacy(t *testing.T) {
	// Legacy format: ss://BASE64(method:password@host:port)#remark
	// chacha20-ietf-poly1305:s3cr3t@10.0.0.1:1080
	encoded := "Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTpzM2NyM3RAMTAuMC4wLjE6MTA4MA=="
	raw := "ss://" + encoded + "#legacy"
	cfg, err := ParseShadowsocks(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Protocol", cfg.Protocol, "shadowsocks")
	assertEqual(t, "Method", cfg.Method, "chacha20-ietf-poly1305")
	assertEqual(t, "Password", cfg.Password, "s3cr3t")
	assertEqual(t, "Address", cfg.Address, "10.0.0.1")
	assertEqual(t, "Port", itoa(cfg.Port), "1080")
	assertEqual(t, "Remark", cfg.Remark, "legacy")
}

func TestParseShadowsocksViaDispatch(t *testing.T) {
	raw := "ss://Y2hhY2hhMjAtaWV0Zi1wb2x5MTMwNTpteXBhc3NAZXhhbXBsZS5jb206NDQz#test"
	cfg, err := ParseProxyURL(raw)
	if err != nil {
		t.Fatalf("dispatch error: %v", err)
	}
	if cfg.Protocol != "shadowsocks" {
		t.Errorf("Protocol: got %q, want shadowsocks", cfg.Protocol)
	}
}

func TestParseShadowsocksToShareURL(t *testing.T) {
	raw := "ss://YWVzLTI1Ni1nY206aHVudGVyMg==@192.0.2.1:8388#test"
	cfg, err := ParseShadowsocks(raw)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	reconstructed := cfg.ToShadowsocksURL()
	cfg2, err := ParseShadowsocks(reconstructed)
	if err != nil {
		t.Fatalf("round-trip parse error: %v", err)
	}
	assertEqual(t, "Method", cfg2.Method, cfg.Method)
	assertEqual(t, "Password", cfg2.Password, cfg.Password)
	assertEqual(t, "Address", cfg2.Address, cfg.Address)
	assertEqual(t, "Port", itoa(cfg2.Port), itoa(cfg.Port))
}

func TestParseShadowsocksInvalidPort(t *testing.T) {
	for _, raw := range []string{
		"ss://YWVzLTI1Ni1nY206dGVzdA==@example.com:0",
		"ss://YWVzLTI1Ni1nY206dGVzdA==@example.com:99999",
	} {
		if _, err := ParseShadowsocks(raw); err == nil {
			t.Fatalf("expected error for %q, got nil", raw)
		}
	}
}
