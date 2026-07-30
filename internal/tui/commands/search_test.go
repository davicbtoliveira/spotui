package commands_test

import (
	"errors"
	"testing"

	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/tui/commands"
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
	if !ok || len(loaded.Tracks) != 1 || loaded.Tracks[0].URI != "spotify:track:hello" {
		t.Fatalf("message: %#v", msg)
	}
}

func TestCmdSearchEngineTracksReturnsActionableError(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetError(spotengine.OperationSearchTracks, errors.New("service unavailable"))

	msg := commands.CmdSearchEngineTracks(engine, "hello", 0)()

	errMsg, ok := msg.(msgs.ErrMsg)
	if !ok || errMsg.Context != "search tracks" || errMsg.Err.Error() != "service unavailable" {
		t.Fatalf("message: %#v", msg)
	}
}
