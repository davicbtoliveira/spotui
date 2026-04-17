package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const redirectURI = "http://localhost:8080/callback"

var scopes = []string{
	spotifyauth.ScopeUserReadPrivate,
	spotifyauth.ScopeUserReadEmail,
	spotifyauth.ScopeUserReadPlaybackState,
	spotifyauth.ScopeUserModifyPlaybackState,
	spotifyauth.ScopeUserReadCurrentlyPlaying,
	spotifyauth.ScopeUserLibraryRead,
	spotifyauth.ScopePlaylistReadPrivate,
	spotifyauth.ScopePlaylistReadCollaborative,
	spotifyauth.ScopeUserTopRead,
	spotifyauth.ScopeUserReadRecentlyPlayed,
}

func tokenPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "spotui", "token.json")
}

func generateVerifier() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	_ = cmd.Start()
}

func waitForCallback(expectedState string) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":8080", Handler: mux}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != expectedState {
			errCh <- fmt.Errorf("state mismatch")
			http.Error(w, "bad state", http.StatusBadRequest)
			return
		}
		if e := q.Get("error"); e != "" {
			errCh <- fmt.Errorf("auth denied: %s", e)
			fmt.Fprintf(w, "<html><body><p>Auth failed: %s. Close this tab.</p></body></html>", e)
			return
		}
		code := q.Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback")
			return
		}
		fmt.Fprint(w, `<html><body style="font-family:sans-serif;text-align:center;padding-top:60px"><h2>&#10003; Logged in to SpotUI</h2><p>You can close this tab.</p></body></html>`)
		codeCh <- code
	})

	go func() { _ = srv.ListenAndServe() }()

	defer func() { _ = srv.Shutdown(context.Background()) }()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-time.After(3 * time.Minute):
		return "", fmt.Errorf("auth timed out after 3 minutes")
	}
}

func saveToken(tok *oauth2.Token) {
	path := tokenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}
	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(tok)
}

func loadToken() (*oauth2.Token, error) {
	f, err := os.Open(tokenPath())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tok oauth2.Token
	if err := json.NewDecoder(f).Decode(&tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

func newAuthenticator(clientID string) *spotifyauth.Authenticator {
	return spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(scopes...),
	)
}

func Authenticate(clientID string) (*spotify.Client, error) {
	auth := newAuthenticator(clientID)

	if tok, err := loadToken(); err == nil && tok.Valid() {
		httpClient := auth.Client(context.Background(), tok)
		return spotify.New(httpClient, spotify.WithRetry(true)), nil
	}

	verifier, err := generateVerifier()
	if err != nil {
		return nil, fmt.Errorf("generate verifier: %w", err)
	}
	challenge := generateChallenge(verifier)

	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	authURL := auth.AuthURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", challenge),
	)

	fmt.Printf("\nOpening Spotify login in browser...\nIf nothing opens, visit:\n%s\n\n", authURL)
	openBrowser(authURL)

	code, err := waitForCallback(state)
	if err != nil {
		return nil, fmt.Errorf("callback: %w", err)
	}

	tok, err := auth.Exchange(context.Background(), code,
		oauth2.SetAuthURLParam("code_verifier", verifier),
	)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	saveToken(tok)

	httpClient := auth.Client(context.Background(), tok)
	return spotify.New(httpClient, spotify.WithRetry(true)), nil
}

func OpenSpotifySettings() {
	openBrowser("https://www.spotify.com/account/overview/")
}
