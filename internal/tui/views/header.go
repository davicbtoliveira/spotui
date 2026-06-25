package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/library"
	"github.com/dcbto/spotui/internal/theme"
)

func RenderTopBar(width int, username string, activeTab library.Tab) string {
	title := theme.TopBarTitle.Render("SPOTUI")
	userStr := theme.TopBarUser.Render(username)

	tabLabels := []string{"Playlists", "Tracks", "Artists", "Search"}
	tabParts := make([]string, len(tabLabels))
	for i, label := range tabLabels {
		sp := theme.DividerStyle.Render(" ")
		num := fmt.Sprintf("%d", i+1)
		numStr := theme.DividerStyle.Render(num)
		var nameStr string
		if library.Tab(i) == activeTab {
			nameStr = theme.ActiveTabStyle.Render(label)
		} else {
			nameStr = theme.InactiveTabStyle.Render(label)
		}
		tabParts[i] = sp + numStr + " " + nameStr
	}
	tabsStr := strings.Join(tabParts, "")

	left := title + "  " + userStr
	right := tabsStr

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := width - 2 - leftW - rightW
	if gap < 1 {
		gap = 1
	}

	return theme.TopBarStyle.Copy().Width(width).Render(
		left + strings.Repeat(" ", gap) + right,
	)
}
