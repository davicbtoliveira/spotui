package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/auth"
	"github.com/dcbto/spotui/internal/library"
	"github.com/dcbto/spotui/internal/spotifyapi"
	"github.com/dcbto/spotui/internal/theme"
	"github.com/dcbto/spotui/internal/tui/commands"
	"github.com/dcbto/spotui/internal/tui/views"
	"github.com/zmb3/spotify/v2"
)

type appState int

const (
	stateAuth    appState = iota
	stateLoading
	stateReady
)

type RootModel struct {
	state    appState
	width    int
	height   int
	clientID string

	client *spotify.Client

	username string

	library *library.Library

	nowPlaying      *spotify.PlayerState
	shuffleOn       bool
	localProgressMs int

	statusMsg   string
	statusIsErr bool
	showHelp    bool

	loadedFlags int
}

const (
	loadedUser      = 1 << 0
	loadedPlaylists = 1 << 1
	loadedTracks    = 1 << 2
	loadedArtists   = 1 << 3
	loadedAll       = loadedUser | loadedPlaylists | loadedTracks | loadedArtists
)

func NewRootModel(clientID string) RootModel {
	return RootModel{
		state:    stateAuth,
		clientID: clientID,
		library:  library.New(),
	}
}

func (m RootModel) Init() tea.Cmd {
	return commands.CmdAuthenticate(m.clientID)
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case AuthDoneMsg:
		m.client = msg.Client
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
		return m, tea.Quit

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

func (m RootModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch msg.String() {
	case KeyQuit, KeyQuitAlt:
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
	case library.TabTracks:
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
	case stateAuth:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			theme.AppTitleStyle.Render("Opening Spotify login in your browser..."))
	case stateLoading:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			theme.SubtextStyle.Render("Loading your library..."))
	case stateReady:
		return m.renderMain()
	}
	return ""
}

func (m RootModel) renderMain() string {
	if m.showHelp {
		return views.RenderHelpOverlay(m.width, m.height)
	}

	header := views.RenderHeader(m.width, m.username, false)
	tabBar := m.library.TabBar(m.width)
	player := views.RenderPlayer(m.width, m.nowPlaying, m.shuffleOn, m.localProgressMs)

	headerH := lipgloss.Height(header)
	tabBarH := lipgloss.Height(tabBar)
	playerH := lipgloss.Height(player)

	libraryH := m.height - headerH - tabBarH - playerH
	if m.statusMsg != "" {
		libraryH--
	}
	if libraryH < 1 {
		libraryH = 1
	}

	libraryContent := m.library.View(m.width, libraryH)
	lib := lipgloss.NewStyle().Height(libraryH).Width(m.width).Render(libraryContent)

	rows := []string{header, tabBar, lib}

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
