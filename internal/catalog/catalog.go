package catalog

import "github.com/dcbto/spotui/internal/spotengine"

type RouteKind string

const (
	RouteLiked       RouteKind = "liked"
	RoutePlaylists   RouteKind = "playlists"
	RouteAlbums      RouteKind = "albums"
	RouteRecommended RouteKind = "recommended"
	RouteSearch      RouteKind = "search"
	RoutePlaylist    RouteKind = "playlist"
	RouteAlbum       RouteKind = "album"
	RouteArtist      RouteKind = "artist"
)

type Route struct {
	Kind  RouteKind
	URI   string
	Query string
}

type CacheKey struct {
	Route  Route
	Offset int
}

func (r Route) Key(offset int) CacheKey {
	return CacheKey{Route: r, Offset: offset}
}

// Payload is the typed response body for a catalog route.
//
// The unexported method keeps the set of valid payloads closed to this
// package, so a route cannot accidentally be paired with an arbitrary value.
type Payload interface {
	isCatalogPayload()
}

type TrackPagePayload struct{ Value spotengine.TrackPage }

func (TrackPagePayload) isCatalogPayload() {}

type PlaylistPagePayload struct {
	Value spotengine.CatalogPage[spotengine.PlaylistSummary]
}

func (PlaylistPagePayload) isCatalogPayload() {}

type AlbumPagePayload struct {
	Value spotengine.CatalogPage[spotengine.AlbumSummary]
}

func (AlbumPagePayload) isCatalogPayload() {}

type SearchPayload struct{ Value spotengine.SearchGroups }

func (SearchPayload) isCatalogPayload() {}

type RecommendedPayload struct{ Value spotengine.RecommendedPage }

func (RecommendedPayload) isCatalogPayload() {}

type PlaylistPayload struct{ Value spotengine.PlaylistDetail }

func (PlaylistPayload) isCatalogPayload() {}

type AlbumPayload struct{ Value spotengine.AlbumDetail }

func (AlbumPayload) isCatalogPayload() {}

type ArtistPayload struct{ Value spotengine.ArtistDetail }

func (ArtistPayload) isCatalogPayload() {}

type Result struct {
	Route     Route
	RequestID uint64
	Payload   Payload
	Err       error
}
