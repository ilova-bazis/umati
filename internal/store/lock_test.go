package store_test

import (
	"testing"

	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/store"
	"github.com/ilova-bazis/umati/internal/workspace"
)

func TestAcquireAndReleaseLock(t *testing.T) {
	ctx := workspace.NewContext(t.TempDir())

	// Acquire lock
	err := store.AcquireLock(ctx, schema.ActorHuman, "test command")
	if err != nil {
		t.Fatalf("AcquireLock() error = %v", err)
	}

	// Verify lock exists
	if !store.IsLocked(ctx) {
		t.Fatal("expected lock to exist")
	}

	// Read lock
	lock, err := store.ReadLock(ctx)
	if err != nil {
		t.Fatalf("ReadLock() error = %v", err)
	}
	if lock.Actor != schema.ActorHuman {
		t.Fatalf("expected actor human, got %s", lock.Actor)
	}
	if lock.Command != "test command" {
		t.Fatalf("expected command 'test command', got %s", lock.Command)
	}

	// Release lock
	err = store.ReleaseLock(ctx)
	if err != nil {
		t.Fatalf("ReleaseLock() error = %v", err)
	}

	// Verify lock is gone
	if store.IsLocked(ctx) {
		t.Fatal("expected lock to be released")
	}
}

func TestAcquireLockContention(t *testing.T) {
	ctx := workspace.NewContext(t.TempDir())

	// First acquire
	err := store.AcquireLock(ctx, schema.ActorHuman, "first")
	if err != nil {
		t.Fatalf("first AcquireLock() error = %v", err)
	}

	// Second acquire should fail
	err = store.AcquireLock(ctx, schema.ActorCodex, "second")
	if !errs.IsKind(err, errs.KindWorkspaceLocked) {
		t.Fatalf("expected workspace locked error, got %v", err)
	}
}

func TestReleaseLockIdempotent(t *testing.T) {
	ctx := workspace.NewContext(t.TempDir())

	// Release when no lock exists should not error
	err := store.ReleaseLock(ctx)
	if err != nil {
		t.Fatalf("ReleaseLock() on non-existent lock error = %v", err)
	}
}
