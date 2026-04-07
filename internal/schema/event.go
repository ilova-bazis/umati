package schema

import "fmt"

type Event struct {
	TaskID    string         `json:"task_id"`
	Type      EventType      `json:"type"`
	Actor     Actor          `json:"actor"`
	Timestamp string         `json:"timestamp"`
	Meta      map[string]any `json:"meta"`
}

func ValidateEvent(event Event) error {
	if _, err := ParseTaskID(event.TaskID); err != nil {
		return err
	}
	if !IsValidEventType(event.Type) {
		return fmt.Errorf("invalid event type: %q", event.Type)
	}
	if !IsValidActor(event.Actor) {
		return fmt.Errorf("invalid actor: %q", event.Actor)
	}
	if err := ValidateTimestamp(event.Timestamp); err != nil {
		return fmt.Errorf("timestamp: %w", err)
	}
	if event.Meta == nil {
		return fmt.Errorf("meta is required")
	}
	return nil
}
