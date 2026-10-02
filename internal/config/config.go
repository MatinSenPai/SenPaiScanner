// Package config holds default value constants for scan and test operations.
// The TUI reads these defaults when building form initial values.
package config

import "time"

// ScanDefaults are the factory defaults shown in the scan config form.
var ScanDefaults = struct {
	Count       int
	Concurrency int
	Timeout     time.Duration
	Tries       int
	Port        int
	Mode        string
	UseV4       bool
	UseV6       bool
	Top         int
	Stability   bool
	InterTryGap time.Duration
}{
	Count: 500,

	// Tries is the per-IP sample count. It went 4 -> 10 earlier, then settled
	// at 8, which is the smallest count that both resolves packet loss to
	// 12.5% steps and still lets the tail percentile discard a worst sample
	// and report something other than the maximum. Higher counts sharpen
	// loss resolution further but cost proportionally more probe time.
	Tries: 8,

	// Timeout was 5s. On Iranian mobile and fibre links a healthy Cloudflare
	// edge frequently needs 6-8s to finish TLS plus the trace GET, so a 5s
	// budget was discarding usable IPs as false loss.
	Timeout: 8 * time.Second,

	Concurrency: 50,
	Port:        443,
	Mode:        "http",
	UseV4:       true,
	UseV6:       false,
	Top:         10,

	// Stability keeps the idle-hold check that catches DPI resetting a
	// connection after the first successful request. It is on by default
	// because turning it off accepts IPs that die mid-session.
	Stability: true,

	// InterTryGap replaces the hardcoded 10-60ms sleep. A 20ms bound keeps the
	// cadence irregular (so traffic does not look metronomic) while removing
	// most of the dead time between tries.
	InterTryGap: 20 * time.Millisecond,
}
