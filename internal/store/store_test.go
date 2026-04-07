package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ilova-bazis/umati/internal/domain"
	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/store"
	"github.com/ilova-bazis/umati/internal/workspace"
)

func TestReadTask(t *testing.T) {
	ctx := mustDiscover(t, filepath.Join("..", "..", "testdata", "workspaces", "flat"))
	task, err := store.ReadTask(ctx, "UM-2")
	if err != nil {
		t.Fatalf("ReadTask() error = %v", err)
	}
	if task.ID != "UM-2" {
		t.Fatalf("expected UM-2, got %s", task.ID)
	}
}

func TestReadTaskNotFound(t *testing.T) {
	ctx := mustDiscover(t, filepath.Join("..", "..", "testdata", "workspaces", "flat"))
	_, err := store.ReadTask(ctx, "UM-999")
	if !errs.IsKind(err, errs.KindTaskNotFound) {
		t.Fatalf("expected task not found error, got %v", err)
	}
}

func TestListTasksSortedByNumericID(t *testing.T) {
	ctx := mustDiscover(t, filepath.Join("..", "..", "testdata", "workspaces", "flat"))
	tasks, err := store.ListTasks(ctx)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	got := []string{tasks[0].ID, tasks[1].ID, tasks[2].ID}
	want := []string{"UM-2", "UM-9", "UM-10"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("task order mismatch: got %v want %v", got, want)
		}
	}
}

func TestWriteTaskAndAppendEvent(t *testing.T) {
	ctx := workspace.NewContext(t.TempDir())
	task := schema.Task{
		ID:          "UM-1",
		Title:       "Task",
		Description: "Valid task.",
		Priority:    schema.PriorityMedium,
		Status:      schema.StatusReady,
		CreatedAt:   "2026-04-03T12:00:00Z",
		UpdatedAt:   "2026-04-03T12:00:00Z",
		CreatedBy:   schema.ActorHuman,
		UpdatedBy:   schema.ActorHuman,
	}
	if err := store.WriteTask(ctx, task); err != nil {
		t.Fatalf("WriteTask() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(ctx.TasksDir, "UM-1.json")); err != nil {
		t.Fatalf("expected task file to exist: %v", err)
	}
	event := schema.Event{TaskID: "UM-1", Type: schema.EventCreated, Actor: schema.ActorHuman, Timestamp: "2026-04-03T12:00:00Z", Meta: map[string]any{}}
	if err := store.AppendEvent(ctx, event); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}
	if _, err := os.Stat(ctx.EventsPath); err != nil {
		t.Fatalf("expected events file to exist: %v", err)
	}
}

func TestWriteDeletedTask(t *testing.T) {
	ctx := workspace.NewContext(t.TempDir())
	deletedAt := "2026-04-03T13:00:00Z"
	deletedBy := schema.ActorHuman
	task := schema.Task{
		ID:          "UM-7",
		Title:       "Archived task",
		Description: "Archived task.",
		Priority:    schema.PriorityLow,
		Status:      schema.StatusDeleted,
		CreatedAt:   "2026-04-03T12:00:00Z",
		UpdatedAt:   "2026-04-03T13:00:00Z",
		CreatedBy:   schema.ActorHuman,
		UpdatedBy:   schema.ActorHuman,
		DeletedAt:   &deletedAt,
		DeletedBy:   &deletedBy,
	}
	if err := store.WriteDeletedTask(ctx, task); err != nil {
		t.Fatalf("WriteDeletedTask() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(ctx.DeletedDir, "UM-7.json")); err != nil {
		t.Fatalf("expected deleted task file to exist: %v", err)
	}
}

func TestNextTaskIDIncludesDeleted(t *testing.T) {
	ctx := mustDiscover(t, filepath.Join("..", "..", "testdata", "workspaces", "deleted"))
	cfg, err := workspace.LoadConfig(ctx)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	id, err := store.NextTaskID(ctx, cfg)
	if err != nil {
		t.Fatalf("NextTaskID() error = %v", err)
	}
	if id != "UM-8" {
		t.Fatalf("expected next id UM-8, got %s", id)
	}
}

func TestDescendants(t *testing.T) {
	ctx := mustDiscover(t, filepath.Join("..", "..", "testdata", "workspaces", "nested"))
	tasks, err := store.ListTasks(ctx)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	descendants := domain.Descendants(tasks, "UM-1")
	if len(descendants) != 2 {
		t.Fatalf("expected 2 descendants, got %d", len(descendants))
	}
}

func mustDiscover(t *testing.T, start string) workspace.Context {
	t.Helper()
	ctx, err := workspace.Discover(start)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	return ctx
}
