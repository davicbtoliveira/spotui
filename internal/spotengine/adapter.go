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

type Adapter struct {
	runtime    engineRuntime
	server     *memoryAPIServer
	hasSession bool

	mu        sync.Mutex
	cancel    context.CancelFunc
	runDone   chan error
	closeOnce sync.Once
	closeDone chan struct{}
	closeErr  error
}

func (a *Adapter) HasSession() bool {
	return a.hasSession
}

func newAdapter(runtime engineRuntime, server *memoryAPIServer) *Adapter {
	return &Adapter{
		runtime:   runtime,
		server:    server,
		closeDone: make(chan struct{}),
	}
}

func (a *Adapter) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		return errors.New("playback engine already started")
	}

	runCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.runDone = make(chan error, 1)
	go func() {
		err := a.runtime.Run(runCtx)
		if err != nil && !errors.Is(err, context.Canceled) {
			a.server.emit(Event{Type: EventTypeError, Err: err})
		}
		a.runDone <- err
	}()
	return nil
}

func (a *Adapter) Close(ctx context.Context) error {
	a.mu.Lock()
	cancel := a.cancel
	runDone := a.runDone
	a.mu.Unlock()

	if cancel == nil {
		return errors.New("playback engine not started")
	}

	a.closeOnce.Do(func() {
		cancel()
		go func() {
			closeErr := a.runtime.Close()
			runErr := <-runDone
			if errors.Is(runErr, context.Canceled) {
				runErr = nil
			}
			a.closeErr = errors.Join(closeErr, runErr, a.server.Close())
			a.server.finish()
			close(a.closeDone)
		}()
	})

	select {
	case <-a.closeDone:
		return a.closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *Adapter) Events() <-chan Event {
	return a.server.events
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
	return a.command(ctx, daemon.ApiRequestTypeSetAutoplay, enabled)
}

func (a *Adapter) command(ctx context.Context, requestType daemon.ApiRequestType, data any) error {
	_, err := a.server.request(ctx, requestType, data)
	return err
}
