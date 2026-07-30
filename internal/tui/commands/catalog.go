package commands

import (
	"context"
	"fmt"
	"image"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotengine"
)

func CmdLoadArtwork(rawURL string) tea.Cmd {
	return func() tea.Msg {
		request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
		if err != nil {
			return msgs.ArtworkLoadedMsg{URL: rawURL, Err: err}
		}
		client := http.Client{Timeout: 5 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			return msgs.ArtworkLoadedMsg{URL: rawURL, Err: err}
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return msgs.ArtworkLoadedMsg{URL: rawURL, Err: fmt.Errorf("artwork request: %s", response.Status)}
		}
		decoded, _, err := image.Decode(response.Body)
		return msgs.ArtworkLoadedMsg{URL: rawURL, Image: decoded, Err: err}
	}
}

func CmdLikedTracks(engine spotengine.Engine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("liked", func() (any, error) { return engine.LikedTracks(context.Background(), request) }, requestID...)
}

func CmdUserPlaylists(engine spotengine.Engine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("playlists", func() (any, error) { return engine.UserPlaylists(context.Background(), request) }, requestID...)
}

func CmdSavedAlbums(engine spotengine.Engine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("albums", func() (any, error) { return engine.SavedAlbums(context.Background(), request) }, requestID...)
}

func CmdSearch(engine spotengine.Engine, request spotengine.SearchRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("search", func() (any, error) { return engine.Search(context.Background(), request) }, requestID...)
}

func CmdPlaylist(engine spotengine.Engine, uri string, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("playlist:"+uri, func() (any, error) { return engine.Playlist(context.Background(), uri, request) }, requestID...)
}

func CmdAlbum(engine spotengine.Engine, uri string, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("album:"+uri, func() (any, error) { return engine.Album(context.Background(), uri, request) }, requestID...)
}

func CmdArtist(engine spotengine.Engine, uri string, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("artist:"+uri, func() (any, error) { return engine.Artist(context.Background(), uri, request) }, requestID...)
}

func CmdTopArtists(engine spotengine.Engine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("recommended", func() (any, error) { return engine.TopArtists(context.Background(), request) }, requestID...)
}

func CmdTopTracks(engine spotengine.Engine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("top_tracks", func() (any, error) { return engine.TopTracks(context.Background(), request) }, requestID...)
}

func CmdRecommended(engine spotengine.Engine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd("recommended", func() (any, error) { return engine.Recommended(context.Background(), request) }, requestID...)
}

func CmdPlayContext(engine spotengine.Engine, contextURI, trackURI string, positionMS int) tea.Cmd {
	return func() tea.Msg {
		if err := engine.PlayContext(context.Background(), contextURI, trackURI, positionMS); err != nil {
			return msgs.ErrMsg{Err: err, Context: "play context"}
		}
		return nil
	}
}

func CmdSetEngineShuffle(engine spotengine.Engine, enabled bool) tea.Cmd {
	return func() tea.Msg {
		if err := engine.SetShuffle(context.Background(), enabled); err != nil {
			return msgs.ErrMsg{Err: err, Context: "set shuffle"}
		}
		return nil
	}
}

func CmdSeekEngine(engine spotengine.Engine, deltaMS int) tea.Cmd {
	return func() tea.Msg {
		if err := engine.SeekRelative(context.Background(), deltaMS); err != nil {
			return msgs.ErrMsg{Err: err, Context: "seek"}
		}
		return nil
	}
}

func catalogCmd(route string, load func() (any, error), requestIDs ...uint64) tea.Cmd {
	return func() tea.Msg {
		data, err := load()
		var requestID uint64
		if len(requestIDs) > 0 {
			requestID = requestIDs[0]
		}
		return msgs.CatalogLoadedMsg{Route: route, RequestID: requestID, Data: data, Err: err}
	}
}
