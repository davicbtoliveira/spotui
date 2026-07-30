package tui

import (
	"fmt"
	"github.com/dcbto/spotui/internal/spotengine"
)

const (
	sectionLibrary = iota
	sectionRecommended
	sectionSearch
)

type browseItem struct {
	kind        string
	URI         string
	Title       string
	Subtitle    string
	DurationMS  int
	ExternalURL string
	ImageURL    string
}

type browseSnapshot struct {
	route       string
	title       string
	items       []browseItem
	cursor      int
	contextURI  string
	contextKind string
	offset      int
	meta        string
}

func trackItem(track spotengine.Track) browseItem {
	subtitle := track.Artist
	if track.Album != "" {
		subtitle += " · " + track.Album
	}
	if track.Explicit {
		subtitle = "Explicit · " + subtitle
	}
	return browseItem{kind: "track", URI: track.URI, Title: track.Name, Subtitle: subtitle, DurationMS: track.DurationMS, ExternalURL: track.ExternalURL, ImageURL: track.ImageURL}
}

func playlistItem(value spotengine.PlaylistSummary) browseItem {
	return browseItem{kind: "playlist", URI: value.URI, Title: value.Name, Subtitle: fmt.Sprintf("%s · %d tracks", value.Owner, value.TrackCount), ExternalURL: value.ExternalURL, ImageURL: value.ImageURL}
}

func albumItem(value spotengine.AlbumSummary) browseItem {
	return browseItem{kind: "album", URI: value.URI, Title: value.Name, Subtitle: fmt.Sprintf("%s · %s · %d tracks", value.Artist, value.ReleaseDate, value.TrackCount), ExternalURL: value.ExternalURL, ImageURL: value.ImageURL}
}

func artistItem(value spotengine.ArtistSummary) browseItem {
	return browseItem{kind: "artist", URI: value.URI, Title: value.Name, Subtitle: "Artist", ExternalURL: value.ExternalURL, ImageURL: value.ImageURL}
}
