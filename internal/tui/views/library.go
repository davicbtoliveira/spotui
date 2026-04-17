package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/theme"
	"github.com/zmb3/spotify/v2"
)

type LibraryTab int

const (
	TabPlaylists LibraryTab = iota
	TabTracks
	TabArtists
)

func RenderTabBar(width int, active LibraryTab) string {
	labels := []string{"Playlists", "Tracks", "Artists"}
	rendered := make([]string, len(labels))
	for i, l := range labels {
		if LibraryTab(i) == active {
			rendered[i] = theme.ActiveTabStyle.Render(l)
		} else {
			rendered[i] = theme.InactiveTabStyle.Render(l)
		}
	}
	bar := "  " + strings.Join(rendered, "  ")
	divider := theme.DividerStyle.Render(strings.Repeat("─", width))
	return bar + "\n" + divider
}

func RenderPlaylists(width, height, cursor int, playlists []spotify.SimplePlaylist) string {
	if len(playlists) == 0 {
		return theme.SubtextStyle.Render("  Your library is empty")
	}

	offset := scrollOffset(cursor, height)
	lines := make([]string, 0, height)

	for i := offset; i < len(playlists) && i < offset+height; i++ {
		pl := playlists[i]

		prefix := "  "
		nameStyle := theme.NormalItemStyle
		if i == cursor {
			prefix = "▶ "
			nameStyle = theme.SelectedItemStyle
		}

		countStr := fmt.Sprintf("%d tracks", pl.Tracks.Total)
		name := nameStyle.Render(truncate(pl.Name, width-len(countStr)-6))
		count := theme.SubtextStyle.Render(countStr)

		nameW := lipgloss.Width(name)
		countW := lipgloss.Width(count)
		gap := width - nameW - countW - len(prefix) - 2
		if gap < 1 {
			gap = 1
		}

		lines = append(lines, prefix+name+strings.Repeat(" ", gap)+count)
	}

	return strings.Join(lines, "\n")
}

func RenderTracks(width, height, cursor int, tracks []spotify.SavedTrack) string {
	if len(tracks) == 0 {
		return theme.SubtextStyle.Render("  No saved tracks yet")
	}

	offset := scrollOffset(cursor, height)
	lines := make([]string, 0, height)

	for i := offset; i < len(tracks) && i < offset+height; i++ {
		t := tracks[i].FullTrack

		artists := make([]string, len(t.Artists))
		for j, a := range t.Artists {
			artists[j] = a.Name
		}
		artistJoined := strings.Join(artists, ", ")

		prefix := "  "
		nameStyle := theme.NormalItemStyle
		artStyle := theme.SubtextStyle
		if i == cursor {
			prefix = "▶ "
			nameStyle = theme.SelectedItemStyle
			artStyle = theme.ArtistNameStyle
		}

		dur := formatDuration(int(t.Duration))
		maxNameW := width/2 - 4
		if maxNameW < 8 {
			maxNameW = 8
		}

		trackName := nameStyle.Render(truncate(t.Name, maxNameW))
		artistStr := artStyle.Render(truncate(artistJoined, width/3))
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

func RenderArtists(width, height, cursor int, artists []spotify.FullArtist) string {
	if len(artists) == 0 {
		return theme.SubtextStyle.Render("  No top artists found")
	}

	offset := scrollOffset(cursor, height)
	lines := make([]string, 0, height)

	for i := offset; i < len(artists) && i < offset+height; i++ {
		a := artists[i]

		prefix := "  "
		nameStyle := theme.NormalItemStyle
		if i == cursor {
			prefix = "▶ "
			nameStyle = theme.SelectedItemStyle
		}

		genres := a.Genres
		if len(genres) > 2 {
			genres = genres[:2]
		}
		genreStr := strings.Join(genres, ", ")

		name := nameStyle.Render(truncate(a.Name, width/2))
		genreRendered := theme.SubtextStyle.Render(truncate(genreStr, width/3))

		nameW := lipgloss.Width(name)
		genreW := lipgloss.Width(genreRendered)
		gap := width - nameW - genreW - len(prefix) - 2
		if gap < 1 {
			gap = 1
		}

		lines = append(lines, prefix+name+strings.Repeat(" ", gap)+genreRendered)
	}

	return strings.Join(lines, "\n")
}

func RenderHelpOverlay(width, height int) string {
	help := strings.Join([]string{
		"",
		"  Keyboard Shortcuts",
		"  " + strings.Repeat("─", 30),
		"",
		"  1 / 2 / 3    Switch tabs (Playlists / Tracks / Artists)",
		"  j / ↓        Move cursor down",
		"  k / ↑        Move cursor up",
		"  enter        Play selected item",
		"  space        Toggle play / pause",
		"  n            Next track",
		"  p            Previous track",
		"  s            Toggle shuffle",
		"  c            Open Spotify account settings",
		"  ?            Toggle this help",
		"  q / ctrl+c   Quit",
		"",
	}, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorGreen).
		Padding(1, 2).
		Foreground(theme.ColorWhite).
		Render(help)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func scrollOffset(cursor, height int) int {
	if cursor < height {
		return 0
	}
	return cursor - height + 1
}
