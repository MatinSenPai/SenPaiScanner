package ui

import (
	"sync/atomic"

	"github.com/matinsenpai/senpaiscanner/internal/antidpi"
)

// ponytail: process-wide profile — the TUI and the desktop GUI run one scan at a time. Pass it per scan if that changes.
var activeAntiDPI atomic.Pointer[antidpi.Profile]

// CurrentAntiDPI is the profile every probe and xray validation uses (default: the published values, enabled).
func CurrentAntiDPI() antidpi.Profile {
	if p := activeAntiDPI.Load(); p != nil {
		return *p
	}
	return antidpi.DefaultProfile()
}

// SetAntiDPI replaces the active profile; an empty one (config files written before Anti-DPI existed) means "defaults".
func SetAntiDPI(p antidpi.Profile) {
	if p.Finalmask == "" && p.Fingerprint == "" && p.ALPN == "" && p.CipherSuites == "" {
		p = antidpi.DefaultProfile()
	}
	activeAntiDPI.Store(&p)
}

// ToggleAntiDPI flips the active profile on/off, keeping its values.
func ToggleAntiDPI() bool {
	p := CurrentAntiDPI()
	p.Enabled = !p.Enabled
	SetAntiDPI(p)
	return p.Enabled
}
