package mobile

import (
	"github.com/matinsenpai/senpaiscanner/internal/antidpi"
	"github.com/matinsenpai/senpaiscanner/internal/ui"
)

// setAntiDPI turns the published fragmentation recipe on or off for the scan that is about to run.
func setAntiDPI(on bool) {
	p := antidpi.DefaultProfile()
	p.Enabled = on
	ui.SetAntiDPI(p)
}
