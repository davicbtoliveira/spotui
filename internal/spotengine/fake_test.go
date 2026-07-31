package spotengine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dcbto/spotui/internal/spotengine"
)

func TestFakeRecordsPlay(t *testing.T) {
	engine := spotengine.NewFake()

	if err := engine.Play(context.Background(), "spotify:track:hello"); err != nil {
		t.Fatalf("play: %v", err)
	}

	calls := engine.Calls()
	if len(calls) != 1 {
		t.Fatalf("calls: want 1, got %d", len(calls))
	}
	if calls[0].Operation != spotengine.OperationPlay {
		t.Fatalf("operation: want %q, got %q", spotengine.OperationPlay, calls[0].Operation)
	}
	if calls[0].URI != "spotify:track:hello" {
		t.Fatalf("URI: want spotify:track:hello, got %q", calls[0].URI)
	}
}

func TestFakeSupportsStableEngineContract(t *testing.T) {
	engine := spotengine.NewFake()

	ctx := context.Background()
	if err := engine.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := engine.Reconnect(ctx); err != nil {
		t.Fatalf("reconnect: %v", err)
	}
	if err := engine.CancelLogin(ctx); err != nil {
		t.Fatalf("cancel login: %v", err)
	}
	if err := engine.Logout(ctx); err != nil {
		t.Fatalf("logout: %v", err)
	}
	operations := []struct {
		name string
		call func() error
	}{
		{"pause", func() error { return engine.Pause(ctx) }},
		{"resume", func() error { return engine.Resume(ctx) }},
		{"next", func() error { return engine.Next(ctx) }},
		{"previous", func() error { return engine.Previous(ctx) }},
		{"volume", func() error { return engine.SetVolume(ctx, 65) }},
		{"autoplay", func() error { return engine.SetAutoplay(ctx, false) }},
		{"close", func() error { return engine.Close(ctx) }},
	}
	for _, operation := range operations {
		if err := operation.call(); err != nil {
			t.Fatalf("%s: %v", operation.name, err)
		}
	}

	want := []spotengine.Operation{
		spotengine.OperationStart,
		spotengine.OperationReconnect,
		spotengine.OperationCancelLogin,
		spotengine.OperationLogout,
		spotengine.OperationPause,
		spotengine.OperationResume,
		spotengine.OperationNext,
		spotengine.OperationPrevious,
		spotengine.OperationSetVolume,
		spotengine.OperationSetAutoplay,
		spotengine.OperationClose,
	}
	calls := engine.Calls()
	if len(calls) != len(want) {
		t.Fatalf("calls: want %d, got %d", len(want), len(calls))
	}
	for i, operation := range want {
		if calls[i].Operation != operation {
			t.Fatalf("call %d: want %q, got %q", i, operation, calls[i].Operation)
		}
	}
	if calls[8].Volume != 65 || calls[9].Enabled {
		t.Fatalf("recorded arguments: %#v", calls)
	}

	var contract spotengine.Engine = engine
	if contract == nil {
		t.Fatal("fake does not implement Engine")
	}
}

func TestFakeReturnsConfiguredOperationError(t *testing.T) {
	engine := spotengine.NewFake()
	want := errors.New("player unavailable")
	engine.SetError(spotengine.OperationNext, want)

	if err := engine.Next(context.Background()); !errors.Is(err, want) {
		t.Fatalf("next error: want %v, got %v", want, err)
	}
}

func TestFakePublishesEngineEvents(t *testing.T) {
	engine := spotengine.NewFake()
	want := spotengine.Event{
		Type: spotengine.EventTypeMetadata,
		Track: &spotengine.Track{
			URI:    "spotify:track:hello",
			Name:   "Hello",
			Artist: "Adele",
		},
	}

	engine.Emit(want)

	got := <-engine.Events()
	if got.Type != want.Type || got.Track == nil || *got.Track != *want.Track {
		t.Fatalf("event: want %#v, got %#v", want, got)
	}
}

func TestFakeAutoplayDefaultsOnAndChangesAtRuntime(t *testing.T) {
	engine := spotengine.NewFake()
	if !engine.AutoplayEnabled() {
		t.Fatal("autoplay default: want enabled")
	}

	if err := engine.SetAutoplay(context.Background(), false); err != nil {
		t.Fatalf("disable autoplay: %v", err)
	}
	if engine.AutoplayEnabled() {
		t.Fatal("autoplay remained enabled")
	}
}
