package store

import (
	"path/filepath"

	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/workspace"
)

func ActiveTaskPath(ctx workspace.Context, id string) (string, error) {
	if _, err := schema.ParseTaskID(id); err != nil {
		return "", errs.E(errs.KindInvalidTaskID, "store.ActiveTaskPath", id, err)
	}
	return filepath.Join(ctx.TasksDir, id+".json"), nil
}

func DeletedTaskPath(ctx workspace.Context, id string) (string, error) {
	if _, err := schema.ParseTaskID(id); err != nil {
		return "", errs.E(errs.KindInvalidTaskID, "store.DeletedTaskPath", id, err)
	}
	return filepath.Join(ctx.DeletedDir, id+".json"), nil
}
