package spotengine

import (
	"sync"

	librespot "github.com/devgianlu/go-librespot"
)

type memoryStateStore struct {
	mu    sync.Mutex
	state *librespot.AppState
}

func newMemoryStateStore() *memoryStateStore {
	return &memoryStateStore{}
}

func (s *memoryStateStore) Load() (*librespot.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return cloneAppState(s.state), nil
}

func (s *memoryStateStore) Save(state *librespot.AppState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = cloneAppState(state)
	return nil
}

func cloneAppState(state *librespot.AppState) *librespot.AppState {
	if state == nil {
		return nil
	}

	state.Lock()
	defer state.Unlock()

	cloned := &librespot.AppState{
		DeviceId:     state.DeviceId,
		EventManager: append([]byte(nil), state.EventManager...),
	}
	cloned.Credentials.Username = state.Credentials.Username
	cloned.Credentials.Data = append([]byte(nil), state.Credentials.Data...)
	if state.LastVolume != nil {
		volume := *state.LastVolume
		cloned.LastVolume = &volume
	}
	return cloned
}
