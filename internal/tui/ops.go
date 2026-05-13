package tui

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ilova-bazis/umati/internal/domain"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/store"
	"github.com/ilova-bazis/umati/internal/workspace"
)

const refreshInterval = 2 * time.Second

var skipDirs = map[string]bool{
	".git": true, ".umati": true,
	"node_modules": true, "vendor": true,
}

func listWorkspaceFiles(root string) []string {
	var files []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && skipDirs[d.Name()] {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel(root, path)
			files = append(files, strings.ReplaceAll(rel, "\\", "/"))
		}
		return nil
	})
	return files
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func loadTasksCmd(ctx workspace.Context) tea.Cmd {
	return func() tea.Msg {
		tasks, err := store.ListTasks(ctx)
		if err != nil {
			return errMsg{err}
		}
		return tasksLoadedMsg{tasks}
	}
}

func loadEventsCmd(ctx workspace.Context, taskID string) tea.Cmd {
	return func() tea.Msg {
		events, err := store.ReadEventsForTask(ctx, taskID, 10)
		if err != nil {
			return eventsLoadedMsg{taskID: taskID, events: nil}
		}
		return eventsLoadedMsg{taskID: taskID, events: events}
	}
}

func claimTaskCmd(ctx workspace.Context, taskID string, agent schema.Actor) tea.Cmd {
	return func() tea.Msg {
		err := store.WithLockAndEvents(ctx, agent, "claim", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
			task, err := store.ReadTask(ctx, taskID)
			if err != nil {
				return nil, err
			}
			if err := domain.CanTransition(task.Status, schema.StatusClaimed); err != nil {
				return nil, err
			}
			if err := domain.ValidateAgentMatch(task, agent, schema.StatusClaimed); err != nil {
				return nil, err
			}
			if task.Status != schema.StatusReady {
				return nil, fmt.Errorf("task is not ready (status: %s)", task.Status)
			}
			if task.Assignee != nil {
				return nil, fmt.Errorf("task is already claimed by %s", *task.Assignee)
			}
			now := schema.NowTimestamp()
			task.Status = schema.StatusClaimed
			task.Assignee = &agent
			task.UpdatedAt = now
			task.UpdatedBy = agent
			if err := store.WriteTask(ctx, task); err != nil {
				return nil, err
			}
			return []schema.Event{{TaskID: taskID, Type: schema.EventClaimed, Actor: agent, Timestamp: now, Meta: map[string]any{}}}, nil
		})
		if err != nil {
			return errMsg{err}
		}
		return mutationDoneMsg{taskID: taskID, command: "claim"}
	}
}

func startTaskCmd(ctx workspace.Context, taskID string, agent schema.Actor) tea.Cmd {
	return func() tea.Msg {
		err := store.WithLockAndEvents(ctx, agent, "start", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
			task, err := store.ReadTask(ctx, taskID)
			if err != nil {
				return nil, err
			}
			if err := domain.CanTransition(task.Status, schema.StatusInProgress); err != nil {
				return nil, err
			}
			if err := domain.ValidateAgentMatch(task, agent, schema.StatusInProgress); err != nil {
				return nil, err
			}
			now := schema.NowTimestamp()
			task.Status = schema.StatusInProgress
			task.UpdatedAt = now
			task.UpdatedBy = agent
			if err := store.WriteTask(ctx, task); err != nil {
				return nil, err
			}
			return []schema.Event{{TaskID: taskID, Type: schema.EventStarted, Actor: agent, Timestamp: now, Meta: map[string]any{}}}, nil
		})
		if err != nil {
			return errMsg{err}
		}
		return mutationDoneMsg{taskID: taskID, command: "start"}
	}
}

func pauseTaskCmd(ctx workspace.Context, taskID string, agent schema.Actor) tea.Cmd {
	return func() tea.Msg {
		err := store.WithLockAndEvents(ctx, agent, "pause", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
			task, err := store.ReadTask(ctx, taskID)
			if err != nil {
				return nil, err
			}
			if err := domain.CanTransition(task.Status, schema.StatusPaused); err != nil {
				return nil, err
			}
			if err := domain.ValidateAgentMatch(task, agent, schema.StatusPaused); err != nil {
				return nil, err
			}
			now := schema.NowTimestamp()
			task.Status = schema.StatusPaused
			task.Assignee = nil
			task.UpdatedAt = now
			task.UpdatedBy = agent
			if err := store.WriteTask(ctx, task); err != nil {
				return nil, err
			}
			return []schema.Event{{TaskID: taskID, Type: schema.EventPaused, Actor: agent, Timestamp: now, Meta: map[string]any{}}}, nil
		})
		if err != nil {
			return errMsg{err}
		}
		return mutationDoneMsg{taskID: taskID, command: "pause"}
	}
}

func releaseTaskCmd(ctx workspace.Context, taskID string, agent schema.Actor) tea.Cmd {
	return func() tea.Msg {
		err := store.WithLockAndEvents(ctx, agent, "release", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
			task, err := store.ReadTask(ctx, taskID)
			if err != nil {
				return nil, err
			}
			if err := domain.CanTransition(task.Status, schema.StatusReady); err != nil {
				return nil, err
			}
			if err := domain.ValidateAgentMatch(task, agent, schema.StatusReady); err != nil {
				return nil, err
			}
			now := schema.NowTimestamp()
			task.Status = schema.StatusReady
			task.Assignee = nil
			task.UpdatedAt = now
			task.UpdatedBy = agent
			if err := store.WriteTask(ctx, task); err != nil {
				return nil, err
			}
			return []schema.Event{{TaskID: taskID, Type: schema.EventReleased, Actor: agent, Timestamp: now, Meta: map[string]any{}}}, nil
		})
		if err != nil {
			return errMsg{err}
		}
		return mutationDoneMsg{taskID: taskID, command: "release"}
	}
}

func completeTaskCmd(ctx workspace.Context, taskID string, agent schema.Actor) tea.Cmd {
	return func() tea.Msg {
		err := store.WithLockAndEvents(ctx, agent, "complete", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
			task, err := store.ReadTask(ctx, taskID)
			if err != nil {
				return nil, err
			}
			allTasks, err := store.ListTasks(ctx)
			if err != nil {
				return nil, err
			}
			if err := domain.CanTransition(task.Status, schema.StatusDone); err != nil {
				return nil, err
			}
			if err := domain.ValidateAgentMatch(task, agent, schema.StatusDone); err != nil {
				return nil, err
			}
			if task.Status != schema.StatusInProgress {
				return nil, fmt.Errorf("task is not in progress (status: %s)", task.Status)
			}
			if err := domain.ValidateParentCompletion(allTasks, taskID); err != nil {
				return nil, err
			}
			now := schema.NowTimestamp()
			task.Status = schema.StatusDone
			task.Assignee = nil
			task.UpdatedAt = now
			task.UpdatedBy = agent
			if err := store.WriteTask(ctx, task); err != nil {
				return nil, err
			}
			return []schema.Event{{TaskID: taskID, Type: schema.EventCompleted, Actor: agent, Timestamp: now, Meta: map[string]any{}}}, nil
		})
		if err != nil {
			return errMsg{err}
		}
		return mutationDoneMsg{taskID: taskID, command: "complete"}
	}
}

func deleteTaskCmd(ctx workspace.Context, taskID string, agent schema.Actor) tea.Cmd {
	return func() tea.Msg {
		err := store.WithLockAndEvents(ctx, agent, "delete", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
			allTasks, err := store.ListTasks(ctx)
			if err != nil {
				return nil, err
			}
			if err := domain.ValidateDeleteEligibility(allTasks, taskID); err != nil {
				return nil, err
			}
			subtree, err := domain.GetSubtree(allTasks, taskID)
			if err != nil {
				return nil, err
			}
			now := schema.NowTimestamp()
			var events []schema.Event
			for _, task := range subtree {
				deleted := task
				deleted.Status = schema.StatusDeleted
				deleted.UpdatedAt = now
				deleted.UpdatedBy = agent
				deletedAt := now
				deleted.DeletedAt = &deletedAt
				deleted.DeletedBy = &agent
				if err := store.WriteDeletedTask(ctx, deleted); err != nil {
					return nil, err
				}
				activePath, err := store.ActiveTaskPath(ctx, task.ID)
				if err != nil {
					return nil, err
				}
				if err := os.Remove(activePath); err != nil && !os.IsNotExist(err) {
					return nil, err
				}
				events = append(events, schema.Event{TaskID: task.ID, Type: schema.EventDeleted, Actor: agent, Timestamp: now, Meta: map[string]any{}})
			}
			return events, nil
		})
		if err != nil {
			return errMsg{err}
		}
		return mutationDoneMsg{taskID: taskID, command: "delete"}
	}
}

func createTaskCmd(ctx workspace.Context, agent schema.Actor, r formResult) tea.Cmd {
	return func() tea.Msg {
		var newTaskID string
		err := store.WithLockAndEvents(ctx, agent, "create", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
			// Validate parent if specified
			var parentID *string
			if r.parentID != "" {
				if _, err := store.ReadTask(ctx, r.parentID); err != nil {
					return nil, fmt.Errorf("parent task not found: %s", r.parentID)
				}
				parentID = &r.parentID
			}
			if r.description == "" && r.status != schema.StatusDraft {
				return nil, fmt.Errorf("description is required unless status is draft")
			}
			id, err := store.NextTaskID(ctx, cfg)
			if err != nil {
				return nil, err
			}
			newTaskID = id
			now := schema.NowTimestamp()

			var assignee *schema.Actor
			if r.assignee != "" {
				a := schema.Actor(r.assignee)
				assignee = &a
			}

			task := schema.Task{
				ID:          id,
				Title:       r.title,
				Description: r.description,
				Priority:    r.priority,
				Status:      r.status,
				Assignee:    assignee,
				ParentID:    parentID,
				Files:       r.files,
				CreatedAt:   now,
				UpdatedAt:   now,
				CreatedBy:   agent,
				UpdatedBy:   agent,
			}
			if err := store.WriteTask(ctx, task); err != nil {
				return nil, err
			}
			return []schema.Event{{TaskID: id, Type: schema.EventCreated, Actor: agent, Timestamp: now, Meta: map[string]any{}}}, nil
		})
		if err != nil {
			return errMsg{err}
		}
		return mutationDoneMsg{taskID: newTaskID, command: "create"}
	}
}

func updateTaskCmd(ctx workspace.Context, agent schema.Actor, taskID string, r formResult) tea.Cmd {
	return func() tea.Msg {
		err := store.WithLockAndEvents(ctx, agent, "update", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
			task, err := store.ReadTask(ctx, taskID)
			if err != nil {
				return nil, err
			}

			now := schema.NowTimestamp()
			updated := false

			if r.title != "" && r.title != task.Title {
				task.Title = r.title
				updated = true
			}
			if r.description != task.Description {
				task.Description = r.description
				updated = true
			}
			if r.priority != task.Priority {
				task.Priority = r.priority
				updated = true
			}
			if r.status != task.Status {
				allTasks, err := store.ListTasks(ctx)
				if err != nil {
					return nil, err
				}
				if err := domain.CanTransition(task.Status, r.status); err != nil {
					return nil, err
				}
				if err := domain.ValidateAgentMatch(task, agent, r.status); err != nil {
					return nil, err
				}
				if r.status == schema.StatusDone {
					if err := domain.ValidateParentCompletion(allTasks, taskID); err != nil {
						return nil, err
					}
				}
				task.Status = r.status
				updated = true
			}
			if r.parentID == "" {
				if task.ParentID != nil {
					task.ParentID = nil
					updated = true
				}
			} else if task.ParentID == nil || *task.ParentID != r.parentID {
				if _, err := store.ReadTask(ctx, r.parentID); err != nil {
					return nil, fmt.Errorf("parent task not found: %s", r.parentID)
				}
				task.ParentID = &r.parentID
				updated = true
			}

			if r.assignee == "" {
				if task.Assignee != nil {
					task.Assignee = nil
					updated = true
				}
			} else {
				a := schema.Actor(r.assignee)
				if task.Assignee == nil || *task.Assignee != a {
					task.Assignee = &a
					updated = true
				}
			}

			// Always update files (empty slice clears them)
			task.Files = r.files
			updated = true

			if !updated {
				return nil, nil
			}

			task.UpdatedAt = now
			task.UpdatedBy = agent
			if err := store.WriteTask(ctx, task); err != nil {
				return nil, err
			}
			return []schema.Event{{TaskID: taskID, Type: schema.EventUpdated, Actor: agent, Timestamp: now, Meta: map[string]any{}}}, nil
		})
		if err != nil {
			return errMsg{err}
		}
		return mutationDoneMsg{taskID: taskID, command: "update"}
	}
}
