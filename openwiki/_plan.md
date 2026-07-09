# Documentation Plan

## Repository: Symphony — Coding Agent Skill Suite

### Purpose
A collection of skills, prompts, and a Go web server (Maestro) for AI coding agents.
Skills encode reusable workflows; Maestro serves structured planning documents with
a web UI, JSON API, and WebSocket live reload.

### Major Domains
1. **Maestro Server** (`maestro/`) — Go web server for structured plans
2. **TOON Format** (`maestro/lib/toon/`) — Token-efficient data format
3. **Skills** (`skills/`) — Agent instruction files for workflows
4. **Setup/Operations** (`setup.sh`, scripts) — Installation and feedback loop

### Pages to write
1. **openwiki/quickstart.md** — Entrypoint. Overview, quick start, links to all sections.
2. **openwiki/architecture.md** — Maestro server architecture, data model, routes, WebSocket protocol, agent state machine.
3. **openwiki/toon-format.md** — TOON syntax, purpose, how Maestro consumes it.
4. **openwiki/skills.md** — All skills summaries, relationships, the agent feedback loop, the Plan Display requirement from AGENTS.md.
5. **openwiki/operations.md** — Setup, configuration, environment variables, heartbeat/listen scripts, running the server.

### Source Evidence per Page

#### quickstart.md
- /README.md — project description, prerequisites, setup
- /maestro/go.mod — Go deps
- /skills/maestro/SKILL.md — quick start section

#### architecture.md
- /maestro/main.go — entrypoint
- /maestro/handler.go — routes
- /maestro/ws.go — WebSocket hub, agent state
- /maestro/store.go — PlanStore, persistence
- /maestro/model.go — data types
- /maestro/watcher.go — file watcher

#### toon-format.md
- /maestro/lib/toon/toon.go — package doc
- /maestro/lib/toon/ (various files) — implementation
- /skills/toon/SKILL.md — human-facing docs
- /maestro/store.go — decodePlan() usage
- /maestro/plans/test-all-modules.toon — example

#### skills.md
- /skills/*/SKILL.md (all 7 skills)
- /skills/maestro/scripts/*.sh — heartbeat + listen
- /AGENTS.md — Plan Display section

#### operations.md
- /setup.sh — installation
- /skills/maestro/SKILL.md — config, scripts
- /skills/maestro/scripts/maestro-heartbeat.sh
- /skills/maestro/scripts/maestro-listen.sh
