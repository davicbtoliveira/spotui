package main

import (
	"testing"

	"github.com/dcbto/spotui/internal/spotengine"
)

func TestShutdownClosesPlaybackEngine(t *testing.T) {
	engine := spotengine.NewFake()

	if err := shutdown(engine); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Operation != spotengine.OperationClose {
		t.Fatalf("engine calls: %#v", calls)
	}
}
