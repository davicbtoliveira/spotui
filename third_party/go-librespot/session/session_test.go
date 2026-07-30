package session

import "testing"

func TestNotifyAuthURLCallsHook(t *testing.T) {
	const want = "https://accounts.spotify.com/authorize?code_challenge=test"

	var got string
	notifyAuthURL(func(url string) {
		got = url
	}, want)

	if got != want {
		t.Fatalf("authorization URL: want %q, got %q", want, got)
	}
}
