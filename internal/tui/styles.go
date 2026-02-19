package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // purple
	ColorSecondary = lipgloss.Color("#A78BFA")
	ColorMuted     = lipgloss.Color("#6B7280")
	ColorError     = lipgloss.Color("#EF4444")
	ColorSuccess   = lipgloss.Color("#10B981")
	ColorWarning   = lipgloss.Color("#F59E0B")

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			MarginTop(1)

	ActiveStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)
)
