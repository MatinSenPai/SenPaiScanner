package xraytest

import (
	"encoding/json"
	"fmt"
)

// BuildXrayConfig generates a minimal xray-core JSON config from a VLESSConfig.
// It creates a SOCKS inbound on the given port and a VLESS outbound.
func BuildXrayConfig(cfg *VLESSConfig, socksPort int) ([]byte, error) {
	config := map[string]interface{}{
		"log": map[string]interface{}{
			"loglevel": "none",
			"access":   "",
			"error":    "",
		},
		"dns": map[string]interface{}{
			"servers": []string{"1.1.1.1", "8.8.8.8"},
		},
		"inbounds": []map[string]interface{}{
			{
				"tag":      "socks-in",
				"port":     socksPort,
				"listen":   "127.0.0.1",
				"protocol": "socks",
				"sniffing": map[string]interface{}{
					"enabled":      true,
					"destOverride": []string{"http", "tls"},
				},
				"settings": map[string]interface{}{
					"udp": true,
				},
			},
		},
		"outbounds": []map[string]interface{}{
			buildOutbound(cfg),
			{
				"tag":      "direct",
				"protocol": "freedom",
				"settings": map[string]interface{}{},
			},
		},
	}

	return json.MarshalIndent(config, "", "  ")
}

func buildOutbound(cfg *VLESSConfig) map[string]interface{} {
	switch cfg.Protocol {
	case "trojan":
		return buildTrojanOutbound(cfg)
	case "shadowsocks":
		return buildShadowsocksOutbound(cfg)
	default:
		return buildVLESSOutbound(cfg)
	}
}

func buildShadowsocksOutbound(cfg *VLESSConfig) map[string]interface{} {
	server := map[string]interface{}{
		"address":  cfg.Address,
		"port":     cfg.Port,
		"method":   cfg.Method,
		"password": cfg.Password,
	}
	return map[string]interface{}{
		"tag":      "proxy",
		"protocol": "shadowsocks",
		"settings": map[string]interface{}{
			"servers": []map[string]interface{}{server},
		},
		"streamSettings": map[string]interface{}{
			"network": "tcp",
		},
	}
}

func buildVLESSOutbound(cfg *VLESSConfig) map[string]interface{} {
	users := []map[string]interface{}{
		{
			"id":         cfg.UUID,
			"encryption": cfg.Encryption,
		},
	}
	if cfg.Flow != "" {
		users[0]["flow"] = cfg.Flow
	}

	return map[string]interface{}{
		"tag":      "proxy",
		"protocol": "vless",
		"settings": map[string]interface{}{
			"vnext": []map[string]interface{}{
				{
					"address": cfg.Address,
					"port":    cfg.Port,
					"users":   users,
				},
			},
		},
		"streamSettings": buildStreamSettings(cfg),
	}
}

func buildTrojanOutbound(cfg *VLESSConfig) map[string]interface{} {
	return map[string]interface{}{
		"tag":      "proxy",
		"protocol": "trojan",
		"settings": map[string]interface{}{
			"servers": []map[string]interface{}{
				{
					"address":  cfg.Address,
					"port":     cfg.Port,
					"password": cfg.Password,
				},
			},
		},
		"streamSettings": buildStreamSettings(cfg),
	}
}

func buildStreamSettings(cfg *VLESSConfig) map[string]interface{} {
	stream := map[string]interface{}{
		"network":  cfg.Network,
		"security": cfg.Security,
	}

	// TLS settings
	if cfg.Security == "tls" {
		tls := map[string]interface{}{}
		if cfg.SNI != "" {
			tls["serverName"] = cfg.SNI
		}
		if cfg.Fingerprint != "" {
			tls["fingerprint"] = cfg.Fingerprint
		}
		if cfg.Insecure {
			tls["allowInsecure"] = true
		}
		if len(cfg.ALPN) > 0 {
			tls["alpn"] = cfg.ALPN
		}
		stream["tlsSettings"] = tls
	}

	// Transport settings
	switch cfg.Network {
	case "ws":
		ws := map[string]interface{}{
			"path": cfg.Path,
		}
		// xray-core expects headers as a map, not a top-level "host" field.
		// Using the correct format ensures the Host header reaches the CDN origin.
		if cfg.Host != "" {
			ws["headers"] = map[string]interface{}{
				"Host": cfg.Host,
			}
		}
		stream["wsSettings"] = ws

	case "grpc":
		grpc := map[string]interface{}{
			"serviceName": cfg.ServiceName,
		}
		if cfg.Authority != "" {
			grpc["authority"] = cfg.Authority
		}
		if cfg.Mode == "multi" {
			grpc["multiMode"] = true
		}
		stream["grpcSettings"] = grpc

	case "xhttp", "splithttp":
		xhttp := map[string]interface{}{
			"path": cfg.Path,
		}
		if cfg.Host != "" {
			xhttp["headers"] = map[string]interface{}{
				"Host": cfg.Host,
			}
		}
		if cfg.Mode != "" {
			xhttp["mode"] = cfg.Mode
		}
		stream["xhttpSettings"] = xhttp
	}

	return stream
}

// BuildXrayConfigJSON is a convenience that returns the config as a formatted string.
func BuildXrayConfigJSON(cfg *VLESSConfig, socksPort int) (string, error) {
	b, err := BuildXrayConfig(cfg, socksPort)
	if err != nil {
		return "", fmt.Errorf("building xray config: %w", err)
	}
	return string(b), nil
}
