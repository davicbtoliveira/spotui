package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/browser"
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
		tui.NewRootModel(engine, browser.OpenURL),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, runErr := p.Run()
	if err := errors.Join(runErr, shutdown(engine)); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func shutdown(engine spotengine.Engine) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return engine.Close(ctx)
}
