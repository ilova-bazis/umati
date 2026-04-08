package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	K       key.Binding
	J       key.Binding
	H       key.Binding
	L       key.Binding
	Enter   key.Binding
	Tab     key.Binding
	Claim   key.Binding
	Start   key.Binding
	Pause   key.Binding
	Release key.Binding
	Done    key.Binding
	Delete  key.Binding
	New     key.Binding
	Edit    key.Binding
	Filter  key.Binding
	Refresh key.Binding
	Help    key.Binding
	Quit    key.Binding
	Escape  key.Binding
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
	Down:    key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
	Left:    key.NewBinding(key.WithKeys("left"), key.WithHelp("←", "left")),
	Right:   key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "right")),
	K:       key.NewBinding(key.WithKeys("k")),
	J:       key.NewBinding(key.WithKeys("j")),
	H:       key.NewBinding(key.WithKeys("h")),
	L:       key.NewBinding(key.WithKeys("l")),
	Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "expand")),
	Tab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "details")),
	Claim:   key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "claim")),
	Start:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
	Pause:   key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pause")),
	Release: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "release")),
	Done:    key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "done")),
	Delete:  key.NewBinding(key.WithKeys("X"), key.WithHelp("X", "delete")),
	New:     key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	Edit:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Filter:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
	Refresh: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "refresh")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Escape:  key.NewBinding(key.WithKeys("esc")),
}
