package domain

import (
	"sort"

	"github.com/ilova-bazis/umati/internal/schema"
)

func DirectChildren(tasks []schema.Task, parentID string) []schema.Task {
	children := make([]schema.Task, 0)
	for _, task := range tasks {
		if task.ParentID != nil && *task.ParentID == parentID {
			children = append(children, task)
		}
	}
	sortTasks(children)
	return children
}

func Descendants(tasks []schema.Task, parentID string) []schema.Task {
	result := make([]schema.Task, 0)
	for _, child := range DirectChildren(tasks, parentID) {
		result = append(result, child)
		result = append(result, Descendants(tasks, child.ID)...)
	}
	return result
}

func sortTasks(tasks []schema.Task) {
	sort.Slice(tasks, func(i, j int) bool {
		cmp, err := schema.CompareTaskIDs(tasks[i].ID, tasks[j].ID)
		if err != nil {
			return tasks[i].ID < tasks[j].ID
		}
		return cmp < 0
	})
}
