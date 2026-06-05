package library

type PlaylistEntry struct {
	Name       string
	TrackCount int
	URI        string
}

type TrackEntry struct {
	Name     string
	Artist   string
	Duration int
	URI      string
}

type ArtistEntry struct {
	Name   string
	Genres []string
}
