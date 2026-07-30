package tui

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/auth"
	"github.com/dcbto/spotui/internal/library"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/spotifyapi"
	"github.com/dcbto/spotui/internal/theme"
	"github.com/dcbto/spotui/internal/tui/commands"
	"github.com/dcbto/spotui/internal/tui/views"
	"github.com/zmb3/spotify/v2"
)

type appState int

const (
	stateLoggedOut appState = iota
	stateAuthenticating
	stateCancelling
	stateLoading
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

	client *spotify.Client

	trackSearcher spotifyapi.TrackSearcher

	username string

	library *library.Library

	nowPlaying        *spotify.PlayerState
	shuffleOn         bool
	localProgressMs   int
	engineTrack       *spotengine.Track
	enginePlaying     bool
	engineBuffering   bool
	engineActive      bool
	engineVolume      int
	engineAutoplay    bool
	engineTransferred bool

	statusMsg        string
	statusIsErr      bool
	showHelp         bool
	confirmingLogout bool
	loggingOut       bool

	searchInputActive bool
	searchQuery       string
	searchLoading     bool
	searchCompleted   bool
	searchOffset      int
	searchTotal       int

	loadedFlags int
}

const (
	loadedUser      = 1 << 0
	loadedPlaylists = 1 << 1
	loadedTracks    = 1 << 2
	loadedArtists   = 1 << 3
	loadedAll       = loadedUser | loadedPlaylists | loadedTracks | loadedArtists
)

func NewRootModel(engine spotengine.Engine, openURL func(string) error) RootModel {
	model := RootModel{
		state:          stateLoggedOut,
		engine:         engine,
		openURL:        openURL,
		library:        library.New(),
		engineVolume:   100,
		engineAutoplay: engine.AutoplayEnabled(),
	}
	if engine.HasSession() {
		model.state = stateLoading
	}
	return model
}

func (m RootModel) Init() tea.Cmd {
	if m.state == stateLoading {
		return commands.CmdStartEngine(m.engine)
	}
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case AuthDoneMsg:
		m.client = msg.Client
		m.trackSearcher = spotifyapi.SpotifyTrackSearcher{Client: msg.Client}
		m.state = stateLoading
		return m, tea.Batch(
			commands.CmdFetchUser(m.client),
			commands.CmdFetchPlaylists(m.client),
			commands.CmdFetchTracks(m.client),
			commands.CmdFetchTopArtists(m.client),
			commands.CmdNowPlaying(m.client),
			commands.CmdTick(),
			commands.CmdProgressTick(),
		)

	case AuthErrMsg:
		m.statusMsg = "Auth failed: " + msg.Err.Error()
		m.statusIsErr = true
		m.state = stateLoggedOut
		return m, nil

	case EngineStartedMsg:
		return m, m.waitEngineEvent()

	case EngineStartErrMsg:
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
			if m.state == stateAuthenticating || m.state == stateLoading {
				clearSession := m.state == stateLoading && m.engine.HasSession()
				m.state = stateCancelling
				m.statusMsg = "Login failed: " + msg.Event.Err.Error()
				m.statusIsErr = true
				return m, commands.CmdResetLogin(m.engine, clearSession)
			}
			m.statusMsg = "Playback engine: " + msg.Event.Err.Error()
			m.statusIsErr = true
		case spotengine.EventTypeAccountProduct:
			if strings.EqualFold(msg.Event.Product, "free") {
				m.state = stateUnsupported
				return m, nil
			}
		case spotengine.EventTypeReady:
			if m.state == stateAuthenticating || m.state == stateLoading {
				m.state = stateReady
				m.library.SetActiveTab(library.TabSearch)
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
			if msg.Event.VolumeMax > 0 {
				m.engineVolume = msg.Event.Volume * 100 / msg.Event.VolumeMax
			} else {
				m.engineVolume = msg.Event.Volume
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

	case LoginResetMsg:
		m.state = stateLoggedOut
		m.authURL = ""
		return m, nil

	case LoginResetErrMsg:
		m.state = stateLoggedOut
		m.authURL = ""
		m.statusMsg = "Could not reset login: " + msg.Err.Error()
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

	case UserLoadedMsg:
		if msg.User.DisplayName != "" {
			m.username = msg.User.DisplayName
		} else {
			m.username = msg.User.ID
		}
		m.loadedFlags |= loadedUser
		m.checkReady()
		return m, nil

	case PlaylistsLoadedMsg:
		m.library.SetPlaylists(msg.Playlists)
		m.loadedFlags |= loadedPlaylists
		m.checkReady()
		return m, nil

	case TracksLoadedMsg:
		translated := make([]library.TrackEntry, len(msg.Tracks))
		for i, t := range msg.Tracks {
			translated[i] = spotifyapi.TranslateTrack(t)
		}
		m.library.SetTracks(translated)
		m.loadedFlags |= loadedTracks
		m.checkReady()
		return m, nil

	case TrackSearchLoadedMsg:
		translated := make([]library.TrackEntry, len(msg.Tracks))
		for i, t := range msg.Tracks {
			translated[i] = spotifyapi.TranslateFullTrack(t)
		}
		m.library.SetSearchResults(translated)
		m.searchLoading = false
		m.searchCompleted = true
		m.searchOffset = msg.Offset
		m.searchTotal = msg.Total
		return m, nil

	case EngineTrackSearchLoadedMsg:
		translated := make([]library.TrackEntry, len(msg.Tracks))
		for i, track := range msg.Tracks {
			translated[i] = library.TrackEntry{
				Name:     track.Name,
				Artist:   track.Artist,
				Duration: track.DurationMS,
				URI:      track.URI,
			}
		}
		m.library.SetSearchResults(translated)
		m.searchLoading = false
		m.searchCompleted = true
		m.searchOffset = msg.Offset
		m.searchTotal = msg.Total
		return m, nil

	case ArtistsLoadedMsg:
		translated := make([]library.ArtistEntry, len(msg.Artists))
		for i, a := range msg.Artists {
			translated[i] = spotifyapi.TranslateArtist(a)
		}
		m.library.SetArtists(translated)
		m.loadedFlags |= loadedArtists
		m.checkReady()
		return m, nil

	case NowPlayingMsg:
		m.nowPlaying = msg.State
		if m.nowPlaying != nil {
			m.shuffleOn = msg.State.ShuffleState
			m.localProgressMs = int(msg.State.Progress)
		}
		return m, nil

	case ProgressTickMsg:
		if m.engineTrack != nil && m.enginePlaying {
			m.localProgressMs += 1000
			if m.localProgressMs > m.engineTrack.DurationMS {
				m.localProgressMs = m.engineTrack.DurationMS
			}
		}
		if m.nowPlaying != nil && m.nowPlaying.Playing && m.nowPlaying.Item != nil {
			dur := int(m.nowPlaying.Item.Duration)
			m.localProgressMs += 1000
			if m.localProgressMs > dur {
				m.localProgressMs = dur
			}
		}
		return m, commands.CmdProgressTick()

	case TickMsg:
		if m.client == nil {
			return m, nil
		}
		return m, tea.Batch(
			commands.CmdNowPlaying(m.client),
			commands.CmdTick(),
		)

	case ErrMsg:
		m.statusMsg = msg.Context + ": " + msg.Err.Error()
		m.statusIsErr = true
		if msg.Context == "search tracks" {
			m.searchLoading = false
		}
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

func (m *RootModel) checkReady() {
	if m.loadedFlags == loadedAll {
		m.state = stateReady
	}
}

func (m *RootModel) clearAccount() {
	m.client = nil
	m.trackSearcher = nil
	m.username = ""
	m.library = library.New()
	m.nowPlaying = nil
	m.shuffleOn = false
	m.localProgressMs = 0
	m.engineTrack = nil
	m.enginePlaying = false
	m.engineBuffering = false
	m.engineActive = false
	m.engineTransferred = false
	m.engineVolume = 100
	m.engineAutoplay = m.engine.AutoplayEnabled()
	m.searchInputActive = false
	m.searchQuery = ""
	m.searchLoading = false
	m.searchCompleted = false
	m.searchOffset = 0
	m.searchTotal = 0
	m.loadedFlags = 0
}

func (m *RootModel) waitEngineEvent() tea.Cmd {
	if m.waitingEngine {
		return nil
	}
	m.waitingEngine = true
	return commands.CmdWaitEngineEvent(m.engine)
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
		default:
			return m, nil
		}
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

	if m.state == stateReady && m.searchInputActive {
		switch msg.Type {
		case tea.KeyRunes:
			m.searchQuery += string(msg.Runes)
			return m, nil
		case tea.KeySpace:
			m.searchQuery += " "
			return m, nil
		case tea.KeyBackspace, tea.KeyCtrlH:
			runes := []rune(m.searchQuery)
			if len(runes) > 0 {
				m.searchQuery = string(runes[:len(runes)-1])
			}
			return m, nil
		case tea.KeyEsc:
			m.searchInputActive = false
			return m, nil
		case tea.KeyEnter:
			if strings.TrimSpace(m.searchQuery) == "" {
				return m, nil
			}
			m.searchInputActive = false
			m.searchLoading = true
			m.searchCompleted = false
			m.searchOffset = 0
			return m, m.searchTracks(0)
		}
	}

	if m.state == stateLoggedOut && msg.Type == tea.KeyEnter {
		m.state = stateAuthenticating
		m.statusMsg = ""
		m.statusIsErr = false
		m.authURL = ""
		return m, commands.CmdStartEngine(m.engine)
	}

	switch msg.String() {
	case KeyQuitAlt:
		return m, tea.Quit
	case KeyHelp:
		m.showHelp = true
		return m, nil
	}

	if m.state != stateReady {
		return m, nil
	}

	switch msg.String() {
	case KeyTab1:
		m.library.SetActiveTab(library.TabPlaylists)
		return m, nil
	case KeyTab2:
		m.library.SetActiveTab(library.TabTracks)
		return m, nil
	case KeyTab3:
		m.library.SetActiveTab(library.TabArtists)
		return m, nil
	case KeyTab4:
		m.library.SetActiveTab(library.TabSearch)
		return m, nil
	case KeySearch:
		m.library.SetActiveTab(library.TabSearch)
		m.searchInputActive = true
		return m, nil

	case KeySpace:
		if m.engineTransferred {
			return m, nil
		}
		if m.enginePlaying {
			return m, commands.CmdPauseEngine(m.engine)
		}
		return m, commands.CmdResumeEngine(m.engine)

	case KeyNext:
		if m.engineTransferred {
			return m, nil
		}
		return m, commands.CmdNextEngine(m.engine)

	case KeyPrev:
		if m.engineTransferred {
			return m, nil
		}
		return m, commands.CmdPreviousEngine(m.engine)

	case KeyVolumeDown:
		if m.engineTransferred {
			return m, nil
		}
		volume := m.engineVolume - 5
		if volume < 0 {
			volume = 0
		}
		return m, commands.CmdSetEngineVolume(m.engine, volume)

	case KeyVolumeUp:
		if m.engineTransferred {
			return m, nil
		}
		volume := m.engineVolume + 5
		if volume > 100 {
			volume = 100
		}
		return m, commands.CmdSetEngineVolume(m.engine, volume)

	case KeyAutoplay:
		if m.engineTrack == nil || m.engineTransferred {
			return m, nil
		}
		return m, commands.CmdSetEngineAutoplay(m.engine, !m.engineAutoplay)

	case KeyLogout:
		m.confirmingLogout = true
		return m, nil

	case KeyShuffle:
		return m, commands.CmdShuffle(m.client, !m.shuffleOn)

	case KeySettings:
		auth.OpenSpotifySettings()
		return m, nil

	case KeySearchNext:
		if m.library.ActiveTab() == library.TabSearch && m.searchCompleted && !m.searchLoading {
			nextOffset := m.searchOffset + commands.TrackSearchLimit
			if nextOffset < m.searchTotal {
				m.searchLoading = true
				m.searchOffset = nextOffset
				return m, m.searchTracks(nextOffset)
			}
		}
		return m, nil
	case KeySearchPrev:
		if m.library.ActiveTab() == library.TabSearch && m.searchCompleted && !m.searchLoading {
			prevOffset := m.searchOffset - commands.TrackSearchLimit
			if prevOffset >= 0 {
				m.searchLoading = true
				m.searchOffset = prevOffset
				return m, m.searchTracks(prevOffset)
			}
		}
		return m, nil
	case KeyUp, KeyUpAlt:
		m.library.MoveUp()
		return m, nil
	case KeyDown, KeyDownAlt:
		m.library.MoveDown()
		return m, nil
	case KeyEnter:
		return m.handleEnter()
	}

	return m, nil
}

func (m RootModel) handleEnter() (tea.Model, tea.Cmd) {
	uri := m.library.SelectedURI()
	if uri == "" {
		return m, nil
	}
	switch m.library.ActiveTab() {
	case library.TabPlaylists:
		if m.client == nil {
			return m, nil
		}
		return m, commands.CmdPlayPlaylist(m.client, spotify.URI(uri))
	case library.TabTracks:
		if m.client == nil {
			return m, nil
		}
		return m, commands.CmdPlayTrack(m.client, spotify.URI(uri))
	case library.TabSearch:
		return m, commands.CmdPlayEngineTrack(m.engine, uri)
	}
	return m, nil
}

func (m RootModel) searchTracks(offset int) tea.Cmd {
	if m.client != nil && m.trackSearcher != nil {
		return commands.CmdSearchTracks(m.trackSearcher, m.searchQuery, offset)
	}
	return commands.CmdSearchEngineTracks(m.engine, m.searchQuery, offset)
}

func (m RootModel) View() string {
	if m.width == 0 {
		return ""
	}

	if m.width < 40 {
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

	topBar := views.RenderTopBar(m.width, m.username, m.library.ActiveTab())
	var player string
	if m.client == nil || m.engineTrack != nil {
		player = views.RenderEnginePlayer(m.width, views.EnginePlayerState{
			Track:       m.engineTrack,
			ProgressMS:  m.localProgressMs,
			Playing:     m.enginePlaying,
			Buffering:   m.engineBuffering,
			Active:      m.engineActive,
			Volume:      m.engineVolume,
			Autoplay:    m.engineAutoplay,
			Transferred: m.engineTransferred,
		})
	} else {
		player = views.RenderPlayer(m.width, m.nowPlaying, m.shuffleOn, m.localProgressMs)
	}

	topBarH := lipgloss.Height(topBar)
	playerH := lipgloss.Height(player)

	libraryH := m.height - topBarH - playerH
	if m.statusMsg != "" {
		libraryH--
	}
	if libraryH < 1 {
		libraryH = 1
	}

	libraryContent := m.library.View(m.width, libraryH)
	if m.library.ActiveTab() == library.TabSearch && m.searchInputActive {
		libraryContent = "  " + theme.ActiveTabStyle.Render("Search: "+m.searchQuery)
	} else if m.library.ActiveTab() == library.TabSearch && m.searchLoading {
		loadingText := "Searching tracks..."
		if m.searchCompleted {
			loadingText = "Loading more tracks..."
		}
		loading := "  " + theme.SubtextStyle.Render(loadingText)
		if m.library.SearchResultCount() > 0 {
			libraryContent = lipgloss.JoinVertical(lipgloss.Left, loading, libraryContent)
		} else {
			libraryContent = loading
		}
	} else if m.library.ActiveTab() == library.TabSearch && m.searchCompleted && m.library.SearchResultCount() == 0 {
		libraryContent = "  " + theme.SubtextStyle.Render("No tracks found")
	}
	lib := lipgloss.NewStyle().Height(libraryH).Width(m.width).Render(libraryContent)

	rows := []string{topBar, lib}

	if m.statusMsg != "" {
		var statusLine string
		if m.statusIsErr {
			statusLine = theme.ErrorStyle.Render("  ✗ " + m.statusMsg)
		} else {
			statusLine = theme.StatusStyle.Render("  " + m.statusMsg)
		}
		rows = append(rows, statusLine)
	}

	rows = append(rows, player)

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
