---
name: kaneo-pm
description: Manage Kaneo projects and tasks via its REST API. Use when the user wants to create, list, update, delete, archive, or manage tasks and projects in their self-hosted Kaneo instance. Includes bulk operations, task import/export, and partial task updates. Do NOT use for Trello, Jira, Linear, Notion, or any project management system other than Kaneo.
compatibility: opencode
---

## Purpose

Provides procedures for interacting with Kaneo (https://github.com/GlenEder/kaneo) — a
self-hosted project tracking system. Covers projects CRUD, tasks CRUD, partial task
updates, bulk operations, and task import/export.

## Quick Start

Requires two environment variables:

- `KANEO_API_URL` — base URL for the API (e.g. `http://localhost:3000/api`)
- `KANEO_API_KEY` — API key for Bearer token authentication

### CLI Script

The script at `scripts/kaneo-api.py` wraps all Kaneo API calls. It reads the env vars
above automatically. Run `scripts/kaneo-api.py -h` for full usage.

Example:

```bash
scripts/kaneo-api.py projects list --workspace-id wksp_abc123
scripts/kaneo-api.py tasks create --project-id proj_456 --title "My task"
```

### Schemas & Workspace Guide

Read `references/schemas.md` when you need exact field names, valid priority values,
query parameter options, or error response format.
Read `references/workspace-guide.md` to understand how workspace IDs are passed or
inferred for different endpoints.

## Project Management

### List Projects

1. Read `references/schemas.md` for the project schema and query params.
2. Run `scripts/kaneo-api.py projects list --workspace-id <id> [--include-archived]`
3. If no workspace ID is known, ask the user.

### Create Project

1. Determine `workspaceId`, pick a unique `slug` (URL-safe, lowercase with hyphens),
   choose an `icon` (string, e.g. `"Layout"`), and a `name`.
2. Read `references/schemas.md` for the project schema.
3. Run:
   ```bash
   scripts/kaneo-api.py projects create --workspace-id <id> --name "Project Name" --icon Layout --slug my-project
   ```
4. On success, the project JSON is returned. On 400, adjust fields per the error message.

### Get Project

```bash
scripts/kaneo-api.py projects get --id <project-id>
```

### Update Project

Pass any combination of updatable fields. Read `references/schemas.md` for valid fields.

```bash
scripts/kaneo-api.py projects update --id <project-id> --name "New Name" --description "New desc"
```

### Delete Project

```bash
scripts/kaneo-api.py projects delete --id <project-id>
```

### Archive / Unarchive

```bash
scripts/kaneo-api.py projects archive --id <project-id>
scripts/kaneo-api.py projects unarchive --id <project-id>
```

## Task Management

### List Tasks in a Project

Accepts optional filters. Read `references/schemas.md` for all filter params.

```bash
scripts/kaneo-api.py tasks list --project-id <id> [--status backlog] [--priority high] [--limit 50] [--sort-by createdAt] [--sort-order desc]
```

### Create Task

Required: `--project-id`, `--title`. Optional: `--status`, `--priority`,
`--description`, `--start-date`, `--due-date`, `--user-id`.

```bash
scripts/kaneo-api.py tasks create --project-id <id> --title "Design new feature" --priority medium --status "to-do"
```

**Priority values**: `no-priority`, `low`, `medium`, `high`, `urgent`.
**Status values**: are dynamic per project (column slugs). Defaults typically include
`backlog`, `to-do`, `in-progress`, `in-review`, `done`. Virtual statuses
`planned` and `archived` are also valid.

### Get Task

```bash
scripts/kaneo-api.py tasks get --id <task-id>
```

### Update Task (Full)

Overwrites all provided fields:

```bash
scripts/kaneo-api.py tasks update --id <task-id> --title "Updated" --priority high --status "in-progress"
```

### Partial Updates

Use these for single-field changes:

```bash
scripts/kaneo-api.py tasks update-status --id <task-id> --status "done"
scripts/kaneo-api.py tasks update-priority --id <task-id> --priority urgent
scripts/kaneo-api.py tasks update-title --id <task-id> --title "New title"
scripts/kaneo-api.py tasks update-description --id <task-id> --description "New description"
scripts/kaneo-api.py tasks update-assignee --id <task-id> --user-id <user-id>
scripts/kaneo-api.py tasks update-due-date --id <task-id> --due-date "2026-08-01"
```

### Move Task to Another Project

```bash
scripts/kaneo-api.py tasks move --id <task-id> --destination-project-id <target-id> [--destination-status <status>]
```

### Delete Task

```bash
scripts/kaneo-api.py tasks delete --id <task-id>
```

## Bulk Operations

Run a single operation on multiple tasks in one call.

1. Read `references/schemas.md` for the bulk payload structure.
2. Run:
   ```bash
   scripts/kaneo-api.py tasks bulk --task-ids "id1,id2,id3" --operation updateStatus --value "done"
   ```

**Operations**: `updateStatus`, `updatePriority`, `updateAssignee`, `delete`,
`addLabel`, `removeLabel`, `updateDueDate`

**Value semantics** by operation:
- `updateStatus` → status slug
- `updatePriority` → priority value
- `updateAssignee` → user ID (empty string to unassign)
- `delete` → omit, value is ignored
- `addLabel` / `removeLabel` → label ID
- `updateDueDate` → ISO date string

## Export / Import

### Export Tasks

Exports all tasks from a project as JSON:

```bash
scripts/kaneo-api.py tasks export --project-id <id>
```

### Import Tasks

Imports a JSON array of tasks into a project:

```bash
scripts/kaneo-api.py tasks import --project-id <id> --file tasks.json
```

The JSON file must contain an array of objects. Each object supports:
`title` (required), `description`, `status`, `priority`, `startDate`, `dueDate`,
`userId`.

## Error Handling

### Authentication Failures (401)

- The response says "Unauthorized" or "Invalid API key."
- Verify `KANEO_API_KEY` is set to a valid key for the target instance.
- Verify the key has the correct format (no extra whitespace, no quotes).

### Validation Errors (400)

- The response says which field is invalid.
- Check field names against `references/schemas.md`.
- For statuses: the error includes `Valid statuses for this project: ...` — list the
  columns for the project to find valid status slugs.
- For priorities: use one of `no-priority`, `low`, `medium`, `high`, `urgent`.

### Not Found (404)

- The requested resource does not exist or the user does not have access.
- Verify the ID is correct and belongs to the user's workspace.

### Server Errors (5xx)

- The Kaneo server is unavailable or encountered an internal error.
- Ask the user to check if the server is running.
- Retry once after a brief wait.

### Connection Errors

- The script fails with "Connection error."
- Verify `KANEO_API_URL` is correct and the server is reachable.
- Check network connectivity and that the server port is accessible.
