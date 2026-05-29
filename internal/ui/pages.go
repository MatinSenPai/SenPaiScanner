package ui

// Page identifies the active screen.
type Page int

const (
	PageHome           Page = iota
	PageQuickScanCount      // count picker for Quick Scan
	PageScanConfig
	PageLiveScan
	PageResults
	PageLiveColos
	PageAbout
)
