package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilova-bazis/umati/internal/domain"
	"github.com/ilova-bazis/umati/internal/schema"
)

// formResult holds the submitted form data.
type formResult struct {
	title       string
	description string
	priority    schema.Priority
	status      schema.Status
	parentID    string
	assignee    string // empty means no assignee
	files       []string
	isEdit      bool
	taskID      string
}

type formField int

const (
	fieldTitle formField = iota
	fieldDescription
	fieldPriority
	fieldStatus
	fieldParent
	fieldAssignee
	fieldFiles
	fieldCount
)

// enumSel cycles through a list of string options.
type enumSel struct {
	options []string
	idx     int
}

func (e *enumSel) next() {
	e.idx = (e.idx + 1) % len(e.options)
}
func (e *enumSel) prev() {
	e.idx = (e.idx - 1 + len(e.options)) % len(e.options)
}
func (e enumSel) value() string {
	if len(e.options) == 0 {
		return ""
	}
	return e.options[e.idx]
}
func (e *enumSel) setTo(v string) {
	for i, o := range e.options {
		if o == v {
			e.idx = i
			return
		}
	}
}

// FormModel is the create/edit task form overlay.
type FormModel struct {
	isEdit bool
	taskID string

	titleInput  textinput.Model
	descInput   textinput.Model
	parentInput textinput.Model

	prioritySel enumSel
	statusSel   enumSel
	assigneeSel enumSel

	focus  formField
	errMsg string

	// File picker state
	allFiles        []string
	files           []string
	fileQuery       string
	fileSuggestions []string
	fileCursor      int
	showFilePicker  bool
}

var priorityOptions = []string{"low", "medium", "high", "urgent"}
var assigneeOptions = []string{"", "human", "claude", "opencode", "codex"}

func newCreateForm(status schema.Status, allFiles []string) FormModel {
	ti := textinput.New()
	ti.Placeholder = "Task title (required)"
	ti.Focus()
	ti.CharLimit = 200

	di := textinput.New()
	di.Placeholder = "Description"
	di.CharLimit = 500

	pi := textinput.New()
	pi.Placeholder = "Parent task ID (e.g. UM-5)"
	pi.CharLimit = 20

	// Status selector only offers valid create-time statuses.
	createStatuses := []string{"draft", "paused", "ready"}
	statusIdx := 0
	for i, s := range createStatuses {
		if schema.Status(s) == status {
			statusIdx = i
			break
		}
	}

	return FormModel{
		titleInput:  ti,
		descInput:   di,
		parentInput: pi,
		prioritySel: enumSel{options: priorityOptions, idx: 1}, // default: medium
		statusSel:   enumSel{options: createStatuses, idx: statusIdx},
		assigneeSel: enumSel{options: assigneeOptions, idx: 0},
		focus:       fieldTitle,
		allFiles:    allFiles,
	}
}

func newEditForm(task schema.Task, allFiles []string) FormModel {
	m := newCreateForm(schema.StatusDraft, allFiles)
	m.isEdit = true
	m.taskID = task.ID

	m.titleInput.SetValue(task.Title)
	m.descInput.SetValue(task.Description)
	if task.ParentID != nil {
		m.parentInput.SetValue(*task.ParentID)
	}
	m.prioritySel.setTo(string(task.Priority))

	// Status options: current + all valid transitions
	statusOpts := []string{string(task.Status)}
	allowed := map[schema.Status][]schema.Status{
		schema.StatusDraft:      {schema.StatusReady, schema.StatusPaused, schema.StatusCancelled},
		schema.StatusPaused:     {schema.StatusReady, schema.StatusCancelled},
		schema.StatusReady:      {schema.StatusClaimed, schema.StatusCancelled},
		schema.StatusClaimed:    {schema.StatusInProgress, schema.StatusPaused, schema.StatusReady, schema.StatusCancelled},
		schema.StatusInProgress: {schema.StatusPaused, schema.StatusReady, schema.StatusDone, schema.StatusCancelled},
	}
	for _, s := range allowed[task.Status] {
		statusOpts = append(statusOpts, string(s))
	}
	m.statusSel = enumSel{options: statusOpts, idx: 0}

	if task.Assignee != nil {
		m.assigneeSel.setTo(string(*task.Assignee))
	}

	m.files = task.Files

	// Unfocus title since it already has content
	m.titleInput.Blur()
	m.syncFocus()

	return m
}

func (m *FormModel) syncFocus() {
	m.titleInput.Blur()
	m.descInput.Blur()
	m.parentInput.Blur()
	switch m.focus {
	case fieldTitle:
		m.titleInput.Focus()
	case fieldDescription:
		m.descInput.Focus()
	case fieldParent:
		m.parentInput.Focus()
	}
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error on any key
		m.errMsg = ""

		// File picker intercepts all keys when active.
		if m.showFilePicker {
			return m.updateFilePicker(msg)
		}

		switch {
		case key.Matches(msg, keys.Escape):
			return m, func() tea.Msg { return formCancelledMsg{} }

		case key.Matches(msg, keys.Tab), msg.String() == "down":
			m.focus = (m.focus + 1) % fieldCount
			m.syncFocus()
			return m, nil

		case msg.String() == "shift+tab", msg.String() == "up":
			m.focus = formField((int(m.focus) - 1 + int(fieldCount)) % int(fieldCount))
			m.syncFocus()
			return m, nil

		case key.Matches(msg, keys.Enter):
			if m.focus == fieldFiles {
				// Open file picker.
				m.showFilePicker = true
				m.fileQuery = ""
				m.fileSuggestions = m.allFiles
				m.fileCursor = 0
				return m, nil
			}
			// Advance focus through other fields.
			m.focus = (m.focus + 1) % fieldCount
			m.syncFocus()
			return m, nil

		case msg.String() == "backspace", msg.String() == "ctrl+h":
			if m.focus == fieldFiles && len(m.files) > 0 {
				m.files = m.files[:len(m.files)-1]
				return m, nil
			}

		case msg.String() == "left":
			if m.focus == fieldPriority || m.focus == fieldStatus || m.focus == fieldAssignee {
				m.cycleEnum(false)
				return m, nil
			}

		case msg.String() == "right":
			if m.focus == fieldPriority || m.focus == fieldStatus || m.focus == fieldAssignee {
				m.cycleEnum(true)
				return m, nil
			}

		case msg.String() == "ctrl+s":
			return m.submit()
		}
	}

	// Route key events to focused text input.
	var cmd tea.Cmd
	switch m.focus {
	case fieldTitle:
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)
	case fieldDescription:
		m.descInput, cmd = m.descInput.Update(msg)
		cmds = append(cmds, cmd)
	case fieldParent:
		m.parentInput, cmd = m.parentInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m FormModel) updateFilePicker(msg tea.KeyMsg) (FormModel, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.showFilePicker = false
		m.fileQuery = ""
		return m, nil

	case msg.String() == "up":
		if m.fileCursor > 0 {
			m.fileCursor--
		}
		return m, nil

	case msg.String() == "down":
		if m.fileCursor < len(m.fileSuggestions)-1 {
			m.fileCursor++
		}
		return m, nil

	case key.Matches(msg, keys.Enter):
		if len(m.fileSuggestions) > 0 && m.fileCursor < len(m.fileSuggestions) {
			m.files = append(m.files, m.fileSuggestions[m.fileCursor])
		}
		m.showFilePicker = false
		m.fileQuery = ""
		return m, nil

	case msg.String() == "backspace", msg.String() == "ctrl+h":
		if len(m.fileQuery) > 0 {
			runes := []rune(m.fileQuery)
			m.fileQuery = string(runes[:len(runes)-1])
			m.fileSuggestions = filterFiles(m.allFiles, m.fileQuery)
			m.fileCursor = 0
		}
		return m, nil

	default:
		if msg.Type == tea.KeyRunes {
			m.fileQuery += string(msg.Runes)
			m.fileSuggestions = filterFiles(m.allFiles, m.fileQuery)
			m.fileCursor = 0
		}
		return m, nil
	}
}

func filterFiles(all []string, query string) []string {
	if query == "" {
		return all
	}
	q := strings.ToLower(query)
	var result []string
	for _, f := range all {
		if strings.Contains(strings.ToLower(f), q) {
			result = append(result, f)
		}
	}
	return result
}

func (m *FormModel) cycleEnum(forward bool) {
	switch m.focus {
	case fieldPriority:
		if forward {
			m.prioritySel.next()
		} else {
			m.prioritySel.prev()
		}
	case fieldStatus:
		if forward {
			m.statusSel.next()
		} else {
			m.statusSel.prev()
		}
	case fieldAssignee:
		if forward {
			m.assigneeSel.next()
		} else {
			m.assigneeSel.prev()
		}
	}
}

func (m FormModel) submit() (FormModel, tea.Cmd) {
	title := strings.TrimSpace(m.titleInput.Value())
	if title == "" {
		m.errMsg = "Title is required"
		return m, nil
	}
	status := schema.Status(m.statusSel.value())
	if !m.isEdit {
		if !schema.IsValidActiveStatus(status) {
			m.errMsg = "Invalid status"
			return m, nil
		}
	}

	// Validate status transition for edit
	if m.isEdit {
		_ = domain.CanTransition // imported but not called here — the server-side op validates
	}

	r := formResult{
		title:       title,
		description: strings.TrimSpace(m.descInput.Value()),
		priority:    schema.Priority(m.prioritySel.value()),
		status:      status,
		parentID:    strings.TrimSpace(m.parentInput.Value()),
		assignee:    m.assigneeSel.value(),
		files:       m.files,
		isEdit:      m.isEdit,
		taskID:      m.taskID,
	}
	return m, func() tea.Msg { return formSubmittedMsg{result: r} }
}

func (m FormModel) View(width, height int) string {
	overlayWidth := 64
	if width < overlayWidth+4 {
		overlayWidth = width - 4
	}

	mode := "New Task"
	if m.isEdit {
		mode = fmt.Sprintf("Edit %s", m.taskID)
	}

	var sb strings.Builder
	sb.WriteString(styleFocusedLabel.Render(mode) + "\n\n")

	// Build files display for the field row.
	var filesView string
	if len(m.files) == 0 {
		filesView = styleCardDim.Render("(none)")
	} else {
		lines := make([]string, len(m.files))
		for i, f := range m.files {
			lines[i] = styleValueFg.Render(f)
		}
		filesView = strings.Join(lines, "\n              ")
	}
	if m.focus == fieldFiles {
		filesView += "  " + styleCardDim.Render("[enter: pick  backspace: remove last]")
	}

	fields := []struct {
		label string
		field formField
		view  string
	}{
		{"Title       ", fieldTitle, m.titleInput.View()},
		{"Description ", fieldDescription, m.descInput.View()},
		{"Priority    ", fieldPriority, m.renderEnum(m.prioritySel, m.focus == fieldPriority)},
		{"Status      ", fieldStatus, m.renderEnum(m.statusSel, m.focus == fieldStatus)},
		{"Parent ID   ", fieldParent, m.parentInput.View()},
		{"Assignee    ", fieldAssignee, m.renderEnum(m.assigneeSel, m.focus == fieldAssignee)},
		{"Files       ", fieldFiles, filesView},
	}

	for _, f := range fields {
		label := styleLabelFg.Render(f.label + ": ")
		if m.focus == f.field {
			label = styleFocusedLabel.Render(f.label + ": ")
		}
		sb.WriteString(label + f.view + "\n")
	}

	sb.WriteString("\n")
	if m.errMsg != "" {
		sb.WriteString(styleStatusErr.Render("  ✗ "+m.errMsg) + "\n\n")
	}

	if m.showFilePicker {
		sb.WriteString(m.renderFilePicker(overlayWidth))
	}

	sb.WriteString(renderHintPairs([][2]string{
		{"ctrl+s", "submit"}, {"esc", "cancel"}, {"tab/↑↓", "navigate"}, {"←→", "cycle options"},
	}))

	return styleOverlay.Width(overlayWidth).Render(sb.String())
}

func (m FormModel) renderFilePicker(width int) string {
	const maxVisible = 6

	var sb strings.Builder

	sep := strings.Repeat("─", max(0, width-4))
	sb.WriteString(styleCardDim.Render(sep) + "\n")
	sb.WriteString(styleFocusedLabel.Render("Search: ") + styleValueFg.Render(m.fileQuery+"_") + "\n")

	if len(m.fileSuggestions) == 0 {
		sb.WriteString(styleCardDim.Render("  (no matches)") + "\n")
	} else {
		start := m.fileCursor - maxVisible + 1
		if start < 0 {
			start = 0
		}
		end := start + maxVisible
		if end > len(m.fileSuggestions) {
			end = len(m.fileSuggestions)
		}
		for i := start; i < end; i++ {
			f := m.fileSuggestions[i]
			if i == m.fileCursor {
				sb.WriteString(styleCardIDSelected.Render("  ▸ "+f) + "\n")
			} else {
				sb.WriteString(styleValueFg.Render("    "+f) + "\n")
			}
		}
	}

	sb.WriteString(renderHintPairs([][2]string{{"enter", "add"}, {"esc", "cancel"}}) + "\n\n")
	return sb.String()
}

func (m FormModel) renderEnum(sel enumSel, focused bool) string {
	v := sel.value()
	if v == "" {
		v = "(none)"
	}
	display := "← " + v + " →"
	if focused {
		return styleCardIDSelected.Render(display)
	}
	return styleValueFg.Render(display)
}
