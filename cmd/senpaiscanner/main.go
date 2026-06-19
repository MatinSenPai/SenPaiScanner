package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/matinsenpai/senpaiscanner/internal/logger"
	"github.com/matinsenpai/senpaiscanner/internal/ui"
	"github.com/matinsenpai/senpaiscanner/pkg/version"
)

func main() {
	// --version flag without launching TUI
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v" || os.Args[1] == "version") {
		fmt.Println("SenPai Scanner", version.String())
		return
	}

	// Initialize the scan logger. Logs are written to
	// %APPDATA%/senpaiscanner/logs/scan-YYYYMMDD-HHMMSS.log (Windows) or
	// ~/.config/senpaiscanner/logs/... (Linux/macOS).
	scanLog := logger.New(logger.Config{MinLevel: logger.LevelDebug})
	defer scanLog.Close()
	ui.SetLogger(scanLog)

	model := ui.NewApp(version.Version)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Give the UI package a reference so background goroutines can send messages.
	ui.SetProgram(p)

	if _, err := p.Run(); err != nil {
		scanLog.Error(logger.PhaseStartup, "TUI exited with error: %v", err)
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
