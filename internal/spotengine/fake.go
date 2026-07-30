package spotengine

import (
	"context"
	"sync"
)

type Operation string

const (
	OperationStart         Operation = "start"
	OperationReconnect     Operation = "reconnect"
	OperationCancelLogin   Operation = "cancel_login"
	OperationLogout        Operation = "logout"
	OperationSearchTracks  Operation = "search_tracks"
	OperationLikedTracks   Operation = "liked_tracks"
	OperationUserPlaylists Operation = "user_playlists"
	OperationSavedAlbums   Operation = "saved_albums"
	OperationSearch        Operation = "search"
	OperationPlaylist      Operation = "playlist"
	OperationAlbum         Operation = "album"
	OperationArtist        Operation = "artist"
	OperationTopArtists    Operation = "top_artists"
	OperationTopTracks     Operation = "top_tracks"
	OperationRecommended   Operation = "recommended"
	OperationPlay          Operation = "play"
	OperationPlayContext   Operation = "play_context"
	OperationPause         Operation = "pause"
	OperationResume        Operation = "resume"
	OperationNext          Operation = "next"
	OperationPrevious      Operation = "previous"
	OperationSetVolume     Operation = "set_volume"
	OperationSetAutoplay   Operation = "set_autoplay"
	OperationSetShuffle    Operation = "set_shuffle"
	OperationSeekRelative  Operation = "seek_relative"
	OperationClose         Operation = "close"
)

type Call struct {
	Operation    Operation
	URI          string
	Search       SearchRequest
	Volume       int
	Enabled      bool
	Page         PageRequest
	URI2         string
	DeltaMS      int
	SearchGroups SearchGroups
}

type Fake struct {
	mu              sync.Mutex
	calls           []Call
	events          chan Event
	errors          map[Operation]error
	searchPage      SearchPage
	searchError     error
	likedTracks     TrackPage
	playlists       CatalogPage[PlaylistSummary]
	albums          CatalogPage[AlbumSummary]
	searchGroups    SearchGroups
	playlistDetails map[string]PlaylistDetail
	albumDetails    map[string]AlbumDetail
	artistDetails   map[string]ArtistDetail
	topArtists      CatalogPage[ArtistSummary]
	topTracks       TrackPage
	recommended     RecommendedPage
	hasSession      bool
	autoplay        bool
}

func (f *Fake) HasSession() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.hasSession
}

func (f *Fake) SetHasSession(hasSession bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.hasSession = hasSession
}

func NewFake() *Fake {
	return &Fake{
		events:          make(chan Event, 16),
		errors:          make(map[Operation]error),
		autoplay:        true,
		playlistDetails: make(map[string]PlaylistDetail),
		albumDetails:    make(map[string]AlbumDetail),
		artistDetails:   make(map[string]ArtistDetail),
	}
}

func (f *Fake) AutoplayEnabled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.autoplay
}

func (f *Fake) Emit(event Event) {
	f.events <- event
}

func (f *Fake) Events() <-chan Event {
	return f.events
}

func (f *Fake) Play(_ context.Context, uri string) error {
	return f.record(Call{Operation: OperationPlay, URI: uri})
}

func (f *Fake) PlayContext(_ context.Context, contextURI, trackURI string, positionMS int) error {
	return f.record(Call{Operation: OperationPlayContext, URI: contextURI, URI2: trackURI, DeltaMS: positionMS})
}

func (f *Fake) Start(_ context.Context) error {
	return f.record(Call{Operation: OperationStart})
}

func (f *Fake) Reconnect(_ context.Context) error {
	return f.record(Call{Operation: OperationReconnect})
}

func (f *Fake) CancelLogin(_ context.Context) error {
	return f.record(Call{Operation: OperationCancelLogin})
}

func (f *Fake) Logout(_ context.Context) error {
	if err := f.record(Call{Operation: OperationLogout}); err != nil {
		return err
	}
	f.SetHasSession(false)
	return nil
}

func (f *Fake) SearchTracks(_ context.Context, request SearchRequest) (SearchPage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, Call{Operation: OperationSearchTracks, Search: request})
	if err := f.errors[OperationSearchTracks]; err != nil {
		return SearchPage{}, err
	}
	return f.searchPage, f.searchError
}

func (f *Fake) LikedTracks(_ context.Context, request PageRequest) (TrackPage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationLikedTracks, Page: request})
	if err := f.errors[OperationLikedTracks]; err != nil {
		return TrackPage{}, err
	}
	return f.likedTracks, nil
}

func (f *Fake) UserPlaylists(_ context.Context, request PageRequest) (CatalogPage[PlaylistSummary], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationUserPlaylists, Page: request})
	if err := f.errors[OperationUserPlaylists]; err != nil {
		return CatalogPage[PlaylistSummary]{}, err
	}
	return f.playlists, nil
}

func (f *Fake) SavedAlbums(_ context.Context, request PageRequest) (CatalogPage[AlbumSummary], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationSavedAlbums, Page: request})
	if err := f.errors[OperationSavedAlbums]; err != nil {
		return CatalogPage[AlbumSummary]{}, err
	}
	return f.albums, nil
}

func (f *Fake) Search(_ context.Context, request SearchRequest) (SearchGroups, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationSearch, Search: request})
	if err := f.errors[OperationSearch]; err != nil {
		return SearchGroups{}, err
	}
	return f.searchGroups, nil
}

func (f *Fake) Playlist(_ context.Context, uri string, request PageRequest) (PlaylistDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationPlaylist, URI: uri, Page: request})
	if err := f.errors[OperationPlaylist]; err != nil {
		return PlaylistDetail{}, err
	}
	return f.playlistDetails[uri], nil
}

func (f *Fake) Album(_ context.Context, uri string, request PageRequest) (AlbumDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationAlbum, URI: uri, Page: request})
	if err := f.errors[OperationAlbum]; err != nil {
		return AlbumDetail{}, err
	}
	return f.albumDetails[uri], nil
}

func (f *Fake) Artist(_ context.Context, uri string, request PageRequest) (ArtistDetail, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationArtist, URI: uri, Page: request})
	if err := f.errors[OperationArtist]; err != nil {
		return ArtistDetail{}, err
	}
	return f.artistDetails[uri], nil
}

func (f *Fake) TopArtists(_ context.Context, request PageRequest) (CatalogPage[ArtistSummary], error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationTopArtists, Page: request})
	if err := f.errors[OperationTopArtists]; err != nil {
		return CatalogPage[ArtistSummary]{}, err
	}
	return f.topArtists, nil
}

func (f *Fake) TopTracks(_ context.Context, request PageRequest) (TrackPage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationTopTracks, Page: request})
	if err := f.errors[OperationTopTracks]; err != nil {
		return TrackPage{}, err
	}
	return f.topTracks, nil
}

func (f *Fake) Recommended(_ context.Context, request PageRequest) (RecommendedPage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, Call{Operation: OperationRecommended, Page: request})
	if err := f.errors[OperationRecommended]; err != nil {
		return RecommendedPage{}, err
	}
	return f.recommended, nil
}

func (f *Fake) Pause(_ context.Context) error {
	return f.record(Call{Operation: OperationPause})
}

func (f *Fake) Resume(_ context.Context) error {
	return f.record(Call{Operation: OperationResume})
}

func (f *Fake) Next(_ context.Context) error {
	return f.record(Call{Operation: OperationNext})
}

func (f *Fake) Previous(_ context.Context) error {
	return f.record(Call{Operation: OperationPrevious})
}

func (f *Fake) SetVolume(_ context.Context, volume int) error {
	return f.record(Call{Operation: OperationSetVolume, Volume: volume})
}

func (f *Fake) SetAutoplay(_ context.Context, enabled bool) error {
	if err := f.record(Call{Operation: OperationSetAutoplay, Enabled: enabled}); err != nil {
		return err
	}
	f.mu.Lock()
	f.autoplay = enabled
	f.mu.Unlock()
	return nil
}

func (f *Fake) SetShuffle(_ context.Context, enabled bool) error {
	return f.record(Call{Operation: OperationSetShuffle, Enabled: enabled})
}

func (f *Fake) SeekRelative(_ context.Context, deltaMS int) error {
	return f.record(Call{Operation: OperationSeekRelative, DeltaMS: deltaMS})
}

func (f *Fake) Close(_ context.Context) error {
	return f.record(Call{Operation: OperationClose})
}

func (f *Fake) SetSearchResult(page SearchPage, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.searchPage = page
	f.searchError = err
}

func (f *Fake) SetLikedTracks(page TrackPage) { f.mu.Lock(); defer f.mu.Unlock(); f.likedTracks = page }
func (f *Fake) SetUserPlaylists(page CatalogPage[PlaylistSummary]) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.playlists = page
}
func (f *Fake) SetSavedAlbums(page CatalogPage[AlbumSummary]) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.albums = page
}
func (f *Fake) SetSearchGroups(groups SearchGroups) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.searchGroups = groups
}
func (f *Fake) SetPlaylistDetail(uri string, detail PlaylistDetail) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.playlistDetails[uri] = detail
}
func (f *Fake) SetAlbumDetail(uri string, detail AlbumDetail) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.albumDetails[uri] = detail
}
func (f *Fake) SetArtistDetail(uri string, detail ArtistDetail) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.artistDetails[uri] = detail
}
func (f *Fake) SetTopArtists(page CatalogPage[ArtistSummary]) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.topArtists = page
}
func (f *Fake) SetTopTracks(page TrackPage) { f.mu.Lock(); defer f.mu.Unlock(); f.topTracks = page }
func (f *Fake) SetRecommended(page RecommendedPage) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recommended = page
}

func (f *Fake) SetError(operation Operation, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.errors[operation] = err
}

func (f *Fake) record(call Call) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, call)
	return f.errors[call.Operation]
}

func (f *Fake) Calls() []Call {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]Call(nil), f.calls...)
}
