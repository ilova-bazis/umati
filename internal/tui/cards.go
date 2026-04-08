package tui

import (
	"sort"

	"github.com/ilova-bazis/umati/internal/schema"
)

// CardItem is a single rendered row in a column — either a task or a child task summary.
type CardItem struct {
	task       schema.Task
	depth      int
	hasKids    bool // has children IN THIS COLUMN
	isExpanded bool // true when children are currently visible
}

// rebuildColumnItems builds the flat list of visible card items for a column.
// Tasks are shown in their own status column. Hierarchy is formed by tasks
// whose parent is also in the same column; those children are shown only
// under their parent (not as standalone cards). Expanding a parent shows
// its same-column children inline.
func rebuildColumnItems(status schema.Status, allTasks []schema.Task, expanded map[string]bool) []CardItem {
	// Collect tasks in this column
	var colTasks []schema.Task
	for _, t := range allTasks {
		if t.Status == status {
			colTasks = append(colTasks, t)
		}
	}

	// Build set of IDs in this column
	colIDs := make(map[string]bool, len(colTasks))
	for _, t := range colTasks {
		colIDs[t.ID] = true
	}

	// Separate roots (parent not in this column) from children
	childrenOf := make(map[string][]schema.Task)
	var roots []schema.Task

	for _, t := range colTasks {
		if t.ParentID != nil && colIDs[*t.ParentID] {
			childrenOf[*t.ParentID] = append(childrenOf[*t.ParentID], t)
		} else {
			roots = append(roots, t)
		}
	}

	// Sort by ID
	sortByID(roots)
	for k := range childrenOf {
		sortByID(childrenOf[k])
	}

	var items []CardItem

	var appendTask func(t schema.Task, depth int)
	appendTask = func(t schema.Task, depth int) {
		kids := childrenOf[t.ID]
		exp := len(kids) > 0 && expanded[t.ID]
		items = append(items, CardItem{
			task:       t,
			depth:      depth,
			hasKids:    len(kids) > 0,
			isExpanded: exp,
		})
		if exp {
			for _, child := range kids {
				appendTask(child, depth+1)
			}
		}
	}

	for _, r := range roots {
		appendTask(r, 0)
	}

	return items
}

func sortByID(tasks []schema.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		cmp, err := schema.CompareTaskIDs(tasks[i].ID, tasks[j].ID)
		if err != nil {
			return tasks[i].ID < tasks[j].ID
		}
		return cmp < 0
	})
}
