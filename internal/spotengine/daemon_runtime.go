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
	server := newMemoryAPIServer()
	store := newFileStateStore(filepath.Join(configDir, "session.json"))
	state, err := store.Load()
	if err != nil {
		return nil, err
	}
	hasSession := state != nil && len(state.Credentials.Data) > 0

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
		DisableAutoplay: false,
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
		return nil, err
	}
	adapter := newAdapter(app, server)
	adapter.hasSession = hasSession
	return adapter, nil
}
