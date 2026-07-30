package commands

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotengine"
)

const TrackSearchLimit = 10

func CmdSearchEngineTracks(engine spotengine.Engine, query string, offset int) tea.Cmd {
	return func() tea.Msg {
		page, err := engine.SearchTracks(context.Background(), spotengine.SearchRequest{
			Query:  query,
			Limit:  TrackSearchLimit,
			Offset: offset,
		})
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "search tracks"}
		}
		return msgs.EngineTrackSearchLoadedMsg{
			Tracks: page.Tracks,
			Total:  page.Total,
			Offset: page.Offset,
		}
	}
}
