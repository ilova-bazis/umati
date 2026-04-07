package store

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/workspace"
)

// AcquireLock atomically acquires the workspace lock.
// Returns an error if the lock already exists.
func AcquireLock(ctx workspace.Context, actor schema.Actor, command string) error {
	op := "store.AcquireLock"

	// Check if lock already exists
	if _, err := os.Stat(ctx.LockPath); err == nil {
		return errs.E(errs.KindWorkspaceLocked, op, ctx.LockPath, fmt.Errorf("workspace is locked by another operation"))
	}

	// Create lock
	lock := schema.Lock{
		PID:       os.Getpid(),
		Actor:     actor,
		Command:   command,
		CreatedAt: schema.NowTimestamp(),
	}

	if err := schema.ValidateLock(lock); err != nil {
		return errs.E(errs.KindInvalidConfig, op, ctx.LockPath, err)
	}

	// Ensure .umati directory exists
	if err := os.MkdirAll(ctx.UmatiDir, 0o755); err != nil {
		return errs.E(errs.KindInvalidPath, op, ctx.UmatiDir, err)
	}

	// Write lock atomically using temp file + rename
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return errs.E(errs.KindInvalidConfig, op, ctx.LockPath, err)
	}
	data = append(data, '\n')

	tmpFile, err := os.CreateTemp(ctx.UmatiDir, ".lock-*.tmp")
	if err != nil {
		return errs.E(errs.KindInvalidPath, op, ctx.UmatiDir, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return errs.E(errs.KindInvalidPath, op, tmpPath, err)
	}
	if err := tmpFile.Close(); err != nil {
		return errs.E(errs.KindInvalidPath, op, tmpPath, err)
	}

	if err := os.Rename(tmpPath, ctx.LockPath); err != nil {
		return errs.E(errs.KindInvalidPath, op, ctx.LockPath, err)
	}

	return nil
}

// ReleaseLock removes the workspace lock.
// Returns nil if the lock doesn't exist (idempotent).
func ReleaseLock(ctx workspace.Context) error {
	op := "store.ReleaseLock"

	err := os.Remove(ctx.LockPath)
	if err != nil && !errorsIsNotExist(err) {
		return errs.E(errs.KindInvalidPath, op, ctx.LockPath, err)
	}

	return nil
}

// ReadLock reads and parses the workspace lock file.
// Returns an error if the lock doesn't exist or is invalid.
func ReadLock(ctx workspace.Context) (schema.Lock, error) {
	op := "store.ReadLock"

	data, err := os.ReadFile(ctx.LockPath)
	if err != nil {
		if errorsIsNotExist(err) {
			return schema.Lock{}, errs.E(errs.KindTaskNotFound, op, ctx.LockPath, fmt.Errorf("no lock exists"))
		}
		return schema.Lock{}, errs.E(errs.KindInvalidPath, op, ctx.LockPath, err)
	}

	var lock schema.Lock
	if err := json.Unmarshal(data, &lock); err != nil {
		return schema.Lock{}, errs.E(errs.KindInvalidConfig, op, ctx.LockPath, err)
	}

	if err := schema.ValidateLock(lock); err != nil {
		return schema.Lock{}, errs.E(errs.KindInvalidConfig, op, ctx.LockPath, err)
	}

	return lock, nil
}

// IsLocked returns true if a lock file exists.
func IsLocked(ctx workspace.Context) bool {
	_, err := os.Stat(ctx.LockPath)
	return err == nil
}
