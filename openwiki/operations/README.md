# Operations — Setup, Running, and Configuration

This page covers how to install, configure, and operate the Symphony tool suite and Maestro planning server.

## Setup

The single `setup.sh` script handles everything:

```bash
# Full setup
./setup.sh

# Preview mode
./setup.sh --dry-run

# Custom Maestro directory
./setup.sh --maestro-dir /custom/path/maestro

# Custom shell config file
./setup.sh --config-file-path ~/.bashrc
```

### What setup.sh does

1. **Symlinks skills** — Each `skills/<name>/` directory is linked to `~/.config/maki/skills/<name>/`
2. **Symlinks config files** — `AGENTS.md`, `init.lua` (if present), and any `commands/` and `providers/` files
3. **Builds Maestro** — Runs `go build -o maestro .` in the `maestro/` directory
4. **Configures PATH** — Adds the Maestro binary to PATH via shell rc file with a marker comment for idempotency

### Idempotency

The script is safe to re-run:
- Existing files are replaced with updated symlinks
- The PATH block in your rc file uses a marker comment (`# Maestro Planning Server`) and is only added if absent
- Non-existent source files are silently skipped

## Prerequisites

- **Go 1.26+** — for building Maestro (see `maestro/go.mod`)
- **Bash** — for running `setup.sh` and helper scripts
- **[fswatch](https://emcrisostomo.github.io/fswatch/)** — recommended for zero-latency file change detection in Maestro listen scripts. Falls back to `stat` polling if unavailable.

## Running Maestro

```bash
# From any directory if setup.sh was run
maestro

# Or from the repo directory
cd maestro && go build -o maestro . && ./maestro

# Or using go run
cd maestro && go run .
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `MAESTRO_PLANS_DIR` | `plans` | Directory with `.toon` files (relative to CWD or absolute) |
| `MAESTRO_DIR` | Binary dir or CWD | Directory containing `templates/` and `static/` assets |

The server runs until Ctrl+C. It logs connections, file changes, and errors to stdout.

### Checking the Server

```bash
# Plans listing
curl http://localhost:8080/api/plans

# A specific plan (replace demo with actual plan ID)
curl http://localhost:8080/api/plan/test-all-modules
```

## Feedback Session Workflow

A feedback session lets an AI agent present a plan to a human reviewer and get real-time updates.

### 1. Start the Server

```bash
maestro &
```

### 2. Start the Heartbeat

The agent runs the heartbeat script to indicate it is listening:

```bash
scripts/maestro-heartbeat.sh --plan-name <plan-id> --port 8080 &
```

This sends `POST /api/agent/<plan-id>/heartbeat` every 15 seconds. The server shows a solid blue dot in the UI for this plan.

### 3. Start the Listen Loop

The agent runs the listen script to block until the plan changes:

```bash
scripts/maestro-listen.sh --plan-name <plan-id> --port 8080 --timeout 7200
```

When the human edits the plan file, the script detects the change (via `fswatch` or `stat` polling), fetches the updated JSON via the API, and outputs it to stdout. The agent can then read the changes and respond.

### 4. Stop the Session

```bash
scripts/maestro-heartbeat.sh --plan-name <plan-id> --stop
```

### Script Reference

#### maestro-heartbeat.sh

```
--plan-name <name>    Plan name (required, matches .toon filename without extension)
--maestro-dir <path>  Path to maestro directory (default: MAESTRO_DIR env or .)
--port <port>         Maestro server port (default: 8080)
--interval <s>        Seconds between heartbeats (default: 15)
--timeout <s>         Max seconds to run (default: 0 = forever)
-s, --stop            Stop a running heartbeat for this plan
```

The script saves its PID to `.maestro-hb-<plan-name>.pid` in the maestro directory. The `--stop` flag reads this file and kills the process.

#### maestro-listen.sh

```
--plan-name <name>    Plan name (required)
--maestro-dir <path>  Path to maestro directory
--port <port>         Maestro server port (default: 8080)
--timeout <s>         Max seconds to wait (default: 7200)
--poll-fallback-sleep <s>  Seconds between stat polls (default: 2)
```

Detection method:
1. **fswatch** — if installed, uses `fswatch -1 --latency 0.5` for instant notification
2. **stat polling** — falls back to `stat -f %m` with configurable interval

The script checks if the plan file exists before starting. Exit code 2 means timeout.

## Plan File Management

### Creating a Plan

Create a `.toon` file in the plans directory:

```toon
title: My Plan
summary: Description of the plan
state: draft

modules[N]:
  - heading: Section Name
    type: <module-type>
    items[N]{fields}:
      ...
```

See the [TOON format](../toon/README.md) and the [Maestro data model](../maestro/README.md#plan-data-model) for full syntax.

### Editing a Plan

Edit the `.toon` file with any text editor. The file watcher detects changes and broadcasts updates over WebSocket to connected browsers.

### Deleting a Plan

Delete the `.toon` file. The server will still serve it from memory until restart. There is no API endpoint for deleting plans — it must be done on disk.

## Troubleshooting

**"Maestro not found" after setup**
Run `source ~/.zshrc` (or your shell's rc file) or open a new terminal.

**"Templates not found" when running Maestro**
Set `MAESTRO_DIR` to the `maestro/` directory: `MAESTRO_DIR=/path/to/maestro maestro`

**WebSocket connections not working**
Ensure the Go version supports `http.HandleFunc` with `r.PathValue` (Go 1.22+). The `maestro/go.mod` specifies Go 1.26.4.

**"Plan not found" in the UI**
Plans are loaded from `MAESTRO_PLANS_DIR` (default `plans/`). Place `.toon` files there. Files are loaded by basename without extension — `plans/demo.toon` becomes plan ID `demo`.

**File watcher errors**
The server logs `file watcher: ... (live updates disabled)` if `fsnotify` fails (e.g., on some network filesystems). The server continues to work, but plan changes won't auto-reload until the next HTTP request triggers a re-read.

## Data Directory

Plans are stored as `.toon` files. Maestro creates the plans directory on startup if it doesn't exist. There is no database — all state is in these files. Messages are persisted inside the `.toon` file alongside the plan data through the TOON round-trip encoding.
