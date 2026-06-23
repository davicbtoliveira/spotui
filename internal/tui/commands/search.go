package commands

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotifyapi"
	"github.com/zmb3/spotify/v2"
)

const trackSearchLimit = 10

func CmdSearchTracks(searcher spotifyapi.TrackSearcher, query string) tea.Cmd {
	return func() tea.Msg {
		page, err := searcher.SearchTracks(context.Background(), spotifyapi.TrackSearchRequest{
			Query:  query,
			Limit:  trackSearchLimit,
			Market: spotify.MarketFromToken,
		})
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "search tracks"}
		}
		return msgs.TrackSearchLoadedMsg{Tracks: page.Tracks, Total: page.Total}
	}
}
