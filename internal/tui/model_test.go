package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/library"
	"github.com/zmb3/spotify/v2"
)

func newReadyModel() RootModel {
	m := NewRootModel("")
	m.state = stateReady
	m.client = &spotify.Client{}
	return m
}

func sendKey(m RootModel, key string) RootModel {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
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
