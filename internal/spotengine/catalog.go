package spotengine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devgianlu/go-librespot/daemon"
)

type spotifyPage struct {
	Items  []json.RawMessage `json:"items"`
	Total  int               `json:"total"`
	Offset int               `json:"offset"`
	Limit  int               `json:"limit"`
}

type spotifyImage struct {
	URL string `json:"url"`
}

type spotifyArtist struct {
	ID           string            `json:"id"`
	URI          string            `json:"uri"`
	Name         string            `json:"name"`
	Genres       []string          `json:"genres"`
	Images       []spotifyImage    `json:"images"`
	ExternalURLs map[string]string `json:"external_urls"`
}

type spotifyAlbum struct {
	ID           string            `json:"id"`
	URI          string            `json:"uri"`
	Name         string            `json:"name"`
	Artists      []spotifyArtist   `json:"artists"`
	ReleaseDate  string            `json:"release_date"`
	TotalTracks  int               `json:"total_tracks"`
	Images       []spotifyImage    `json:"images"`
	ExternalURLs map[string]string `json:"external_urls"`
}

type spotifyTrack struct {
	ID           string            `json:"id"`
	URI          string            `json:"uri"`
	Name         string            `json:"name"`
	DurationMS   int               `json:"duration_ms"`
	Explicit     bool              `json:"explicit"`
	TrackNumber  int               `json:"track_number"`
	DiscNumber   int               `json:"disc_number"`
	Artists      []spotifyArtist   `json:"artists"`
	Album        spotifyAlbum      `json:"album"`
	ExternalURLs map[string]string `json:"external_urls"`
}

type spotifyPlaylist struct {
	ID           string            `json:"id"`
	URI          string            `json:"uri"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Owner        spotifyArtist     `json:"owner"`
	Images       []spotifyImage    `json:"images"`
	ExternalURLs map[string]string `json:"external_urls"`
	Tracks       struct {
		Total int               `json:"total"`
		Items []json.RawMessage `json:"items"`
	} `json:"tracks"`
}

type spotifySavedTrack struct {
	Track spotifyTrack `json:"track"`
}
type spotifySavedAlbum struct {
	Album spotifyAlbum `json:"album"`
}

// PlaylistContextLoader is the narrow seam for followed/non-owned contexts.
// The fork integration can provide it without leaking protocol types to the TUI.
type PlaylistContextLoader func(context.Context, string, PageRequest) (TrackPage, error)

func (a *Adapter) SetPlaylistContextLoader(loader PlaylistContextLoader) {
	a.playlistContextLoader = loader
}

func (a *Adapter) nativeCatalogGet(ctx context.Context, data daemon.ApiRequestDataNativeCatalog, out any) error {
	response, err := a.server.request(ctx, daemon.ApiRequestTypeNativeCatalog, data)
	if err != nil {
		return err
	}
	nativeResponse, ok := response.(daemon.ApiResponseNativeCatalog)
	if !ok {
		return fmt.Errorf("decode native catalog response: unexpected %T", response)
	}
	if err := json.Unmarshal(nativeResponse.Payload, out); err != nil {
		return fmt.Errorf("decode native catalog response: %w", err)
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func imageURL(images []spotifyImage) string {
	if len(images) == 0 {
		return ""
	}
	return images[0].URL
}
func externalURL(values map[string]string) string { return values["spotify"] }
func joinArtists(values []spotifyArtist) string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		names = append(names, value.Name)
	}
	return strings.Join(names, ", ")
}
func artistName(value spotifyArtist) string { return value.Name }

func translateTrack(value spotifyTrack) Track {
	uri := value.URI
	if uri == "" && value.ID != "" {
		uri = "spotify:track:" + value.ID
	}
	return Track{URI: uri, Name: value.Name, Artist: joinArtists(value.Artists), Album: value.Album.Name, AlbumURI: value.Album.URI, DurationMS: value.DurationMS, Explicit: value.Explicit, ImageURL: imageURL(value.Album.Images), ExternalURL: externalURL(value.ExternalURLs), TrackNumber: value.TrackNumber, DiscNumber: value.DiscNumber}
}

func translatePlaylist(value spotifyPlaylist) PlaylistSummary {
	uri := value.URI
	if uri == "" && value.ID != "" {
		uri = "spotify:playlist:" + value.ID
	}
	return PlaylistSummary{URI: uri, Name: value.Name, Owner: artistName(value.Owner), Description: value.Description, TrackCount: value.Tracks.Total, ImageURL: imageURL(value.Images), ExternalURL: externalURL(value.ExternalURLs)}
}

func translateAlbum(value spotifyAlbum) AlbumSummary {
	uri := value.URI
	if uri == "" && value.ID != "" {
		uri = "spotify:album:" + value.ID
	}
	return AlbumSummary{URI: uri, Name: value.Name, Artist: joinArtists(value.Artists), ReleaseDate: value.ReleaseDate, TrackCount: value.TotalTracks, ImageURL: imageURL(value.Images), ExternalURL: externalURL(value.ExternalURLs)}
}

func translateArtist(value spotifyArtist) ArtistSummary {
	uri := value.URI
	if uri == "" && value.ID != "" {
		uri = "spotify:artist:" + value.ID
	}
	return ArtistSummary{URI: uri, Name: value.Name, ImageURL: imageURL(value.Images), ExternalURL: externalURL(value.ExternalURLs)}
}

func (a *Adapter) LikedTracks(ctx context.Context, request PageRequest) (TrackPage, error) {
	var response spotifyPage
	if err := a.nativeCatalogGet(ctx, daemon.ApiRequestDataNativeCatalog{Kind: "liked", Offset: max(0, request.Offset), Limit: request.Limit}, &response); err != nil {
		return TrackPage{}, err
	}
	page := TrackPage{Total: response.Total, Offset: response.Offset, Limit: response.Limit}
	for _, raw := range response.Items {
		var item spotifySavedTrack
		if err := json.Unmarshal(raw, &item); err != nil {
			return TrackPage{}, err
		}
		page.Items = append(page.Items, translateTrack(item.Track))
	}
	return page, nil
}

func (a *Adapter) UserPlaylists(ctx context.Context, request PageRequest) (CatalogPage[PlaylistSummary], error) {
	var response spotifyPage
	if err := a.nativeCatalogGet(ctx, daemon.ApiRequestDataNativeCatalog{Kind: "playlists", Offset: max(0, request.Offset), Limit: request.Limit}, &response); err != nil {
		return CatalogPage[PlaylistSummary]{}, err
	}
	page := CatalogPage[PlaylistSummary]{Total: response.Total, Offset: response.Offset, Limit: response.Limit}
	for _, raw := range response.Items {
		var item spotifyPlaylist
		if err := json.Unmarshal(raw, &item); err != nil {
			return page, err
		}
		page.Items = append(page.Items, translatePlaylist(item))
	}
	return page, nil
}

func (a *Adapter) SavedAlbums(ctx context.Context, request PageRequest) (CatalogPage[AlbumSummary], error) {
	var response spotifyPage
	if err := a.nativeCatalogGet(ctx, daemon.ApiRequestDataNativeCatalog{Kind: "saved_albums", Offset: max(0, request.Offset), Limit: request.Limit}, &response); err != nil {
		return CatalogPage[AlbumSummary]{}, err
	}
	page := CatalogPage[AlbumSummary]{Total: response.Total, Offset: response.Offset, Limit: response.Limit}
	for _, raw := range response.Items {
		var item spotifySavedAlbum
		if err := json.Unmarshal(raw, &item); err != nil {
			return page, err
		}
		page.Items = append(page.Items, translateAlbum(item.Album))
	}
	return page, nil
}

func (a *Adapter) Search(ctx context.Context, request SearchRequest) (SearchGroups, error) {
	trackPage, err := a.SearchTracks(ctx, request)
	if err != nil {
		return SearchGroups{}, err
	}
	// Native search currently exposes track results. Returning the explicit
	// unavailable marker keeps the other groups honest and, importantly, never
	// falls back to the quota-bound public Web API.
	return SearchGroups{
		Tracks:              TrackPage{Items: trackPage.Tracks, Total: trackPage.Total, Offset: trackPage.Offset, Limit: request.Limit},
		NonTrackUnavailable: true,
	}, nil
}

func translateSearchTracks(response spotifyPage) TrackPage {
	page := TrackPage{Total: response.Total, Offset: response.Offset, Limit: response.Limit}
	for _, raw := range response.Items {
		var value spotifyTrack
		if json.Unmarshal(raw, &value) == nil {
			page.Items = append(page.Items, translateTrack(value))
		}
	}
	return page
}
func translateAlbumPage(response spotifyPage) CatalogPage[AlbumSummary] {
	page := CatalogPage[AlbumSummary]{Total: response.Total, Offset: response.Offset, Limit: response.Limit}
	for _, raw := range response.Items {
		var value spotifyAlbum
		if json.Unmarshal(raw, &value) == nil {
			page.Items = append(page.Items, translateAlbum(value))
		}
	}
	return page
}
func translateArtistPage(response spotifyPage) CatalogPage[ArtistSummary] {
	page := CatalogPage[ArtistSummary]{Total: response.Total, Offset: response.Offset, Limit: response.Limit}
	for _, raw := range response.Items {
		var value spotifyArtist
		if json.Unmarshal(raw, &value) == nil {
			page.Items = append(page.Items, translateArtist(value))
		}
	}
	return page
}
func translatePlaylistPage(response spotifyPage) CatalogPage[PlaylistSummary] {
	page := CatalogPage[PlaylistSummary]{Total: response.Total, Offset: response.Offset, Limit: response.Limit}
	for _, raw := range response.Items {
		var value spotifyPlaylist
		if json.Unmarshal(raw, &value) == nil {
			page.Items = append(page.Items, translatePlaylist(value))
		}
	}
	return page
}

func entityID(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) >= 3 {
		return parts[len(parts)-1]
	}
	return uri
}

func (a *Adapter) Playlist(ctx context.Context, uri string, request PageRequest) (PlaylistDetail, error) {
	var response struct {
		Playlist spotifyPlaylist `json:"playlist"`
		Tracks   spotifyPage     `json:"tracks"`
	}
	if err := a.nativeCatalogGet(ctx, daemon.ApiRequestDataNativeCatalog{Kind: "playlist", URI: uri, Offset: max(0, request.Offset), Limit: request.Limit}, &response); err != nil {
		return PlaylistDetail{}, err
	}
	detail := PlaylistDetail{PlaylistSummary: translatePlaylist(response.Playlist)}
	detail.Tracks = TrackPage{Total: response.Tracks.Total, Offset: response.Tracks.Offset, Limit: response.Tracks.Limit}
	for _, raw := range response.Tracks.Items {
		var value spotifyTrack
		if err := json.Unmarshal(raw, &value); err == nil {
			detail.Tracks.Items = append(detail.Tracks.Items, translateTrack(value))
		}
	}
	if a.playlistContextLoader != nil {
		if tracks, err := a.playlistContextLoader(ctx, uri, request); err == nil {
			detail.Tracks = tracks
		}
	}
	return detail, nil
}

func (a *Adapter) Album(ctx context.Context, uri string, request PageRequest) (AlbumDetail, error) {
	var response struct {
		Album  spotifyAlbum `json:"album"`
		Tracks spotifyPage  `json:"tracks"`
	}
	if err := a.nativeCatalogGet(ctx, daemon.ApiRequestDataNativeCatalog{Kind: "album", URI: uri, Offset: max(0, request.Offset), Limit: request.Limit}, &response); err != nil {
		return AlbumDetail{}, err
	}
	detail := AlbumDetail{AlbumSummary: translateAlbum(response.Album)}
	detail.Tracks = translateSearchTracks(response.Tracks)
	return detail, nil
}

func (a *Adapter) Artist(ctx context.Context, uri string, request PageRequest) (ArtistDetail, error) {
	var response struct {
		Artist  spotifyArtist `json:"artist"`
		Genres  []string      `json:"genres"`
		Popular spotifyPage   `json:"popular"`
		Albums  spotifyPage   `json:"albums"`
	}
	if err := a.nativeCatalogGet(ctx, daemon.ApiRequestDataNativeCatalog{Kind: "artist", URI: uri, Offset: max(0, request.Offset), Limit: request.Limit}, &response); err != nil {
		return ArtistDetail{}, err
	}
	return ArtistDetail{ArtistSummary: translateArtist(response.Artist), Genres: response.Genres, Popular: translateSearchTracks(response.Popular), Albums: translateAlbumPage(response.Albums).Items}, nil
}

func (a *Adapter) TopArtists(ctx context.Context, request PageRequest) (CatalogPage[ArtistSummary], error) {
	var response spotifyPage
	if err := a.nativeCatalogGet(ctx, daemon.ApiRequestDataNativeCatalog{Kind: "top_artists", Offset: max(0, request.Offset), Limit: request.Limit}, &response); err != nil {
		return CatalogPage[ArtistSummary]{}, err
	}
	return translateArtistPage(response), nil
}

func (a *Adapter) TopTracks(ctx context.Context, request PageRequest) (TrackPage, error) {
	var response spotifyPage
	if err := a.nativeCatalogGet(ctx, daemon.ApiRequestDataNativeCatalog{Kind: "top_tracks", Offset: max(0, request.Offset), Limit: request.Limit}, &response); err != nil {
		return TrackPage{}, err
	}
	return translateSearchTracks(response), nil
}

func (a *Adapter) Recommended(ctx context.Context, request PageRequest) (RecommendedPage, error) {
	artists, err := a.TopArtists(ctx, request)
	if err != nil {
		return RecommendedPage{}, err
	}
	tracks, err := a.TopTracks(ctx, request)
	if err != nil {
		return RecommendedPage{Artists: artists, TracksUnavailable: true}, nil
	}
	page := RecommendedPage{Artists: artists, Tracks: tracks}
	seenAlbums := make(map[string]bool)
	for _, track := range tracks.Items {
		if track.AlbumURI != "" && !seenAlbums[track.AlbumURI] {
			seenAlbums[track.AlbumURI] = true
			page.Albums = append(page.Albums, AlbumSummary{URI: track.AlbumURI, Name: track.Album, Artist: track.Artist, ImageURL: track.ImageURL})
		}
	}
	return page, nil
}

func (a *Adapter) PlayContext(ctx context.Context, contextURI, trackURI string, positionMS int) error {
	_, err := a.server.request(ctx, daemon.ApiRequestTypePlay, daemon.ApiRequestDataPlay{Uri: contextURI, SkipToUri: trackURI, Position: int64(max(0, positionMS))})
	return err
}

func (a *Adapter) SetShuffle(ctx context.Context, enabled bool) error {
	return a.command(ctx, daemon.ApiRequestTypeSetShufflingContext, enabled)
}

func (a *Adapter) SeekRelative(ctx context.Context, deltaMS int) error {
	_, err := a.server.request(ctx, daemon.ApiRequestTypeSeek, daemon.ApiRequestDataSeek{Position: int64(deltaMS), Relative: true})
	return err
}
