package session

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	librespot "github.com/devgianlu/go-librespot"
)

func TestOAuth2ServerReportsDeniedCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, results, err := NewOAuth2Server(ctx, &librespot.NullLogger{}, 0)
	if err != nil {
		t.Fatalf("start OAuth callback: %v", err)
	}

	client := &http.Client{Timeout: time.Second}
	response, err := client.Get(
		"http://127.0.0.1:" + strconv.Itoa(port) + "/login?error=access_denied",
	)
	if err != nil {
		t.Fatalf("send denied callback: %v", err)
	}
	_ = response.Body.Close()

	result := <-results
	if result.Err == nil || !strings.Contains(result.Err.Error(), "authorization denied") {
		t.Fatalf("callback error: %v", result.Err)
	}
	if result.Code != "" {
		t.Fatalf("callback code: want empty, got %q", result.Code)
	}
}

func TestOAuth2ServerReportsMalformedCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, results, err := NewOAuth2Server(ctx, &librespot.NullLogger{}, 0)
	if err != nil {
		t.Fatalf("start OAuth callback: %v", err)
	}

	client := &http.Client{Timeout: time.Second}
	response, err := client.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/login")
	if err != nil {
		t.Fatalf("send malformed callback: %v", err)
	}
	_ = response.Body.Close()

	result := <-results
	if result.Err == nil || !strings.Contains(result.Err.Error(), "missing authorization code") {
		t.Fatalf("callback error: %v", result.Err)
	}
}

func TestOAuth2ServerCancellationClosesListener(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	port, results, err := NewOAuth2Server(ctx, &librespot.NullLogger{}, 0)
	if err != nil {
		t.Fatalf("start OAuth callback: %v", err)
	}

	cancel()
	result := <-results
	if !errors.Is(result.Err, context.Canceled) {
		t.Fatalf("callback error: want context canceled, got %v", result.Err)
	}

	address := "127.0.0.1:" + strconv.Itoa(port)
	deadline := time.Now().Add(time.Second)
	for {
		connection, dialErr := net.DialTimeout("tcp", address, 20*time.Millisecond)
		if dialErr != nil {
			break
		}
		_ = connection.Close()
		if time.Now().After(deadline) {
			t.Fatalf("callback listener remains open at %s", address)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
