package store

import (
	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/workspace"
)

// MutateFunc is a function that performs a mutation while holding the workspace lock.
type MutateFunc func(ctx workspace.Context, cfg schema.Config) error

// WithLock acquires the workspace lock, executes the mutation function, and releases the lock.
// The lock is always released, even if the mutation returns an error or panics.
func WithLock(ctx workspace.Context, actor schema.Actor, command string, fn MutateFunc) error {
	// Acquire lock
	if err := AcquireLock(ctx, actor, command); err != nil {
		return err
	}

	// Load config for the mutation
	cfg, err := workspace.LoadConfig(ctx)
	if err != nil {
		ReleaseLock(ctx) // Best effort release
		return err
	}

	// Execute mutation and always release lock
	mutateErr := fn(ctx, cfg)

	// Release lock (ignore release errors, mutation result is more important)
	ReleaseLock(ctx)

	return mutateErr
}

// WithLockAndEvents acquires lock, executes mutation, and appends events.
// This is a convenience wrapper for mutations that need to log events.
func WithLockAndEvents(ctx workspace.Context, actor schema.Actor, command string, fn func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error)) error {
	op := "store.WithLockAndEvents"

	return WithLock(ctx, actor, command, func(ctx workspace.Context, cfg schema.Config) error {
		events, err := fn(ctx, cfg)
		if err != nil {
			return err
		}

		// Append all events
		for _, event := range events {
			if err := AppendEvent(ctx, event); err != nil {
				return errs.E(errs.KindInvalidEvent, op, ctx.EventsPath, err)
			}
		}

		return nil
	})
}
