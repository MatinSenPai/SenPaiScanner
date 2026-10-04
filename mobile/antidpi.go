package mobile

import (
	"encoding/json"
	"strings"

	"github.com/matinsenpai/senpaiscanner/internal/antidpi"
	"github.com/matinsenpai/senpaiscanner/internal/ui"
)

// profileOf builds the Anti-DPI profile from the scan config; empty fields fall back to the published values.
func profileOf(c ScanConfig) antidpi.Profile {
	d := antidpi.DefaultProfile()
	p := antidpi.Profile{Enabled: c.AntiDpiEnabled, Finalmask: strings.TrimSpace(c.AdFinalmask), Fingerprint: strings.TrimSpace(c.AdFingerprint),
		ALPN: strings.TrimSpace(c.AdAlpn), CipherSuites: strings.TrimSpace(c.AdCiphers)}
	if p.Finalmask == "" {
		p.Finalmask = d.Finalmask
	}
	if p.Fingerprint == "" {
		p.Fingerprint = d.Fingerprint
	}
	if p.ALPN == "" {
		p.ALPN = d.ALPN
	}
	if p.CipherSuites == "" {
		p.CipherSuites = d.CipherSuites
	}
	return p
}

// applyAntiDpi validates the recipe and makes it the active one for the scan that is about to run.
func applyAntiDpi(c ScanConfig) error {
	p := profileOf(c)
	if err := p.Validate(); err != nil {
		return err
	}
	ui.SetAntiDPI(p)
	return nil
}

// AntiDpiDefaults returns the published recipe as JSON for the "Suggested values" button.
func AntiDpiDefaults() string {
	b, _ := json.Marshal(antidpi.DefaultProfile())
	return string(b)
}

// ValidateAntiDpi returns "" when the recipe in configJson is usable, else the reason.
func ValidateAntiDpi(configJson string) string {
	var c ScanConfig
	if err := json.Unmarshal([]byte(configJson), &c); err != nil {
		return err.Error()
	}
	if err := profileOf(c).Validate(); err != nil {
		return err.Error()
	}
	return ""
}
