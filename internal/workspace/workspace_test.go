package workspace_test

import (
	"path/filepath"
	"testing"

	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/workspace"
)

func TestDiscoverFromNestedDirectory(t *testing.T) {
	start := filepath.Join("..", "..", "testdata", "workspaces", "flat", "nested", "dir")
	ctx, err := workspace.Discover(start)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if filepath.Base(ctx.Root) != "flat" {
		t.Fatalf("expected root to end with flat, got %s", ctx.Root)
	}
}

func TestDiscoverMissingWorkspace(t *testing.T) {
	_, err := workspace.Discover(t.TempDir())
	if !errs.IsKind(err, errs.KindWorkspaceNotFound) {
		t.Fatalf("expected workspace not found error, got %v", err)
	}
}

func TestLoadConfig(t *testing.T) {
	ctx, err := workspace.Discover(filepath.Join("..", "..", "testdata", "workspaces", "empty"))
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	cfg, err := workspace.LoadConfig(ctx)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if cfg.IDPrefix != "UM" {
		t.Fatalf("expected id prefix UM, got %s", cfg.IDPrefix)
	}
}
