package player_test

import (
	"testing"

	"github.com/dcbto/spotui/internal/player"
)

func TestProgressMsIsZeroInitially(t *testing.T) {
	p := player.New()
	if p.ProgressMs() != 0 {
		t.Fatalf("want 0, got %d", p.ProgressMs())
	}
}

func TestSetNowPlayingSetsProgress(t *testing.T) {
	p := player.New()
	p.SetNowPlaying(player.NowPlayingEntry{ProgressMs: 30000})
	if p.ProgressMs() != 30000 {
		t.Fatalf("want 30000, got %d", p.ProgressMs())
	}
}

func TestTickAdvancesProgressWhenPlaying(t *testing.T) {
	p := player.New()
	p.SetNowPlaying(player.NowPlayingEntry{ProgressMs: 0, Playing: true, Duration: 200000})
	p.Tick()
	if p.ProgressMs() != 1000 {
		t.Fatalf("want 1000, got %d", p.ProgressMs())
	}
}

func TestTickDoesNothingWhenNotPlaying(t *testing.T) {
	p := player.New()
	p.SetNowPlaying(player.NowPlayingEntry{ProgressMs: 5000, Playing: false})
	p.Tick()
	if p.ProgressMs() != 5000 {
		t.Fatalf("want 5000, got %d", p.ProgressMs())
	}
}

func TestTickClampsAtDuration(t *testing.T) {
	p := player.New()
	p.SetNowPlaying(player.NowPlayingEntry{ProgressMs: 199500, Duration: 200000, Playing: true})
	p.Tick()
	if p.ProgressMs() != 200000 {
		t.Fatalf("want clamp at 200000, got %d", p.ProgressMs())
	}
}
