package views

import (
	"fmt"
	"image"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dcbto/spotui/internal/artwork"
	"github.com/dcbto/spotui/internal/theme"
)

type BrowseItem struct {
	Header      bool
	Title       string
	Subtitle    string
	DurationMS  int
	ImageURL    string
	ExternalURL string
}

type BrowseViewState struct {
	Width, Height    int
	NavigationCursor int
	NavigationFocus  bool
	NavigationLabels []string
	Title, Meta      string
	Search           bool
	SearchQuery      string
	Loading          bool
	Error            string
	Cursor           int
	Focus            int
	Items            []BrowseItem
	Artwork          map[string]image.Image
	StatusMessage    string
	StatusIsError    bool
	Player           EnginePlayerState
}

func RenderBrowseShell(state BrowseViewState) string {
	top := RenderBrowseTopBar(state.Width, navigationLabel(state))
	wide := state.Width >= 80 && state.Height >= 24
	playerWidth := state.Width
	if wide {
		playerWidth -= 24
	}
	player := RenderEnginePlayer(playerWidth, state.Player)
	playerHeight := lipgloss.Height(player)
	contentHeight := state.Height - lipgloss.Height(top) - playerHeight
	if state.StatusMessage != "" {
		contentHeight--
	}
	if contentHeight < 1 {
		contentHeight = 1
	}
	contentWidth := state.Width
	if wide {
		contentWidth -= 24
	}
	if contentWidth < 1 {
		contentWidth = 1
	}
	content := renderBrowseContent(state, contentWidth, contentHeight)
	var body string
	if wide {
		navigation := renderNavigation(state.NavigationCursor, state.NavigationFocus, 24, contentHeight, state.NavigationLabels)
		body = lipgloss.JoinHorizontal(lipgloss.Top, navigation, lipgloss.NewStyle().Width(contentWidth).Height(contentHeight).Render(content))
	} else if state.NavigationFocus {
		body = renderNavigation(state.NavigationCursor, true, state.Width, contentHeight, state.NavigationLabels)
	} else {
		body = lipgloss.NewStyle().Width(contentWidth).Height(contentHeight).Render(content)
	}
	rows := []string{top, body}
	if state.StatusMessage != "" {
		if state.StatusIsError {
			rows = append(rows, theme.ErrorStyle.Render("  ✗ "+state.StatusMessage))
		} else {
			rows = append(rows, theme.StatusStyle.Render("  "+state.StatusMessage))
		}
	}
	if wide {
		rows = append(rows, lipgloss.NewStyle().MarginLeft(24).Render(player))
	} else {
		rows = append(rows, player)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func navigationLabel(state BrowseViewState) string {
	if state.NavigationCursor < 0 || state.NavigationCursor >= len(state.NavigationLabels) {
		return "Browse"
	}
	return state.NavigationLabels[state.NavigationCursor]
}

func renderNavigation(cursor int, focused bool, width, height int, labels []string) string {
	rows := []string{"", theme.TopBarTitle.Render("  Browse")}
	for i, label := range labels {
		style := theme.InactiveTabStyle
		if i == cursor {
			style = theme.SelectedItemStyle
			if focused {
				style = theme.ActiveTabStyle
			}
		}
		rows = append(rows, style.Render("  "+label))
	}
	rows = append(rows, "", theme.SubtextStyle.Render("  Tab focus"), theme.SubtextStyle.Render("  Enter open"), theme.SubtextStyle.Render("  Esc back"))
	return lipgloss.NewStyle().Width(width).Height(height).Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(theme.ColorDim).Render(strings.Join(rows, "\n"))
}

func renderBrowseContent(state BrowseViewState, width, height int) string {
	title := state.Title
	if state.Meta != "" {
		title += " · " + state.Meta
	}
	rows := []string{"  " + theme.TopBarTitle.Render(title)}
	if state.Cursor >= 0 && state.Cursor < len(state.Items) {
		if img := state.Artwork[state.Items[state.Cursor].ImageURL]; img != nil {
			rows = append(rows, artwork.RenderANSI(img, artwork.CellRect{Columns: 14, Rows: 6}))
		}
	}
	if state.Search {
		rows = append(rows, "  "+theme.SubtextStyle.Render("Search: "+state.SearchQuery))
	}
	if state.Loading && len(state.Items) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, append(rows, "", "  "+theme.SubtextStyle.Render("Loading..."))...)
	}
	if state.Error != "" {
		rows = append(rows, "", "  "+theme.ErrorStyle.Render(state.Error), "  "+theme.SubtextStyle.Render("Press r to retry"))
		return strings.Join(rows, "\n")
	}
	if len(state.Items) == 0 {
		rows = append(rows, "", "  "+theme.SubtextStyle.Render("No items found"))
		return strings.Join(rows, "\n")
	}
	for i, item := range state.Items {
		prefix := "  "
		style := theme.NormalItemStyle
		if i == state.Cursor && state.Focus == 1 {
			prefix = "▸ "
			style = theme.SelectedItemStyle
		}
		if item.Header {
			rows = append(rows, "", "  "+theme.ActiveTabStyle.Render(item.Title))
			continue
		}
		detail := item.Subtitle
		if item.ImageURL != "" {
			detail = "▣ " + detail
		}
		if item.DurationMS > 0 {
			detail += " · " + formatBrowseDuration(item.DurationMS)
		}
		rows = append(rows, prefix+style.Render(truncateBrowse(item.Title, width-6))+"  "+theme.SubtextStyle.Render(truncateBrowse(detail, width/2)))
		if len(rows) >= height {
			break
		}
	}
	return strings.Join(rows, "\n")
}

func formatBrowseDuration(ms int) string {
	return fmt.Sprintf("%02d:%02d", ms/60000, (ms/1000)%60)
}

func truncateBrowse(value string, width int) string {
	if width < 4 {
		width = 4
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-1]) + "…"
}
