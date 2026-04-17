package clientid

import (
	"fmt"
	"os"
)

var Value = ""

func Resolve() (string, error) {
	if Value != "" {
		return Value, nil
	}
	if env := os.Getenv("SPOTIFY_CLIENT_ID"); env != "" {
		return env, nil
	}
	return "", fmt.Errorf(
		"no Spotify client ID configured\n" +
			"  development: export SPOTIFY_CLIENT_ID=<your_id>\n" +
			"  release build: -ldflags \"-X github.com/dcbto/spotui/internal/clientid.Value=<your_id>\"",
	)
}
