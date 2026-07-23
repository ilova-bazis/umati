package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilova-bazis/umati/internal/schema"
)

func TestDescriptionEnterInsertsNewline(t *testing.T) {
	form := newCreateForm(schema.StatusDraft, nil)
	form.focus = fieldDescription
	form.syncFocus()
	form.descInput.SetValue("first line")

	form, _ = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if got := form.descInput.Value(); got != "first line\n" {
		t.Fatalf("description after enter = %q, want newline", got)
	}
}

func TestEditFormPreservesExistingDescriptionAboveInputLimit(t *testing.T) {
	description := strings.Repeat("界", schema.MaxDescriptionInputLength+1) + "\nsecond line"
	task := schema.Task{
		ID:          "UM-1",
		Title:       "Original title",
		Description: description,
		Priority:    schema.PriorityMedium,
		Status:      schema.StatusReady,
	}

	form := newEditForm(task, nil)
	if got := form.descInput.Value(); got != description {
		t.Fatal("newEditForm() truncated the existing description")
	}

	form.titleInput.SetValue("Updated title")
	form, cmd := form.submit()
	if form.errMsg != "" {
		t.Fatalf("submit() error = %q", form.errMsg)
	}
	if cmd == nil {
		t.Fatal("submit() returned no command")
	}

	rawMsg := cmd()
	msg, ok := rawMsg.(formSubmittedMsg)
	if !ok {
		t.Fatalf("submit() message type = %T, want formSubmittedMsg", rawMsg)
	}
	if msg.result.description != description {
		t.Fatal("submit() changed an untouched existing description")
	}
}

func TestEditFormRejectsChangedDescriptionAboveInputLimit(t *testing.T) {
	task := schema.Task{
		ID:          "UM-1",
		Title:       "Task",
		Description: "Original description",
		Priority:    schema.PriorityMedium,
		Status:      schema.StatusReady,
	}

	form := newEditForm(task, nil)
	form.descInput.SetValue(strings.Repeat("a", schema.MaxDescriptionInputLength+1))
	form, cmd := form.submit()
	if cmd != nil {
		t.Fatal("submit() should not submit an oversized changed description")
	}
	want := "description exceeds 10000 characters"
	if form.errMsg != want {
		t.Fatalf("submit() error = %q, want %q", form.errMsg, want)
	}
}
