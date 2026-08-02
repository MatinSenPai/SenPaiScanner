package export

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/matinsenpai/senpaiscanner/internal/xraytest"
)

// Endpoint is a working proxy endpoint (Cloudflare IP + port).
type Endpoint struct {
	IP   string
	Port int
}

// Bundle holds every export format generated from a single template config
// combined with many working endpoints.
type Bundle struct {
	Subscription string   // one share URL per line
	ShareURLs    []string // the vless/trojan/vmess URLs, one per endpoint
	SingBox      string   // sing-box client config JSON
	Clash        string   // clash proxy list YAML
}

// Generate combines one template config with a list of working endpoints and
// produces subscription URLs, a sing-box JSON config, and a Clash YAML.
func Generate(template *xraytest.VLESSConfig, endpoints []Endpoint) (*Bundle, error) {
	if template == nil {
		return nil, fmt.Errorf("template config is nil")
	}
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("no working endpoints to export")
	}

	b := &Bundle{}

	for _, ep := range endpoints {
		port := ep.Port
		if port <= 0 {
			port = template.Port
		}
		swapped := template.WithEndpoint(ep.IP, port)
		url := swapped.ToShareURL()
		b.ShareURLs = append(b.ShareURLs, url)
		b.Subscription += url + "\n"
	}

	singbox, err := singBox(template, endpoints)
	if err != nil {
		return nil, fmt.Errorf("building sing-box config: %w", err)
	}
	b.SingBox = singbox
	b.Clash = clash(template, endpoints)
	return b, nil
}

// ParseEndpoints converts "ip:port" strings into Endpoint values.
func ParseEndpoints(raw []string) []Endpoint {
	var eps []Endpoint
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		ip := r
		port := 0
		if i := strings.LastIndex(r, ":"); i != -1 {
			ip = r[:i]
			fmt.Sscanf(r[i+1:], "%d", &port)
		}
		if ip == "" {
			continue
		}
		eps = append(eps, Endpoint{IP: ip, Port: port})
	}
	return eps
}

// --- sing-box ---------------------------------------------------------------

type sbOutbound struct {
	Type       string                 `json:"type"`
	Tag        string                 `json:"tag"`
	Server     string                 `json:"server"`
	ServerPort int                    `json:"server_port"`
	UUID       string                 `json:"uuid,omitempty"`
	Password   string                 `json:"password,omitempty"`
	TLS        map[string]interface{} `json:"tls,omitempty"`
	Transport  map[string]interface{} `json:"transport,omitempty"`
}

func singBox(cfg *xraytest.VLESSConfig, endpoints []Endpoint) (string, error) {
	var outbounds []interface{}
	for i, ep := range endpoints {
		port := ep.Port
		if port <= 0 {
			port = cfg.Port
		}
		o := sbOutbound{
			Type:       cfg.Protocol,
			Tag:        fmt.Sprintf("CF-Endpoint-%d", i+1),
			Server:     ep.IP,
			ServerPort: port,
		}
		if cfg.Protocol == "trojan" {
			o.Password = cfg.Password
		} else {
			o.UUID = cfg.UUID
		}
		if cfg.Security == "tls" {
			tlsConf := map[string]interface{}{
				"enabled":     true,
				"server_name": cfg.SNI,
			}
			if cfg.Fingerprint != "" {
				tlsConf["utls"] = map[string]interface{}{
					"enabled":     true,
					"fingerprint": cfg.Fingerprint,
				}
			}
			if len(cfg.ALPN) > 0 {
				tlsConf["alpn"] = cfg.ALPN
			}
			o.TLS = tlsConf
		}
		switch cfg.Network {
		case "ws":
			wsConf := map[string]interface{}{
				"type": "ws",
				"path": cfg.Path,
			}
			if cfg.Host != "" {
				wsConf["headers"] = map[string]interface{}{
					"Host": cfg.Host,
				}
			}
			o.Transport = wsConf
		case "grpc":
			grpcConf := map[string]interface{}{
				"type":         "grpc",
				"service_name": cfg.ServiceName,
			}
			o.Transport = grpcConf
		case "xhttp", "splithttp":
			xhttpConf := map[string]interface{}{
				"type": "xhttp",
				"path": cfg.Path,
			}
			if cfg.Host != "" {
				xhttpConf["headers"] = map[string]interface{}{
					"Host": cfg.Host,
				}
			}
			o.Transport = xhttpConf
		}
		outbounds = append(outbounds, o)
	}
	config := map[string]interface{}{"outbounds": outbounds}
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// --- clash ------------------------------------------------------------------

func clash(cfg *xraytest.VLESSConfig, endpoints []Endpoint) string {
	var lines []string
	lines = append(lines, "proxies:")
	for i, ep := range endpoints {
		port := ep.Port
		if port <= 0 {
			port = cfg.Port
		}
		lines = append(lines, fmt.Sprintf("  - name: \"CF-Endpoint-%d\"", i+1))
		lines = append(lines, fmt.Sprintf("    type: %s", cfg.Protocol))
		lines = append(lines, fmt.Sprintf("    server: %s", ep.IP))
		lines = append(lines, fmt.Sprintf("    port: %d", port))

		if cfg.Protocol == "trojan" {
			lines = append(lines, fmt.Sprintf("    password: %s", cfg.Password))
			lines = append(lines, "    udp: true")
		} else {
			lines = append(lines, fmt.Sprintf("    uuid: %s", cfg.UUID))
			if cfg.Protocol == "vmess" {
				lines = append(lines, "    alterId: 0")
				lines = append(lines, "    cipher: auto")
			}
		}

		if cfg.Security == "tls" {
			lines = append(lines, "    tls: true")
			if cfg.SNI != "" {
				lines = append(lines, fmt.Sprintf("    servername: %s", cfg.SNI))
			}
			if cfg.Fingerprint != "" {
				lines = append(lines, fmt.Sprintf("    client-fingerprint: %s", cfg.Fingerprint))
			}
		}

		switch cfg.Network {
		case "ws":
			lines = append(lines, "    network: ws")
			lines = append(lines, "    ws-opts:")
			lines = append(lines, fmt.Sprintf("      path: %s", cfg.Path))
			if cfg.Host != "" {
				lines = append(lines, "      headers:")
				lines = append(lines, fmt.Sprintf("        Host: %s", cfg.Host))
			}
		case "grpc":
			lines = append(lines, "    network: grpc")
			lines = append(lines, "    grpc-opts:")
			lines = append(lines, fmt.Sprintf("      grpc-service-name: %s", cfg.ServiceName))
		case "xhttp", "splithttp":
			lines = append(lines, "    network: h2")
			lines = append(lines, "    http-opts:")
			lines = append(lines, "      path:")
			lines = append(lines, fmt.Sprintf("        - %s", cfg.Path))
			if cfg.Host != "" {
				lines = append(lines, "      headers:")
				lines = append(lines, "        Host:")
				lines = append(lines, fmt.Sprintf("          - %s", cfg.Host))
			}
		}
	}
	return strings.Join(lines, "\n") + "\n"
}
