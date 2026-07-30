package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/artwork"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/theme"
	"github.com/dcbto/spotui/internal/tui/commands"
	"github.com/dcbto/spotui/internal/tui/views"
)

var navLabels = []string{"Library", "  Liked Tracks", "  Playlists", "  Saved Albums", "Recommended", "Search"}

func (m *RootModel) loadBrowseRoute() tea.Cmd {
	m.browseLoading = true
	m.browseError = ""
	m.browseRequestID++
	requestID := m.browseRequestID
	cacheKey := fmt.Sprintf("%s|%d|%s", m.browseRoute, m.browseOffset, m.searchQuery)
	if cached, ok := m.browseCache[cacheKey]; ok {
		cached.RequestID = requestID
		return func() tea.Msg { return cached }
	}
	switch {
	case m.browseRoute == "liked":
		return commands.CmdLikedTracks(m.engine, m.browsePageRequest(), requestID)
	case m.browseRoute == "playlists":
		return commands.CmdUserPlaylists(m.engine, m.browsePageRequest(), requestID)
	case m.browseRoute == "albums":
		return commands.CmdSavedAlbums(m.engine, m.browsePageRequest(), requestID)
	case m.browseRoute == "recommended":
		return commands.CmdRecommended(m.engine, spotengine.PageRequest{Offset: m.browseOffset, Limit: 10}, requestID)
	case m.browseRoute == "search":
		return commands.CmdSearch(m.engine, spotengine.SearchRequest{Query: m.searchQuery, Offset: m.browseOffset, Limit: 10}, requestID)
	case strings.HasPrefix(m.browseRoute, "playlist:"):
		return commands.CmdPlaylist(m.engine, strings.TrimPrefix(m.browseRoute, "playlist:"), m.browsePageRequest(), requestID)
	case strings.HasPrefix(m.browseRoute, "album:"):
		return commands.CmdAlbum(m.engine, strings.TrimPrefix(m.browseRoute, "album:"), m.browsePageRequest(), requestID)
	case strings.HasPrefix(m.browseRoute, "artist:"):
		return commands.CmdArtist(m.engine, strings.TrimPrefix(m.browseRoute, "artist:"), m.browsePageRequest(), requestID)
	default:
		m.browseLoading = false
		return nil
	}
}

func (m RootModel) browsePageRequest() spotengine.PageRequest {
	return spotengine.PageRequest{Offset: m.browseOffset, Limit: 10}
}

func (m *RootModel) applyCatalogMessage(msg CatalogLoadedMsg) tea.Cmd {
	if (msg.Route != m.browseRoute && !(msg.Route == "search" && m.browseRoute == "search")) || (msg.RequestID != 0 && msg.RequestID != m.browseRequestID) {
		return nil
	}
	m.browseLoading = false
	requestedOffset := m.browseOffset
	if msg.Err != nil {
		m.browseError = msg.Err.Error()
		return nil
	}
	if m.browseCache == nil {
		m.browseCache = make(map[string]CatalogLoadedMsg)
	}
	cacheKey := fmt.Sprintf("%s|%d|%s", msg.Route, m.browseOffset, m.searchQuery)
	cacheMsg := msg
	cacheMsg.RequestID = 0
	m.browseCache[cacheKey] = cacheMsg
	m.browseError = ""
	m.browseCursor = 0
	// Preserve the requested page offset; stale responses are rejected above.
	m.browseTotal = 0
	m.browseContextURI = ""
	m.browseContextKind = ""
	m.browseMeta = ""
	switch value := msg.Data.(type) {
	case spotengine.TrackPage:
		m.browseOffset, m.browseTotal = pageOffset(value.Offset, requestedOffset), value.Total
		m.browseItems = make([]browseItem, 0, len(value.Items))
		for _, item := range value.Items {
			m.browseItems = append(m.browseItems, trackItem(item))
		}
	case spotengine.CatalogPage[spotengine.PlaylistSummary]:
		m.browseOffset, m.browseTotal = pageOffset(value.Offset, requestedOffset), value.Total
		m.browseItems = make([]browseItem, 0, len(value.Items))
		for _, item := range value.Items {
			m.browseItems = append(m.browseItems, playlistItem(item))
		}
	case spotengine.CatalogPage[spotengine.AlbumSummary]:
		m.browseOffset, m.browseTotal = pageOffset(value.Offset, requestedOffset), value.Total
		m.browseItems = make([]browseItem, 0, len(value.Items))
		for _, item := range value.Items {
			m.browseItems = append(m.browseItems, albumItem(item))
		}
	case spotengine.CatalogPage[spotengine.ArtistSummary]:
		m.browseOffset, m.browseTotal = pageOffset(value.Offset, requestedOffset), value.Total
		m.browseItems = make([]browseItem, 0, len(value.Items))
		for _, item := range value.Items {
			m.browseItems = append(m.browseItems, artistItem(item))
		}
	case spotengine.SearchGroups:
		m.groupedSearch = value
		m.browseTotal = maxSearchTotal(value)
		m.browseItems = searchItems(value)
	case spotengine.RecommendedPage:
		m.browseItems = recommendedItems(value)
		m.browseTotal = max(value.Artists.Total, value.Tracks.Total)
	case spotengine.PlaylistDetail:
		m.browseOffset, m.browseTotal = pageOffset(value.Tracks.Offset, requestedOffset), value.Tracks.Total
		m.browseTitle = value.Name
		m.browseContextURI = value.URI
		m.browseContextKind = "playlist"
		m.browseMeta = fmt.Sprintf("%s · %d tracks", value.Owner, value.TrackCount)
		m.browseItems = make([]browseItem, 0, len(value.Tracks.Items))
		for _, item := range value.Tracks.Items {
			m.browseItems = append(m.browseItems, trackItem(item))
		}
	case spotengine.AlbumDetail:
		m.browseOffset, m.browseTotal = pageOffset(value.Tracks.Offset, requestedOffset), value.Tracks.Total
		m.browseTitle = value.Name
		m.browseContextURI = value.URI
		m.browseContextKind = "album"
		m.browseMeta = fmt.Sprintf("%s · %s · %d tracks", value.Artist, value.ReleaseDate, value.TrackCount)
		m.browseItems = make([]browseItem, 0, len(value.Tracks.Items))
		for _, item := range value.Tracks.Items {
			m.browseItems = append(m.browseItems, trackItem(item))
		}
	case spotengine.ArtistDetail:
		m.browseTitle = value.Name
		m.browseMeta = strings.Join(value.Genres, " · ")
		if value.ImageURL != "" {
			m.browseItems = []browseItem{{kind: "header", Title: "Artist artwork", ImageURL: value.ImageURL}}
		}
		m.browseItems = append(m.browseItems, make([]browseItem, 0, len(value.Popular.Items)+len(value.Albums))...)
		for _, item := range value.Popular.Items {
			m.browseItems = append(m.browseItems, trackItem(item))
		}
		for _, item := range value.Albums {
			m.browseItems = append(m.browseItems, albumItem(item))
		}
	}
	return m.requestArtwork()
}

func pageOffset(responseOffset, requestedOffset int) int {
	if responseOffset == 0 && requestedOffset > 0 {
		return requestedOffset
	}
	return responseOffset
}

func (m *RootModel) requestArtwork() tea.Cmd {
	if m.browseCursor < 0 || m.browseCursor >= len(m.browseItems) {
		return nil
	}
	url := m.browseItems[m.browseCursor].ImageURL
	if url == "" || m.artwork[url] != nil || m.artworkLoading[url] {
		return nil
	}
	if m.artworkLoading == nil {
		m.artworkLoading = make(map[string]bool)
	}
	m.artworkLoading[url] = true
	return commands.CmdLoadArtwork(url)
}

func searchItems(groups spotengine.SearchGroups) []browseItem {
	items := make([]browseItem, 0)
	items = append(items, browseItem{kind: "header", Title: fmt.Sprintf("Tracks (%d)", groups.Tracks.Total)})
	for _, item := range groups.Tracks.Items {
		i := trackItem(item)
		i.Subtitle = "Track · " + i.Subtitle
		items = append(items, i)
	}
	items = append(items, browseItem{kind: "header", Title: searchGroupTitle("Albums", groups.Albums.Total, groups.NonTrackUnavailable)})
	for _, item := range groups.Albums.Items {
		i := albumItem(item)
		i.Subtitle = "Album · " + i.Subtitle
		items = append(items, i)
	}

	items = append(items, browseItem{kind: "header", Title: searchGroupTitle("Artists", groups.Artists.Total, groups.NonTrackUnavailable)})
	for _, item := range groups.Artists.Items {
		i := artistItem(item)
		i.Subtitle = "Artist"
		items = append(items, i)
	}
	// Playlists are the final search group.
	items = append(items, browseItem{kind: "header", Title: searchGroupTitle("Playlists", groups.Playlists.Total, groups.NonTrackUnavailable)})
	for _, item := range groups.Playlists.Items {
		i := playlistItem(item)
		i.Subtitle = "Playlist · " + i.Subtitle
		items = append(items, i)
	}
	return items
}

func searchGroupTitle(name string, total int, unavailable bool) string {
	if unavailable {
		return name + " (unavailable)"
	}
	if total == 0 {
		return name + " (empty)"
	}
	return fmt.Sprintf("%s (%d)", name, total)
}

func maxSearchTotal(groups spotengine.SearchGroups) int {
	return max(max(groups.Tracks.Total, groups.Albums.Total), max(groups.Artists.Total, groups.Playlists.Total))
}

func recommendedItems(value spotengine.RecommendedPage) []browseItem {
	items := []browseItem{{kind: "header", Title: fmt.Sprintf("Your top artists (%d)", value.Artists.Total)}}
	for _, artist := range value.Artists.Items {
		items = append(items, artistItem(artist))
	}
	trackTitle := fmt.Sprintf("Your top tracks (%d)", value.Tracks.Total)
	if value.TracksUnavailable {
		trackTitle = "Your top tracks (unavailable)"
	}
	items = append(items, browseItem{kind: "header", Title: trackTitle})
	for _, track := range value.Tracks.Items {
		items = append(items, trackItem(track))
	}
	items = append(items, browseItem{kind: "header", Title: fmt.Sprintf("Albums from top artists (%d)", len(value.Albums))})
	for _, album := range value.Albums {
		items = append(items, albumItem(album))
	}
	items = append(items, browseItem{kind: "header", Title: fmt.Sprintf("Playlists from top artists (%d)", len(value.Playlists))})
	for _, playlist := range value.Playlists {
		items = append(items, playlistItem(playlist))
	}
	return items
}

func (m *RootModel) selectNav(index int) tea.Cmd {
	if index < 0 {
		index = 0
	}
	if index >= len(navLabels) {
		index = len(navLabels) - 1
	}
	m.navCursor = index
	m.browseFocus = 0
	if m.width < 80 || m.height < 24 {
		m.browseFocus = 1
	}
	switch index {
	case 0, 1:
		m.activeSection = sectionLibrary
		m.browseRoute = "liked"
		m.browseTitle = "Liked Tracks"
	case 2:
		m.activeSection = sectionLibrary
		m.browseRoute = "playlists"
		m.browseTitle = "Playlists"
	case 3:
		m.activeSection = sectionLibrary
		m.browseRoute = "albums"
		m.browseTitle = "Saved Albums"
	case 4:
		m.activeSection = sectionRecommended
		m.browseRoute = "recommended"
		m.browseTitle = "Recommended"
	case 5:
		m.activeSection = sectionSearch
		m.browseRoute = "search"
		m.browseTitle = "Search"
	}
	m.browseItems = nil
	m.browseCursor = 0
	return m.loadBrowseRoute()
}

func (m *RootModel) activateBrowseItem() tea.Cmd {
	if m.browseCursor < 0 || m.browseCursor >= len(m.browseItems) {
		return nil
	}
	item := m.browseItems[m.browseCursor]
	switch item.kind {
	case "track":
		if m.browseContextURI != "" {
			return commands.CmdPlayContext(m.engine, m.browseContextURI, item.URI, 0)
		}
		return commands.CmdPlayEngineTrack(m.engine, item.URI)
	case "playlist", "album", "artist":
		m.browseStack = append(m.browseStack, browseSnapshot{route: m.browseRoute, title: m.browseTitle, items: append([]browseItem(nil), m.browseItems...), cursor: m.browseCursor, contextURI: m.browseContextURI, contextKind: m.browseContextKind, offset: m.browseOffset, meta: m.browseMeta})
		m.browseRoute = item.kind + ":" + item.URI
		m.browseTitle = item.Title
		m.browseItems = nil
		m.browseCursor = 0
		return m.loadBrowseRoute()
	}
	return nil
}

func (m *RootModel) popBrowseRoute() bool {
	if len(m.browseStack) == 0 {
		return false
	}
	last := m.browseStack[len(m.browseStack)-1]
	m.browseStack = m.browseStack[:len(m.browseStack)-1]
	m.browseRoute, m.browseTitle, m.browseItems, m.browseCursor, m.browseContextURI, m.browseContextKind, m.browseOffset, m.browseMeta = last.route, last.title, last.items, last.cursor, last.contextURI, last.contextKind, last.offset, last.meta
	m.browseLoading = false
	m.browseError = ""
	return true
}

func (m RootModel) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.groupedSearchActive && m.searchInputActive {
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
			m.browseRoute = "search"
			m.browseTitle = "Search"
			return m, m.loadBrowseRoute()
		}
	}

	if msg.Type == tea.KeyTab || msg.String() == "tab" {
		m.browseFocus = 1 - m.browseFocus
		return m, nil
	}
	if msg.String() == "shift+tab" {
		m.browseFocus = 1 - m.browseFocus
		return m, nil
	}
	if msg.String() == KeySearch {
		m.navCursor = 5
		m.activeSection = sectionSearch
		m.browseFocus = 1
		m.groupedSearchActive = true
		m.searchInputActive = true
		m.browseRoute = "search"
		m.browseTitle = "Search"
		m.browseItems = nil
		m.browseError = ""
		return m, nil
	}

	if msg.String() == "r" && m.browseError != "" {
		return m, m.loadBrowseRoute()
	}
	if msg.String() == KeySearchNext && !m.browseLoading && m.browseOffset+10 < m.browseTotal {
		m.browseOffset += 10
		return m, m.loadBrowseRoute()
	}
	if msg.String() == KeySearchPrev && !m.browseLoading && m.browseOffset >= 10 {
		m.browseOffset -= 10
		return m, m.loadBrowseRoute()
	}
	if msg.String() == "o" && m.browseCursor >= 0 && m.browseCursor < len(m.browseItems) {
		if url := m.browseItems[m.browseCursor].ExternalURL; url != "" {
			return m, commands.CmdOpenURL(m.openURL, url)
		}
	}

	if msg.String() == KeySpace {
		if m.engineTransferred {
			return m, nil
		}
		if m.enginePlaying {
			return m, commands.CmdPauseEngine(m.engine)
		}
		return m, commands.CmdResumeEngine(m.engine)
	}
	if msg.String() == KeyNext && !m.engineTransferred {
		return m, commands.CmdNextEngine(m.engine)
	}
	if msg.String() == KeyPrev && !m.engineTransferred {
		return m, commands.CmdPreviousEngine(m.engine)
	}
	if msg.String() == KeyVolumeDown && !m.engineTransferred {
		value := m.engineVolume - 5
		if value < 0 {
			value = 0
		}
		return m, commands.CmdSetEngineVolume(m.engine, value)
	}
	if msg.String() == KeyVolumeUp && !m.engineTransferred {
		value := m.engineVolume + 5
		if value > 100 {
			value = 100
		}
		return m, commands.CmdSetEngineVolume(m.engine, value)
	}
	if msg.String() == KeyAutoplay && m.engineTrack != nil && !m.engineTransferred {
		return m, commands.CmdSetEngineAutoplay(m.engine, !m.engineAutoplay)
	}
	if msg.String() == "s" && m.browseContextURI != "" && !m.engineTransferred {
		return m, commands.CmdSetEngineShuffle(m.engine, !m.engineShuffle)
	}
	if msg.String() == "h" && !m.engineTransferred {
		if m.localProgressMs <= 0 {
			return m, nil
		}
		return m, commands.CmdSeekEngine(m.engine, -10000)
	}
	if msg.String() == "l" && !m.engineTransferred {
		if m.engineTrack != nil && m.engineTrack.DurationMS > 0 && m.localProgressMs >= m.engineTrack.DurationMS {
			return m, nil
		}
		return m, commands.CmdSeekEngine(m.engine, 10000)
	}

	if msg.Type == tea.KeyEsc || msg.Type == tea.KeyBackspace {
		if m.searchInputActive {
			m.searchInputActive = false
			return m, nil
		}
		if m.popBrowseRoute() {
			return m, nil
		}
		m.browseFocus = 0
		return m, nil
	}

	if m.browseFocus == 0 {
		switch msg.String() {
		case KeyUp, KeyUpAlt:
			m.navCursor--
			if m.navCursor < 0 {
				m.navCursor = len(navLabels) - 1
			}
			return m, nil
		case KeyDown, KeyDownAlt:
			m.navCursor++
			if m.navCursor >= len(navLabels) {
				m.navCursor = 0
			}
			return m, nil
		case KeyEnter:
			if m.navCursor == 5 {
				m.groupedSearchActive = true
				m.searchInputActive = true
				m.searchQuery = ""
				m.browseItems = nil
				m.browseFocus = 1
				return m, nil
			}
			m.groupedSearchActive = false
			return m, m.selectNav(m.navCursor)
		}
		return m, nil
	}

	switch msg.String() {
	case KeyUp, KeyUpAlt:
		m.browseCursor = previousBrowseItem(m.browseItems, m.browseCursor)
		return m, m.requestArtwork()
	case KeyDown, KeyDownAlt:
		m.browseCursor = nextBrowseItem(m.browseItems, m.browseCursor)
		return m, m.requestArtwork()
	case KeyEnter:
		return m, m.activateBrowseItem()
	}
	return m, nil
}

func previousBrowseItem(items []browseItem, cursor int) int {
	for i := cursor - 1; i >= 0; i-- {
		if items[i].kind != "header" {
			return i
		}
	}
	return cursor
}

func nextBrowseItem(items []browseItem, cursor int) int {
	for i := cursor + 1; i < len(items); i++ {
		if items[i].kind != "header" {
			return i
		}
	}
	return cursor
}

func (m RootModel) renderBrowseShell() string {
	top := views.RenderBrowseTopBar(m.width, navLabels[m.navCursor])
	wide := m.width >= 80 && m.height >= 24
	playerWidth := m.width
	if wide {
		playerWidth -= 24
	}
	player := views.RenderEnginePlayer(playerWidth, views.EnginePlayerState{Track: m.engineTrack, ProgressMS: m.localProgressMs, Playing: m.enginePlaying, Buffering: m.engineBuffering, Active: m.engineActive, Volume: m.engineVolume, Autoplay: m.engineAutoplay, Shuffle: m.engineShuffle, Transferred: m.engineTransferred})
	playerH := lipgloss.Height(player)
	contentH := m.height - lipgloss.Height(top) - playerH
	if m.statusMsg != "" {
		contentH--
	}
	if contentH < 1 {
		contentH = 1
	}
	contentW := m.width
	if wide {
		contentW -= 24
	}
	if contentW < 1 {
		contentW = 1
	}
	content := renderBrowseContent(m, contentW, contentH)
	var body string
	if wide {
		nav := renderNavigation(m.navCursor, m.browseFocus == 0, 24, contentH)
		body = lipgloss.JoinHorizontal(lipgloss.Top, nav, lipgloss.NewStyle().Width(contentW).Height(contentH).Render(content))
	} else if m.browseFocus == 0 {
		body = renderNavigation(m.navCursor, true, m.width, contentH)
	} else {
		body = lipgloss.NewStyle().Width(contentW).Height(contentH).Render(content)
	}
	rows := []string{top, body}
	if m.statusMsg != "" {
		if m.statusIsErr {
			rows = append(rows, theme.ErrorStyle.Render("  ✗ "+m.statusMsg))
		} else {
			rows = append(rows, theme.StatusStyle.Render("  "+m.statusMsg))
		}
	}
	if wide {
		rows = append(rows, lipgloss.NewStyle().MarginLeft(24).Render(player))
	} else {
		rows = append(rows, player)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderNavigation(cursor int, focused bool, width, height int) string {
	rows := []string{"", theme.TopBarTitle.Render("  Browse")}
	for i, label := range navLabels {
		style := theme.InactiveTabStyle
		if i == cursor {
			style = theme.SelectedItemStyle
			if focused {
				style = theme.ActiveTabStyle
			}
		}
		rows = append(rows, style.Render("  "+label))
	}
	rows = append(rows, "", theme.SubtextStyle.Render("  Tab focus"), theme.SubtextStyle.Render("  Enter open"), theme.SubtextStyle.Render("  Esc back"))
	return lipgloss.NewStyle().Width(width).Height(height).Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(theme.ColorDim).Render(strings.Join(rows, "\n"))
}

func renderBrowseContent(m RootModel, width, height int) string {
	title := m.browseTitle
	if m.browseMeta != "" {
		title += " · " + m.browseMeta
	}
	rows := []string{"  " + theme.TopBarTitle.Render(title)}
	if m.browseCursor >= 0 && m.browseCursor < len(m.browseItems) {
		if img := m.artwork[m.browseItems[m.browseCursor].ImageURL]; img != nil {
			rows = append(rows, artwork.RenderANSI(img, artwork.CellRect{Columns: 14, Rows: 6}))
		}
	}
	if m.browseRoute == "search" {
		rows = append(rows, "  "+theme.SubtextStyle.Render("Search: "+m.searchQuery))
	}
	if m.browseLoading && len(m.browseItems) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, append(rows, "", "  "+theme.SubtextStyle.Render("Loading..."))...)
	}
	if m.browseError != "" {
		rows = append(rows, "", "  "+theme.ErrorStyle.Render(m.browseError), "  "+theme.SubtextStyle.Render("Press r to retry"))
		return strings.Join(rows, "\n")
	}
	if len(m.browseItems) == 0 {
		rows = append(rows, "", "  "+theme.SubtextStyle.Render("No items found"))
		return strings.Join(rows, "\n")
	}
	for i, item := range m.browseItems {
		prefix := "  "
		style := theme.NormalItemStyle
		if i == m.browseCursor && m.browseFocus == 1 {
			prefix = "▸ "
			style = theme.SelectedItemStyle
		}
		if item.kind == "header" {
			rows = append(rows, "", "  "+theme.ActiveTabStyle.Render(item.Title))
			continue
		}
		detail := item.Subtitle
		if item.ImageURL != "" {
			detail = "▣ " + detail
		}
		if item.DurationMS > 0 {
			detail += " · " + formatBrowseDuration(item.DurationMS)
		}
		rows = append(rows, prefix+style.Render(truncateBrowse(item.Title, width-6))+"  "+theme.SubtextStyle.Render(truncateBrowse(detail, width/2)))
		if len(rows) >= height {
			break
		}
	}
	return strings.Join(rows, "\n")
}

func formatBrowseDuration(ms int) string { return fmt.Sprintf("%02d:%02d", ms/60000, (ms/1000)%60) }
func truncateBrowse(value string, width int) string {
	if width < 4 {
		width = 4
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-1]) + "…"
}
