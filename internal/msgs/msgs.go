package msgs

import (
	"github.com/dcbto/spotui/internal/library"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/zmb3/spotify/v2"
)

type AuthDoneMsg struct {
	Client *spotify.Client
}

type AuthErrMsg struct {
	Err error
}

type UserLoadedMsg struct {
	User *spotify.PrivateUser
}

type PlaylistsLoadedMsg struct {
	Playlists []library.PlaylistEntry
}

type TracksLoadedMsg struct {
	Tracks []spotify.SavedTrack
	Total  int
}

type TrackSearchLoadedMsg struct {
	Tracks []spotify.FullTrack
	Total  int
	Offset int
}

type ArtistsLoadedMsg struct {
	Artists []spotify.FullArtist
}

type NowPlayingMsg struct {
	State *spotify.PlayerState
}

type TickMsg struct{}

type ErrMsg struct {
	Err     error
	Context string
}

type ClearStatusMsg struct{}

type ProgressTickMsg struct{}

type EngineStartedMsg struct{}

type EngineStartErrMsg struct {
	Err error
}

type EngineEventMsg struct {
	Event spotengine.Event
}

type EngineEventsClosedMsg struct{}

type BrowserOpenedMsg struct{}

type BrowserOpenErrMsg struct {
	Err error
}

type LoginResetMsg struct{}

type LoginResetErrMsg struct {
	Err error
}
