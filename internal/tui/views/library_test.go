package views

import (
	"strings"
	"testing"
)

func TestHelpDisclosesStandalonePlaybackConstraints(t *testing.T) {
	help := RenderHelpOverlay(100, 40)

	for _, expected := range []string{
		"Login",
		"Spotify Premium",
		"Linux and macOS",
		"local terminal",
		"unofficial protocol",
	} {
		if !strings.Contains(help, expected) {
			t.Errorf("help missing %q", expected)
		}
	}
}
