package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/library"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
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
		token, err := client.Token()
		if err != nil {
			return msgs.ErrMsg{Err: fmt.Errorf("get token: %w", err), Context: "fetch playlists"}
		}

		httpClient := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
		resp, err := httpClient.Get("https://api.spotify.com/v1/me/playlists?limit=50")
		if err != nil {
			return msgs.ErrMsg{Err: fmt.Errorf("request: %w", err), Context: "fetch playlists"}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return msgs.ErrMsg{Err: fmt.Errorf("API returned status %d", resp.StatusCode), Context: "fetch playlists"}
		}

		var page struct {
			Playlists []apiPlaylist `json:"items"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			return msgs.ErrMsg{Err: fmt.Errorf("decode: %w", err), Context: "fetch playlists"}
		}

		entries := make([]library.PlaylistEntry, len(page.Playlists))
		for i, pl := range page.Playlists {
			entries[i] = library.PlaylistEntry{
				Name:       pl.Name,
				TrackCount: pl.trackCount(),
				URI:        pl.URI,
			}
		}

		return msgs.PlaylistsLoadedMsg{Playlists: entries}
	}
}

type apiPlaylist struct {
	Name   string                `json:"name"`
	URI    string                `json:"uri"`
	Tracks spotify.PlaylistTracks `json:"tracks"`
	Items  spotify.PlaylistTracks `json:"items"`
}

func (p apiPlaylist) trackCount() int {
	if tc := int(p.Items.Total); tc > 0 {
		return tc
	}
	return int(p.Tracks.Total)
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
