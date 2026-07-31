package tui

import (
	"image"

	"github.com/dcbto/spotui/internal/catalog"
	"github.com/dcbto/spotui/internal/spotengine"
)

type playerState struct {
	localProgressMs       int
	engineTrack           *spotengine.Track
	enginePlaying         bool
	engineBuffering       bool
	engineActive          bool
	engineVolume          int
	confirmedEngineVolume int
	volumeCommandInFlight bool
	volumeDebouncePending bool
	volumeDebounceID      uint64
	engineAutoplay        bool
	engineShuffle         bool
	engineTransferred     bool
}

type browseState struct {
	browseInitialized bool
	navCursor         int
	browseFocus       int
	browseRoute       catalog.Route
	browseTitle       string
	browseItems       []browseItem
	browseCursor      int
	browseLoading     bool
	browseError       string
	browseOffset      int
	browseTotal       int
	browseContextURI  string
	browseStack       []browseSnapshot
	browseRequestID   uint64
	browseMeta        string
	browseCache       map[catalog.CacheKey]CatalogLoadedMsg
	browseCacheOrder  []catalog.CacheKey
	artwork           map[string]image.Image
	artworkLoading    map[string]bool
	artworkOrder      []string
	searchInputActive bool
	searchQuery       string
}
