package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/auth"
	"github.com/dcbto/spotui/internal/msgs"
)

func CmdAuthenticate(clientID string) tea.Cmd {
	return func() tea.Msg {
		client, err := auth.Authenticate(clientID)
		if err != nil {
			return msgs.AuthErrMsg{Err: err}
		}
		return msgs.AuthDoneMsg{Client: client}
	}
}
