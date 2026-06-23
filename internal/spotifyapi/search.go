package spotifyapi

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

type TrackSearchRequest struct {
	Query  string
	Limit  int
	Market string
}

type TrackSearchPage struct {
	Tracks []spotify.FullTrack
	Total  int
}

type TrackSearcher interface {
	SearchTracks(context.Context, TrackSearchRequest) (TrackSearchPage, error)
}

type SpotifyTrackSearcher struct {
	Client *spotify.Client
}

func (s SpotifyTrackSearcher) SearchTracks(ctx context.Context, req TrackSearchRequest) (TrackSearchPage, error) {
	result, err := s.Client.Search(
		ctx,
		req.Query,
		spotify.SearchTypeTrack,
		spotify.Limit(req.Limit),
		spotify.Market(req.Market),
	)
	if err != nil {
		return TrackSearchPage{}, err
	}
	if result.Tracks == nil {
		return TrackSearchPage{}, nil
	}
	return TrackSearchPage{
		Tracks: result.Tracks.Tracks,
		Total:  int(result.Tracks.Total),
	}, nil
}
