package xraytest

import (
	"net"
	"strconv"

	"github.com/matinsenpai/senpaiscanner/internal/antidpi"
)

// dialTarget is where xray connects: the loopback fragmenter when StartAntiDPI is active, else the config's address.
func (c *VLESSConfig) dialTarget() (string, int) {
	if c.dialHost != "" {
		return c.dialHost, c.dialPort
	}
	return c.Address, c.Port
}

// StartAntiDPI returns a copy of cfg whose xray outbound goes through a local fragmenting forwarder (stock
// xray-core cannot parse the per-segment finalmask lists, so the fragmentation runs in antidpi.Proxy instead).
// Call stop when the validation is done. With Anti-DPI off it returns cfg itself and a no-op stop.
func StartAntiDPI(cfg *VLESSConfig) (*VLESSConfig, func(), error) {
	px, err := antidpi.StartProxy(cfg.AntiDPI, net.JoinHostPort(cfg.Address, strconv.Itoa(cfg.Port)))
	if err != nil || px == nil {
		return cfg, func() {}, err
	}
	c := *cfg
	c.dialHost, c.dialPort = px.Addr()
	return &c, px.Close, nil
}
