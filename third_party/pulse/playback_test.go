package pulse

import "testing"

func TestEnsurePlaybackBuffersGrowsForServerRequest(t *testing.T) {
	front := make([]byte, 4080)
	back := make([]byte, 4080)

	front, back = ensurePlaybackBuffers(front, back, 4112)

	if cap(front) < 4112 || cap(back) < 4112 {
		t.Fatalf("buffer capacities = (%d, %d), want both at least 4112", cap(front), cap(back))
	}
}

func TestEnsurePlaybackBuffersRetainsSufficientBuffers(t *testing.T) {
	front := make([]byte, 4096)
	back := make([]byte, 4096)

	gotFront, gotBack := ensurePlaybackBuffers(front, back, 2048)

	if &gotFront[0] != &front[0] || &gotBack[0] != &back[0] {
		t.Fatal("sufficient buffers were unexpectedly replaced")
	}
}

func TestPlaybackStartedSignalSurvivesUnderflow(t *testing.T) {
	stream := &PlaybackStream{
		state:     newStateMachine(),
		underflow: true,
		started:   make(chan bool, 1),
	}
	stream.state.set(running)

	stream.notifyStarted()

	select {
	case <-stream.started:
	default:
		t.Fatal("started signal was discarded after underflow")
	}
}

func TestPrepareResumeMakesStreamReadableBeforeUncork(t *testing.T) {
	stream := &PlaybackStream{
		state:     newStateMachine(),
		underflow: true,
	}
	stream.state.set(paused)

	if !stream.prepareResume() {
		t.Fatal("paused stream was not prepared for resume")
	}
	if !stream.state.is(running) {
		t.Fatal("stream must accept buffer requests before it is uncorked")
	}
	if stream.underflow {
		t.Fatal("resume did not clear underflow")
	}
}

func TestConsumeStartedDoesNotWaitForMissingServerEvent(t *testing.T) {
	started := make(chan bool, 1)

	consumeStarted(started)
}
