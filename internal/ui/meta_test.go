package ui

import (
	"net"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMetaHelpersNormalizeProviderData(t *testing.T) {
	if got := parseASN("AS15169 Google LLC"); got != 15169 {
		t.Fatalf("parseASN = %d, want 15169", got)
	}
	if got := cleanASOrganization("AS15169 Google LLC"); got != "Google LLC" {
		t.Fatalf("cleanASOrganization = %q", got)
	}
	merged := mergeMeta(
		MetaMsg{IP: "203.0.113.8", Colo: "FRA", Source: "Cloudflare"},
		MetaMsg{ASN: 64500, ASOrganization: "Example ISP", Country: "DE", Source: "IPWhois"},
	)
	if merged.ASN != 64500 || merged.ASOrganization != "Example ISP" || merged.Country != "DE" {
		t.Fatalf("mergeMeta = %+v", merged)
	}
	if !strings.Contains(merged.Source, "Cloudflare") || !strings.Contains(merged.Source, "IPWhois") {
		t.Fatalf("merged source = %q", merged.Source)
	}
}

func TestCymruOriginNames(t *testing.T) {
	if got := cymruOriginName(net.ParseIP("8.8.4.4")); got != "4.4.8.8.origin.asn.cymru.com" {
		t.Fatalf("IPv4 Cymru name = %q", got)
	}
	if got := cymruOriginName(net.ParseIP("2001:4860:4860::8888")); !strings.HasSuffix(got, ".origin6.asn.cymru.com") {
		t.Fatalf("IPv6 Cymru name = %q", got)
	}
}

func TestNeighborScanIsOptInAndPassedToPhase1(t *testing.T) {
	m := newTestApp(t)
	if m.configNeighborScan {
		t.Fatal("neighbor scan should be disabled by default")
	}
	m.page = PageScanWithConfig
	m.configSetupRow = 6
	updated, _ := m.handleScanWithConfigKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m2 := updated.(AppModel)
	if !m2.configNeighborScan {
		t.Fatal("space did not enable neighbor scan")
	}
	if !m2.resolvePhase1Options().neighborScan {
		t.Fatal("neighbor choice was not passed to Phase 1 options")
	}
}
