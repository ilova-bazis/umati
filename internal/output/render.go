package output

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ilova-bazis/umati/internal/domain"
	"github.com/ilova-bazis/umati/internal/schema"
)

// Event is an alias for schema.Event for convenience.
type Event = schema.Event

// RenderListAll renders all active tasks in a hierarchical view.
func RenderListAll(tasks []schema.Task, idPrefix string) {
	if len(tasks) == 0 {
		fmt.Println("No tasks.")
		return
	}

	// Find top-level tasks
	childrenMap := buildChildrenMap(tasks)
	topLevel := findTopLevel(tasks)

	for _, task := range topLevel {
		renderTaskRow(task, 0)
		renderChildren(tasks, childrenMap, task.ID, 1)
	}
}

// RenderListReady renders ready tasks with their descendants.
func RenderListReady(tasks []schema.Task, idPrefix string) {
	if len(tasks) == 0 {
		fmt.Println("No tasks.")
		return
	}

	// Find ready tasks
	childrenMap := buildChildrenMap(tasks)
	readyTasks := filterByStatus(tasks, schema.StatusReady)

	if len(readyTasks) == 0 {
		fmt.Println("No ready tasks.")
		return
	}

	// Track which tasks have been rendered to avoid duplicates
	rendered := make(map[string]bool)

	for _, task := range readyTasks {
		if rendered[task.ID] {
			continue
		}
		renderTaskRow(task, 0)
		rendered[task.ID] = true
		renderChildrenWithTracking(tasks, childrenMap, task.ID, 1, rendered)
	}
}

// RenderShow renders full task details.
func RenderShow(task schema.Task, allTasks []schema.Task, events []schema.Event) {
	fmt.Printf("%s\n", task.ID)
	fmt.Printf("Title: %s\n", task.Title)
	fmt.Printf("Status: %s\n", task.Status)
	fmt.Printf("Priority: %s\n", task.Priority)
	fmt.Printf("Assignee: %s\n", formatAssignee(task.Assignee))
	fmt.Printf("Parent: %s\n", formatParent(task.ParentID))
	fmt.Println("Description:")
	if task.Description != "" {
		fmt.Println(task.Description)
	}
	fmt.Println()

	// Render subtasks
	children := domain.DirectChildren(allTasks, task.ID)
	if len(children) > 0 {
		fmt.Println("Subtasks:")
		for _, child := range children {
			fmt.Printf("- %s %s %s\n", child.ID, child.Status, child.Title)
		}
		fmt.Println()
	}

	// Render recent events
	if len(events) > 0 {
		fmt.Println("Recent events:")
		for _, event := range events {
			fmt.Printf("- %s by %s at %s\n", event.Type, event.Actor, event.Timestamp)
		}
	}
}

func renderTaskRow(task schema.Task, depth int) {
	indent := strings.Repeat("  ", depth)
	assignee := formatAssignee(task.Assignee)
	fmt.Fprintf(os.Stderr, "%s%s  %-6s  %-12s  %-8s  %s\n",
		indent, task.ID, task.Priority, task.Status, assignee, task.Title)
}

func renderChildren(tasks []schema.Task, childrenMap map[string][]string, parentID string, depth int) {
	children := childrenMap[parentID]
	if len(children) == 0 {
		return
	}

	// Sort children by ID
	sort.Strings(children)

	// Find task objects for IDs
	taskMap := buildTaskMap(tasks)
	for _, childID := range children {
		if task, ok := taskMap[childID]; ok {
			renderTaskRow(task, depth)
			renderChildren(tasks, childrenMap, childID, depth+1)
		}
	}
}

func renderChildrenWithTracking(tasks []schema.Task, childrenMap map[string][]string, parentID string, depth int, rendered map[string]bool) {
	children := childrenMap[parentID]
	if len(children) == 0 {
		return
	}

	// Sort children by ID
	sort.Strings(children)

	taskMap := buildTaskMap(tasks)
	for _, childID := range children {
		if rendered[childID] {
			continue
		}
		if task, ok := taskMap[childID]; ok {
			renderTaskRow(task, depth)
			rendered[childID] = true
			renderChildrenWithTracking(tasks, childrenMap, childID, depth+1, rendered)
		}
	}
}

func buildChildrenMap(tasks []schema.Task) map[string][]string {
	children := make(map[string][]string)
	for _, task := range tasks {
		if task.ParentID != nil {
			parent := *task.ParentID
			children[parent] = append(children[parent], task.ID)
		}
	}
	return children
}

func buildTaskMap(tasks []schema.Task) map[string]schema.Task {
	m := make(map[string]schema.Task)
	for _, task := range tasks {
		m[task.ID] = task
	}
	return m
}

func findTopLevel(tasks []schema.Task) []schema.Task {
	var top []schema.Task
	for _, task := range tasks {
		if task.ParentID == nil {
			top = append(top, task)
		}
	}
	return top
}

func filterByStatus(tasks []schema.Task, status schema.Status) []schema.Task {
	var result []schema.Task
	for _, task := range tasks {
		if task.Status == status {
			result = append(result, task)
		}
	}
	return result
}

func formatAssignee(assignee *schema.Actor) string {
	if assignee == nil {
		return "-"
	}
	return string(*assignee)
}

func formatParent(parentID *string) string {
	if parentID == nil {
		return "-"
	}
	return *parentID
}
