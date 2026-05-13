package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/workspace"
)

type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayForm
	overlayFilter
	overlayHelp
)

// linesPerCard is the number of terminal lines each card occupies (3 content + 1 blank).
const linesPerCard = 4

// BoardModel is the root bubbletea model.
type BoardModel struct {
	ctx   workspace.Context
	cfg   schema.Config
	agent schema.Actor

	tasks   []schema.Task
	taskMap map[string]schema.Task

	columns  []schema.Status
	activeCol int
	colState map[schema.Status]*ColumnState

	filter FilterState

	overlay overlayKind
	detail  DetailModel
	form    FormModel
	filter_ FilterModel
	help    HelpModel

	showDetail   bool
	detailEvents []schema.Event

	statusMsg    string
	statusIsErr  bool

	width, height int
	ready         bool
}

func NewBoardModel(ctx workspace.Context, cfg schema.Config, agent schema.Actor) BoardModel {
	cs := make(map[schema.Status]*ColumnState, len(defaultColumns))
	for _, col := range defaultColumns {
		cs[col] = newColumnState()
	}
	return BoardModel{
		ctx:      ctx,
		cfg:      cfg,
		agent:    agent,
		columns:  defaultColumns,
		colState: cs,
	}
}

func (m BoardModel) Init() tea.Cmd {
	return tea.Batch(loadTasksCmd(m.ctx), tickCmd())
}

// Run starts the bubbletea program.
func Run(ctx workspace.Context, cfg schema.Config, agent schema.Actor) error {
	m := NewBoardModel(ctx, cfg, agent)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// --- Update ---

func (m BoardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always handle window resize.
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		m.ready = true
		return m, nil
	}

	// Overlay completion messages must be handled at the board level regardless
	// of which overlay is active (they are dispatched on the next tick, after the
	// overlay's cmd runs, so m.overlay is still set when they arrive).
	switch msg.(type) {
	case formSubmittedMsg, formCancelledMsg,
		filterAppliedMsg, filterCancelledMsg:
		return m.updateBoard(msg)
	}

	// Route to overlay if active.
	switch m.overlay {
	case overlayForm:
		return m.updateFormOverlay(msg)
	case overlayFilter:
		return m.updateFilterOverlay(msg)
	case overlayHelp:
		return m.updateHelpOverlay(msg)
	}

	return m.updateBoard(msg)
}

func (m BoardModel) updateBoard(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tasksLoadedMsg:
		m.tasks = msg.tasks
		m.taskMap = make(map[string]schema.Task, len(msg.tasks))
		for _, t := range msg.tasks {
			m.taskMap[t.ID] = t
		}
		m.rebuildAllColumns()
		// Refresh detail panel events if it's open.
		if m.showDetail {
			if t := m.selectedTask(); t != nil {
				return m, loadEventsCmd(m.ctx, t.ID)
			}
		}
		return m, nil

	case eventsLoadedMsg:
		m.detailEvents = msg.events
		m.detail = DetailModel{
			task:     m.selectedTask(),
			allTasks: m.tasks,
			events:   m.detailEvents,
		}
		return m, nil

	case tickMsg:
		return m, tea.Batch(loadTasksCmd(m.ctx), tickCmd())

	case mutationDoneMsg:
		m.setStatus(fmt.Sprintf("✓ %s %s", msg.command, msg.taskID), false)
		return m, tea.Batch(loadTasksCmd(m.ctx), m.clearStatusAfter())

	case errMsg:
		m.setStatus("✗ "+msg.err.Error(), true)
		return m, m.clearStatusAfter()

	case clearStatusMsg:
		m.statusMsg = ""
		m.statusIsErr = false
		return m, nil

	case formSubmittedMsg:
		m.overlay = overlayNone
		r := msg.result
		if r.isEdit {
			return m, updateTaskCmd(m.ctx, m.agent, r.taskID, r)
		}
		return m, createTaskCmd(m.ctx, m.agent, r)

	case formCancelledMsg:
		m.overlay = overlayNone
		return m, nil

	case filterAppliedMsg:
		m.filter = FilterState{priority: msg.priority, agent: msg.agent}
		m.overlay = overlayNone
		m.rebuildAllColumns()
		return m, nil

	case filterCancelledMsg:
		m.overlay = overlayNone
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m BoardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Refresh):
		return m, loadTasksCmd(m.ctx)

	case key.Matches(msg, keys.Help):
		m.overlay = overlayHelp
		return m, nil

	case key.Matches(msg, keys.Filter):
		m.overlay = overlayFilter
		m.filter_ = newFilterModel(m.filter)
		return m, nil

	case key.Matches(msg, keys.Tab):
		m.showDetail = !m.showDetail
		if m.showDetail {
			if t := m.selectedTask(); t != nil {
				m.detail = DetailModel{task: t, allTasks: m.tasks}
				return m, loadEventsCmd(m.ctx, t.ID)
			}
		}
		return m, nil

	case key.Matches(msg, keys.Up), key.Matches(msg, keys.K):
		m.moveCursor(-1)
		m.updateDetailTask()
		return m, nil

	case key.Matches(msg, keys.Down), key.Matches(msg, keys.J):
		m.moveCursor(1)
		m.updateDetailTask()
		return m, nil

	case key.Matches(msg, keys.Left), key.Matches(msg, keys.H):
		if m.activeCol > 0 {
			m.activeCol--
		}
		m.updateDetailTask()
		return m, nil

	case key.Matches(msg, keys.Right), key.Matches(msg, keys.L):
		if m.activeCol < len(m.columns)-1 {
			m.activeCol++
		}
		m.updateDetailTask()
		return m, nil

	case key.Matches(msg, keys.Enter):
		cs := m.colState[m.columns[m.activeCol]]
		if cs.cursor == len(cs.items) {
			// Cursor is on the add row — open create form for this column's status.
			m.overlay = overlayForm
			m.form = newCreateForm(m.columns[m.activeCol], listWorkspaceFiles(m.ctx.Root))
			return m, nil
		}
		m.toggleExpand()
		return m, nil

	case key.Matches(msg, keys.New):
		m.overlay = overlayForm
		m.form = newCreateForm(m.columns[m.activeCol], listWorkspaceFiles(m.ctx.Root))
		return m, nil

	case key.Matches(msg, keys.Edit):
		if t := m.selectedTask(); t != nil {
			m.overlay = overlayForm
			m.form = newEditForm(*t, listWorkspaceFiles(m.ctx.Root))
		}
		return m, nil

	// Status transitions
	case key.Matches(msg, keys.Claim):
		if t := m.selectedTask(); t != nil {
			if t.Status != schema.StatusReady {
				m.setStatus(fmt.Sprintf("✗ %s is not ready (status: %s)", t.ID, t.Status), true)
				return m, m.clearStatusAfter()
			}
			return m, claimTaskCmd(m.ctx, t.ID, m.agent)
		}

	case key.Matches(msg, keys.Start):
		if t := m.selectedTask(); t != nil {
			if t.Status != schema.StatusClaimed {
				m.setStatus(fmt.Sprintf("✗ %s is not claimed (status: %s)", t.ID, t.Status), true)
				return m, m.clearStatusAfter()
			}
			return m, startTaskCmd(m.ctx, t.ID, m.agent)
		}

	case key.Matches(msg, keys.Pause):
		if t := m.selectedTask(); t != nil {
			if t.Status != schema.StatusClaimed && t.Status != schema.StatusInProgress {
				m.setStatus(fmt.Sprintf("✗ cannot pause %s (status: %s)", t.ID, t.Status), true)
				return m, m.clearStatusAfter()
			}
			return m, pauseTaskCmd(m.ctx, t.ID, m.agent)
		}

	case key.Matches(msg, keys.Release):
		if t := m.selectedTask(); t != nil {
			if t.Status != schema.StatusClaimed && t.Status != schema.StatusInProgress {
				m.setStatus(fmt.Sprintf("✗ cannot release %s (status: %s)", t.ID, t.Status), true)
				return m, m.clearStatusAfter()
			}
			return m, releaseTaskCmd(m.ctx, t.ID, m.agent)
		}

	case key.Matches(msg, keys.Done):
		if t := m.selectedTask(); t != nil {
			if t.Status != schema.StatusInProgress {
				m.setStatus(fmt.Sprintf("✗ %s is not in progress (status: %s)", t.ID, t.Status), true)
				return m, m.clearStatusAfter()
			}
			return m, completeTaskCmd(m.ctx, t.ID, m.agent)
		}

	case key.Matches(msg, keys.Delete):
		if t := m.selectedTask(); t != nil {
			return m, deleteTaskCmd(m.ctx, t.ID, m.agent)
		}
	}

	return m, nil
}

func (m BoardModel) updateFormOverlay(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m BoardModel) updateFilterOverlay(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.filter_, cmd = m.filter_.Update(msg)
	return m, cmd
}

func (m BoardModel) updateHelpOverlay(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.help, cmd = m.help.Update(msg)
	// formCancelledMsg is reused to signal "close overlay"
	return m, cmd
}

// --- Helpers ---

func (m *BoardModel) rebuildAllColumns() {
	filtered := m.filteredTasks()
	for _, col := range m.columns {
		m.colState[col].rebuild(col, filtered)
	}
}

func (m BoardModel) filteredTasks() []schema.Task {
	if !m.filter.isActive() {
		return m.tasks
	}
	var out []schema.Task
	for _, t := range m.tasks {
		if m.filter.priority != nil && t.Priority != *m.filter.priority {
			continue
		}
		if m.filter.agent != nil && (t.Assignee == nil || *t.Assignee != *m.filter.agent) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (m *BoardModel) moveCursor(delta int) {
	cs := m.colState[m.columns[m.activeCol]]
	cs.cursor += delta
	visibleRows := m.cardsAreaHeight()
	// +1 to account for the add-row slot
	cs.clamp((visibleRows / linesPerCard) + 1)
}

func (m BoardModel) selectedTask() *schema.Task {
	if !m.ready || len(m.columns) == 0 {
		return nil
	}
	return m.colState[m.columns[m.activeCol]].selectedTask()
}

func (m *BoardModel) toggleExpand() {
	cs := m.colState[m.columns[m.activeCol]]
	if len(cs.items) == 0 {
		return
	}
	item := cs.items[cs.cursor]
	if item.hasKids {
		cs.expanded[item.task.ID] = !cs.expanded[item.task.ID]
		cs.rebuild(m.columns[m.activeCol], m.filteredTasks())
	}
}

func (m *BoardModel) updateDetailTask() {
	if m.showDetail {
		m.detail = DetailModel{
			task:     m.selectedTask(),
			allTasks: m.tasks,
			events:   m.detailEvents,
		}
	}
}

func (m *BoardModel) setStatus(msg string, isErr bool) {
	m.statusMsg = msg
	m.statusIsErr = isErr
}

func (m BoardModel) clearStatusAfter() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// --- View ---

func (m BoardModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	switch m.overlay {
	case overlayForm:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.form.View(m.width, m.height))
	case overlayFilter:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.filter_.View(m.width, m.height))
	case overlayHelp:
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			m.help.View(m.width, m.height))
	}

	var parts []string
	parts = append(parts, m.renderBoard())
	if m.showDetail {
		parts = append(parts, m.renderDetailPanel())
	}
	parts = append(parts, m.renderStatusBar())
	parts = append(parts, m.renderHelpBar())

	return strings.Join(parts, "\n")
}


func (m BoardModel) renderBoard() string {
	colW := m.columnWidth()
	cardsH := m.cardsAreaHeight()
	visibleCards := cardsH / linesPerCard

	colViews := make([]string, len(m.columns))
	for i, col := range m.columns {
		cs := m.colState[col]
		cs.clamp(visibleCards)
		isActive := i == m.activeCol
		colViews[i] = m.renderColumn(col, cs, isActive, colW, cardsH)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, colViews...)
}

func (m BoardModel) renderColumn(status schema.Status, cs *ColumnState, isActive bool, colW, cardsH int) string {
	// Header
	label := columnLabel(status)
	count := fmt.Sprintf("(%d)", len(cs.items))
	headerText := label + " " + count
	var header string
	if isActive {
		header = styleColHeaderActive.Width(colW).Render(headerText)
	} else {
		header = styleColHeader.Width(colW).Render(headerText)
	}

	// Separator
	sep := styleColSep.Render(strings.Repeat("─", colW))

	// In the active column, reserve one card-slot for the add row.
	visibleCards := cardsH / linesPerCard
	if isActive && visibleCards > 0 {
		visibleCards--
	}

	var cardLines []string

	end := cs.offset + visibleCards
	if end > len(cs.items) {
		end = len(cs.items)
	}

	for idx := cs.offset; idx < end; idx++ {
		item := cs.items[idx]
		isSelected := isActive && idx == cs.cursor
		for line := 0; line < linesPerCard; line++ {
			cardLines = append(cardLines, renderCardLine(item, line, isSelected, colW))
		}
	}

	// Pad card area to fill the reserved card slots.
	totalCardLines := visibleCards * linesPerCard
	for len(cardLines) < totalCardLines {
		cardLines = append(cardLines, strings.Repeat(" ", colW))
	}
	if len(cardLines) > totalCardLines {
		cardLines = cardLines[:totalCardLines]
	}

	// Add row (active column only).
	if isActive {
		addSelected := cs.cursor == len(cs.items)
		addLine0 := renderAddRow(addSelected, colW)
		cardLines = append(cardLines,
			addLine0,
			strings.Repeat(" ", colW),
			strings.Repeat(" ", colW),
			strings.Repeat(" ", colW),
		)
	}

	content := header + "\n" + sep + "\n" + strings.Join(cardLines, "\n")

	borderStyle := lipgloss.NewStyle().
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder)

	return borderStyle.Render(content)
}

func (m BoardModel) renderDetailPanel() string {
	return m.detail.View(m.width)
}

func (m BoardModel) renderStatusBar() string {
	// Left: workspace + agent identity badge (always shown).
	left := styleWorkspaceBadge.Render("workspace:" + m.cfg.IDPrefix + "  agent:" + string(m.agent))
	if m.filter.isActive() {
		left += "  " + styleStatusMsg.Render(m.filter.label())
	}

	// Right: status message or selected-task summary.
	var right string
	if m.statusMsg != "" {
		if m.statusIsErr {
			right = styleStatusErr.Render(m.statusMsg)
		} else {
			right = styleStatusMsg.Render(m.statusMsg)
		}
	} else if t := m.selectedTask(); t != nil {
		right = styleLabelFg.Render(fmt.Sprintf("%s · %s · %s · %s",
			t.ID, t.Title, string(t.Status), assigneeDisplay(t.Assignee)))
	}

	if right != "" {
		pad := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
		if pad < 1 {
			pad = 1
		}
		return styleStatusBar.Width(m.width).Render(left + strings.Repeat(" ", pad) + right)
	}
	return styleStatusBar.Width(m.width).Render(left)
}

func (m BoardModel) renderHelpBar() string {
	var hints string
	if t := m.selectedTask(); t != nil {
		hints = renderActionHints(t, m.agent)
	} else {
		hints = renderNavHints()
	}
	return styleStatusBar.Width(m.width).Render(hints)
}

func renderNavHints() string {
	return renderHintPairs([][2]string{
		{"↑↓", "nav"}, {"←→", "col"}, {"enter", "expand"}, {"tab", "details"},
		{"n", "new"}, {"f", "filter"}, {"R", "refresh"}, {"?", "help"}, {"q", "quit"},
	})
}

// --- Layout calculations ---

func (m BoardModel) fixedRows() int {
	// status-bar(1) + help-bar(1) = 2
	// plus column-header(1) + separator(1) = 2
	rows := 4
	if m.showDetail {
		rows += m.detailHeight()
	}
	return rows
}

func (m BoardModel) cardsAreaHeight() int {
	h := m.height - m.fixedRows()
	if h < linesPerCard {
		h = linesPerCard
	}
	return h
}

func (m BoardModel) detailHeight() int {
	return 9 // fixed detail panel height
}

func (m BoardModel) columnWidth() int {
	n := len(m.columns)
	if n == 0 {
		return 20
	}
	w := (m.width - n) / n
	if w < 14 {
		w = 14
	}
	return w
}

// renderAddRow renders the "+ Add task" row at the bottom of the active column.
func renderAddRow(isSelected bool, width int) string {
	label := "  + Add task"
	raw := []rune(label)
	pad := width - len(raw)
	if pad < 0 {
		pad = 0
		label = string(raw[:width])
	}
	content := label + strings.Repeat(" ", pad)
	if isSelected {
		return styleCardIDSelected.Render(content)
	}
	return styleCardDim.Render(content)
}
