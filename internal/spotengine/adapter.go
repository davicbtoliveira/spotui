package spotengine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/devgianlu/go-librespot/ap"
	"github.com/devgianlu/go-librespot/daemon"
	"github.com/devgianlu/go-librespot/login5"
	spotifypb "github.com/devgianlu/go-librespot/proto/spotify"
	login5pb "github.com/devgianlu/go-librespot/proto/spotify/login5/v3"
)

type engineRuntime interface {
	Run(context.Context) error
	Close() error
}

type attemptFactory func() (engineRuntime, *memoryAPIServer, error)

type Adapter struct {
	runtime          engineRuntime
	server           *memoryAPIServer
	factory          attemptFactory
	clearState       func() error
	saveAutoplay     func(bool) error
	sessionAvailable func() bool
	events           chan Event
	hasSession       bool
	autoplay         bool
	mu               sync.Mutex
	cancel           context.CancelFunc
	runContext       context.Context
	runDone          chan error
	stopDone         chan struct{}
	stopErr          error
	closed           bool
	eventsOnce       sync.Once
}

func (a *Adapter) HasSession() bool {
	a.mu.Lock()
	available := a.sessionAvailable
	hasSession := a.hasSession
	a.mu.Unlock()
	if available != nil {
		return available()
	}
	return hasSession
}

func (a *Adapter) AutoplayEnabled() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.autoplay
}

func newAdapter(runtime engineRuntime, server *memoryAPIServer) *Adapter {
	return &Adapter{
		runtime: runtime,
		server:  server,
		events:  server.events,
	}
}

func (a *Adapter) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return errors.New("playback engine closed")
	}
	if a.stopDone != nil {
		select {
		case <-a.stopDone:
			a.stopDone = nil
			a.stopErr = nil
		default:
			return errors.New("playback engine is stopping")
		}
	}
	if a.cancel != nil {
		return errors.New("playback engine already started")
	}
	if a.runtime == nil {
		if a.factory == nil {
			return errors.New("playback engine unavailable")
		}
		runtime, server, err := a.factory()
		if err != nil {
			return fmt.Errorf("create playback engine attempt: %w", err)
		}
		a.runtime = runtime
		a.server = server
	}

	runCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.runContext = runCtx
	a.runDone = make(chan error, 1)
	runDone := a.runDone
	runtime := a.runtime
	server := a.server
	go func() {
		err := runtime.Run(runCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			server.emit(Event{
				Type:      EventTypeError,
				Err:       err,
				ErrorKind: classifyRuntimeError(err),
			})
		}
		runDone <- err
	}()
	return nil
}

func classifyRuntimeError(err error) ErrorKind {
	var accesspointErr *ap.AccesspointLoginError
	if errors.As(err, &accesspointErr) && accesspointErr.Message != nil &&
		accesspointErr.Message.GetErrorCode() == spotifypb.ErrorCode_BadCredentials {
		return ErrorKindCredentialRejected
	}

	var loginErr *login5.LoginError
	if errors.As(err, &loginErr) &&
		(loginErr.Code == login5pb.LoginError_INVALID_CREDENTIALS ||
			loginErr.Code == login5pb.LoginError_UNKNOWN_IDENTIFIER) {
		return ErrorKindCredentialRejected
	}
	return ErrorKindTransient
}

func (a *Adapter) CancelLogin(ctx context.Context) error {
	return a.stop(ctx, true, false, false)
}

func (a *Adapter) Reconnect(ctx context.Context) error {
	return a.stop(ctx, true, false, false)
}

func (a *Adapter) Logout(ctx context.Context) error {
	return a.stop(ctx, true, true, false)
}

func (a *Adapter) Close(ctx context.Context) error {
	return a.stop(ctx, false, false, true)
}

func (a *Adapter) stop(ctx context.Context, restart, clearState, final bool) error {
	a.mu.Lock()
	if a.stopDone != nil && final && !a.closed {
		select {
		case <-a.stopDone:
			a.stopDone = nil
			a.stopErr = nil
		default:
		}
	}
	if a.stopDone == nil {
		done := make(chan struct{})
		a.stopDone = done
		cancel := a.cancel
		runDone := a.runDone
		runtime := a.runtime
		server := a.server
		if cancel != nil {
			cancel()
		}
		go func() {
			var serverErr error
			// Close the API boundary before waiting for Run. Runtime callbacks
			// may be blocked publishing an event; closing the boundary releases
			// them so the runtime can finish and report through runDone.
			if server != nil {
				serverErr = server.Close()
			}
			var closeErr error
			if runtime != nil {
				closeErr = runtime.Close()
			}
			var runErr error
			if runDone != nil {
				runErr = <-runDone
			}
			if errors.Is(runErr, context.Canceled) {
				runErr = nil
			}
			stopErr := errors.Join(closeErr, runErr, serverErr)
			clearSucceeded := !clearState || a.clearState == nil
			if clearState && a.clearState != nil {
				clearErr := a.clearState()
				clearSucceeded = clearErr == nil
				stopErr = errors.Join(stopErr, clearErr)
			}
			if restart {
				drainEvents(a.events)
			}
			var nextRuntime engineRuntime
			var nextServer *memoryAPIServer
			if restart && a.factory != nil && !(clearState && clearSucceeded) {
				var err error
				nextRuntime, nextServer, err = a.factory()
				stopErr = errors.Join(stopErr, err)
			}

			a.mu.Lock()
			a.cancel = nil
			a.runContext = nil
			a.runDone = nil
			a.stopErr = stopErr
			if clearState && clearSucceeded {
				a.hasSession = false
			}
			if final {
				a.closed = true
				a.eventsOnce.Do(func() { close(a.events) })
			} else if restart {
				a.runtime = nextRuntime
				a.server = nextServer
			}
			if restart {
				a.publishSessionEndedLocked()
			}
			close(done)
			a.mu.Unlock()
		}()
	}
	done := a.stopDone
	a.mu.Unlock()

	select {
	case <-done:
		a.mu.Lock()
		err := a.stopErr
		closed := a.closed
		a.mu.Unlock()
		if final && !closed {
			if ctx.Err() != nil {
				return errors.Join(err, ctx.Err())
			}
			return errors.Join(err, a.stop(ctx, false, false, true))
		}
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *Adapter) publishSessionEndedLocked() {
	if a.closed {
		return
	}
	select {
	case a.events <- Event{Type: EventTypeSessionEnded}:
	default:
	}
}

func drainEvents(events <-chan Event) {
	for {
		select {
		case <-events:
		default:
			return
		}
	}
}

func (a *Adapter) Events() <-chan Event {
	return a.events
}

func (a *Adapter) request(ctx context.Context, requestType daemon.ApiRequestType, data any) (any, error) {
	a.mu.Lock()
	server := a.server
	runContext := a.runContext
	a.mu.Unlock()
	if server == nil {
		return nil, errors.New("playback engine unavailable")
	}
	if runContext == nil {
		return nil, errors.New("playback engine is not running")
	}
	requestContext, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(runContext, cancel)
	defer func() {
		stop()
		cancel()
	}()
	return server.request(requestContext, requestType, data)
}

func (a *Adapter) Play(ctx context.Context, uri string) error {
	_, err := a.request(ctx, daemon.ApiRequestTypePlay, daemon.ApiRequestDataPlay{
		Uri:       uri,
		SkipToUri: uri,
		Position:  0,
	})
	return err
}

func (a *Adapter) searchTracks(ctx context.Context, request SearchRequest) (TrackPage, error) {
	response, err := a.request(ctx, daemon.ApiRequestTypeSearch, daemon.ApiRequestDataSearch{
		Query:  request.Query,
		Offset: request.Offset,
		Limit:  request.Limit,
	})
	if err != nil {
		return TrackPage{}, err
	}

	searchResponse, ok := response.(daemon.ApiResponseSearch)
	if !ok {
		return TrackPage{}, fmt.Errorf("decode track search response: unexpected %T", response)
	}

	page := TrackPage{
		Items:  make([]Track, len(searchResponse.Tracks)),
		Total:  searchResponse.Total,
		Offset: searchResponse.Offset,
	}
	for i, item := range searchResponse.Tracks {
		page.Items[i] = Track{
			URI:        item.Uri,
			Name:       item.Name,
			Artist:     strings.Join(item.ArtistNames, ", "),
			Album:      item.AlbumName,
			DurationMS: item.Duration,
		}
	}
	return page, nil
}

func (a *Adapter) Pause(ctx context.Context) error {
	return a.command(ctx, daemon.ApiRequestTypePause, nil)
}

func (a *Adapter) Resume(ctx context.Context) error {
	return a.command(ctx, daemon.ApiRequestTypeResume, nil)
}

func (a *Adapter) Next(ctx context.Context) error {
	return a.command(ctx, daemon.ApiRequestTypeNext, daemon.ApiRequestDataNext{})
}

func (a *Adapter) Previous(ctx context.Context) error {
	return a.command(ctx, daemon.ApiRequestTypePrev, nil)
}

func (a *Adapter) SetVolume(ctx context.Context, volume int) error {
	return a.command(ctx, daemon.ApiRequestTypeSetVolume, daemon.ApiRequestDataVolume{
		Volume: int32(volume),
	})
}

func (a *Adapter) SetAutoplay(ctx context.Context, enabled bool) error {
	if err := a.command(ctx, daemon.ApiRequestTypeSetAutoplay, enabled); err != nil {
		return err
	}
	if a.saveAutoplay != nil {
		if err := a.saveAutoplay(enabled); err != nil {
			return err
		}
	}
	a.mu.Lock()
	a.autoplay = enabled
	a.mu.Unlock()
	return nil
}

func (a *Adapter) command(ctx context.Context, requestType daemon.ApiRequestType, data any) error {
	_, err := a.request(ctx, requestType, data)
	return err
}
