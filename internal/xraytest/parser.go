package xraytest

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// VLESSConfig holds parsed parameters from a VLESS or Trojan share URL.
// Check the Protocol field to know which type this is.
type VLESSConfig struct {
	// Protocol is "vless" or "trojan".
	Protocol string

	// VLESS-specific
	UUID       string
	Encryption string
	Flow       string

	// Trojan-specific
	Password string

	// Shadowsocks-specific
	Method string // cipher, e.g. "aes-256-gcm", "chacha20-ietf-poly1305"

	// Common
	Address string
	Port    int

	// Transport
	Network     string // ws, grpc, xhttp, tcp
	Path        string
	Host        string
	ServiceName string // gRPC
	Mode        string // gRPC multi/gun, xhttp auto
	Authority   string // gRPC

	// TLS
	Security    string // tls, reality, none
	SNI         string
	Fingerprint string
	ALPN        []string
	Insecure    bool

	// Metadata
	Remark string
}

// ParseProxyURL auto-detects the protocol (vless://, trojan://, or ss://) and parses
// the share URL into a VLESSConfig. Returns an error if the scheme is unknown.
func ParseProxyURL(raw string) (*VLESSConfig, error) {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)
	switch {
	case strings.HasPrefix(lower, "vless://"):
		return ParseVLESS(raw)
	case strings.HasPrefix(lower, "trojan://"):
		return ParseTrojan(raw)
	case strings.HasPrefix(lower, "ss://"):
		return ParseShadowsocks(raw)
	default:
		return nil, fmt.Errorf("unsupported URL scheme — must start with vless://, trojan://, or ss://")
	}
}

// ParseVLESS parses a vless:// share URL into a VLESSConfig.
func ParseVLESS(raw string) (*VLESSConfig, error) {
	if !hasScheme(raw, "vless") {
		return nil, fmt.Errorf("not a vless:// URL")
	}

	// vless://UUID@address:port?params#remark
	// URL parse doesn't handle the UUID as userinfo well, so we do it manually
	raw = stripScheme(raw, "vless")

	// Split remark
	remark := ""
	if idx := strings.LastIndex(raw, "#"); idx != -1 {
		remark = raw[idx+1:]
		raw = raw[:idx]
	}
	remark, _ = url.QueryUnescape(remark)

	// Split params
	params := url.Values{}
	if idx := strings.Index(raw, "?"); idx != -1 {
		var err error
		params, err = url.ParseQuery(raw[idx+1:])
		if err != nil {
			return nil, fmt.Errorf("parsing query params: %w", err)
		}
		raw = raw[:idx]
	}

	// Split UUID@address:port
	atIdx := strings.Index(raw, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("missing @ in URL")
	}
	uuid := raw[:atIdx]
	hostPort := raw[atIdx+1:]

	// Parse host:port
	host, portStr, err := splitHostPort(hostPort)
	if err != nil {
		return nil, fmt.Errorf("parsing host:port %q: %w", hostPort, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		// The '?' separator may have been silently dropped by some paste
		// handlers. Recover: extract leading digits as port and treat the
		// remainder as additional query params.
		port, params, err = recoverMissingQuerySep(portStr, params)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q", portStr)
		}
	}
	if err := validatePort(port); err != nil {
		return nil, err
	}

	cfg := &VLESSConfig{
		Protocol:    "vless",
		UUID:        uuid,
		Address:     host,
		Port:        port,
		Encryption:  paramOr(params, "encryption", "none"),
		Flow:        params.Get("flow"),
		Network:     paramOr(params, "type", "tcp"),
		Security:    paramOr(params, "security", "none"),
		SNI:         params.Get("sni"),
		Fingerprint: paramOr(params, "fp", ""),
		Insecure:    params.Get("insecure") == "1" || params.Get("allowInsecure") == "1",
		Remark:      remark,
	}

	// Transport-specific
	switch cfg.Network {
	case "ws":
		cfg.Path = paramOr(params, "path", "/")
		cfg.Host = paramOr(params, "host", cfg.SNI)
	case "grpc":
		cfg.ServiceName = params.Get("serviceName")
		cfg.Authority = params.Get("authority")
		cfg.Mode = paramOr(params, "mode", "gun")
	case "xhttp", "splithttp":
		cfg.Path = paramOr(params, "path", "/")
		cfg.Host = paramOr(params, "host", cfg.SNI)
		cfg.Mode = paramOr(params, "mode", "auto")
	}

	// ALPN
	if alpnStr := params.Get("alpn"); alpnStr != "" {
		cfg.ALPN = strings.Split(alpnStr, ",")
	}

	return cfg, nil
}

// WithAddress returns a copy of the config with the address replaced.
// Port is preserved. This is used to swap in a candidate CF IP.
func (c *VLESSConfig) WithAddress(newAddr string) *VLESSConfig {
	copy := *c
	copy.Address = newAddr
	return &copy
}

// WithEndpoint returns a copy of the config with the address and port replaced.
func (c *VLESSConfig) WithEndpoint(newAddr string, newPort int) *VLESSConfig {
	copy := *c
	copy.Address = newAddr
	copy.Port = newPort
	return &copy
}

// ToShareURL reconstructs a share URL from the config.
// For Shadowsocks configs it delegates to ToShadowsocksURL.
func (c *VLESSConfig) ToShareURL() string {
	if c.Protocol == "shadowsocks" {
		return c.ToShadowsocksURL()
	}
	params := url.Values{}
	params.Set("encryption", c.Encryption)
	params.Set("security", c.Security)
	params.Set("type", c.Network)

	if c.SNI != "" {
		params.Set("sni", c.SNI)
	}
	if c.Fingerprint != "" {
		params.Set("fp", c.Fingerprint)
	}
	if c.Insecure {
		params.Set("allowInsecure", "1")
	}
	if len(c.ALPN) > 0 {
		params.Set("alpn", strings.Join(c.ALPN, ","))
	}

	switch c.Network {
	case "ws":
		params.Set("path", c.Path)
		if c.Host != "" {
			params.Set("host", c.Host)
		}
	case "grpc":
		params.Set("serviceName", c.ServiceName)
		if c.Authority != "" {
			params.Set("authority", c.Authority)
		}
		if c.Mode != "" {
			params.Set("mode", c.Mode)
		}
	case "xhttp", "splithttp":
		params.Set("path", c.Path)
		if c.Host != "" {
			params.Set("host", c.Host)
		}
		if c.Mode != "" {
			params.Set("mode", c.Mode)
		}
	}

	remark := url.QueryEscape(c.Remark)
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s", c.UUID, c.Address, c.Port, params.Encode(), remark)
}

func splitHostPort(hostPort string) (string, string, error) {
	// Handle IPv6 [addr]:port
	if strings.HasPrefix(hostPort, "[") {
		end := strings.Index(hostPort, "]")
		if end == -1 {
			return "", "", fmt.Errorf("missing ] in IPv6 address")
		}
		host := hostPort[1:end]
		if end+1 >= len(hostPort) || hostPort[end+1] != ':' {
			return "", "", fmt.Errorf("missing port after IPv6 address")
		}
		return host, hostPort[end+2:], nil
	}

	// Regular host:port
	lastColon := strings.LastIndex(hostPort, ":")
	if lastColon == -1 {
		return "", "", fmt.Errorf("missing port")
	}
	return hostPort[:lastColon], hostPort[lastColon+1:], nil
}

func hasScheme(raw, scheme string) bool {
	prefix := scheme + "://"
	return strings.HasPrefix(strings.ToLower(raw), prefix)
}

func stripScheme(raw, scheme string) string {
	return raw[len(scheme)+3:]
}

func validatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

// recoverMissingQuerySep handles URLs where the '?' separator between port and
// query params was silently dropped (common with certain terminal paste modes).
// Input: portStr like "2053encryption=none&security=tls&sni=..."
// It extracts the leading digit run as the port and merges the rest into params.
func recoverMissingQuerySep(portStr string, params url.Values) (int, url.Values, error) {
	i := 0
	for i < len(portStr) && portStr[i] >= '0' && portStr[i] <= '9' {
		i++
	}
	if i == 0 || i == len(portStr) {
		return 0, params, fmt.Errorf("cannot recover port from %q", portStr)
	}
	port, err := strconv.Atoi(portStr[:i])
	if err != nil {
		return 0, params, err
	}
	extra, _ := url.ParseQuery(portStr[i:])
	if params == nil {
		params = make(url.Values)
	}
	for k, vs := range extra {
		if _, exists := params[k]; !exists {
			params[k] = vs
		}
	}
	return port, params, nil
}

func paramOr(params url.Values, key, fallback string) string {
	v := params.Get(key)
	if v == "" {
		return fallback
	}
	return v
}

// ParseTrojan parses a trojan:// share URL.
// Format: trojan://password@address:port?params#remark
func ParseTrojan(raw string) (*VLESSConfig, error) {
	if !hasScheme(raw, "trojan") {
		return nil, fmt.Errorf("not a trojan:// URL")
	}

	raw = stripScheme(raw, "trojan")

	// Split remark
	remark := ""
	if idx := strings.LastIndex(raw, "#"); idx != -1 {
		remark = raw[idx+1:]
		raw = raw[:idx]
	}
	remark, _ = url.QueryUnescape(remark)

	// Split params
	params := url.Values{}
	if idx := strings.Index(raw, "?"); idx != -1 {
		var err error
		params, err = url.ParseQuery(raw[idx+1:])
		if err != nil {
			return nil, fmt.Errorf("parsing query params: %w", err)
		}
		raw = raw[:idx]
	}

	// Split password@address:port
	atIdx := strings.Index(raw, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("missing @ in URL")
	}
	password, _ := url.QueryUnescape(raw[:atIdx])
	hostPort := raw[atIdx+1:]

	host, portStr, err := splitHostPort(hostPort)
	if err != nil {
		return nil, fmt.Errorf("parsing host:port %q: %w", hostPort, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port, params, err = recoverMissingQuerySep(portStr, params)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q", portStr)
		}
	}
	if err := validatePort(port); err != nil {
		return nil, err
	}

	cfg := &VLESSConfig{
		Protocol:    "trojan",
		Password:    password,
		Address:     host,
		Port:        port,
		Network:     paramOr(params, "type", "tcp"),
		Security:    paramOr(params, "security", "tls"),
		SNI:         params.Get("sni"),
		Fingerprint: paramOr(params, "fp", ""),
		Insecure:    params.Get("insecure") == "1" || params.Get("allowInsecure") == "1",
		Remark:      remark,
	}

	switch cfg.Network {
	case "ws":
		cfg.Path = paramOr(params, "path", "/")
		cfg.Host = paramOr(params, "host", cfg.SNI)
	case "grpc":
		cfg.ServiceName = params.Get("serviceName")
		cfg.Authority = params.Get("authority")
		cfg.Mode = paramOr(params, "mode", "gun")
	case "xhttp", "splithttp":
		cfg.Path = paramOr(params, "path", "/")
		cfg.Host = paramOr(params, "host", cfg.SNI)
		cfg.Mode = paramOr(params, "mode", "auto")
	}

	if alpnStr := params.Get("alpn"); alpnStr != "" {
		cfg.ALPN = strings.Split(alpnStr, ",")
	}

	return cfg, nil
}

// ParseShadowsocks parses a ss:// share URL (SIP002 and legacy formats).
//
// SIP002 (preferred):
//
//	ss://BASE64(method:password)@hostname:port[/?plugin=...][#remark]
//
// Legacy:
//
//	ss://BASE64(method:password@hostname:port)[#remark]
func ParseShadowsocks(raw string) (*VLESSConfig, error) {
	if !hasScheme(raw, "ss") {
		return nil, fmt.Errorf("not a ss:// URL")
	}

	raw = stripScheme(raw, "ss")

	// Split remark
	remark := ""
	if idx := strings.LastIndex(raw, "#"); idx != -1 {
		remark = raw[idx+1:]
		raw = raw[:idx]
	}
	remark, _ = url.QueryUnescape(remark)

	// Strip plugin query string (informational; not needed for scanning)
	if idx := strings.Index(raw, "?"); idx != -1 {
		raw = raw[:idx]
	}

	// Some clients append a trailing "/" before "?"
	raw = strings.TrimSuffix(raw, "/")

	var method, password, address string
	var port int

	if atIdx := strings.Index(raw, "@"); atIdx != -1 {
		// ── SIP002 ──────────────────────────────────────────────────────────
		// userinfo is BASE64(method:password) or plain method:password
		userinfoRaw := raw[:atIdx]
		hostPort := raw[atIdx+1:]

		methodPassword, err := ssDecodeUserinfo(userinfoRaw)
		if err != nil {
			return nil, fmt.Errorf("decoding ss:// userinfo: %w", err)
		}

		colonIdx := strings.Index(methodPassword, ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid ss:// userinfo: missing \":\" between method and password")
		}
		method = strings.ToLower(methodPassword[:colonIdx])
		password = methodPassword[colonIdx+1:]

		host, portStr, err := splitHostPort(hostPort)
		if err != nil {
			return nil, fmt.Errorf("parsing host:port %q: %w", hostPort, err)
		}
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q", portStr)
		}
		address = host
	} else {
		// ── Legacy ───────────────────────────────────────────────────────────
		// entire body is BASE64(method:password@hostname:port)
		decoded, err := base64DecodeSS(raw)
		if err != nil {
			return nil, fmt.Errorf("decoding legacy ss:// URL: %w", err)
		}

		atIdx := strings.LastIndex(decoded, "@")
		if atIdx == -1 {
			return nil, fmt.Errorf("invalid legacy ss:// URL: missing @")
		}

		methodPassword := decoded[:atIdx]
		hostPort := decoded[atIdx+1:]

		colonIdx := strings.Index(methodPassword, ":")
		if colonIdx == -1 {
			return nil, fmt.Errorf("invalid method:password in legacy ss:// URL")
		}
		method = strings.ToLower(methodPassword[:colonIdx])
		password = methodPassword[colonIdx+1:]

		host, portStr, err := splitHostPort(hostPort)
		if err != nil {
			return nil, fmt.Errorf("parsing host:port %q: %w", hostPort, err)
		}
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port %q", portStr)
		}
		address = host
	}

	if err := validatePort(port); err != nil {
		return nil, err
	}

	return &VLESSConfig{
		Protocol: "shadowsocks",
		Method:   method,
		Password: password,
		Address:  address,
		Port:     port,
		Network:  "tcp",
		Security: "none",
		Remark:   remark,
	}, nil
}

// ToShadowsocksURL reconstructs a ss:// SIP002 share URL from the config.
func (c *VLESSConfig) ToShadowsocksURL() string {
	userinfo := base64.RawURLEncoding.EncodeToString([]byte(c.Method + ":" + c.Password))
	remark := url.QueryEscape(c.Remark)
	return fmt.Sprintf("ss://%s@%s:%d#%s", userinfo, c.Address, c.Port, remark)
}

// ssDecodeUserinfo tries to base64-decode a SIP002 userinfo field.
// Falls back to treating it as plain (URL-decoded) "method:password".
func ssDecodeUserinfo(s string) (string, error) {
	decoded, err := base64DecodeSS(s)
	if err == nil && strings.Contains(decoded, ":") {
		return decoded, nil
	}
	// Try plain URL-encoded method:password
	unescaped, _ := url.QueryUnescape(s)
	if strings.Contains(unescaped, ":") {
		return unescaped, nil
	}
	if strings.Contains(s, ":") {
		return s, nil
	}
	return "", fmt.Errorf("cannot interpret %q as base64 or plain method:password", s)
}

// base64DecodeSS decodes a base64 string, tolerating missing padding
// and both standard / URL-safe alphabets.
func base64DecodeSS(s string) (string, error) {
	stripped := strings.TrimRight(s, "=")
	for _, enc := range []*base64.Encoding{base64.RawURLEncoding, base64.RawStdEncoding} {
		if b, err := enc.DecodeString(stripped); err == nil {
			return string(b), nil
		}
	}
	// Restore padding and retry
	switch len(stripped) % 4 {
	case 2:
		stripped += "=="
	case 3:
		stripped += "="
	}
	if b, err := base64.URLEncoding.DecodeString(stripped); err == nil {
		return string(b), nil
	}
	if b, err := base64.StdEncoding.DecodeString(stripped); err == nil {
		return string(b), nil
	}
	return "", fmt.Errorf("base64 decode failed for %q", s)
}
