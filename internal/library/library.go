package library

type Library struct {
	search []TrackEntry
	cursor int
}

func New() *Library {
	return &Library{}
}

func (l *Library) SetSearchResults(t []TrackEntry) {
	l.search = t
	if l.cursor >= len(l.search) {
		l.cursor = len(l.search) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
}

func (l *Library) SearchResultCount() int {
	return len(l.search)
}

func (l *Library) Cursor() int {
	return l.cursor
}

func (l *Library) MoveDown() {
	if l.cursor < len(l.search)-1 {
		l.cursor++
	}
}

func (l *Library) MoveUp() {
	if l.cursor > 0 {
		l.cursor--
	}
}

func (l *Library) SelectedURI() string {
	if l.cursor < 0 || l.cursor >= len(l.search) {
		return ""
	}
	return l.search[l.cursor].URI
}
