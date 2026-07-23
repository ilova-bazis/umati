package schema_test

import (
	"strings"
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

func TestValidateDescriptionInput(t *testing.T) {
	if err := schema.ValidateDescriptionInput(strings.Repeat("a", schema.MaxDescriptionInputLength)); err != nil {
		t.Fatalf("ValidateDescriptionInput() at limit error = %v", err)
	}
	if err := schema.ValidateDescriptionInput(strings.Repeat("a", schema.MaxDescriptionInputLength+1)); err == nil {
		t.Fatal("ValidateDescriptionInput() expected error above limit")
	}
	if err := schema.ValidateDescriptionInput(strings.Repeat("界", schema.MaxDescriptionInputLength)); err != nil {
		t.Fatalf("ValidateDescriptionInput() should count Unicode characters, error = %v", err)
	}
}

func TestValidateActiveTaskAllowsDescriptionAboveInputLimit(t *testing.T) {
	task := validTask()
	task.Description = strings.Repeat("a", schema.MaxDescriptionInputLength+1)
	if err := schema.ValidateActiveTask(task); err != nil {
		t.Fatalf("ValidateActiveTask() should allow existing long descriptions, error = %v", err)
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

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"nil", nil, nil},
		{"empty", []string{}, nil},
		{"single", []string{"bug"}, []string{"bug"}},
		{"lowercase", []string{"BUG", "Api"}, []string{"bug", "api"}},
		{"deduplicate", []string{"bug", "BUG", "bug"}, []string{"bug"}},
		{"trimAndDropEmpty", []string{" bug ", "", "api "}, []string{"bug", "api"}},
		{"mixed", []string{"  Bug ", "", "API", "bug"}, []string{"bug", "api"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schema.NormalizeTags(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("NormalizeTags() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("NormalizeTags() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestValidateTagsRejectsNonCanonical(t *testing.T) {
	tests := []struct {
		name string
		tags []string
	}{
		{"empty string", []string{"bug", "", "api"}},
		{"mixed case", []string{"Bug"}},
		{"duplicate", []string{"bug", "bug"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := validTask()
			task.Tags = tt.tags
			if err := schema.ValidateActiveTask(task); err == nil {
				t.Fatalf("expected validation error for tags %v", tt.tags)
			}
		})
	}
}

func TestValidateTagsAllowsCanonical(t *testing.T) {
	task := validTask()
	task.Tags = []string{"bug", "api", "backend"}
	if err := schema.ValidateActiveTask(task); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestIsValidKind(t *testing.T) {
	valid := []schema.Kind{
		schema.KindTask, schema.KindBug, schema.KindFeature,
		schema.KindChore, schema.KindImprovement, schema.KindDastan,
	}
	for _, k := range valid {
		if !schema.IsValidKind(k) {
			t.Errorf("IsValidKind(%q) = false, want true", k)
		}
	}
	if schema.IsValidKind("epic") {
		t.Errorf("IsValidKind(\"epic\") = true, want false")
	}
}

func TestValidateActiveTaskRejectsInvalidKind(t *testing.T) {
	task := validTask()
	task.Kind = schema.Kind("epic")
	if err := schema.ValidateActiveTask(task); err == nil {
		t.Fatal("expected validation error for invalid kind")
	}
}

func TestKindDisplay(t *testing.T) {
	if schema.KindDisplay("") != "task" {
		t.Errorf("KindDisplay(\"\") = %q, want \"task\"", schema.KindDisplay(""))
	}
	if schema.KindDisplay("bug") != "bug" {
		t.Errorf("KindDisplay(\"bug\") = %q, want \"bug\"", schema.KindDisplay("bug"))
	}
}

func TestHasAllTags(t *testing.T) {
	tests := []struct {
		name       string
		taskTags   []string
		filterTags []string
		want       bool
	}{
		{"all match", []string{"bug", "api"}, []string{"bug"}, true},
		{"all match multiple", []string{"bug", "api", "backend"}, []string{"bug", "api"}, true},
		{"partial match", []string{"bug"}, []string{"bug", "api"}, false},
		{"no match", []string{"bug"}, []string{"api"}, false},
		{"empty filter", []string{"bug"}, nil, true},
		{"empty tasks", nil, []string{"bug"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := schema.HasAllTags(tt.taskTags, tt.filterTags); got != tt.want {
				t.Fatalf("HasAllTags(%v, %v) = %v, want %v", tt.taskTags, tt.filterTags, got, tt.want)
			}
		})
	}
}
