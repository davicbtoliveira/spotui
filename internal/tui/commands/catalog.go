package commands

import (
	"context"
	"fmt"
	"image"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/catalog"
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

func CmdLikedTracks(engine spotengine.CatalogEngine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd(catalog.Route{Kind: catalog.RouteLiked}, func() (catalog.Payload, error) {
		value, err := engine.LikedTracks(context.Background(), request)
		return catalog.TrackPagePayload{Value: value}, err
	}, requestID...)
}

func CmdUserPlaylists(engine spotengine.CatalogEngine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd(catalog.Route{Kind: catalog.RoutePlaylists}, func() (catalog.Payload, error) {
		value, err := engine.UserPlaylists(context.Background(), request)
		return catalog.PlaylistPagePayload{Value: value}, err
	}, requestID...)
}

func CmdSavedAlbums(engine spotengine.CatalogEngine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd(catalog.Route{Kind: catalog.RouteAlbums}, func() (catalog.Payload, error) {
		value, err := engine.SavedAlbums(context.Background(), request)
		return catalog.AlbumPagePayload{Value: value}, err
	}, requestID...)
}

func CmdSearch(engine spotengine.CatalogEngine, request spotengine.SearchRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd(catalog.Route{Kind: catalog.RouteSearch, Query: request.Query}, func() (catalog.Payload, error) {
		value, err := engine.Search(context.Background(), request)
		return catalog.SearchPayload{Value: value}, err
	}, requestID...)
}

func CmdPlaylist(engine spotengine.CatalogEngine, uri string, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd(catalog.Route{Kind: catalog.RoutePlaylist, URI: uri}, func() (catalog.Payload, error) {
		value, err := engine.Playlist(context.Background(), uri, request)
		return catalog.PlaylistPayload{Value: value}, err
	}, requestID...)
}

func CmdAlbum(engine spotengine.CatalogEngine, uri string, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd(catalog.Route{Kind: catalog.RouteAlbum, URI: uri}, func() (catalog.Payload, error) {
		value, err := engine.Album(context.Background(), uri, request)
		return catalog.AlbumPayload{Value: value}, err
	}, requestID...)
}

func CmdArtist(engine spotengine.CatalogEngine, uri string, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd(catalog.Route{Kind: catalog.RouteArtist, URI: uri}, func() (catalog.Payload, error) {
		value, err := engine.Artist(context.Background(), uri, request)
		return catalog.ArtistPayload{Value: value}, err
	}, requestID...)
}

func CmdRecommended(engine spotengine.CatalogEngine, request spotengine.PageRequest, requestID ...uint64) tea.Cmd {
	return catalogCmd(catalog.Route{Kind: catalog.RouteRecommended}, func() (catalog.Payload, error) {
		value, err := engine.Recommended(context.Background(), request)
		return catalog.RecommendedPayload{Value: value}, err
	}, requestID...)
}

func CmdPlayContext(engine spotengine.PlaybackEngine, contextURI, trackURI string, positionMS int) tea.Cmd {
	return func() tea.Msg {
		if err := engine.PlayContext(context.Background(), contextURI, trackURI, positionMS); err != nil {
			return msgs.ErrMsg{Err: err, Context: "play context"}
		}
		return nil
	}
}

func CmdSetEngineShuffle(engine spotengine.PlaybackEngine, enabled bool) tea.Cmd {
	return func() tea.Msg {
		if err := engine.SetShuffle(context.Background(), enabled); err != nil {
			return msgs.ErrMsg{Err: err, Context: "set shuffle"}
		}
		return nil
	}
}

func CmdSeekEngine(engine spotengine.PlaybackEngine, deltaMS int) tea.Cmd {
	return func() tea.Msg {
		if err := engine.SeekRelative(context.Background(), deltaMS); err != nil {
			return msgs.ErrMsg{Err: err, Context: "seek"}
		}
		return nil
	}
}

func catalogCmd(route catalog.Route, load func() (catalog.Payload, error), requestIDs ...uint64) tea.Cmd {
	return func() tea.Msg {
		payload, err := load()
		var requestID uint64
		if len(requestIDs) > 0 {
			requestID = requestIDs[0]
		}
		return msgs.CatalogLoadedMsg{Route: route, RequestID: requestID, Payload: payload, Err: err}
	}
}
