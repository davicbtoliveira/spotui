package spotengine

import (
	"runtime"

	librespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/daemon"
)

func NewAdapter() (*Adapter, error) {
	server := newMemoryAPIServer()

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
		StateStore: newMemoryStateStore(),
		APIServer:  server,
		OnAuthURL: func(url string) {
			server.emit(Event{Type: EventTypeAuthorizationURL, URL: url})
		},
	})
	if err != nil {
		return nil, err
	}
	return newAdapter(app, server), nil
}
