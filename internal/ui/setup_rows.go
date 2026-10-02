package ui

// configSetupRow indices for the "Find Working IPs" setup page.
//
// These were bare integer literals scattered through the view and key-handling
// code. When a Tries row was inserted above Ports, the port pill highlight kept
// testing row 4 and so lit up while the cursor was on Tries, and went dark the
// moment the cursor moved onto Ports. Named constants make a future insert or
// reorder fail loudly at review time instead of silently mis-highlighting.
const (
	rowSource   = 0
	rowCount    = 1
	rowWorkers  = 2
	rowTimeout  = 3
	rowTries    = 4
	rowPorts    = 5
	rowWS       = 6
	rowNeighbor = 7
)

// lastSetupRow is the highest navigable row on the setup page.
const lastSetupRow = rowNeighbor