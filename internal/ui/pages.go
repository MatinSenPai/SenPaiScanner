package ui

// Page identifies the active screen.
type Page int

const (
	PageHome           Page = iota
	PageScanConfig
	PageLiveScan
	PageResults
	PageAbout
	PageScanWithConfig // setup: source, count, workers, timeout, ports
	PageConfigOptional // optional config URL + Phase 2 top N
	PageConfigPhase1   // xray config - fast connectivity scan
	PageConfigPhase2   // xray config - xray validation
	PageLogViewer      // scan log viewer
)
