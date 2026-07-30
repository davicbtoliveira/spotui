package spotengine

import (
	"context"
	"sync"
)

type Operation string

const (
	OperationStart        Operation = "start"
	OperationCancelLogin  Operation = "cancel_login"
	OperationLogout       Operation = "logout"
	OperationSearchTracks Operation = "search_tracks"
	OperationPlay         Operation = "play"
	OperationPause        Operation = "pause"
	OperationResume       Operation = "resume"
	OperationNext         Operation = "next"
	OperationPrevious     Operation = "previous"
	OperationSetVolume    Operation = "set_volume"
	OperationSetAutoplay  Operation = "set_autoplay"
	OperationClose        Operation = "close"
)

type Call struct {
	Operation Operation
	URI       string
	Search    SearchRequest
	Volume    int
	Enabled   bool
}

type Fake struct {
	mu          sync.Mutex
	calls       []Call
	events      chan Event
	errors      map[Operation]error
	searchPage  SearchPage
	searchError error
	hasSession  bool
	autoplay    bool
}

func (f *Fake) HasSession() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.hasSession
}

func (f *Fake) SetHasSession(hasSession bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.hasSession = hasSession
}

func NewFake() *Fake {
	return &Fake{
		events:   make(chan Event, 16),
		errors:   make(map[Operation]error),
		autoplay: true,
	}
}

func (f *Fake) AutoplayEnabled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.autoplay
}

func (f *Fake) Emit(event Event) {
	f.events <- event
}

func (f *Fake) Events() <-chan Event {
	return f.events
}

func (f *Fake) Play(_ context.Context, uri string) error {
	return f.record(Call{Operation: OperationPlay, URI: uri})
}

func (f *Fake) Start(_ context.Context) error {
	return f.record(Call{Operation: OperationStart})
}

func (f *Fake) CancelLogin(_ context.Context) error {
	return f.record(Call{Operation: OperationCancelLogin})
}

func (f *Fake) Logout(_ context.Context) error {
	if err := f.record(Call{Operation: OperationLogout}); err != nil {
		return err
	}
	f.SetHasSession(false)
	return nil
}

func (f *Fake) SearchTracks(_ context.Context, request SearchRequest) (SearchPage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, Call{Operation: OperationSearchTracks, Search: request})
	if err := f.errors[OperationSearchTracks]; err != nil {
		return SearchPage{}, err
	}
	return f.searchPage, f.searchError
}

func (f *Fake) Pause(_ context.Context) error {
	return f.record(Call{Operation: OperationPause})
}

func (f *Fake) Resume(_ context.Context) error {
	return f.record(Call{Operation: OperationResume})
}

func (f *Fake) Next(_ context.Context) error {
	return f.record(Call{Operation: OperationNext})
}

func (f *Fake) Previous(_ context.Context) error {
	return f.record(Call{Operation: OperationPrevious})
}

func (f *Fake) SetVolume(_ context.Context, volume int) error {
	return f.record(Call{Operation: OperationSetVolume, Volume: volume})
}

func (f *Fake) SetAutoplay(_ context.Context, enabled bool) error {
	if err := f.record(Call{Operation: OperationSetAutoplay, Enabled: enabled}); err != nil {
		return err
	}
	f.mu.Lock()
	f.autoplay = enabled
	f.mu.Unlock()
	return nil
}

func (f *Fake) Close(_ context.Context) error {
	return f.record(Call{Operation: OperationClose})
}

func (f *Fake) SetSearchResult(page SearchPage, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.searchPage = page
	f.searchError = err
}

func (f *Fake) SetError(operation Operation, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.errors[operation] = err
}

func (f *Fake) record(call Call) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, call)
	return f.errors[call.Operation]
}

func (f *Fake) Calls() []Call {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]Call(nil), f.calls...)
}
