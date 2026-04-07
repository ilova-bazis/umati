# Umati Disk Schema Specification

## Overview

This document defines the v1 on-disk schema for `umati`.

`umati` stores its workspace-local state in JSON files under `.umati/`. The schema is designed to be easy for humans to inspect, easy for coding agents to read and update, and simple enough to evolve during early iterations.

## Design Principles

- use plain JSON files as the primary storage model
- keep the source of truth local to the workspace
- make task records easy to inspect and edit when necessary
- keep history append-only
- avoid unnecessary denormalization in v1

## Directory Layout

The v1 `.umati/` directory layout is:

```text
.umati/
  config.json
  tasks/
    UM-1.json
    UM-2.json
  deleted/
    UM-9.json
  events/
    events.jsonl
```

## File Roles

- `config.json` stores minimal workspace-level metadata
- `tasks/` stores all active task records, one file per task
- `deleted/` stores archived deleted task records, one file per task
- `events/events.jsonl` stores append-only lifecycle history

## Workspace Config

Path:

```text
.umati/config.json
```

Purpose:

- identify the schema version
- store the task ID prefix used in the workspace
- record workspace creation time

Suggested shape:

```json
{
  "schema_version": 1,
  "id_prefix": "UM",
  "created_at": "2026-04-03T12:00:00Z"
}
```

Rules:

- `schema_version` is required
- `id_prefix` is required
- `created_at` is required
- `config.json` should remain minimal in v1

## Task Storage

Active tasks are stored in:

```text
.umati/tasks/
```

Each task is stored in its own JSON file named after the task ID.

Example:

```text
.umati/tasks/UM-12.json
```

Deleted tasks are moved to:

```text
.umati/deleted/
```

Example:

```text
.umati/deleted/UM-12.json
```

## Active Task Schema

Example active task:

```json
{
  "id": "UM-12",
  "title": "Implement auth refactor",
  "description": "Refactor authentication flow to support token refresh.",
  "priority": "high",
  "status": "ready",
  "assignee": null,
  "parent_id": null,
  "created_at": "2026-04-03T10:00:00Z",
  "updated_at": "2026-04-03T10:15:00Z",
  "created_by": "human",
  "updated_by": "human"
}
```

### Required Fields

- `id`
- `title`
- `description`
- `priority`
- `status`
- `assignee`
- `parent_id`
- `created_at`
- `updated_at`
- `created_by`
- `updated_by`

### Field Definitions

- `id`: task identifier such as `UM-12`
- `title`: short human-readable task summary
- `description`: task details; may be empty only when status is `draft`
- `priority`: one of `low`, `medium`, `high`, `urgent`
- `status`: one of `draft`, `paused`, `ready`, `claimed`, `in_progress`, `done`, `cancelled`
- `assignee`: one of `opencode`, `codex`, `claude`, `human`, or `null`
- `parent_id`: parent task ID or `null` for top-level tasks
- `created_at`: ISO 8601 UTC timestamp
- `updated_at`: ISO 8601 UTC timestamp
- `created_by`: one of `opencode`, `codex`, `claude`, `human`
- `updated_by`: one of `opencode`, `codex`, `claude`, `human`

## Deleted Task Schema

Deleted tasks are archived rather than erased completely.

Example deleted task:

```json
{
  "id": "UM-21",
  "title": "Old API cleanup",
  "description": "Remove deprecated API paths.",
  "priority": "low",
  "status": "deleted",
  "assignee": null,
  "parent_id": null,
  "created_at": "2026-04-03T10:00:00Z",
  "updated_at": "2026-04-03T11:00:00Z",
  "created_by": "human",
  "updated_by": "human",
  "deleted_at": "2026-04-03T11:00:00Z",
  "deleted_by": "human"
}
```

Additional deleted-task fields:

- `deleted_at`: ISO 8601 UTC timestamp of archival
- `deleted_by`: actor who triggered deletion

Deleted-task rules:

- deleted task files live only in `.umati/deleted/`
- deleted tasks use `status: "deleted"`
- deleted tasks are excluded from `umati list all`
- deleted tasks are excluded from `umati show <task-id>`
- when a task is recursively deleted, each child gets its own `deleted_at` and `deleted_by`

## Status Rules

Active task statuses are:

- `draft`
- `paused`
- `ready`
- `claimed`
- `in_progress`
- `done`
- `cancelled`

Archived-only status:

- `deleted`

`deleted` is not part of the normal active task workflow. It exists only for archived task files moved into `.umati/deleted/`.

## Assignee Rules

Allowed assignee values are:

- `opencode`
- `codex`
- `claude`
- `human`
- `null`

Rules:

- `assignee` should be `null` when a task is not actively claimed or in progress
- only the known actors are valid assignee values in v1

## Hierarchy Model

Task hierarchy is represented through `parent_id`.

Rules:

- `parent_id = null` means the task is top-level
- any non-null `parent_id` references another active task
- child relationships are derived by scanning tasks for matching `parent_id`
- v1 does not store denormalized `child_ids`

This keeps the source of truth simple and avoids sync errors between parent and child records.

## Delete Semantics

Delete is archival, not permanent erasure.

When a task is deleted:

- its file is moved from `.umati/tasks/` to `.umati/deleted/`
- its status is changed to `deleted`
- `deleted_at` is set
- `deleted_by` is set
- `updated_at` and `updated_by` should also reflect the deletion action

Recursive delete rules:

- if a task has children in `claimed` or `in_progress`, deletion must fail
- otherwise, deleting a parent task recursively deletes its children
- each archived child receives its own `deleted_at` and `deleted_by`

## Description Rules

- `description` may be empty only when `status` is `draft`
- all non-draft tasks must have a non-empty description

## Event Log

Path:

```text
.umati/events/events.jsonl
```

Format:

- append-only JSON Lines file
- one event object per line

Example:

```json
{"task_id":"UM-12","type":"created","actor":"human","timestamp":"2026-04-03T10:00:00Z","meta":{}}
{"task_id":"UM-12","type":"updated","actor":"human","timestamp":"2026-04-03T10:15:00Z","meta":{"field":"description"}}
{"task_id":"UM-20","type":"claimed","actor":"opencode","timestamp":"2026-04-03T10:30:00Z","meta":{}}
```

### Event Schema

Each event record should contain:

- `task_id`
- `type`
- `actor`
- `timestamp`
- `meta`

Example event object:

```json
{
  "task_id": "UM-12",
  "type": "claimed",
  "actor": "opencode",
  "timestamp": "2026-04-03T10:30:00Z",
  "meta": {}
}
```

### Recommended Event Types

- `created`
- `updated`
- `claimed`
- `released`
- `started`
- `paused`
- `completed`
- `cancelled`
- `deleted`

## Lookup And Visibility Rules

- `umati list all` scans active tasks only
- `umati list ready` scans active tasks only
- `umati show <task-id>` looks only in active tasks
- deleted tasks are archival records and are not returned by normal CLI lookups

## ID Rules

- task IDs use the configured prefix from `config.json`
- the expected format is `<PREFIX>-<number>`
- example: `UM-12`

The exact ID allocation mechanism can be defined in implementation, but stored task IDs must follow the configured prefix.

## Timestamps

All timestamps should use ISO 8601 UTC format.

Examples:

- `2026-04-03T10:00:00Z`
- `2026-04-03T11:15:42Z`

## Validation Summary

- each active task must have all required fields
- each deleted task must have all required fields plus `deleted_at` and `deleted_by`
- `description` may be empty only for `draft`
- `assignee` must be a known actor or `null`
- active task status must not be `deleted`
- archived task status must be `deleted`
- `parent_id` may be `null` or reference another active task

## V1 Summary

The v1 `.umati/` schema uses simple JSON files as the source of truth. Each active task is stored as its own file, deleted tasks are archived separately, and task history is captured in a global append-only event log. The schema is intentionally minimal, explicit, and well-suited for both human inspection and agent-driven workflows.
