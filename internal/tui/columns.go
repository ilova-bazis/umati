package tui

import "github.com/ilova-bazis/umati/internal/schema"

// defaultColumns is the ordered list of columns shown in the board.
var defaultColumns = []schema.Status{
	schema.StatusDraft,
	schema.StatusPaused,
	schema.StatusReady,
	schema.StatusClaimed,
	schema.StatusInProgress,
	schema.StatusDone,
	schema.StatusCancelled,
}

// ColumnState holds the display state for one column.
type ColumnState struct {
	items    []CardItem
	cursor   int
	offset   int          // scroll offset (first visible item index)
	expanded map[string]bool
}

func newColumnState() *ColumnState {
	return &ColumnState{
		expanded: make(map[string]bool),
	}
}

// rebuild refreshes the items slice from the full task list, preserving
// expanded state and clamping cursor/offset to valid bounds.
func (cs *ColumnState) rebuild(status schema.Status, allTasks []schema.Task) {
	cs.items = rebuildColumnItems(status, allTasks, cs.expanded)
	cs.clamp(0)
}

// clamp ensures cursor and offset are within valid bounds.
// Cursor may range from 0 to len(items) inclusive — len(items) is the add-row position.
// visibleRows=0 means don't constrain offset.
func (cs *ColumnState) clamp(visibleRows int) {
	// Valid cursor range: [0, len(items)] — the add row is at len(items).
	n := len(cs.items)
	if cs.cursor > n {
		cs.cursor = n
	}
	if cs.cursor < 0 {
		cs.cursor = 0
	}
	if visibleRows > 0 {
		// Include the add-row slot when computing scroll bounds.
		total := n + 1
		if cs.cursor < cs.offset {
			cs.offset = cs.cursor
		}
		if cs.cursor >= cs.offset+visibleRows {
			cs.offset = cs.cursor - visibleRows + 1
		}
		maxOffset := max(0, total-visibleRows)
		if cs.offset > maxOffset {
			cs.offset = maxOffset
		}
	}
	if cs.offset < 0 {
		cs.offset = 0
	}
}

// selectedTask returns the task at the cursor, or nil if the column is empty.
func (cs *ColumnState) selectedTask() *schema.Task {
	if len(cs.items) == 0 || cs.cursor >= len(cs.items) {
		return nil
	}
	t := cs.items[cs.cursor].task
	return &t
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
