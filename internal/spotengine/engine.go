package spotengine

import "context"

type SessionEngine interface {
	HasSession() bool
	AutoplayEnabled() bool
	Start(context.Context) error
	Reconnect(context.Context) error
	CancelLogin(context.Context) error
	Logout(context.Context) error
	Close(context.Context) error
}

type CatalogEngine interface {
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
}

type PlaybackEngine interface {
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
}

type EventSource interface {
	Events() <-chan Event
}

type Engine interface {
	SessionEngine
	CatalogEngine
	PlaybackEngine
	EventSource
}

type SearchRequest struct {
	Query  string
	Offset int
	Limit  int
}
