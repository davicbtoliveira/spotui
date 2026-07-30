package views

import (
	"fmt"
	"strings"
	"time"

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
	Shuffle     bool
	Transferred bool
}

func renderProgress(width, progress, total int) string {
	if total <= 0 {
		return theme.ProgressEmptyStyle.Render(strings.Repeat("░", width))
	}
	if progress > total {
		progress = total
	}
	filled := int(float64(width) * float64(progress) / float64(total))
	empty := width - filled
	if empty < 0 {
		empty = 0
	}
	return theme.ProgressStyle.Render(strings.Repeat("█", filled)) +
		theme.ProgressEmptyStyle.Render(strings.Repeat("░", empty))
}

func formatDuration(ms int) string {
	duration := time.Duration(ms) * time.Millisecond
	return fmt.Sprintf("%02d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)
}

func truncate(value string, maxLength int) string {
	if maxLength < 4 {
		maxLength = 4
	}
	runes := []rune(value)
	if len(runes) <= maxLength {
		return value
	}
	return string(runes[:maxLength-1]) + "…"
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
		theme.ArtistNameStyle.Render(truncate(track.Artist, width/4)) +
		theme.SubtextStyle.Render(" · ") +
		theme.SubtextStyle.Render(truncate(track.Album, width/4))
	autoplay := "Off"
	if state.Autoplay {
		autoplay = "On"
	}
	shuffle := "Off"
	if state.Shuffle {
		shuffle = "On"
	}
	right := theme.SubtextStyle.Render(
		fmt.Sprintf("%s · Vol %d%% · Autoplay %s · Shuffle %s", status, state.Volume, autoplay, shuffle),
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
