package library

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/theme"
)

func (l *Library) View(width, height int) string {
	if len(l.search) > 0 {
		return renderTracks(width, height, l.cursor, l.search)
	}
	return renderSearchEmpty()
}

func renderSearchEmpty() string {
	return theme.SubtextStyle.Render("  Press / to search tracks")
}

func renderTracks(width, height, cursor int, tracks []TrackEntry) string {
	if len(tracks) == 0 {
		return theme.SubtextStyle.Render("  No saved tracks yet")
	}

	offset := scrollOffset(cursor, height)
	lines := make([]string, 0, height)

	for i := offset; i < len(tracks) && i < offset+height; i++ {
		t := tracks[i]

		prefix := "  "
		nameStyle := theme.NormalItemStyle
		artStyle := theme.SubtextStyle
		if i == cursor {
			prefix = "▸ "
			nameStyle = theme.SelectedItemStyle
			artStyle = theme.ArtistNameStyle
		}

		dur := formatDuration(t.Duration)
		maxNameW := width/2 - 4
		if maxNameW < 8 {
			maxNameW = 8
		}

		trackName := nameStyle.Render(truncate(t.Name, maxNameW))
		artistStr := artStyle.Render(truncate(t.Artist, width/3))
		durStr := theme.SubtextStyle.Render(dur)

		nameW := lipgloss.Width(trackName) + lipgloss.Width(artistStr) + 3
		gap := width - nameW - lipgloss.Width(durStr) - len(prefix) - 2
		if gap < 1 {
			gap = 1
		}

		line := prefix + trackName + theme.SubtextStyle.Render(" · ") + artistStr +
			strings.Repeat(" ", gap) + durStr
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
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

func formatDuration(ms int) string {
	mins := ms / 60000
	secs := (ms % 60000) / 1000
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func scrollOffset(cursor, height int) int {
	if cursor < height {
		return 0
	}
	return cursor - height + 1
}
