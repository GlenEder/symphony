# Maestro — Planning Server

Maestro is a lightweight Go web server that serves structured planning documents. It provides a web UI for human reviewers, a JSON API for agents, and WebSocket-based live updates.

## Quick Start

```bash
# Start the server
maestro

# With custom port and plans directory
PORT=9090 MAESTRO_PLANS_DIR=/tmp/plans maestro

# Open in browser
open http://localhost:8080/plans
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `MAESTRO_PLANS_DIR` | `plans` | Directory containing `.toon` plan files |
| `MAESTRO_DIR` | Binary directory or CWD | Directory with `templates/` and `static/` assets |

The `MAESTRO_DIR` env var is particularly important when running `maestro` from outside the repo directory — it tells the server where to find HTML templates, CSS, and JavaScript. The `setup.sh` script automatically exports it.

## Plan Data Model

A plan is stored as a `.toon` file and loaded into a `Plan` struct.

### Core Types (from `maestro/model.go`)

```
Plan
├── title: string (required)
├── summary: string
├── state: "draft" | "approved"
├── updated_at: RFC3339 timestamp
├── messages: Message[]
│   └── Message { id, role ("agent"|"human"), text, item_ref ("moduleIndex:itemIndex"), created_at }
└── modules: Module[]
    └── Module { type, heading, items: Item[], columns?: Column[], hideRowNum?: bool }
        └── Item { text, severity, impact, mitigation, status, owner, answered, answer, changeType }
        └── Column { heading, key }
```

### Module Types

| Type | Purpose | Item-Specific Fields |
|------|---------|---------------------|
| `criteria` | Acceptance criteria (checkbox list) | `text` |
| `steps` | Implementation steps (numbered) | `text`, `status` (pending/in-progress/done/blocked), `owner` |
| `risks` | Risk items | `text`, `severity` (high/medium/low), `impact`, `mitigation` |
| `assumptions` | Assumptions being made | `text` |
| `changes` | Files or resources that change | `text`, `changeType` (terraform/config/docs/etc) |
| `notes` | Freeform notes | `text` |
| `questions` | Open questions | `text`, `answered` (bool), `answer` |
| `table` | Schema-driven table display | `text` (via Column key mapping) |

### Table Module

The `table` module type renders items in a structured table with user-defined columns:

```toon
- type: table
  heading: Encoding Validation
  columns[3]:
    - heading: Character
      key: text
    - heading: Name
      key: severity
    - heading: Unicode
      key: impact
  items[9]:
    - impact: U+2014
      severity: em dash
      text: —
```

Columns map item fields (`text`, `severity`, `impact`, etc.) to table headings. See `maestro/templates/components/module-table.html` for rendering.

## API Endpoints

All API routes return JSON. The base URL is `http://localhost:PORT`.

### Plan Listing

```
GET /api/plans
```
Returns `[{id, title, summary, updated_at}, ...]` sorted by recency.

### Get Plan

```
GET /api/plan/{id}
```
Returns a flat JSON plan with agent status.

### Add a Message

```
POST /api/plan/{id}/messages
Content-Type: application/json

{"role": "human", "text": "Your feedback", "item_ref": "2:1"}
```

- `role`: `"agent"` or `"human"`
- `text`: message body (required)
- `item_ref`: optional positional reference `"moduleIndex:itemIndex"` (e.g., `"2:1"`)

Returns the created message with server-assigned ID and timestamp.

### Delete a Message

```
DELETE /api/plan/{id}/messages/{msgId}
```
Returns the updated flat plan JSON.

### Set Plan State

```
POST /api/plan/{id}/state
Content-Type: application/json

{"state": "approved"}
```

Valid states: `"draft"`, `"approved"`. Returns the updated flat plan JSON.

### Agent Heartbeat

```
POST /api/agent/{id}/heartbeat
```
The agent sends this periodically (default every 15s) while listening. Returns `{"status":"ok"}`.

### Set Agent Status

```
POST /api/agent/{id}/status
Content-Type: application/json

{"status": "offline"}
```

Only `"offline"` is accepted from external callers. `"thinking"` and `"listening"` are auto-transitioned by the server based on message roles.

### Get Agent Status

```
GET /api/agent/{id}/status
```
Returns `{"status": "listening"|"thinking"|"offline"}`. Polled by the browser.

## WebSocket

```
ws://host/ws/plan/{id}
```

On connect, the server immediately sends the full flat plan JSON. On each file change or mutation, the server broadcasts the updated plan to all connected clients. The read loop detects disconnection and cleans up.

## Web UI Routes

| Route | Description |
|-------|-------------|
| `/` | Redirects to `/plans` |
| `/plans` | Plan listing with cards sorted by recency |
| `/plan/{id}` | Plan detail page with modules, discussion sidebar, commenting |

## Agent Status System

Maestro tracks per-plan agent presence with a finite state machine.

### States

| State | Visual | Meaning |
|-------|--------|---------|
| `listening` | Solid blue dot | Agent is running and receiving updates |
| `thinking` | Blinking blue dot | Human just sent a message, agent is processing |
| `offline` | Red dot | No heartbeat received in 10 minutes |

### Transitions

- **Agent starts heartbeat** → `listening`
- **Human sends message** → `thinking` (auto)
- **Agent sends message** → `listening` (auto)
- **No heartbeat for 10 minutes** → `offline` (GC goroutine)
- **Agent calls status API with "offline"** → `offline`

The GC goroutine runs every 5 seconds and marks any plan whose last heartbeat is older than 10 minutes as offline.

## File Watcher

Maestro uses `fsnotify` to watch the plans directory for changes. When a `.toon` file is written or created:
1. The file is reloaded into the PlanStore
2. The onChange callback fires
3. The Hub broadcasts the updated plan over WebSocket

The watcher is non-critical — if `fsnotify` fails (e.g., on some filesystems), live updates are disabled but the server continues to work.

## Templates

Maestro uses Go `html/template` with a two-tier system:
- **Base template** (`templates/base.html`) — layout with header, nav, footer
- **Page templates** (`templates/plans.html`, `templates/plan.html`) — page-specific content
- **Component partials** (`templates/components/*.html`) — reusable module renderers

Key components:
- `plan-module.html` — dispatches to type-specific component modules
- `module-*.html` — one per module type (steps, criteria, risks, etc.)
- `module-table.html` — schema-driven table module renderer
- `plan-header.html` — plan title, summary, state, actions
- `approval-panel.html` — approve/draft state toggle

## Source Map

| File | Responsibility |
|------|---------------|
| `main.go` | Server bootstrap, template parsing, route registration, dependency wiring |
| `handler.go` | HTTP handler functions for all routes, template rendering helpers |
| `model.go` | Plan/Module/Item/Message/FlatPlan types, JSON conversion |
| `store.go` | PlanStore — disk I/O, TOON encoding/decoding, message management |
| `ws.go` | AgentState, WebSocket Hub, upgrade handler, GC |
| `watcher.go` | fsnotify-based file watcher |
| `templates/` | `base.html`, page templates, component partials |
| `static/` | `style.css` (33KB, all-in-one stylesheet), `script.js` (stub) |

## Important Notes for Developers

- **No tests exist yet** — see [Testing](../testing/README.md) for guidance on adding test coverage.
- **The TOON library** at `maestro/lib/toon/` is a vendored Go port of the TypeScript `@toon-format/toon` library. It has its own `go.mod`. The main `maestro` module replaces the import path to point at this local copy.
- **Messages are persisted inside the .toon file** via a JSON round-trip (Plan → JSON → TOON → disk). This means `decodePlan` uses two steps: TOON decode to generic map, then JSON marshal/unmarshal to typed struct.
- **Positional references** (`item_ref: "moduleIndex:itemIndex"`) are 0-indexed and point into the current plan state. They are not stable across edits that reorder modules or items.
- **The flat plan** (`FlatPlan` struct) is the serialization format used by both the API and WebSocket. It includes an `agent_status` field not present in the raw Plan struct.
- **WebSocket connections** are per-plan. The Hub is a `map[planID]map[*websocket.Conn]bool`. Each plan detail page opens its own WebSocket connection.
