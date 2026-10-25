# Operations — Setup, Running, and Configuration

This page covers how to install, configure, and operate the Symphony tool suite and the legacy Maestro planning server (pending deprecation).

## Setup

The `setup` script is a dispatcher for modular subcommands:

```bash
# List all subcommands
./setup

# Preview mode for a subcommand
./setup skills --dry-run

# Only build the legacy Maestro server (with custom options)
./setup maestro --maestro-dir /custom/path/maestro --config-file-path ~/.bashrc

# Only symlink skills
./setup skills

# Install for Claude Code (skills into ~/.claude/skills/, AGENTS.md into ~/.claude/CLAUDE.md)
./setup skills --claude
./setup agents --claude

# Install skills for Cursor (~/.cursor/skills/)
./setup skills --cursor
```

Run `./setup -h` to list all subcommands.

### What the setup subcommands do

| Subcommand | Description |
|------------|-------------|
| `agents` | Copies `AGENTS.md` to `~/.config/maki/AGENTS.md` (default) or `~/.claude/CLAUDE.md` with `--claude` |
| `skills` | Symlinks each `skills/<name>/` directory to `~/.config/maki/skills/<name>/` (default), `~/.cursor/skills/` with `--cursor`, or `~/.claude/skills/` with `--claude` |
| `commands` | Symlinks each `commands/*` file to `~/.config/maki/commands/` |
| `providers` | Symlinks each `providers/*` file to `~/.config/maki/providers/` |
| `init` | Symlinks `init.lua` to `~/.config/maki/init.lua` |
| `maestro` | Builds the legacy Maestro server and configures PATH in shell rc file |

Running `./setup` with no arguments prints the help message; run each subcommand explicitly.

### Idempotency

The script is safe to re-run:
- Existing files are replaced with updated symlinks
- The PATH block in your rc file uses a marker comment (`# Maestro Planning Server`) and is only added if absent
- Non-existent source files are silently skipped

## Prerequisites

- **Bash** — for running `setup` and helper scripts
- **Go 1.26+** — for building the legacy Maestro server (see `maestro/go.mod`); not needed for the skill suite

## Running Maestro (Legacy)

The Maestro server is no longer part of the skill suite's planning workflow — it remains available as a standalone legacy server pending replacement by the plannotator project.

```bash
# From any directory if setup was run
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
| `MAESTRO_PLANS_DIR` | `plans` | Directory with `.json` plan files (relative to CWD or absolute) |
| `MAESTRO_DIR` | Binary dir or CWD | Directory containing `templates/` and `static/` assets |

The server runs until Ctrl+C. It logs connections, file changes, and errors to stdout.

### Checking the Server

```bash
# Plans listing
curl http://localhost:8080/api/plans

# A specific plan (replace demo with actual plan ID)
curl http://localhost:8080/api/plan/test-all-modules
```

## Plan File Management

### Creating a Plan

Create a `.json` file in the plans directory:

```json
{
  "title": "My Plan",
  "summary": "Description of the plan",
  "state": "draft",
  "modules": [
    {
      "type": "criteria",
      "heading": "Acceptance Criteria",
      "items": [{"text": "First criterion"}]
    }
  ]
}
```

See the [Maestro data model](../maestro/README.md#plan-data-model) for full syntax.

### Editing a Plan

Edit the `.json` file with any text editor. The file watcher detects changes and broadcasts updates over WebSocket to connected browsers.

### Deleting a Plan

Delete the `.json` file. The server will still serve it from memory until restart. There is no API endpoint for deleting plans — it must be done on disk.

## Troubleshooting

**"Maestro not found" after setup**
Run `source ~/.zshrc` (or your shell's rc file) or open a new terminal.

**"Templates not found" when running Maestro**
Set `MAESTRO_DIR` to the `maestro/` directory: `MAESTRO_DIR=/path/to/maestro maestro`

**WebSocket connections not working**
Ensure the Go version supports `http.HandleFunc` with `r.PathValue` (Go 1.22+). The `maestro/go.mod` specifies Go 1.26.4.

**"Plan not found" in the UI**
Plans are loaded from `MAESTRO_PLANS_DIR` (default `plans/`). Place `.json` files there. Files are loaded by basename without extension — `plans/demo.json` becomes plan ID `demo`.

**File watcher errors**
The server logs `file watcher: ... (live updates disabled)` if `fsnotify` fails (e.g., on some network filesystems). The server continues to work, but plan changes won't auto-reload until the next HTTP request triggers a re-read.

## Data Directory

Plans are stored as `.json` files. Maestro creates the plans directory on startup if it doesn't exist. There is no database — all state is in these files. Messages are persisted inside the `.json` file alongside the plan data.
