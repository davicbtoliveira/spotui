package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/theme"
	"github.com/zmb3/spotify/v2"
)

func RenderPlayer(width int, state *spotify.PlayerState, shuffleOn bool, localProgressMs int) string {
	divider := theme.DividerStyle.Render(strings.Repeat("─", width))

	if state == nil || state.Item == nil {
		content := lipgloss.JoinVertical(lipgloss.Left,
			divider,
			theme.SubtextStyle.Render("  No active device — open Spotify on a device first"),
			"",
			"  "+theme.SubtextStyle.Render("[⏮]  [▶]  [⏭]"),
		)
		return theme.PlayerBarStyle.Copy().Width(width).Render(content)
	}

	track := state.Item
	artists := make([]string, len(track.Artists))
	for i, a := range track.Artists {
		artists[i] = a.Name
	}

	trackName := theme.TrackNameStyle.Render(truncate(track.Name, width/2))
	artistStr := theme.ArtistNameStyle.Render(truncate(strings.Join(artists, ", "), width/3))
	albumStr := theme.SubtextStyle.Render(truncate(track.Album.Name, width/2))

	barWidth := width - 14
	if barWidth < 10 {
		barWidth = 10
	}
	progressMs := localProgressMs
	if progressMs == 0 {
		progressMs = int(state.Progress)
	}
	progressBar := renderProgress(barWidth, progressMs, int(track.Duration))
	elapsed := formatDuration(progressMs)
	total := formatDuration(int(track.Duration))
	timeLabel := theme.SubtextStyle.Render(elapsed + "/" + total)

	playPause := "▶"
	if state.Playing {
		playPause = "⏸"
	}
	controls := fmt.Sprintf("  [⏮]  [%s]  [⏭]", playPause)

	shuffleLabel := theme.ShuffleOffStyle.Render("🔀 OFF")
	if shuffleOn {
		shuffleLabel = theme.ShuffleOnStyle.Render("🔀 ON ")
	}

	line1 := "  " + trackName + theme.SubtextStyle.Render("  ·  ") + artistStr
	line2 := "  " + albumStr
	line3 := "  " + progressBar + "  " + timeLabel
	line4 := controls + "      " + shuffleLabel

	content := lipgloss.JoinVertical(lipgloss.Left,
		divider,
		line1,
		line2,
		line3,
		line4,
	)

	return theme.PlayerBarStyle.Copy().Width(width).Render(content)
}

func renderProgress(width, progress, total int) string {
	if total <= 0 {
		return theme.ProgressEmptyStyle.Render(strings.Repeat("─", width))
	}
	if progress > total {
		progress = total
	}
	filled := int(float64(width) * float64(progress) / float64(total))
	empty := width - filled - 1
	if empty < 0 {
		empty = 0
	}
	return theme.ProgressStyle.Render(strings.Repeat("─", filled)) +
		theme.ProgressStyle.Render("●") +
		theme.ProgressEmptyStyle.Render(strings.Repeat("─", empty))
}

func formatDuration(ms int) string {
	d := time.Duration(ms) * time.Millisecond
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func truncate(s string, maxLen int) string {
	if maxLen < 4 {
		maxLen = 4
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}
