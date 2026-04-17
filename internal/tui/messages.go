package tui

// Message types live in internal/msgs to avoid import cycles.
// Re-export as type aliases so callers inside this package use the short name.

import "github.com/dcbto/spotui/internal/msgs"

type AuthDoneMsg = msgs.AuthDoneMsg
type AuthErrMsg = msgs.AuthErrMsg
type UserLoadedMsg = msgs.UserLoadedMsg
type PlaylistsLoadedMsg = msgs.PlaylistsLoadedMsg
type TracksLoadedMsg = msgs.TracksLoadedMsg
type ArtistsLoadedMsg = msgs.ArtistsLoadedMsg
type NowPlayingMsg = msgs.NowPlayingMsg
type TickMsg = msgs.TickMsg
type ErrMsg = msgs.ErrMsg
type ClearStatusMsg = msgs.ClearStatusMsg
type ProgressTickMsg = msgs.ProgressTickMsg
