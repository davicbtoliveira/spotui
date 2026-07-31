package spclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	librespot "github.com/devgianlu/go-librespot"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestPathfinderQueryUsesAuthenticatedPersistedQuery(t *testing.T) {
	client, err := NewSpclient(context.Background(), &librespot.NullLogger{}, &http.Client{Transport: roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.Host != "api-partner.spotify.com" || request.URL.Path != "/pathfinder/v1/query" {
			t.Fatalf("request: %s %s", request.Method, request.URL)
		}
		if request.Header.Get("Authorization") != "Bearer access-token" || request.Header.Get("Client-Token") != "client-token" {
			t.Fatalf("authorization headers: %#v", request.Header)
		}
		if request.URL.Query().Get("operationName") != "queryArtistPlaylists" || !strings.Contains(request.URL.Query().Get("variables"), "spotify:artist:arctic") || !strings.Contains(request.URL.Query().Get("extensions"), "persistedQuery") {
			t.Fatalf("query: %s", request.URL.RawQuery)
		}
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":{"artistUnion":{"profile":{"playlistsV2":{"items":[]}}}}}`))}, nil
	})}, func(context.Context) string {
		return "spclient.wg.spotify.com"
	}, func(context.Context, bool) (string, error) {
		return "access-token", nil
	}, "device", "client-token")
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := client.PathfinderQuery(context.Background(), "queryArtistPlaylists", "hash", map[string]string{"uri": "spotify:artist:arctic"}, &result); err != nil {
		t.Fatal(err)
	}
	if result["artistUnion"] == nil {
		t.Fatalf("result: %#v", result)
	}
}
