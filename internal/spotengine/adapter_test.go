package spotengine

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/devgianlu/go-librespot/daemon"
	"go.uber.org/goleak"
)

type lifecycleRuntime struct {
	started chan struct{}
	closed  chan struct{}
	once    sync.Once
}

type stubbornRuntime struct {
	started chan struct{}
	release chan struct{}
}

type failingRuntime struct {
	err error
}

func (r failingRuntime) Run(context.Context) error {
	return r.err
}

func (failingRuntime) Close() error {
	return nil
}

type requestRuntime struct {
	server   *memoryAPIServer
	requests chan daemon.ApiRequest
	reply    func(daemon.ApiRequest) (any, error)
}

func (r *requestRuntime) Run(ctx context.Context) error {
	for {
		select {
		case request := <-r.server.Receive():
			r.requests <- request
			var data any
			var err error
			if r.reply != nil {
				data, err = r.reply(request)
			}
			request.Reply(data, err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func TestAdapterTranslatesTrackSearchResponse(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	server := newMemoryAPIServer()
	runtime := &requestRuntime{
		server:   server,
		requests: make(chan daemon.ApiRequest, 1),
		reply: func(daemon.ApiRequest) (any, error) {
			return map[string]any{
				"tracks": map[string]any{
					"items": []any{
						map[string]any{
							"uri":         "spotify:track:hello",
							"name":        "Hello",
							"duration_ms": float64(295000),
							"artists": []any{
								map[string]any{"name": "Adele"},
							},
							"album": map[string]any{"name": "25"},
						},
					},
					"total":  float64(1),
					"offset": float64(20),
				},
			}, nil
		},
	}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	page, err := adapter.SearchTracks(context.Background(), SearchRequest{
		Query:  "hello",
		Offset: 20,
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("search tracks: %v", err)
	}
	if len(page.Tracks) != 1 {
		t.Fatalf("tracks: want 1, got %d", len(page.Tracks))
	}
	if got := page.Tracks[0]; got.URI != "spotify:track:hello" || got.Name != "Hello" ||
		got.Artist != "Adele" || got.Album != "25" || got.DurationMS != 295000 {
		t.Fatalf("track: %#v", got)
	}
	if page.Total != 1 || page.Offset != 20 {
		t.Fatalf("page: %#v", page)
	}

	request := <-runtime.requests
	data, ok := request.Data.(daemon.ApiRequestDataWebApi)
	if request.Type != daemon.ApiRequestTypeWebApi || !ok {
		t.Fatalf("request: type %q, data %T", request.Type, request.Data)
	}
	if data.Method != "GET" || data.Path != "/v1/search" ||
		data.Query.Get("q") != "hello" || data.Query.Get("type") != "track" ||
		data.Query.Get("offset") != "20" || data.Query.Get("limit") != "20" {
		t.Fatalf("request data: %#v", data)
	}

	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestAdapterTranslatesMetadataEvent(t *testing.T) {
	server := newMemoryAPIServer()
	adapter := newAdapter(newLifecycleRuntime(), server)

	server.Emit(&daemon.ApiEvent{
		Type: daemon.ApiEventTypeMetadata,
		Data: daemon.ApiEventDataMetadata(daemon.ApiResponseStatusTrack{
			Uri:         "spotify:track:hello",
			Name:        "Hello",
			ArtistNames: []string{"Adele"},
			AlbumName:   "25",
			Position:    12000,
			Duration:    295000,
		}),
	})

	select {
	case event := <-adapter.Events():
		if event.Type != EventTypeMetadata || event.Track == nil {
			t.Fatalf("event: %#v", event)
		}
		if event.Track.URI != "spotify:track:hello" || event.Track.Artist != "Adele" ||
			event.PositionMS != 12000 {
			t.Fatalf("event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("metadata event was not published")
	}
}

func TestAdapterTranslatesPlaybackEvents(t *testing.T) {
	server := newMemoryAPIServer()
	adapter := newAdapter(newLifecycleRuntime(), server)

	upstream := []*daemon.ApiEvent{
		{Type: daemon.ApiEventTypePlaybackReady},
		{
			Type: daemon.ApiEventTypePlaying,
			Data: daemon.ApiEventDataPlaying{
				ContextUri: "spotify:album:25",
				Uri:        "spotify:track:hello",
			},
		},
		{
			Type: daemon.ApiEventTypePaused,
			Data: daemon.ApiEventDataPaused{
				ContextUri: "spotify:album:25",
				Uri:        "spotify:track:hello",
			},
		},
		{
			Type: daemon.ApiEventTypeVolume,
			Data: daemon.ApiEventDataVolume{Value: 65, Max: 100},
		},
		{
			Type: daemon.ApiEventTypeSeek,
			Data: daemon.ApiEventDataSeek{Position: 12000, Duration: 295000},
		},
		{Type: daemon.ApiEventTypeInactive},
		{Type: daemon.ApiEventTypeStopped},
	}
	want := []EventType{
		EventTypeReady,
		EventTypePlaying,
		EventTypePaused,
		EventTypeVolume,
		EventTypeSeek,
		EventTypeInactive,
		EventTypeStopped,
	}

	for i, event := range upstream {
		server.Emit(event)
		got := <-adapter.Events()
		if got.Type != want[i] {
			t.Fatalf("event %d: want %q, got %q", i, want[i], got.Type)
		}
		if got.Type == EventTypePlaying &&
			(got.ContextURI != "spotify:album:25" || got.URI != "spotify:track:hello") {
			t.Fatalf("playing event: %#v", got)
		}
		if got.Type == EventTypeVolume && (got.Volume != 65 || got.VolumeMax != 100) {
			t.Fatalf("volume event: %#v", got)
		}
		if got.Type == EventTypeSeek && (got.PositionMS != 12000 || got.DurationMS != 295000) {
			t.Fatalf("seek event: %#v", got)
		}
	}
}

func TestAdapterTranslatesAccountProductEvent(t *testing.T) {
	server := newMemoryAPIServer()
	adapter := newAdapter(newLifecycleRuntime(), server)

	server.Emit(&daemon.ApiEvent{
		Type: daemon.ApiEventTypeAccountProduct,
		Data: daemon.ApiEventDataAccountProduct{Product: "free"},
	})

	event := <-adapter.Events()
	if event.Type != EventTypeAccountProduct || event.Product != "free" {
		t.Fatalf("event: %#v", event)
	}
}

func TestNewAdapterReturnsStableEngine(t *testing.T) {
	adapter, err := newAdapterAtDir(t.TempDir())
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	var engine Engine = adapter
	if engine == nil {
		t.Fatal("adapter does not implement Engine")
	}
}

func TestAdapterPublishesRuntimeFailure(t *testing.T) {
	want := errors.New("playback runtime failed")
	adapter := newAdapter(failingRuntime{err: want}, newMemoryAPIServer())

	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	select {
	case event := <-adapter.Events():
		if event.Type != EventTypeError || !errors.Is(event.Err, want) {
			t.Fatalf("event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("runtime error event was not published")
	}

	if err := adapter.Close(context.Background()); !errors.Is(err, want) {
		t.Fatalf("close: want runtime error, got %v", err)
	}
}

func (r *requestRuntime) Close() error {
	return nil
}

func (r *stubbornRuntime) Run(context.Context) error {
	close(r.started)
	<-r.release
	return nil
}

func (r *stubbornRuntime) Close() error {
	return nil
}

func newLifecycleRuntime() *lifecycleRuntime {
	return &lifecycleRuntime{
		started: make(chan struct{}),
		closed:  make(chan struct{}),
	}
}

func (r *lifecycleRuntime) Run(ctx context.Context) error {
	close(r.started)
	<-ctx.Done()
	return ctx.Err()
}

func (r *lifecycleRuntime) Close() error {
	r.once.Do(func() { close(r.closed) })
	return nil
}

func TestAdapterStartsAndClosesRuntimeInProcess(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	runtime := newLifecycleRuntime()
	adapter := newAdapter(runtime, newMemoryAPIServer())

	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	select {
	case <-runtime.started:
	case <-time.After(time.Second):
		t.Fatal("runtime did not start")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := adapter.Close(ctx); err != nil {
		t.Fatalf("close: %v", err)
	}
	select {
	case <-runtime.closed:
	default:
		t.Fatal("runtime Close was not called")
	}
	select {
	case _, ok := <-adapter.Events():
		if ok {
			t.Fatal("events channel published data after close")
		}
	case <-time.After(time.Second):
		t.Fatal("events channel remains open after close")
	}
}

func TestAdapterReportsRuntimeCancellationFailure(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	runtime := &stubbornRuntime{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	adapter := newAdapter(runtime, newMemoryAPIServer())
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	<-runtime.started

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := adapter.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("close: want deadline exceeded, got %v", err)
	}

	close(runtime.release)
	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close after runtime release: %v", err)
	}
}

func TestAdapterCancelsLoginAndStartsFreshAttempt(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	events := make(chan Event, 64)
	var runtimes []*lifecycleRuntime
	factory := func() (engineRuntime, *memoryAPIServer, error) {
		runtime := newLifecycleRuntime()
		runtimes = append(runtimes, runtime)
		return runtime, newMemoryAPIServerWithEvents(events), nil
	}
	runtime, server, err := factory()
	if err != nil {
		t.Fatalf("create first attempt: %v", err)
	}
	adapter := newAdapter(runtime, server)
	adapter.factory = factory

	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start first attempt: %v", err)
	}
	<-runtimes[0].started

	if err := adapter.CancelLogin(context.Background()); err != nil {
		t.Fatalf("cancel login: %v", err)
	}
	select {
	case <-runtimes[0].closed:
	default:
		t.Fatal("first runtime was not closed")
	}
	select {
	case event, ok := <-adapter.Events():
		if !ok {
			t.Fatal("cancel closed the stable events channel")
		}
		if event.Type != EventTypeSessionEnded {
			t.Fatalf("cancel event: want session ended, got %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("cancel did not release the pending event listener")
	}

	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start second attempt: %v", err)
	}
	if len(runtimes) != 2 {
		t.Fatalf("runtime attempts: want 2, got %d", len(runtimes))
	}
	<-runtimes[1].started

	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestAdapterLogoutClearsSessionBeforeFreshAttempt(t *testing.T) {
	events := make(chan Event, 64)
	factory := func() (engineRuntime, *memoryAPIServer, error) {
		return newLifecycleRuntime(), newMemoryAPIServerWithEvents(events), nil
	}
	runtime, server, err := factory()
	if err != nil {
		t.Fatalf("create attempt: %v", err)
	}
	adapter := newAdapter(runtime, server)
	adapter.factory = factory
	adapter.hasSession = true
	server.emit(Event{Type: EventTypeReady})
	cleared := false
	adapter.clearState = func() error {
		cleared = true
		return nil
	}

	if err := adapter.Logout(context.Background()); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !cleared {
		t.Fatal("logout did not clear persisted session")
	}
	if adapter.HasSession() {
		t.Fatal("logout retained session state")
	}
	if event := <-adapter.Events(); event.Type != EventTypeSessionEnded {
		t.Fatalf("logout retained stale event: %#v", event)
	}

	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start after logout: %v", err)
	}
	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestAdapterRetriesAttemptCreationAfterResetFailure(t *testing.T) {
	events := make(chan Event, 64)
	firstRuntime := newLifecycleRuntime()
	adapter := newAdapter(firstRuntime, newMemoryAPIServerWithEvents(events))
	want := errors.New("create attempt")
	factoryCalls := 0
	adapter.factory = func() (engineRuntime, *memoryAPIServer, error) {
		factoryCalls++
		if factoryCalls == 1 {
			return nil, nil, want
		}
		return newLifecycleRuntime(), newMemoryAPIServerWithEvents(events), nil
	}

	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start first attempt: %v", err)
	}
	<-firstRuntime.started
	if err := adapter.CancelLogin(context.Background()); !errors.Is(err, want) {
		t.Fatalf("cancel error: want %v, got %v", want, err)
	}

	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("retry start: %v", err)
	}
	if factoryCalls != 2 {
		t.Fatalf("factory calls: want 2, got %d", factoryCalls)
	}
	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestAdapterPlaysTrackThroughInMemoryRequest(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	server := newMemoryAPIServer()
	runtime := &requestRuntime{
		server:   server,
		requests: make(chan daemon.ApiRequest, 1),
	}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	if err := adapter.Play(context.Background(), "spotify:track:hello"); err != nil {
		t.Fatalf("play: %v", err)
	}
	request := <-runtime.requests
	if request.Type != daemon.ApiRequestTypePlay {
		t.Fatalf("request type: want %q, got %q", daemon.ApiRequestTypePlay, request.Type)
	}
	data, ok := request.Data.(daemon.ApiRequestDataPlay)
	if !ok {
		t.Fatalf("request data: want ApiRequestDataPlay, got %T", request.Data)
	}
	if data.Uri != "spotify:track:hello" || data.SkipToUri != "spotify:track:hello" || data.Position != 0 {
		t.Fatalf("request data: %#v", data)
	}

	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestAdapterSendsPlayerControlsThroughInMemoryRequests(t *testing.T) {
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	server := newMemoryAPIServer()
	runtime := &requestRuntime{
		server:   server,
		requests: make(chan daemon.ApiRequest, 8),
	}
	adapter := newAdapter(runtime, server)
	if err := adapter.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	ctx := context.Background()
	controls := []struct {
		name string
		call func() error
		want daemon.ApiRequestType
	}{
		{"pause", func() error { return adapter.Pause(ctx) }, daemon.ApiRequestTypePause},
		{"resume", func() error { return adapter.Resume(ctx) }, daemon.ApiRequestTypeResume},
		{"next", func() error { return adapter.Next(ctx) }, daemon.ApiRequestTypeNext},
		{"previous", func() error { return adapter.Previous(ctx) }, daemon.ApiRequestTypePrev},
		{"volume", func() error { return adapter.SetVolume(ctx, 65) }, daemon.ApiRequestTypeSetVolume},
		{"autoplay", func() error { return adapter.SetAutoplay(ctx, false) }, daemon.ApiRequestTypeSetAutoplay},
	}
	for _, control := range controls {
		if err := control.call(); err != nil {
			t.Fatalf("%s: %v", control.name, err)
		}
		if request := <-runtime.requests; request.Type != control.want {
			t.Fatalf("%s request: want %q, got %q", control.name, control.want, request.Type)
		}
	}

	if err := adapter.Close(context.Background()); err != nil {
		t.Fatalf("close: %v", err)
	}
}
