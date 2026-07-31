package tui

import (
	"errors"
	"image"
	"math/rand/v2"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/catalog"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/theme"
	"github.com/dcbto/spotui/internal/tui/commands"
	"github.com/dcbto/spotui/internal/tui/views"
)

const spotifyAccessErrorMessage = "Could not access Spotify. Check your connection and try again."

type appState int

const (
	stateLoggedOut appState = iota
	stateAuthenticating
	stateCancelling
	stateLoading
	stateReconnecting
	stateUnsupported
	stateReady
)

type RootModel struct {
	state   appState
	width   int
	height  int
	engine  spotengine.Engine
	openURL func(string) error
	authURL string

	waitingEngine bool

	playerState
	browseState

	statusMsg        string
	statusIsErr      bool
	showHelp         bool
	confirmingLogout bool
	loggingOut       bool

	waitForRecovery func(time.Duration) tea.Cmd
	recoveryJitter  func() float64
	recoveryAttempt int
}

func NewRootModel(engine spotengine.Engine, openURL func(string) error) RootModel {
	return NewRootModelWithRecovery(
		engine,
		openURL,
		func(delay time.Duration) tea.Cmd {
			return tea.Tick(delay, func(time.Time) tea.Msg { return ReconnectTimerMsg{} })
		},
		rand.Float64,
	)
}

func NewRootModelWithRecovery(
	engine spotengine.Engine,
	openURL func(string) error,
	waitForRecovery func(time.Duration) tea.Cmd,
	recoveryJitter func() float64,
) RootModel {
	model := RootModel{
		state:   stateLoggedOut,
		engine:  engine,
		openURL: openURL,
		playerState: playerState{
			engineVolume:          100,
			confirmedEngineVolume: 100,
			engineAutoplay:        engine.AutoplayEnabled(),
		},
		browseState: browseState{
			navCursor:      0,
			browseFocus:    1,
			browseCache:    make(map[catalog.CacheKey]CatalogLoadedMsg),
			artwork:        make(map[string]image.Image),
			artworkLoading: make(map[string]bool),
		},
		waitForRecovery: waitForRecovery,
		recoveryJitter:  recoveryJitter,
	}
	if engine.HasSession() {
		model.state = stateLoading
	}
	return model
}

func (m RootModel) Init() tea.Cmd {
	if m.state == stateLoading {
		return tea.Batch(commands.CmdStartEngine(m.engine), commands.CmdProgressTick())
	}
	return commands.CmdProgressTick()
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case EngineStartedMsg:
		return m, m.waitEngineEvent()

	case EngineStartErrMsg:
		if m.state == stateReconnecting {
			m.statusMsg = "Reconnect failed: " + msg.Err.Error()
			m.statusIsErr = true
			return m, m.scheduleReconnect()
		}
		m.state = stateLoggedOut
		m.statusMsg = "Login failed: " + msg.Err.Error()
		m.statusIsErr = true
		return m, nil

	case EngineEventMsg:
		m.waitingEngine = false
		switch msg.Event.Type {
		case spotengine.EventTypeAuthorizationURL:
			if m.state != stateAuthenticating {
				return m, nil
			}
			m.authURL = msg.Event.URL
			m.statusMsg = ""
			m.statusIsErr = false
			return m, commands.CmdOpenURL(m.openURL, msg.Event.URL)
		case spotengine.EventTypeError:
			if msg.Event.ErrorKind == spotengine.ErrorKindCredentialRejected &&
				(m.state == stateLoading || m.state == stateReady || m.state == stateReconnecting) &&
				m.engine.HasSession() {
				m.state = stateLoading
				m.statusMsg = "Session expired. Removing invalid Local Session..."
				m.statusIsErr = true
				return m, commands.CmdExpireSession(m.engine)
			}
			if m.state == stateAuthenticating || m.state == stateLoading {
				clearSession := m.state == stateLoading && m.engine.HasSession()
				m.state = stateCancelling
				if isSpotifyAccessError(msg.Event.Err) {
					m.statusMsg = spotifyAccessErrorMessage
				} else {
					m.statusMsg = "Login failed: " + msg.Event.Err.Error()
				}
				m.statusIsErr = true
				return m, commands.CmdResetLogin(m.engine, clearSession)
			}
			if (m.state == stateReady || m.state == stateReconnecting) && m.engine.HasSession() {
				m.enterReconnecting(msg.Event.Err)
				return m, commands.CmdReconnectEngine(m.engine)
			}
			m.statusMsg = "Playback engine: " + msg.Event.Err.Error()
			m.statusIsErr = true
		case spotengine.EventTypeAccountProduct:
			if strings.EqualFold(msg.Event.Product, "free") {
				m.state = stateUnsupported
				return m, nil
			}
		case spotengine.EventTypeReady:
			if m.state == stateAuthenticating || m.state == stateLoading || m.state == stateReconnecting {
				m.state = stateReady
				m.recoveryAttempt = 0
				m.statusMsg = ""
				m.statusIsErr = false
				if !m.browseInitialized {
					m.browseInitialized = true
					m.browseRoute = catalog.Route{Kind: catalog.RouteLiked}
					m.browseTitle = "Liked Tracks"
					return m, tea.Batch(m.waitEngineEvent(), m.loadBrowseRoute())
				}
				return m, m.waitEngineEvent()
			}
		case spotengine.EventTypeMetadata:
			if msg.Event.Track != nil {
				track := *msg.Event.Track
				if track.DurationMS == 0 {
					track.DurationMS = msg.Event.DurationMS
				}
				m.engineTrack = &track
			}
			m.localProgressMs = msg.Event.PositionMS
		case spotengine.EventTypeSeek:
			m.localProgressMs = msg.Event.PositionMS
		case spotengine.EventTypeShuffle:
			m.engineShuffle = msg.Event.Shuffle
		case spotengine.EventTypeBuffering:
			m.engineBuffering = true
			m.enginePlaying = false
		case spotengine.EventTypePlaying:
			m.engineBuffering = false
			m.enginePlaying = true
		case spotengine.EventTypePaused:
			m.engineBuffering = false
			m.enginePlaying = false
		case spotengine.EventTypeStopped:
			m.engineBuffering = false
			m.enginePlaying = false
		case spotengine.EventTypeActive:
			m.engineActive = true
			m.engineTransferred = false
		case spotengine.EventTypeInactive:
			m.engineActive = false
			m.enginePlaying = false
			m.engineBuffering = false
			m.engineTransferred = true
		case spotengine.EventTypeVolume:
			if !m.volumeCommandInFlight && !m.volumeDebouncePending {
				if msg.Event.VolumeMax > 0 {
					m.engineVolume = msg.Event.Volume * 100 / msg.Event.VolumeMax
				} else {
					m.engineVolume = msg.Event.Volume
				}
				m.confirmedEngineVolume = m.engineVolume
			}
		case spotengine.EventTypeSessionEnded:
			if m.state == stateLoggedOut || m.state == stateCancelling || m.state == stateUnsupported {
				return m, nil
			}
		}
		return m, m.waitEngineEvent()

	case BrowserOpenedMsg:
		if m.state != stateAuthenticating {
			return m, nil
		}
		m.statusMsg = ""
		m.statusIsErr = false
		return m, m.waitEngineEvent()

	case BrowserOpenErrMsg:
		if m.state != stateAuthenticating {
			return m, nil
		}
		m.statusMsg = "Could not open browser: " + msg.Err.Error()
		m.statusIsErr = true
		return m, m.waitEngineEvent()

	case EngineEventsClosedMsg:
		m.waitingEngine = false
		if m.state != stateLoggedOut {
			m.state = stateLoggedOut
			m.statusMsg = "Login failed: playback engine closed"
			m.statusIsErr = true
		}
		return m, nil

	case EngineReconnectedMsg:
		return m, m.scheduleReconnect()

	case EngineReconnectErrMsg:
		m.statusMsg = "Reconnect cleanup: " + msg.Err.Error()
		m.statusIsErr = true
		return m, m.scheduleReconnect()

	case ReconnectTimerMsg:
		return m, commands.CmdStartEngine(m.engine)

	case SessionExpiredMsg:
		m.clearAccount()
		m.state = stateLoggedOut
		m.statusMsg = "Session expired. Log in with Spotify again."
		m.statusIsErr = true
		return m, nil

	case SessionExpireErrMsg:
		m.state = stateLoading
		m.statusMsg = "Session expired, but cleanup failed: " + msg.Err.Error()
		m.statusIsErr = true
		return m, nil

	case LoginResetMsg:
		m.state = stateLoggedOut
		m.authURL = ""
		return m, nil

	case LoginResetErrMsg:
		m.state = stateLoggedOut
		m.authURL = ""
		if isSpotifyAccessError(msg.Err) {
			m.statusMsg = spotifyAccessErrorMessage
		} else {
			m.statusMsg = "Could not reset login: " + msg.Err.Error()
		}
		m.statusIsErr = true
		return m, nil

	case AutoplayChangedMsg:
		m.engineAutoplay = msg.Enabled
		if msg.Enabled {
			m.statusMsg = "Autoplay enabled"
		} else {
			m.statusMsg = "Autoplay disabled"
		}
		m.statusIsErr = false
		return m, commands.CmdClearStatus()

	case LogoutDoneMsg:
		m.clearAccount()
		m.state = stateLoggedOut
		m.confirmingLogout = false
		m.loggingOut = false
		m.authURL = ""
		m.statusMsg = ""
		m.statusIsErr = false
		return m, nil

	case LogoutErrMsg:
		m.confirmingLogout = false
		m.loggingOut = false
		m.statusMsg = "Logout failed: " + msg.Err.Error()
		m.statusIsErr = true
		m.enginePlaying = false
		m.engineBuffering = false
		m.engineActive = false
		m.state = stateLoading
		return m, commands.CmdStartEngine(m.engine)

	case CatalogLoadedMsg:
		return m, m.applyCatalogMessage(msg)

	case ArtworkLoadedMsg:
		delete(m.artworkLoading, msg.URL)
		if msg.Err == nil && msg.Image != nil {
			m.cacheArtwork(msg.URL, msg.Image)
		}
		return m, nil

	case ShuffleChangedMsg:
		m.engineShuffle = msg.Enabled
		if msg.Enabled {
			m.statusMsg = "Shuffle enabled"
		} else {
			m.statusMsg = "Shuffle disabled"
		}
		m.statusIsErr = false
		return m, commands.CmdClearStatus()

	case ProgressTickMsg:
		if m.engineTrack != nil && m.enginePlaying {
			m.localProgressMs += 1000
			if m.localProgressMs > m.engineTrack.DurationMS {
				m.localProgressMs = m.engineTrack.DurationMS
			}
		}
		return m, commands.CmdProgressTick()

	case VolumeDebounceElapsedMsg:
		if msg.Generation != m.volumeDebounceID {
			return m, nil
		}
		m.volumeDebouncePending = false
		if m.volumeCommandInFlight {
			return m, nil
		}
		m.volumeCommandInFlight = true
		return m, commands.CmdSetEngineVolume(m.engine, m.engineVolume)

	case VolumeSetMsg:
		m.volumeCommandInFlight = false
		if msg.Err != nil {
			if m.engineVolume != msg.Volume {
				if m.volumeDebouncePending {
					return m, nil
				}
				m.volumeCommandInFlight = true
				return m, commands.CmdSetEngineVolume(m.engine, m.engineVolume)
			}
			m.engineVolume = m.confirmedEngineVolume
			m.statusMsg = "set volume: " + msg.Err.Error()
			m.statusIsErr = true
			return m, commands.CmdClearStatus()
		}
		m.confirmedEngineVolume = msg.Volume
		if m.volumeDebouncePending {
			return m, nil
		}
		if m.engineVolume != msg.Volume {
			m.volumeCommandInFlight = true
			return m, commands.CmdSetEngineVolume(m.engine, m.engineVolume)
		}
		return m, nil

	case ErrMsg:
		m.statusMsg = msg.Context + ": " + msg.Err.Error()
		m.statusIsErr = true
		return m, commands.CmdClearStatus()

	case ClearStatusMsg:
		m.statusMsg = ""
		m.statusIsErr = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func isSpotifyAccessError(err error) bool {
	var requestErr *url.Error
	if !errors.As(err, &requestErr) {
		return false
	}
	requestURL, parseErr := url.Parse(requestErr.URL)
	if parseErr != nil {
		return false
	}
	host := strings.ToLower(requestURL.Hostname())
	return host == "spotify.com" || strings.HasSuffix(host, ".spotify.com")
}

func (m *RootModel) clearAccount() {
	m.localProgressMs = 0
	m.engineTrack = nil
	m.enginePlaying = false
	m.engineBuffering = false
	m.engineActive = false
	m.engineTransferred = false
	m.engineVolume = 100
	m.confirmedEngineVolume = 100
	m.volumeCommandInFlight = false
	m.volumeDebouncePending = false
	m.volumeDebounceID++
	m.engineAutoplay = m.engine.AutoplayEnabled()
	m.engineShuffle = false
	m.searchInputActive = false
	m.searchQuery = ""
	m.browseInitialized = false
	m.navCursor = 0
	m.browseFocus = 1
	m.browseRoute = catalog.Route{}
	m.browseTitle = ""
	m.browseItems = nil
	m.browseCursor = 0
	m.browseLoading = false
	m.browseError = ""
	m.browseOffset = 0
	m.browseTotal = 0
	m.browseContextURI = ""
	m.browseStack = nil
	m.browseCache = make(map[catalog.CacheKey]CatalogLoadedMsg)
	m.browseCacheOrder = nil
	m.artwork = make(map[string]image.Image)
	m.artworkLoading = make(map[string]bool)
	m.artworkOrder = nil
}

func (m *RootModel) waitEngineEvent() tea.Cmd {
	if m.waitingEngine {
		return nil
	}
	m.waitingEngine = true
	return commands.CmdWaitEngineEvent(m.engine)
}

func (m *RootModel) enterReconnecting(err error) {
	m.state = stateReconnecting
	m.statusMsg = "Connection lost: " + err.Error()
	m.statusIsErr = true
	m.engineTrack = nil
	m.enginePlaying = false
	m.engineBuffering = false
	m.engineActive = false
	m.engineTransferred = false
	m.localProgressMs = 0
}

func (m *RootModel) scheduleReconnect() tea.Cmd {
	delay := reconnectDelay(m.recoveryAttempt, m.recoveryJitter())
	m.recoveryAttempt++
	return m.waitForRecovery(delay)
}

func reconnectDelay(attempt int, jitter float64) time.Duration {
	base := time.Second
	for i := 0; i < attempt && base < 30*time.Second; i++ {
		base *= 2
	}
	if base > 30*time.Second {
		base = 30 * time.Second
	}
	delay := time.Duration(float64(base) * (0.8 + 0.4*jitter))
	if delay > 30*time.Second {
		return 30 * time.Second
	}
	return delay
}

func (m RootModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	if msg.String() == KeyQuit {
		return m, tea.Quit
	}

	if m.confirmingLogout {
		switch {
		case msg.String() == "y":
			m.confirmingLogout = false
			m.loggingOut = true
			return m, commands.CmdLogout(m.engine)
		case msg.String() == "n" || msg.Type == tea.KeyEsc || msg.Type == tea.KeyEnter:
			m.confirmingLogout = false
			return m, nil
		case msg.String() == KeyQuitAlt:
			return m, tea.Quit
		default:
			return m, nil
		}
	}

	if m.loggingOut {
		if msg.String() == KeyQuitAlt {
			return m, tea.Quit
		}
		return m, nil
	}

	if m.state == stateUnsupported {
		switch msg.String() {
		case KeyQuitAlt:
			return m, tea.Quit
		case KeyLogout:
			m.confirmingLogout = true
			return m, nil
		default:
			return m, nil
		}
	}

	if m.state == stateAuthenticating {
		switch {
		case msg.Type == tea.KeyEsc:
			m.state = stateCancelling
			return m, commands.CmdResetLogin(m.engine, false)
		case msg.String() == KeyRetry && m.authURL != "":
			return m, commands.CmdOpenURL(m.openURL, m.authURL)
		}
	}

	if m.state == stateCancelling {
		if msg.String() == KeyQuitAlt {
			return m, tea.Quit
		}
		return m, nil
	}

	if m.state == stateReconnecting {
		if msg.String() == KeyQuitAlt {
			return m, tea.Quit
		}
		return m, nil
	}

	if m.state == stateReady && msg.String() == KeyLogout {
		m.confirmingLogout = true
		return m, nil
	}
	if m.state == stateReady {
		switch msg.String() {
		case KeyQuitAlt:
			return m, tea.Quit
		case KeyHelp:
			m.showHelp = true
			return m, nil
		}
	}

	if m.state == stateReady && m.browseInitialized {
		return m.handleBrowseKey(msg)
	}

	if m.state == stateLoggedOut && msg.Type == tea.KeyEnter {
		m.state = stateAuthenticating
		m.statusMsg = ""
		m.statusIsErr = false
		m.authURL = ""
		return m, commands.CmdStartEngine(m.engine)
	}

	return m, nil
}

func (m RootModel) View() string {
	if m.width == 0 {
		return ""
	}

	if m.width < 50 || m.height < 16 {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			theme.ErrorStyle.Render("Terminal too small"))
	}

	if m.confirmingLogout {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				theme.TopBarTitle.Render("Log out of Spotify?"),
				theme.SubtextStyle.Render("Playback will stop and the Local Session will be removed."),
				theme.SubtextStyle.Render("Confirm? y/N"),
			))
	}
	if m.loggingOut {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			theme.SubtextStyle.Render("Logging out..."))
	}

	switch m.state {
	case stateLoggedOut:
		rows := []string{
			theme.TopBarTitle.Render("SpotUI"),
			theme.SubtextStyle.Render("Spotify Premium is required"),
			theme.SubtextStyle.Render("Press Enter to Log in with Spotify"),
			theme.SubtextStyle.Render("Press q to quit"),
		}
		if m.statusMsg != "" {
			rows = append(rows, theme.ErrorStyle.Render(m.statusMsg))
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, rows...))
	case stateAuthenticating:
		rows := []string{theme.SubtextStyle.Render("Waiting for Spotify authorization...")}
		if m.authURL != "" {
			rows = append(rows,
				theme.SubtextStyle.Render("If the browser did not open, use this URL:"),
				m.authURL,
				theme.SubtextStyle.Render("Press r to retry browser • esc to cancel"),
			)
		}
		if m.statusMsg != "" {
			rows = append(rows, theme.ErrorStyle.Render(m.statusMsg))
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, rows...))
	case stateCancelling:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			theme.SubtextStyle.Render("Closing Spotify login..."))
	case stateLoading:
		rows := []string{theme.SubtextStyle.Render("Loading your library...")}
		if m.statusMsg != "" {
			rows = append(rows, theme.ErrorStyle.Render(m.statusMsg))
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, rows...))
	case stateReconnecting:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				theme.TopBarTitle.Render("Reconnecting..."),
				theme.SubtextStyle.Render("Playback controls and Track Search are temporarily disabled."),
				theme.ErrorStyle.Render(m.statusMsg),
				theme.SubtextStyle.Render("Press q to quit"),
			))
	case stateUnsupported:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center,
				theme.TopBarTitle.Render("Spotify Free is not supported"),
				theme.SubtextStyle.Render("SpotUI requires Spotify Premium for playback"),
				theme.SubtextStyle.Render("Press L to Log out"),
				theme.SubtextStyle.Render("Press q to quit"),
			))
	case stateReady:
		return m.renderMain()
	}
	return ""
}

func (m RootModel) renderMain() string {
	if m.showHelp {
		return views.RenderHelpOverlay(m.width, m.height)
	}
	return m.renderBrowseShell()
}
