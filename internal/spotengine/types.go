package spotengine

type EventType string
type ErrorKind string

const (
	ErrorKindTransient          ErrorKind = "transient"
	ErrorKindCredentialRejected ErrorKind = "credential_rejected"
)

const (
	EventTypeReady            EventType = "ready"
	EventTypeBuffering        EventType = "buffering"
	EventTypePlaying          EventType = "playing"
	EventTypePaused           EventType = "paused"
	EventTypeStopped          EventType = "stopped"
	EventTypeActive           EventType = "active"
	EventTypeInactive         EventType = "inactive"
	EventTypeMetadata         EventType = "metadata"
	EventTypeVolume           EventType = "volume"
	EventTypeSeek             EventType = "seek"
	EventTypeShuffle          EventType = "shuffle"
	EventTypeAuthorizationURL EventType = "authorization_url"
	EventTypeError            EventType = "error"
	EventTypeAccountProduct   EventType = "account_product"
	EventTypeSessionEnded     EventType = "session_ended"
)

type Track struct {
	URI         string
	Name        string
	Artist      string
	Album       string
	AlbumURI    string
	DurationMS  int
	Explicit    bool
	ImageURL    string
	ExternalURL string
	TrackNumber int
	DiscNumber  int
}

type PageRequest struct {
	Offset int
	Limit  int
}

type TrackPage struct {
	Items  []Track
	Total  int
	Offset int
	Limit  int
}

type PlaylistSummary struct {
	URI         string
	Name        string
	Owner       string
	Description string
	TrackCount  int
	ImageURL    string
	ExternalURL string
}

type PlaylistDetail struct {
	PlaylistSummary
	Tracks TrackPage
}

type AlbumSummary struct {
	URI         string
	Name        string
	Artist      string
	ReleaseDate string
	TrackCount  int
	ImageURL    string
	ExternalURL string
}

type AlbumDetail struct {
	AlbumSummary
	Tracks TrackPage
}

type ArtistSummary struct {
	URI         string
	Name        string
	ImageURL    string
	ExternalURL string
}

type ArtistDetail struct {
	ArtistSummary
	Genres               []string
	Popular              TrackPage
	Albums               []AlbumSummary
	Playlists            []PlaylistSummary
	PlaylistsUnavailable bool
}

type CatalogPage[T any] struct {
	Items  []T
	Total  int
	Offset int
	Limit  int
}

type SearchGroups struct {
	Tracks                      TrackPage
	Albums                      CatalogPage[AlbumSummary]
	Artists                     CatalogPage[ArtistSummary]
	Playlists                   CatalogPage[PlaylistSummary]
	AlbumsAndArtistsUnavailable bool
	PlaylistsUnavailable        bool
}

type RecommendedPage struct {
	Artists              CatalogPage[ArtistSummary]
	Tracks               TrackPage
	Albums               []AlbumSummary
	Playlists            []PlaylistSummary
	TracksUnavailable    bool
	PlaylistsUnavailable bool
}

type Event struct {
	Type       EventType
	Track      *Track
	ContextURI string
	URI        string
	PositionMS int
	DurationMS int
	Volume     int
	VolumeMax  int
	Shuffle    bool
	URL        string
	Err        error
	ErrorKind  ErrorKind
	Product    string
}
