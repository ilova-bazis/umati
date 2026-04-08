package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// HelpModel renders the help overlay.
type HelpModel struct{}

func (h HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, keys.Escape) || key.Matches(msg, keys.Help) || key.Matches(msg, keys.Quit) {
			return h, func() tea.Msg { return formCancelledMsg{} }
		}
	}
	return h, nil
}

func (h HelpModel) View(width, height int) string {
	var sb strings.Builder
	sb.WriteString(styleFocusedLabel.Render("Keyboard Shortcuts") + "\n\n")

	sections := []struct {
		heading string
		rows    [][2]string
	}{
		{"Navigation", [][2]string{
			{"↑ / k", "move up"},
			{"↓ / j", "move down"},
			{"← / h", "previous column"},
			{"→ / l", "next column"},
			{"enter", "expand / collapse subtasks"},
			{"tab", "toggle detail panel"},
		}},
		{"Task Actions", [][2]string{
			{"c", "claim (ready → claimed)"},
			{"s", "start (claimed → in progress)"},
			{"p", "pause"},
			{"r", "release (back to ready)"},
			{"D", "done (complete task)"},
			{"X", "delete task"},
		}},
		{"Board", [][2]string{
			{"n", "create new task"},
			{"e", "edit selected task"},
			{"f", "filter by priority / agent"},
			{"R", "refresh from disk"},
			{"?", "toggle this help"},
			{"q / ctrl+c", "quit"},
		}},
	}

	for _, sec := range sections {
		sb.WriteString(styleLabelFg.Render(sec.heading) + "\n")
		for _, row := range sec.rows {
			key := styleCardID.Render(padRight(row[0], 14))
			desc := styleCardDim.Render(row[1])
			sb.WriteString("  " + key + " " + desc + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(styleCardDim.Render("press ? or esc to close"))

	overlayWidth := 52
	if width < overlayWidth+4 {
		overlayWidth = width - 4
	}
	return styleOverlay.Width(overlayWidth).Render(sb.String())
}
