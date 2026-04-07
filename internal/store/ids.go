package store

import (
	"path/filepath"
	"strings"

	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/workspace"
)

func NextTaskID(ctx workspace.Context, cfg schema.Config) (string, error) {
	paths, err := readTaskFilesFromDir(ctx.TasksDir)
	if err != nil {
		return "", errs.E(errs.KindInvalidPath, "store.NextTaskID", ctx.TasksDir, err)
	}
	deletedPaths, err := readTaskFilesFromDir(ctx.DeletedDir)
	if err != nil {
		return "", errs.E(errs.KindInvalidPath, "store.NextTaskID", ctx.DeletedDir, err)
	}
	paths = append(paths, deletedPaths...)

	maxNumber := 0
	for _, path := range paths {
		id := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		parsed, err := schema.ParseTaskID(id)
		if err != nil {
			continue
		}
		if parsed.Prefix != cfg.IDPrefix {
			continue
		}
		if parsed.Number > maxNumber {
			maxNumber = parsed.Number
		}
	}

	return cfg.IDPrefix + "-" + itoa(maxNumber+1), nil
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for v > 0 {
		buf = append([]byte{byte('0' + v%10)}, buf...)
		v /= 10
	}
	return string(buf)
}
