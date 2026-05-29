package banner

import (
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Art is the multi-line ASCII art for "SenPai Scanner".
// Uses box-drawing block characters for a bold, retro look.
const Art = `
 ░██████╗███████╗███╗░░██╗██████╗░░█████╗░██╗
 ██╔════╝██╔════╝████╗░██║██╔══██╗██╔══██╗██║
 ╚█████╗░█████╗░░██╔██╗██║██████╔╝███████║██║
 ░╚═══██╗██╔══╝░░██║╚████║██╔═══╝░██╔══██║██║
 ██████╔╝███████╗██║░╚███║██║░░░░░██║░░██║██║
 ╚═════╝░╚══════╝╚═╝░░╚══╝╚═╝░░░░░╚═╝░░╚═╝╚═╝

 ░██████╗░█████╗░░█████╗░███╗░░██╗███╗░░██╗███████╗██████╗░
 ██╔════╝██╔══██╗██╔══██╗████╗░██║████╗░██║██╔════╝██╔══██╗
 ╚█████╗░██║░░╚═╝███████║██╔██╗██║██╔██╗██║█████╗░░██████╔╝
 ░╚═══██╗██║░░██╗██╔══██║██║╚████║██║╚████║██╔══╝░░██╔══██╗
 ██████╔╝╚█████╔╝██║░░██║██║░╚███║██║░╚███║███████╗██║░░██║
 ╚═════╝░░╚════╝░╚═╝░░╚═╝╚═╝░░╚══╝╚═╝░░╚══╝╚══════╝╚═╝░░╚═╝`

// Tagline is shown beneath the art.
const Tagline = "  Cloudflare IP Scanner — tuned for restricted networks"

// rainbowPalette is a smooth warm→cool gradient used for color cycling.
var rainbowPalette = []string{
	"#FF4C4C", "#FF6B35", "#FF8C42", "#FFAE5E", "#FFC94E",
	"#FFE066", "#C8FF66", "#66FFB2", "#4CF2FF", "#4CB8FF",
	"#7B6FFF", "#B066FF", "#FF66E0", "#FF4CA8", "#FF4C6E",
}

// The gradient repeats every len(rainbowPalette) columns, so a frame's output
// depends only on frame mod len(rainbowPalette). All distinct frames are
// rendered once and cached; animation is then a constant-time slice lookup
// instead of thousands of per-rune style allocations per tick.
var (
	frameOnce  sync.Once
	frameCache []string
)

func buildFrames() {
	n := len(rainbowPalette)
	styles := make([]lipgloss.Style, n)
	for i, c := range rainbowPalette {
		styles[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Bold(true)
	}
	tagline := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Italic(true).Render(Tagline)

	lines := strings.Split(Art, "\n")
	frameCache = make([]string, n)
	for f := 0; f < n; f++ {
		var out strings.Builder
		for _, line := range lines {
			for col, r := range []rune(line) {
				out.WriteString(styles[(col+f)%n].Render(string(r)))
			}
			out.WriteRune('\n')
		}
		out.WriteString(tagline)
		out.WriteRune('\n')
		frameCache[f] = out.String()
	}
}

// Render returns the rainbow-gradient ASCII art for the given animation frame.
// Increment frame each tick to animate the color cycle.
func Render(frame int) string {
	frameOnce.Do(buildFrames)
	n := len(frameCache)
	return frameCache[((frame%n)+n)%n]
}
