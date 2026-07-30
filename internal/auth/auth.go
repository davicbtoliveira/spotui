package auth

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenURL(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("opening URLs is unsupported on %s", runtime.GOOS)
	}
	return cmd.Start()
}

func OpenSpotifySettings() {
	_ = OpenURL("https://www.spotify.com/account/overview/")
}
