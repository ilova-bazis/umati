package store

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/workspace"
)

func AppendEvent(ctx workspace.Context, event schema.Event) error {
	if err := schema.ValidateEvent(event); err != nil {
		return errs.E(errs.KindInvalidEvent, "store.AppendEvent", ctx.EventsPath, err)
	}
	if err := os.MkdirAll(ctx.EventsDir, 0o755); err != nil {
		return errs.E(errs.KindInvalidPath, "store.AppendEvent", ctx.EventsDir, err)
	}
	file, err := os.OpenFile(ctx.EventsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return errs.E(errs.KindInvalidPath, "store.AppendEvent", ctx.EventsPath, err)
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return errs.E(errs.KindInvalidEvent, "store.AppendEvent", ctx.EventsPath, err)
	}
	data = append(data, '\n')
	if _, err := file.Write(data); err != nil {
		return errs.E(errs.KindInvalidPath, "store.AppendEvent", ctx.EventsPath, err)
	}
	return nil
}

// ReadEventsForTask reads the last N events for a specific task.
// Events are returned in chronological order (oldest first).
func ReadEventsForTask(ctx workspace.Context, taskID string, limit int) ([]schema.Event, error) {
	op := "store.ReadEventsForTask"

	// Validate task ID
	if _, err := schema.ParseTaskID(taskID); err != nil {
		return nil, errs.E(errs.KindInvalidTaskID, op, taskID, err)
	}

	// Open events file
	file, err := os.Open(ctx.EventsPath)
	if err != nil {
		if errorsIsNotExist(err) {
			return []schema.Event{}, nil
		}
		return nil, errs.E(errs.KindInvalidPath, op, ctx.EventsPath, err)
	}
	defer file.Close()

	// Read all events for this task
	var events []schema.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event schema.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue // Skip malformed lines
		}

		if event.TaskID == taskID {
			events = append(events, event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errs.E(errs.KindInvalidPath, op, ctx.EventsPath, err)
	}

	// Return last N events in chronological order
	if len(events) > limit {
		events = events[len(events)-limit:]
	}

	return events, nil
}
