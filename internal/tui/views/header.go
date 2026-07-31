package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/theme"
)

func RenderBrowseTopBar(width int, section string) string {
	title := theme.TopBarTitle.Render("SPOTUI")
	right := theme.ActiveTabStyle.Render(section)
	gap := width - 2 - lipgloss.Width(title) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return theme.TopBarStyle.Copy().Width(width).Render(title + strings.Repeat(" ", gap) + right)
}
