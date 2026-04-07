package errs

import (
	"errors"
	"fmt"
)

type Kind string

const (
	KindWorkspaceNotFound  Kind = "workspace_not_found"
	KindWorkspaceLocked    Kind = "workspace_locked"
	KindInvalidConfig      Kind = "invalid_config"
	KindInvalidTaskID      Kind = "invalid_task_id"
	KindTaskNotFound       Kind = "task_not_found"
	KindInvalidTaskFile    Kind = "invalid_task_file"
	KindInvalidDeletedTask Kind = "invalid_deleted_task_file"
	KindInvalidEvent       Kind = "invalid_event_record"
	KindInvalidPath        Kind = "invalid_path"
)

type Error struct {
	Kind Kind
	Op   string
	Path string
	Err  error
}

func (e *Error) Error() string {
	parts := make([]string, 0, 3)
	if e.Op != "" {
		parts = append(parts, e.Op)
	}
	if e.Path != "" {
		parts = append(parts, e.Path)
	}
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}
	if len(parts) == 0 {
		return string(e.Kind)
	}
	return fmt.Sprintf("%s: %s", e.Kind, join(parts))
}

func (e *Error) Unwrap() error { return e.Err }

func E(kind Kind, op, path string, err error) error {
	return &Error{Kind: kind, Op: op, Path: path, Err: err}
}

func IsKind(err error, kind Kind) bool {
	var target *Error
	if !errors.As(err, &target) {
		return false
	}
	return target.Kind == kind
}

func join(parts []string) string {
	result := parts[0]
	for _, part := range parts[1:] {
		result += ": " + part
	}
	return result
}
