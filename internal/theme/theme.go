package theme

import "github.com/charmbracelet/lipgloss"

var (
	ColorBG      = lipgloss.Color("#1A1817")
	ColorSurface = lipgloss.Color("#252220")
	ColorAmber   = lipgloss.Color("#E8A838")
	ColorText    = lipgloss.Color("#E8E0D5")
	ColorMuted   = lipgloss.Color("#8A7E72")
	ColorDim     = lipgloss.Color("#5A5250")
	ColorRose    = lipgloss.Color("#C4526E")
)

var (
	TopBarStyle = lipgloss.NewStyle().
			Background(ColorSurface).
			Padding(0, 1)

	TopBarTitle = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Bold(true)

	TopBarUser = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorAmber).
			Bold(true)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorDim)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorAmber).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	SubtextStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	ArtistNameStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	PlayerBarStyle = lipgloss.NewStyle().
			Background(ColorSurface).
			Padding(0, 1)

	TrackNameStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Bold(true)

	ProgressStyle = lipgloss.NewStyle().
			Foreground(ColorAmber)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorDim)

	ShuffleOnStyle = lipgloss.NewStyle().
			Foreground(ColorRose)

	ShuffleOffStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRose).
			Bold(true)

	StatusStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)
