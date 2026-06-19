package xraytest

import (
	"fmt"
	"strings"
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

func TestParseVLESS_WorkersDevMixedSNIHost(t *testing.T) {
	raw := "vless://3a24f4fb-c574-43f4-8041-2c1a381779af@188.114.97.3:443?encryption=none&security=tls&sni=MmM.matin3sALEaZIRAn.WoRKerS.DEv&fp=chrome&alpn=http%2F1.1&insecure=0&allowInsecure=0&type=ws&host=mmm.matin3saleaziran.workers.dev&path=%2FeyJqdW5rIjoidmlIaTNjaHNmWE1pIiwicHJvdG9jb2wiOiJ2bCIsIm1vZGUiOiJwcmVmaXgiLCJwYW5lbElQcyI6WyJbMjYwMjpmYzU5OmIwOjY0OjpdIl19%3Fed%3D2560#test"

	cfg, err := ParseVLESS(raw)
	if err != nil {
		t.Fatalf("ParseVLESS failed: %v", err)
	}
	if cfg.SNI != "MmM.matin3sALEaZIRAn.WoRKerS.DEv" {
		t.Fatalf("SNI = %q", cfg.SNI)
	}
	if cfg.Host != "mmm.matin3saleaziran.workers.dev" {
		t.Fatalf("Host = %q", cfg.Host)
	}
	if cfg.ALPN[0] != "http/1.1" {
		t.Fatalf("ALPN = %v", cfg.ALPN)
	}
	if !strings.HasPrefix(cfg.Path, "/eyJ") {
		t.Fatalf("Path = %q", cfg.Path)
	}

	swapped := cfg.WithEndpoint("188.114.97.3", 443)
	b, err := BuildXrayConfig(swapped, 20002)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"allowInsecure"`) {
		t.Fatalf("allowInsecure must not appear in xray config, got %s", string(b))
	}
	if !strings.Contains(string(b), `"serverName": "MmM.matin3sALEaZIRAn.WoRKerS.DEv"`) {
		t.Fatalf("expected serverName in TLS settings, got %s", string(b))
	}
	if !strings.Contains(string(b), `"verifyPeerCertByName"`) {
		t.Fatalf("expected verifyPeerCertByName for IP endpoint, got %s", string(b))
	}
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

func TestParseVMess_WS(t *testing.T) {
	raw := "vmess://eyJ2IjoiMiIsInBzIjoiQ0YtVk1lc3MtVGVzdCIsImFkZCI6ImV4YW1wbGUuY29tIiwicG9ydCI6NDQzLCJpZCI6IjEyMzQ1Njc4LTEyMzQtMTIzNC0xMjM0LTEyMzQ1Njc4OWFiYyIsImFpZCI6MCwic2N5IjoiYXV0byIsIm5ldCI6IndzIiwidHlwZSI6Im5vbmUiLCJob3N0IjoiZXhhbXBsZS5jb20iLCJwYXRoIjoiL2Rvd25sb2FkIiwidGxzIjoidGxzIiwic25pIjoiZXhhbXBsZS5jb20ifQ=="

	cfg, err := ParseProxyURL(raw)
	if err != nil {
		t.Fatalf("ParseProxyURL failed: %v", err)
	}

	assertEqual(t, "Protocol", cfg.Protocol, "vmess")
	assertEqual(t, "UUID", cfg.UUID, "12345678-1234-1234-1234-123456789abc")
	assertEqual(t, "Address", cfg.Address, "example.com")
	assertEqual(t, "Port", itoa(cfg.Port), "443")
	assertEqual(t, "Network", cfg.Network, "ws")
	assertEqual(t, "Security", cfg.Security, "tls")
	assertEqual(t, "SNI", cfg.SNI, "example.com")
	assertEqual(t, "Path", cfg.Path, "/download")
	assertEqual(t, "Host", cfg.Host, "example.com")
	assertEqual(t, "Remark", cfg.Remark, "CF-VMess-Test")
}

func TestParseShadowsocks_Base64(t *testing.T) {
	// ss://base64(aes-256-gcm:password123)@198.51.100.1:8388#SS-Test
	raw := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQxMjM=@198.51.100.1:8388#SS-Test"

	cfg, err := ParseShadowsocks(raw)
	if err != nil {
		t.Fatalf("ParseShadowsocks failed: %v", err)
	}

	assertEqual(t, "Protocol", cfg.Protocol, "shadowsocks")
	assertEqual(t, "Method", cfg.Method, "aes-256-gcm")
	assertEqual(t, "Password", cfg.Password, "password123")
	assertEqual(t, "Address", cfg.Address, "198.51.100.1")
	assertEqual(t, "Port", itoa(cfg.Port), "8388")
	assertEqual(t, "Remark", cfg.Remark, "SS-Test")
}

func TestParseShadowsocks_Plaintext(t *testing.T) {
	// ss://chacha20-ietf-poly1305:secret@10.0.0.1:443#My-SS
	raw := "ss://chacha20-ietf-poly1305:secret@10.0.0.1:443#My-SS"

	cfg, err := ParseShadowsocks(raw)
	if err != nil {
		t.Fatalf("ParseShadowsocks failed: %v", err)
	}

	assertEqual(t, "Protocol", cfg.Protocol, "shadowsocks")
	assertEqual(t, "Method", cfg.Method, "chacha20-ietf-poly1305")
	assertEqual(t, "Password", cfg.Password, "secret")
	assertEqual(t, "Address", cfg.Address, "10.0.0.1")
	assertEqual(t, "Port", itoa(cfg.Port), "443")
	assertEqual(t, "Remark", cfg.Remark, "My-SS")
}

func TestParseShadowsocks_viaParseProxyURL(t *testing.T) {
	raw := "ss://YWVzLTEyOC1nY206dGVzdA=@192.168.1.1:1080#auto"
	cfg, err := ParseProxyURL(raw)
	if err != nil {
		t.Fatalf("ParseProxyURL failed for ss://: %v", err)
	}
	assertEqual(t, "Protocol", cfg.Protocol, "shadowsocks")
	assertEqual(t, "Method", cfg.Method, "aes-128-gcm")
	assertEqual(t, "Password", cfg.Password, "test")
}

func TestParseShadowsocks_RoundTrip(t *testing.T) {
	raw := "ss://YWVzLTI1Ni1nY206cGFzc3dvcmQxMjM=@198.51.100.1:8388#SS-Test"
	cfg, err := ParseShadowsocks(raw)
	if err != nil {
		t.Fatalf("ParseShadowsocks failed: %v", err)
	}
	roundTrip := cfg.ToShareURL()
	if !strings.HasPrefix(roundTrip, "ss://") {
		t.Errorf("ToShareURL should start with ss://, got %q", roundTrip)
	}
	if !strings.Contains(roundTrip, "198.51.100.1:8388") {
		t.Errorf("ToShareURL should contain host:port, got %q", roundTrip)
	}
}

func TestBuildShadowsocksConfig(t *testing.T) {
	cfg := &VLESSConfig{
		Protocol: "shadowsocks",
		Method:   "aes-256-gcm",
		Password: "testpass",
		Address:  "1.2.3.4",
		Port:     8388,
	}
	b, err := BuildXrayConfig(cfg, 1080)
	if err != nil {
		t.Fatalf("BuildXrayConfig failed: %v", err)
	}
	json := string(b)
	if !strings.Contains(json, `"protocol": "shadowsocks"`) {
		t.Error("expected shadowsocks protocol in config")
	}
	if !strings.Contains(json, `"method": "aes-256-gcm"`) {
		t.Error("expected method in config")
	}
	if !strings.Contains(json, `"password": "testpass"`) {
		t.Error("expected password in config")
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
