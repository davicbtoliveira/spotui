package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	librespot "github.com/devgianlu/go-librespot"
)

type OAuth2Result struct {
	Code string
	Err  error
}

func NewOAuth2Server(ctx context.Context, _ librespot.Logger, callbackPort int) (int, <-chan OAuth2Result, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", callbackPort))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to listen: %w", err)
	}

	results := make(chan OAuth2Result, 1)
	var finishOnce sync.Once
	finish := func(result OAuth2Result) {
		finishOnce.Do(func() {
			results <- result
			close(results)
			_ = listener.Close()
		})
	}

	server := &http.Server{Handler: http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		if query.Get("error") != "" {
			http.Error(response, "Spotify authorization denied", http.StatusBadRequest)
			finish(OAuth2Result{Err: errors.New("authorization denied")})
			return
		}

		code := query.Get("code")
		if code == "" {
			http.Error(response, "Spotify callback missing code", http.StatusBadRequest)
			finish(OAuth2Result{Err: errors.New("callback missing authorization code")})
			return
		}

		_, _ = response.Write([]byte("Go back to go-librespot!"))
		finish(OAuth2Result{Code: code})
	})}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) &&
			!errors.Is(err, net.ErrClosed) {
			finish(OAuth2Result{Err: fmt.Errorf("callback server failed: %w", err)})
		}
	}()

	go func() {
		<-ctx.Done()
		finish(OAuth2Result{Err: ctx.Err()})
	}()

	return listener.Addr().(*net.TCPAddr).Port, results, nil
}
