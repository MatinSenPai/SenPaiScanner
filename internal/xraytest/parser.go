package xraytest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// VLESSConfig holds parsed parameters from a VLESS, Trojan or VMess share URL.
// Check the Protocol field to know which type this is.
type VLESSConfig struct {
	// Protocol is "vless", "trojan", or "vmess".
	Protocol string

	// VLESS-specific
	UUID       string
	Encryption string
	Flow       string

	// Trojan-specific
	Password string

	// Shadowsocks-specific
	Method string // cipher, e.g. "aes-256-gcm"

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

	// Custom configurations passed to speed runner
	SpeedURL  string
	SpeedSize int64

	// Download test flag — when true, Phase 2 measures download throughput.
	DownloadTest bool

	// Upload test flag — when true, Phase 2 measures upload throughput.
	UploadTest bool
}

// ParseProxyURL auto-detects the protocol (vless://, trojan://, or vmess://) and parses
// the share URL into a VLESSConfig. Returns an error if the scheme is unknown.
func ParseProxyURL(raw string) (*VLESSConfig, error) {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)
	var cfg *VLESSConfig
	var err error
	switch {
	case strings.HasPrefix(lower, "vless://"):
		cfg, err = ParseVLESS(raw)
	case strings.HasPrefix(lower, "trojan://"):
		cfg, err = ParseTrojan(raw)
	case strings.HasPrefix(lower, "vmess://"):
		cfg, err = ParseVMess(raw)
	case strings.HasPrefix(lower, "ss://"):
		cfg, err = ParseShadowsocks(raw)
	default:
		return nil, fmt.Errorf("unsupported URL scheme — must start with vless://, trojan://, vmess://, or ss://")
	}
	if err != nil {
		return nil, err
	}
	if cfg != nil && cfg.Insecure {
		return nil, fmt.Errorf("allowInsecure/insecure is set to true (unsupported in latest xray-core)")
	}
	return cfg, nil
}

// ParseVMess parses a vmess:// share URL (base64-encoded JSON) into a VLESSConfig.
func ParseVMess(raw string) (*VLESSConfig, error) {
	if !hasScheme(raw, "vmess") {
		return nil, fmt.Errorf("not a vmess:// URL")
	}
	b64 := stripScheme(raw, "vmess")
	if idx := strings.Index(b64, "?"); idx != -1 {
		b64 = b64[:idx]
	}
	if idx := strings.Index(b64, "#"); idx != -1 {
		b64 = b64[:idx]
	}
	b64 = strings.TrimSpace(b64)
	b64 = strings.ReplaceAll(b64, " ", "")

	// Fix base64 padding
	if l := len(b64) % 4; l > 0 {
		b64 += strings.Repeat("=", 4-l)
	}

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		data, err = base64.URLEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("decode vmess base64: %w", err)
		}
	}

	type VMessJSON struct {
		V    interface{} `json:"v"`
		Ps   string      `json:"ps"`
		Add  string      `json:"add"`
		Port interface{} `json:"port"`
		Id   string      `json:"id"`
		Aid  interface{} `json:"aid"`
		Scy  string      `json:"scy"`
		Net  string      `json:"net"`
		Type string      `json:"type"`
		Host string      `json:"host"`
		Path string      `json:"path"`
		Tls  string      `json:"tls"`
		Sni  string      `json:"sni"`
		Alpn string      `json:"alpn"`
		Fp   string      `json:"fp"`
	}

	var vj VMessJSON
	if err := json.Unmarshal(data, &vj); err != nil {
		return nil, fmt.Errorf("parse vmess json: %w", err)
	}

	var port int
	switch p := vj.Port.(type) {
	case float64:
		port = int(p)
	case string:
		port, _ = strconv.Atoi(p)
	}
	if port == 0 {
		port = 443
	}

	security := vj.Tls
	if security == "" {
		security = "none"
	}

	cfg := &VLESSConfig{
		Protocol:    "vmess",
		UUID:        vj.Id,
		Address:     vj.Add,
		Port:        port,
		Network:     vj.Net,
		Security:    security,
		SNI:         vj.Sni,
		Host:        vj.Host,
		Path:        vj.Path,
		Fingerprint: vj.Fp,
		Remark:      vj.Ps,
	}
	if cfg.Network == "" {
		cfg.Network = "tcp"
	}
	if cfg.Security == "tls" && cfg.SNI == "" {
		cfg.SNI = cfg.Host
	}
	if vj.Alpn != "" {
		cfg.ALPN = strings.Split(vj.Alpn, ",")
	}
	return cfg, nil
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

	var query string
	if idx := strings.Index(raw, "?"); idx != -1 {
		query = raw[idx+1:]
		raw = raw[:idx]
	}
	params := parseShareQuery(query)

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

	cfg.SNI = normalizeKnownHostTypos(cfg.SNI)
	cfg.Host = normalizeKnownHostTypos(cfg.Host)

	return cfg, nil
}

// normalizeKnownHostTypos fixes common paste/save typos that break Phase 2 while
// Phase 1 still passes (Phase 1 falls back to generic Cloudflare trace SNIs).
func normalizeKnownHostTypos(host string) string {
	if strings.Contains(host, ".worers.dev") {
		return strings.ReplaceAll(host, ".worers.dev", ".workers.dev")
	}
	return host
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

// ToShareURL reconstructs a vless://, trojan://, or vmess:// share URL from the config.
func (c *VLESSConfig) ToShareURL() string {
	if c.Protocol == "vmess" {
		type VMessJSON struct {
			V    string `json:"v"`
			Ps   string `json:"ps"`
			Add  string `json:"add"`
			Port int    `json:"port"`
			Id   string `json:"id"`
			Aid  int    `json:"aid"`
			Scy  string `json:"scy"`
			Net  string `json:"net"`
			Type string `json:"type"`
			Host string `json:"host"`
			Path string `json:"path"`
			Tls  string `json:"tls"`
			Sni  string `json:"sni"`
			Alpn string `json:"alpn"`
			Fp   string `json:"fp"`
		}
		tlsVal := ""
		if c.Security == "tls" {
			tlsVal = "tls"
		}
		vj := VMessJSON{
			V:    "2",
			Ps:   c.Remark,
			Add:  c.Address,
			Port: c.Port,
			Id:   c.UUID,
			Aid:  0,
			Scy:  "auto",
			Net:  c.Network,
			Type: "none",
			Host: c.Host,
			Path: c.Path,
			Tls:  tlsVal,
			Sni:  c.SNI,
			Fp:   c.Fingerprint,
		}
		if len(c.ALPN) > 0 {
			vj.Alpn = strings.Join(c.ALPN, ",")
		}
		b, _ := json.Marshal(vj)
		return "vmess://" + base64.StdEncoding.EncodeToString(b)
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
	if c.Protocol == "trojan" {
		return fmt.Sprintf("trojan://%s@%s:%d?%s#%s", c.Password, c.Address, c.Port, params.Encode(), remark)
	}
	if c.Protocol == "shadowsocks" {
		encoded := base64.RawURLEncoding.EncodeToString([]byte(c.Method + ":" + c.Password))
		return fmt.Sprintf("ss://%s@%s:%d#%s", encoded, c.Address, c.Port, remark)
	}
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

// vlessQueryKeys lists share-link parameter names used to delimit values that
// may contain '&' or '?' (common in CF worker WS paths).
var vlessQueryKeys = []string{
	"encryption", "security", "sni", "fp", "alpn", "insecure", "allowInsecure",
	"type", "host", "path", "serviceName", "authority", "mode", "flow", "ed",
	"packetEncoding", "headerType", "seed", "pbk", "sid", "spx",
}

func parseShareQuery(query string) url.Values {
	params := make(url.Values)
	if query == "" {
		return params
	}
	parsed, err := url.ParseQuery(query)
	if err == nil {
		for k, vs := range parsed {
			params[k] = vs
		}
	}
	for _, key := range []string{"path", "host", "sni"} {
		if v := extractQueryValue(query, key); v != "" {
			params.Set(key, v)
		}
	}
	normalizeSharePath(params)
	return params
}

func extractQueryValue(query, key string) string {
	prefix := key + "="
	idx := strings.Index(query, prefix)
	if idx < 0 {
		idx = strings.Index(query, "&"+prefix)
		if idx >= 0 {
			idx++
		}
	}
	if idx < 0 {
		return ""
	}
	start := idx + len(prefix)
	end := len(query)
	for _, k := range vlessQueryKeys {
		if k == key {
			continue
		}
		sep := "&" + k + "="
		if j := strings.Index(query[start:], sep); j >= 0 && start+j < end {
			end = start + j
		}
	}
	val, err := url.QueryUnescape(query[start:end])
	if err != nil {
		return query[start:end]
	}
	return val
}

func normalizeSharePath(params url.Values) {
	path := params.Get("path")
	if path == "" {
		return
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if ed := params.Get("ed"); ed != "" && !strings.Contains(path, "ed=") {
		if strings.Contains(path, "?") {
			path += "&ed=" + ed
		} else {
			path += "?ed=" + ed
		}
	}
	params.Set("path", path)
}

// Phase2SanityError reports config problems that make Phase 2 fail while Phase 1
// can still pass (Phase 1 does not require a valid WS path).
func (c *VLESSConfig) Phase2SanityError() string {
	if c == nil {
		return "empty config"
	}
	if c.Insecure {
		return "allowInsecure/insecure is set to true (unsupported in latest xray-core)"
	}
	switch c.Network {
	case "ws", "xhttp", "splithttp":
		if c.Host == "" && c.SNI == "" {
			return "missing host/sni in VLESS link"
		}
		if len(c.Path) < 24 && strings.Contains(c.Path, "eyJ") {
			return fmt.Sprintf("WS path looks truncated (%d chars) — re-paste the full vless:// link with encoded path", len(c.Path))
		}
	}
	return ""
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
	var query string
	if idx := strings.Index(raw, "?"); idx != -1 {
		query = raw[idx+1:]
		raw = raw[:idx]
	}
	params := parseShareQuery(query)

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

// ParseShadowsocks parses an ss:// share URL into a VLESSConfig.
// Supported formats:
//
//	ss://base64(method:password)@host:port#remark
//	ss://method:password@host:port#remark
func ParseShadowsocks(raw string) (*VLESSConfig, error) {
	if !hasScheme(raw, "ss") {
		return nil, fmt.Errorf("not an ss:// URL")
	}

	raw = stripScheme(raw, "ss")

	// Split remark
	remark := ""
	if idx := strings.LastIndex(raw, "#"); idx >= 0 {
		remark = raw[idx+1:]
		remark, _ = url.QueryUnescape(remark)
		raw = raw[:idx]
	}

	// Extract userinfo@host:port
	atIdx := strings.LastIndex(raw, "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("ss:// URL missing @ separator")
	}
	userinfo := raw[:atIdx]
	hostPort := raw[atIdx+1:]

	// Decode userinfo — may be base64(method:password) or plaintext method:password
	method, password, err := decodeSSUserinfo(userinfo)
	if err != nil {
		return nil, fmt.Errorf("decoding ss:// userinfo: %w", err)
	}

	// Parse host:port
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return nil, fmt.Errorf("parsing ss:// host:port: %w", err)
	}
	port, _ := strconv.Atoi(portStr)
	if port <= 0 {
		return nil, fmt.Errorf("ss:// invalid port %q", portStr)
	}

	return &VLESSConfig{
		Protocol: "shadowsocks",
		Method:   method,
		Password: password,
		Address:  host,
		Port:     port,
		Remark:   remark,
	}, nil
}

// decodeSSUserinfo decodes the userinfo part of an ss:// URL.
// It first tries base64 decoding (standard format), then falls back to
// plaintext method:password.
func decodeSSUserinfo(raw string) (method, password string, err error) {
	// Try base64 decode first (standard Shadowsocks format).
	// Try multiple base64 variants to handle various padding styles.
	candidates := []string{raw, strings.TrimRight(raw, "=")}
	for _, c := range candidates {
		if decoded, decErr := base64.StdEncoding.DecodeString(c); decErr == nil {
			raw = string(decoded)
			goto decoded
		}
		if decoded, decErr := base64.RawStdEncoding.DecodeString(c); decErr == nil {
			raw = string(decoded)
			goto decoded
		}
		if decoded, decErr := base64.URLEncoding.DecodeString(c); decErr == nil {
			raw = string(decoded)
			goto decoded
		}
		if decoded, decErr := base64.RawURLEncoding.DecodeString(c); decErr == nil {
			raw = string(decoded)
			goto decoded
		}
	}
	// else: treat as plaintext method:password

decoded:
	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected method:password, got %q", raw)
	}
	return parts[0], parts[1], nil
}
