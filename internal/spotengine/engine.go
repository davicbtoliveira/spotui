package spotengine

import "context"

type Engine interface {
	HasSession() bool
	AutoplayEnabled() bool
	Start(context.Context) error
	Reconnect(context.Context) error
	CancelLogin(context.Context) error
	Logout(context.Context) error
	SearchTracks(context.Context, SearchRequest) (SearchPage, error)
	LikedTracks(context.Context, PageRequest) (TrackPage, error)
	UserPlaylists(context.Context, PageRequest) (CatalogPage[PlaylistSummary], error)
	SavedAlbums(context.Context, PageRequest) (CatalogPage[AlbumSummary], error)
	Search(context.Context, SearchRequest) (SearchGroups, error)
	Playlist(context.Context, string, PageRequest) (PlaylistDetail, error)
	Album(context.Context, string, PageRequest) (AlbumDetail, error)
	Artist(context.Context, string, PageRequest) (ArtistDetail, error)
	TopArtists(context.Context, PageRequest) (CatalogPage[ArtistSummary], error)
	TopTracks(context.Context, PageRequest) (TrackPage, error)
	Recommended(context.Context, PageRequest) (RecommendedPage, error)
	Play(context.Context, string) error
	PlayContext(context.Context, string, string, int) error
	Pause(context.Context) error
	Resume(context.Context) error
	Next(context.Context) error
	Previous(context.Context) error
	SetVolume(context.Context, int) error
	SetAutoplay(context.Context, bool) error
	SetShuffle(context.Context, bool) error
	SeekRelative(context.Context, int) error
	Events() <-chan Event
	Close(context.Context) error
}

type SearchRequest struct {
	Query  string
	Offset int
	Limit  int
}

type SearchPage struct {
	Tracks []Track
	Total  int
	Offset int
}
