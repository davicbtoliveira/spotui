package spotengine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
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
type spotifySearchResponse struct {
	Tracks    spotifyPage `json:"tracks"`
	Albums    spotifyPage `json:"albums"`
	Artists   spotifyPage `json:"artists"`
	Playlists spotifyPage `json:"playlists"`
}

// PlaylistContextLoader is the narrow seam for followed/non-owned contexts.
// The fork integration can provide it without leaking protocol types to the TUI.
type PlaylistContextLoader func(context.Context, string, PageRequest) (TrackPage, error)

func (a *Adapter) SetPlaylistContextLoader(loader PlaylistContextLoader) {
	a.playlistContextLoader = loader
}

func (a *Adapter) webAPIGet(ctx context.Context, path string, query url.Values, out any) error {
	response, err := a.server.request(ctx, daemon.ApiRequestTypeWebApi, daemon.ApiRequestDataWebApi{
		Method: "GET", Path: path, Query: query,
	})
	if err != nil {
		return err
	}
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode web api response: %w", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode web api response: %w", err)
	}
	return nil
}

func pageQuery(request PageRequest) url.Values {
	limit := request.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	return url.Values{"limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(max(0, request.Offset))}}
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
	if err := a.webAPIGet(ctx, "/v1/me/tracks", pageQuery(request), &response); err != nil {
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
	if err := a.webAPIGet(ctx, "/v1/me/playlists", pageQuery(request), &response); err != nil {
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
	if err := a.webAPIGet(ctx, "/v1/me/albums", pageQuery(request), &response); err != nil {
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
	limit := request.Limit
	if limit <= 0 || limit > 10 {
		limit = 10
	}
	query := url.Values{"q": {request.Query}, "type": {"album,artist,playlist"}, "limit": {strconv.Itoa(limit)}, "offset": {strconv.Itoa(max(0, request.Offset))}}
	var response spotifySearchResponse
	if err := a.webAPIGet(ctx, "/v1/search", query, &response); err != nil {
		return SearchGroups{Tracks: TrackPage{Items: trackPage.Tracks, Total: trackPage.Total, Offset: trackPage.Offset, Limit: request.Limit}, NonTrackUnavailable: true}, nil
	}
	groups := SearchGroups{Tracks: TrackPage{Items: trackPage.Tracks, Total: trackPage.Total, Offset: trackPage.Offset, Limit: request.Limit}}
	groups.Albums = translateAlbumPage(response.Albums)
	groups.Artists = translateArtistPage(response.Artists)
	groups.Playlists = translatePlaylistPage(response.Playlists)
	return groups, nil
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
	var value spotifyPlaylist
	if err := a.webAPIGet(ctx, "/v1/playlists/"+url.PathEscape(entityID(uri)), nil, &value); err != nil {
		return PlaylistDetail{}, err
	}
	detail := PlaylistDetail{PlaylistSummary: translatePlaylist(value)}
	if a.playlistContextLoader != nil {
		tracks, err := a.playlistContextLoader(ctx, uri, request)
		if err == nil {
			detail.Tracks = tracks
			return detail, nil
		}
	}
	if request.Limit == 0 {
		request.Limit = 10
	}
	query := pageQuery(request)
	var items spotifyPage
	if err := a.webAPIGet(ctx, "/v1/playlists/"+url.PathEscape(entityID(uri))+"/tracks", query, &items); err != nil {
		return detail, err
	}
	detail.Tracks = TrackPage{Total: items.Total, Offset: items.Offset, Limit: items.Limit}
	for _, raw := range items.Items {
		var item struct {
			Track spotifyTrack `json:"track"`
		}
		if err := json.Unmarshal(raw, &item); err == nil {
			detail.Tracks.Items = append(detail.Tracks.Items, translateTrack(item.Track))
		}
	}
	return detail, nil
}

func (a *Adapter) Album(ctx context.Context, uri string, request PageRequest) (AlbumDetail, error) {
	var value spotifyAlbum
	if err := a.webAPIGet(ctx, "/v1/albums/"+url.PathEscape(entityID(uri)), nil, &value); err != nil {
		return AlbumDetail{}, err
	}
	detail := AlbumDetail{AlbumSummary: translateAlbum(value)}
	var items spotifyPage
	if err := a.webAPIGet(ctx, "/v1/albums/"+url.PathEscape(entityID(uri))+"/tracks", pageQuery(request), &items); err != nil {
		return detail, err
	}
	detail.Tracks = translateSearchTracks(items)
	return detail, nil
}

func (a *Adapter) Artist(ctx context.Context, uri string, request PageRequest) (ArtistDetail, error) {
	var value spotifyArtist
	if err := a.webAPIGet(ctx, "/v1/artists/"+url.PathEscape(entityID(uri)), nil, &value); err != nil {
		return ArtistDetail{}, err
	}
	detail := ArtistDetail{ArtistSummary: translateArtist(value), Genres: value.Genres}
	var popular spotifyPage
	if err := a.webAPIGet(ctx, "/v1/artists/"+url.PathEscape(entityID(uri))+"/top-tracks", url.Values{"market": {"from_token"}}, &popular); err == nil {
		detail.Popular = translateSearchTracks(popular)
	}
	var albums spotifyPage
	if err := a.webAPIGet(ctx, "/v1/artists/"+url.PathEscape(entityID(uri))+"/albums", pageQuery(request), &albums); err == nil {
		detail.Albums = translateAlbumPage(albums).Items
	}
	return detail, nil
}

func (a *Adapter) TopArtists(ctx context.Context, request PageRequest) (CatalogPage[ArtistSummary], error) {
	var response spotifyPage
	if err := a.webAPIGet(ctx, "/v1/me/top/artists", pageQuery(request), &response); err != nil {
		return CatalogPage[ArtistSummary]{}, err
	}
	return translateArtistPage(response), nil
}

func (a *Adapter) TopTracks(ctx context.Context, request PageRequest) (TrackPage, error) {
	var response spotifyPage
	if err := a.webAPIGet(ctx, "/v1/me/top/tracks", pageQuery(request), &response); err != nil {
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
	seenPlaylists := make(map[string]bool)
	for _, artist := range artists.Items {
		detail, err := a.Artist(ctx, artist.URI, PageRequest{Limit: 5})
		if err == nil {
			for _, album := range detail.Albums {
				if !seenAlbums[album.URI] {
					seenAlbums[album.URI] = true
					page.Albums = append(page.Albums, album)
				}
			}
		}
		groups, err := a.Search(ctx, SearchRequest{Query: artist.Name, Limit: 5})
		if err == nil {
			for _, playlist := range groups.Playlists.Items {
				if !seenPlaylists[playlist.URI] {
					seenPlaylists[playlist.URI] = true
					page.Playlists = append(page.Playlists, playlist)
				}
			}
		}
	}
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
