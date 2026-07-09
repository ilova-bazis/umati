package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilova-bazis/umati/internal/schema"
)

// FilterState holds the active board filter values.
type FilterState struct {
	priority *schema.Priority
	agent    *schema.Actor
	tags     []string
}

func (fs FilterState) isActive() bool {
	return fs.priority != nil || fs.agent != nil || len(fs.tags) > 0
}

func (fs FilterState) label() string {
	var parts []string
	if fs.priority != nil {
		parts = append(parts, "priority:"+string(*fs.priority))
	}
	if fs.agent != nil {
		parts = append(parts, "agent:"+string(*fs.agent))
	}
	if len(fs.tags) > 0 {
		parts = append(parts, "tags:"+strings.Join(fs.tags, ","))
	}
	if len(parts) == 0 {
		return ""
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// FilterModel is the filter panel overlay.
type FilterModel struct {
	prioritySel enumSel // index 0 = "any"
	agentSel    enumSel // index 0 = "any"
	tagsInput   textinput.Model
	focus       int // 0 = priority, 1 = agent, 2 = tags
}

var filterPriorityOptions = []string{"any", "low", "medium", "high", "urgent"}
var filterAgentOptions = []string{"any", "human", "claude", "opencode", "codex"}

func newFilterModel(current FilterState) FilterModel {
	ps := enumSel{options: filterPriorityOptions}
	as := enumSel{options: filterAgentOptions}

	if current.priority != nil {
		ps.setTo(string(*current.priority))
	}
	if current.agent != nil {
		as.setTo(string(*current.agent))
	}

	ti := textinput.New()
	ti.Placeholder = "comma-separated (bug, api)"
	ti.CharLimit = 200
	if len(current.tags) > 0 {
		ti.SetValue(strings.Join(current.tags, ", "))
	}

	return FilterModel{
		prioritySel: ps,
		agentSel:    as,
		tagsInput:   ti,
	}
}

func (m *FilterModel) syncFocus() {
	if m.focus == 2 {
		m.tagsInput.Focus()
	} else {
		m.tagsInput.Blur()
	}
}

func (m FilterModel) Update(msg tea.Msg) (FilterModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Escape):
			return m, func() tea.Msg { return filterCancelledMsg{} }

		case key.Matches(msg, keys.Enter):
			return m, m.applyCmd()

		case key.Matches(msg, keys.Tab), msg.String() == "down":
			m.focus = (m.focus + 1) % 3
			m.syncFocus()
		case msg.String() == "shift+tab", msg.String() == "up":
			m.focus = (m.focus - 1 + 3) % 3
			m.syncFocus()

		case msg.String() == "left":
			if m.focus == 0 {
				m.prioritySel.prev()
			} else if m.focus == 1 {
				m.agentSel.prev()
			}
		case msg.String() == "right":
			if m.focus == 0 {
				m.prioritySel.next()
			} else if m.focus == 1 {
				m.agentSel.next()
			}

		default:
			// Route to tagsInput when focused
			if m.focus == 2 {
				var cmd tea.Cmd
				m.tagsInput, cmd = m.tagsInput.Update(msg)
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m FilterModel) applyCmd() tea.Cmd {
	var priority *schema.Priority
	if m.prioritySel.value() != "any" {
		p := schema.Priority(m.prioritySel.value())
		priority = &p
	}
	var agent *schema.Actor
	if m.agentSel.value() != "any" {
		a := schema.Actor(m.agentSel.value())
		agent = &a
	}
	tags := parseTags(m.tagsInput.Value())
	return func() tea.Msg {
		return filterAppliedMsg{priority: priority, agent: agent, tags: tags}
	}
}

func (m FilterModel) View(width, height int) string {
	var sb strings.Builder
	sb.WriteString(styleFocusedLabel.Render("Filter Board") + "\n\n")

	prioLabel := styleLabelFg.Render("Priority   : ")
	if m.focus == 0 {
		prioLabel = styleFocusedLabel.Render("Priority   : ")
	}
	sb.WriteString(prioLabel + renderEnumSel(m.prioritySel, m.focus == 0) + "\n")

	agentLabel := styleLabelFg.Render("Agent      : ")
	if m.focus == 1 {
		agentLabel = styleFocusedLabel.Render("Agent      : ")
	}
	sb.WriteString(agentLabel + renderEnumSel(m.agentSel, m.focus == 1) + "\n")

	tagsLabel := styleLabelFg.Render("Tags       : ")
	if m.focus == 2 {
		tagsLabel = styleFocusedLabel.Render("Tags       : ")
	}
	sb.WriteString(tagsLabel + m.tagsInput.View() + "\n")

	sb.WriteString("\n")
	sb.WriteString(styleCardDim.Render("enter: apply  esc: cancel  ←→: change  tab/↑↓: navigate"))

	overlayWidth := 48
	if width < overlayWidth+4 {
		overlayWidth = width - 4
	}
	return styleOverlay.Width(overlayWidth).Render(sb.String())
}

func renderEnumSel(sel enumSel, focused bool) string {
	v := sel.value()
	display := "← " + v + " →"
	if focused {
		return styleCardIDSelected.Render(display)
	}
	return styleValueFg.Render(display)
}
