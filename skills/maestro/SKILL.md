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

modules[N]:
  - type: <module-type>
    heading: Module Heading
    items[N]:
      - text: Item description
        # optional fields per item:
        severity: high|medium|low
        impact: Description of impact
        mitigation: How to mitigate
        status: pending|in-progress|done|blocked
        owner: team-or-person
        answered: true|false
        answer: The answer text
        type: terraform|config|docs   # for changes module

# Optional top-level field for user-submitted responses
# response: Your feedback here
```

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

### Top-Level Fields

| Field      | Type   | Description                                      |
|------------|--------|--------------------------------------------------|
| `title`    | string | Plan title (required)                            |
| `summary`  | string | Short description of the plan                    |
| `response` | string | User-submitted response/feedback (optional)      |
| `modules`  | array  | Plan module list                                 |

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
summary: Migrate from PostgreSQL 12 to PostgreSQL 15 with zero downtime across all environments.

modules[7]:
  - type: criteria
    heading: Acceptance Criteria
    items[4]:
      - text: All existing data is preserved after migration
      - text: Read replicas sync within 5 seconds of primary
      - text: Rollback completes in under 15 minutes
      - text: All application integration tests pass against new database
        answered: true

  - type: steps
    heading: Implementation Steps
    items[6]:
      - text: Provision PostgreSQL 15 instance in staging
        owner: infra-team
        status: done
      - text: Configure logical replication between old and new instances
        owner: infra-team
        status: pending
      - text: Run schema compatibility checks on all databases
        owner: app-team
        status: pending
      - text: Switch read traffic to new instance and monitor
        owner: app-team
        status: pending
      - text: Switch write traffic during maintenance window
        owner: both
        status: blocked
      - text: Decommission old PostgreSQL 12 instance
        owner: infra-team
        status: pending

  - type: risks
    heading: Risks
    items[3]:
      - text: Replication lag could cause data inconsistency during cutover
        severity: high
        impact: Loss of transactions during the final sync window
        mitigation: Run in read-only mode for 15 minutes before cutover
      - text: Application connection strings need updates across all services
        severity: medium
        impact: Services unable to connect to new database
        mitigation: Use a DNS alias so the connection string remains unchanged
      - text: Minor PostgreSQL extension version mismatch
        severity: low
        impact: Some advanced features may be temporarily unavailable
        mitigation: Verify all extensions are compatible with PG15 ahead of time

  - type: assumptions
    heading: Assumptions
    items[3]:
      - text: Application uses connection pooling via PgBouncer
      - text: No schema migrations will be deployed during the migration window
      - text: Network latency between old and new instances is under 1ms

  - type: changes
    heading: Changes Required
    items[4]:
      - text: infra/terraform/database.tf
        type: terraform
      - text: config/deploy.yml
        type: config
      - text: docker-compose.yml
        type: config
      - text: docs/runbooks/migration.md
        type: docs

  - type: notes
    heading: Notes
    items[2]:
      - text: Coordinate with DevOps team to schedule the maintenance window. Suggested window is Saturday 02:00–04:00 UTC.
      - text: Run the migration script with --dry-run first to verify all steps before the actual cutover.

  - type: questions
    heading: Open Questions
    items[3]:
      - text: Should we keep the old PG12 instance running for 30 days as a fallback?
        answered: true
        answer: Yes — keep for 30 days at reduced cost (scale down to minimal size).
      - text: What is the acceptable replication lag threshold for cutover?
        answered: true
        answer: Maximum 5 seconds lag before we abort the cutover.
      - text: Do we need to update any monitoring dashboards or alerts?
        answered: false
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
  "response": "...",
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

### Submit Plan Response

```
POST /api/plan/{id}/response
Content-Type: application/json

{"text": "Your feedback or response to the plan"}
```

Returns the full updated flat JSON plan. The response is saved to the `.toon` plan file on disk, which triggers the file watcher and broadcasts the update to all WebSocket clients viewing that plan.

Response:
```json
{
  "title": "Database Migration Plan",
  "summary": "...",
  "response": "Your feedback or response to the plan",
  "modules": [...]
}
```

### WebSocket (Live Updates)

```
ws://host/ws/plan/{id}
```

When the plan file is modified, the server sends the full flat JSON plan over the WebSocket. The client can then reload or patch the view.

## Web UI Routes

| Route              | Description                              |
|--------------------|------------------------------------------|
| `/`                | Redirects to `/plans`                    |
| `/plans`           | Plan listing page                        |
| `/plan/{id}`       | Plan detail page with response box       |

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

1. Create a `.toon` file in the `plans/` directory following the TOON format above.
2. The server loads it automatically on startup (or on next request if placed while running).
3. The file watcher detects edits and triggers a WebSocket broadcast to connected clients.
4. To update a plan, edit the `.toon` file — the server picks up changes live.

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
- `- field: value` for array elements
- Indentation (2 spaces) is significant — it defines nesting

## Use Cases

- **AI-driven planning**: An agent can write a `.toon` file to the plans directory, and the Maestro server immediately serves it with live reload.
- **Plan review workflow**: Edit plans in the filesystem while viewing them in the browser — changes propagate instantly via WebSocket.
- **CI/CD plan publishing**: Automated pipelines can write `.toon` plan files and the server serves them without restart.
