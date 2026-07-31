package tui

import (
	"fmt"
	"image"
	"testing"

	"github.com/dcbto/spotui/internal/catalog"
)

func TestCatalogCacheRemainsBounded(t *testing.T) {
	m := RootModel{browseState: browseState{browseCache: make(map[catalog.CacheKey]CatalogLoadedMsg)}}
	for i := 0; i < maxBrowseCacheEntries+4; i++ {
		m.browseOffset = i
		m.cacheCatalog(CatalogLoadedMsg{Route: catalog.Route{Kind: catalog.RouteLiked, Query: fmt.Sprint(i)}})
	}

	if got := len(m.browseCache); got != maxBrowseCacheEntries {
		t.Fatalf("catalog cache size: want %d, got %d", maxBrowseCacheEntries, got)
	}
	if _, exists := m.browseCache[(catalog.Route{Kind: catalog.RouteLiked, Query: "0"}).Key(0)]; exists {
		t.Fatal("catalog cache retained its oldest entry")
	}
}

func TestArtworkCacheRemainsBounded(t *testing.T) {
	m := RootModel{browseState: browseState{artwork: make(map[string]image.Image)}}
	for i := 0; i < maxArtworkEntries+4; i++ {
		m.cacheArtwork(fmt.Sprintf("https://image.test/%d", i), image.NewRGBA(image.Rect(0, 0, 1, 1)))
	}

	if got := len(m.artwork); got != maxArtworkEntries {
		t.Fatalf("artwork cache size: want %d, got %d", maxArtworkEntries, got)
	}
	if _, exists := m.artwork["https://image.test/0"]; exists {
		t.Fatal("artwork cache retained its oldest entry")
	}
}
