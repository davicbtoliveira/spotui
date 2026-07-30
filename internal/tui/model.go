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

	nowPlaying      *spotify.PlayerState
	shuffleOn       bool
	localProgressMs int

	statusMsg   string
	statusIsErr bool
	showHelp    bool

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
		state:   stateLoggedOut,
		engine:  engine,
		openURL: openURL,
		library: library.New(),
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

	if m.state == stateUnsupported {
		switch msg.String() {
		case KeyQuitAlt:
			return m, tea.Quit
		case KeyLogout:
			return m, commands.CmdLogout(m.engine)
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
			return m, commands.CmdSearchTracks(m.trackSearcher, m.searchQuery, 0)
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

	if m.state != stateReady || m.client == nil {
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
		playing := m.nowPlaying != nil && m.nowPlaying.Playing
		return m, commands.CmdPlayPause(m.client, playing)

	case KeyNext:
		return m, commands.CmdNext(m.client)

	case KeyPrev:
		return m, commands.CmdPrevious(m.client)

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
				return m, commands.CmdSearchTracks(m.trackSearcher, m.searchQuery, nextOffset)
			}
		}
		return m, nil
	case KeySearchPrev:
		if m.library.ActiveTab() == library.TabSearch && m.searchCompleted && !m.searchLoading {
			prevOffset := m.searchOffset - commands.TrackSearchLimit
			if prevOffset >= 0 {
				m.searchLoading = true
				m.searchOffset = prevOffset
				return m, commands.CmdSearchTracks(m.trackSearcher, m.searchQuery, prevOffset)
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
		return m, commands.CmdPlayPlaylist(m.client, spotify.URI(uri))
	case library.TabTracks, library.TabSearch:
		return m, commands.CmdPlayTrack(m.client, spotify.URI(uri))
	}
	return m, nil
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
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			theme.SubtextStyle.Render("Loading your library..."))
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
	player := views.RenderPlayer(m.width, m.nowPlaying, m.shuffleOn, m.localProgressMs)

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
