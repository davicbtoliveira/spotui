package spotifyapi_test

import (
	"testing"

	"github.com/dcbto/spotui/internal/spotifyapi"
	"github.com/zmb3/spotify/v2"
)

func TestTranslatePlaylist(t *testing.T) {
	got := spotifyapi.TranslatePlaylist(spotify.SimplePlaylist{
		Name:   "My Mix",
		Tracks: spotify.PlaylistTracks{Total: 42},
		URI:    "spotify:playlist:abc",
	})
	if got.Name != "My Mix" {
		t.Fatalf("Name: want %q, got %q", "My Mix", got.Name)
	}
	if got.TrackCount != 42 {
		t.Fatalf("TrackCount: want 42, got %d", got.TrackCount)
	}
	if got.URI != "spotify:playlist:abc" {
		t.Fatalf("URI: want %q, got %q", "spotify:playlist:abc", got.URI)
	}
}

func TestTranslateTrack(t *testing.T) {
	got := spotifyapi.TranslateTrack(spotify.SavedTrack{
		FullTrack: spotify.FullTrack{
			SimpleTrack: spotify.SimpleTrack{
				Name: "Song",
				Artists: []spotify.SimpleArtist{
					{Name: "Artist 1"},
					{Name: "Artist 2"},
				},
				URI:      "spotify:track:xyz",
				Duration: 180000,
			},
		},
	})
	if got.Name != "Song" {
		t.Fatalf("Name: want %q, got %q", "Song", got.Name)
	}
	if got.Artist != "Artist 1, Artist 2" {
		t.Fatalf("Artist: want %q, got %q", "Artist 1, Artist 2", got.Artist)
	}
	if got.Duration != 180000 {
		t.Fatalf("Duration: want 180000, got %d", got.Duration)
	}
	if got.URI != "spotify:track:xyz" {
		t.Fatalf("URI: want %q, got %q", "spotify:track:xyz", got.URI)
	}
}

func TestTranslateArtist(t *testing.T) {
	got := spotifyapi.TranslateArtist(spotify.FullArtist{
		SimpleArtist: spotify.SimpleArtist{Name: "Artist"},
		Genres:       []string{"rock", "indie", "alternative"},
	})
	if got.Name != "Artist" {
		t.Fatalf("Name: want %q, got %q", "Artist", got.Name)
	}
	if len(got.Genres) != 3 || got.Genres[0] != "rock" || got.Genres[2] != "alternative" {
		t.Fatalf("Genres: want [rock indie alternative], got %v", got.Genres)
	}
}

func TestTranslateNowPlayingProgress(t *testing.T) {
	got := spotifyapi.TranslateNowPlayingEntry(spotify.PlayerState{
		CurrentlyPlaying: spotify.CurrentlyPlaying{Progress: 30000},
	})
	if got.ProgressMs != 30000 {
		t.Fatalf("ProgressMs: want 30000, got %d", got.ProgressMs)
	}
}

func TestTranslateNowPlayingPlaying(t *testing.T) {
	got := spotifyapi.TranslateNowPlayingEntry(spotify.PlayerState{
		CurrentlyPlaying: spotify.CurrentlyPlaying{Playing: true},
	})
	if !got.Playing {
		t.Fatalf("Playing: want true, got false")
	}
}

func TestTranslateNowPlayingShuffle(t *testing.T) {
	got := spotifyapi.TranslateNowPlayingEntry(spotify.PlayerState{
		ShuffleState: true,
	})
	if !got.ShuffleOn {
		t.Fatalf("ShuffleOn: want true, got false")
	}
}

func TestTranslateNowPlayingTrackName(t *testing.T) {
	got := spotifyapi.TranslateNowPlayingEntry(spotify.PlayerState{
		CurrentlyPlaying: spotify.CurrentlyPlaying{
			Item: &spotify.FullTrack{
				SimpleTrack: spotify.SimpleTrack{
					Name: "Song",
				},
			},
		},
	})
	if got.TrackName != "Song" {
		t.Fatalf("TrackName: want %q, got %q", "Song", got.TrackName)
	}
}

func TestTranslateNowPlayingDuration(t *testing.T) {
	got := spotifyapi.TranslateNowPlayingEntry(spotify.PlayerState{
		CurrentlyPlaying: spotify.CurrentlyPlaying{
			Item: &spotify.FullTrack{
				SimpleTrack: spotify.SimpleTrack{
					Duration: 180000,
				},
			},
		},
	})
	if got.Duration != 180000 {
		t.Fatalf("Duration: want 180000, got %d", got.Duration)
	}
}
