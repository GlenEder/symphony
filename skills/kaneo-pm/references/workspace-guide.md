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

Three methods:

### 1. API Discovery (Recommended)

Use the Better Auth organizations endpoint to list all workspaces the API key has access to:

```bash
curl -s "$KANEO_API_URL/auth/organization/list" \
  -H "Authorization: Bearer $KANEO_API_KEY"
```

Response:
```json
[
  {
    "id": "wksp_abc123",
    "name": "My Workspace",
    "slug": "my-workspace",
    "createdAt": "2026-06-18T22:53:43.889Z"
  }
]
```

### 2. Web UI URL

The workspace ID is visible in the URL when using the Kaneo web UI:
  `https://kaneo.example.com/workspace/{workspaceId}/...`

### 3. Ask the User

Fallback if other methods fail.
