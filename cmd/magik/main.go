package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ehsatn/MAGIK/internal/ui"
	"github.com/ehsatn/MAGIK/pkg/version"
)

func main() {
	// --version flag without launching TUI
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println("MAGIK", version.String())
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
