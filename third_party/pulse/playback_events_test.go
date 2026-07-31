package pulse

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/jfreymuth/pulse/proto"
)

func TestPlaybackEventTimeoutDoesNotPanic(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	t.Cleanup(func() {
		_ = clientConn.Close()
		_ = serverConn.Close()
	})
	go func() { _, _ = io.Copy(io.Discard, serverConn) }()

	protocolClient := &proto.Client{}
	protocolClient.Open(clientConn)
	protocolClient.SetTimeout(5 * time.Millisecond)

	stream := &PlaybackStream{
		c:             &Client{c: protocolClient},
		state:         newStateMachine(),
		volumeChanges: make(chan proto.ChannelVolumes, 1),
		createRequest: proto.CreatePlaybackStream{
			ChannelMap: proto.ChannelMap{proto.ChannelMono},
		},
	}
	events := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		stream.handleEvents(events)
		close(done)
	}()

	events <- struct{}{}
	close(events)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("event handler did not return after the request timed out")
	}
}
