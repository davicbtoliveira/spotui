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
		"  /            Search tracks",
		"  [ / ]        Previous / next search page",
		"  esc          Cancel search input",
		"  j / ↓        Move cursor down",
		"  k / ↑        Move cursor up",
		"  enter        Play selected item",
		"  space        Toggle play / pause",
		"  n            Next track",
		"  p            Previous track",
		"  - / +        Lower / raise player volume",
		"  a            Toggle Autoplay",
		"  L            Log out",
		"  ?            Toggle this help",
		"  q / ctrl+c   Quit",
		"",
		"  Login requires Spotify Premium.",
		"  First Login: Linux and macOS local terminal.",
		"  Uses an unofficial protocol; service changes may break it.",
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
