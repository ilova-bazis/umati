package domain

import (
	"fmt"

	"github.com/ilova-bazis/umati/internal/schema"
)

// CanTransition checks if a status transition is valid.
// Returns nil if valid, error otherwise.
func CanTransition(from, to schema.Status) error {
	// Define valid transitions
	allowed := map[schema.Status][]schema.Status{
		schema.StatusDraft:      {schema.StatusReady, schema.StatusPaused, schema.StatusCancelled},
		schema.StatusPaused:     {schema.StatusReady, schema.StatusCancelled},
		schema.StatusReady:      {schema.StatusClaimed, schema.StatusCancelled},
		schema.StatusClaimed:    {schema.StatusInProgress, schema.StatusPaused, schema.StatusReady, schema.StatusCancelled},
		schema.StatusInProgress: {schema.StatusPaused, schema.StatusReady, schema.StatusDone, schema.StatusCancelled},
		schema.StatusDone:       {},
		schema.StatusCancelled:  {},
		schema.StatusDeleted:    {},
	}

	validTargets, ok := allowed[from]
	if !ok {
		return fmt.Errorf("unknown source status: %s", from)
	}

	for _, target := range validTargets {
		if target == to {
			return nil
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", from, to)
}

// RequiresAgentMatch returns true if the given status requires the acting agent
// to match the task's assignee for the specified target status.
func RequiresAgentMatch(current, target schema.Status) bool {
	// Agent must match when:
	// - Starting work (claimed -> in_progress)
	// - Pausing active work (claimed -> paused, in_progress -> paused)
	// - Releasing work (claimed -> ready, in_progress -> ready)
	// - Completing work (in_progress -> done)
	// - Cancelling claimed/in_progress work
	switch current {
	case schema.StatusClaimed:
		switch target {
		case schema.StatusInProgress, schema.StatusPaused, schema.StatusReady, schema.StatusCancelled:
			return true
		}
	case schema.StatusInProgress:
		switch target {
		case schema.StatusPaused, schema.StatusReady, schema.StatusDone, schema.StatusCancelled:
			return true
		}
	}
	return false
}

// IsTerminalStatus returns true if the status is terminal (no further transitions allowed).
func IsTerminalStatus(status schema.Status) bool {
	return status == schema.StatusDone || status == schema.StatusCancelled || status == schema.StatusDeleted
}

// ValidateAgentMatch checks if the acting agent matches the task's assignee.
// Returns nil if valid or no match required, error otherwise.
func ValidateAgentMatch(task schema.Task, actor schema.Actor, targetStatus schema.Status) error {
	if !RequiresAgentMatch(task.Status, targetStatus) {
		return nil
	}

	if task.Assignee == nil {
		return fmt.Errorf("task has no assignee")
	}

	if *task.Assignee != actor {
		return fmt.Errorf("task is assigned to %s, not %s", *task.Assignee, actor)
	}

	return nil
}

// ValidateParentCompletion checks if a parent task can be completed.
// All descendants must be done or cancelled.
func ValidateParentCompletion(tasks []schema.Task, parentID string) error {
	descendants := Descendants(tasks, parentID)

	var unfinished []string
	for _, desc := range descendants {
		if desc.Status != schema.StatusDone && desc.Status != schema.StatusCancelled {
			unfinished = append(unfinished, fmt.Sprintf("%s (%s)", desc.ID, desc.Status))
		}
	}

	if len(unfinished) > 0 {
		return fmt.Errorf("cannot complete: unfinished descendants: %v", unfinished)
	}

	return nil
}

// ValidateDeleteEligibility checks if a task and its descendants can be deleted.
// No task in the subtree may be claimed or in_progress.
func ValidateDeleteEligibility(tasks []schema.Task, rootID string) error {
	// Find the root task
	var root *schema.Task
	for i := range tasks {
		if tasks[i].ID == rootID {
			root = &tasks[i]
			break
		}
	}
	if root == nil {
		return fmt.Errorf("task not found: %s", rootID)
	}

	// Check root
	if root.Status == schema.StatusClaimed || root.Status == schema.StatusInProgress {
		return fmt.Errorf("cannot delete %s: status is %s", root.ID, root.Status)
	}

	// Check all descendants
	descendants := Descendants(tasks, rootID)
	var blocked []string
	for _, desc := range descendants {
		if desc.Status == schema.StatusClaimed || desc.Status == schema.StatusInProgress {
			blocked = append(blocked, fmt.Sprintf("%s (%s)", desc.ID, desc.Status))
		}
	}

	if len(blocked) > 0 {
		return fmt.Errorf("cannot delete: active descendants: %v", blocked)
	}

	return nil
}

// GetSubtree returns the root task and all its descendants.
func GetSubtree(tasks []schema.Task, rootID string) ([]schema.Task, error) {
	var result []schema.Task

	// Find root
	var root *schema.Task
	for i := range tasks {
		if tasks[i].ID == rootID {
			root = &tasks[i]
			break
		}
	}
	if root == nil {
		return nil, fmt.Errorf("task not found: %s", rootID)
	}

	result = append(result, *root)
	result = append(result, Descendants(tasks, rootID)...)

	return result, nil
}
