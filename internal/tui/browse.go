package tui

import (
	"fmt"

	"github.com/dcbto/spotui/internal/catalog"
	"github.com/dcbto/spotui/internal/spotengine"
)

type browseItemKind uint8

const (
	browseItemHeader browseItemKind = iota
	browseItemTrack
	browseItemPlaylist
	browseItemAlbum
	browseItemArtist
)

type browseItem struct {
	kind        browseItemKind
	URI         string
	Title       string
	Subtitle    string
	DurationMS  int
	ExternalURL string
	ImageURL    string
}

type browseSnapshot struct {
	route      catalog.Route
	title      string
	items      []browseItem
	cursor     int
	contextURI string
	offset     int
	meta       string
}

func trackItem(track spotengine.Track) browseItem {
	subtitle := track.Artist
	if track.Album != "" {
		subtitle += " · " + track.Album
	}
	if track.Explicit {
		subtitle = "Explicit · " + subtitle
	}
	return browseItem{kind: browseItemTrack, URI: track.URI, Title: track.Name, Subtitle: subtitle, DurationMS: track.DurationMS, ExternalURL: track.ExternalURL, ImageURL: track.ImageURL}
}

func playlistItem(value spotengine.PlaylistSummary) browseItem {
	return browseItem{kind: browseItemPlaylist, URI: value.URI, Title: value.Name, Subtitle: fmt.Sprintf("%s · %d tracks", value.Owner, value.TrackCount), ExternalURL: value.ExternalURL, ImageURL: value.ImageURL}
}

func albumItem(value spotengine.AlbumSummary) browseItem {
	return browseItem{kind: browseItemAlbum, URI: value.URI, Title: value.Name, Subtitle: fmt.Sprintf("%s · %s · %d tracks", value.Artist, value.ReleaseDate, value.TrackCount), ExternalURL: value.ExternalURL, ImageURL: value.ImageURL}
}

func artistItem(value spotengine.ArtistSummary) browseItem {
	return browseItem{kind: browseItemArtist, URI: value.URI, Title: value.Name, Subtitle: "Artist", ExternalURL: value.ExternalURL, ImageURL: value.ImageURL}
}
