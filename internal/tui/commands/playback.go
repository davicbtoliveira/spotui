package commands

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/msgs"
)

func CmdClearStatus() tea.Cmd {
	return tea.Tick(4*time.Second, func(_ time.Time) tea.Msg {
		return msgs.ClearStatusMsg{}
	})
}

func CmdProgressTick() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return msgs.ProgressTickMsg{}
	})
}
