package daemon

import (
	"context"
	"testing"

	"github.com/devgianlu/go-librespot/ap"
)

type captureApiServer struct {
	events chan *ApiEvent
}

func (s *captureApiServer) Emit(event *ApiEvent) {
	s.events <- event
}

func (*captureApiServer) Receive() <-chan ApiRequest {
	return make(<-chan ApiRequest)
}

func (*captureApiServer) Close() error {
	return nil
}

func TestSetAutoplayRequestChangesRunningPlayer(t *testing.T) {
	cfg := &Config{DisableAutoplay: true}
	p := &AppPlayer{app: &App{cfg: cfg}}

	_, err := p.handleApiRequest(context.Background(), ApiRequest{
		Type: ApiRequestTypeSetAutoplay,
		Data: true,
	})
	if err != nil {
		t.Fatalf("enable autoplay: %v", err)
	}
	if cfg.DisableAutoplay {
		t.Fatal("autoplay remains disabled")
	}

	_, err = p.handleApiRequest(context.Background(), ApiRequest{
		Type: ApiRequestTypeSetAutoplay,
		Data: false,
	})
	if err != nil {
		t.Fatalf("disable autoplay: %v", err)
	}
	if !cfg.DisableAutoplay {
		t.Fatal("autoplay remains enabled")
	}
}

func TestProductInfoEmitsAccountProductBeforePlaybackReady(t *testing.T) {
	server := &captureApiServer{events: make(chan *ApiEvent, 2)}
	player := &AppPlayer{
		app:             &App{server: server},
		playbackReadyCh: make(chan struct{}),
		countryCode:     new(string),
	}

	err := player.handleAccesspointPacket(ap.PacketTypeProductInfo, []byte(
		`<products><product><type>premium</type></product></products>`,
	))
	if err != nil {
		t.Fatalf("handle product info: %v", err)
	}

	event := <-server.events
	if event.Type != ApiEventTypeAccountProduct {
		t.Fatalf("event type: want %q, got %q", ApiEventTypeAccountProduct, event.Type)
	}
	data, ok := event.Data.(ApiEventDataAccountProduct)
	if !ok || data.Product != "premium" {
		t.Fatalf("event data: %#v", event.Data)
	}
	select {
	case extra := <-server.events:
		t.Fatalf("unexpected event before player readiness: %#v", extra)
	default:
	}
}
