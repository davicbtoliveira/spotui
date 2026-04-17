package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/theme"
)

func RenderHeader(width int, username string, settingsFocused bool) string {
	title := theme.AppTitleStyle.Render("SpotUI")

	var settingsBtn string
	if settingsFocused {
		settingsBtn = theme.SettingsBtnStyle.Copy().
			BorderForeground(theme.ColorGreen).
			Foreground(theme.ColorWhite).
			Render("⚙ Settings")
	} else {
		settingsBtn = theme.SettingsBtnStyle.Render("⚙ Settings")
	}

	userStr := theme.UsernameStyle.Render(username)
	right := userStr + "  " + settingsBtn

	titleW := lipgloss.Width(title)
	rightW := lipgloss.Width(right)
	innerW := width - 2
	gap := innerW - titleW - rightW
	if gap < 0 {
		gap = 0
	}

	line := theme.HeaderStyle.Copy().Width(width).Render(
		title + strings.Repeat(" ", gap) + right,
	)

	divider := theme.DividerStyle.Render(strings.Repeat("─", width))

	return line + "\n" + divider
}
