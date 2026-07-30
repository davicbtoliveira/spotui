package commands_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/spotifyapi"
	"github.com/dcbto/spotui/internal/tui/commands"
	"github.com/zmb3/spotify/v2"
)

func TestCmdSearchEngineTracksUsesAuthenticatedEngine(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetSearchResult(spotengine.SearchPage{
		Tracks: []spotengine.Track{{
			URI:        "spotify:track:hello",
			Name:       "Hello",
			Artist:     "Adele",
			DurationMS: 295000,
		}},
		Total:  1,
		Offset: 20,
	}, nil)

	msg := commands.CmdSearchEngineTracks(engine, "track:Hello artist:Adele", 20)()

	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Operation != spotengine.OperationSearchTracks {
		t.Fatalf("engine calls: %#v", calls)
	}
	request := calls[0].Search
	if request.Query != "track:Hello artist:Adele" || request.Offset != 20 || request.Limit != 10 {
		t.Fatalf("search request: %#v", request)
	}
	loaded, ok := msg.(msgs.EngineTrackSearchLoadedMsg)
	if !ok {
		t.Fatalf("message: want EngineTrackSearchLoadedMsg, got %T", msg)
	}
	if len(loaded.Tracks) != 1 || loaded.Tracks[0].URI != "spotify:track:hello" {
		t.Fatalf("tracks: %#v", loaded.Tracks)
	}
}

type fakeTrackSearcher struct {
	req spotifyapi.TrackSearchRequest
	err error
}

func (f *fakeTrackSearcher) SearchTracks(_ context.Context, req spotifyapi.TrackSearchRequest) (spotifyapi.TrackSearchPage, error) {
	f.req = req
	if f.err != nil {
		return spotifyapi.TrackSearchPage{}, f.err
	}
	return spotifyapi.TrackSearchPage{
		Tracks: []spotify.FullTrack{
			{
				SimpleTrack: spotify.SimpleTrack{
					Name: "Hello",
					URI:  "spotify:track:hello",
				},
			},
		},
		Total: 1,
	}, nil
}

func TestCmdSearchTracksUsesSpotifyTrackSearch(t *testing.T) {
	searcher := &fakeTrackSearcher{}

	msg := commands.CmdSearchTracks(searcher, "track:Hello artist:Adele", 0)()

	if searcher.req.Query != "track:Hello artist:Adele" {
		t.Fatalf("Query: want raw query, got %q", searcher.req.Query)
	}
	if searcher.req.Limit != 10 {
		t.Fatalf("Limit: want 10, got %d", searcher.req.Limit)
	}
	if searcher.req.Market != spotify.MarketFromToken {
		t.Fatalf("Market: want %q, got %q", spotify.MarketFromToken, searcher.req.Market)
	}
	loaded, ok := msg.(msgs.TrackSearchLoadedMsg)
	if !ok {
		t.Fatalf("message: want TrackSearchLoadedMsg, got %T", msg)
	}
	if len(loaded.Tracks) != 1 || loaded.Tracks[0].Name != "Hello" {
		t.Fatalf("Tracks: got %#v", loaded.Tracks)
	}
}

func TestCmdSearchTracksPassesOffset(t *testing.T) {
	searcher := &fakeTrackSearcher{}

	commands.CmdSearchTracks(searcher, "hello", 20)()

	if searcher.req.Offset != 20 {
		t.Fatalf("Offset: want 20, got %d", searcher.req.Offset)
	}
}

func TestCmdSearchTracksReturnsSearchErrorContext(t *testing.T) {
	searcher := &fakeTrackSearcher{err: errors.New("rate limited")}

	msg := commands.CmdSearchTracks(searcher, "hello", 0)()

	got, ok := msg.(msgs.ErrMsg)
	if !ok {
		t.Fatalf("message: want ErrMsg, got %T", msg)
	}
	if got.Context != "search tracks" {
		t.Fatalf("Context: want %q, got %q", "search tracks", got.Context)
	}
	if got.Err.Error() != "rate limited" {
		t.Fatalf("Err: got %v", got.Err)
	}
}

func TestCmdSearchTracksUsesValidSpotifySearchLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"status":400,"message":"invalid limit"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"tracks":{"items":[],"total":0}}`))
	}))
	defer server.Close()

	client := spotify.New(server.Client(), spotify.WithBaseURL(server.URL+"/"))

	msg := commands.CmdSearchTracks(spotifyapi.SpotifyTrackSearcher{Client: client}, "hello", 0)()

	if _, ok := msg.(msgs.TrackSearchLoadedMsg); !ok {
		t.Fatalf("message: want TrackSearchLoadedMsg, got %T: %#v", msg, msg)
	}
}
