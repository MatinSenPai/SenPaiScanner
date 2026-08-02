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
	}
	b, err := Generate(cfg, endpoints)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.ShareURLs) != 3 {
		t.Fatalf("got %d URLs, want 3", len(b.ShareURLs))
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
	eps := ParseEndpoints([]string{"1.1.1.1:443", "2.2.2.2", "", " 3.3.3.3:8443 "})
	if len(eps) != 3 {
		t.Fatalf("got %d endpoints, want 3", len(eps))
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
}
