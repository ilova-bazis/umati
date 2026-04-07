# Umati Project Specification

## Overview

`umati` is a workspace-local, CLI-first task coordination tool for humans and coding agents. It is designed to help a developer define work, expose that work to AI agents, let agents claim tasks safely, and track the minimal amount of task state needed to coordinate implementation.

The project aims to stay lightweight, practical, and easy to use. It should feel closer to a simple task board for coding agents than a full project management platform.

## Product Intent

`umati` is a developer tool that accompanies AI coding workflows. It is meant to be used alongside a `skill.md` file that teaches agents how to interact with the tool correctly.

Initial target agents include:

- `opencode`
- `codex`
- `claude`
- `human`

The long-term goal is to allow humans to create tasks and let agents periodically inspect available work, claim tasks, implement them, and update their state through the CLI.

## Product Goals

- Provide simple task management for coding workflows
- Allow humans and agents to coordinate through a shared workspace-local tool
- Support agent task claiming so work is not duplicated
- Track enough history to understand what happened to a task
- Support breaking tasks into nested subtasks
- Keep the interface simple and avoid unnecessary complexity

## Non-Goals For V1

- No web UI in the first version
- No advanced permissions or authentication system
- No deep reporting or analytics
- No rich collaboration features like threaded discussion
- No heavy Jira-style workflow complexity

## Interface Roadmap

### V1

- CLI only

### Later Stages

- TUI once the product is mostly complete and the workflow is validated
- Potential future expansion into a web UI

## Workspace Model

`umati` is scoped to a workspace or folder, similar in spirit to Git. It should not require storing repository paths on each task because the current workspace is the context.

The source of truth should live under:

```text
.umati/
```

## Core Concepts

### Tasks

A task is the main unit of work. Tasks should remain intentionally simple.

Minimum fields for a task:

- `id`
- `title`
- `description`
- `priority`
- `status`
- `assignee`
- `parent_id`

Recommended additional metadata for v1:

- `created_at`
- `updated_at`
- `created_by`
- `updated_by`

Deleted task records should also store:

- `deleted_at`
- `deleted_by`

### Task IDs

Task IDs should be short and readable.

Example format:

- `UM-12`

### Priority Levels

Supported priorities in v1:

- `low`
- `medium`
- `high`
- `urgent`

### Task Statuses

Supported statuses in v1:

- `draft`
- `paused`
- `ready`
- `claimed`
- `in_progress`
- `done`
- `cancelled`

Status meanings:

- `draft` - task is still being shaped and is not claimable
- `paused` - task exists but should not be implemented yet
- `ready` - task is available to be claimed
- `claimed` - task is reserved by one agent so no other agent can take it
- `in_progress` - active implementation has started
- `done` - task has been completed
- `cancelled` - task has been intentionally abandoned

## Claiming And Assignment Model

Tasks are claimable exclusively by one agent at a time.

Agents identify themselves explicitly through the CLI using an argument such as:

```bash
umati claim UM-12 --agent opencode
```

Claiming behavior:

- sets `assignee` to the provided agent
- changes status to `claimed`
- records an event
- prevents other agents from claiming the same task

Starting behavior:

- requires the task to already be `claimed`
- changes status to `in_progress`
- records an event

Release behavior:

- happens through an explicit CLI command
- clears the assignee
- should normally return the task to `ready`
- records an event

The CLI should not depend on distinguishing whether the caller is a human or an agent.

## Task Lifecycle Rules

Recommended lifecycle rules for v1:

- only `ready` tasks can be claimed
- only `claimed` tasks can be started
- only `in_progress` tasks can be completed
- `draft` and `paused` tasks are not claimable
- `done` and `cancelled` are terminal states for v1
- releasing a task should be explicit

Completion rule for parent tasks:

- a parent task cannot be completed while it has unfinished subtasks

## Hierarchy And Subtasks

Tasks support a hierarchical structure through `parent_id`.

Subtasks are full tasks, not lightweight checklist items. This allows the same schema and workflow to be used at any depth.

Implications:

- tasks can be nested arbitrarily deep
- subtasks use the same status model as top-level tasks
- parent completion depends on the completion state of child tasks

## Events And History

The event system should stay minimal and useful.

`umati` should record basic operational events such as:

- task created
- task updated
- task claimed
- task released
- task started
- task paused
- task completed
- task cancelled
- task deleted
- assignee changed
- parent changed

Each event should capture:

- `task_id`
- `type`
- `actor`
- `timestamp`
- minimal metadata relevant to the event

The goal is traceability, not a full discussion system.

## CLI Scope For V1

The must-have commands for the first version are:

- `umati list all`
- `umati list ready`
- `umati show <task-id>`
- `umati claim <task-id> --agent <agent>`
- `umati release <task-id> --agent <agent>`
- `umati start <task-id> --agent <agent>`
- `umati complete <task-id> --agent <agent>`
- `umati pause <task-id> --agent <agent>`
- `umati delete <task-id> --agent <agent>`
- `umati create ...`

Useful create patterns:

```bash
umati create --title "..." --description "..."
umati create --parent UM-12 --title "..." --description "..."
```

## Listing Behavior

For `umati list all`:

- show all non-deleted tasks in the workspace
- include nested tasks in the output hierarchy

For `umati list ready`:

- show all ready tasks
- if a ready task has nested tasks, also show its nested tasks regardless of their own status

This allows both users and agents to understand the structure of a ready task before acting on it.

## Agent Workflow Contract

The companion `skill.md` should instruct agents to follow a strict workflow:

1. Query available work with `umati list ready`
2. Inspect the selected task with `umati show <task-id>`
3. Claim the task with `umati claim <task-id> --agent <name>`
4. Start work with `umati start <task-id> --agent <name>`
5. Perform implementation outside `umati`
6. Update task state using `complete`, `pause`, or `release`

Expected agent behavior:

- never start work without claiming first
- never claim a task already claimed by another agent
- always inspect a task before acting on it
- release a task if unable to proceed
- create subtasks if the work needs decomposition
- keep task state changes visible through CLI actions

## Storage Direction

The source of truth should live under `.umati/` with some core files inside it.

Likely future structure:

```text
.umati/
  config.json
  tasks/
  deleted/
  events/
```

Two possible storage approaches were identified:

- JSON files for simplicity and agent readability
- SQLite for stronger consistency and richer queries later

Current leaning for early implementation:

- start with simple files under `.umati/`

Storage behavior for v1:

- active tasks live in `.umati/tasks/`
- deleted tasks are archived in `.umati/deleted/`
- deleted tasks should not appear in normal commands like `list all` or `show`

This keeps v1 inspectable, easy to debug, and easy for agents to reason about.

## Guiding Principles

- keep the workflow simple
- optimize for local developer use
- make agent interaction explicit and deterministic
- avoid overbuilding features before they are justified
- prefer clarity over configurability in v1

## V1 Summary

`umati` is a workspace-local CLI task system for human and AI-agent collaboration. It uses a minimal but structured task model, supports exclusive claiming, allows nested subtasks, and records lightweight history. The system is designed to coordinate implementation work safely across multiple coding agents while staying fast and easy to use.
