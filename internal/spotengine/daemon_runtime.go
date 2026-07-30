package spotengine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	librespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/daemon"
)

func NewAdapter() (*Adapter, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user config directory: %w", err)
	}
	return newAdapterAtDir(filepath.Join(userConfigDir, "spotui"))
}

func newAdapterAtDir(configDir string) (*Adapter, error) {
	store := newFileStateStore(filepath.Join(configDir, "session.json"))
	preferences := newPreferenceStore(filepath.Join(configDir, "settings.json"))
	autoplay, err := preferences.LoadAutoplay()
	if err != nil {
		return nil, err
	}
	state, err := store.Load()
	if err != nil {
		if clearErr := store.Clear(); clearErr != nil {
			return nil, fmt.Errorf("recover local session: %w", clearErr)
		}
		state = nil
	}
	hasSession := state != nil && len(state.Credentials.Data) > 0

	events := make(chan Event, 64)
	factory := func() (engineRuntime, *memoryAPIServer, error) {
		autoplay, err := preferences.LoadAutoplay()
		if err != nil {
			return nil, nil, err
		}
		server := newMemoryAPIServerWithEvents(events)
		audioBackend := "alsa"
		if runtime.GOOS == "darwin" {
			audioBackend = "audio-toolbox"
		}

		config := &daemon.Config{
			DeviceName:      "SpotUI",
			DeviceType:      "computer",
			AudioBackend:    audioBackend,
			AudioDevice:     "default",
			Bitrate:         160,
			VolumeSteps:     100,
			InitialVolume:   100,
			ZeroconfEnabled: true,
			ZeroconfBackend: "builtin",
			ImageSize:       "default",
			DisableAutoplay: !autoplay,
			Credentials: daemon.CredentialsConfig{
				Type: "interactive",
				Interactive: daemon.InteractiveCredentials{
					CallbackPort: 0,
				},
			},
		}
		config.Cache.Enabled = false

		app, err := daemon.New(&daemon.Options{
			Logger:     &librespot.NullLogger{},
			Config:     config,
			StateStore: store,
			APIServer:  server,
			OnAuthURL: func(url string) {
				server.emit(Event{Type: EventTypeAuthorizationURL, URL: url})
			},
		})
		if err != nil {
			return nil, nil, err
		}
		return app, server, nil
	}
	app, server, err := factory()
	if err != nil {
		return nil, err
	}
	adapter := newAdapter(app, server)
	adapter.factory = factory
	adapter.clearState = store.Clear
	adapter.saveAutoplay = preferences.SaveAutoplay
	adapter.sessionAvailable = func() bool {
		state, err := store.Load()
		return err == nil && state != nil && len(state.Credentials.Data) > 0
	}
	adapter.hasSession = hasSession
	adapter.autoplay = autoplay
	return adapter, nil
}
