package commands

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotengine"
)

func CmdStartEngine(engine spotengine.Engine) tea.Cmd {
	return func() tea.Msg {
		if err := engine.Start(context.Background()); err != nil {
			return msgs.EngineStartErrMsg{Err: err}
		}
		return msgs.EngineStartedMsg{}
	}
}

func CmdWaitEngineEvent(engine spotengine.Engine) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-engine.Events()
		if !ok {
			return msgs.EngineEventsClosedMsg{}
		}
		return msgs.EngineEventMsg{Event: event}
	}
}

func CmdOpenURL(openURL func(string) error, url string) tea.Cmd {
	return func() tea.Msg {
		if err := openURL(url); err != nil {
			return msgs.BrowserOpenErrMsg{Err: err}
		}
		return msgs.BrowserOpenedMsg{}
	}
}

func CmdResetLogin(engine spotengine.Engine, clearSession bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if clearSession {
			err = engine.Logout(context.Background())
		} else {
			err = engine.CancelLogin(context.Background())
		}
		if err != nil {
			return msgs.LoginResetErrMsg{Err: err}
		}
		return msgs.LoginResetMsg{}
	}
}

func CmdLogout(engine spotengine.Engine) tea.Cmd {
	return CmdResetLogin(engine, true)
}
