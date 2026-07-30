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
