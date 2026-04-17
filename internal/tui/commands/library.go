package commands

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/zmb3/spotify/v2"
)

func CmdFetchUser(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		user, err := client.CurrentUser(context.Background())
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "fetch user"}
		}
		return msgs.UserLoadedMsg{User: user}
	}
}

func CmdFetchPlaylists(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		page, err := client.CurrentUsersPlaylists(context.Background(), spotify.Limit(50))
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "fetch playlists"}
		}
		return msgs.PlaylistsLoadedMsg{Playlists: page.Playlists}
	}
}

func CmdFetchTracks(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		page, err := client.CurrentUsersTracks(context.Background(), spotify.Limit(50))
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "fetch tracks"}
		}
		tracks := make([]spotify.SavedTrack, len(page.Tracks))
		copy(tracks, page.Tracks)
		return msgs.TracksLoadedMsg{Tracks: tracks, Total: int(page.Total)}
	}
}

func CmdFetchTopArtists(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		page, err := client.CurrentUsersTopArtists(
			context.Background(),
			spotify.Timerange(spotify.MediumTermRange),
			spotify.Limit(50),
		)
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "fetch artists"}
		}
		return msgs.ArtistsLoadedMsg{Artists: page.Artists}
	}
}
