package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/matinsenpai/senpaiscanner/internal/plaincli"
	"github.com/matinsenpai/senpaiscanner/internal/ui"
	"github.com/matinsenpai/senpaiscanner/pkg/version"
)

func main() {
	// --version flag without launching TUI
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println("SenPai Scanner", version.String())
		return
	}

	// Plain text mode for screen readers and scripts: "scan ..." or --plain / --no-tui / --simple / --accessible.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "scan", "--plain", "--no-tui", "--simple", "--accessible":
			os.Exit(plaincli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
		case "-h", "--help", "help":
			fmt.Print(plaincli.Usage)
			return
		}
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
