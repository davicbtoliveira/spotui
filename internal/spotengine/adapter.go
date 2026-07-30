package spotengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/devgianlu/go-librespot/daemon"
)

type engineRuntime interface {
	Run(context.Context) error
	Close() error
}

type attemptFactory func() (engineRuntime, *memoryAPIServer, error)

type Adapter struct {
	runtime      engineRuntime
	server       *memoryAPIServer
	factory      attemptFactory
	clearState   func() error
	saveAutoplay func(bool) error
	events       chan Event
	hasSession   bool
	autoplay     bool

	mu         sync.Mutex
	cancel     context.CancelFunc
	runDone    chan error
	stopDone   chan struct{}
	stopErr    error
	closed     bool
	eventsOnce sync.Once
}

func (a *Adapter) HasSession() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.hasSession
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
	a.runDone = make(chan error, 1)
	runDone := a.runDone
	runtime := a.runtime
	server := a.server
	go func() {
		err := runtime.Run(runCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			server.emit(Event{Type: EventTypeError, Err: err})
		}
		runDone <- err
	}()
	return nil
}

func (a *Adapter) CancelLogin(ctx context.Context) error {
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
			var serverErr error
			if server != nil {
				serverErr = server.Close()
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
				a.events <- Event{Type: EventTypeSessionEnded}
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
		a.mu.Unlock()
		return err
	case <-ctx.Done():
		return ctx.Err()
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

func (a *Adapter) Play(ctx context.Context, uri string) error {
	_, err := a.server.request(ctx, daemon.ApiRequestTypePlay, daemon.ApiRequestDataPlay{
		Uri:       uri,
		SkipToUri: uri,
		Position:  0,
	})
	return err
}

func (a *Adapter) SearchTracks(ctx context.Context, request SearchRequest) (SearchPage, error) {
	response, err := a.server.request(ctx, daemon.ApiRequestTypeWebApi, daemon.ApiRequestDataWebApi{
		Method: "GET",
		Path:   "/v1/search",
		Query: url.Values{
			"q":      {request.Query},
			"type":   {"track"},
			"offset": {strconv.Itoa(request.Offset)},
			"limit":  {strconv.Itoa(request.Limit)},
		},
	})
	if err != nil {
		return SearchPage{}, err
	}

	var payload struct {
		Tracks struct {
			Items []struct {
				URI        string `json:"uri"`
				Name       string `json:"name"`
				DurationMS int    `json:"duration_ms"`
				Artists    []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					Name string `json:"name"`
				} `json:"album"`
			} `json:"items"`
			Total  int `json:"total"`
			Offset int `json:"offset"`
		} `json:"tracks"`
	}
	data, err := json.Marshal(response)
	if err != nil {
		return SearchPage{}, fmt.Errorf("encode track search response: %w", err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return SearchPage{}, fmt.Errorf("decode track search response: %w", err)
	}

	page := SearchPage{
		Tracks: make([]Track, len(payload.Tracks.Items)),
		Total:  payload.Tracks.Total,
		Offset: payload.Tracks.Offset,
	}
	for i, item := range payload.Tracks.Items {
		artists := make([]string, len(item.Artists))
		for j, artist := range item.Artists {
			artists[j] = artist.Name
		}
		page.Tracks[i] = Track{
			URI:        item.URI,
			Name:       item.Name,
			Artist:     strings.Join(artists, ", "),
			Album:      item.Album.Name,
			DurationMS: item.DurationMS,
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
	a.mu.Lock()
	server := a.server
	a.mu.Unlock()
	_, err := server.request(ctx, requestType, data)
	return err
}
