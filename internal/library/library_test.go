package library_test

import (
	"testing"

	"github.com/dcbto/spotui/internal/library"
)

func TestSearchCursorMovesWithinResults(t *testing.T) {
	lib := library.New()
	lib.SetSearchResults([]library.TrackEntry{
		{Name: "A", URI: "spotify:track:a"},
		{Name: "B", URI: "spotify:track:b"},
	})

	lib.MoveDown()
	lib.MoveDown()
	if got := lib.Cursor(); got != 1 {
		t.Fatalf("cursor = %d, want 1", got)
	}
	if got := lib.SelectedURI(); got != "spotify:track:b" {
		t.Fatalf("selected URI = %q, want spotify:track:b", got)
	}

	lib.MoveUp()
	lib.MoveUp()
	if got := lib.Cursor(); got != 0 {
		t.Fatalf("cursor = %d, want 0", got)
	}
}

func TestReplacingSearchResultsClampsCursor(t *testing.T) {
	lib := library.New()
	lib.SetSearchResults([]library.TrackEntry{{URI: "a"}, {URI: "b"}})
	lib.MoveDown()

	lib.SetSearchResults(nil)

	if got := lib.Cursor(); got != 0 {
		t.Fatalf("cursor = %d, want 0", got)
	}
	if got := lib.SelectedURI(); got != "" {
		t.Fatalf("selected URI = %q, want empty", got)
	}
}
