# Kaneo API Schemas

## Project Schema

```
{
  id: string          // cuid2
  workspaceId: string // FK → workspace
  slug: string        // URL-safe unique identifier
  icon: string|null   // defaults to "Layout"
  name: string
  description: string|null
  createdAt: string   // ISO date
  isPublic: boolean|null
  archivedAt: string|null // ISO date, non-null = archived
}
```

## Task Schema

```
{
  id: string              // cuid2
  projectId: string       // FK → project
  position: number|null
  number: number|null     // task #, unique per project
  userId: string|null     // assignee ID (DB column: assignee_id)
  title: string
  description: string|null
  status: string          // dynamic per project (column slug); virtual: "planned", "archived"
  priority: "no-priority"|"low"|"medium"|"high"|"urgent"
  startDate: string|null  // ISO date
  dueDate: string|null    // ISO date
  createdAt: string       // ISO date
  updatedAt: string       // ISO date
}
```

## Valid Priorities

| Value | Label |
|---|---|
| `no-priority` | No priority |
| `low` | Low |
| `medium` | Medium |
| `high` | High |
| `urgent` | Urgent |

## Valid Statuses

**Statuses are dynamic per project** — they come from the project's columns (column slugs).
Virtual statuses (always valid): `planned`, `archived`.

To discover valid statuses for a project, list its columns:
`GET /api/column?projectId={projectId}` — returns `[{ slug, name, ... }]`

Default columns typically include: `backlog`, `to-do`, `in-progress`, `in-review`, `done`.

## Task Query Params (listTasks)

All optional:

| Param | Type | Values |
|---|---|---|
| `status` | string | Filter by status slug |
| `priority` | string | Filter by priority: `no-priority\|low\|medium\|high\|urgent` |
| `assigneeId` | string | Filter by assignee user ID |
| `page` | number | Page number (1-based) |
| `limit` | number | Items per page |
| `sortBy` | enum | `createdAt\|priority\|dueDate\|position\|title\|number` |
| `sortOrder` | `asc\|desc` | Sort direction |
| `dueBefore` | string | ISO date — due date before |
| `dueAfter` | string | ISO date — due date after |

## Bulk Operation Payload

PATCH `/api/task/bulk`

```
{
  taskIds: string[]        // min 1 task IDs
  operation: "updateStatus"|"updatePriority"|"updateAssignee"|"delete"|"addLabel"|"removeLabel"|"updateDueDate"
  value: string|null       // operation-specific:
                           //   updateStatus → status slug
                           //   updatePriority → priority value
                           //   updateAssignee → user ID (empty string = unassign)
                           //   delete → null
                           //   addLabel → label ID
                           //   removeLabel → label ID
                           //   updateDueDate → ISO date
}
```

Response: `{ success: boolean, updatedCount: number }`

## Error Response

```
{
  message: string
  status: number  // HTTP status code
}
```

Common statuses:
- `400` — validation error (invalid field, missing field)
- `401` — missing/invalid API key
- `404` — resource not found
- `500` — server error
