package library

type Library struct {
	playlists []PlaylistEntry
	tracks    []TrackEntry
	artists   []ArtistEntry
	search    []TrackEntry
	cursors   [4]int
	active    Tab
}

func New() *Library {
	return &Library{}
}

func (l *Library) SetPlaylists(p []PlaylistEntry) {
	l.playlists = p
}

func (l *Library) SetTracks(t []TrackEntry) {
	l.tracks = t
}

func (l *Library) SetArtists(a []ArtistEntry) {
	l.artists = a
}

func (l *Library) SetSearchResults(t []TrackEntry) {
	l.search = t
	if l.cursors[TabSearch] >= len(l.search) {
		l.cursors[TabSearch] = len(l.search) - 1
	}
	if l.cursors[TabSearch] < 0 {
		l.cursors[TabSearch] = 0
	}
}

func (l *Library) SearchResultCount() int {
	return len(l.search)
}

func (l *Library) ActiveTab() Tab {
	return l.active
}

func (l *Library) SetActiveTab(tab Tab) {
	l.active = tab
}

func (l *Library) Cursor() int {
	return l.cursors[l.active]
}

func (l *Library) MoveDown() {
	if l.cursors[l.active] < l.listLen()-1 {
		l.cursors[l.active]++
	}
}

func (l *Library) MoveUp() {
	if l.cursors[l.active] > 0 {
		l.cursors[l.active]--
	}
}

func (l *Library) SelectedURI() string {
	cur := l.cursors[l.active]
	switch l.active {
	case TabPlaylists:
		if cur < 0 || cur >= len(l.playlists) {
			return ""
		}
		return l.playlists[cur].URI
	case TabTracks:
		if cur < 0 || cur >= len(l.tracks) {
			return ""
		}
		return l.tracks[cur].URI
	case TabSearch:
		if cur < 0 || cur >= len(l.search) {
			return ""
		}
		return l.search[cur].URI
	}
	return ""
}

func (l *Library) listLen() int {
	switch l.active {
	case TabPlaylists:
		return len(l.playlists)
	case TabTracks:
		return len(l.tracks)
	case TabArtists:
		return len(l.artists)
	case TabSearch:
		return len(l.search)
	}
	return 0
}
