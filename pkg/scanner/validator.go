package scanner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ehsatn/MAGIK/internal/xraytest"
)

// ProxyConfig represents a parsed proxy share URL that can be validated against
// candidate Cloudflare endpoints.
type ProxyConfig struct {
	Protocol string // vless, trojan, or vmess
	Address  string
	Port     int

	UserID   string // VLESS/VMess UUID or Trojan password
	Password string // Trojan password alias for callers that expect it

	Network     string // tcp, ws, grpc, xhttp, splithttp
	TLS         bool
	SNI         string
	Path        string
	Host        string
	Fingerprint string
	Remark      string

	raw *xraytest.VLESSConfig
}

// ParseProxyURL parses a VLESS, Trojan, or VMess share URL.
func ParseProxyURL(rawURL string) (*ProxyConfig, error) {
	cfg, err := xraytest.ParseProxyURL(rawURL)
	if err != nil {
		return nil, err
	}
	return proxyConfigFromXray(cfg), nil
}

func proxyConfigFromXray(cfg *xraytest.VLESSConfig) *ProxyConfig {
	if cfg == nil {
		return nil
	}
	raw := *cfg
	userID := cfg.UUID
	password := cfg.Password
	if cfg.Protocol == "trojan" {
		userID = cfg.Password
	}
	return &ProxyConfig{
		Protocol:    cfg.Protocol,
		Address:     cfg.Address,
		Port:        cfg.Port,
		UserID:      userID,
		Password:    password,
		Network:     cfg.Network,
		TLS:         cfg.Security == "tls" || cfg.Security == "reality",
		SNI:         cfg.SNI,
		Path:        cfg.Path,
		Host:        cfg.Host,
		Fingerprint: cfg.Fingerprint,
		Remark:      cfg.Remark,
		raw:         &raw,
	}
}

func (c *ProxyConfig) toXrayConfig() (*xraytest.VLESSConfig, error) {
	if c == nil {
		return nil, fmt.Errorf("proxy config is nil")
	}
	if c.raw != nil {
		cfg := *c.raw
		return &cfg, nil
	}

	protocol := strings.ToLower(strings.TrimSpace(c.Protocol))
	if protocol == "" {
		return nil, fmt.Errorf("proxy protocol is empty")
	}
	switch protocol {
	case "vless", "trojan", "vmess":
	default:
		return nil, fmt.Errorf("unsupported proxy protocol: %s", c.Protocol)
	}

	network := c.Network
	if network == "" {
		network = "tcp"
	}
	security := "none"
	if c.TLS {
		security = "tls"
	}
	port := c.Port
	if port <= 0 {
		port = 443
	}

	cfg := &xraytest.VLESSConfig{
		Protocol:    protocol,
		UUID:        c.UserID,
		Password:    c.Password,
		Address:     c.Address,
		Port:        port,
		Network:     network,
		Security:    security,
		SNI:         c.SNI,
		Path:        c.Path,
		Host:        c.Host,
		Fingerprint: c.Fingerprint,
		Remark:      c.Remark,
	}
	if protocol == "trojan" {
		if cfg.Password == "" {
			cfg.Password = c.UserID
		}
		cfg.UUID = ""
	}
	return cfg, nil
}

// Validator validates candidate endpoints by running the project's embedded
// Xray validation flow.
type Validator struct {
	mu      sync.Mutex
	cfg     *xraytest.VLESSConfig
	timeout time.Duration
}

// NewValidator creates a Validator with a conservative default timeout.
func NewValidator() *Validator {
	return &Validator{timeout: 30 * time.Second}
}

// SetTimeout changes the per-endpoint validation timeout.
func (v *Validator) SetTimeout(timeout time.Duration) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if timeout > 0 {
		v.timeout = timeout
	}
}

// SetupXrayClient stores the proxy configuration used by Validate. The cleanup
// function clears the stored config to match the old public API.
func (v *Validator) SetupXrayClient(cfg *ProxyConfig) (func(), error) {
	xcfg, err := cfg.toXrayConfig()
	if err != nil {
		return nil, err
	}

	v.mu.Lock()
	v.cfg = xcfg
	v.mu.Unlock()

	return func() {
		v.mu.Lock()
		v.cfg = nil
		v.mu.Unlock()
	}, nil
}

// Validate tests targetIP:port with the configured proxy settings.
func (v *Validator) Validate(ctx context.Context, targetIP net.IP, port int) (time.Duration, error) {
	if targetIP == nil {
		return 0, fmt.Errorf("target IP is nil")
	}

	v.mu.Lock()
	cfg := v.cfg
	timeout := v.timeout
	v.mu.Unlock()
	if cfg == nil {
		return 0, fmt.Errorf("Xray config is not set up")
	}

	endpointPort := port
	if endpointPort <= 0 {
		endpointPort = cfg.Port
	}
	endpoint := cfg.WithEndpoint(targetIP.String(), endpointPort)
	res := xraytest.ValidateConfig(ctx, endpoint, timeout)
	if !res.Success {
		if res.Error == "" {
			res.Error = "validation failed"
		}
		return res.Latency, errors.New(res.Error)
	}
	return res.Latency, nil
}
