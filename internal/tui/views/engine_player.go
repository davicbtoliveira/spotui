package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/spotengine"
	"github.com/dcbto/spotui/internal/theme"
)

type EnginePlayerState struct {
	Track       *spotengine.Track
	ProgressMS  int
	Playing     bool
	Buffering   bool
	Active      bool
	Volume      int
	Autoplay    bool
	Transferred bool
}

func RenderEnginePlayer(width int, state EnginePlayerState) string {
	if state.Transferred && state.Track == nil {
		content := theme.SubtextStyle.Render("  Transferred Playback — active on another device")
		return theme.PlayerBarStyle.Copy().Width(width).Render(content)
	}
	if state.Track == nil {
		content := theme.SubtextStyle.Render("  Ready — press / to search tracks")
		return theme.PlayerBarStyle.Copy().Width(width).Render(content)
	}

	status := "Paused"
	switch {
	case state.Transferred:
		status = "Transferred Playback"
	case state.Buffering:
		status = "Buffering"
	case state.Playing:
		status = "Playing"
	}

	track := state.Track
	left := "  " + theme.TrackNameStyle.Render(truncate(track.Name, width/3)) +
		theme.SubtextStyle.Render(" · ") +
		theme.ArtistNameStyle.Render(truncate(track.Artist, width/3))
	autoplay := "Off"
	if state.Autoplay {
		autoplay = "On"
	}
	right := theme.SubtextStyle.Render(
		fmt.Sprintf("%s · Vol %d%% · Autoplay %s", status, state.Volume, autoplay),
	)
	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}

	timeText := theme.SubtextStyle.Render(
		fmt.Sprintf("%s/%s", formatDuration(state.ProgressMS), formatDuration(track.DurationMS)),
	)
	barWidth := width - lipgloss.Width(timeText) - 6
	if barWidth < 5 {
		barWidth = 5
	}
	progress := "  " + renderProgress(barWidth, state.ProgressMS, track.DurationMS) + "  " + timeText

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		left+strings.Repeat(" ", gap)+right,
		progress,
	)
	return theme.PlayerBarStyle.Copy().Width(width).Render(content)
}
