package banner

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Art is the multi-line ASCII art for "SenPai Scanner": the first word in white, the second in red.
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
const Tagline = "  CLOUDFLARE IP SCANNER  /  CHAIN  /  ANTI-DPI"

// Palette: black, white, red. Nothing else.
var (
	white = lipgloss.NewStyle().Foreground(lipgloss.Color("#F4F4F1")).Bold(true)
	red   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EE2B38")).Bold(true)
	grey  = lipgloss.NewStyle().Foreground(lipgloss.Color("#8D8D88"))
)

// sweepWidth is how many columns the scanning beam lights up.
const sweepWidth = 4

// Render draws the art with a scanner beam travelling left to right.
// frame controls the beam position for animation — increment it each tick.
func Render(frame int) string {
	var out strings.Builder
	word := 0
	beam := frame % 70 // a little longer than the widest line, so the beam leaves the art before it returns

	for _, line := range strings.Split(Art, "\n") {
		if strings.TrimSpace(line) == "" {
			word++
		}
		for col, r := range []rune(line) {
			style := white
			if word > 0 {
				style = red
			}
			if d := col - beam; d >= 0 && d < sweepWidth && r != ' ' {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("#F4F4F1")).Background(lipgloss.Color("#EE2B38")).Bold(true)
			}
			out.WriteString(style.Render(string(r)))
		}
		out.WriteRune('\n')
	}

	out.WriteString(grey.Render(Tagline))
	out.WriteRune('\n')
	return out.String()
}

// RenderStatic returns a non-animated render, suitable for non-TUI contexts like `--version`.
func RenderStatic() string {
	return Render(-100) // beam parked off-screen
}
