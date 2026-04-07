package workspace

import "path/filepath"

type Context struct {
	Root       string
	UmatiDir   string
	ConfigPath string
	LockPath   string
	TasksDir   string
	DeletedDir string
	EventsDir  string
	EventsPath string
}

func NewContext(root string) Context {
	umatiDir := filepath.Join(root, ".umati")
	eventsDir := filepath.Join(umatiDir, "events")
	return Context{
		Root:       root,
		UmatiDir:   umatiDir,
		ConfigPath: filepath.Join(umatiDir, "config.json"),
		LockPath:   filepath.Join(umatiDir, ".lock"),
		TasksDir:   filepath.Join(umatiDir, "tasks"),
		DeletedDir: filepath.Join(umatiDir, "deleted"),
		EventsDir:  eventsDir,
		EventsPath: filepath.Join(eventsDir, "events.jsonl"),
	}
}
