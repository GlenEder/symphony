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

### Scripts

Two scripts in `scripts/` help run feedback sessions:

**`maestro-heartbeat.sh`** — starts a background heartbeat so the server knows the agent is listening. Runs until stopped or the plan is approved.

```bash
scripts/maestro-heartbeat.sh --plan-name demo --port 8080
# Stop it when done:
scripts/maestro-heartbeat.sh --plan-name demo --stop
```

**`maestro-listen.sh`** — watches the plan file for changes using fswatch (or stat polling). Outputs plan JSON on change.

```bash
scripts/maestro-listen.sh --plan-name demo --port 8080 --timeout 7200
```

Run both from the project root or `maestro/` dir. See `--help` on each for all options.

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

## Feedback Session Workflow

After crafting a plan (via `plan-modules` or manually), start a feedback session so the user can review, comment, and approve before implementation begins.

### 1. Start the Session

1. Build the server (if not already running):
   ```bash
   cd maestro && go build -o maestro . && ./maestro &
   ```
   The server starts on port 8080 by default (`PORT` env to override).

2. Wait for the server to be ready:
   ```bash
   while ! curl -s http://localhost:8080/api/plans > /dev/null 2>&1; do sleep 0.2; done
   ```

3. Write your plan as a `.toon` file into `maestro/plans/{plan-id}.toon`.

4. Open the plan in the browser:
   ```bash
   open http://localhost:8080/plan/{plan-id}
   ```

5. Inform the user:
   > The plan is ready for review at http://localhost:8080/plan/{plan-id}
   > You can comment on individual items by clicking them, send general feedback in the discussion sidebar, and click "Approve Plan" when satisfied.
   > I'll wait here for your feedback.

### 2. Start the Heartbeat

Before entering the listen loop, start a background heartbeat so the server tracks the agent as **listening**:

```bash
scripts/maestro-heartbeat.sh --plan-name "{plan-name}" --port 8080 --interval 15
```

This runs in the background and persists across listen loop iterations. Stop it when the plan is approved:

```bash
scripts/maestro-heartbeat.sh --plan-name "{plan-name}" --stop
```

### 3. The Listen Loop

Now enter the feedback loop. The plan file on disk (`maestro/plans/{plan-name}.toon`) is updated by the Maestro server whenever a message is added or the state changes (because `persistPlan()` rewrites the file). Use `maestro-listen.sh` to watch for changes:

```bash
# Blocks until the plan file changes (zero token burn during idle)
# Outputs plan JSON on change, exits with code 2 on timeout.
scripts/maestro-listen.sh --plan-name "{plan-name}" --port 8080 --timeout 7200
```

**Flag reference (`maestro-listen.sh`):**

| Flag | Default | Description |
|------|---------|-------------|
  | `--plan-name` | *(required)* | Plan name to watch |
| `--maestro-dir` | `.` | Path to maestro directory |
| `--port` | `8080` | Maestro server port |
| `--timeout` | `7200` | Max seconds to wait (0 = no limit) |
| `--poll-fallback-sleep` | `2` | Seconds between stat polls (fallback only) |

Exit codes: `0` = change detected (plan JSON on stdout), `1` = error, `2` = timeout.

**Status transitions are automatic:**
- When the **user** sends a message, the server auto-sets the dot to **thinking** (pulsing blue).
- When the **agent** posts a response message, the server auto-sets the dot back to **listening** (solid blue).
- The agent only needs to set **offline** explicitly (e.g. when the plan is approved).
- No more manual `thinking`/`listening` status calls needed.

### 4. Process Changes

After `maestro-listen.sh` returns, the plan JSON is already on stdout (or pipe it in). Parse it with `jq` or inline bash. On each wake, compare against the previous state (tracked in variables you maintain throughout the loop). Detect:

1. **New human messages** — any message where `role == "human"` and `id` is not in your set of previously seen IDs.
2. **State change** — `state == "approved"` means the user approved the plan.

**Example: piping the listen output into your processor:**

```bash
plan_json=$(scripts/maestro-listen.sh --plan-name "{plan-name}" --port 8080 --timeout 7200)
# Now parse $plan_json...
```

### 4. Respond to Feedback

For each new human message you detect:

1. If `item_ref` is set (e.g. `"2:1"` = module index 2, item index 1), resolve the referenced item from the plan's `modules` array for context.
2. Formulate a response addressing the feedback.
3. If the feedback implies a plan change, update the `.toon` file directly (the server's file watcher reloads it and broadcasts via WebSocket).
4. Post your response via the messages API:
   ```bash
   curl -s -X POST http://localhost:8080/api/plan/{plan-id}/messages \
     -H "Content-Type: application/json" \
     -d '{"role": "agent", "text": "Good point. I\'ve updated the risk section." }'
   ```
   The server auto-sets the dot back to **listening** when you post an agent message.
5. After posting, immediately update your local `last_seen_msg_id` tracking so you don't reprocess the same message.

> **Note:** The thinking/listening transitions are now handled automatically by the server. You no longer need to call `POST /api/agent/{id}/status` for these.

### 5. Handle Approval

When `plan.state == "approved"`:

1. **Set status to `offline`** to indicate the agent is done:
   ```bash
   curl -s -X POST http://localhost:8080/api/agent/{plan-id}/status \
     -H "Content-Type: application/json" \
     -d '{"status": "offline"}'
   ```
2. Acknowledge to the user via a final agent message.
3. Exit the listen loop.
4. Proceed with implementation using the **plan-implementation-procedure** skill.

> **Note:** The server will also auto-set the agent to `offline` if no heartbeat is received for 1 minute. Setting it explicitly ensures immediate feedback.

### 6. Handle Cancellation (User Interrupt)

If the user presses Ctrl-C during the file-watch (the bash call fails/interrupts):

1. Ask the user what they want to do — resume the session, discard it, or proceed anyway.
2. If resuming: restart the listen loop.
3. If discarding: clean up (stop the server if you started it) and move on.

### 7. Full Loop Pseudocode

```
1. Start server, create plan, open browser
2. Initialize last_seen_msg_ids = {}  (from current plan messages)
3. Loop:
   a. Run maestro-listen.sh (handles heartbeat + file watch, returns plan JSON)
      plan_json=$(scripts/maestro-listen.sh --plan-id {id} --port 8080 --timeout 7200)
   b. Parse JSON from $plan_json
   c. For each msg in plan.messages where role=="human" and id not in last_seen_msg_ids:
      - Analyze: resolve item_ref if set, read context from plan.modules
      - Respond: POST /api/plan/{id}/messages (role="agent", text="...")
      - Update last_seen_msg_ids
      (thinking/listening transitions are automatic — no status calls needed)
   d. If plan.state == "approved":
      - Set offline (already done by maestro-listen.sh cleanup, but explicit is fine too):
        curl -X POST http://localhost:8080/api/agent/{id}/status -d '{"status":"offline"}'
      - Post final acknowledgment
      - Break loop → proceed to implementation
   e. If user interrupted → ask what to do
   f. Goto 3a
```

### 8. Configuration

| Setting | Default | Description |
|---|---|---|
| `PLANS_DIR` | `maestro/plans` | Directory with `.toon` files |
| `MAESTRO_PORT` | `8080` | Port for the Maestro server |

For listen loop script flags, run `scripts/maestro-listen.sh --help`.

### 9. Edge Cases

| Scenario | Handling |
|---|---|---|
| No `fswatch` available | `maestro-listen.sh` auto-falls back to `stat` polling (--poll-fallback-sleep flag) |
| User edits `.toon` file directly | `maestro-listen.sh` wakes on file change; detect module changes via API diff; acknowledge the edits |
| User revokes approval | `state` goes back to `draft` — stay in the loop; acknowledge the reversal |
| Multiple rapid messages | Process all new messages in batch on one wake; group related responses |
| Message deleted | Messages array shrinks — just update your tracking set; no action needed |
| Server crash | If the API is unreachable, try to restart the server and re-check |
| User idle / timeout | After 30 minutes without any change, ask the user if they're still reviewing |

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
