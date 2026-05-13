package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ilova-bazis/umati/internal/domain"
	"github.com/ilova-bazis/umati/internal/errs"
	"github.com/ilova-bazis/umati/internal/output"
	"github.com/ilova-bazis/umati/internal/schema"
	"github.com/ilova-bazis/umati/internal/store"
	"github.com/ilova-bazis/umati/internal/workspace"
)

// Run is the root CLI entrypoint.
func Run(args []string) error {
	// Handle version flags before command dispatch
	if len(args) > 0 && (args[0] == "-v" || args[0] == "--version") {
		fmt.Println("umati version", Version)
		return nil
	}

	if len(args) == 0 {
		printUsage()
		return nil
	}

	cmd := args[0]
	switch cmd {
	case "list":
		return runList(args[1:])
	case "show":
		return runShow(args[1:])
	case "create":
		return runCreate(args[1:])
	case "claim":
		return runClaim(args[1:])
	case "start":
		return runStart(args[1:])
	case "pause":
		return runPause(args[1:])
	case "release":
		return runRelease(args[1:])
	case "complete":
		return runComplete(args[1:])
	case "cancel":
		return runCancel(args[1:])
	case "delete":
		return runDelete(args[1:])
	case "init":
		return runInit(args[1:])
	case "search":
		return runSearch(args[1:])
	case "update":
		return runUpdate(args[1:])
	case "board":
		return runBoard(args[1:])
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: umati <command> [args]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -v, --version          Show version information")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  init [directory]                   Initialize a new workspace")
	fmt.Fprintln(os.Stderr, "  list all                           List all active tasks")
	fmt.Fprintln(os.Stderr, "  list ready                         List ready tasks with descendants")
	fmt.Fprintln(os.Stderr, "  list mine --agent <agent>          List tasks assigned to agent")
	fmt.Fprintln(os.Stderr, "  show <task-id>                     Show details for a task")
	fmt.Fprintln(os.Stderr, "  search <query>                     Search tasks by title/description")
	fmt.Fprintln(os.Stderr, "  create [options]                   Create a new task")
	fmt.Fprintln(os.Stderr, "  update <task-id> [options]         Update task fields")
	fmt.Fprintln(os.Stderr, "  claim <task-id> --agent <agent>    Claim a ready task")
	fmt.Fprintln(os.Stderr, "  start <task-id> --agent <agent>    Start work on a claimed task")
	fmt.Fprintln(os.Stderr, "  pause <task-id> --agent <agent>    Pause a claimed or in-progress task")
	fmt.Fprintln(os.Stderr, "  release <task-id> --agent <agent>  Release a claimed task")
	fmt.Fprintln(os.Stderr, "  complete <task-id> --agent <agent> Complete an in-progress task")
	fmt.Fprintln(os.Stderr, "  cancel <task-id> --agent <agent>   Cancel a task (any non-terminal status)")
	fmt.Fprintln(os.Stderr, "  delete <task-id> --agent <agent>   Delete a task")
	fmt.Fprintln(os.Stderr, "  board --agent <agent>              Open the interactive kanban board")
	fmt.Fprintln(os.Stderr, "  help                               Show this help message")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "List filtering (works with all, ready, mine):")
	fmt.Fprintln(os.Stderr, "  --status <status>       Filter by status")
	fmt.Fprintln(os.Stderr, "  --priority <priority>   Filter by priority")
	fmt.Fprintln(os.Stderr, "  --agent <agent>         Filter by assignee")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Create options:")
	fmt.Fprintln(os.Stderr, "  --title <title>         Task title (required)")
	fmt.Fprintln(os.Stderr, "  --description <text>    Task description")
	fmt.Fprintln(os.Stderr, "  --priority <level>      low|medium|high|urgent (default: medium)")
	fmt.Fprintln(os.Stderr, "  --status <status>       draft|paused|ready (default: draft)")
	fmt.Fprintln(os.Stderr, "  --parent <task-id>      Parent task ID for subtasks")
	fmt.Fprintln(os.Stderr, "  --agent <agent>         opencode|codex|claude|human (required)")
	fmt.Fprintln(os.Stderr, "  -i, --interactive       Interactive mode (prompts for all fields)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Update options:")
	fmt.Fprintln(os.Stderr, "  --title <title>         New title")
	fmt.Fprintln(os.Stderr, "  --description <text>    New description")
	fmt.Fprintln(os.Stderr, "  --priority <level>      New priority")
	fmt.Fprintln(os.Stderr, "  --status <status>       New status (with transition validation)")
	fmt.Fprintln(os.Stderr, "  --parent <task-id>      New parent (or 'none' for top-level)")
	fmt.Fprintln(os.Stderr, "  --agent <agent>         Required for all updates")
}

func runList(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: umati list <all|ready|mine> [filters]")
	}

	subcmd := args[0]

	// Parse filter flags (applies to all list types)
	var filters struct {
		status   *string
		priority *string
		agent    *string
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--status":
			i++
			if i >= len(args) {
				return fmt.Errorf("--status requires a value")
			}
			filters.status = &args[i]
		case "--priority":
			i++
			if i >= len(args) {
				return fmt.Errorf("--priority requires a value")
			}
			filters.priority = &args[i]
		case "--agent":
			i++
			if i >= len(args) {
				return fmt.Errorf("--agent requires a value")
			}
			filters.agent = &args[i]
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	switch subcmd {
	case "all":
		return runListAll(filters)
	case "ready":
		return runListReady(filters)
	case "mine":
		return runListMine(filters)
	default:
		return fmt.Errorf("unknown list subcommand: %s", subcmd)
	}
}

func runListAll(filters struct {
	status   *string
	priority *string
	agent    *string
}) error {
	ctx, cfg, err := loadWorkspace()
	if err != nil {
		return renderError(err)
	}

	tasks, err := store.ListTasks(ctx)
	if err != nil {
		return renderError(err)
	}

	tasks = applyFilters(tasks, filters)

	output.RenderListAll(tasks, cfg.IDPrefix)
	return nil
}

func runListReady(filters struct {
	status   *string
	priority *string
	agent    *string
}) error {
	ctx, cfg, err := loadWorkspace()
	if err != nil {
		return renderError(err)
	}

	tasks, err := store.ListTasks(ctx)
	if err != nil {
		return renderError(err)
	}

	tasks = applyFilters(tasks, filters)

	output.RenderListReady(tasks, cfg.IDPrefix)
	return nil
}

func runListMine(filters struct {
	status   *string
	priority *string
	agent    *string
}) error {
	if filters.agent == nil {
		return fmt.Errorf("--agent is required for 'list mine'")
	}

	actor := schema.Actor(*filters.agent)
	if !schema.IsValidActor(actor) {
		return fmt.Errorf("invalid agent: %s", *filters.agent)
	}

	ctx, cfg, err := loadWorkspace()
	if err != nil {
		return renderError(err)
	}

	tasks, err := store.ListTasks(ctx)
	if err != nil {
		return renderError(err)
	}

	// Filter to tasks assigned to this agent with claimed or in-progress status
	var mine []schema.Task
	for _, task := range tasks {
		if task.Assignee != nil && *task.Assignee == actor &&
			(task.Status == schema.StatusClaimed || task.Status == schema.StatusInProgress) {
			mine = append(mine, task)
		}
	}

	mine = applyFilters(mine, filters)

	if len(mine) == 0 {
		fmt.Printf("No active tasks assigned to %s\n", actor)
		return nil
	}

	output.RenderListAll(mine, cfg.IDPrefix)
	return nil
}

func applyFilters(tasks []schema.Task, filters struct {
	status   *string
	priority *string
	agent    *string
}) []schema.Task {
	var result []schema.Task

	for _, task := range tasks {
		// Status filter
		if filters.status != nil {
			if string(task.Status) != *filters.status {
				continue
			}
		}

		// Priority filter
		if filters.priority != nil {
			if string(task.Priority) != *filters.priority {
				continue
			}
		}

		// Agent filter (for assigned tasks)
		if filters.agent != nil {
			if task.Assignee == nil || string(*task.Assignee) != *filters.agent {
				continue
			}
		}

		result = append(result, task)
	}

	return result
}

func runShow(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: umati show <task-id>")
	}

	taskID := args[0]

	ctx, _, err := loadWorkspace()
	if err != nil {
		return renderError(err)
	}

	task, err := store.ReadTask(ctx, taskID)
	if err != nil {
		return renderError(err)
	}

	// Load all tasks for hierarchy
	allTasks, err := store.ListTasks(ctx)
	if err != nil {
		return renderError(err)
	}

	// Load recent events
	events, err := store.ReadEventsForTask(ctx, taskID, 10)
	if err != nil {
		return renderError(err)
	}

	output.RenderShow(task, allTasks, events)
	return nil
}

func runCreate(args []string) error {
	// Check for interactive flag
	interactive := false
	var filteredArgs []string

	for _, arg := range args {
		if arg == "-i" || arg == "--interactive" {
			interactive = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// If interactive mode, run interactive prompt
	if interactive {
		return runCreateInteractive()
	}

	// Otherwise, parse arguments normally
	var opts struct {
		title       string
		description string
		priority    string
		status      string
		parent      string
		agent       string
	}

	for i := 0; i < len(filteredArgs); i++ {
		switch filteredArgs[i] {
		case "--title":
			i++
			if i >= len(filteredArgs) {
				return fmt.Errorf("--title requires a value")
			}
			opts.title = filteredArgs[i]
		case "--description":
			i++
			if i >= len(filteredArgs) {
				return fmt.Errorf("--description requires a value")
			}
			opts.description = filteredArgs[i]
		case "--priority":
			i++
			if i >= len(filteredArgs) {
				return fmt.Errorf("--priority requires a value")
			}
			opts.priority = filteredArgs[i]
		case "--status":
			i++
			if i >= len(filteredArgs) {
				return fmt.Errorf("--status requires a value")
			}
			opts.status = filteredArgs[i]
		case "--parent":
			i++
			if i >= len(filteredArgs) {
				return fmt.Errorf("--parent requires a value")
			}
			opts.parent = filteredArgs[i]
		case "--agent":
			i++
			if i >= len(filteredArgs) {
				return fmt.Errorf("--agent requires a value")
			}
			opts.agent = filteredArgs[i]
		default:
			return fmt.Errorf("unknown flag: %s", filteredArgs[i])
		}
	}

	return runCreateWithOpts(opts)
}

func runCreateInteractive() error {
	// Run interactive prompt
	result, err := interactivePrompt()
	if err != nil {
		return err
	}

	// Convert to opts structure
	opts := struct {
		title       string
		description string
		priority    string
		status      string
		parent      string
		agent       string
	}{
		title:       result.title,
		description: result.description,
		priority:    result.priority,
		status:      result.status,
		parent:      result.parent,
		agent:       result.agent,
	}

	return runCreateWithOpts(opts)
}

func runCreateWithOpts(opts struct {
	title       string
	description string
	priority    string
	status      string
	parent      string
	agent       string
}) error {
	// Validate required arguments
	if opts.title == "" {
		return fmt.Errorf("--title is required")
	}
	if opts.agent == "" {
		return fmt.Errorf("--agent is required")
	}

	// Validate agent
	actor := schema.Actor(opts.agent)
	if !schema.IsValidActor(actor) {
		return fmt.Errorf("invalid agent: %s", opts.agent)
	}

	// Set defaults
	if opts.priority == "" {
		opts.priority = "medium"
	}
	if opts.status == "" {
		opts.status = "draft"
	}

	// Parse priority and status
	priority := schema.Priority(opts.priority)
	if !schema.IsValidPriority(priority) {
		return fmt.Errorf("invalid priority: %s", opts.priority)
	}

	status := schema.Status(opts.status)
	if !schema.IsValidActiveStatus(status) {
		return fmt.Errorf("invalid status: %s", opts.status)
	}

	// Load workspace
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, err := workspace.Discover(wd)
	if err != nil {
		return renderError(err)
	}

	// Perform create with lock
	var taskID string
	err = store.WithLockAndEvents(ctx, actor, "create", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
		// Validate parent if specified
		var parentID *string
		if opts.parent != "" {
			if _, err := store.ReadTask(ctx, opts.parent); err != nil {
				return nil, fmt.Errorf("parent task not found: %s", opts.parent)
			}
			parentID = &opts.parent
		}

		// Validate description requirement
		if opts.description == "" && status != schema.StatusDraft {
			return nil, fmt.Errorf("description is required unless status is draft")
		}

		// Allocate task ID
		newID, err := store.NextTaskID(ctx, cfg)
		if err != nil {
			return nil, err
		}
		taskID = newID

		now := schema.NowTimestamp()

		// Create task
		task := schema.Task{
			ID:          newID,
			Title:       opts.title,
			Description: opts.description,
			Priority:    priority,
			Status:      status,
			Assignee:    nil,
			ParentID:    parentID,
			CreatedAt:   now,
			UpdatedAt:   now,
			CreatedBy:   actor,
			UpdatedBy:   actor,
		}

		if err := store.WriteTask(ctx, task); err != nil {
			return nil, err
		}

		// Create event
		event := schema.Event{
			TaskID:    newID,
			Type:      schema.EventCreated,
			Actor:     actor,
			Timestamp: now,
			Meta:      map[string]any{},
		}

		return []schema.Event{event}, nil
	})

	if err != nil {
		return renderError(err)
	}

	fmt.Printf("Created task %s\n", taskID)
	return nil
}

func runClaim(args []string) error {
	return runStatusChange(args, "claim", schema.StatusClaimed, func(task schema.Task) error {
		if task.Status != schema.StatusReady {
			return fmt.Errorf("task is not ready (status: %s)", task.Status)
		}
		if task.Assignee != nil {
			return fmt.Errorf("task is already claimed by %s", *task.Assignee)
		}
		return nil
	})
}

func runStart(args []string) error {
	return runStatusChange(args, "start", schema.StatusInProgress, func(task schema.Task) error {
		if task.Status != schema.StatusClaimed {
			return fmt.Errorf("task is not claimed (status: %s)", task.Status)
		}
		return nil
	})
}

func runPause(args []string) error {
	return runStatusChange(args, "pause", schema.StatusPaused, func(task schema.Task) error {
		if task.Status != schema.StatusClaimed && task.Status != schema.StatusInProgress {
			return fmt.Errorf("task cannot be paused (status: %s)", task.Status)
		}
		return nil
	})
}

func runRelease(args []string) error {
	return runStatusChange(args, "release", schema.StatusReady, func(task schema.Task) error {
		if task.Status != schema.StatusClaimed && task.Status != schema.StatusInProgress {
			return fmt.Errorf("task cannot be released (status: %s)", task.Status)
		}
		return nil
	})
}

func runCancel(args []string) error {
	return runStatusChange(args, "cancel", schema.StatusCancelled, func(task schema.Task) error {
		if domain.IsTerminalStatus(task.Status) {
			return fmt.Errorf("task is already %s", task.Status)
		}
		return nil
	})
}

func runComplete(args []string) error {
	return runStatusChangeWithValidation(args, "complete", schema.StatusDone, func(task schema.Task, allTasks []schema.Task) error {
		if task.Status != schema.StatusInProgress {
			return fmt.Errorf("task is not in progress (status: %s)", task.Status)
		}
		// Check parent completion
		if err := domain.ValidateParentCompletion(allTasks, task.ID); err != nil {
			return err
		}
		return nil
	})
}

func runDelete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: umati delete <task-id> --agent <agent>")
	}

	taskID := args[0]

	// Parse --agent flag
	var agent string
	for i := 1; i < len(args); i++ {
		if args[i] == "--agent" {
			i++
			if i >= len(args) {
				return fmt.Errorf("--agent requires a value")
			}
			agent = args[i]
		}
	}

	if agent == "" {
		return fmt.Errorf("--agent is required")
	}

	actor := schema.Actor(agent)
	if !schema.IsValidActor(actor) {
		return fmt.Errorf("invalid agent: %s", agent)
	}

	// Load workspace
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, err := workspace.Discover(wd)
	if err != nil {
		return renderError(err)
	}

	// Perform delete with lock
	err = store.WithLockAndEvents(ctx, actor, "delete", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
		// Load all tasks to find subtree
		allTasks, err := store.ListTasks(ctx)
		if err != nil {
			return nil, err
		}

		// Validate delete eligibility
		if err := domain.ValidateDeleteEligibility(allTasks, taskID); err != nil {
			return nil, err
		}

		// Get subtree (root + descendants)
		subtree, err := domain.GetSubtree(allTasks, taskID)
		if err != nil {
			return nil, err
		}

		now := schema.NowTimestamp()
		var events []schema.Event

		// Archive each task in subtree
		for _, task := range subtree {
			// Mark as deleted
			deletedTask := task
			deletedTask.Status = schema.StatusDeleted
			deletedTask.UpdatedAt = now
			deletedTask.UpdatedBy = actor
			deletedAt := now
			deletedTask.DeletedAt = &deletedAt
			deletedTask.DeletedBy = &actor

			// Write to deleted/, remove from tasks/
			if err := store.WriteDeletedTask(ctx, deletedTask); err != nil {
				return nil, err
			}

			// Remove active task file
			activePath, err := store.ActiveTaskPath(ctx, task.ID)
			if err != nil {
				return nil, err
			}
			if err := os.Remove(activePath); err != nil && !os.IsNotExist(err) {
				return nil, err
			}

			// Create event
			event := schema.Event{
				TaskID:    task.ID,
				Type:      schema.EventDeleted,
				Actor:     actor,
				Timestamp: now,
				Meta:      map[string]any{},
			}
			events = append(events, event)
		}

		return events, nil
	})

	if err != nil {
		return renderError(err)
	}

	fmt.Printf("Deleted %s\n", taskID)
	return nil
}

// runStatusChange handles simple status changes with agent validation.
func runStatusChange(args []string, command string, newStatus schema.Status, preValidate func(schema.Task) error) error {
	return runStatusChangeWithValidation(args, command, newStatus, func(task schema.Task, allTasks []schema.Task) error {
		return preValidate(task)
	})
}

// runStatusChangeWithValidation handles status changes with full validation including agent matching.
func runStatusChangeWithValidation(args []string, command string, newStatus schema.Status, validate func(schema.Task, []schema.Task) error) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: umati %s <task-id> --agent <agent>", command)
	}

	taskID := args[0]

	// Parse --agent flag
	var agent string
	for i := 1; i < len(args); i++ {
		if args[i] == "--agent" {
			i++
			if i >= len(args) {
				return fmt.Errorf("--agent requires a value")
			}
			agent = args[i]
		}
	}

	if agent == "" {
		return fmt.Errorf("--agent is required")
	}

	actor := schema.Actor(agent)
	if !schema.IsValidActor(actor) {
		return fmt.Errorf("invalid agent: %s", agent)
	}

	// Load workspace
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, err := workspace.Discover(wd)
	if err != nil {
		return renderError(err)
	}

	// Map command to event type
	var eventType schema.EventType
	switch command {
	case "claim":
		eventType = schema.EventClaimed
	case "start":
		eventType = schema.EventStarted
	case "pause":
		eventType = schema.EventPaused
	case "release":
		eventType = schema.EventReleased
	case "complete":
		eventType = schema.EventCompleted
	case "cancel":
		eventType = schema.EventCancelled
	}

	// Perform mutation with lock
	err = store.WithLockAndEvents(ctx, actor, command, func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
		// Load task
		task, err := store.ReadTask(ctx, taskID)
		if err != nil {
			return nil, err
		}

		// Load all tasks for validation
		allTasks, err := store.ListTasks(ctx)
		if err != nil {
			return nil, err
		}

		// Validate status transition
		if err := domain.CanTransition(task.Status, newStatus); err != nil {
			return nil, err
		}

		// Validate agent match if required
		if err := domain.ValidateAgentMatch(task, actor, newStatus); err != nil {
			return nil, err
		}

		// Run additional validation
		if err := validate(task, allTasks); err != nil {
			return nil, err
		}

		now := schema.NowTimestamp()

		// Update task
		task.Status = newStatus
		task.UpdatedAt = now
		task.UpdatedBy = actor

		// Clear assignee for pause and release
		if newStatus == schema.StatusPaused || newStatus == schema.StatusReady {
			task.Assignee = nil
		}

		// Set assignee for claim
		if command == "claim" {
			task.Assignee = &actor
		}

		// Clear assignee for complete or cancel
		if newStatus == schema.StatusDone || newStatus == schema.StatusCancelled {
			task.Assignee = nil
		}

		if err := store.WriteTask(ctx, task); err != nil {
			return nil, err
		}

		// Create event
		event := schema.Event{
			TaskID:    taskID,
			Type:      eventType,
			Actor:     actor,
			Timestamp: now,
			Meta:      map[string]any{},
		}

		return []schema.Event{event}, nil
	})

	if err != nil {
		return renderError(err)
	}

	fmt.Printf("%s %s\n", strings.ToUpper(command[:1])+command[1:], taskID)
	return nil
}

func loadWorkspace() (workspace.Context, schema.Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return workspace.Context{}, schema.Config{}, err
	}

	ctx, err := workspace.Discover(wd)
	if err != nil {
		return workspace.Context{}, schema.Config{}, err
	}

	cfg, err := workspace.LoadConfig(ctx)
	if err != nil {
		return workspace.Context{}, schema.Config{}, err
	}

	return ctx, cfg, nil
}

func renderError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific error kinds
	switch {
	case errs.IsKind(err, errs.KindWorkspaceNotFound):
		return fmt.Errorf("umati: no workspace found (looked for .umati/)")
	case errs.IsKind(err, errs.KindWorkspaceLocked):
		return fmt.Errorf("umati: workspace is locked by another operation")
	case errs.IsKind(err, errs.KindTaskNotFound):
		return fmt.Errorf("umati: task not found")
	case errs.IsKind(err, errs.KindInvalidConfig):
		return fmt.Errorf("umati: invalid config: %v", unwrap(err))
	case errs.IsKind(err, errs.KindInvalidTaskFile):
		return fmt.Errorf("umati: invalid task file: %v", unwrap(err))
	case errs.IsKind(err, errs.KindInvalidTaskID):
		return fmt.Errorf("umati: invalid task id: %v", unwrap(err))
	default:
		// Clean up the error message
		msg := err.Error()
		msg = strings.TrimPrefix(msg, "workspace_not_found: ")
		msg = strings.TrimPrefix(msg, "task_not_found: ")
		msg = strings.TrimPrefix(msg, "invalid_config: ")
		return fmt.Errorf("umati: %s", msg)
	}
}

func unwrap(err error) error {
	type unwrapper interface {
		Unwrap() error
	}
	if u, ok := err.(unwrapper); ok {
		return u.Unwrap()
	}
	return err
}

func runInit(args []string) error {
	var targetDir string
	var idPrefix string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		if args[i] == "--id-prefix" {
			i++
			if i >= len(args) {
				return fmt.Errorf("--id-prefix requires a value")
			}
			idPrefix = args[i]
		} else if targetDir == "" && !strings.HasPrefix(args[i], "-") {
			targetDir = args[i]
		} else {
			return fmt.Errorf("unknown argument: %s", args[i])
		}
	}

	// Set defaults
	if targetDir == "" {
		targetDir = "."
	}
	if idPrefix == "" {
		idPrefix = "UM"
	}

	// Validate prefix format
	if _, err := schema.ParseTaskID(idPrefix + "-1"); err != nil {
		return fmt.Errorf("invalid id-prefix format: %s", idPrefix)
	}

	// Resolve absolute path
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("cannot resolve path: %w", err)
	}

	// Check if workspace already exists
	ctx := workspace.NewContext(absDir)
	if _, err := os.Stat(ctx.UmatiDir); err == nil {
		return fmt.Errorf("workspace already exists at %s", ctx.UmatiDir)
	}

	// Create directory structure
	if err := os.MkdirAll(ctx.TasksDir, 0755); err != nil {
		return fmt.Errorf("cannot create tasks directory: %w", err)
	}
	if err := os.MkdirAll(ctx.DeletedDir, 0755); err != nil {
		return fmt.Errorf("cannot create deleted directory: %w", err)
	}
	if err := os.MkdirAll(ctx.EventsDir, 0755); err != nil {
		return fmt.Errorf("cannot create events directory: %w", err)
	}

	// Create config.json
	cfg := schema.Config{
		SchemaVersion: 1,
		IDPrefix:      idPrefix,
		CreatedAt:     schema.NowTimestamp(),
	}

	configData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}
	configData = append(configData, '\n')

	if err := os.WriteFile(ctx.ConfigPath, configData, 0644); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	fmt.Printf("Initialized umati workspace at %s\n", ctx.UmatiDir)
	fmt.Printf("Task ID prefix: %s\n", idPrefix)

	// Create .agents/skills/umati/ directory and write SKILL.md
	agentsSkillDir := filepath.Join(absDir, ".agents", "skills", "umati")
	if err := os.MkdirAll(agentsSkillDir, 0755); err != nil {
		return fmt.Errorf("cannot create .agents/skills/umati directory: %w", err)
	}

	skillPath := filepath.Join(agentsSkillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(skillMarkdownContent), 0644); err != nil {
		return fmt.Errorf("cannot write SKILL.md: %w", err)
	}

	fmt.Printf("Created agent skill documentation at %s\n", skillPath)
	return nil
}

func runSearch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: umati search <query>")
	}

	query := strings.ToLower(strings.Join(args, " "))

	ctx, _, err := loadWorkspace()
	if err != nil {
		return renderError(err)
	}

	tasks, err := store.ListTasks(ctx)
	if err != nil {
		return renderError(err)
	}

	var matches []schema.Task
	for _, task := range tasks {
		if strings.Contains(strings.ToLower(task.Title), query) ||
			strings.Contains(strings.ToLower(task.Description), query) {
			matches = append(matches, task)
		}
	}

	if len(matches) == 0 {
		fmt.Println("No matching tasks found.")
		return nil
	}

	output.RenderListAll(matches, "")
	return nil
}

func runUpdate(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: umati update <task-id> [options] --agent <agent>")
	}

	taskID := args[0]

	// Parse options
	var opts struct {
		title       *string
		description *string
		priority    *string
		status      *string
		parent      *string
		agent       string
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--title":
			i++
			if i >= len(args) {
				return fmt.Errorf("--title requires a value")
			}
			opts.title = &args[i]
		case "--description":
			i++
			if i >= len(args) {
				return fmt.Errorf("--description requires a value")
			}
			opts.description = &args[i]
		case "--priority":
			i++
			if i >= len(args) {
				return fmt.Errorf("--priority requires a value")
			}
			opts.priority = &args[i]
		case "--status":
			i++
			if i >= len(args) {
				return fmt.Errorf("--status requires a value")
			}
			opts.status = &args[i]
		case "--parent":
			i++
			if i >= len(args) {
				return fmt.Errorf("--parent requires a value")
			}
			opts.parent = &args[i]
		case "--agent":
			i++
			if i >= len(args) {
				return fmt.Errorf("--agent requires a value")
			}
			opts.agent = args[i]
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	if opts.agent == "" {
		return fmt.Errorf("--agent is required")
	}

	actor := schema.Actor(opts.agent)
	if !schema.IsValidActor(actor) {
		return fmt.Errorf("invalid agent: %s", opts.agent)
	}

	// Check if any field is being updated
	if opts.title == nil && opts.description == nil && opts.priority == nil &&
		opts.status == nil && opts.parent == nil {
		return fmt.Errorf("no fields to update")
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, err := workspace.Discover(wd)
	if err != nil {
		return renderError(err)
	}

	// Perform update with lock
	err = store.WithLockAndEvents(ctx, actor, "update", func(ctx workspace.Context, cfg schema.Config) ([]schema.Event, error) {
		task, err := store.ReadTask(ctx, taskID)
		if err != nil {
			return nil, err
		}

		now := schema.NowTimestamp()
		updated := false

		// Update title
		if opts.title != nil {
			if *opts.title == "" {
				return nil, fmt.Errorf("title cannot be empty")
			}
			task.Title = *opts.title
			updated = true
		}

		// Update description
		if opts.description != nil {
			task.Description = *opts.description
			updated = true
		}

		// Update priority
		if opts.priority != nil {
			priority := schema.Priority(*opts.priority)
			if !schema.IsValidPriority(priority) {
				return nil, fmt.Errorf("invalid priority: %s", *opts.priority)
			}
			task.Priority = priority
			updated = true
		}

		// Update status
		if opts.status != nil {
			newStatus := schema.Status(*opts.status)
			if !schema.IsValidActiveStatus(newStatus) {
				return nil, fmt.Errorf("invalid status: %s", *opts.status)
			}
			if err := domain.CanTransition(task.Status, newStatus); err != nil {
				return nil, err
			}
			task.Status = newStatus
			updated = true
		}

		// Update parent
		if opts.parent != nil {
			if *opts.parent == "none" {
				task.ParentID = nil
			} else {
				// Validate parent exists
				if _, err := store.ReadTask(ctx, *opts.parent); err != nil {
					return nil, fmt.Errorf("parent task not found: %s", *opts.parent)
				}
				// Prevent cycles
				if *opts.parent == taskID {
					return nil, fmt.Errorf("task cannot be its own parent")
				}
				allTasks, _ := store.ListTasks(ctx)
				descendants := domain.Descendants(allTasks, taskID)
				for _, desc := range descendants {
					if desc.ID == *opts.parent {
						return nil, fmt.Errorf("cannot set descendant as parent (would create cycle)")
					}
				}
				task.ParentID = opts.parent
			}
			updated = true
		}

		if !updated {
			return nil, fmt.Errorf("no fields were updated")
		}

		task.UpdatedAt = now
		task.UpdatedBy = actor

		if err := store.WriteTask(ctx, task); err != nil {
			return nil, err
		}

		event := schema.Event{
			TaskID:    taskID,
			Type:      schema.EventUpdated,
			Actor:     actor,
			Timestamp: now,
			Meta:      map[string]any{},
		}

		return []schema.Event{event}, nil
	})

	if err != nil {
		return renderError(err)
	}

	fmt.Printf("Updated %s\n", taskID)
	return nil
}

// skillMarkdownContent contains the SKILL.md documentation for AI agents
const skillMarkdownContent = `# Umati Skill Documentation

## Overview

**Umati** is a workspace-local CLI task management tool designed for coordination between humans and AI coding agents. It provides a simple, file-based task system that supports task creation, assignment, lifecycle management, and archival.

### Why Use Umati?

- **Workspace-local**: Tasks live in ` + "`.umati/`" + ` within your project
- **AI-friendly**: Simple JSON format, clear CLI interface
- **Agent coordination**: Exclusive claiming prevents duplicate work
- **Hierarchy support**: Break down work into subtasks
- **Event history**: Track what happened to each task

## Installation

Umati is a Go binary. Install it and ensure it's in your PATH.

## Quick Start

Initialize a workspace:
` + "```bash" + `
umati init
` + "```" + `

Create your first task:
` + "```bash" + `
umati create --title "Implement auth" --description "Add JWT authentication" --status ready --agent human
` + "```" + `

View tasks:
` + "```bash" + `
umati list all
` + "```" + `

## Core Concepts

### Workspace

Umati operates within a workspace directory containing a ` + "`.umati/`" + ` folder:

` + "```" + `
.umati/
  config.json      # Workspace configuration
  tasks/           # Active task files
  deleted/         # Archived deleted tasks
  events/          # Event log (events.jsonl)
  .lock            # Workspace lock (during mutations)
` + "```" + `

### Task Lifecycle

Tasks flow through these statuses:

` + "```" + `
draft → ready → claimed → in_progress → done
  ↓       ↓         ↓           ↓
pause  cancel    pause      cancel
` + "```" + `

**Status meanings:**
- ` + "`draft`" + ` - Task is being written, not yet actionable
- ` + "`paused`" + ` - Intentionally not being worked on
- ` + "`ready`" + ` - Available for claiming
- ` + "`claimed`" + ` - Reserved by an agent (no other agent can claim)
- ` + "`in_progress`" + ` - Actively being implemented
- ` + "`done`" + ` - Completed
- ` + "`cancelled`" + ` - Abandoned
- ` + "`deleted`" + ` - Archived (not visible in normal operations)

### Agents

Valid agent identifiers:
- ` + "`opencode`" + ` - OpenCode agent
- ` + "`codex`" + ` - Codex agent
- ` + "`claude`" + ` - Claude agent
- ` + "`human`" + ` - Human user

**Important**: Always identify yourself with ` + "`--agent <name>`" + ` when running mutating commands.

## Command Reference

### Workspace Commands

#### ` + "`umati init [directory]`" + `

Initialize a new umati workspace.

Options:
- ` + "`--id-prefix <prefix>`" + ` - Task ID prefix (default: UM)

Examples:
` + "```bash" + `
umati init                           # Initialize in current directory
umati init ./my-project              # Initialize in specific directory
umati init --id-prefix PROJ          # Use PROJ-1, PROJ-2, etc.
` + "```" + `

### Read Commands

#### ` + "`umati list all [filters]`" + `

List all active tasks hierarchically.

Filters:
- ` + "`--status <status>`" + ` - Filter by status
- ` + "`--priority <priority>`" + ` - Filter by priority
- ` + "`--agent <agent>`" + ` - Filter by assignee

Example:
` + "```bash" + `
umati list all
umati list all --status ready
umati list all --priority high --agent human
` + "```" + `

#### ` + "`umati list ready [filters]`" + `

List ready tasks with their descendants (regardless of descendant status).

Same filters as ` + "`list all`" + `.

#### ` + "`umati list mine --agent <agent>`" + `

List tasks assigned to you (claimed or in-progress).

Example:
` + "```bash" + `
umati list mine --agent claude
` + "```" + `

#### ` + "`umati show <task-id>`" + `

Show full details for a task.

Example:
` + "```bash" + `
umati show UM-12
` + "```" + `

#### ` + "`umati search <query>`" + `

Search tasks by title or description (case-insensitive).

Example:
` + "```bash" + `
umati search auth
umati search "api key"
` + "```" + `

### Write Commands (Mutations)

**All write commands require ` + "`--agent <name>`" + ` and acquire a workspace lock.**

#### ` + "`umati create [options]`" + `

Create a new task.

Required:
- ` + "`--title <title>`" + ` - Task title
- ` + "`--agent <agent>`" + ` - Your agent identifier

Optional:
- ` + "`--description <text>`" + ` - Task description (required unless status=draft)
- ` + "`--priority <level>`" + ` - low|medium|high|urgent (default: medium)
- ` + "`--status <status>`" + ` - draft|paused|ready (default: draft)
- ` + "`--parent <task-id>`" + ` - Parent task for subtasks
- ` + "`-i, --interactive`" + ` - Interactive mode (prompts for fields)

Example:
` + "```bash" + `
umati create --title "Fix bug" --description "Fix the login bug" --priority high --status ready --agent human
umati create --title "Subtask" --description "Part of the work" --parent UM-12 --agent human
umati create -i  # Interactive mode
` + "```" + `

#### ` + "`umati update <task-id> [options]`" + `

Update task fields.

Required:
- ` + "`--agent <agent>`" + ` - Your agent identifier

Optional (at least one required):
- ` + "`--title <title>`" + ` - New title
- ` + "`--description <text>`" + ` - New description
- ` + "`--priority <level>`" + ` - New priority
- ` + "`--status <status>`" + ` - New status (must be valid transition)
- ` + "`--parent <task-id>`" + ` - New parent (use 'none' for top-level)

Example:
` + "```bash" + `
umati update UM-12 --priority urgent --agent human
umati update UM-12 --status ready --agent human
umati update UM-12 --parent UM-5 --agent human
umati update UM-12 --parent none --agent human
` + "```" + `

#### ` + "`umati claim <task-id> --agent <agent>`" + `

Claim a ready task exclusively.

Requirements:
- Task must be ` + "`ready`" + `
- Task must have no assignee

Example:
` + "```bash" + `
umati claim UM-12 --agent claude
` + "```" + `

#### ` + "`umati start <task-id> --agent <agent>`" + `

Start work on a claimed task.

Requirements:
- Task must be ` + "`claimed`" + `
- You must be the claiming agent

Example:
` + "```bash" + `
umati start UM-12 --agent claude
` + "```" + `

#### ` + "`umati pause <task-id> --agent <agent>`" + `

Pause a claimed or in-progress task.

Requirements:
- Task must be ` + "`claimed`" + ` or ` + "`in_progress`" + `
- For in_progress, you must be the assignee

Effect:
- Status becomes ` + "`paused`" + `
- Assignee is cleared

Example:
` + "```bash" + `
umati pause UM-12 --agent claude
` + "```" + `

#### ` + "`umati release <task-id> --agent <agent>`" + `

Release a claimed task back to ready.

Requirements:
- Task must be ` + "`claimed`" + ` or ` + "`in_progress`" + `
- You must be the assignee

Effect:
- Status becomes ` + "`ready`" + `
- Assignee is cleared

Example:
` + "```bash" + `
umati release UM-12 --agent claude
` + "```" + `

#### ` + "`umati complete <task-id> --agent <agent>`" + `

Mark an in-progress task as done.

Requirements:
- Task must be ` + "`in_progress`" + `
- You must be the assignee
- All descendants must be ` + "`done`" + ` or ` + "`cancelled`" + `

Effect:
- Status becomes ` + "`done`" + `
- Assignee is cleared

Example:
` + "```bash" + `
umati complete UM-12 --agent claude
` + "```" + `

#### ` + "`umati delete <task-id> --agent <agent>`" + `

Archive a task and its descendants.

Requirements:
- Task and all descendants must NOT be ` + "`claimed`" + ` or ` + "`in_progress`" + `

Effect:
- Task and descendants are moved to ` + "`.umati/deleted/`" + `
- Status becomes ` + "`deleted`" + `
- Deleted tasks don't appear in normal commands

Example:
` + "```bash" + `
umati delete UM-12 --agent human
` + "```" + `

## Agent Workflow

### Starting Work

1. **Find available work:**
   ` + "```bash" + `
   umati list ready
   ` + "```" + `

2. **Inspect a task:**
   ` + "```bash" + `
   umati show UM-12
   ` + "```" + `

3. **Claim the task:**
   ` + "```bash" + `
   umati claim UM-12 --agent claude
   ` + "```" + `

4. **Start working:**
   ` + "```bash" + `
   umati start UM-12 --agent claude
   ` + "```" + `

5. **Do the implementation work** (outside umati)

### Finishing Work

1. **Complete the task:**
   ` + "```bash" + `
   umati complete UM-12 --agent claude
   ` + "```" + `

### If Blocked

1. **Pause the task:**
   ` + "```bash" + `
   umati pause UM-12 --agent claude
   ` + "```" + `

2. **Create a subtask or new task for the blocker:**
   ` + "```bash" + `
   umati create --title "Investigate blocker" --description "Need to understand X before continuing" --status ready --parent UM-12 --agent claude
   ` + "```" + `

### If You Can't Complete

1. **Release the task:**
   ` + "```bash" + `
   umati release UM-12 --agent claude
   ` + "```" + `

## Best Practices

### Do

- **Always claim before starting** - Prevents duplicate work
- **Always use ` + "`--agent`" + `** - Required for mutations, helps tracking
- **Create subtasks for complex work** - Break down large tasks
- **Write clear descriptions** - Help others (and future you) understand
- **Update status promptly** - Keep the board accurate
- **Search before creating** - Avoid duplicate tasks
- **Use interactive mode** - ` + "`umati create -i`" + ` for guided task creation

### Don't

- **Don't work without claiming** - Someone else might pick it up
- **Don't complete parent tasks with unfinished children** - System will block this
- **Don't delete tasks with active work** - Claimed/in-progress tasks block deletion
- **Don't panic if locked** - Wait and retry, or check what's happening

## Common Workflows

### Human Creates Work for AI

` + "```bash" + `
# Create parent task
umati create --title "Implement API" --description "Build the REST API" --status ready --agent human

# Create subtasks
umati create --title "Setup routes" --description "Define API routes" --status ready --parent UM-1 --agent human
umati create --title "Add validation" --description "Input validation" --status draft --parent UM-1 --agent human
` + "```" + `

### AI Takes and Completes Work

` + "```bash" + `
# Find work
umati list ready

# Claim and start
umati claim UM-2 --agent claude
umati start UM-2 --agent claude

# [Do implementation work]

# Complete
umati complete UM-2 --agent claude
` + "```" + `

### Updating Task Priority

` + "```bash" + `
umati update UM-12 --priority urgent --agent human
` + "```" + `

### Moving a Subtask

` + "```bash" + `
# Make top-level
umati update UM-5 --parent none --agent human

# Move to different parent
umati update UM-5 --parent UM-10 --agent human
` + "```" + `

## Troubleshooting

### "workspace is locked by another operation"

Another agent is currently mutating the workspace. Wait a moment and retry.

### "cannot transition from X to Y"

You're trying an invalid state transition. Check the lifecycle diagram.

### "task is assigned to Z, not you"

You tried to modify a task claimed by another agent. Only the claiming agent can start/pause/release/complete.

### "cannot complete: unfinished descendants"

Complete all subtasks before completing a parent task.

### "cannot delete: active descendants"

A task in the subtree is claimed or in-progress. Release or complete those first.

## Tips for AI Agents

1. **Always inspect before claiming** - Read the full task with ` + "`show`" + `
2. **Check for subtasks** - Parent tasks may have work items defined
3. **Use search** - ` + "`umati search auth`" + ` to find related tasks
4. **Update your tasks** - Use ` + "`umati list mine --agent <you>`" + ` to see your work
5. **Create clear titles** - Other agents and humans will read them
6. **Respect the hierarchy** - Complete leaf tasks before parents
`
