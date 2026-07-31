package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/catalog"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/tui/commands"
)

var navLabels = []string{"Library", "  Liked Tracks", "  Playlists", "  Saved Albums", "Recommended"}

func (m *RootModel) loadBrowseRoute() tea.Cmd {
	m.browseLoading = true
	m.browseError = ""
	m.browseRequestID++
	requestID := m.browseRequestID
	if m.browseRoute.Kind == catalog.RouteSearch {
		m.browseRoute.Query = m.searchQuery
	}
	cacheKey := m.browseRoute.Key(m.browseOffset)
	if cached, ok := m.browseCache[cacheKey]; ok {
		cached.RequestID = requestID
		return func() tea.Msg { return cached }
	}
	switch m.browseRoute.Kind {
	case catalog.RouteLiked:
		return commands.CmdLikedTracks(m.engine, m.browsePageRequest(), requestID)
	case catalog.RoutePlaylists:
		return commands.CmdUserPlaylists(m.engine, m.browsePageRequest(), requestID)
	case catalog.RouteAlbums:
		return commands.CmdSavedAlbums(m.engine, m.browsePageRequest(), requestID)
	case catalog.RouteRecommended:
		return commands.CmdRecommended(m.engine, spotengine.PageRequest{Offset: m.browseOffset, Limit: 10}, requestID)
	case catalog.RouteSearch:
		return commands.CmdSearch(m.engine, spotengine.SearchRequest{Query: m.browseRoute.Query, Offset: m.browseOffset, Limit: 10}, requestID)
	case catalog.RoutePlaylist:
		return commands.CmdPlaylist(m.engine, m.browseRoute.URI, m.browsePageRequest(), requestID)
	case catalog.RouteAlbum:
		return commands.CmdAlbum(m.engine, m.browseRoute.URI, m.browsePageRequest(), requestID)
	case catalog.RouteArtist:
		return commands.CmdArtist(m.engine, m.browseRoute.URI, m.browsePageRequest(), requestID)
	default:
		m.browseLoading = false
		return nil
	}
}

func (m RootModel) browsePageRequest() spotengine.PageRequest {
	return spotengine.PageRequest{Offset: m.browseOffset, Limit: 10}
}

func (m *RootModel) applyCatalogMessage(msg CatalogLoadedMsg) tea.Cmd {
	if msg.Route != m.browseRoute || (msg.RequestID != 0 && msg.RequestID != m.browseRequestID) {
		return nil
	}
	m.browseLoading = false
	requestedOffset := m.browseOffset
	if msg.Err != nil {
		m.browseError = msg.Err.Error()
		return nil
	}
	m.browseError = ""
	m.browseCursor = 0
	// Preserve the requested page offset; stale responses are rejected above.
	m.browseTotal = 0
	m.browseContextURI = ""
	m.browseMeta = ""
	switch msg.Route.Kind {
	case catalog.RouteLiked:
		payload, ok := msg.Payload.(catalog.TrackPagePayload)
		if !ok {
			m.browseError = "catalog response missing tracks"
			return nil
		}
		value := payload.Value
		m.browseOffset, m.browseTotal = pageOffset(value.Offset, requestedOffset), value.Total
		m.browseItems = make([]browseItem, 0, len(value.Items))
		for _, item := range value.Items {
			m.browseItems = append(m.browseItems, trackItem(item))
		}
	case catalog.RoutePlaylists:
		payload, ok := msg.Payload.(catalog.PlaylistPagePayload)
		if !ok {
			m.browseError = "catalog response missing playlists"
			return nil
		}
		value := payload.Value
		m.browseOffset, m.browseTotal = pageOffset(value.Offset, requestedOffset), value.Total
		m.browseItems = make([]browseItem, 0, len(value.Items))
		for _, item := range value.Items {
			m.browseItems = append(m.browseItems, playlistItem(item))
		}
	case catalog.RouteAlbums:
		payload, ok := msg.Payload.(catalog.AlbumPagePayload)
		if !ok {
			m.browseError = "catalog response missing albums"
			return nil
		}
		value := payload.Value
		m.browseOffset, m.browseTotal = pageOffset(value.Offset, requestedOffset), value.Total
		m.browseItems = make([]browseItem, 0, len(value.Items))
		for _, item := range value.Items {
			m.browseItems = append(m.browseItems, albumItem(item))
		}
	case catalog.RouteSearch:
		payload, ok := msg.Payload.(catalog.SearchPayload)
		if !ok {
			m.browseError = "catalog response missing search results"
			return nil
		}
		value := payload.Value
		m.browseTotal = maxSearchTotal(value)
		m.browseItems = searchItems(value)
	case catalog.RouteRecommended:
		payload, ok := msg.Payload.(catalog.RecommendedPayload)
		if !ok {
			m.browseError = "catalog response missing recommendations"
			return nil
		}
		value := payload.Value
		m.browseItems = recommendedItems(value)
		m.browseTotal = max(value.Artists.Total, value.Tracks.Total)
	case catalog.RoutePlaylist:
		payload, ok := msg.Payload.(catalog.PlaylistPayload)
		if !ok {
			m.browseError = "catalog response missing playlist"
			return nil
		}
		value := payload.Value
		m.browseOffset, m.browseTotal = pageOffset(value.Tracks.Offset, requestedOffset), value.Tracks.Total
		m.browseTitle = value.Name
		m.browseContextURI = value.URI
		m.browseMeta = fmt.Sprintf("%s · %d tracks", value.Owner, value.TrackCount)
		m.browseItems = make([]browseItem, 0, len(value.Tracks.Items))
		for _, item := range value.Tracks.Items {
			m.browseItems = append(m.browseItems, trackItem(item))
		}
	case catalog.RouteAlbum:
		payload, ok := msg.Payload.(catalog.AlbumPayload)
		if !ok {
			m.browseError = "catalog response missing album"
			return nil
		}
		value := payload.Value
		m.browseOffset, m.browseTotal = pageOffset(value.Tracks.Offset, requestedOffset), value.Tracks.Total
		m.browseTitle = value.Name
		m.browseContextURI = value.URI
		m.browseMeta = fmt.Sprintf("%s · %s · %d tracks", value.Artist, value.ReleaseDate, value.TrackCount)
		m.browseItems = make([]browseItem, 0, len(value.Tracks.Items))
		for _, item := range value.Tracks.Items {
			if item.Name == "" {
				continue
			}
			m.browseItems = append(m.browseItems, trackItem(item))
		}
	case catalog.RouteArtist:
		payload, ok := msg.Payload.(catalog.ArtistPayload)
		if !ok {
			m.browseError = "catalog response missing artist"
			return nil
		}
		value := payload.Value
		m.browseTitle = value.Name
		m.browseMeta = strings.Join(value.Genres, " · ")
		m.browseItems = make([]browseItem, 0, len(value.Popular.Items)+len(value.Albums)+len(value.Playlists)+2)
		if value.ImageURL != "" {
			m.browseItems = append(m.browseItems, browseItem{kind: browseItemHeader, Title: "Artist artwork", ImageURL: value.ImageURL})
		}
		for _, item := range value.Popular.Items {
			if item.Name == "" {
				continue
			}
			m.browseItems = append(m.browseItems, trackItem(item))
		}
		for _, item := range value.Albums {
			if item.Name == "" {
				continue
			}
			m.browseItems = append(m.browseItems, albumItem(item))
		}
		if len(value.Playlists) > 0 || value.PlaylistsUnavailable {
			title := "Playlists featuring this artist"
			if value.PlaylistsUnavailable {
				title += " (unavailable)"
			}
			m.browseItems = append(m.browseItems, browseItem{kind: browseItemHeader, Title: title})
		}
		for _, item := range value.Playlists {
			if item.Name == "" {
				continue
			}
			m.browseItems = append(m.browseItems, playlistItem(item))
		}
	}
	cacheMsg := msg
	cacheMsg.RequestID = 0
	m.cacheCatalog(cacheMsg)
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
	if len(m.artworkLoading) >= maxArtworkEntries {
		return nil
	}
	m.artworkLoading[url] = true
	return commands.CmdLoadArtwork(url)
}

func searchItems(groups spotengine.SearchGroups) []browseItem {
	items := make([]browseItem, 0)
	items = append(items, browseItem{kind: browseItemHeader, Title: fmt.Sprintf("Tracks (%d)", groups.Tracks.Total)})
	for _, item := range groups.Tracks.Items {
		i := trackItem(item)
		i.Subtitle = "Track · " + i.Subtitle
		items = append(items, i)
	}
	items = append(items, browseItem{kind: browseItemHeader, Title: searchGroupTitle("Albums", groups.Albums.Total, groups.AlbumsAndArtistsUnavailable)})
	for _, item := range groups.Albums.Items {
		i := albumItem(item)
		i.Subtitle = "Album · " + i.Subtitle
		items = append(items, i)
	}

	items = append(items, browseItem{kind: browseItemHeader, Title: searchGroupTitle("Artists", groups.Artists.Total, groups.AlbumsAndArtistsUnavailable)})
	for _, item := range groups.Artists.Items {
		i := artistItem(item)
		i.Subtitle = "Artist"
		items = append(items, i)
	}
	// Playlists are the final search group.
	items = append(items, browseItem{kind: browseItemHeader, Title: searchGroupTitle("Playlists", groups.Playlists.Total, groups.PlaylistsUnavailable)})
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
	items := []browseItem{{kind: browseItemHeader, Title: fmt.Sprintf("Your top artists (%d)", value.Artists.Total)}}
	for _, artist := range value.Artists.Items {
		items = append(items, artistItem(artist))
	}
	trackTitle := fmt.Sprintf("Your top tracks (%d)", value.Tracks.Total)
	if value.TracksUnavailable {
		trackTitle = "Your top tracks (unavailable)"
	}
	items = append(items, browseItem{kind: browseItemHeader, Title: trackTitle})
	for _, track := range value.Tracks.Items {
		items = append(items, trackItem(track))
	}
	items = append(items, browseItem{kind: browseItemHeader, Title: fmt.Sprintf("Albums from top artists (%d)", len(value.Albums))})
	for _, album := range value.Albums {
		items = append(items, albumItem(album))
	}
	playlistTitle := fmt.Sprintf("Recommended playlists (%d)", len(value.Playlists))
	if value.PlaylistsUnavailable {
		playlistTitle = "Recommended playlists (unavailable)"
	}
	items = append(items, browseItem{kind: browseItemHeader, Title: playlistTitle})
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
		m.browseRoute = catalog.Route{Kind: catalog.RouteLiked}
		m.browseTitle = "Liked Tracks"
	case 2:
		m.browseRoute = catalog.Route{Kind: catalog.RoutePlaylists}
		m.browseTitle = "Playlists"
	case 3:
		m.browseRoute = catalog.Route{Kind: catalog.RouteAlbums}
		m.browseTitle = "Saved Albums"
	case 4:
		m.browseRoute = catalog.Route{Kind: catalog.RouteRecommended}
		m.browseTitle = "Recommended"
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
	case browseItemTrack:
		if m.browseContextURI != "" {
			return commands.CmdPlayContext(m.engine, m.browseContextURI, item.URI, 0)
		}
		return commands.CmdPlayEngineTrack(m.engine, item.URI)
	case browseItemPlaylist, browseItemAlbum, browseItemArtist:
		m.browseStack = append(m.browseStack, browseSnapshot{route: m.browseRoute, title: m.browseTitle, items: append([]browseItem(nil), m.browseItems...), cursor: m.browseCursor, contextURI: m.browseContextURI, offset: m.browseOffset, meta: m.browseMeta})
		switch item.kind {
		case browseItemPlaylist:
			m.browseRoute = catalog.Route{Kind: catalog.RoutePlaylist, URI: item.URI}
		case browseItemAlbum:
			m.browseRoute = catalog.Route{Kind: catalog.RouteAlbum, URI: item.URI}
		case browseItemArtist:
			m.browseRoute = catalog.Route{Kind: catalog.RouteArtist, URI: item.URI}
		default:
			return nil
		}
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
	m.browseRoute, m.browseTitle, m.browseItems, m.browseCursor, m.browseContextURI, m.browseOffset, m.browseMeta = last.route, last.title, last.items, last.cursor, last.contextURI, last.offset, last.meta
	m.browseLoading = false
	m.browseError = ""
	return true
}

func (m RootModel) handleBrowseKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.searchInputActive {
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
			m.browseRoute = catalog.Route{Kind: catalog.RouteSearch, Query: m.searchQuery}
			m.browseTitle = "Search"
			return m, m.loadBrowseRoute()
		}
	}

	if msg.Type == tea.KeyTab || msg.String() == "tab" {
		m.browseFocus = 1 - m.browseFocus
		if m.browseFocus == 0 {
			m.searchInputActive = false
		}
		return m, nil
	}
	if msg.String() == "shift+tab" {
		m.browseFocus = 1 - m.browseFocus
		if m.browseFocus == 0 {
			m.searchInputActive = false
		}
		return m, nil
	}
	if msg.String() == KeySearch {
		m.browseFocus = 1
		m.searchInputActive = true
		m.browseRoute = catalog.Route{Kind: catalog.RouteSearch, Query: m.searchQuery}
		m.browseTitle = "Search"
		m.browseItems = nil
		m.browseCursor = 0
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

	if cmd, handled := m.handlePlaybackKey(msg); handled {
		return m, cmd
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
	if len(items) == 0 || cursor <= 0 {
		return 0
	}
	if cursor > len(items) {
		cursor = len(items)
	}
	for i := cursor - 1; i >= 0; i-- {
		if items[i].kind != browseItemHeader {
			return i
		}
	}
	return cursor
}

func nextBrowseItem(items []browseItem, cursor int) int {
	if len(items) == 0 {
		return 0
	}
	if cursor < -1 {
		cursor = -1
	}
	if cursor >= len(items) {
		return len(items) - 1
	}
	for i := cursor + 1; i < len(items); i++ {
		if items[i].kind != browseItemHeader {
			return i
		}
	}
	return cursor
}
