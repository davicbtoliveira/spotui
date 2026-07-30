package msgs

import "github.com/dcbto/spotui/internal/spotengine"

type EngineTrackSearchLoadedMsg struct {
	Tracks []spotengine.Track
	Total  int
	Offset int
}

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

type AutoplayChangedMsg struct {
	Enabled bool
}

type LogoutDoneMsg struct{}

type LogoutErrMsg struct {
	Err error
}

type EngineReconnectedMsg struct{}

type EngineReconnectErrMsg struct {
	Err error
}

type ReconnectTimerMsg struct{}

type SessionExpiredMsg struct{}

type SessionExpireErrMsg struct {
	Err error
}
