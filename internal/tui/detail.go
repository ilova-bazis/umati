package tui

import (
	"fmt"
	"strings"

	"github.com/ilova-bazis/umati/internal/schema"
)

// DetailModel renders the bottom detail panel for the selected task.
type DetailModel struct {
	task     *schema.Task
	allTasks []schema.Task
	events   []schema.Event
}

// View renders the detail panel to a string of the given width.
func (d DetailModel) View(width int) string {
	if d.task == nil {
		return styleDetailPanel.Width(width - 2).Render("No task selected")
	}

	t := *d.task

	// Title line
	titleLine := styleCardID.Render(t.ID) + "  " +
		styleValueFg.Render(t.Title)

	// Meta line
	metaLine := fmt.Sprintf(
		"Priority: %s  Status: %s  Assignee: %s  Parent: %s  Created by: %s",
		priorityStyle(t.Priority).Render(strings.ToUpper(priorityAbbr(t.Priority))),
		statusStyle(t.Status).Render(string(t.Status)),
		styleLabelFg.Render(assigneeDisplay(t.Assignee)),
		styleLabelFg.Render(parentDisplay(t.ParentID)),
		styleLabelFg.Render(string(t.CreatedBy)),
	)

	// Description
	desc := t.Description
	if desc == "" {
		desc = styleLabelFg.Render("(no description)")
	}

	// Subtasks
	var subtaskParts []string
	for _, sub := range d.allTasks {
		if sub.ParentID != nil && *sub.ParentID == t.ID {
			subtaskParts = append(subtaskParts, fmt.Sprintf("  %s [%s] %s",
				styleCardID.Render(sub.ID),
				statusStyle(sub.Status).Render(string(sub.Status)),
				sub.Title,
			))
		}
	}

	// Events
	var eventParts []string
	for _, e := range d.events {
		ts := e.Timestamp
		if len(ts) > 16 {
			ts = ts[:16]
		}
		eventParts = append(eventParts, fmt.Sprintf("  %s by %s at %s",
			styleValueFg.Render(string(e.Type)),
			styleLabelFg.Render(string(e.Actor)),
			styleLabelFg.Render(ts),
		))
	}

	inner := titleLine + "\n" + metaLine + "\n" + desc
	if len(t.Files) > 0 {
		inner += "\n" + styleLabelFg.Render("Files:")
		for _, f := range t.Files {
			inner += "\n  " + styleValueFg.Render(f)
		}
	}
	if len(subtaskParts) > 0 {
		inner += "\n" + styleLabelFg.Render("Subtasks:") + "\n" + strings.Join(subtaskParts, "\n")
	}
	if len(eventParts) > 0 {
		inner += "\n" + styleLabelFg.Render("Recent events:") + "\n" + strings.Join(eventParts, "\n")
	}

	panelWidth := width - 2
	if panelWidth < 20 {
		panelWidth = 20
	}
	return styleDetailPanel.Width(panelWidth).Render(inner)
}

// renderActionHints returns a line of available action key hints for the given task and agent.
func renderActionHints(t *schema.Task, agent schema.Actor) string {
	if t == nil {
		return ""
	}

	var hints []string

	switch t.Status {
	case schema.StatusReady:
		hints = append(hints, "[c]laim")
	case schema.StatusClaimed:
		if t.Assignee != nil && *t.Assignee == agent {
			hints = append(hints, "[s]tart", "[p]ause", "[r]elease")
		}
	case schema.StatusInProgress:
		if t.Assignee != nil && *t.Assignee == agent {
			hints = append(hints, "[p]ause", "[r]elease", "[D]one")
		}
	case schema.StatusDraft, schema.StatusPaused:
		// no special actions
	}

	hints = append(hints, "[e]dit", "[X]delete")

	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = renderHint(h)
	}
	return styleNavLabel.Render("Actions: ") + strings.Join(parts, styleNavLabel.Render("  "))
}

// renderHint renders a hint like "[e]dit" — bracketed key in bright white, rest in cyan.
func renderHint(h string) string {
	start := strings.Index(h, "[")
	end := strings.Index(h, "]")
	if start == -1 || end == -1 || end <= start {
		return styleNavLabel.Render(h)
	}
	return styleNavKey.Render(h[start:end+1]) + styleNavLabel.Render(h[end+1:])
}

// renderHintPairs renders key:label pairs with consistent nav-bar styling.
func renderHintPairs(pairs [][2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = styleNavKey.Render(p[0]) + styleNavLabel.Render(":"+p[1])
	}
	return strings.Join(parts, styleNavLabel.Render("  "))
}

func assigneeDisplay(a *schema.Actor) string {
	if a == nil {
		return "-"
	}
	return string(*a)
}

func parentDisplay(p *string) string {
	if p == nil {
		return "-"
	}
	return *p
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(runes[:n-1]) + "…"
}

func padRight(s string, n int) string {
	runes := []rune(s)
	if len(runes) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(runes))
}

// renderCardLine renders a single line for a card item.
// The line is exactly `width` visible characters wide.
func renderCardLine(item CardItem, lineNum int, isSelected bool, width int) string {
	t := item.task
	depth := item.depth
	indent := strings.Repeat("  ", depth)

	bg := func(s string) string {
		// Pad s to width and apply selection background.
		raw := []rune(s)
		pad := width - len(raw)
		if pad < 0 {
			pad = 0
			s = string(raw[:width])
		}
		return styleCardSelected.Render(s + strings.Repeat(" ", pad))
	}
	plain := func(s string) string {
		raw := []rune(s)
		pad := width - len(raw)
		if pad < 0 {
			pad = 0
			s = string(raw[:width])
		}
		return s + strings.Repeat(" ", pad)
	}

	switch lineNum {
	case 0: // ID + expander
		var exp string
		switch {
		case item.hasKids && item.isExpanded:
			exp = "▾ "
		case item.hasKids:
			exp = "▸ "
		case depth > 0:
			exp = "└ "
		default:
			exp = "  "
		}
		content := indent + exp + t.ID
		if isSelected {
			return styleCardIDSelected.Render(content +
				strings.Repeat(" ", max(0, width-len([]rune(content)))))
		}
		return styleCardID.Render(indent+exp+t.ID) +
			strings.Repeat(" ", max(0, width-len([]rune(content))))

	case 1: // Title
		inner := indent + "  " + truncate(t.Title, width-len([]rune(indent))-2)
		if isSelected {
			return bg(inner)
		}
		return plain(inner)

	case 2: // Priority + assignee
		prio := priorityAbbr(t.Priority)
		asgn := assigneeDisplay(t.Assignee)
		inner := indent + "  " + prio + "  " + asgn
		if isSelected {
			return bg(inner)
		}
		// Colorize priority on non-selected rows
		coloredInner := indent + "  " + priorityStyle(t.Priority).Render(prio) + "  " + styleLabelFg.Render(asgn)
		rawLen := len([]rune(inner))
		pad := width - rawLen
		if pad < 0 {
			pad = 0
		}
		return coloredInner + strings.Repeat(" ", pad)

	default: // Blank separator
		if isSelected {
			return bg("")
		}
		return strings.Repeat(" ", width)
	}
}
