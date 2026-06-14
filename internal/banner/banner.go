package banner

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Art is the multi-line ASCII art for "MAGIK".
const Art = `
 __  __    _    ____ ___ _  __
|  \/  |  / \  / ___|_ _| |/ /
| |\/| | / _ \| |  _ | || ' /
| |  | |/ ___ \ |_| || || . \
|_|  |_/_/   \_\____|___|_|\_\`

// Tagline is shown beneath the art.
const Tagline = "  Cloudflare IP Scanner - tuned for restricted networks"

// rainbowPalette is a smooth warm→cool gradient used for color cycling.
var rainbowPalette = []string{
	"#FF4C4C", "#FF6B35", "#FF8C42", "#FFAE5E", "#FFC94E",
	"#FFE066", "#C8FF66", "#66FFB2", "#4CF2FF", "#4CB8FF",
	"#7B6FFF", "#B066FF", "#FF66E0", "#FF4CA8", "#FF4C6E",
}

// Render applies a rainbow gradient to the ASCII art.
// frame controls the color offset for animation — increment it each tick.
func Render(frame int) string {
	lines := strings.Split(Art, "\n")
	var out strings.Builder

	for _, line := range lines {
		runes := []rune(line)
		for col, r := range runes {
			idx := ((col+frame)%len(rainbowPalette) + len(rainbowPalette)) % len(rainbowPalette)
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(rainbowPalette[idx])).Bold(true)
			out.WriteString(style.Render(string(r)))
		}
		out.WriteRune('\n')
	}

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Italic(true)
	out.WriteString(dim.Render(Tagline))
	out.WriteRune('\n')

	return out.String()
}

// Version returns a static (non-animated) render at frame=0, suitable for
// non-TUI contexts like `--version`.
func RenderStatic() string {
	return Render(0)
}
