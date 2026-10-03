// Package antidpi fragments the first TLS flight of a connection so that
// deep-packet-inspection boxes cannot match the ClientHello (SNI) in one packet.
//
// The recipe and its defaults are the published Iran values from
// https://t.me/MatinSenPaii/5469. They change from time to time, so every
// field of Profile is user-editable. The "finalmask" JSON follows the PattN
// fork of xray-core: stock xray only knows one length/delay pair per mask, the
// recipe needs per-segment lists, so the algorithm runs here instead.
package antidpi

import (
	"context"
	"crypto/tls"
	"net"
	"strings"
)

// Profile is the editable Anti-DPI recipe.
type Profile struct {
	Enabled      bool   `json:"enabled"`
	Finalmask    string `json:"finalmask"`    // {"tcp":[{"type":"fragment","settings":{...}}]}
	Fingerprint  string `json:"fingerprint"`  // xray uTLS fingerprint; "unsafe" = native Go TLS
	ALPN         string `json:"alpn"`         // comma separated
	CipherSuites string `json:"cipherSuites"` // colon separated TLS 1.2 suite names
}

// DefaultProfile returns the values from the Telegram post.
func DefaultProfile() Profile {
	return Profile{
		Enabled:     true,
		Finalmask:   `{"tcp": [{"type": "fragment", "settings": {"packets": "tlshello", "lengths": ["0", "104", "1"], "delays": ["0"], "maxSplit": "0"}},{"type": "fragment", "settings": {"packets": "1-1", "lengths": ["114", "1"], "delays": ["1"], "maxSplit": "11"}}]}`,
		Fingerprint: "unsafe",
		ALPN:        "http/1.1",
		CipherSuites: "TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_128_GCM_SHA256:TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:" +
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:" +
			"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256:TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:" +
			"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256:TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
	}
}

// ALPNList splits the comma separated ALPN field.
func (p Profile) ALPNList() []string {
	var r []string
	for _, x := range strings.Split(p.ALPN, ",") {
		if x = strings.TrimSpace(x); x != "" {
			r = append(r, x)
		}
	}
	return r
}

// Validate reports whether the profile can be used (bad finalmask JSON, unsupported mask type, ...).
func (p Profile) Validate() error {
	if !p.Enabled {
		return nil
	}
	_, err := parseFinalmask(p.Finalmask)
	return err
}

// state is what a context carries: the compiled masks plus the TLS tweaks.
type state struct {
	masks  []fragCfg
	alpn   []string
	suites []uint16
}

type ctxKey struct{}

// With returns a context whose probes use p. A disabled profile returns ctx unchanged.
func With(ctx context.Context, p Profile) (context.Context, error) {
	if !p.Enabled {
		return ctx, nil
	}
	masks, err := parseFinalmask(p.Finalmask)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, ctxKey{}, &state{masks: masks, alpn: p.ALPNList(), suites: suiteIDs(p.CipherSuites)}), nil
}

func from(ctx context.Context) *state {
	s, _ := ctx.Value(ctxKey{}).(*state)
	return s
}

// Dial dials like d.DialContext and fragments the writes when ctx carries a profile.
func Dial(ctx context.Context, d *net.Dialer, network, addr string) (net.Conn, error) {
	c, err := d.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	if s := from(ctx); s != nil && len(s.masks) > 0 {
		return &fragConn{Conn: c, w: chainWriters(s.masks, c)}, nil
	}
	return c, nil
}

// TLS applies the profile's ALPN and cipher suites to cfg (no-op without a profile) and returns cfg.
func TLS(ctx context.Context, cfg *tls.Config) *tls.Config {
	if s := from(ctx); s != nil {
		if len(s.alpn) > 0 {
			cfg.NextProtos = s.alpn
		}
		if len(s.suites) > 0 {
			cfg.CipherSuites = s.suites // only TLS 1.2 suites are configurable in Go; TLS 1.3 ones are ignored
		}
	}
	return cfg
}

// suiteIDs maps suite names to Go IDs; unknown names (and TLS 1.3 suites Go does not let you choose) are skipped.
func suiteIDs(list string) []uint16 {
	byName := map[string]uint16{}
	for _, s := range tls.CipherSuites() {
		byName[s.Name] = s.ID
	}
	for _, s := range tls.InsecureCipherSuites() {
		byName[s.Name] = s.ID
	}
	var ids []uint16
	for _, n := range strings.Split(list, ":") {
		if id, ok := byName[strings.TrimSpace(n)]; ok && !strings.HasPrefix(n, "TLS_AES_") && !strings.HasPrefix(n, "TLS_CHACHA20_") {
			ids = append(ids, id)
		}
	}
	return ids
}

// fragConn routes writes through the mask chain.
type fragConn struct {
	net.Conn
	w interface{ Write([]byte) (int, error) }
}

func (c *fragConn) Write(p []byte) (int, error) { return c.w.Write(p) }
