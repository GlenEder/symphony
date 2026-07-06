---
name: maestro
description: Using the Maestro planning server — build, run, manage plans, and use its API for AI-driven planning workflows.
compatibility: opencode
---

## Purpose

Maestro is a lightweight Go web server that serves structured planning documents from TOON-encoded files. It provides a web UI, a JSON API, and WebSocket-based live reload for plans. Plans are composed of typed modules (criteria, steps, risks, assumptions, changes, notes, questions) and stored as `.toon` files in a configurable directory.

## Quick Start

```bash
cd maestro

# Build and run with default settings (port 8080, plans/ directory)
go build -o maestro . && ./maestro

# Or run directly
go run .
```

### Configuration

| Env Var     | Default   | Description               |
|-------------|-----------|---------------------------|
| `PORT`      | `8080`    | HTTP server port          |
| `PLANS_DIR` | `plans`   | Directory with `.toon` files |

## Plan Data Model

A plan is a TOON file with the following structure:

```toon
title: Plan Title
summary: Short description of the plan
state: draft|approved

modules[N]:
  - heading: Module Heading
    items[N]{field1,field2,…}:
      # Data rows using tabular tuple format
    type: <module-type>

# Messages are stored separately and not part of the TOON file (they are persisted as JSON).
# See the Messages API below.
```

### Tabular Tuple Format

Items use TOON's tabular tuple format for compactness:

```toon
items[4]{text}:
  All existing data is preserved after migration
  "Read replicas sync within 5 seconds of primary"
  "Rollback completes in under 15 minutes"
  "All application integration tests pass against new database"
```

Multi-field items:

```toon
items[3]{impact,mitigation,severity,text}:
  "Loss of transactions during the final sync window","Run in read-only mode for 15 minutes before cutover",high,"Replication lag could cause data inconsistency during cutover"
```

**Important**: The data rows must be **indented deeper** (2 more spaces) than the `items[N]{fields}:` header line.

### Module Types

| Type          | Purpose                              |
|---------------|--------------------------------------|
| `criteria`    | Acceptance criteria (checkbox list)  |
| `steps`       | Implementation steps (numbered list) |
| `risks`       | Risk items with severity/impact      |
| `assumptions` | Assumptions being made               |
| `changes`     | Files or resources that change       |
| `notes`       | Freeform notes                       |
| `questions`   | Open questions with answered/answer  |

| Field      | Type   | Description                                      |
|------------|--------|--------------------------------------------------|
| `title`    | string | Plan title (required)                            |
| `summary`  | string | Short description of the plan                    |
| `state`    | string | Plan state: `draft` or `approved`                |
| `modules`  | array  | Plan module list                                 |

Messages are stored alongside the plan in-memory and persisted to the `.toon` file on disk. They are not part of the raw TOON format — they are added by the server during load/save round-trip.

### Item Fields by Module Type

**All module types** use `text` (required) as the primary description.

**`risks` items** additionally support:
- `severity`: `high`, `medium`, or `low`
- `impact`: description of the potential impact
- `mitigation`: how the risk is mitigated

**`steps` items** additionally support:
- `status`: `pending`, `in-progress`, `done`, `blocked`
- `owner`: who is responsible

**`changes` items** additionally support:
- `type`: the kind of change (e.g. `terraform`, `config`, `docs`)

**`questions` items** additionally support:
- `answered`: `true` or `false`
- `answer`: the resolved answer (when `answered: true`)

## Plan File Storage

- Plans are stored as `.toon` files in the `PLANS_DIR` directory.
- The file name (without `.toon` extension) becomes the plan ID used in URLs and API calls.
- The server watches the directory for file changes and reloads automatically.
- When a plan file changes, the server broadcasts the updated plan via WebSocket to all connected clients viewing that plan.

### Example: `plans/demo.toon`

```toon
title: Database Migration Plan
summary: "Migrate from PostgreSQL 12 to PostgreSQL 15 with zero downtime across all environments."
state: draft

modules[7]:
  - heading: Acceptance Criteria
    items[4]{text}:
      All existing data is preserved after migration
      "Read replicas sync within 5 seconds of primary"
      "Rollback completes in under 15 minutes"
      "All application integration tests pass against new database"
    type: criteria
  - heading: Implementation Steps
    items[6]{owner,status,text}:
      infra-team,done,"Provision PostgreSQL 15 instance in staging"
      infra-team,pending,"Configure logical replication between old and new instances"
      app-team,pending,"Run schema compatibility checks on all databases"
      app-team,pending,"Switch read traffic to new instance and monitor"
      both,blocked,"Switch write traffic during maintenance window"
      infra-team,pending,"Decommission old PostgreSQL 12 instance"
    type: steps
  - heading: Risks
    items[3]{impact,mitigation,severity,text}:
      "Loss of transactions during the final sync window","Run in read-only mode for 15 minutes before cutover",high,"Replication lag could cause data inconsistency during cutover"
      "Services unable to connect to new database","Use a DNS alias so the connection string remains unchanged",medium,"Application connection strings need updates across all services"
      "Some advanced features may be temporarily unavailable","Verify all extensions are compatible with PG15 ahead of time",low,"Minor PostgreSQL extension version mismatch"
    type: risks
  - heading: Assumptions
    items[3]{text}:
      "Application uses connection pooling via PgBouncer"
      "No schema migrations will be deployed during the migration window"
      "Network latency between old and new instances is under 1ms"
    type: assumptions
  - heading: Changes Required
    items[4]{text,type}:
      infra/terraform/database.tf,terraform
      config/deploy.yml,config
      docker-compose.yml,config
      docs/runbooks/migration.md,docs
    type: changes
  - heading: Notes
    items[2]{text}:
      "Coordinate with DevOps team to schedule the maintenance window. Suggested window is Saturday 02:00–04:00 UTC."
      "Run the migration script with --dry-run first to verify all steps before the actual cutover."
    type: notes
  - heading: Open Questions
    items[3]{answer,answered,text}:
      "Yes — keep for 30 days at reduced cost (scale down to minimal size).",true,"Should we keep the old PG12 instance running for 30 days as a fallback?"
      "Maximum 5 seconds lag before we abort the cutover.",true,"What is the acceptable replication lag threshold for cutover?"
      ,false,"Do we need to update any monitoring dashboards or alerts?"
    type: questions
```

## API Endpoints

All API routes return JSON.

### List Plans

```
GET /api/plans
```

Response:
```json
[
  {"id": "demo", "title": "Database Migration Plan", "summary": "Migrate from PostgreSQL 12 to PostgreSQL 15..."}
]
```

### Get Plan

```
GET /api/plan/{id}
```

Response is a flat JSON structure:
```json
{
  "title": "Database Migration Plan",
  "summary": "...",
  "state": "draft",
  "messages": [],
  "modules": [
    {
      "type": "criteria",
      "heading": "Acceptance Criteria",
      "items": [
        {"text": "All existing data is preserved after migration"}
      ]
    }
  ]
}
```

### Add a Message (Conversation Thread)

```
POST /api/plan/{id}/messages
Content-Type: application/json

{"role": "human", "text": "Your feedback", "item_ref": "2:1"}
```

- `role`: `"agent"` or `"human"`
- `text`: message body (required)
- `item_ref`: optional positional reference `"moduleIndex:itemIndex"` (e.g., `"2:1"` = module 2, item 1)

Returns the created message:
```json
{
  "id": "msg_18bfc3e196bafae0",
  "role": "human",
  "text": "Your feedback",
  "item_ref": "2:1",
  "created_at": "2026-07-06T17:35:51Z"
}
```

The message is appended to the plan's conversation thread and the plan is persisted.

### Set Plan State

```
POST /api/plan/{id}/state
Content-Type: application/json

{"state": "approved"}
```

Valid states: `"draft"`, `"approved"`.
Returns the full updated flat JSON plan.

### WebSocket (Live Updates)

```
ws://host/ws/plan/{id}
```

When the plan file is modified, the server sends the full flat JSON plan over the WebSocket. The client can then reload or patch the view.

## Web UI Routes

| Route              | Description                                    |
|--------------------|------------------------------------------------|
| `/`                | Redirects to `/plans`                          |
| `/plans`           | Plan listing page                              |
| `/plan/{id}`       | Plan detail page with modules, sidebar, messages |

### Example: `plans/regression-suite.toon`

```toon
title: Regression Test Suite Migration
summary: Migrate end-to-end regression tests from Cypress to Playwright across all product modules.

modules[6]:
  - type: criteria
    heading: Acceptance Criteria
    items[5]:
      - text: All existing Cypress test scenarios have an equivalent Playwright test
        answered: true
      - text: CI pipeline runs Playwright suite and reports results to GitHub Checks
      - text: Playwright tests pass consistently across Chrome, Firefox, and WebKit
      - text: Test execution time does not exceed current Cypress runtime by more than 20%
      - text: Flaky test detection and auto-retry is configured for all new Playwright tests

  - type: steps
    heading: Migration Steps
    items[7]:
      - text: Audit existing Cypress test suite and catalog all test scenarios
        owner: qa-team
        status: done
      - text: Set up Playwright project with TypeScript, ESLint, and Prettier
        owner: qa-team
        status: done
      - text: Migrate authentication and session tests (12 scenarios)
        owner: qa-team
        status: in-progress
      - text: Migrate checkout flow tests (8 scenarios)
        owner: qa-team
        status: pending
      - text: Migrate search and browse tests (15 scenarios)
        owner: qa-team
        status: pending
      - text: Update CI pipeline to run Playwright and remove Cypress step
        owner: devops
        status: pending
      - text: Run full regression suite in staging for 5 consecutive days before cutover
        owner: both
        status: pending

  - type: risks
    heading: Risks
    items[4]:
      - text: Playwright may not support certain Cypress-specific plugins (cy.route, cy.intercept patterns)
        severity: medium
        impact: Some tests may need redesign or alternative approaches
        mitigation: Audit plugin usage before migration begins; prototype alternatives early
      - text: Flaky E2E tests could erode team confidence in the new suite
        severity: high
        impact: Teams may push back on Playwright adoption
        mitigation: Configure Playwright flaky test detection with 3x auto-retry; track flake rate in dashboard
      - text: Test execution time could regress due to Playwright's multi-browser matrix
        severity: medium
        impact: CI pipeline duration increases, blocking developer velocity
        mitigation: Run browsers in parallel via Playwright sharding; set a hard 10-minute CI limit
      - text: Team lacks Playwright expertise initially
        severity: low
        impact: Migration takes longer than estimated
        mitigation: Pair a senior QA engineer on first 3 migration sprints; schedule workshop

  - type: assumptions
    heading: Assumptions
    items[4]:
      - text: All current test environments are compatible with Playwright browser binaries
      - text: The team can maintain both test suites during a 2-week parallel run period
      - text: No major application architecture changes are planned during the migration window
      - text: Playwright's request interception API covers all existing mock scenarios

  - type: changes
    heading: Changes Required
    items[5]:
      - text: e2e/playwright/ (new directory)
        type: config
      - text: e2e/cypress/ (remove after migration)
        type: config
      - text: .github/workflows/e2e.yml
        type: config
      - text: playwright.config.ts
        type: config
      - text: docs/testing/regression-checklist.md
        type: docs

  - type: notes
    heading: Notes
    items[3]:
      - text: Playwright's codegen tool (`npx playwright codegen`) should be used to accelerate initial test scaffolding.
      - text: Create a shared test fixture factory to reduce duplication across migrated scenarios.
      - text: Flag any test that requires visual regression comparison — those may need percy.io or Playwright's built-in snapshot.
```

## Workflow: Creating and Editing Plans

1. Create a `.toon` file in the `plans/` directory using the tabular tuple format shown above.
2. The server loads it automatically on startup (or on next request if placed while running).
3. The file watcher detects edits and triggers a WebSocket broadcast to connected clients.
4. To update a plan, edit the `.toon` file — the server picks up changes live.
5. To add feedback or comments, use the API's messages endpoint or the sidebar in the web UI.

## Code Layout

```
maestro/
├── main.go              # Entry point, env config, server setup
├── handler.go           # HTTP route handlers (UI + JSON API)
├── model.go             # Data model types (Plan, Module, Item)
├── store.go             # PlanStore — loads, caches, saves plans to disk
├── watcher.go           # File system watcher (fsnotify) for live reload
├── ws.go                # WebSocket hub for plan update broadcasts
├── go.mod / go.sum      # Go module dependencies
├── templates/           # Go html/template files
│   ├── base.html        # Layout shell
│   ├── plan.html        # Plan detail page + WebSocket client + response box
│   ├── plans.html       # Plan listing page
│   └── components/      # Reusable template components per module type
├── static/              # Static assets
│   ├── style.css        # Styling (includes response box styles)
│   └── script.js        # Client-side JS
└── plans/               # Default plans directory
    └── demo.toon        # Example plan
```

## Dependencies

- Go 1.26+
- `github.com/fsnotify/fsnotify` — file system notifications
- `github.com/gorilla/websocket` — WebSocket support
- `github.com/sstraus/toon_go/toon` — TOON format parser

## TOON Format Notes

TOON is Token-Oriented Object Notation — a compact, schema-aware JSON encoding. The `toon_go` library parses `.toon` files and the server converts them to JSON internally. Key syntax:

- `key: value` for simple fields
- `key[N]:` for arrays followed by indented items
- `key[N]{fields}:` for tabular tuple arrays with data rows on subsequent lines (must be indented deeper)
- `- field: value` for array elements in list format
- Indentation (2 spaces) is significant — it defines nesting

## Use Cases

- **AI-driven planning**: An agent can write a `.toon` file to the plans directory, and the Maestro server immediately serves it with live reload.
- **Plan review workflow**: Edit plans in the filesystem while viewing them in the browser — changes propagate instantly via WebSocket.
- **CI/CD plan publishing**: Automated pipelines can write `.toon` plan files and the server serves them without restart.
