package tui

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/catalog"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/dcbto/spotui/internal/spotengine"
)

type modelBrowser struct {
	urls []string
	err  error
}

func (b *modelBrowser) Open(url string) error {
	b.urls = append(b.urls, url)
	return b.err
}

type recoveryClock struct {
	delays []time.Duration
}

func (c *recoveryClock) Wait(delay time.Duration) tea.Cmd {
	c.delays = append(c.delays, delay)
	return func() tea.Msg { return ReconnectTimerMsg{} }
}

func readyModel(engine *spotengine.Fake) RootModel {
	m := NewRootModel(engine, func(string) error { return nil })
	m.state = stateReady
	m.browseInitialized = true
	m.width = 90
	m.height = 24
	return m
}

func key(m RootModel, value string) (RootModel, tea.Cmd) {
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)})
	return updated.(RootModel), cmd
}

func TestLoggedOutScreenWaitsForExplicitLogin(t *testing.T) {
	engine := spotengine.NewFake()
	browser := &modelBrowser{}
	m := NewRootModel(engine, browser.Open)
	m.width, m.height = 80, 24

	if m.Init() == nil || len(engine.Calls()) != 0 || len(browser.urls) != 0 {
		t.Fatal("initial screen started Login")
	}
	for _, want := range []string{"Spotify Premium", "Enter", "Log in"} {
		if !strings.Contains(m.View(), want) {
			t.Fatalf("logged-out screen missing %q:\n%s", want, m.View())
		}
	}

	updated, startCmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(RootModel)
	if _, ok := startCmd().(msgs.EngineStartedMsg); !ok || m.state != stateAuthenticating {
		t.Fatal("Enter did not start Login")
	}
}

func TestReadyEventInitializesBrowseRoute(t *testing.T) {
	engine := spotengine.NewFake()
	m := NewRootModel(engine, func(string) error { return nil })
	m.state = stateLoading

	updated, cmd := m.Update(msgs.EngineEventMsg{Event: spotengine.Event{Type: spotengine.EventTypeReady}})
	m = updated.(RootModel)
	if m.state != stateReady || !m.browseInitialized || m.browseRoute.Kind != catalog.RouteLiked ||
		m.browseTitle != "Liked Tracks" || !m.browseLoading || cmd == nil {
		t.Fatalf("ready transition: state=%v initialized=%v route=%#v title=%q loading=%v cmd=%v", m.state, m.browseInitialized, m.browseRoute, m.browseTitle, m.browseLoading, cmd != nil)
	}
}

func TestPlaybackProgressStartsTickingAndAdvancesWhilePlaying(t *testing.T) {
	engine := spotengine.NewFake()
	m := NewRootModel(engine, func(string) error { return nil })
	m.state = stateReady
	m.width, m.height = 90, 24
	m.engineTrack = &spotengine.Track{Name: "Hello", DurationMS: 100000}
	m.enginePlaying = true

	if m.Init() == nil {
		t.Fatal("initialization did not start the progress clock")
	}

	updated, nextTick := m.Update(msgs.ProgressTickMsg{})
	m = updated.(RootModel)
	if m.localProgressMs != 1000 {
		t.Fatalf("progress = %dms, want 1000ms", m.localProgressMs)
	}
	if !strings.Contains(m.View(), "00:01/01:40") {
		t.Fatalf("visible progress did not advance:\n%s", m.View())
	}
	if nextTick == nil {
		t.Fatal("progress clock did not schedule the next tick")
	}
}

func TestPlaybackProgressDoesNotAdvanceWhilePaused(t *testing.T) {
	engine := spotengine.NewFake()
	m := readyModel(engine)
	m.engineTrack = &spotengine.Track{Name: "Hello", DurationMS: 100000}
	m.localProgressMs = 12000
	m.enginePlaying = false

	updated, _ := m.Update(msgs.ProgressTickMsg{})
	m = updated.(RootModel)
	if m.localProgressMs != 12000 {
		t.Fatalf("paused progress = %dms, want 12000ms", m.localProgressMs)
	}
}

func TestPlaybackProgressStopsAtTrackDuration(t *testing.T) {
	engine := spotengine.NewFake()
	m := readyModel(engine)
	m.engineTrack = &spotengine.Track{Name: "Hello", DurationMS: 12500}
	m.localProgressMs = 12000
	m.enginePlaying = true

	updated, _ := m.Update(msgs.ProgressTickMsg{})
	m = updated.(RootModel)
	if m.localProgressMs != 12500 {
		t.Fatalf("progress = %dms, want track duration 12500ms", m.localProgressMs)
	}
}

func TestPendingLoginShowsURLRetriesAndCancels(t *testing.T) {
	engine := spotengine.NewFake()
	browser := &modelBrowser{}
	m := NewRootModel(engine, browser.Open)
	m.state = stateAuthenticating
	m.width, m.height = 120, 24
	const authURL = "https://accounts.spotify.com/authorize?code_challenge=test"

	updated, openCmd := m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeAuthorizationURL, URL: authURL},
	})
	m = updated.(RootModel)
	if !strings.Contains(m.View(), authURL) {
		t.Fatalf("authorization URL missing:\n%s", m.View())
	}
	_ = openCmd()
	m, openCmd = key(m, "r")
	_ = openCmd()
	if len(browser.urls) != 2 || browser.urls[0] != authURL || browser.urls[1] != authURL {
		t.Fatalf("browser URLs: %v", browser.urls)
	}

	updated, cancelCmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(RootModel)
	updated, _ = m.Update(cancelCmd())
	m = updated.(RootModel)
	if m.state != stateLoggedOut {
		t.Fatalf("cancel state: %v", m.state)
	}
}

func TestBrowserFailureKeepsLoginRetryable(t *testing.T) {
	engine := spotengine.NewFake()
	browser := &modelBrowser{err: errors.New("no browser")}
	m := NewRootModel(engine, browser.Open)
	m.state = stateAuthenticating
	m.authURL = "https://accounts.spotify.com/authorize?test"
	m.width, m.height = 100, 24

	updated, waitCmd := m.Update(msgs.BrowserOpenErrMsg{Err: browser.err})
	m = updated.(RootModel)
	if m.state != stateAuthenticating || waitCmd == nil ||
		!strings.Contains(m.View(), "no browser") || !strings.Contains(m.View(), m.authURL) {
		t.Fatalf("browser recovery:\n%s", m.View())
	}
}

func TestBlockedSpotifyURLShowsFriendlyLoginError(t *testing.T) {
	engine := spotengine.NewFake()
	blocked := fmt.Errorf("failed getting endpoints from resolver: %w", &url.Error{
		Op:  "Get",
		URL: "https://apresolve.spotify.com/",
		Err: errors.New("access denied"),
	})
	engine.SetError(spotengine.OperationCancelLogin, blocked)
	m := NewRootModel(engine, func(string) error { return nil })
	m.state = stateAuthenticating
	m.width, m.height = 100, 24

	updated, resetCmd := m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeError, Err: blocked},
	})
	m = updated.(RootModel)
	updated, _ = m.Update(resetCmd())
	m = updated.(RootModel)

	const want = "Could not access Spotify. Check your connection and try again."
	if !strings.Contains(m.View(), want) {
		t.Fatalf("friendly connection error missing:\n%s", m.View())
	}
	if strings.Contains(m.View(), "apresolve") || strings.Contains(m.View(), "Could not reset login") {
		t.Fatalf("technical login error leaked:\n%s", m.View())
	}
}

func TestFreeAccountBlocksPlayerUntilConfirmedLogout(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetHasSession(true)
	m := NewRootModel(engine, func(string) error { return nil })
	m.width, m.height = 80, 24

	updated, _ := m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeAccountProduct, Product: "free"},
	})
	m = updated.(RootModel)
	if m.state != stateUnsupported || !strings.Contains(m.View(), "Spotify Free") {
		t.Fatalf("unsupported state: %v\n%s", m.state, m.View())
	}
	m, cmd := key(m, "L")
	if cmd != nil || !m.confirmingLogout {
		t.Fatal("L skipped Logout confirmation")
	}
	m, cmd = key(m, "y")
	updated, _ = m.Update(cmd())
	m = updated.(RootModel)
	if m.state != stateLoggedOut || engine.HasSession() {
		t.Fatalf("logout state=%v session=%v", m.state, engine.HasSession())
	}
}

func TestPlaybackEventsDrivePlayerAndControls(t *testing.T) {
	engine := spotengine.NewFake()
	m := readyModel(engine)
	for _, event := range []spotengine.Event{
		{Type: spotengine.EventTypeMetadata, Track: &spotengine.Track{Name: "Hello", Artist: "Adele", DurationMS: 295000}, PositionMS: 12000},
		{Type: spotengine.EventTypeBuffering},
		{Type: spotengine.EventTypePlaying},
		{Type: spotengine.EventTypeVolume, Volume: 65, VolumeMax: 100},
		{Type: spotengine.EventTypeActive},
	} {
		updated, _ := m.Update(msgs.EngineEventMsg{Event: event})
		m = updated.(RootModel)
	}
	for _, want := range []string{"Hello", "Adele", "Playing", "Vol 65%", "00:12/04:55"} {
		if !strings.Contains(m.View(), want) {
			t.Fatalf("player missing %q:\n%s", want, m.View())
		}
	}

	var cmd tea.Cmd
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(RootModel)
	_ = cmd()
	m, cmd = key(m, "n")
	_ = cmd()
	m, cmd = key(m, "p")
	_ = cmd()
	m, cmd = key(m, "+")
	updated, cmd = m.Update(msgs.VolumeDebounceElapsedMsg{Generation: m.volumeDebounceID})
	m = updated.(RootModel)
	_ = cmd()
	want := []spotengine.Operation{
		spotengine.OperationPause, spotengine.OperationNext,
		spotengine.OperationPrevious, spotengine.OperationSetVolume,
	}
	calls := engine.Calls()
	for i, operation := range want {
		if calls[i].Operation != operation {
			t.Fatalf("call %d: want %q, got %#v", i, operation, calls[i])
		}
	}
	if calls[3].Volume != 70 {
		t.Fatalf("volume: want 70, got %d", calls[3].Volume)
	}
}

func TestRapidVolumeKeysCoalesceToTheLatestTarget(t *testing.T) {
	engine := spotengine.NewFake()
	m := readyModel(engine)
	m.engineVolume = 90
	m.confirmedEngineVolume = 90

	m, first := key(m, "-")
	if first == nil {
		t.Fatal("first volume key did not schedule a command")
	}
	m, second := key(m, "-")
	if second == nil {
		t.Fatal("second volume key did not replace the pending command")
	}
	m, third := key(m, "-")
	if third == nil {
		t.Fatal("third volume key did not replace the pending command")
	}
	if m.engineVolume != 75 {
		t.Fatalf("optimistic volume = %d, want 75", m.engineVolume)
	}
	if len(engine.Calls()) != 0 {
		t.Fatal("volume was sent before the debounce elapsed")
	}

	updated, next := m.Update(msgs.VolumeDebounceElapsedMsg{Generation: m.volumeDebounceID})
	m = updated.(RootModel)
	if next == nil {
		t.Fatal("latest volume target was not sent after the debounce elapsed")
	}
	updated, final := m.Update(next())
	m = updated.(RootModel)
	if final != nil {
		t.Fatal("volume command remained in flight after reaching the latest target")
	}

	calls := engine.Calls()
	if len(calls) != 1 || calls[0].Volume != 75 {
		t.Fatalf("volume calls = %#v, want only 75", calls)
	}
}

func TestVolumeChangeInFlightSendsOnlyTheLatestFollowUpTarget(t *testing.T) {
	engine := spotengine.NewFake()
	m := readyModel(engine)
	m.engineVolume = 90
	m.confirmedEngineVolume = 90

	m, _ = key(m, "-")
	updated, firstSet := m.Update(msgs.VolumeDebounceElapsedMsg{Generation: m.volumeDebounceID})
	m = updated.(RootModel)

	m, _ = key(m, "-")
	m, _ = key(m, "-")
	updated, pending := m.Update(msgs.VolumeDebounceElapsedMsg{Generation: m.volumeDebounceID})
	m = updated.(RootModel)
	if pending != nil {
		t.Fatal("volume command started while the previous command was in flight")
	}

	updated, latestSet := m.Update(firstSet())
	m = updated.(RootModel)
	updated, final := m.Update(latestSet())
	m = updated.(RootModel)
	if final != nil || m.volumeCommandInFlight {
		t.Fatal("volume command remained in flight after applying the latest target")
	}

	calls := engine.Calls()
	if len(calls) != 2 || calls[0].Volume != 85 || calls[1].Volume != 75 {
		t.Fatalf("volume calls = %#v, want 85 then 75", calls)
	}
}

func TestFailedFinalVolumeChangeRestoresConfirmedVolume(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetError(spotengine.OperationSetVolume, errors.New("engine unavailable"))
	m := readyModel(engine)
	m.engineVolume = 65
	m.confirmedEngineVolume = 65

	m, _ = key(m, "+")
	updated, setVolume := m.Update(msgs.VolumeDebounceElapsedMsg{Generation: m.volumeDebounceID})
	m = updated.(RootModel)
	updated, _ = m.Update(setVolume())
	m = updated.(RootModel)

	if m.engineVolume != 65 || m.confirmedEngineVolume != 65 {
		t.Fatalf("volume after failed command = %d (confirmed %d), want 65", m.engineVolume, m.confirmedEngineVolume)
	}
	if !m.statusIsErr || !strings.Contains(m.statusMsg, "engine unavailable") {
		t.Fatalf("missing volume error: %q", m.statusMsg)
	}
}

func TestAutoplayAndTransferStayEngineDriven(t *testing.T) {
	engine := spotengine.NewFake()
	m := readyModel(engine)
	m.engineTrack = &spotengine.Track{Name: "Hello", DurationMS: 100000}
	m.engineActive = true

	m, autoplayCmd := key(m, "a")
	updated, _ := m.Update(autoplayCmd())
	m = updated.(RootModel)
	if engine.AutoplayEnabled() || !strings.Contains(m.View(), "Autoplay Off") {
		t.Fatalf("autoplay state:\n%s", m.View())
	}

	updated, _ = m.Update(msgs.EngineEventMsg{Event: spotengine.Event{Type: spotengine.EventTypeInactive}})
	m = updated.(RootModel)
	if !strings.Contains(m.View(), "Transferred Playback") {
		t.Fatalf("transfer state:\n%s", m.View())
	}
	m, cmd := key(m, "n")
	if cmd != nil {
		t.Fatal("transferred playback accepted remote control")
	}
}

func TestLogoutConfirmationCancelAndSuccess(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetHasSession(true)
	m := readyModel(engine)
	m.engineTrack = &spotengine.Track{Name: "Hello", DurationMS: 100000}
	m.enginePlaying = true

	m, _ = key(m, "L")
	m, _ = key(m, "n")
	if !m.enginePlaying || !strings.Contains(m.View(), "Hello") || len(engine.Calls()) != 0 {
		t.Fatal("cancel changed playback")
	}

	m, _ = key(m, "L")
	m, logoutCmd := key(m, "y")
	updated, _ := m.Update(logoutCmd())
	m = updated.(RootModel)
	if m.state != stateLoggedOut || engine.HasSession() || m.engineTrack != nil {
		t.Fatalf("logout state=%v session=%v track=%#v", m.state, engine.HasSession(), m.engineTrack)
	}
}

func TestTransientFailureReconnectsWithoutLoginOrResume(t *testing.T) {
	engine := spotengine.NewFake()
	engine.SetHasSession(true)
	browser := &modelBrowser{}
	clock := &recoveryClock{}
	m := NewRootModelWithRecovery(engine, browser.Open, clock.Wait, func() float64 { return 0.5 })
	m.state = stateReady
	m.width, m.height = 80, 24
	m.engineTrack = &spotengine.Track{Name: "Hello"}
	m.enginePlaying = true

	updated, reconnectCmd := m.Update(msgs.EngineEventMsg{
		Event: spotengine.Event{Type: spotengine.EventTypeError, Err: errors.New("network unavailable")},
	})
	m = updated.(RootModel)
	updated, waitCmd := m.Update(reconnectCmd())
	m = updated.(RootModel)
	updated, startCmd := m.Update(waitCmd())
	m = updated.(RootModel)
	_ = startCmd()
	if m.state != stateReconnecting || len(clock.delays) != 1 || clock.delays[0] != time.Second ||
		len(browser.urls) != 0 {
		t.Fatalf("recovery state=%v delays=%v browser=%v", m.state, clock.delays, browser.urls)
	}
	updated, _ = m.Update(msgs.EngineEventMsg{Event: spotengine.Event{Type: spotengine.EventTypeReady}})
	m = updated.(RootModel)
	if m.state != stateReady || m.engineTrack != nil || m.enginePlaying {
		t.Fatal("recovery resumed audio")
	}
}

func TestReconnectBackoffAndCredentialRejection(t *testing.T) {
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 30 * time.Second}
	for attempt, expected := range want {
		if got := reconnectDelay(attempt, 0.5); got != expected {
			t.Fatalf("attempt %d: want %s, got %s", attempt, expected, got)
		}
	}
	if reconnectDelay(0, 0) != 800*time.Millisecond || reconnectDelay(20, 1) != 30*time.Second {
		t.Fatal("jitter or cap incorrect")
	}

	engine := spotengine.NewFake()
	engine.SetHasSession(true)
	m := NewRootModel(engine, func(string) error { return nil })
	m.state = stateLoading
	m.width, m.height = 80, 24
	updated, expireCmd := m.Update(msgs.EngineEventMsg{Event: spotengine.Event{
		Type: spotengine.EventTypeError, Err: errors.New("bad credentials"),
		ErrorKind: spotengine.ErrorKindCredentialRejected,
	}})
	m = updated.(RootModel)
	updated, _ = m.Update(expireCmd())
	m = updated.(RootModel)
	if m.state != stateLoggedOut || engine.HasSession() || !strings.Contains(m.View(), "Session expired") {
		t.Fatalf("expired session state=%v session=%v\n%s", m.state, engine.HasSession(), m.View())
	}
}
