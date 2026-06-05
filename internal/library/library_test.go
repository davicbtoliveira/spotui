package library_test

import (
	"testing"

	"github.com/dcbto/spotui/internal/library"
)

func TestMoveDownMovesCursor(t *testing.T) {
	lib := library.New()
	lib.SetPlaylists([]library.PlaylistEntry{
		{Name: "A"}, {Name: "B"},
	})
	lib.MoveDown()
	if lib.Cursor() != 1 {
		t.Fatalf("want cursor 1, got %d", lib.Cursor())
	}
}

func TestMoveDownClampsAtLast(t *testing.T) {
	lib := library.New()
	lib.SetPlaylists([]library.PlaylistEntry{{Name: "A"}})
	lib.MoveDown()
	if lib.Cursor() != 0 {
		t.Fatalf("want clamp at 0, got %d", lib.Cursor())
	}
}

func TestSetActiveTab(t *testing.T) {
	lib := library.New()
	lib.SetActiveTab(library.TabTracks)
	if lib.ActiveTab() != library.TabTracks {
		t.Fatalf("want TabArtists, got %v", lib.ActiveTab())
	}
}

func TestSetTracksEnablesTrackCursor(t *testing.T) {
	lib := library.New()
	lib.SetTracks([]library.TrackEntry{{Name: "T1"}, {Name: "T2"}})
	lib.SetActiveTab(library.TabTracks)
	lib.MoveDown()
	if lib.Cursor() != 1 {
		t.Fatalf("want 1, got %d", lib.Cursor())
	}
}

func TestMoveUpMovesCursor(t *testing.T) {
	lib := library.New()
	lib.SetPlaylists([]library.PlaylistEntry{{Name: "A"}, {Name: "B"}})
	lib.MoveDown()
	lib.MoveUp()
	if lib.Cursor() != 0 {
		t.Fatalf("want 0, got %d", lib.Cursor())
	}
}

func TestMoveUpClampsAtTop(t *testing.T) {
	lib := library.New()
	lib.SetPlaylists([]library.PlaylistEntry{{Name: "A"}})
	lib.MoveUp()
	if lib.Cursor() != 0 {
		t.Fatalf("want 0, got %d", lib.Cursor())
	}
}
