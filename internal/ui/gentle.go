package ui

import (
	"sync/atomic"
	"time"
)

// The "Gentle" profile is for ISPs that cut the connection when a scan is too aggressive (#25 #56 #62 #96):
// few workers, a patient timeout and a global cap on how many probes start per second.
const (
	GentleWorkers = 25
	GentleTimeout = 6 * time.Second
	GentleRate    = 40.0 // probes started per second
)

// ponytail: process-wide like the Anti-DPI profile - one scan at a time in the TUI.
var gentleOn atomic.Bool

// Gentle reports whether the low-impact profile is active for the terminal UI.
func Gentle() bool { return gentleOn.Load() }

// SetGentle turns the low-impact profile on or off.
func SetGentle(on bool) { gentleOn.Store(on) }

// gentleLimits applies the profile to the worker count and timeout; rate is 0 when the profile is off.
func gentleLimits(workers int, timeout time.Duration) (int, time.Duration, float64) {
	if !Gentle() {
		return workers, timeout, 0
	}
	return min(workers, GentleWorkers), max(timeout, GentleTimeout), GentleRate
}
