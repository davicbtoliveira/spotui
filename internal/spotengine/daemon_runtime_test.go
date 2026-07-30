package spotengine

import (
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestAudioOutputUsesPulseAudioWhenWSLgSocketIsAvailable(t *testing.T) {
	backend, runtimeSocket := audioOutput("linux", "", true)

	if backend != "pulseaudio" {
		t.Fatalf("backend = %q, want pulseaudio", backend)
	}
	if runtimeSocket != "unix:/mnt/wslg/PulseServer" {
		t.Fatalf("runtime socket = %q, want unix:/mnt/wslg/PulseServer", runtimeSocket)
	}
}

func TestAudioOutputUsesConfiguredPulseAudioServer(t *testing.T) {
	backend, runtimeSocket := audioOutput("linux", "unix:/run/user/1000/pulse/native", false)

	if backend != "pulseaudio" {
		t.Fatalf("backend = %q, want pulseaudio", backend)
	}
	if runtimeSocket != "unix:/run/user/1000/pulse/native" {
		t.Fatalf("runtime socket = %q, want configured PulseAudio server", runtimeSocket)
	}
}

func TestAudioOutputFallsBackToALSAOnLinux(t *testing.T) {
	backend, runtimeSocket := audioOutput("linux", "", false)

	if backend != "alsa" {
		t.Fatalf("backend = %q, want alsa", backend)
	}
	if runtimeSocket != "" {
		t.Fatalf("runtime socket = %q, want empty", runtimeSocket)
	}
}

func TestAudioOutputUsesUserPulseAudioSocket(t *testing.T) {
	backend, runtimeSocket := audioOutputWithUserPulseSocket("linux", "", false, "/run/user/1000/pulse/native")

	if backend != "pulseaudio" {
		t.Fatalf("backend = %q, want pulseaudio", backend)
	}
	if runtimeSocket != "unix:/run/user/1000/pulse/native" {
		t.Fatalf("runtime socket = %q, want user PulseAudio socket", runtimeSocket)
	}
}

func TestUserPulseAudioSocketChecksRuntimeSocket(t *testing.T) {
	runtimeDir := t.TempDir()
	pulseDir := filepath.Join(runtimeDir, "pulse")
	if err := os.Mkdir(pulseDir, 0o700); err != nil {
		t.Fatal(err)
	}
	socket := filepath.Join(pulseDir, "native")
	listener, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	if got := userPulseAudioSocket("linux", runtimeDir); got != socket {
		t.Fatalf("socket = %q, want %q", got, socket)
	}
	if got := userPulseAudioSocket("darwin", runtimeDir); got != "" {
		t.Fatalf("darwin socket = %q, want empty", got)
	}
}

func TestAudioOutputUsesAudioToolboxOnMacOS(t *testing.T) {
	backend, runtimeSocket := audioOutput("darwin", "", false)

	if backend != "audio-toolbox" {
		t.Fatalf("backend = %q, want audio-toolbox", backend)
	}
	if runtimeSocket != "" {
		t.Fatalf("runtime socket = %q, want empty", runtimeSocket)
	}
}
