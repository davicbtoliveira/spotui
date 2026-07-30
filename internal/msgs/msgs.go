package msgs

import (
	"github.com/dcbto/spotui/internal/spotengine"
	"image"
)

type EngineTrackSearchLoadedMsg struct {
	Tracks []spotengine.Track
	Total  int
	Offset int
}

type CatalogLoadedMsg struct {
	Route     string
	RequestID uint64
	Data      any
	Err       error
}

type ShuffleChangedMsg struct{ Enabled bool }

type ArtworkLoadedMsg struct {
	URL   string
	Image image.Image
	Err   error
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
