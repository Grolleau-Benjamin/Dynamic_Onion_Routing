package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Align(lipgloss.Center)

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	blurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))

	noStyle = lipgloss.NewStyle()

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center)

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	lineNoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81"))

	formBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

	logsBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(1, 2)

	activeFormBoxStyle = formBoxStyle.
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("205"))

	inactiveFormBoxStyle = formBoxStyle.
				BorderForeground(lipgloss.Color("240"))

	activeLogsBoxStyle = logsBoxStyle.
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("205"))

	inactiveLogsBoxStyle = logsBoxStyle.
				BorderForeground(lipgloss.Color("240"))

	errorPopupStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("196")).
			Padding(1, 2).
			Background(lipgloss.Color("52"))

	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = blurredStyle.Render("[ Submit ]")
)
