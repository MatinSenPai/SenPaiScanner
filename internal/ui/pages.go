package ui

// Page identifies the active screen.
type Page int

const (
	PageHome           Page = iota
	PageQuickScanCount      // count picker for Quick Scan
	PageScanConfig
	PageLiveScan
	PageResults
	PageColos
	PageLiveColos
	PageAbout
	PageScanWithConfig // setup: source, count, workers, timeout, ports
	PageConfigOptional // optional config URL + Phase 2 top N
	PageConfigSetup    // legacy setup (unused)
	PageConfigPhase1   // xray config - fast connectivity scan
	PageConfigPhase2   // xray config - xray validation
)
