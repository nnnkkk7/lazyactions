package app

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the application
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Tab         key.Binding
	ShiftTab    key.Binding
	Enter       key.Binding
	Trigger     key.Binding
	Cancel      key.Binding
	Rerun       key.Binding
	RerunFailed key.Binding
	Yank        key.Binding
	Filter      key.Binding
	Refresh     key.Binding
	FullLog     key.Binding
	Help        key.Binding
	Quit        key.Binding
	Escape      key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "previous pane"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "next pane"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next pane"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous pane"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/confirm"),
		),
		Trigger: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "trigger workflow"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "cancel run"),
		),
		Rerun: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rerun workflow"),
		),
		RerunFailed: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "rerun failed jobs"),
		),
		Yank: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "copy to clipboard"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		FullLog: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "full log view"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),
	}
}
