# Umati CLI Specification

## Overview

This document defines the v1 command-line interface for `umati`.

`umati` is a workspace-local CLI tool for managing tasks that can be created by humans and worked on by coding agents. The CLI is designed to be simple enough for direct human use while also being predictable enough for agent automation.

## Command Shape

Base form:

```bash
umati <command> [subcommand] [args] [flags]
```

State-changing commands should accept explicit actor identity through `--agent`.

Example:

```bash
umati claim UM-12 --agent opencode
```

## Output Principles

- default output should be human-readable
- output should remain stable and easy for agents to parse
- machine-friendly output such as `--json` can be added later

## Core Commands

The v1 command set is:

- `umati list all`
- `umati list ready`
- `umati show <task-id>`
- `umati create --title <title> --description <text> [options]`
- `umati claim <task-id> --agent <agent>`
- `umati start <task-id> --agent <agent>`
- `umati pause <task-id> --agent <agent>`
- `umati release <task-id> --agent <agent>`
- `umati complete <task-id> --agent <agent>`
- `umati delete <task-id> --agent <agent>`

## Supported Agents

Initial supported actor values:

- `opencode`
- `codex`
- `claude`
- `human`

## Command Reference

### `umati list ready`

Purpose:

- show all tasks in `ready` state

Behavior:

- returns all ready tasks in the workspace
- if a ready task has nested tasks, include those nested tasks in the output regardless of their own status

Example:

```bash
umati list ready
```

Example output:

```text
UM-12  high    ready        Implement auth refactor
  UM-13  medium  draft        Add token parsing helper
  UM-14  ready   ready        Update session middleware
UM-20  urgent  ready        Fix broken CI pipeline
```

### `umati list all`

Purpose:

- show all tasks in the workspace

Behavior:

- returns top-level and nested tasks
- includes task status, priority, assignee, and title in a compact view
- excludes deleted tasks
- useful for humans reviewing the whole workspace state

Example:

```bash
umati list all
```

Example output:

```text
UM-12  high    ready        -         Implement auth refactor
  UM-13  medium  draft        -         Add token parsing helper
  UM-14  ready   claimed      claude    Update session middleware
UM-20  urgent  done         opencode  Fix broken CI pipeline
UM-21  low     paused       -         Revisit API cleanup
```

### `umati show <task-id>`

Purpose:

- show full details for a single task

Behavior:

- shows task metadata
- shows parent and child relationships
- shows assignee if present
- shows recent events
- does not return deleted tasks

Example:

```bash
umati show UM-12
```

Example output:

```text
UM-12
Title: Implement auth refactor
Status: ready
Priority: high
Assignee: -
Parent: -
Description:
Refactor authentication flow to support token refresh and cleaner session boundaries.

Subtasks:
- UM-13 draft Add token parsing helper
- UM-14 ready Update session middleware

Recent events:
- created by human at 2026-04-03T10:00:00Z
- updated by human at 2026-04-03T10:15:00Z
```

### `umati create`

Purpose:

- create a top-level task or subtask

Arguments:

- `--title <title>`
- `--description <text>`

Optional arguments:

- `--priority <low|medium|high|urgent>`
- `--status <draft|paused|ready|claimed|in_progress|done|cancelled>`
- `--parent <task-id>`
- `--agent <actor>`

Recommended defaults:

- `priority=medium`
- `status=draft`
- `assignee=null`

Examples:

```bash
umati create --title "Implement auth refactor" --description "Refactor authentication flow to support token refresh."
umati create --parent UM-12 --title "Update session middleware" --description "Adjust middleware to use refreshed tokens." --priority high --status ready
umati create --title "Fix CI pipeline" --description "Repair failing checks in GitHub Actions." --priority urgent --agent human
```

Example output:

```text
Created task UM-21
```

### `umati claim <task-id> --agent <agent>`

Purpose:

- exclusively reserve a ready task for one agent

Behavior:

- requires task status to be `ready`
- sets `assignee` to the provided agent
- changes status to `claimed`
- records a claim event
- prevents another agent from claiming the same task

Example:

```bash
umati claim UM-20 --agent opencode
```

Example output:

```text
Claimed UM-20 as opencode
```

### `umati start <task-id> --agent <agent>`

Purpose:

- mark claimed work as actively being implemented

Behavior:

- requires task status to be `claimed`
- should validate that the starting agent matches the assignee
- changes status to `in_progress`
- records a start event

Example:

```bash
umati start UM-20 --agent opencode
```

Example output:

```text
Started UM-20 as opencode
```

### `umati pause <task-id> --agent <agent>`

Purpose:

- stop work and mark the task as intentionally not actionable right now

Behavior:

- allowed from `claimed` or `in_progress`
- clears `assignee`
- changes status to `paused`
- records a pause event

Example:

```bash
umati pause UM-20 --agent opencode
```

Example output:

```text
Paused UM-20
```

### `umati release <task-id> --agent <agent>`

Purpose:

- give up a claim without completing the task

Behavior:

- allowed from `claimed` or `in_progress`
- clears `assignee`
- changes status to `ready`
- records a release event

Example:

```bash
umati release UM-20 --agent opencode
```

Example output:

```text
Released UM-20 back to ready
```

### `umati complete <task-id> --agent <agent>`

Purpose:

- mark active work as finished

Behavior:

- requires task status to be `in_progress`
- should validate that the completing agent matches the assignee
- parent tasks cannot be completed while any child task remains unfinished
- changes status to `done`
- records a completion event

Example:

```bash
umati complete UM-20 --agent opencode
```

Example output:

```text
Completed UM-20
```

### `umati delete <task-id> --agent <agent>`

Purpose:

- archive a task and remove it from active task views

Behavior:

- intended for mistaken, obsolete, or abandoned tasks that should no longer exist
- moves the task file from `.umati/tasks/` to `.umati/deleted/`
- changes task status to `deleted`
- stamps `deleted_at` and `deleted_by`
- should be allowed only when the task and all descendants are not actively claimed or in progress
- recursively archives child tasks when allowed
- records a delete event for each archived task

Recommended allowed states:

- `draft`
- `paused`
- `ready`
- `done`
- `cancelled`

Example:

```bash
umati delete UM-21 --agent human
```

Example output:

```text
Deleted UM-21
```

## Validation Rules

Recommended v1 validation rules:

- `claim` fails if the task is not `ready`
- `claim` fails if the task already has an assignee
- `start` fails unless the task is `claimed`
- `pause` fails from `draft`, `done`, or `cancelled`
- `release` fails unless the task is currently `claimed` or `in_progress`
- `complete` fails unless the task is `in_progress`
- `complete` fails if the task has unfinished subtasks
- `delete` fails for `claimed` or `in_progress` tasks
- `delete` fails if any descendant is `claimed` or `in_progress`
- `show` should fail clearly for unknown task IDs
- `show` should treat deleted task IDs as not found

For parent completion checks, tasks in `done` or `cancelled` should count as finished.

## State Transition Intent

The intended task flow is:

```text
draft -> ready
draft -> paused
paused -> ready
ready -> claimed
claimed -> in_progress
claimed -> paused
claimed -> ready
in_progress -> paused
in_progress -> ready
in_progress -> done
ready -> cancelled
claimed -> cancelled
in_progress -> cancelled
paused -> cancelled
draft -> cancelled
draft -> deleted
paused -> deleted
ready -> deleted
done -> deleted
cancelled -> deleted
```

In command terms:

- `claim` moves `ready -> claimed`
- `start` moves `claimed -> in_progress`
- `pause` moves `claimed|in_progress -> paused`
- `release` moves `claimed|in_progress -> ready`
- `complete` moves `in_progress -> done`
- `delete` archives eligible tasks into `.umati/deleted/`

## Suggested Global Flags

Useful global flags for v1 or near-v1:

- `--agent <name>`
- `--help`
- `--version`

Possible later additions:

- `--json`
- `--no-color`

## Example Workflows

### Human Creates Work

```bash
umati create --title "Implement auth refactor" --description "Refactor authentication flow."
umati create --parent UM-12 --title "Update session middleware" --description "Use refreshed tokens." --status ready
```

### Agent Takes A Task

```bash
umati list all
```

### Human Deletes A Task

```bash
umati show UM-21
umati delete UM-21 --agent human
```

### Agent Takes A Task

```bash
umati list ready
umati show UM-20
umati claim UM-20 --agent claude
umati start UM-20 --agent claude
umati complete UM-20 --agent claude
```

### Agent Gets Blocked

```bash
umati list ready
umati show UM-12
umati claim UM-12 --agent codex
umati start UM-12 --agent codex
umati pause UM-12 --agent codex
```

### Agent Releases Work

```bash
umati list ready
umati show UM-30
umati claim UM-30 --agent opencode
umati start UM-30 --agent opencode
umati release UM-30 --agent opencode
```

## Design Notes

- `pause` means the task should not be acted on right now
- `release` means the task is available again for someone else to claim
- `delete` is archival, not permanent erasure
- both humans and agents use the same command contract
- the CLI should stay small and focused in v1
