package constants

import (
	"github.com/treramey/nuget-deploy/entry"
	"github.com/treramey/nuget-deploy/project"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* CONSTANTS */

var (
	// P the current tea program
	P *tea.Program
	// Er the entry repository for the tui
	Er *entry.GormRepository
	// Pr the project repository for the tui
	Pr *project.GormRepository
	// WindowSize store the size of the terminal window
	WindowSize tea.WindowSizeMsg
)

/* STYLING */

// DocStyle styling for viewports
var DocStyle = lipgloss.NewStyle().Margin(0, 2)

// HelpStyle styling for help context menu
var HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

// ErrStyle provides styling for error messages
var ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#bd534b")).Render

// AlertStyle provides styling for alert messages
var AlertStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render

type keymap struct {
	Enter  key.Binding
	Back   key.Binding
	Quit   key.Binding
}

// Keymap reusable key mappings shared across models
var Keymap = keymap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
}
