package daemon

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	librespot "github.com/devgianlu/go-librespot"
	connectpb "github.com/devgianlu/go-librespot/proto/spotify/connectstate"
	extmetadatapb "github.com/devgianlu/go-librespot/proto/spotify/extendedmetadata"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"github.com/devgianlu/go-librespot/session"
	"github.com/devgianlu/go-librespot/spclient"
)

// requestNativeCatalog keeps catalog reads on the authenticated native
// spclient. The public Web API has an application quota shared by unofficial
// clients; playback and native metadata use a different protocol and are not
// routed through that quota.
func requestNativeCatalog(ctx context.Context, log librespot.Logger, sess *session.Session, request ApiRequestDataNativeCatalog) (any, error) {
	request.Limit = catalogLimit(request.Limit)
	if request.Offset < 0 {
		return nil, errors.New("catalog offset must not be negative")
	}

	client := sess.Spclient()
	var payload any
	var err error
	switch request.Kind {
	case "liked":
		payload, err = nativeContextPage(ctx, log, client, "spotify:user:"+sess.Username()+":collection", request, true)
	case "playlist":
		payload, err = nativePlaylist(ctx, log, client, request)
	case "album":
		payload, err = nativeAlbum(ctx, client, request)
	case "artist":
		payload, err = nativeArtist(ctx, client, request)
	case "playlists":
		payload, err = nativeUserPlaylists(ctx, client, sess.Username(), request)
	case "saved_albums":
		payload, err = nativeSavedAlbums(ctx, log, client, sess.Username(), request)
	case "top_tracks":
		// The native protocol does not expose the Web API's affinity windows.
		// The user's native collection is a stable, authenticated fallback that
		// keeps this tab useful without making a quota-bound request.
		payload, err = nativeContextPage(ctx, log, client, "spotify:user:"+sess.Username()+":collection", request, false)
	case "top_artists":
		payload, err = nativeTopArtists(ctx, log, client, sess.Username(), request)
	case "search":
		payload, err = nativeSearch(ctx, log, client, request)
	default:
		return nil, fmt.Errorf("unknown native catalog kind %q", request.Kind)
	}
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode native catalog response: %w", err)
	}
	return ApiResponseNativeCatalog{Payload: encoded}, nil
}

func catalogLimit(limit int) int {
	if limit <= 0 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}

type nativePage struct {
	Items  []any `json:"items"`
	Total  int   `json:"total"`
	Offset int   `json:"offset"`
	Limit  int   `json:"limit"`
}

func (p nativePage) value() map[string]any {
	return map[string]any{
		"items": p.Items, "total": p.Total, "offset": p.Offset, "limit": p.Limit,
	}
}

func nativeContextPage(ctx context.Context, log librespot.Logger, client *spclient.Spclient, uri string, request ApiRequestDataNativeCatalog, saved bool) (any, error) {
	tracks, total, err := nativeContextTracks(ctx, log, client, uri, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}
	items := make([]any, 0, len(tracks))
	for _, track := range tracks {
		item, err := nativeTrack(ctx, client, track)
		if err != nil {
			log.WithError(err).Debugf("skipping native catalog track %s", track.Uri)
			continue
		}
		if saved {
			items = append(items, map[string]any{"track": item})
		} else {
			items = append(items, item)
		}
	}
	return nativePage{Items: items, Total: total, Offset: request.Offset, Limit: request.Limit}.value(), nil
}

func nativeContextTracks(ctx context.Context, log librespot.Logger, client *spclient.Spclient, uri string, offset, limit int) ([]*connectpb.ContextTrack, int, error) {
	spotCtx, err := client.ContextResolve(ctx, uri)
	if err != nil {
		return nil, 0, fmt.Errorf("resolve native context %s: %w", uri, err)
	}
	resolver, err := spclient.NewContextResolver(ctx, log, client, spotCtx)
	if err != nil {
		return nil, 0, fmt.Errorf("initialize native context %s: %w", uri, err)
	}

	needed := offset + limit
	if needed < limit {
		needed = limit
	}
	all := make([]*connectpb.ContextTrack, 0, needed)
	for pageIndex := 0; len(all) < needed; pageIndex++ {
		page, pageErr := resolver.Page(ctx, pageIndex)
		if errors.Is(pageErr, io.EOF) {
			break
		}
		if pageErr != nil {
			return nil, 0, fmt.Errorf("load native context page %s: %w", uri, pageErr)
		}
		all = append(all, page...)
		if len(page) == 0 {
			break
		}
	}
	start := minInt(offset, len(all))
	end := minInt(start+limit, len(all))
	total := contextTotal(resolver.Metadata(), len(all))
	return all[start:end], total, nil
}

func contextTotal(metadata map[string]string, loaded int) int {
	for _, key := range []string{"playlist_number_of_tracks", "number_of_tracks", "total_tracks"} {
		if value, err := strconv.Atoi(metadata[key]); err == nil && value >= loaded {
			return value
		}
	}
	return loaded
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func nativeTrack(ctx context.Context, client *spclient.Spclient, contextTrack *connectpb.ContextTrack) (map[string]any, error) {
	uri := contextTrack.Uri
	if uri == "" && len(contextTrack.Gid) == 16 {
		uri = librespot.SpotifyIdFromGid(librespot.SpotifyIdTypeTrack, contextTrack.Gid).Uri()
	}
	if uri == "" {
		return nil, errors.New("native context track has no uri")
	}
	id, err := librespot.SpotifyIdFromUri(uri)
	if err != nil {
		return nil, err
	}
	var metadata metadatapb.Track
	if err := client.ExtendedMetadataSimple(ctx, *id, extmetadatapb.ExtensionKind_TRACK_V4, &metadata); err != nil {
		return nil, err
	}
	return trackValue(&metadata, uri, ""), nil
}

func trackValue(track *metadatapb.Track, uri, parentAlbumURI string) map[string]any {
	if uri == "" && len(track.Gid) == 16 {
		uri = librespot.SpotifyIdFromGid(librespot.SpotifyIdTypeTrack, track.Gid).Uri()
	}
	albumURI := parentAlbumURI
	if track.Album != nil && len(track.Album.Gid) == 16 {
		albumURI = librespot.SpotifyIdFromGid(librespot.SpotifyIdType("album"), track.Album.Gid).Uri()
	}
	album := albumValue(track.Album, albumURI)
	artists := make([]any, 0, len(track.Artist))
	for _, artist := range track.Artist {
		artists = append(artists, artistValue(artist))
	}
	return map[string]any{
		"id":            entityID(uri),
		"uri":           uri,
		"name":          track.GetName(),
		"duration_ms":   track.GetDuration(),
		"explicit":      track.GetExplicit(),
		"track_number":  track.GetNumber(),
		"disc_number":   track.GetDiscNumber(),
		"artists":       artists,
		"album":         album,
		"external_urls": map[string]any{"spotify": externalURL(uri)},
	}
}

func albumValue(album *metadatapb.Album, fallbackURI string) map[string]any {
	if album == nil {
		return map[string]any{"uri": fallbackURI, "name": ""}
	}
	uri := albumURI(album, fallbackURI)
	artists := make([]any, 0, len(album.Artist))
	for _, artist := range album.Artist {
		artists = append(artists, artistValue(artist))
	}
	tracks := 0
	for _, disc := range album.Disc {
		tracks += len(disc.Track)
	}
	return map[string]any{
		"id":            entityID(uri),
		"uri":           uri,
		"name":          album.GetName(),
		"artists":       artists,
		"release_date":  albumDate(album.Date),
		"total_tracks":  tracks,
		"images":        imageValues(album.Cover, album.CoverGroup),
		"external_urls": map[string]any{"spotify": externalURL(uri)},
	}
}

func albumURI(album *metadatapb.Album, fallbackURI string) string {
	if album != nil && len(album.Gid) == 16 {
		return librespot.SpotifyIdFromGid(librespot.SpotifyIdType("album"), album.Gid).Uri()
	}
	return fallbackURI
}

func artistValue(artist *metadatapb.Artist) map[string]any {
	if artist == nil {
		return map[string]any{"name": ""}
	}
	uri := ""
	if len(artist.Gid) == 16 {
		uri = librespot.SpotifyIdFromGid(librespot.SpotifyIdType("artist"), artist.Gid).Uri()
	}
	return map[string]any{
		"id":            entityID(uri),
		"uri":           uri,
		"name":          artist.GetName(),
		"images":        imageValues(artist.Portrait, artist.PortraitGroup),
		"external_urls": map[string]any{"spotify": externalURL(uri)},
	}
}

func imageValues(images []*metadatapb.Image, group *metadatapb.ImageGroup) []any {
	if len(images) == 0 && group != nil {
		images = group.Image
	}
	values := make([]any, 0, len(images))
	for _, image := range images {
		if image == nil || len(image.FileId) == 0 {
			continue
		}
		values = append(values, map[string]any{
			"url":    "https://i.scdn.co/image/" + hex.EncodeToString(image.FileId),
			"width":  image.GetWidth(),
			"height": image.GetHeight(),
		})
	}
	return values
}

func albumDate(date *metadatapb.Date) string {
	if date == nil || date.GetYear() == 0 {
		return ""
	}
	if date.GetMonth() == 0 {
		return strconv.Itoa(int(date.GetYear()))
	}
	if date.GetDay() == 0 {
		return fmt.Sprintf("%04d-%02d", date.GetYear(), date.GetMonth())
	}
	return fmt.Sprintf("%04d-%02d-%02d", date.GetYear(), date.GetMonth(), date.GetDay())
}

func entityID(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) == 3 {
		return parts[2]
	}
	return ""
}

func externalURL(uri string) string {
	parts := strings.Split(uri, ":")
	if len(parts) != 3 || parts[1] == "" || parts[2] == "" {
		return ""
	}
	return "https://open.spotify.com/" + parts[1] + "/" + parts[2]
}

func nativeAlbum(ctx context.Context, client *spclient.Spclient, request ApiRequestDataNativeCatalog) (any, error) {
	id, err := librespot.SpotifyIdFromUri(request.URI)
	if err != nil {
		return nil, err
	}
	var album metadatapb.Album
	if err := client.ExtendedMetadataSimple(ctx, *id, extmetadatapb.ExtensionKind_ALBUM_V4, &album); err != nil {
		return nil, fmt.Errorf("load album metadata: %w", err)
	}
	tracks := make([]any, 0)
	for _, disc := range album.Disc {
		for _, track := range disc.Track {
			tracks = append(tracks, trackValue(track, "", request.URI))
		}
	}
	start := minInt(request.Offset, len(tracks))
	end := minInt(start+request.Limit, len(tracks))
	return map[string]any{
		"album":  albumValue(&album, request.URI),
		"tracks": nativePage{Items: tracks[start:end], Total: len(tracks), Offset: request.Offset, Limit: request.Limit}.value(),
	}, nil
}

func nativeArtist(ctx context.Context, client *spclient.Spclient, request ApiRequestDataNativeCatalog) (any, error) {
	id, err := librespot.SpotifyIdFromUri(request.URI)
	if err != nil {
		return nil, err
	}
	var artist metadatapb.Artist
	if err := client.ExtendedMetadataSimple(ctx, *id, extmetadatapb.ExtensionKind_ARTIST_V4, &artist); err != nil {
		return nil, fmt.Errorf("load artist metadata: %w", err)
	}
	popular := make([]any, 0)
	for _, top := range artist.TopTrack {
		for _, track := range top.Track {
			popular = append(popular, trackValue(track, "", ""))
		}
	}
	albumURIs := make([]string, 0)
	seen := map[string]bool{}
	for _, group := range append(append(append(artist.AlbumGroup, artist.SingleGroup...), artist.CompilationGroup...), artist.AppearsOnGroup...) {
		for _, album := range group.Album {
			uri := albumURI(album, "")
			if uri == "" || seen[uri] {
				continue
			}
			seen[uri] = true
			albumURIs = append(albumURIs, uri)
		}
	}
	start := minInt(request.Offset, len(albumURIs))
	end := minInt(start+request.Limit, len(albumURIs))
	albums := make([]any, 0, end-start)
	for _, uri := range albumURIs[start:end] {
		value, ok := nativeAlbumSummary(ctx, client, uri)
		if ok {
			albums = append(albums, value)
		}
	}
	return map[string]any{
		"artist":  artistValue(&artist),
		"genres":  []string{},
		"popular": nativePage{Items: popular, Total: len(popular), Offset: 0, Limit: len(popular)}.value(),
		"albums":  nativePage{Items: albums, Total: len(albumURIs), Offset: request.Offset, Limit: request.Limit}.value(),
	}, nil
}

func nativeAlbumSummary(ctx context.Context, client *spclient.Spclient, uri string) (map[string]any, bool) {
	id, err := librespot.SpotifyIdFromUri(uri)
	if err != nil {
		return nil, false
	}
	var album metadatapb.Album
	if err := client.ExtendedMetadataSimple(ctx, *id, extmetadatapb.ExtensionKind_ALBUM_V4, &album); err != nil {
		return nil, false
	}
	value := albumValue(&album, uri)
	name, _ := value["name"].(string)
	return value, name != ""
}

func nativePlaylist(ctx context.Context, log librespot.Logger, client *spclient.Spclient, request ApiRequestDataNativeCatalog) (any, error) {
	tracks, total, err := nativeContextTracks(ctx, log, client, request.URI, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}
	items := make([]any, 0, len(tracks))
	for _, track := range tracks {
		value, trackErr := nativeTrack(ctx, client, track)
		if trackErr != nil {
			log.WithError(trackErr).Debugf("skipping playlist track %s", track.Uri)
			continue
		}
		items = append(items, value)
	}
	summary := map[string]any{
		"id":            entityID(request.URI),
		"uri":           request.URI,
		"name":          "Playlist",
		"description":   "",
		"images":        []any{},
		"external_urls": map[string]any{"spotify": externalURL(request.URI)},
		"tracks":        map[string]any{"total": total},
	}
	if spotCtx, resolveErr := client.ContextResolve(ctx, request.URI); resolveErr == nil {
		for key, destination := range map[string]string{"title": "name", "description": "description", "image_url": "image_url"} {
			if value := spotCtx.Metadata[key]; value != "" {
				if destination == "image_url" {
					summary["images"] = []any{map[string]any{"url": value}}
				} else {
					summary[destination] = value
				}
			}
		}
		if owner := spotCtx.Metadata["owner_name"]; owner != "" {
			summary["owner"] = map[string]any{"name": owner}
		}
	}
	return map[string]any{
		"playlist": summary,
		"tracks":   nativePage{Items: items, Total: total, Offset: request.Offset, Limit: request.Limit}.value(),
	}, nil
}

func nativeSearch(ctx context.Context, log librespot.Logger, client *spclient.Spclient, request ApiRequestDataNativeCatalog) (any, error) {
	response, err := requestSearch(ctx, log, client, ApiRequestDataSearch{Query: request.Query, Offset: request.Offset, Limit: request.Limit})
	if err != nil {
		return nil, err
	}
	items := make([]any, 0, len(response.(ApiResponseSearch).Tracks))
	for _, track := range response.(ApiResponseSearch).Tracks {
		items = append(items, map[string]any{
			"uri": track.Uri, "name": track.Name, "duration_ms": track.Duration,
			"artists": []any{map[string]any{"name": strings.Join(track.ArtistNames, ", ")}},
			"album":   map[string]any{"name": track.AlbumName},
		})
	}
	search := response.(ApiResponseSearch)
	return map[string]any{"tracks": nativePage{Items: items, Total: search.Total, Offset: search.Offset, Limit: request.Limit}.value()}, nil
}

func nativeTopArtists(ctx context.Context, log librespot.Logger, client *spclient.Spclient, username string, request ApiRequestDataNativeCatalog) (any, error) {
	tracks, total, err := nativeContextTracks(ctx, log, client, "spotify:user:"+username+":collection", 0, request.Offset+request.Limit)
	if err != nil {
		return nil, err
	}
	values := make([]any, 0)
	seen := map[string]bool{}
	for _, contextTrack := range tracks {
		item, trackErr := nativeTrack(ctx, client, contextTrack)
		if trackErr != nil {
			continue
		}
		artists, _ := item["artists"].([]any)
		for _, raw := range artists {
			artist, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			uri, _ := artist["uri"].(string)
			if uri != "" && seen[uri] {
				continue
			}
			seen[uri] = true
			values = append(values, artist)
		}
	}
	start := minInt(request.Offset, len(values))
	end := minInt(start+request.Limit, len(values))
	_ = total
	return nativePage{Items: values[start:end], Total: len(values), Offset: request.Offset, Limit: request.Limit}.value(), nil
}

func nativeUserPlaylists(ctx context.Context, client *spclient.Spclient, username string, request ApiRequestDataNativeCatalog) (any, error) {
	var profile any
	if err := nativeJSON(ctx, client, "GET", "/user-profile-view/v3/profile/"+url.PathEscape(username), url.Values{"playlist_limit": {strconv.Itoa(request.Offset + request.Limit)}}, &profile); err != nil {
		return nil, fmt.Errorf("load native user playlists: %w", err)
	}
	values := make([]any, 0)
	collectPlaylists(profile, &values)
	values = uniqueEntityValues(values, "uri")
	start := minInt(request.Offset, len(values))
	end := minInt(start+request.Limit, len(values))
	return nativePage{Items: values[start:end], Total: len(values), Offset: request.Offset, Limit: request.Limit}.value(), nil
}

func uniqueEntityValues(values []any, key string) []any {
	unique := make([]any, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		item, ok := value.(map[string]any)
		if !ok {
			continue
		}
		identity, _ := item[key].(string)
		if identity == "" || seen[identity] {
			continue
		}
		seen[identity] = true
		unique = append(unique, value)
	}
	return unique
}

func collectPlaylists(value any, result *[]any) {
	switch typed := value.(type) {
	case []any:
		for _, child := range typed {
			collectPlaylists(child, result)
		}
	case map[string]any:
		if candidate := playlistValue(typed); candidate != nil {
			*result = append(*result, candidate)
			return
		}
		for _, child := range typed {
			collectPlaylists(child, result)
		}
	}
}

func playlistValue(value map[string]any) map[string]any {
	typ, _ := value["type"].(string)
	uri, _ := value["uri"].(string)
	id, _ := value["id"].(string)
	if typ != "playlist" && !strings.HasPrefix(uri, "spotify:playlist:") && id == "" {
		return nil
	}
	if uri == "" && id != "" {
		uri = "spotify:playlist:" + id
	}
	if !strings.HasPrefix(uri, "spotify:playlist:") {
		return nil
	}
	name, _ := value["name"].(string)
	trackCount := intValue(value, "track_count")
	if tracks, ok := value["tracks"].(map[string]any); ok {
		trackCount = intValue(tracks, "total")
	}
	owner := value["owner"]
	switch typed := owner.(type) {
	case string:
		owner = map[string]any{"name": typed}
	case map[string]any:
		if stringValue(typed, "name") == "" && stringValue(typed, "username") != "" {
			typed["name"] = stringValue(typed, "username")
		}
	case nil:
		owner = map[string]any{"name": stringValue(value, "owner_name")}
	default:
		owner = map[string]any{"name": ""}
	}
	images := value["images"]
	if images == nil && value["image_url"] != nil {
		images = []any{map[string]any{"url": value["image_url"]}}
	} else {
		switch typed := images.(type) {
		case string:
			images = []any{map[string]any{"url": typed}}
		case map[string]any:
			images = []any{typed}
		}
	}
	return map[string]any{
		"id": id, "uri": uri, "name": name, "description": stringValue(value, "description"),
		"owner": owner, "images": images, "external_urls": map[string]any{"spotify": externalURL(uri)},
		"tracks": map[string]any{"total": trackCount},
	}
}

func stringValue(value map[string]any, key string) string {
	result, _ := value[key].(string)
	return result
}

func intValue(value map[string]any, key string) int {
	switch result := value[key].(type) {
	case int:
		return result
	case float64:
		return int(result)
	}
	return 0
}

func nativeSavedAlbums(ctx context.Context, log librespot.Logger, client *spclient.Spclient, username string, request ApiRequestDataNativeCatalog) (any, error) {
	candidates := []string{
		"spotify:user:" + username + ":collection:albums",
		"spotify:internal:collection:albums",
		"spotify:collection:albums",
	}
	var tracks []*connectpb.ContextTrack
	var err error
	for _, uri := range candidates {
		tracks, _, err = nativeContextTracks(ctx, log, client, uri, 0, request.Offset+request.Limit)
		if err == nil {
			break
		}
	}
	if err != nil {
		// Some accounts do not expose the album collection context. An empty
		// page is preferable to falling back to the quota-bound Web API.
		return nativePage{Items: []any{}, Total: 0, Offset: request.Offset, Limit: request.Limit}.value(), nil
	}
	values := make([]any, 0)
	seen := map[string]bool{}
	for _, contextTrack := range tracks {
		item, trackErr := nativeTrack(ctx, client, contextTrack)
		if trackErr != nil {
			continue
		}
		album, _ := item["album"].(map[string]any)
		if album == nil {
			continue
		}
		uri, _ := album["uri"].(string)
		if uri == "" || seen[uri] {
			continue
		}
		seen[uri] = true
		values = append(values, map[string]any{"album": album})
	}
	start := minInt(request.Offset, len(values))
	end := minInt(start+request.Limit, len(values))
	return nativePage{Items: values[start:end], Total: len(values), Offset: request.Offset, Limit: request.Limit}.value(), nil
}

func nativeJSON(ctx context.Context, client *spclient.Spclient, method, path string, query url.Values, output any) error {
	response, err := client.Request(ctx, method, path, query, nil, nil)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != 200 {
		return fmt.Errorf("native request returned status %d", response.StatusCode)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, output); err != nil {
		return fmt.Errorf("decode native response: %w", err)
	}
	return nil
}
