package schema

import "fmt"

// Lock represents a workspace lock file.
type Lock struct {
	PID       int    `json:"pid"`
	Actor     Actor  `json:"actor"`
	Command   string `json:"command"`
	CreatedAt string `json:"created_at"`
}

// ValidateLock validates the lock file content.
func ValidateLock(lock Lock) error {
	if lock.PID <= 0 {
		return Errorf("invalid pid: %d", lock.PID)
	}
	if !IsValidActor(lock.Actor) {
		return Errorf("invalid actor: %q", lock.Actor)
	}
	if lock.Command == "" {
		return Errorf("command is required")
	}
	if err := ValidateTimestamp(lock.CreatedAt); err != nil {
		return Errorf("created_at: %w", err)
	}
	return nil
}

// Errorf is a helper for formatted errors.
func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
