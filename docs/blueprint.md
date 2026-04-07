# Umati Implementation Blueprint

## Overview

This document describes how to implement `umati` v1 based on the current product, CLI, and disk schema specifications.

The implementation should optimize for simplicity, correctness, and predictable behavior for both humans and coding agents.

## Implementation Goals

- keep `.umati/` as the single source of truth
- implement the full v1 CLI contract defined in `cli.md`
- enforce lifecycle and schema rules consistently
- make JSON storage safe enough for local multi-agent usage
- keep the internals simple enough to evolve later

## Source Of Truth

The workspace state lives entirely in `.umati/`.

Directory layout:

```text
.umati/
  config.json
  .lock
  tasks/
    UM-1.json
    UM-2.json
  deleted/
    UM-9.json
  events/
    events.jsonl
```

Storage rules:

- active tasks live in `.umati/tasks/`
- deleted tasks are archived in `.umati/deleted/`
- lifecycle history is appended to `.umati/events/events.jsonl`
- `.umati/.lock` protects mutating operations

## Architecture

The implementation should be split into a small number of focused layers.

### CLI Layer

Responsibilities:

- parse command-line arguments
- route commands to handlers
- print human-readable output
- map domain errors to clean CLI messages

### Workspace Layer

Responsibilities:

- locate `.umati/` from the current working directory
- load and validate `config.json`
- provide resolved paths for `tasks/`, `deleted/`, `events/`, and `.lock`

### Schema Layer

Responsibilities:

- define allowed enums and field rules
- validate task records
- validate config records
- validate event records
- provide timestamp and actor helpers

### Store Layer

Responsibilities:

- read active task files
- list active task files
- write active task files atomically
- move tasks to `.umati/deleted/`
- append events to `events.jsonl`
- scan tasks by `parent_id`

### Domain Layer

Responsibilities:

- implement lifecycle transitions
- implement task hierarchy rules
- validate allowed state changes
- enforce delete recursion rules
- enforce completion rules for parent tasks

### Output Layer

Responsibilities:

- render `list all`
- render `list ready`
- render `show`
- render short success and error messages

## Core Runtime Model

The CLI should treat task files as the main source of truth.

Implementation choices:

- no cache layer in v1
- no index file in v1
- no denormalized `child_ids`
- derive children by scanning active tasks for matching `parent_id`

This keeps the implementation easy to reason about and reduces synchronization bugs.

## Locking Design

Because `umati` uses JSON files and supports multiple agents, mutating commands must use a workspace-scoped lock.

Lock file path:

```text
.umati/.lock
```

Mutating commands that must acquire the lock:

- `create`
- `claim`
- `start`
- `pause`
- `release`
- `complete`
- `delete`

Read-only commands that do not require the lock:

- `list all`
- `list ready`
- `show`

### Lock Behavior

- acquire the lock before reading files for a mutation
- keep the lock through validation, write, move, and event append
- release the lock only after the mutation finishes or fails cleanly
- if the lock already exists, fail fast with a clear error

Recommended lock error:

```text
Workspace is locked by another umati operation
```

### Lock File Shape

Suggested lock file content:

```json
{
  "pid": 12345,
  "actor": "opencode",
  "command": "claim UM-12",
  "created_at": "2026-04-03T12:30:00Z"
}
```

### Locking Notes

- lock creation should be atomic
- lock cleanup should happen in success and error paths
- stale lock recovery can be handled later with a dedicated command if needed

## Command Implementation Order

Recommended implementation order:

1. workspace loading and config validation
2. task schema validation and storage primitives
3. lock acquisition and release
4. `show`
5. `list all`
6. `list ready`
7. `create`
8. `claim`
9. `start`
10. `pause`
11. `release`
12. `complete`
13. `delete`

This order gets read commands working early and reduces debugging complexity.

## Command Blueprint

### `show`

Implementation responsibilities:

- load active task from `.umati/tasks/`
- if not found, return not found
- never return archived tasks from `.umati/deleted/`
- load child tasks by scanning active tasks for matching `parent_id`
- load recent task events from `events.jsonl`

### `list all`

Implementation responsibilities:

- scan `.umati/tasks/`
- exclude deleted tasks by definition
- render full active hierarchy
- show compact fields: `id`, `priority`, `status`, `assignee`, `title`

### `list ready`

Implementation responsibilities:

- scan `.umati/tasks/`
- identify all tasks with `status=ready`
- when a ready task has descendants, include them regardless of descendant status
- render hierarchy for context

### `create`

Implementation responsibilities:

- acquire lock
- allocate next task ID
- validate title
- require non-empty description unless `status=draft`
- validate parent if provided
- write active task file
- append `created` event

Defaults:

- `priority=medium`
- `status=draft`
- `assignee=null`

### `claim`

Implementation responsibilities:

- acquire lock
- load task
- require `status=ready`
- set `assignee=<agent>`
- set `status=claimed`
- update timestamps and actor fields
- append `claimed` event

### `start`

Implementation responsibilities:

- acquire lock
- load task
- require `status=claimed`
- require `assignee=<agent>`
- set `status=in_progress`
- update timestamps and actor fields
- append `started` event

### `pause`

Implementation responsibilities:

- acquire lock
- load task
- require `status=claimed` or `status=in_progress`
- clear `assignee`
- set `status=paused`
- update timestamps and actor fields
- append `paused` event

### `release`

Implementation responsibilities:

- acquire lock
- load task
- require `status=claimed` or `status=in_progress`
- clear `assignee`
- set `status=ready`
- update timestamps and actor fields
- append `released` event

### `complete`

Implementation responsibilities:

- acquire lock
- load task
- require `status=in_progress`
- require `assignee=<agent>`
- verify all descendants are `done` or `cancelled`
- set `status=done`
- keep assignee as-is or clear it based on implementation choice, but be consistent
- update timestamps and actor fields
- append `completed` event

Recommendation:

- clear `assignee` on completion for a cleaner terminal state

### `delete`

Implementation responsibilities:

- acquire lock
- load task and all descendants
- fail if task or any descendant is `claimed` or `in_progress`
- recursively archive the task and its descendants
- for each archived task:
  - set `status=deleted`
  - set `deleted_at`
  - set `deleted_by`
  - update `updated_at`
  - update `updated_by`
  - move file to `.umati/deleted/`
  - append `deleted` event

Delete is archival, not permanent erasure.

## Domain Rules

### Status Rules

Allowed active statuses:

- `draft`
- `paused`
- `ready`
- `claimed`
- `in_progress`
- `done`
- `cancelled`

Archived-only status:

- `deleted`

### Priority Rules

Allowed priorities:

- `low`
- `medium`
- `high`
- `urgent`

### Assignee Rules

Allowed assignee values:

- `opencode`
- `codex`
- `claude`
- `human`
- `null`

### Description Rules

- description may be empty only if status is `draft`
- non-draft tasks must have a non-empty description

### Completion Rule Track enough history to understand what s

- only `in_progress` tasks can be completed
- parent tasks cannot be completed while any descendant is not `done` or `cancelled`

### Delete Rules

- `claimed` tasks cannot be deleted
- `in_progress` tasks cannot be deleted
- if any descendant is `claimed` or `in_progress`, deletion must fail
- otherwise, deletion recursively archives the full subtree

## Storage Operations

### Atomic Writes

Task writes should be atomic.

Recommended approach:

- write to a temporary file in the same directory
- fsync if available in the implementation language
- rename the temp file to the final file name

### Event Appends

Event writes should:

- append one JSON object per line
- happen while the mutation lock is held
- occur after the corresponding task mutation is prepared

### Delete Moves

Deleting a task should:

- update the archived task payload first
- write it into `.umati/deleted/`
- remove the active file from `.umati/tasks/`

## Suggested Error Model

Errors should be concise and readable.

Examples:

- `Task not found: UM-99`
- `Cannot claim UM-12: status is paused`
- `Cannot start UM-12: assignee is codex`
- `Cannot complete UM-5: unfinished subtasks exist`
- `Cannot delete UM-3: descendant UM-4 is in progress`
- `Workspace is locked by another umati operation`

## Testing Blueprint

### Unit Tests

Test:

- config validation
- task schema validation
- actor validation
- description rules
- status transition rules
- parent-child traversal
- delete eligibility checks

### Integration Tests

Test full command flows:

- create -> claim -> start -> complete
- create parent -> create child -> complete child -> complete parent
- create parent -> child in progress -> delete parent fails
- create parent -> children done/cancelled -> delete parent archives subtree
- deleted tasks do not appear in `list all`
- deleted tasks do not appear in `show`

### Locking Tests

Test:

- concurrent claim attempts on the same task
- concurrent mutation attempts in the same workspace
- lock cleanup on mutation failure

## Fixtures

Implementation should include sample fixture workspaces for tests.

Recommended fixtures:

- empty workspace
- workspace with flat tasks
- workspace with nested tasks
- workspace with archived deleted tasks
- workspace with mixed ready and non-ready tasks

## Implementation Milestones

### Milestone 1

- workspace detection
- config loading
- schema definitions
- storage helpers

### Milestone 2

- lock implementation
- `show`
- `list all`
- `list ready`

### Milestone 3

- `create`
- `claim`
- `start`
- `pause`
- `release`
- `complete`

### Milestone 4

- recursive archival delete
- event polishing
- integration tests

### Milestone 5

- agent-facing output cleanup
- `skill.md` integration support
- implementation hardening

## Future Extensions

Possible future work after v1:

- `umati init`
- `--json` output mode
- `list deleted`
- `restore` for archived tasks
- stale lock recovery command
- TUI layer
- alternate storage backend such as SQLite

## V1 Summary

The v1 implementation should stay intentionally small: JSON-backed, lock-protected, CLI-first, and strict about lifecycle rules. The main goal is to make task coordination reliable for humans and coding agents without introducing unnecessary complexity.
