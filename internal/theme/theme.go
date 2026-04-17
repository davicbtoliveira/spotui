package theme

import "github.com/charmbracelet/lipgloss"

var (
	ColorGreen = lipgloss.Color("#1DB954")
	ColorBlack = lipgloss.Color("#191414")
	ColorWhite = lipgloss.Color("#FFFFFF")
	ColorGray  = lipgloss.Color("#535353")
	ColorLGray = lipgloss.Color("#B3B3B3")
	ColorDGray = lipgloss.Color("#282828")
)

var (
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorDGray).
			Padding(0, 1)

	AppTitleStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true)

	UsernameStyle = lipgloss.NewStyle().
			Foreground(ColorLGray)

	SettingsBtnStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorGray).
				Padding(0, 1)

	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(ColorGreen).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorGreen).
			Padding(0, 1)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorGray).
				Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorWhite)

	SubtextStyle = lipgloss.NewStyle().
			Foreground(ColorLGray)

	ArtistNameStyle = lipgloss.NewStyle().
			Foreground(ColorLGray)

	PlayerBarStyle = lipgloss.NewStyle().
			Background(ColorDGray).
			Padding(0, 1)

	TrackNameStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	ProgressStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	ProgressEmptyStyle = lipgloss.NewStyle().
				Foreground(ColorGray)

	ShuffleOnStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	ShuffleOffStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	DividerStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	StatusStyle = lipgloss.NewStyle().
			Foreground(ColorLGray)
)
