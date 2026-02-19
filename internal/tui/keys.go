package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the global key bindings.
type KeyMap struct {
	Quit key.Binding
	Back key.Binding
}

// DefaultKeyMap returns the default global key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}
