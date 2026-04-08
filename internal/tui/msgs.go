package tui

import "github.com/ilova-bazis/umati/internal/schema"

type tasksLoadedMsg struct {
	tasks []schema.Task
}

type errMsg struct {
	err error
}

type mutationDoneMsg struct {
	taskID  string
	command string
}

type clearStatusMsg struct{}

type eventsLoadedMsg struct {
	taskID string
	events []schema.Event
}

type formSubmittedMsg struct {
	result formResult
}

type formCancelledMsg struct{}

type filterAppliedMsg struct {
	priority *schema.Priority
	agent    *schema.Actor
}

type filterCancelledMsg struct{}
