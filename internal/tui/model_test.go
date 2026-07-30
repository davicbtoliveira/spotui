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
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/spotifyapi"
	"github.com/zmb3/spotify/v2"
)

type modelBrowser struct {
	urls []string
}

func (b *modelBrowser) Open(url string) error {
	b.urls = append(b.urls, url)
	return nil
}

func TestFirstStartWaitsOnLoggedOutScreen(t *testing.T) {
	engine := spotengine.NewFake()
	browser := &modelBrowser{}
	m := NewRootModel(engine, browser.Open)
	m.width = 80
	m.height = 24

	if cmd := m.Init(); cmd != nil {
		t.Fatal("first start returned a command")
	}
	view := m.View()
	for _, want := range []string{"SpotUI", "Spotify Premium", "Enter", "Log in"} {
		if !strings.Contains(view, want) {
			t.Fatalf("logged-out screen missing %q:\n%s", want, view)
		}
	}
	if len(browser.urls) != 0 {
		t.Fatalf("browser opened on first start: %v", browser.urls)
	}
	if calls := engine.Calls(); len(calls) != 0 {
		t.Fatalf("engine called on first start: %#v", calls)
	}
}

func TestEnterStartsExactlyOneLogin(t *testing.T) {
	engine := spotengine.NewFake()
	m := NewRootModel(engine, (&modelBrowser{}).Open)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)
	if cmd == nil {
		t.Fatal("Enter returned no Login command")
	}
	if _, ok := cmd().(msgs.EngineStartedMsg); !ok {
		t.Fatal("Login command did not start Playback Engine")
	}

	_, duplicate := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if duplicate != nil {
		t.Fatal("second Enter started another Login command")
	}
	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Operation != spotengine.OperationStart {
		t.Fatalf("engine calls: %#v", calls)
	}
}

func TestAuthorizationURLHookOpensBrowser(t *testing.T) {
	engine := spotengine.NewFake()
	browser := &modelBrowser{}
	m := NewRootModel(engine, browser.Open)

	updated, startCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)
	updated, waitCmd := m.Update(startCmd())
	m = updated.(RootModel)
	if waitCmd == nil {
		t.Fatal("engine start did not wait for authorization event")
	}

	const authURL = "https://accounts.spotify.com/authorize?code_challenge=test"
	engine.Emit(spotengine.Event{Type: spotengine.EventTypeAuthorizationURL, URL: authURL})
	eventMsg := waitCmd()
	if _, ok := eventMsg.(msgs.EngineEventMsg); !ok {
		t.Fatalf("event command: want EngineEventMsg, got %T", eventMsg)
	}

	updated, openCmd := m.Update(eventMsg)
	m = updated.(RootModel)
	if openCmd == nil {
		t.Fatal("authorization event returned no browser command")
	}
	if _, ok := openCmd().(msgs.BrowserOpenedMsg); !ok {
		t.Fatal("browser command did not report success")
	}
	if len(browser.urls) != 1 || browser.urls[0] != authURL {
		t.Fatalf("browser URLs: %v", browser.urls)
	}
}

func TestPendingAuthorizationShowsURLAndRetriesSameTransaction(t *testing.T) {
	engine := spotengine.NewFake()
	browser := &modelBrowser{}
	m := NewRootModel(engine, browser.Open)
	m.state = stateAuthenticating
	m.width = 120
	m.height = 24
	const authURL = "https://accounts.spotify.com/authorize?code_challenge=same"

	updated, openCmd := m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeAuthorizationURL, URL: authURL},
	})
	m = updated.(RootModel)
	view := m.View()
	for _, want := range []string{authURL, "r", "esc"} {
		if !strings.Contains(view, want) {
			t.Fatalf("authorization view missing %q:\n%s", want, view)
		}
	}
	if _, ok := openCmd().(msgs.BrowserOpenedMsg); !ok {
		t.Fatal("initial browser command did not report success")
	}

	updated, retryCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	m = updated.(RootModel)
	if retryCmd == nil {
		t.Fatal("retry returned no browser command")
	}
	if _, ok := retryCmd().(msgs.BrowserOpenedMsg); !ok {
		t.Fatal("retry browser command did not report success")
	}
	if len(browser.urls) != 2 || browser.urls[0] != authURL || browser.urls[1] != authURL {
		t.Fatalf("browser URLs: %v", browser.urls)
	}
	for _, call := range engine.Calls() {
		if call.Operation == spotengine.OperationStart {
			t.Fatalf("retry restarted engine transaction: %#v", engine.Calls())
		}
	}
}

func TestEscapeCancelsPendingLoginAndReturnsLoggedOut(t *testing.T) {
	engine := spotengine.NewFake()
	m := NewRootModel(engine, (&modelBrowser{}).Open)
	m.state = stateAuthenticating

	updated, cancelCmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(RootModel)
	if cancelCmd == nil {
		t.Fatal("escape returned no cancellation command")
	}
	updated, _ = m.Update(cancelCmd())
	m = updated.(RootModel)

	if m.state != stateLoggedOut {
		t.Fatalf("state: want logged out, got %v", m.state)
	}
	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Operation != spotengine.OperationCancelLogin {
		t.Fatalf("engine calls: %#v", calls)
	}
}

func TestBrowserFailureKeepsLoginRetryable(t *testing.T) {
	engine := spotengine.NewFake()
	m := NewRootModel(engine, func(string) error { return errors.New("no browser") })
	m.state = stateAuthenticating
	m.width = 100
	m.height = 24
	m.authURL = "https://accounts.spotify.com/authorize?test"

	updated, waitCmd := m.Update(msgs.BrowserOpenErrMsg{Err: errors.New("no browser")})
	m = updated.(RootModel)

	if m.state != stateAuthenticating {
		t.Fatalf("state: want authenticating, got %v", m.state)
	}
	if waitCmd == nil {
		t.Fatal("browser failure stopped listening for callback")
	}
	view := m.View()
	for _, want := range []string{"no browser", m.authURL, "r"} {
		if !strings.Contains(view, want) {
			t.Fatalf("browser failure view missing %q:\n%s", want, view)
		}
	}
}

func TestLoginFailureReturnsRetryableLoggedOutScreen(t *testing.T) {
	engine := spotengine.NewFake()
	m := NewRootModel(engine, (&modelBrowser{}).Open)
	m.state = stateAuthenticating
	m.width = 80
	m.height = 24

	updated, resetCmd := m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeError, Err: errors.New("access denied")},
	})
	m = updated.(RootModel)
	if resetCmd == nil {
		t.Fatal("login failure returned no reset command")
	}
	updated, quitCmd := m.Update(resetCmd())
	m = updated.(RootModel)

	if quitCmd != nil {
		t.Fatal("login failure quit the application")
	}
	if m.state != stateLoggedOut {
		t.Fatalf("state: want logged out, got %v", m.state)
	}
	view := m.View()
	for _, want := range []string{"access denied", "Enter"} {
		if !strings.Contains(view, want) {
			t.Fatalf("failure view missing %q:\n%s", want, view)
		}
	}
}

func TestFreeAccountBlocksPlayerUntilLogout(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetHasSession(true)
	m := NewRootModel(engine, (&modelBrowser{}).Open)
	m.width = 80
	m.height = 24

	updated, cmd := m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeAccountProduct, Product: "free"},
	})
	m = updated.(RootModel)
	if cmd != nil {
		t.Fatal("unsupported account kept entering player flow")
	}
	if m.state != stateUnsupported {
		t.Fatalf("state: want unsupported, got %v", m.state)
	}
	view := m.View()
	for _, want := range []string{"Free", "Premium", "L", "q"} {
		if !strings.Contains(view, want) {
			t.Fatalf("unsupported view missing %q:\n%s", want, view)
		}
	}

	updated, blockedCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)
	if blockedCmd != nil || m.state != stateUnsupported {
		t.Fatal("unsupported account accepted player input")
	}
	updated, logoutCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")})
	m = updated.(RootModel)
	if logoutCmd == nil {
		t.Fatal("L returned no logout command")
	}
	updated, _ = m.Update(logoutCmd())
	m = updated.(RootModel)
	if m.state != stateLoggedOut || engine.HasSession() {
		t.Fatalf("logout state=%v hasSession=%v", m.state, engine.HasSession())
	}
}

func TestSuccessfulPremiumLoginReachesReadyState(t *testing.T) {
	engine := spotengine.NewFake()
	m := NewRootModel(engine, (&modelBrowser{}).Open)
	m.state = stateAuthenticating
	m.width = 80
	m.height = 24

	updated, cmd := m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeReady},
	})
	m = updated.(RootModel)

	if m.state != stateReady {
		t.Fatalf("state: want ready, got %v", m.state)
	}
	if cmd == nil {
		t.Fatal("ready state stopped listening for engine events")
	}
	if view := m.View(); strings.Contains(view, "Waiting for Spotify authorization") {
		t.Fatalf("ready view remains authenticating:\n%s", view)
	}
	for _, call := range engine.Calls() {
		if call.Operation == spotengine.OperationPlay {
			t.Fatalf("successful Login resumed audio: %#v", engine.Calls())
		}
	}
}

func TestValidLocalSessionRestoresWithoutBrowserOrAudio(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetHasSession(true)
	browser := &modelBrowser{}
	m := NewRootModel(engine, browser.Open)

	startCmd := m.Init()
	if startCmd == nil {
		t.Fatal("valid Local Session did not start restoration")
	}
	updated, waitCmd := m.Update(startCmd())
	m = updated.(RootModel)
	engine.Emit(spotengine.Event{Type: spotengine.EventTypeReady})
	updated, _ = m.Update(waitCmd())
	m = updated.(RootModel)

	if m.state != stateReady {
		t.Fatalf("state: want ready, got %v", m.state)
	}
	if len(browser.urls) != 0 {
		t.Fatalf("restoration opened browser: %v", browser.urls)
	}
	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Operation != spotengine.OperationStart {
		t.Fatalf("restoration engine calls: %#v", calls)
	}
}

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
	m := NewRootModel(spotengine.NewFake(), func(string) error { return nil })
	m.state = stateReady
	m.client = &spotify.Client{}
	m.width = 80
	m.height = 24
	return m
}

func TestReadyEngineSearchesAndRendersStableResults(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetSearchResult(spotengine.SearchPage{
		Tracks: []spotengine.Track{{
			URI:        "spotify:track:hello",
			Name:       "Hello",
			Artist:     "Adele",
			DurationMS: 295000,
		}},
		Total: 1,
	}, nil)
	m := NewRootModel(engine, func(string) error { return nil })
	m.state = stateReady
	m.library.SetActiveTab(library.TabSearch)
	m.width = 80
	m.height = 24

	m = sendKey(m, "/")
	for _, key := range []string{"h", "e", "l", "l", "o"} {
		m = sendKey(m, key)
	}
	updated, searchCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)
	if searchCmd == nil {
		t.Fatal("search returned no command")
	}
	loaded := searchCmd()
	if _, ok := loaded.(msgs.EngineTrackSearchLoadedMsg); !ok {
		t.Fatalf("message: want EngineTrackSearchLoadedMsg, got %T", loaded)
	}
	updated, _ = m.Update(loaded)
	m = updated.(RootModel)

	view := m.View()
	for _, want := range []string{"Hello", "Adele", "04:55"} {
		if !strings.Contains(view, want) {
			t.Fatalf("search view missing %q:\n%s", want, view)
		}
	}
	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Operation != spotengine.OperationSearchTracks {
		t.Fatalf("engine calls: %#v", calls)
	}
}

func TestEnterOnEngineSearchResultStartsLocalPlayback(t *testing.T) {
	engine := spotengine.NewFake()
	m := NewRootModel(engine, func(string) error { return nil })
	m.state = stateReady
	m.library.SetActiveTab(library.TabSearch)
	updated, _ := m.Update(msgs.EngineTrackSearchLoadedMsg{
		Tracks: []spotengine.Track{{
			URI:  "spotify:track:hello",
			Name: "Hello",
		}},
		Total: 1,
	})
	m = updated.(RootModel)

	_, playCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if playCmd == nil {
		t.Fatal("Enter returned no playback command")
	}
	if msg := playCmd(); msg != nil {
		if errMsg, ok := msg.(msgs.ErrMsg); ok {
			t.Fatalf("playback error: %v", errMsg.Err)
		}
	}

	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Operation != spotengine.OperationPlay ||
		calls[0].URI != "spotify:track:hello" {
		t.Fatalf("engine calls: %#v", calls)
	}
}

func TestPlaybackEngineEventsDriveVisiblePlayerState(t *testing.T) {
	m := NewRootModel(spotengine.NewFake(), func(string) error { return nil })
	m.state = stateReady
	m.width = 80
	m.height = 24

	events := []spotengine.Event{
		{
			Type:       spotengine.EventTypeMetadata,
			Track:      &spotengine.Track{Name: "Hello", Artist: "Adele", DurationMS: 295000},
			PositionMS: 12000,
			DurationMS: 295000,
		},
		{Type: spotengine.EventTypeBuffering},
	}
	for _, event := range events {
		updated, _ := m.Update(msgs.EngineEventMsg{Event: event})
		m = updated.(RootModel)
	}
	view := m.View()
	for _, want := range []string{"Hello", "Adele", "Buffering", "00:12/04:55"} {
		if !strings.Contains(view, want) {
			t.Fatalf("buffering player missing %q:\n%s", want, view)
		}
	}

	updated, _ := m.Update(msgs.EngineEventMsg{Event: spotengine.Event{Type: spotengine.EventTypePlaying}})
	m = updated.(RootModel)
	updated, _ = m.Update(ProgressTickMsg{})
	m = updated.(RootModel)
	view = m.View()
	for _, want := range []string{"Playing", "00:13/04:55"} {
		if !strings.Contains(view, want) {
			t.Fatalf("playing player missing %q:\n%s", want, view)
		}
	}

	updated, _ = m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeError, Err: errors.New("check default audio output")},
	})
	m = updated.(RootModel)
	if view = m.View(); !strings.Contains(view, "check default audio output") {
		t.Fatalf("playback failure not actionable:\n%s", view)
	}
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
