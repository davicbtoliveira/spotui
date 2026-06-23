package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/library"
	"github.com/zmb3/spotify/v2"
)

func newReadyModel() RootModel {
	m := NewRootModel("")
	m.state = stateReady
	m.client = &spotify.Client{}
	m.width = 80
	m.height = 24
	return m
}

func sendKey(m RootModel, key string) RootModel {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(RootModel)
}

func sendSpecialKey(m RootModel, key tea.KeyType) RootModel {
	updated, _ := m.Update(tea.KeyMsg{Type: key})
	return updated.(RootModel)
}

func TestTabKeysRouteToLibrary(t *testing.T) {
	m := newReadyModel()

	m = sendKey(m, "2")
	if got := m.library.ActiveTab(); got != library.TabTracks {
		t.Fatalf("Tab 2: want TabTracks, got %v", got)
	}
	m = sendKey(m, "3")
	if got := m.library.ActiveTab(); got != library.TabArtists {
		t.Fatalf("Tab 3: want TabArtists, got %v", got)
	}
	m = sendKey(m, "1")
	if got := m.library.ActiveTab(); got != library.TabPlaylists {
		t.Fatalf("Tab 1: want TabPlaylists, got %v", got)
	}
	m = sendKey(m, "4")
	if got := m.library.ActiveTab(); got != library.TabSearch {
		t.Fatalf("Tab 4: want TabSearch, got %v", got)
	}
	if view := m.View(); !strings.Contains(view, "Press / to search tracks") {
		t.Fatalf("search empty state missing from view:\n%s", view)
	}
}

func TestCursorMovesPerActiveTab(t *testing.T) {
	m := newReadyModel()
	m.library.SetPlaylists([]library.PlaylistEntry{{Name: "P1"}, {Name: "P2"}})
	m.library.SetTracks([]library.TrackEntry{{Name: "T1"}, {Name: "T2"}})

	m = sendKey(m, "2")
	m = sendKey(m, "j")
	if got := m.library.Cursor(); got != 1 {
		t.Fatalf("tracks j: want 1, got %d", got)
	}
	m = sendKey(m, "1")
	if got := m.library.Cursor(); got != 0 {
		t.Fatalf("playlists untouched: want 0, got %d", got)
	}
}

func TestSelectedURIFromActiveTab(t *testing.T) {
	m := newReadyModel()
	m.library.SetPlaylists([]library.PlaylistEntry{
		{Name: "P1", URI: "spotify:playlist:1"},
		{Name: "P2", URI: "spotify:playlist:2"},
	})
	m = sendKey(m, "j")
	if got := m.library.SelectedURI(); got != "spotify:playlist:2" {
		t.Fatalf("selected: want spotify:playlist:2, got %q", got)
	}
}

func TestSlashOpensTrackSearchInput(t *testing.T) {
	m := newReadyModel()

	m = sendKey(m, "/")

	if got := m.library.ActiveTab(); got != library.TabSearch {
		t.Fatalf("slash: want TabSearch, got %v", got)
	}
	if view := m.View(); !strings.Contains(view, "Search: ") {
		t.Fatalf("search input missing from view:\n%s", view)
	}
}

func TestTrackSearchInputAppendsTypedCharacters(t *testing.T) {
	m := newReadyModel()

	m = sendKey(m, "/")
	m = sendKey(m, "a")
	m = sendKey(m, "d")
	m = sendKey(m, "e")
	m = sendKey(m, "l")
	m = sendKey(m, "e")

	if view := m.View(); !strings.Contains(view, "Search: adele") {
		t.Fatalf("search query missing from view:\n%s", view)
	}
}

func TestTrackSearchInputAcceptsSpace(t *testing.T) {
	m := newReadyModel()

	m = sendKey(m, "/")
	m = sendKey(m, "a")
	m = sendSpecialKey(m, tea.KeySpace)
	m = sendKey(m, "b")

	if view := m.View(); !strings.Contains(view, "Search: a b") {
		t.Fatalf("search query missing space:\n%s", view)
	}
}

func TestTrackSearchInputBackspaceEditsQuery(t *testing.T) {
	m := newReadyModel()

	m = sendKey(m, "/")
	m = sendKey(m, "a")
	m = sendKey(m, "b")
	m = sendSpecialKey(m, tea.KeyBackspace)

	if view := m.View(); !strings.Contains(view, "Search: a") || strings.Contains(view, "Search: ab") {
		t.Fatalf("backspace did not edit query:\n%s", view)
	}
}

func TestEscapeCancelsTrackSearchInput(t *testing.T) {
	m := newReadyModel()

	m = sendKey(m, "/")
	m = sendKey(m, "a")
	m = sendSpecialKey(m, tea.KeyEsc)

	if got := m.library.ActiveTab(); got != library.TabSearch {
		t.Fatalf("escape: want TabSearch, got %v", got)
	}
	view := m.View()
	if strings.Contains(view, "Search: ") {
		t.Fatalf("escape left search input visible:\n%s", view)
	}
	if !strings.Contains(view, "Press / to search tracks") {
		t.Fatalf("escape did not return to search empty state:\n%s", view)
	}
}

func TestBlankTrackSearchSubmitDoesNotStartCommand(t *testing.T) {
	m := newReadyModel()

	m = sendKey(m, "/")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)

	if cmd != nil {
		t.Fatal("blank search submit returned a command")
	}
	if got := m.library.ActiveTab(); got != library.TabSearch {
		t.Fatalf("blank submit: want TabSearch, got %v", got)
	}
	if view := m.View(); !strings.Contains(view, "Search: ") {
		t.Fatalf("blank submit should keep input visible:\n%s", view)
	}
}
