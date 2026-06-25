package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/theme"
)

func RenderHelpOverlay(width, height int) string {
	help := strings.Join([]string{
		"",
		"  Keyboard Shortcuts",
		"  " + strings.Repeat("─", 30),
		"",
		"  1 / 2 / 3 / 4 Switch tabs (Playlists / Tracks / Artists / Search)",
		"  /            Search tracks",
		"  [ / ]        Previous / next search page",
		"  esc          Cancel search input",
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
		BorderForeground(theme.ColorAmber).
		Padding(1, 2).
		Foreground(theme.ColorText).
		Render(help)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
