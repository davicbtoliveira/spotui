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
	if state == nil || state.Item == nil {
		msg := theme.SubtextStyle.Render("  No active device — open Spotify on a device first")
		pad := width - lipgloss.Width(msg) - 2
		if pad < 0 {
			pad = 0
		}
		content := msg + strings.Repeat(" ", pad)
		return theme.PlayerBarStyle.Copy().Width(width).Render(content)
	}

	track := state.Item
	artists := make([]string, len(track.Artists))
	for i, a := range track.Artists {
		artists[i] = a.Name
	}

	artistStr := strings.Join(artists, ", ")

	playPause := "▶"
	if state.Playing {
		playPause = "⏸"
	}

	shuffleLabel := theme.ShuffleOffStyle.Render("  🔀")
	if shuffleOn {
		shuffleLabel = theme.ShuffleOnStyle.Render("  🔀")
	}

	controls := fmt.Sprintf("  %s  %s  %s", theme.SubtextStyle.Render("⏮"), theme.TrackNameStyle.Render(playPause), theme.SubtextStyle.Render("⏭"))

	totalDur := formatDuration(int(track.Duration))

	progressMs := localProgressMs
	if progressMs == 0 {
		progressMs = int(state.Progress)
	}
	elapsed := formatDuration(progressMs)
	timeStr := theme.SubtextStyle.Render(elapsed + "/" + totalDur)
	timeWidth := lipgloss.Width(timeStr)

	progressBarWidth := width - timeWidth - 6
	if progressBarWidth < 5 {
		progressBarWidth = 5
	}
	bar := renderProgress(progressBarWidth, progressMs, int(track.Duration))
	progressLine := "  " + bar + "  " + timeStr

	trackInfo := truncate(track.Name, width/3)
	artistInfo := truncate(artistStr, width/3)

	leftLabel := theme.TrackNameStyle.Render(trackInfo)
	metaLabel := theme.ArtistNameStyle.Render(artistInfo)
	left := "  " + leftLabel + theme.SubtextStyle.Render(" · ") + metaLabel

	right := controls + shuffleLabel

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := width - 2 - leftW - rightW
	if gap < 1 {
		gap = 1
	}

	line1 := left + strings.Repeat(" ", gap) + right

	content := lipgloss.JoinVertical(lipgloss.Left, line1, progressLine)
	return theme.PlayerBarStyle.Copy().Width(width).Render(content)
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
