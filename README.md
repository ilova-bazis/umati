# Umati

A workspace-local CLI task management tool for humans and AI coding agents.

## Features

- **Simple & Lightweight** - File-based storage, no database or server needed
- **AI-Agent Friendly** - Clear CLI interface, JSON data format, documented workflows
- **Task Coordination** - Exclusive claiming prevents duplicate work across agents
- **Hierarchy Support** - Break down work into parent tasks and subtasks
- **Event History** - Track what happened to each task over time
- **Workspace-Local** - Tasks live with your code in `.umati/`

## Installation

```bash
go install github.com/ilova-bazis/umati/cmd/umati@latest
```

Or clone and build:

```bash
git clone https://github.com/ilova-bazis/umati.git
cd umati
go build -o umati ./cmd/umati
mv umati $GOPATH/bin/
```

## Quick Start

Initialize a workspace:

```bash
cd your-project
umati init
```

Create your first task:

```bash
umati create --title "Implement feature X" \
             --description "Add the new feature" \
             --status ready \
             --agent human
```

View tasks:

```bash
umati list all
```

Claim and start work:

```bash
umati claim UM-1 --agent claude
umati start UM-1 --agent claude
```

Complete work:

```bash
umati complete UM-1 --agent claude
```

## Commands

### Workspace

- `umati init [directory]` - Initialize a new workspace

### Reading

- `umati list all` - List all active tasks
- `umati list ready` - List ready tasks with descendants
- `umati list mine --agent <agent>` - List your assigned tasks
- `umati show <task-id>` - Show task details
- `umati search <query>` - Search tasks by title/description

### Writing

- `umati create [options]` - Create a new task
- `umati update <task-id> [options]` - Update task fields
- `umati claim <task-id> --agent <agent>` - Claim a ready task
- `umati start <task-id> --agent <agent>` - Start claimed work
- `umati pause <task-id> --agent <agent>` - Pause active work
- `umati release <task-id> --agent <agent>` - Release claimed work
- `umati complete <task-id> --agent <agent>` - Complete in-progress work
- `umati delete <task-id> --agent <agent>` - Archive a task

### List Filtering

All list commands support filtering:

```bash
umati list all --status ready
umati list all --priority high
umati list all --agent human
umati list ready --priority urgent
```

## Task Lifecycle

```
draft → ready → claimed → in_progress → done
  ↓       ↓         ↓           ↓
pause  cancel    pause      cancel
```

**Statuses:**
- `draft` - Being written, not yet actionable
- `paused` - Intentionally not being worked on
- `ready` - Available for claiming
- `claimed` - Reserved by an agent
- `in_progress` - Actively being implemented
- `done` - Completed
- `cancelled` - Abandoned
- `deleted` - Archived

## Agents

Valid agent identifiers:
- `human` - Human user
- `claude` - Claude AI agent
- `codex` - Codex AI agent
- `opencode` - OpenCode AI agent

## Example Workflows

### Creating Work

```bash
# Create parent task
umati create --title "Build API" --description "REST API implementation" --status ready --agent human

# Create subtasks
umati create --title "Setup routes" --description "Define routes" --status ready --parent UM-1 --agent human
umati create --title "Add auth" --description "Authentication" --status draft --parent UM-1 --agent human
```

### Taking Work

```bash
# Find available work
umati list ready

# Inspect a task
umati show UM-2

# Claim and start
umati claim UM-2 --agent claude
umati start UM-2 --agent claude

# Do implementation work...

# Complete
umati complete UM-2 --agent claude
```

### Managing Your Tasks

```bash
# See your active tasks
umati list mine --agent claude

# Update priority
umati update UM-3 --priority urgent --agent claude

# Search for related tasks
umati search auth
```

## Workspace Structure

```
.umati/
  config.json      # Workspace configuration
  tasks/           # Active task files (JSON)
  deleted/         # Archived deleted tasks
  events/
    events.jsonl   # Event log (append-only)
  .lock            # Lock file (during mutations)
```

## Configuration

Create `config.json` via `umati init`:

```json
{
  "schema_version": 1,
  "id_prefix": "UM",
  "created_at": "2026-04-07T12:00:00Z"
}
```

Customize the ID prefix:

```bash
umati init --id-prefix PROJ
# Creates tasks like PROJ-1, PROJ-2, etc.
```

## For AI Agents

See [`skill.md`](skill.md) for comprehensive documentation on using umati as an AI agent, including:

- Recommended workflows
- Best practices
- Command reference
- Troubleshooting

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o umati ./cmd/umati
```

## Contributing

Contributions welcome! Please ensure:

- Tests pass: `go test ./...`
- Code is formatted: `gofmt -w .`
- Documentation is updated

## License

MIT (placeholder - update with actual license)