package tui

import (
	"image"

	"github.com/dcbto/spotui/internal/catalog"
)

const (
	maxBrowseCacheEntries = 16
	maxArtworkEntries     = 32
)

func (m *RootModel) cacheCatalog(msg CatalogLoadedMsg) {
	if m.browseCache == nil {
		m.browseCache = make(map[catalog.CacheKey]CatalogLoadedMsg)
	}
	key := msg.Route.Key(m.browseOffset)
	setBounded(m.browseCache, &m.browseCacheOrder, maxBrowseCacheEntries, key, msg)
}

func (m *RootModel) cacheArtwork(url string, value image.Image) {
	if m.artwork == nil {
		m.artwork = make(map[string]image.Image)
	}
	setBounded(m.artwork, &m.artworkOrder, maxArtworkEntries, url, value)
}

func setBounded[K comparable, V any](values map[K]V, order *[]K, limit int, key K, value V) {
	if _, exists := values[key]; !exists {
		if len(*order) >= limit {
			oldest := (*order)[0]
			*order = (*order)[1:]
			delete(values, oldest)
		}
		*order = append(*order, key)
	}
	values[key] = value
}
