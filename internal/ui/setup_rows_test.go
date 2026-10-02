package ui

import "testing"

// setupRows pins the display order of the "Find Working IPs" setup page. The
// ports-highlight bug happened because inserting the Tries row above Ports moved
// the ports row index from 4 to 5 while the highlight kept comparing against a
// bare 4, so the pills lit up on the Tries row and went dark on Ports. These
// tests fail loudly if a future row insert breaks that mapping again.
var setupRows = []struct {
	name string
	idx  int
}{
	{"Source", rowSource},
	{"Count", rowCount},
	{"Workers", rowWorkers},
	{"Timeout", rowTimeout},
	{"Tries", rowTries},
	{"Ports", rowPorts},
	{"WebSocket", rowWS},
	{"Neighbors", rowNeighbor},
}

func TestSetupRowConstantsMatchDisplayOrder(t *testing.T) {
	for i, row := range setupRows {
		if row.idx != i {
			t.Errorf("row %q = %d, want %d — setup rows must stay contiguous and in display order",
				row.name, row.idx, i)
		}
	}
	if lastSetupRow != rowNeighbor {
		t.Errorf("lastSetupRow = %d, want %d (Neighbors)", lastSetupRow, rowNeighbor)
	}
}

func TestPortsRowIsNotTriesRow(t *testing.T) {
	if rowPorts == rowTries {
		t.Fatalf("rowPorts == rowTries (%d): the port pills would highlight while the "+
			"cursor sits on Tries", rowPorts)
	}
	if rowPorts != rowTries+1 {
		t.Errorf("rowPorts = %d, want %d — Ports must sit directly below Tries",
			rowPorts, rowTries+1)
	}
}

// TestPortPillFocusedOnlyOnPortsRow is the direct regression test: moving the
// setup cursor onto Ports must light exactly one pill, the focused one, and
// moving the cursor anywhere else -- including Tries -- must light none.
//
// Each focus value needs its own model. portPillFocused reports whether a
// specific pill is highlighted for one particular focus, so holding a single
// focus fixed and asking about every other pill is the case that matters: only
// the pill matching that focus may report true.
func TestPortPillFocusedOnlyOnPortsRow(t *testing.T) {
	for _, row := range setupRows {
		for focus := range configPortChoices {
			m := AppModel{configSetupRow: row.idx, configPortFocus: focus}

			if got := m.portPillFocused(focus); got != (row.idx == rowPorts) {
				t.Errorf("cursor on %-9s (row %d), pill %d: portPillFocused = %v, want %v",
					row.name, row.idx, focus, got, row.idx == rowPorts)
			}

			// Every other pill must be dark, whichever pill holds focus.
			// configPortChoices holds {label, port} structs, so iterate indices.
			for other := range configPortChoices {
				if other == focus {
					continue
				}
				if got := m.portPillFocused(other); got {
					t.Errorf("cursor on %-9s (row %d): pill %d reported highlighted while "+
						"focus is on pill %d", row.name, row.idx, other, focus)
				}
			}
		}
	}
}

// TestPortPillHighlightFollowsFocusOnPortsRow checks that the left/right cursor
// moves the highlight to a different pill once the ports row is active.
func TestPortPillHighlightFollowsFocusOnPortsRow(t *testing.T) {
	m := AppModel{configSetupRow: rowPorts, configPortFocus: 0}

	if !m.portPillFocused(0) {
		t.Fatal("pill 0 should be highlighted on the ports row")
	}
	for i := 1; i < len(configPortChoices); i++ {
		if m.portPillFocused(i) {
			t.Errorf("pill %d highlighted while focus is on pill 0", i)
		}
	}

	last := len(configPortChoices) - 1
	m.configPortFocus = last
	if !m.portPillFocused(last) {
		t.Errorf("pill %d should be highlighted after moving focus right", last)
	}
	if m.portPillFocused(0) {
		t.Error("pill 0 still highlighted after moving focus away")
	}
}