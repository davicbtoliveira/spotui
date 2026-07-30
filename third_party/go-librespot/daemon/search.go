package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	librespot "github.com/devgianlu/go-librespot"
	connectpb "github.com/devgianlu/go-librespot/proto/spotify/connectstate"
	extmetadatapb "github.com/devgianlu/go-librespot/proto/spotify/extendedmetadata"
	metadatapb "github.com/devgianlu/go-librespot/proto/spotify/metadata"
	"github.com/devgianlu/go-librespot/spclient"
)

func requestSearch(
	ctx context.Context,
	log librespot.Logger,
	client *spclient.Spclient,
	request ApiRequestDataSearch,
) (any, error) {
	query := strings.TrimSpace(request.Query)
	if query == "" {
		return nil, errors.New("search query is empty")
	}
	if request.Offset < 0 {
		return nil, errors.New("search offset must not be negative")
	}
	if request.Limit <= 0 {
		return nil, errors.New("search limit must be positive")
	}

	searchContext, err := client.ContextResolve(ctx, searchContextURI(query))
	if err != nil {
		return nil, fmt.Errorf("resolve search context: %w", err)
	}
	resolver, err := spclient.NewContextResolver(ctx, log, client, searchContext)
	if err != nil {
		return nil, fmt.Errorf("initialize search context: %w", err)
	}

	needed := request.Offset + request.Limit + 1
	var contextTracks []*connectpb.ContextTrack
	for pageIndex := 0; len(contextTracks) < needed; pageIndex++ {
		page, err := resolver.Page(ctx, pageIndex)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("load search page: %w", err)
		}
		contextTracks = append(contextTracks, page...)
	}

	start := min(request.Offset, len(contextTracks))
	end := min(start+request.Limit, len(contextTracks))
	response := ApiResponseSearch{
		Tracks: make([]ApiResponseSearchTrack, 0, end-start),
		Total:  len(contextTracks),
		Offset: request.Offset,
	}
	for _, contextTrack := range contextTracks[start:end] {
		uri := contextTrack.Uri
		if uri == "" && len(contextTrack.Gid) > 0 {
			uri = librespot.SpotifyIdFromGid(librespot.SpotifyIdTypeTrack, contextTrack.Gid).Uri()
		}
		trackID, err := librespot.SpotifyIdFromUri(uri)
		if err != nil {
			log.WithError(err).Warnf("skipping invalid search track")
			continue
		}

		var metadata metadatapb.Track
		if err := client.ExtendedMetadataSimple(
			ctx,
			*trackID,
			extmetadatapb.ExtensionKind_TRACK_V4,
			&metadata,
		); err != nil {
			log.WithError(err).Warnf("skipping search track metadata")
			continue
		}
		response.Tracks = append(response.Tracks, searchTrackResponse(uri, &metadata))
	}
	return response, nil
}

func searchContextURI(query string) string {
	return "spotify:search:" + url.QueryEscape(query)
}

func searchTrackResponse(uri string, metadata *metadatapb.Track) ApiResponseSearchTrack {
	artists := make([]string, 0, len(metadata.Artist))
	for _, artist := range metadata.Artist {
		if artist.Name != nil {
			artists = append(artists, *artist.Name)
		}
	}
	var album string
	if metadata.Album != nil && metadata.Album.Name != nil {
		album = *metadata.Album.Name
	}
	return ApiResponseSearchTrack{
		Uri:         uri,
		Name:        metadata.GetName(),
		ArtistNames: artists,
		AlbumName:   album,
		Duration:    int(metadata.GetDuration()),
	}
}
