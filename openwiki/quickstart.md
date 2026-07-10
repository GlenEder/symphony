# Symphony — Quickstart

Symphony is a collection of tools and agent instructions for AI coding agents. It has three main layers:

1. **Skills** — Markdown instruction files (SKILL.md) that teach coding agents specialized workflows like GitHub operations, planning, project management, and Maestro interaction.
2. **Maestro** — A lightweight Go web server that serves structured planning documents with a web UI, JSON API, and WebSocket live updates.
3. **TOON** — A compact, token-efficient data format (Token-Oriented Object Notation) used by Maestro for plan files.

The project is designed for [Maki](https://github.com/gleneder/maki) but ports easily to other agent systems (Claude Code, Cline, Aider).

## Quick Setup

```bash
# From the repo root
./setup

# Preview without making changes
./setup --dry-run

# Only build Maestro
./setup maestro

# Only symlink skills
./setup skills --dry-run

# With custom config (maestro subcommand)
./setup maestro --maestro-dir /path/to/maestro --config-file-path ~/.bashrc
```

This symlinks skills into `~/.config/maki/skills/`, builds the Maestro binary, and adds `maestro` to your PATH. See [Operations](operations/README.md) for details.

Run `./setup -h` to list all available subcommands, or `./setup <command> -h` for subcommand-specific help.

## Start Using Maestro

```bash
# Start the planning server (from any directory if on PATH)
maestro

# Then open http://localhost:8080/plans in your browser
```

See [Maestro](maestro/README.md) for the full API, data model, and workflow.

## Repository Map

| Path | Description |
|------|-------------|
| `setup` | Installation dispatcher — run subcommands like `./setup maestro` |
| `setup_cmds/` | Modular setup subcommand scripts |
| `AGENTS.md` | Global agent baseline instructions (alwaysApply) |
| `maestro/` | Go planning server source, templates, static assets |
| `maestro/lib/toon/` | Go library for TOON encode/decode |
| `maestro/plans/` | Plan files (`.toon` format) |
| `skills/` | Agent skill definitions (one subdirectory per skill) |
| `skills/maestro/` | Maestro skill — teaches agents to use the planning server |
| `skills/toon/` | TOON format skill — teaches agents to encode/decode TOON |

## Documentation Sections

- [Architecture](architecture/README.md) — system design, data flow, component relationships
- [Maestro](maestro/README.md) — planning server, API, data model, agent status, WebSocket
- [Skills](skills/README.md) — agent skill definitions, structure, and how to use them
- [TOON Format](toon/README.md) — token-efficient data format used by Maestro
- [Operations](operations/README.md) — setup, running, environment variables, scripts

## Quick Links to Key Files

| File | What it does |
|------|-------------|
| `/maestro/main.go` | Server entrypoint, dependency wiring, route registration |
| `/maestro/handler.go` | HTTP handler functions and route registration |
| `/maestro/model.go` | Plan, Module, Item, Message, FlatPlan data types |
| `/maestro/store.go` | PlanStore — load, persist, manage plan lifecycle |
| `/maestro/ws.go` | AgentState, WebSocket Hub, live broadcasts |
| `/maestro/watcher.go` | File watcher for live plan reloads |
| `/setup` | Installation dispatcher |
| `/setup_cmds/` | Setup subcommand scripts |
| `/skills/maestro/SKILL.md` | Agent instructions for using Maestro |
| `/skills/toon/SKILL.md` | Agent instructions for using TOON |

## Next Steps

- Read [Architecture](architecture/README.md) to understand how the pieces fit together
- Read [Maestro](maestro/README.md) to learn the API and plan workflows
- Read [Skills](skills/README.md) to understand how agent skills work
- Read [TOON Format](toon/README.md) to understand plan file syntax
