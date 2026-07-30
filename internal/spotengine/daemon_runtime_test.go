package spotengine

import "testing"

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

func TestAudioOutputUsesAudioToolboxOnMacOS(t *testing.T) {
	backend, runtimeSocket := audioOutput("darwin", "", false)

	if backend != "audio-toolbox" {
		t.Fatalf("backend = %q, want audio-toolbox", backend)
	}
	if runtimeSocket != "" {
		t.Fatalf("runtime socket = %q, want empty", runtimeSocket)
	}
}
