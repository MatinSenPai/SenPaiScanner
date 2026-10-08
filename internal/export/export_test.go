package export

import (
	"strings"
	"testing"

	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

func TestGenerateSubscription(t *testing.T) {
	cfg, err := xraytest.ParseProxyURL("vless://12345678-1234-1234-1234-123456789abc@template.example.com:443?encryption=none&security=tls&sni=cdn.example.com&type=ws&host=cdn.example.com&path=%2Fws#CF")
	if err != nil {
		t.Fatal(err)
	}

	endpoints := []Endpoint{
		{IP: "1.1.1.1", Port: 443},
		{IP: "2.2.2.2", Port: 443},
		{IP: "3.3.3.3"},
		{IP: "2606:4700::1", Port: 443},
	}
	b, err := Generate(cfg, endpoints)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.ShareURLs) != 4 {
		t.Fatalf("got %d URLs, want 4", len(b.ShareURLs))
	}
	for _, u := range b.ShareURLs {
		if !strings.HasPrefix(u, "vless://") {
			t.Errorf("unexpected URL %q", u)
		}
	}
	if !strings.Contains(b.Subscription, "1.1.1.1:443") {
		t.Error("subscription missing endpoint IP")
	}
	if !strings.Contains(b.Subscription, "3.3.3.3:443") {
		t.Error("endpoint without port should fall back to template port 443")
	}
	if !strings.Contains(b.Subscription, "[2606:4700::1]:443") {
		t.Error("subscription missing bracketed IPv6 endpoint")
	}
	if !strings.Contains(b.SingBox, `"server": "2606:4700::1"`) {
		t.Error("sing-box missing clean IPv6 server address")
	}
	if !strings.Contains(b.Clash, "server: 2606:4700::1") {
		t.Error("clash missing clean IPv6 server address")
	}
	if b.SingBox == "" {
		t.Error("sing-box output empty")
	}
	if b.Clash == "" {
		t.Error("clash output empty")
	}
}

func TestGenerateNoEndpoints(t *testing.T) {
	cfg, err := xraytest.ParseProxyURL("vless://12345678-1234-1234-1234-123456789abc@example.com:443")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(cfg, nil); err == nil {
		t.Fatal("expected error with no endpoints")
	}
}

func TestParseEndpoints(t *testing.T) {
	eps := ParseEndpoints([]string{
		"1.1.1.1:443",
		"2.2.2.2",
		"",
		" 3.3.3.3:8443 ",
		"[2606:4700::1]:443",
		"2606:4700::2",
		"[2606:4700::3]",
	})
	if len(eps) != 6 {
		t.Fatalf("got %d endpoints, want 6", len(eps))
	}
	if eps[0].IP != "1.1.1.1" || eps[0].Port != 443 {
		t.Errorf("bad first endpoint: %+v", eps[0])
	}
	if eps[1].IP != "2.2.2.2" || eps[1].Port != 0 {
		t.Errorf("bad second endpoint: %+v", eps[1])
	}
	if eps[2].IP != "3.3.3.3" || eps[2].Port != 8443 {
		t.Errorf("bad third endpoint: %+v", eps[2])
	}
	if eps[3].IP != "2606:4700::1" || eps[3].Port != 443 {
		t.Errorf("bad fourth endpoint: %+v", eps[3])
	}
	if eps[4].IP != "2606:4700::2" || eps[4].Port != 0 {
		t.Errorf("bad fifth endpoint: %+v", eps[4])
	}
	if eps[5].IP != "2606:4700::3" || eps[5].Port != 0 {
		t.Errorf("bad sixth endpoint: %+v", eps[5])
	}
}
