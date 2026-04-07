package store

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/workspace"
)

func ReadTask(ctx workspace.Context, id string) (schema.Task, error) {
	op := "store.ReadTask"
	path, err := ActiveTaskPath(ctx, id)
	if err != nil {
		return schema.Task{}, err
	}
	task, err := readTaskFile(path)
	if err != nil {
		if errs.IsKind(err, errs.KindTaskNotFound) {
			return schema.Task{}, errs.E(errs.KindTaskNotFound, op, path, err)
		}
		return schema.Task{}, errs.E(errs.KindInvalidTaskFile, op, path, err)
	}
	if err := schema.ValidateActiveTask(task); err != nil {
		return schema.Task{}, errs.E(errs.KindInvalidTaskFile, op, path, err)
	}
	return task, nil
}

func ListTasks(ctx workspace.Context) ([]schema.Task, error) {
	op := "store.ListTasks"
	entries, err := os.ReadDir(ctx.TasksDir)
	if err != nil {
		if errorsIsNotExist(err) {
			return []schema.Task{}, nil
		}
		return nil, errs.E(errs.KindInvalidPath, op, ctx.TasksDir, err)
	}

	tasks := make([]schema.Task, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(ctx.TasksDir, entry.Name())
		task, err := readTaskFile(path)
		if err != nil {
			return nil, errs.E(errs.KindInvalidTaskFile, op, path, err)
		}
		if err := schema.ValidateActiveTask(task); err != nil {
			return nil, errs.E(errs.KindInvalidTaskFile, op, path, err)
		}
		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i, j int) bool {
		cmp, err := schema.CompareTaskIDs(tasks[i].ID, tasks[j].ID)
		if err != nil {
			return tasks[i].ID < tasks[j].ID
		}
		return cmp < 0
	})

	return tasks, nil
}

func WriteTask(ctx workspace.Context, task schema.Task) error {
	if err := schema.ValidateActiveTask(task); err != nil {
		return errs.E(errs.KindInvalidTaskFile, "store.WriteTask", task.ID, err)
	}
	path, err := ActiveTaskPath(ctx, task.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(ctx.TasksDir, 0o755); err != nil {
		return errs.E(errs.KindInvalidPath, "store.WriteTask", ctx.TasksDir, err)
	}
	return writeJSONAtomically(path, task)
}

func WriteDeletedTask(ctx workspace.Context, task schema.Task) error {
	if err := schema.ValidateDeletedTask(task); err != nil {
		return errs.E(errs.KindInvalidDeletedTask, "store.WriteDeletedTask", task.ID, err)
	}
	path, err := DeletedTaskPath(ctx, task.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(ctx.DeletedDir, 0o755); err != nil {
		return errs.E(errs.KindInvalidPath, "store.WriteDeletedTask", ctx.DeletedDir, err)
	}
	return writeJSONAtomically(path, task)
}

func readTaskFile(path string) (schema.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errorsIsNotExist(err) {
			return schema.Task{}, errs.E(errs.KindTaskNotFound, "store.readTaskFile", path, err)
		}
		return schema.Task{}, err
	}
	var task schema.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return schema.Task{}, err
	}
	return task, nil
}

func writeJSONAtomically(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func errorsIsNotExist(err error) bool {
	return err != nil && os.IsNotExist(err)
}

func readTaskFilesFromDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errorsIsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		paths = append(paths, filepath.Join(dir, entry.Name()))
	}
	return paths, nil
}

var _ fs.FileInfo
