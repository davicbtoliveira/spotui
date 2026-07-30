package spotengine

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/devgianlu/go-librespot/daemon"
)

type memoryAPIServer struct {
	requests chan daemon.ApiRequest
	events   chan Event
	closed   chan struct{}
	once     sync.Once
}

func newMemoryAPIServer() *memoryAPIServer {
	return newMemoryAPIServerWithEvents(make(chan Event, 64))
}

func newMemoryAPIServerWithEvents(events chan Event) *memoryAPIServer {
	return &memoryAPIServer{
		requests: make(chan daemon.ApiRequest),
		events:   events,
		closed:   make(chan struct{}),
	}
}

func (s *memoryAPIServer) Emit(event *daemon.ApiEvent) {
	var translated Event
	switch event.Type {
	case daemon.ApiEventTypePlaybackReady:
		translated.Type = EventTypeReady
	case daemon.ApiEventTypeWillPlay:
		translated.Type = EventTypeBuffering
	case daemon.ApiEventTypePlaying:
		data := event.Data.(daemon.ApiEventDataPlaying)
		translated = Event{
			Type:       EventTypePlaying,
			ContextURI: data.ContextUri,
			URI:        data.Uri,
		}
	case daemon.ApiEventTypePaused:
		data := event.Data.(daemon.ApiEventDataPaused)
		translated = Event{
			Type:       EventTypePaused,
			ContextURI: data.ContextUri,
			URI:        data.Uri,
		}
	case daemon.ApiEventTypeStopped:
		translated.Type = EventTypeStopped
	case daemon.ApiEventTypeActive:
		translated.Type = EventTypeActive
	case daemon.ApiEventTypeInactive:
		translated.Type = EventTypeInactive
	case daemon.ApiEventTypeMetadata:
		data := event.Data.(daemon.ApiEventDataMetadata)
		translated = Event{
			Type: EventTypeMetadata,
			Track: &Track{
				URI:        data.Uri,
				Name:       data.Name,
				Artist:     strings.Join(data.ArtistNames, ", "),
				Album:      data.AlbumName,
				DurationMS: data.Duration,
			},
			PositionMS: int(data.Position),
			DurationMS: data.Duration,
		}
	case daemon.ApiEventTypeVolume:
		data := event.Data.(daemon.ApiEventDataVolume)
		translated = Event{
			Type:      EventTypeVolume,
			Volume:    int(data.Value),
			VolumeMax: int(data.Max),
		}
	case daemon.ApiEventTypeSeek:
		data := event.Data.(daemon.ApiEventDataSeek)
		translated = Event{
			Type:       EventTypeSeek,
			ContextURI: data.ContextUri,
			URI:        data.Uri,
			PositionMS: data.Position,
			DurationMS: data.Duration,
		}
	case daemon.ApiEventTypeAccountProduct:
		data := event.Data.(daemon.ApiEventDataAccountProduct)
		translated = Event{
			Type:    EventTypeAccountProduct,
			Product: data.Product,
		}
	default:
		return
	}
	s.emit(translated)
}

func (s *memoryAPIServer) Receive() <-chan daemon.ApiRequest {
	return s.requests
}

func (s *memoryAPIServer) Close() error {
	s.once.Do(func() { close(s.closed) })
	return nil
}

func (s *memoryAPIServer) emit(event Event) {
	select {
	case s.events <- event:
	case <-s.closed:
	}
}

func (s *memoryAPIServer) request(ctx context.Context, requestType daemon.ApiRequestType, data any) (any, error) {
	request, wait := daemon.NewApiRequest(requestType, data)
	select {
	case s.requests <- request:
		return wait(ctx)
	case <-s.closed:
		return nil, errors.New("playback engine closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
