# Symphony — Quickstart

Symphony is a collection of tools and agent instructions for AI coding agents. It has three main layers:

1. **Skills** — Markdown instruction files (SKILL.md) that teach coding agents specialized workflows like GitHub operations, planning, project management.
2. **Maestro** — A legacy Go planning server, retained in `maestro/` pending deprecation; no longer part of the planning workflow and slated for replacement by the plannotator project.
3. **TOON** — A compact, token-efficient data format (Token-Oriented Object Notation), documented at [toonformat.dev](https://toonformat.dev).

The project is designed for [Maki](https://github.com/gleneder/maki) but ports easily to other agent systems (Claude Code, Cline, Aider).

## Quick Setup

```bash
# List all subcommands (from the repo root)
./setup

# Only symlink skills (preview with --dry-run first)
./setup skills --dry-run
./setup skills

# Only build the legacy Maestro server
./setup maestro

# Install for Claude Code instead of Maki
./setup skills --claude
./setup agents --claude

# With custom config (maestro subcommand)
./setup maestro --maestro-dir /path/to/maestro --config-file-path ~/.bashrc
```

The install symlinks skills into `~/.config/maki/skills/`; the `./setup maestro` subcommand additionally builds the legacy Maestro binary and adds it to your PATH.

Pass `--claude` to install for Claude Code instead — `./setup skills --claude` symlinks into `~/.claude/skills/` and `./setup agents --claude` copies `AGENTS.md` to `~/.claude/CLAUDE.md`.

See [Operations](operations/README.md) for details.

Run `./setup -h` to list all available subcommands, or `./setup <command> -h` for subcommand-specific help.

## Maestro (Legacy)

The Maestro planning server is no longer part of the skill suite's planning workflow — it remains in `maestro/` pending deprecation and will be replaced by the plannotator project.
See [Maestro](maestro/README.md) for the full API, data model, and workflow.

## Repository Map

| Path | Description |
|------|-------------|
| `setup` | Installation dispatcher — run subcommands like `./setup maestro` |
| `setup_cmds/` | Modular setup subcommand scripts |
| `AGENTS.md` | Global agent baseline instructions (alwaysApply) |
| `maestro/` | Legacy Go planning server source, templates, static assets (pending deprecation) |
| `maestro/plans/` | Plan files (`.json` format) |
| `skills/` | Agent skill definitions (one subdirectory per skill) |
| `skills/toon/` | TOON format skill — teaches agents to encode/decode TOON |

## Documentation Sections

- [Architecture](architecture/README.md) — system design, data flow, component relationships
- [Maestro](maestro/README.md) — legacy planning server, API, data model, agent status, WebSocket
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
| `/skills/toon/SKILL.md` | Agent instructions for using TOON |

## Next Steps

- Read [Architecture](architecture/README.md) to understand how the pieces fit together
- Read [Skills](skills/README.md) to understand how agent skills work
- Read [TOON Format](toon/README.md) to understand plan file syntax
