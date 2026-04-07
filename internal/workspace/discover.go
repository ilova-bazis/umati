package workspace

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ilova-bazis/umati/internal/errs"
)

func Discover(start string) (Context, error) {
	op := "workspace.Discover"
	resolved, err := filepath.Abs(start)
	if err != nil {
		return Context{}, errs.E(errs.KindInvalidPath, op, start, err)
	}

	for {
		candidate := filepath.Join(resolved, ".umati")
		info, statErr := os.Stat(candidate)
		if statErr == nil && info.IsDir() {
			return NewContext(resolved), nil
		}
		if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
			return Context{}, errs.E(errs.KindInvalidPath, op, candidate, statErr)
		}

		parent := filepath.Dir(resolved)
		if parent == resolved {
			return Context{}, errs.E(errs.KindWorkspaceNotFound, op, start, fs.ErrNotExist)
		}
		resolved = parent
	}
}
