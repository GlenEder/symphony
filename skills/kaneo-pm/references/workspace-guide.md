# Workspace Scoping

Kaneo is workspace-scoped. Most resources belong to a workspace, and many
endpoints require a `workspaceId` to operate.

## How to Pass Workspace ID

There are three patterns depending on the endpoint:

### 1. Query parameter (list endpoints)

`GET /api/project?workspaceId={workspaceId}`

Used by: `listProjects`

### 2. Body field (create endpoints)

```
POST /api/project
{ "workspaceId": "…", "name": "…", "icon": "…", "slug": "…" }
```

Used by: `createProject`

### 3. Inferred from resource ID (get/update/delete by ID)

When operating on an existing resource by its ID (e.g., `GET /api/project/:id`),
the workspace is inferred from the resource itself. You do NOT need to pass
workspaceId in these cases.

Used by: all project and task endpoints that use `:id` or `:projectId` params.

## Finding a Workspace ID

- The user likely knows their workspace ID.
- If not, there may be a `GET /api/workspace` endpoint (not yet confirmed).
- The workspace ID is visible in the URL when using the Kaneo web UI:
  `https://kaneo.example.com/workspace/{workspaceId}/...`
