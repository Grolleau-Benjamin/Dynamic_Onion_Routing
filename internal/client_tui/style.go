package clienttui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen = lipgloss.Color("#73F59F")
	colorGray  = lipgloss.Color("#626262")
	colorText  = lipgloss.Color("#c6d0f5")

	ContainerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray).
			Padding(0, 1)

	SubContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorGreen).
				Padding(0, 1)

	TabActiveStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorGreen).
			Foreground(colorText).
			Bold(true).
			Padding(0, 1)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorGray).
				Padding(0, 1)

	KeyStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Bold(true)

	DescStyle = lipgloss.NewStyle().
			Foreground(colorGray)
)
