package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotengine"
)

func browseReadyModel(engine *spotengine.Fake) RootModel {
	m := readyModel(engine)
	m.browseInitialized = true
	m.browseRoute = "liked"
	m.browseTitle = "Liked Tracks"
	m.navCursor = 1
	m.browseFocus = 1
	return m
}

func TestBrowseShellLoadsLikedTracksAndPlaysSelection(t *testing.T) {
	engine := spotengine.NewFake()
	m := browseReadyModel(engine)
	updated, _ := m.Update(msgs.CatalogLoadedMsg{Route: "liked", Data: spotengine.TrackPage{Items: []spotengine.Track{{URI: "spotify:track:hello", Name: "Hello", Artist: "Adele", DurationMS: 1000}}}})
	m = updated.(RootModel)
	if !strings.Contains(m.View(), "Liked Tracks") || !strings.Contains(m.View(), "Hello") || !strings.Contains(m.View(), "Library") {
		t.Fatalf("browse view:\n%s", m.View())
	}
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)
	if cmd == nil {
		t.Fatal("Enter did not create playback command")
	}
	_ = cmd()
	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Operation != spotengine.OperationPlay || calls[0].URI != "spotify:track:hello" {
		t.Fatalf("calls: %#v", calls)
	}
}

func TestBrowsePlaylistPreservesContextPlayback(t *testing.T) {
	engine := spotengine.NewFake()
	m := browseReadyModel(engine)
	m.browseRoute = "playlists"
	m.browseTitle = "Playlists"
	updated, _ := m.Update(msgs.CatalogLoadedMsg{Route: "playlists", Data: spotengine.CatalogPage[spotengine.PlaylistSummary]{Items: []spotengine.PlaylistSummary{{URI: "spotify:playlist:one", Name: "Mix"}}}})
	m = updated.(RootModel)
	updated, load := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)
	if load == nil {
		t.Fatal("playlist Enter did not load detail")
	}
	engine.SetPlaylistDetail("spotify:playlist:one", spotengine.PlaylistDetail{PlaylistSummary: spotengine.PlaylistSummary{URI: "spotify:playlist:one", Name: "Mix"}, Tracks: spotengine.TrackPage{Items: []spotengine.Track{{URI: "spotify:track:hello", Name: "Hello"}}}})
	updated, _ = m.Update(load())
	m = updated.(RootModel)
	updated, play := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)
	_ = play()
	calls := engine.Calls()
	if len(calls) != 2 || calls[1].Operation != spotengine.OperationPlayContext || calls[1].URI != "spotify:playlist:one" {
		t.Fatalf("calls: %#v", calls)
	}
}

func TestBrowseArtistDoesNotRenderEmptyAlbumRows(t *testing.T) {
	engine := spotengine.NewFake()
	m := browseReadyModel(engine)
	m.browseRoute = "artist:spotify:artist:arctic-monkeys"
	updated, _ := m.Update(msgs.CatalogLoadedMsg{
		Route: m.browseRoute,
		Data: spotengine.ArtistDetail{
			ArtistSummary: spotengine.ArtistSummary{Name: "Arctic Monkeys", ImageURL: "https://image.test/arctic"},
			Albums: []spotengine.AlbumSummary{
				{URI: "spotify:album:incomplete"},
				{URI: "spotify:album:am", Name: "AM", Artist: "Arctic Monkeys", ReleaseDate: "2013", TrackCount: 12},
			},
		},
	})
	m = updated.(RootModel)
	view := m.View()
	if strings.Contains(view, "0 tracks") || strings.Contains(view, "· ·") {
		t.Fatalf("artist view rendered an empty album row:\n%s", view)
	}
	if !strings.Contains(view, "AM") || !strings.Contains(view, "12 tracks") {
		t.Fatalf("artist view lost the valid album:\n%s", view)
	}
}
