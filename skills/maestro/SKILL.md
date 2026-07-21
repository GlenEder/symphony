---
name: maestro
description: Run the Maestro planning server to publish a plan as a JSON file and drive the live feedback loop with the user until the plan is approved. Use when the user wants to publish a plan for review, mentions Maestro, plan files, the planning server, or "approve the plan", or when another skill needs to hand a plan to the user for review.
compatibility: opencode
---

## Purpose

Maestro is a lightweight Go web server that serves structured planning documents from JSON plan files.
It gives a plan a web UI, a JSON API, and WebSocket live reload, so an agent can publish a plan and then run a feedback loop with the user until the plan is approved.

Plans are composed of typed modules — criteria, steps, risks, assumptions, changes, notes, questions — stored as `.json` files in a configurable directory.

## Quick Start

Make sure `maestro` is on your PATH (run `./setup.sh --add-to-path` from the repo root):

```bash
# Start the server from any directory
maestro

# Or with custom settings
PORT=9090 MAESTRO_PLANS_DIR=/tmp/plans maestro
```

Before starting a new server, always check for one already running and reuse it — run `scripts/maestro-discover.sh` (see [Start the session](#1-start-the-session) below).
Three scripts in `scripts/` support the feedback session — `maestro-discover.sh` finds an already-running server to reuse, `maestro-heartbeat.sh` keeps the server informed the agent is **listening**, and `maestro-listen.sh` watches the plan file and returns plan JSON on change.
Run all three from the project root or `maestro/` dir; see `--help` on each for all flags.

> `--plan-name` is the primary flag for selecting a plan (`--plan-id` is an alias).

### Configuration

| Env var | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `MAESTRO_PLANS_DIR` | `plans` | Directory with `.json` files |
| `MAESTRO_DIR` | Binary dir or CWD | Directory containing `templates/` and `static/` assets; also the default for the scripts' `--maestro-dir` flag |
| `MAESTRO_POLL_INTERVAL` | `500ms` | File polling interval (e.g. `100ms`, `2s`) |

## Plan Data Model

A plan is a JSON file with these fields:

| Field | Type | Description |
|---|---|---|
| `title` | string | Plan title (required) |
| `summary` | string | Short description |
| `state` | string | `draft` or `approved` |
| `modules` | array | Plan module list |

```json
{
  "title": "Plan Title",
  "summary": "Short description of the plan",
  "state": "draft",
  "modules": [
    {
      "type": "criteria",
      "heading": "Module Heading",
      "items": [
        {"text": "First item"},
        {"text": "Second item"}
      ]
    }
  ]
}
```

Each module has a `type` and `heading`, plus a `items` array of objects whose fields depend on the module type (see [Module Types](#module-types) below).

For module type field reference and worked examples, see [`GLOSSARY.md`](GLOSSARY.md).

### Module Types

A plan's `modules` are typed. **Bold types** below are shown off in [`GLOSSARY.md`](GLOSSARY.md) — each type's fields and a worked example in both tuple and list form; consult it when authoring a module.

| Type | Purpose |
|---|---|
| **criteria** | Acceptance criteria (checkbox list) |
| **decision** | Resolved decisions with alternatives and rationale |
| **steps** | Implementation steps (numbered list) |
| **risks** | Risk items with severity/impact |
| **assumptions** | Assumptions being made |
| **changes** | Files or resources that change |
| **notes** | Freeform notes |
| **questions** | Open questions with answered/answer |

All types use `text` (required) as the primary description; the other fields vary by type and are shown alongside each example in the glossary.

### Plan File Storage

- Plans are stored as `.json` files in `MAESTRO_PLANS_DIR`.
- The file name (without `.json`) is the plan ID used in URLs and API calls.
- The server polls for `.json` changes (default 500ms) and reloads automatically, broadcasting to connected clients via WebSocket.
- Server-initiated writes are tracked and skipped by the poller to avoid redundant reloads.
- `POST /api/admin/reload` forces an immediate full rescan.

See `examples/demo.json` and `examples/regression-suite.json` for worked examples.

## API

The feedback loop only needs these endpoints; the full reference is in [`API.md`](API.md).

### Create or update a plan

```
POST /api/plan/{id}
Content-Type: application/json

{ full plan JSON body }
```

Creates a new plan or overwrites an existing one (upsert).
Returns the full plan JSON including any existing messages.

### Partial plan update

```
PATCH /api/plan/{id}
Content-Type: application/json

{"title": "...", "summary": "...", "state": "...", "modules": [...]}
```

Only specified fields are updated. Messages are always preserved.
Returns the full updated plan.

### Add a message

```
POST /api/plan/{id}/messages
Content-Type: application/json

{"role": "agent"|"human", "text": "...", "item_ref": "moduleIndex:itemIndex"}
```

`item_ref` is optional, e.g. `"2:1"` = module 2, item 1.
Returns the created message with `id` and `created_at`; the plan is persisted.

### Set plan state

```
POST /api/plan/{id}/state
{"state": "approved"}
```

Valid states: `"draft"`, `"approved"`.
Returns the full updated plan.

### Set agent status

```
POST /api/agent/{id}/status
{"status": "offline"}
```

The agent dot has three states — **listening** (solid blue, awaiting user input), **thinking** (pulsing blue, the server sets this when a human message arrives), and **offline** (done).
The server drives `listening`/`thinking` automatically from message roles; the agent sets `offline` explicitly when the plan is approved.

## Workflow: Authoring a Plan

1. POST the plan JSON to the server:
   ```bash
   curl -s -X POST "http://localhost:$port/api/plan/{plan-id}" \
     -H "Content-Type: application/json" \
     -d '{ "title": "...", "summary": "...", "state": "draft", "modules": [...] }'
   ```
   The server validates the JSON and returns the parsed plan.
2. To update a plan's structure during the feedback loop, use `PATCH /api/plan/{id}`:
   ```bash
   curl -s -X PATCH "http://localhost:$port/api/plan/{plan-id}" \
     -H "Content-Type: application/json" \
     -d '{ "modules": [...] }'
   ```
3. After bulk external edits, call `POST /api/admin/reload` to force an immediate rescan.

Done when: `GET /api/plan/{id}` returns the plan.

## Feedback Session Workflow

After authoring a plan, run a feedback session so the user can review, comment, and approve before implementation begins.
The agent dot is the session's heartbeat: it is **listening** while you wait, **thinking** while you respond, and **offline** when the plan is approved.

### 1. Start the session

Always check for an existing server before starting one — a maestro server left running from a previous session must be reused, not duplicated.
Run the discovery helper from the project root or `maestro/` dir:

```bash
port=$(scripts/maestro-discover.sh --port 8080 --max-port 8089 2>/dev/null || echo "")
```

It probes ports 8080–8089 and prints the first one whose `GET /api/plans` returns a JSON array, exiting `1` if none is found.
Then:

1. If `port` is non-empty, reuse the live server — set `started_server=false`.
   Do not start another server, and do not stop this one when the session ends (you did not start it).
2. If `port` is empty, start a fresh server (assuming `maestro` is on your PATH — run `setup.sh` if not):
   ```bash
   port=8080
   maestro &
   while ! curl -s "http://localhost:$port/api/plans" > /dev/null 2>&1; do sleep 0.2; done
   started_server=true
   ```
   The server defaults to port 8080 (`PORT` env to override); discovery scans 8080–8089 so a non-default port in that range is still reused.
 3. POST your plan JSON to the server:
    ```bash
    curl -s -X POST "http://localhost:$port/api/plan/{plan-id}" \
      -H "Content-Type: application/json" \
      -d '{ "title": "...", "summary": "...", "state": "draft", "modules": [...] }'
    ```
    The server writes the plan file and returns the parsed plan. No filesystem path knowledge needed — the server owns its plans directory.
 4. Open the plan in the browser:
   ```bash
   open "http://localhost:$port/plan/{plan-id}"
   ```
5. Tell the user:
   > The plan is ready for review at http://localhost:$port/plan/{plan-id}
   > You can comment on individual items by clicking them, send general feedback in the discussion sidebar, and click "Approve Plan" when satisfied.
   > I'll wait here for your feedback.

Carry `$port` and `$started_server` through every later step — all API calls, the heartbeat, and the listen loop use `--port "$port"`.

Done when: a server is reachable (reused or started), the plan file is written and confirmed via `GET /api/plan/{plan-id}`, the browser is open, and the user has been told where to review.

### 2. Start the heartbeat

Before entering the listen loop, start a background heartbeat so the server tracks the agent as **listening**:

```bash
scripts/maestro-heartbeat.sh --plan-name "{plan-name}" --port "$port" --interval 15
```

This runs in the background and persists across listen loop iterations.
Stop it when the plan is approved (pass `--port "$port"` so the offline status hits the right server):

```bash
scripts/maestro-heartbeat.sh --plan-name "{plan-name}" --port "$port" --stop
```

Done when: the heartbeat process is running (its PID is saved to `.maestro-hb-{plan-name}.pid`).

### 3. Run the listen loop

The plan file on disk is rewritten by the server whenever a message is added or the state changes.
Watch it with `maestro-listen.sh` — it blocks until the file changes, then prints plan JSON and exits:

```bash
plan_json=$(scripts/maestro-listen.sh --plan-name "{plan-name}" --port "$port" --timeout 7200)
```

Exit codes: `0` = change detected (plan JSON on stdout), `1` = error, `2` = timeout.

Done when: `maestro-listen.sh` returns plan JSON on stdout.

### 4. Process changes

Parse the returned JSON with `jq` or inline bash.
Compare against the previous state (tracked in variables you maintain across iterations).
Detect:

1. **New human messages** — any message where `role == "human"` and `id` is not in your set of previously seen IDs.
2. **State change** — `state == "approved"` means the user approved the plan.

Done when: every new human message has been read and `last_seen_msg_ids` updated to include it.

### 5. Respond to feedback

For each new human message:

1. If `item_ref` is set (e.g. `"2:1"` = module index 2, item index 1), resolve the referenced item from `plan.modules` for context.
2. Formulate a response addressing the feedback.
 3. If the feedback implies a plan change, update the plan via PATCH:
    ```bash
    curl -s -X PATCH "http://localhost:$port/api/plan/{plan-id}" \
      -H "Content-Type: application/json" \
      -d '{ "modules": [...] }'
    ```
    The server updates the file and broadcasts via WebSocket. Messages are preserved automatically.
 4. Post your response:
   ```bash
   curl -s -X POST "http://localhost:$port/api/plan/{plan-id}/messages" \
     -H "Content-Type: application/json" \
     -d '{"role": "agent", "text": "Good point. I'\''ve updated the risk section."}'
   ```
   The server sets the dot back to **listening** when you post an agent message.
5. Update `last_seen_msg_ids` immediately so you do not reprocess the same message.

Done when: every new human message has an agent reply posted and tracked.

### 6. Handle approval

When `plan.state == "approved"`:

1. Set the dot to **offline**:
   ```bash
   curl -s -X POST "http://localhost:$port/api/agent/{plan-id}/status" \
     -H "Content-Type: application/json" \
     -d '{"status": "offline"}'
   ```
2. Post a final acknowledgment message.
3. Stop the heartbeat: `scripts/maestro-heartbeat.sh --plan-name "{plan-name}" --port "$port" --stop`.
4. Exit the listen loop.
5. Proceed with implementation by entering the **plan-implementation-procedure** orchestration loop (implementer + reviewer subagents) on the approved plan.

The server also sets the agent **offline** automatically after 1 minute with no heartbeat; setting it explicitly gives the user immediate feedback.

Done when: the dot is offline, the heartbeat is stopped, the user has been acknowledged, and control has passed to the plan-implementation-procedure orchestration loop.

### 7. Handle interruption

If the user presses Ctrl-C during the file-watch (the bash call fails or interrupts):

1. Ask the user whether to resume, discard, or proceed anyway.
2. If resuming: restart the listen loop.
3. If discarding: stop the heartbeat, and stop the server only if `started_server=true` (never kill a server you reused).

Done when: the user has chosen a path and you have acted on it.

### Loop summary

```
1. Discover a running server (reuse it) or start one; write plan, open browser, tell the user.
2. Initialize last_seen_msg_ids from the plan's current messages.
3. Loop:
   a. plan_json=$(scripts/maestro-listen.sh --plan-name {plan-name} --port "$port" --timeout 7200)
   b. Parse $plan_json.
   c. For each msg in plan.messages where role=="human" and id not in last_seen_msg_ids:
      - Resolve item_ref against plan.modules for context.
      - POST /api/plan/{id}/messages (role="agent", text="...").
      - Add msg id to last_seen_msg_ids.
      (listening/thinking transitions are automatic.)
   d. If plan.state == "approved":
      - POST /api/agent/{id}/status {"status":"offline"}.
      - Stop the heartbeat.
      - Post final acknowledgment.
      - Break → proceed to the plan-implementation-procedure orchestration loop.
   e. If interrupted → ask the user what to do.
   f. Goto 3a.
```

## Edge Cases

| Scenario | Handling |
|---|---|
| **External .json edits** | The poller detects mtime changes within one interval cycle (default 500ms); `fswatch` is not required |
| **No `fswatch` available** | `maestro-listen.sh` falls back to `stat` polling (`--poll-fallback-sleep` flag); the server-side poller works regardless |
| **Server-initiated writes** | The poller skips them via `IsSelfWrite()` mtime matching, avoiding redundant reloads |
| **Direct .json mtime edits** | `POST /api/admin/reload` forces an immediate rescan |
| **User edits `.json` file directly** | `maestro-listen.sh` wakes on file change; detect module changes via API diff; acknowledge the edits |
| **User revokes approval** | `state` goes back to `draft` — stay in the loop; acknowledge the reversal |
| **Multiple rapid messages** | Process all new messages in batch on one wake; group related responses |
| **Message deleted** | Messages array shrinks — update your tracking set; no action needed |
| **Existing server on another port** | `maestro-discover.sh` scans 8080–8089 by default; widen with `--port`/`--max-port` if you run on a custom port |
| **Existing server, different plans dir** | The `POST /api/plan/{id}` endpoint writes to the server's own plans dir, so this mismatch never occurs |
| **Server crash** | Re-run `maestro-discover.sh`; if still unreachable, start a fresh server and re-check |
| **User idle / timeout** | After 30 minutes without any change, ask the user if they are still reviewing |

For the Go source layout and build dependencies, see [`DEVELOPMENT.md`](DEVELOPMENT.md).
