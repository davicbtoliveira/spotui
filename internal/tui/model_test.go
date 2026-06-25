package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/library"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotifyapi"
	"github.com/zmb3/spotify/v2"
)

type modelTrackSearcher struct {
	req spotifyapi.TrackSearchRequest
}

func (s *modelTrackSearcher) SearchTracks(_ context.Context, req spotifyapi.TrackSearchRequest) (spotifyapi.TrackSearchPage, error) {
	s.req = req
	return spotifyapi.TrackSearchPage{
		Tracks: []spotify.FullTrack{
			{SimpleTrack: spotify.SimpleTrack{Name: "Hello", URI: "spotify:track:hello"}},
		},
		Total:  1,
		Offset: req.Offset,
	}, nil
}

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

func TestTrackSearchSubmitStartsLoading(t *testing.T) {
	m := newReadyModel()

	m = sendKey(m, "/")
	m = sendKey(m, "a")
	m = sendKey(m, "d")
	m = sendKey(m, "e")
	m = sendKey(m, "l")
	m = sendKey(m, "e")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)

	if cmd == nil {
		t.Fatal("track search submit returned no command")
	}
	view := m.View()
	if strings.Contains(view, "Search: adele") {
		t.Fatalf("search input still visible after submit:\n%s", view)
	}
	if !strings.Contains(view, "Searching tracks...") {
		t.Fatalf("search loading state missing from view:\n%s", view)
	}
}

func TestTrackSearchSubmitDispatchesSearchCommand(t *testing.T) {
	searcher := &modelTrackSearcher{}
	m := newReadyModel()
	m.trackSearcher = searcher

	m = sendKey(m, "/")
	m = sendKey(m, "h")
	m = sendKey(m, "i")
	m = sendSpecialKey(m, tea.KeySpace)
	m = sendKey(m, "a")
	m = sendKey(m, "d")
	m = sendKey(m, "e")
	m = sendKey(m, "l")
	m = sendKey(m, "e")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	msg := cmd()
	if searcher.req.Query != "hi adele" {
		t.Fatalf("Query: want raw query, got %q", searcher.req.Query)
	}
	if _, ok := msg.(msgs.TrackSearchLoadedMsg); !ok {
		t.Fatalf("message: want TrackSearchLoadedMsg, got %T", msg)
	}
}

func TestTrackSearchResultsRenderLikeTracks(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)
	m.searchLoading = true

	updated, _ := m.Update(TrackSearchLoadedMsg{Tracks: []spotify.FullTrack{
		{
			SimpleTrack: spotify.SimpleTrack{
				Name:     "Hello",
				Artists:  []spotify.SimpleArtist{{Name: "Adele"}},
				Duration: 295000,
				URI:      "spotify:track:hello",
			},
		},
	}})
	m = updated.(RootModel)

	view := m.View()
	if strings.Contains(view, "Searching tracks...") {
		t.Fatalf("search loading still visible after results:\n%s", view)
	}
	for _, want := range []string{"Hello", "Adele", "04:55"} {
		if !strings.Contains(view, want) {
			t.Fatalf("search result missing %q from view:\n%s", want, view)
		}
	}
}

func TestTrackSearchKeepsPriorResultsWhileLoading(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)

	updated, _ := m.Update(TrackSearchLoadedMsg{Tracks: []spotify.FullTrack{
		{
			SimpleTrack: spotify.SimpleTrack{
				Name:     "Old Result",
				Artists:  []spotify.SimpleArtist{{Name: "Adele"}},
				Duration: 180000,
				URI:      "spotify:track:old",
			},
		},
	}})
	m = updated.(RootModel)

	m = sendKey(m, "/")
	m = sendKey(m, "n")
	m = sendKey(m, "e")
	m = sendKey(m, "w")
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)

	view := m.View()
	if !strings.Contains(view, "Searching tracks...") {
		t.Fatalf("loading state missing from view:\n%s", view)
	}
	if !strings.Contains(view, "Old Result") {
		t.Fatalf("prior result missing while loading:\n%s", view)
	}
}

func TestEmptyTrackSearchShowsNoTracksFound(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)
	m.searchLoading = true

	updated, _ := m.Update(TrackSearchLoadedMsg{})
	m = updated.(RootModel)

	view := m.View()
	if strings.Contains(view, "Searching tracks...") {
		t.Fatalf("search loading still visible after empty results:\n%s", view)
	}
	if !strings.Contains(view, "No tracks found") {
		t.Fatalf("empty search state missing from view:\n%s", view)
	}
}

func TestTrackSearchErrorClearsLoadingAndShowsStatus(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)
	m.searchLoading = true

	updated, _ := m.Update(ErrMsg{Err: errors.New("spotify unavailable"), Context: "search tracks"})
	m = updated.(RootModel)

	view := m.View()
	if strings.Contains(view, "Searching tracks...") {
		t.Fatalf("search loading still visible after error:\n%s", view)
	}
	if !strings.Contains(view, "search tracks: spotify unavailable") {
		t.Fatalf("search error missing from status:\n%s", view)
	}
}

func TestEnterOnSearchResultStartsPlaybackCommand(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)
	updated, _ := m.Update(TrackSearchLoadedMsg{Tracks: []spotify.FullTrack{
		{SimpleTrack: spotify.SimpleTrack{Name: "Hello", URI: "spotify:track:hello"}},
	}})
	m = updated.(RootModel)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("enter on search result returned no playback command")
	}
}

func makeSearchTracks(n int) []spotify.FullTrack {
	tracks := make([]spotify.FullTrack, n)
	for i := range tracks {
		tracks[i] = spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				Name:     fmt.Sprintf("Track %d", i),
				Duration: 200000,
				URI:      spotify.URI(fmt.Sprintf("spotify:track:%d", i)),
			},
		}
	}
	return tracks
}

func sendKeyAndGetCmd(m RootModel, key string) (RootModel, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(RootModel), cmd
}

type paginationTrackSearcher struct {
	callCount int
	offsets   []int
	pageSize  int
	total     int
}

func (s *paginationTrackSearcher) SearchTracks(_ context.Context, req spotifyapi.TrackSearchRequest) (spotifyapi.TrackSearchPage, error) {
	s.offsets = append(s.offsets, req.Offset)
	s.callCount++
	tracks := make([]spotify.FullTrack, s.pageSize)
	for i := range tracks {
		tracks[i] = spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				Name:     fmt.Sprintf("Page%d-Track%d", req.Offset/s.pageSize, i),
				Duration: 200000,
				URI:      spotify.URI(fmt.Sprintf("spotify:track:page%d_%d", req.Offset/s.pageSize, i)),
			},
		}
	}
	return spotifyapi.TrackSearchPage{
		Tracks: tracks,
		Total:  s.total,
		Offset: req.Offset,
	}, nil
}

func TestSearchNextPageDispatchesCommand(t *testing.T) {
	searcher := &paginationTrackSearcher{pageSize: 10, total: 30}
	m := newReadyModel()
	m.trackSearcher = searcher
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30
	m.library.SetSearchResults(make([]library.TrackEntry, 10))

	_, cmd := sendKeyAndGetCmd(m, "]")

	if cmd == nil {
		t.Fatal("] on search results: want a command, got nil")
	}
	msg := cmd()
	loaded, ok := msg.(msgs.TrackSearchLoadedMsg)
	if !ok {
		t.Fatalf("] command: want TrackSearchLoadedMsg, got %T", msg)
	}
	if loaded.Offset != 10 {
		t.Fatalf("next page offset: want 10, got %d", loaded.Offset)
	}
	if len(searcher.offsets) != 1 || searcher.offsets[0] != 10 {
		t.Fatalf("searcher offset: want [10], got %v", searcher.offsets)
	}
}

func TestSearchPrevPageDispatchesCommand(t *testing.T) {
	searcher := &paginationTrackSearcher{pageSize: 10, total: 30}
	m := newReadyModel()
	m.trackSearcher = searcher
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 20
	m.searchTotal = 30
	m.library.SetSearchResults(make([]library.TrackEntry, 10))

	_, cmd := sendKeyAndGetCmd(m, "[")

	if cmd == nil {
		t.Fatal("[ on search results: want a command, got nil")
	}
	msg := cmd()
	loaded, ok := msg.(msgs.TrackSearchLoadedMsg)
	if !ok {
		t.Fatalf("[ command: want TrackSearchLoadedMsg, got %T", msg)
	}
	if loaded.Offset != 10 {
		t.Fatalf("prev page offset: want 10, got %d", loaded.Offset)
	}
	if len(searcher.offsets) != 1 || searcher.offsets[0] != 10 {
		t.Fatalf("searcher offset: want [10], got %v", searcher.offsets)
	}
}

func TestSearchNextPageNoOpWhenOnLastPage(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 20
	m.searchTotal = 30
	m.library.SetSearchResults(make([]library.TrackEntry, 10))

	_, cmd := sendKeyAndGetCmd(m, "]")

	if cmd != nil {
		t.Fatal("] on last page: should not dispatch a command")
	}
}

func TestSearchPrevPageNoOpWhenOnFirstPage(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30
	m.library.SetSearchResults(make([]library.TrackEntry, 10))

	_, cmd := sendKeyAndGetCmd(m, "[")

	if cmd != nil {
		t.Fatal("[ on first page: should not dispatch a command")
	}
}

func TestSearchNextPageNoOpBeforeAnySearch(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)

	_, cmd := sendKeyAndGetCmd(m, "]")

	if cmd != nil {
		t.Fatal("] before search: should not dispatch a command")
	}
}

func TestSearchPrevPageNoOpBeforeAnySearch(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)

	_, cmd := sendKeyAndGetCmd(m, "[")

	if cmd != nil {
		t.Fatal("[ before search: should not dispatch a command")
	}
}

func TestSearchPaginationSetsLoadingState(t *testing.T) {
	searcher := &paginationTrackSearcher{pageSize: 10, total: 30}
	m := newReadyModel()
	m.trackSearcher = searcher
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30
	m.library.SetSearchResults(make([]library.TrackEntry, 10))

	updated, _ := sendKeyAndGetCmd(m, "]")
	m = updated

	if !m.searchLoading {
		t.Fatal("searchLoading should be true after pressing ]")
	}
	view := m.View()
	if !strings.Contains(view, "Loading more tracks...") {
		t.Fatalf("loading state should show 'Loading more tracks...':\n%s", view)
	}
}

func TestSearchPaginationKeepsPriorResultsWhileLoading(t *testing.T) {
	searcher := &paginationTrackSearcher{pageSize: 10, total: 30}
	m := newReadyModel()
	m.trackSearcher = searcher
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30

	tracks := make([]library.TrackEntry, 10)
	for i := range tracks {
		tracks[i] = library.TrackEntry{Name: fmt.Sprintf("Old %d", i)}
	}
	m.library.SetSearchResults(tracks)

	updated, _ := sendKeyAndGetCmd(m, "]")
	m = updated

	view := m.View()
	if !strings.Contains(view, "Old 0") {
		t.Fatalf("prior results should remain visible while loading:\n%s", view)
	}
}

func TestSearchPaginationReplacesResults(t *testing.T) {
	searcher := &paginationTrackSearcher{pageSize: 10, total: 30}
	m := newReadyModel()
	m.trackSearcher = searcher
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30
	m.library.SetSearchResults(make([]library.TrackEntry, 10))

	updated, cmd := sendKeyAndGetCmd(m, "]")
	m = updated

	msg := cmd()
	next, _ := m.Update(msg)
	m = next.(RootModel)

	if m.library.SearchResultCount() != 10 {
		t.Fatalf("after page load: want 10 results, got %d", m.library.SearchResultCount())
	}
	if m.searchOffset != 10 {
		t.Fatalf("after page load: want offset 10, got %d", m.searchOffset)
	}
}

func TestSearchPaginationPreservesCursorWhenPossible(t *testing.T) {
	searcher := &paginationTrackSearcher{pageSize: 10, total: 30}
	m := newReadyModel()
	m.trackSearcher = searcher
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30

	tracks := make([]library.TrackEntry, 10)
	for i := range tracks {
		tracks[i] = library.TrackEntry{Name: fmt.Sprintf("Track %d", i)}
	}
	m.library.SetSearchResults(tracks)
	m.library.MoveDown()

	updated, cmd := sendKeyAndGetCmd(m, "]")
	msg := cmd()
	next, _ := updated.Update(msg)
	m = next.(RootModel)

	if m.library.Cursor() != 1 {
		t.Fatalf("cursor after page change: want 1 (same index), got %d", m.library.Cursor())
	}
}

func TestSearchPaginationClampsCursorWhenNewPageSmaller(t *testing.T) {
	searcher := &paginationTrackSearcher{pageSize: 10, total: 30}
	m := newReadyModel()
	m.trackSearcher = searcher
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30

	tracks := make([]library.TrackEntry, 10)
	for i := range tracks {
		tracks[i] = library.TrackEntry{Name: fmt.Sprintf("Track %d", i)}
	}
	m.library.SetSearchResults(tracks)

	// Move cursor to 9 (last item on page)
	for i := 0; i < 9; i++ {
		m.library.MoveDown()
	}
	if m.library.Cursor() != 9 {
		t.Fatalf("setup: want cursor 9, got %d", m.library.Cursor())
	}

	// Next page command returns page with only 5 tracks (simulate last page)
	tracks5 := make([]spotify.FullTrack, 5)
	for i := range tracks5 {
		tracks5[i] = spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				Name: fmt.Sprintf("Track %d", i),
				URI:  spotify.URI(fmt.Sprintf("spotify:track:%d", i)),
			},
		}
	}
	msg := TrackSearchLoadedMsg{Tracks: tracks5, Total: 15, Offset: 10}
	updated, _ := m.Update(msg)
	m = updated.(RootModel)

	if m.library.Cursor() != 4 {
		t.Fatalf("cursor clamp: want 4 (last index of 5), got %d", m.library.Cursor())
	}
}

func TestSearchPaginationErrorShowsStatus(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30
	m.library.SetSearchResults(make([]library.TrackEntry, 10))
	m.searchLoading = true

	updated, _ := m.Update(ErrMsg{Err: errors.New("pagination failed"), Context: "search tracks"})
	m = updated.(RootModel)

	if m.searchLoading {
		t.Fatal("searchLoading should be false after error")
	}
	view := m.View()
	if !strings.Contains(view, "search tracks: pagination failed") {
		t.Fatalf("pagination error missing from status:\n%s", view)
	}
}

func TestSearchNextPageDoesNotFireDuringLoading(t *testing.T) {
	searcher := &paginationTrackSearcher{pageSize: 10, total: 30}
	m := newReadyModel()
	m.trackSearcher = searcher
	m.library.SetActiveTab(library.TabSearch)
	m.searchCompleted = true
	m.searchOffset = 0
	m.searchTotal = 30
	m.library.SetSearchResults(make([]library.TrackEntry, 10))
	m.searchLoading = true

	_, cmd := sendKeyAndGetCmd(m, "]")

	if cmd != nil {
		t.Fatal("] during loading: should not dispatch a command")
	}
}

func TestSearchResultCursorMovesWithNavigationKeys(t *testing.T) {
	m := newReadyModel()
	m.library.SetActiveTab(library.TabSearch)
	updated, _ := m.Update(TrackSearchLoadedMsg{Tracks: []spotify.FullTrack{
		{SimpleTrack: spotify.SimpleTrack{Name: "First", URI: "spotify:track:first"}},
		{SimpleTrack: spotify.SimpleTrack{Name: "Second", URI: "spotify:track:second"}},
	}})
	m = updated.(RootModel)

	m = sendKey(m, "j")

	if got := m.library.SelectedURI(); got != "spotify:track:second" {
		t.Fatalf("selected after j: want second track, got %q", got)
	}

	m = sendKey(m, "k")

	if got := m.library.SelectedURI(); got != "spotify:track:first" {
		t.Fatalf("selected after k: want first track, got %q", got)
	}
}
