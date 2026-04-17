package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/clientid"
	"github.com/dcbto/spotui/internal/tui"
)

func main() {
	id, err := clientid.Resolve()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		tui.NewRootModel(id),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
