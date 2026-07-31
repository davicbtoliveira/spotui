package tui

import (
	"github.com/dcbto/spotui/internal/catalog"
	"github.com/dcbto/spotui/internal/tui/views"
)

func (m RootModel) renderBrowseShell() string {
	items := make([]views.BrowseItem, len(m.browseItems))
	for i, item := range m.browseItems {
		items[i] = views.BrowseItem{
			Header:      item.kind == browseItemHeader,
			Title:       item.Title,
			Subtitle:    item.Subtitle,
			DurationMS:  item.DurationMS,
			ImageURL:    item.ImageURL,
			ExternalURL: item.ExternalURL,
		}
	}
	return views.RenderBrowseShell(views.BrowseViewState{
		Width:            m.width,
		Height:           m.height,
		NavigationCursor: m.navCursor,
		NavigationFocus:  m.browseFocus == 0,
		NavigationLabels: navLabels,
		Title:            m.browseTitle,
		Meta:             m.browseMeta,
		Search:           m.browseRoute.Kind == catalog.RouteSearch,
		SearchQuery:      m.searchQuery,
		Loading:          m.browseLoading,
		Error:            m.browseError,
		Cursor:           m.browseCursor,
		Focus:            m.browseFocus,
		Items:            items,
		Artwork:          m.artwork,
		StatusMessage:    m.statusMsg,
		StatusIsError:    m.statusIsErr,
		Player: views.EnginePlayerState{
			Track:       m.engineTrack,
			ProgressMS:  m.localProgressMs,
			Playing:     m.enginePlaying,
			Buffering:   m.engineBuffering,
			Active:      m.engineActive,
			Volume:      m.engineVolume,
			Autoplay:    m.engineAutoplay,
			Shuffle:     m.engineShuffle,
			Transferred: m.engineTransferred,
		},
	})
}
