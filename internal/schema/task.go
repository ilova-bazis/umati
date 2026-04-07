package schema

import "fmt"

type Task struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Status      Status   `json:"status"`
	Assignee    *Actor   `json:"assignee"`
	ParentID    *string  `json:"parent_id"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	CreatedBy   Actor    `json:"created_by"`
	UpdatedBy   Actor    `json:"updated_by"`
	DeletedAt   *string  `json:"deleted_at,omitempty"`
	DeletedBy   *Actor   `json:"deleted_by,omitempty"`
}

func ValidateActiveTask(task Task) error {
	if err := validateBaseTask(task); err != nil {
		return err
	}
	if !IsValidActiveStatus(task.Status) {
		return fmt.Errorf("invalid active status: %q", task.Status)
	}
	if task.DeletedAt != nil || task.DeletedBy != nil {
		return fmt.Errorf("active task must not include deleted metadata")
	}
	return nil
}

func ValidateDeletedTask(task Task) error {
	if err := validateBaseTask(task); err != nil {
		return err
	}
	if task.Status != StatusDeleted {
		return fmt.Errorf("deleted task status must be %q", StatusDeleted)
	}
	if task.DeletedAt == nil || *task.DeletedAt == "" {
		return fmt.Errorf("deleted_at is required")
	}
	if err := ValidateTimestamp(*task.DeletedAt); err != nil {
		return fmt.Errorf("deleted_at: %w", err)
	}
	if task.DeletedBy == nil || !IsValidActor(*task.DeletedBy) {
		return fmt.Errorf("deleted_by is required and must be valid")
	}
	return nil
}

func validateBaseTask(task Task) error {
	if _, err := ParseTaskID(task.ID); err != nil {
		return err
	}
	if task.Title == "" {
		return fmt.Errorf("title is required")
	}
	if task.Description == "" && task.Status != StatusDraft {
		return fmt.Errorf("description is required unless status is draft")
	}
	if !IsValidPriority(task.Priority) {
		return fmt.Errorf("invalid priority: %q", task.Priority)
	}
	if task.Assignee != nil && !IsValidActor(*task.Assignee) {
		return fmt.Errorf("invalid assignee: %q", *task.Assignee)
	}
	if task.ParentID != nil {
		if _, err := ParseTaskID(*task.ParentID); err != nil {
			return fmt.Errorf("invalid parent_id: %w", err)
		}
	}
	if !IsValidActor(task.CreatedBy) {
		return fmt.Errorf("invalid created_by: %q", task.CreatedBy)
	}
	if !IsValidActor(task.UpdatedBy) {
		return fmt.Errorf("invalid updated_by: %q", task.UpdatedBy)
	}
	if err := ValidateTimestamp(task.CreatedAt); err != nil {
		return fmt.Errorf("created_at: %w", err)
	}
	if err := ValidateTimestamp(task.UpdatedAt); err != nil {
		return fmt.Errorf("updated_at: %w", err)
	}
	return nil
}
