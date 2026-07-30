package tui

// Message types live in internal/msgs to avoid import cycles.
// Re-export as type aliases so callers inside this package use the short name.

import "github.com/dcbto/spotui/internal/msgs"

type AuthDoneMsg = msgs.AuthDoneMsg
type AuthErrMsg = msgs.AuthErrMsg
type UserLoadedMsg = msgs.UserLoadedMsg
type PlaylistsLoadedMsg = msgs.PlaylistsLoadedMsg
type TracksLoadedMsg = msgs.TracksLoadedMsg
type TrackSearchLoadedMsg = msgs.TrackSearchLoadedMsg
type EngineTrackSearchLoadedMsg = msgs.EngineTrackSearchLoadedMsg
type ArtistsLoadedMsg = msgs.ArtistsLoadedMsg
type NowPlayingMsg = msgs.NowPlayingMsg
type TickMsg = msgs.TickMsg
type ErrMsg = msgs.ErrMsg
type ClearStatusMsg = msgs.ClearStatusMsg
type ProgressTickMsg = msgs.ProgressTickMsg
type EngineStartedMsg = msgs.EngineStartedMsg
type EngineStartErrMsg = msgs.EngineStartErrMsg
type EngineEventMsg = msgs.EngineEventMsg
type EngineEventsClosedMsg = msgs.EngineEventsClosedMsg
type BrowserOpenedMsg = msgs.BrowserOpenedMsg
type BrowserOpenErrMsg = msgs.BrowserOpenErrMsg
type LoginResetMsg = msgs.LoginResetMsg
type LoginResetErrMsg = msgs.LoginResetErrMsg
type AutoplayChangedMsg = msgs.AutoplayChangedMsg
type LogoutDoneMsg = msgs.LogoutDoneMsg
type LogoutErrMsg = msgs.LogoutErrMsg
