package ui

import "github.com/charmbracelet/lipgloss"

const (
	defaultWidth  = 104
	defaultHeight = 34
	minWidth      = 44
	minHeight     = 18
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("14"))
	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("14")).
			Padding(0, 1)
	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")).
				Padding(0, 1)
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(0, 1)
	focusedPanelStyle = panelStyle.Copy().
				BorderForeground(lipgloss.Color("14"))
	statusInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true)
	statusWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true)
	statusErrorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true)
	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")).
				Bold(true)
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("14"))
	overlayStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("14")).
			Padding(0, 1)
)
