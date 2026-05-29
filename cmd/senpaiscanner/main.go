package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/matinsenpai/senpaiscanner/internal/ui"
	"github.com/matinsenpai/senpaiscanner/internal/web"
	"github.com/matinsenpai/senpaiscanner/pkg/version"
)

func main() {
	// --version flag without launching TUI
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println("SenPai Scanner", version.String())
		return
	}
	if len(os.Args) > 1 && (os.Args[1] == "web" || os.Args[1] == "--web") {
		if err := runWeb(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	model := ui.NewApp(version.Version)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Give the UI package a reference so background goroutines can send messages.
	ui.SetProgram(p)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// runWeb parses the flags for the embedded web UI and starts the server.
//
// The listen address is built from --web-host and --web-port so the port can be
// changed without remembering the full host:port syntax. --addr still wins when
// callers want full control (e.g. binding a Unix-style "0.0.0.0:0"), and a bare
// positional host:port is accepted for backward compatibility.
func runWeb(args []string) error {
	fs := flag.NewFlagSet("web", flag.ContinueOnError)
	var (
		host string
		port int
		addr string
	)
	fs.StringVar(&host, "web-host", "127.0.0.1", "interface the web UI binds to (use 0.0.0.0 to expose on the LAN)")
	fs.IntVar(&port, "web-port", 8787, "TCP port the web UI listens on")
	fs.StringVar(&addr, "addr", "", "full host:port listen address (overrides --web-host/--web-port)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: senpaiscanner web [--web-host HOST] [--web-port PORT] [--addr HOST:PORT]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil // usage already printed by the FlagSet
		}
		return err
	}

	listen := strings.TrimSpace(addr)
	if listen == "" {
		// Backward compatibility: `senpaiscanner web 0.0.0.0:9000`.
		if rest := fs.Args(); len(rest) > 0 && strings.Contains(rest[0], ":") {
			listen = rest[0]
		} else {
			if port < 0 || port > 65535 {
				return fmt.Errorf("web port %d out of range (0-65535)", port)
			}
			listen = net.JoinHostPort(host, strconv.Itoa(port))
		}
	}
	return web.Serve(listen, version.Version)
}
