package spotifyapi

import (
	"strings"

	"github.com/dcbto/spotui/internal/library"
	"github.com/dcbto/spotui/internal/player"
	"github.com/zmb3/spotify/v2"
)

func TranslatePlaylist(pl spotify.SimplePlaylist) library.PlaylistEntry {
	return library.PlaylistEntry{
		Name:       pl.Name,
		TrackCount: int(pl.Tracks.Total),
		URI:        string(pl.URI),
	}
}

func TranslateTrack(t spotify.SavedTrack) library.TrackEntry {
	artistNames := make([]string, len(t.Artists))
	for i, a := range t.Artists {
		artistNames[i] = a.Name
	}
	return library.TrackEntry{
		Name:     t.Name,
		Artist:   strings.Join(artistNames, ", "),
		Duration: int(t.Duration),
		URI:      string(t.URI),
	}
}

func TranslateArtist(a spotify.FullArtist) library.ArtistEntry {
	return library.ArtistEntry{
		Name:   a.Name,
		Genres: a.Genres,
	}
}

func TranslateNowPlayingEntry(s spotify.PlayerState) player.NowPlayingEntry {
	var trackName string
	if s.Item != nil {
		trackName = s.Item.Name
	}
	return player.NowPlayingEntry{
		ProgressMs: int(s.Progress),
		Playing:    s.Playing,
		ShuffleOn:  s.ShuffleState,
		TrackName:  trackName,
	}
}
