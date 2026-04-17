package commands

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dcbto/spotui/internal/msgs"
	"github.com/zmb3/spotify/v2"
)

func CmdClearStatus() tea.Cmd {
	return tea.Tick(4*time.Second, func(_ time.Time) tea.Msg {
		return msgs.ClearStatusMsg{}
	})
}

func CmdProgressTick() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return msgs.ProgressTickMsg{}
	})
}

func CmdNowPlaying(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		state, err := client.PlayerState(context.Background())
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "now playing"}
		}
		return msgs.NowPlayingMsg{State: state}
	}
}

func CmdTick() tea.Cmd {
	return tea.Tick(3*time.Second, func(_ time.Time) tea.Msg {
		return msgs.TickMsg{}
	})
}

func CmdPlayPause(client *spotify.Client, playing bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if playing {
			err = client.Pause(context.Background())
		} else {
			err = client.Play(context.Background())
		}
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "play/pause"}
		}
		return CmdNowPlaying(client)()
	}
}

func CmdNext(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		if err := client.Next(context.Background()); err != nil {
			return msgs.ErrMsg{Err: err, Context: "next track"}
		}
		return CmdNowPlaying(client)()
	}
}

func CmdPrevious(client *spotify.Client) tea.Cmd {
	return func() tea.Msg {
		if err := client.Previous(context.Background()); err != nil {
			return msgs.ErrMsg{Err: err, Context: "previous track"}
		}
		return CmdNowPlaying(client)()
	}
}

func CmdShuffle(client *spotify.Client, enable bool) tea.Cmd {
	return func() tea.Msg {
		if err := client.Shuffle(context.Background(), enable); err != nil {
			return msgs.ErrMsg{Err: err, Context: "shuffle"}
		}
		return CmdNowPlaying(client)()
	}
}

func CmdPlayPlaylist(client *spotify.Client, playlistURI spotify.URI) tea.Cmd {
	return func() tea.Msg {
		uri := playlistURI
		err := client.PlayOpt(context.Background(), &spotify.PlayOptions{
			PlaybackContext: &uri,
		})
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "play playlist"}
		}
		return CmdNowPlaying(client)()
	}
}

func CmdPlayTrack(client *spotify.Client, trackURI spotify.URI) tea.Cmd {
	return func() tea.Msg {
		err := client.PlayOpt(context.Background(), &spotify.PlayOptions{
			URIs: []spotify.URI{trackURI},
		})
		if err != nil {
			return msgs.ErrMsg{Err: err, Context: "play track"}
		}
		return CmdNowPlaying(client)()
	}
}
