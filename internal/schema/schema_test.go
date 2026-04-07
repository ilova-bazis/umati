package schema_test

import (
	"testing"

	"github.com/ilova-bazis/umati/internal/schema"
)

func TestValidateConfig(t *testing.T) {
	cfg := schema.Config{SchemaVersion: 1, IDPrefix: "UM", CreatedAt: "2026-04-03T12:00:00Z"}
	if err := schema.ValidateConfig(cfg); err != nil {
		t.Fatalf("ValidateConfig() error = %v", err)
	}
}

func TestValidateActiveTaskRejectsEmptyDescriptionOnNonDraft(t *testing.T) {
	task := validTask()
	task.Description = ""
	task.Status = schema.StatusReady
	if err := schema.ValidateActiveTask(task); err == nil {
		t.Fatal("ValidateActiveTask() expected error for empty non-draft description")
	}
}

func TestValidateDeletedTask(t *testing.T) {
	task := validTask()
	task.Status = schema.StatusDeleted
	deletedAt := "2026-04-03T13:00:00Z"
	deletedBy := schema.ActorHuman
	task.DeletedAt = &deletedAt
	task.DeletedBy = &deletedBy
	if err := schema.ValidateDeletedTask(task); err != nil {
		t.Fatalf("ValidateDeletedTask() error = %v", err)
	}
}

func TestCompareTaskIDsNumeric(t *testing.T) {
	cmp, err := schema.CompareTaskIDs("UM-9", "UM-10")
	if err != nil {
		t.Fatalf("CompareTaskIDs() error = %v", err)
	}
	if cmp >= 0 {
		t.Fatalf("expected UM-9 to sort before UM-10, got %d", cmp)
	}
}

func TestValidateEventRequiresMeta(t *testing.T) {
	event := schema.Event{TaskID: "UM-1", Type: schema.EventCreated, Actor: schema.ActorHuman, Timestamp: "2026-04-03T12:00:00Z"}
	if err := schema.ValidateEvent(event); err == nil {
		t.Fatal("ValidateEvent() expected error for missing meta")
	}
}

func validTask() schema.Task {
	return schema.Task{
		ID:          "UM-1",
		Title:       "Task",
		Description: "Valid task.",
		Priority:    schema.PriorityMedium,
		Status:      schema.StatusDraft,
		Assignee:    nil,
		ParentID:    nil,
		CreatedAt:   "2026-04-03T12:00:00Z",
		UpdatedAt:   "2026-04-03T12:00:00Z",
		CreatedBy:   schema.ActorHuman,
		UpdatedBy:   schema.ActorHuman,
	}
}
