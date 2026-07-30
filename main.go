package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/auth"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/tui"
)

func main() {
	engine, err := spotengine.NewAdapter()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		tui.NewRootModel(engine, auth.OpenURL),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
