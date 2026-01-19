package app

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings for the application
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	PanelUp     key.Binding
	PanelDown   key.Binding
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
	InfoTab     key.Binding
	LogsTab     key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "move up in list"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "move down in list"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "detail view"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "detail view"),
		),
		PanelUp: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "previous panel"),
		),
		PanelDown: key.NewBinding(
			key.WithKeys("j"),
			key.WithHelp("j", "next panel"),
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
		InfoTab: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "info tab"),
		),
		LogsTab: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "logs tab"),
		),
	}
}
