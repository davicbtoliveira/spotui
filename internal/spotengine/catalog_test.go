package spotengine

import (
	"context"
	"fmt"
	"testing"

	"github.com/devgianlu/go-librespot/daemon"
)

func TestAdapterTranslatesLikedTracksFromNativeCatalog(t *testing.T) {
	server := newMemoryAPIServer()
	runtime := &requestRuntime{server: server, requests: make(chan daemon.ApiRequest, 1), reply: func(request daemon.ApiRequest) (any, error) {
		if request.Type == daemon.ApiRequestTypeWebApi {
			return nil, daemon.ErrTooManyRequests
		}
		if request.Type != daemon.ApiRequestTypeNativeCatalog {
			return nil, fmt.Errorf("unexpected request type: %q", request.Type)
		}
		data := request.Data.(daemon.ApiRequestDataNativeCatalog)
		if data.Kind != "liked" || data.Limit != 10 {
			t.Fatalf("catalog request: %#v", data)
		}
		return daemon.ApiResponseNativeCatalog{Payload: []byte(`{"total":1,"offset":0,"limit":10,"items":[{"track":{"uri":"spotify:track:hello","name":"Hello","duration_ms":1000,"artists":[{"name":"Adele"}],"album":{"name":"25","uri":"spotify:album:25"}}}]}`)}, nil
	}}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer adapter.Close(context.Background())
	page, err := adapter.LikedTracks(context.Background(), PageRequest{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "Hello" || page.Items[0].AlbumURI != "spotify:album:25" {
		t.Fatalf("page: %#v", page)
	}
}

func TestAdapterTranslatesSavedAlbumsFromNativeCatalog(t *testing.T) {
	server := newMemoryAPIServer()
	runtime := &requestRuntime{server: server, requests: make(chan daemon.ApiRequest, 1), reply: func(request daemon.ApiRequest) (any, error) {
		if request.Type != daemon.ApiRequestTypeNativeCatalog {
			return nil, fmt.Errorf("unexpected request type: %q", request.Type)
		}
		return daemon.ApiResponseNativeCatalog{Payload: []byte(`{"total":1,"offset":0,"limit":10,"items":[{"album":{"uri":"spotify:album:25","name":"25","artists":[{"name":"Adele"}],"total_tracks":11}}]}`)}, nil
	}}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer adapter.Close(context.Background())
	page, err := adapter.SavedAlbums(context.Background(), PageRequest{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].Name != "25" || page.Items[0].TrackCount != 11 {
		t.Fatalf("page: %#v", page)
	}
}

func TestAdapterTranslatesArtistPublicPlaylistsFromNativeCatalog(t *testing.T) {
	server := newMemoryAPIServer()
	runtime := &requestRuntime{server: server, requests: make(chan daemon.ApiRequest, 1), reply: func(request daemon.ApiRequest) (any, error) {
		data := request.Data.(daemon.ApiRequestDataNativeCatalog)
		if request.Type != daemon.ApiRequestTypeNativeCatalog || data.Kind != "artist" {
			return nil, fmt.Errorf("catalog request: %#v", request)
		}
		return daemon.ApiResponseNativeCatalog{Payload: []byte(`{"artist":{"uri":"spotify:artist:arctic","name":"Arctic Monkeys"},"popular":{"items":[]},"albums":{"items":[]},"playlists":{"items":[{"uri":"spotify:playlist:indie","name":"Indie Forever","owner":{"name":"Spotify"},"tracks":{"total":50}}]}}`)}, nil
	}}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer adapter.Close(context.Background())
	detail, err := adapter.Artist(context.Background(), "spotify:artist:arctic", PageRequest{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Playlists) != 1 || detail.Playlists[0].URI != "spotify:playlist:indie" || detail.Playlists[0].Owner != "Spotify" {
		t.Fatalf("playlists: %#v", detail.Playlists)
	}
}

func TestAdapterSearchTranslatesPublicPlaylistsFromNativeCatalog(t *testing.T) {
	server := newMemoryAPIServer()
	runtime := &requestRuntime{server: server, requests: make(chan daemon.ApiRequest, 1), reply: func(request daemon.ApiRequest) (any, error) {
		data, ok := request.Data.(daemon.ApiRequestDataNativeCatalog)
		if request.Type != daemon.ApiRequestTypeNativeCatalog || !ok || data.Kind != "search" {
			return nil, fmt.Errorf("catalog request: %#v", request)
		}
		return daemon.ApiResponseNativeCatalog{Payload: []byte(`{"tracks":{"total":1,"offset":0,"limit":10,"items":[{"uri":"spotify:track:hello","name":"Hello","artists":[{"name":"Adele"}],"album":{"name":"25"}}]},"playlists":{"total":1,"offset":0,"limit":10,"items":[{"uri":"spotify:playlist:indie","name":"Indie Forever","owner":{"name":"Spotify"},"tracks":{"total":50}}]}}`)}, nil
	}}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer adapter.Close(context.Background())
	groups, err := adapter.Search(context.Background(), SearchRequest{Query: "indie", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups.Playlists.Items) != 1 || groups.Playlists.Items[0].URI != "spotify:playlist:indie" || groups.PlaylistsUnavailable {
		t.Fatalf("search groups: %#v", groups)
	}
}

func TestAdapterRecommendedLoadsPublicPlaylistsWithoutLosingTopContent(t *testing.T) {
	server := newMemoryAPIServer()
	runtime := &requestRuntime{server: server, requests: make(chan daemon.ApiRequest, 3), reply: func(request daemon.ApiRequest) (any, error) {
		data := request.Data.(daemon.ApiRequestDataNativeCatalog)
		switch data.Kind {
		case "top_artists":
			return daemon.ApiResponseNativeCatalog{Payload: []byte(`{"items":[{"uri":"spotify:artist:arctic","name":"Arctic Monkeys"}]}`)}, nil
		case "top_tracks":
			return daemon.ApiResponseNativeCatalog{Payload: []byte(`{"items":[{"uri":"spotify:track:hello","name":"Hello","artists":[{"name":"Arctic Monkeys"}],"album":{"name":"AM","uri":"spotify:album:am"}}]}`)}, nil
		case "recommended_playlists":
			return daemon.ApiResponseNativeCatalog{Payload: []byte(`{"items":[{"uri":"spotify:playlist:discover","name":"Discover Weekly","owner":{"name":"Spotify"},"tracks":{"total":30}}]}`)}, nil
		default:
			return nil, fmt.Errorf("catalog request: %#v", request)
		}
	}}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer adapter.Close(context.Background())
	page, err := adapter.Recommended(context.Background(), PageRequest{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Artists.Items) != 1 || len(page.Tracks.Items) != 1 || len(page.Playlists) != 1 || page.Playlists[0].URI != "spotify:playlist:discover" || page.PlaylistsUnavailable {
		t.Fatalf("recommended page: %#v", page)
	}
}

func TestAdapterSendsContextShuffleAndRelativeSeek(t *testing.T) {
	server := newMemoryAPIServer()
	runtime := &requestRuntime{server: server, requests: make(chan daemon.ApiRequest, 3)}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer adapter.Close(context.Background())
	if err := adapter.PlayContext(context.Background(), "spotify:album:25", "spotify:track:hello", 0); err != nil {
		t.Fatal(err)
	}
	play := <-runtime.requests
	data := play.Data.(daemon.ApiRequestDataPlay)
	if play.Type != daemon.ApiRequestTypePlay || data.Uri != "spotify:album:25" || data.SkipToUri != "spotify:track:hello" {
		t.Fatalf("play: %#v", play)
	}
	if err := adapter.SetShuffle(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	shuffle := <-runtime.requests
	if shuffle.Type != daemon.ApiRequestTypeSetShufflingContext || shuffle.Data.(bool) != true {
		t.Fatalf("shuffle: %#v", shuffle)
	}
	if err := adapter.SeekRelative(context.Background(), -10000); err != nil {
		t.Fatal(err)
	}
	seek := <-runtime.requests
	seekData := seek.Data.(daemon.ApiRequestDataSeek)
	if seek.Type != daemon.ApiRequestTypeSeek || seekData.Position != -10000 || !seekData.Relative {
		t.Fatalf("seek: %#v", seek)
	}
}
