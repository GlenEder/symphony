# Architecture

Symphony has two independent layers that work together: the **skill suite** (agent instructions) and **Maestro** (the planning server).

## High-Level Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Coding Agent                          │
│  (Maki / Claude Code / Cline / Aider)                   │
│                                                         │
│  ┌──────────────────────────────────────────────┐      │
│  │  Skills (~/.config/maki/skills/)             │      │
│  │  ┌─────────┐ ┌──────────┐ ┌─────────────┐  │      │
│  │  │ maestro │ │ toon     │ │ gh, pm, ... │  │      │
│  │  └────┬────┘ └──────────┘ └─────────────┘  │      │
│  └───────┼──────────────────────────────────────┘      │
│          │                                             │
│          │ HTTP + WebSocket                            │
└──────────┼─────────────────────────────────────────────┘
           │
┌──────────▼─────────────────────────────────────────────┐
│                    Maestro Server                        │
│  (Go web server, port 8080)                            │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌────────────────────┐  │
│  │ Handler  │  │ PlanStore│  │ WebSocket Hub      │  │
│  │ (routes) │  │ (load/   │  │ (live broadcasts)  │  │
│  │          │  │  persist)│  │                    │  │
│  └────┬─────┘  └────┬─────┘  └────────┬───────────┘  │
│       │              │                  │              │
│  ┌────▼──────────────▼──────────────────▼───────────┐  │
│  │              TOON (.toon) files                    │  │
│  │              (maestro/plans/ or MAESTRO_PLANS_DIR)│  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
           │
           │ Browser (human reviewer)
```

## Layer 1: Skill Suite

The skill suite is a collection of **SKILL.md** files organized under `skills/`. Each skill teaches an AI agent a specific workflow. The `setup.sh` script symlinks these into the agent's config directory (`~/.config/maki/skills/`).

**Key design choices:**
- Skills are plain Markdown + shell scripts — they are agent-system-agnostic.
- They are installed via symlinks so changes in the repo are immediately available.
- The `AGENTS.md` file at the repo root provides global baseline instructions (alwaysApply: true).

See [Skills](skills/README.md) for details.

## Layer 2: Maestro Planning Server

Maestro is a Go web server that:
1. Loads `.toon` files from a configurable directory into memory
2. Serves them via HTTP (HTML UI + JSON API)
3. Broadcasts live updates via WebSocket when files change
4. Tracks AI agent status (listening/thinking/offline) via heartbeat polling

**Source files and their roles:**

| File | Role |
|------|------|
| `main.go` | Server entrypoint — wires up PlanStore, AgentState, Hub, watcher, templates, routes |
| `handler.go` | HTTP route handlers — plan listing, detail pages, JSON API, agent status endpoints |
| `model.go` | Data types — Plan, Module, Item, Message, FlatPlan, and conversion helpers |
| `store.go` | PlanStore — loads `.toon` files from disk, persists changes, manages conversation threads |
| `ws.go` | AgentState (per-plan agent status + GC), WebSocket Hub (subscribe/broadcast) |
| `watcher.go` | fsnotify-based file watcher for live reload of plan files |

**Dependencies:** `gorilla/websocket`, `fsnotify`, and the local TOON library at `maestro/lib/toon/`.

## Data Flow

### Plan Loading

```
Disk (.toon file) ──► TOON decoder ──► JSON marshal ──► Plan struct ──► map[string]*Plan
                           │                    │
                    (toon.Unmarshal)      (json.Unmarshal)
```

### Serving a Plan

```
HTTP GET /plan/{id}
  └─► store.Get(id) ──► Plan struct ──► PlanPageData ──► HTML template render
HTTP GET /api/plan/{id}
  └─► store.Get(id) ──► toFlatPlan() ──► JSON response
```

### Live Updates

```
File change ──► fsnotify ──► store.loadFile() ──► onChange callback
                                                     │
                                              hub.Broadcast(id, json)
                                                     │
                                           WebSocket clients receive
                                              full flat plan JSON
```

### Agent Status

```
Agent heartbeat (POST /api/agent/{id}/heartbeat, every 15s)
  └─► agentState.Heartbeat(id) ──► Status=Listening, LastHeartbeat=now
                                              │
                                     GC goroutine (5s interval)
                                              │
                              Expired (10 min timeout) → Status=Offline

Human sends message ──► agentState.SetThinking(id)
Agent sends message  ──► agentState.SetListening(id)

Browser polls ──► GET /api/agent/{id}/status ──► returns JSON {"status":"..."}
```

## Layer 3: TOON Format

TOON (Token-Oriented Object Notation) is a compact JSON alternative designed for LLM token efficiency. It uses indentation-based structure with CSV-style tabular arrays. Maestro uses the local Go library at `maestro/lib/toon/` to encode and decode plan files.

See [TOON Format](toon/README.md) for the full reference.

## Key Architectural Decisions

1. **TOON as the on-disk format** — Compact, human-readable, token-efficient. The Go library at `maestro/lib/toon/` provides round-trip encoding (TOON ↔ JSON ↔ Go struct).
2. **File-watcher-based live reload** — Plans update in real-time without server restart (fsnotify). The watcher triggers WebSocket broadcasts to connected clients.
3. **In-memory + file persistence** — Plans are loaded into memory on startup and persisted back to disk on each mutation. Messages are stored inline in the `.toon` file as part of the JSON round-trip.
4. **Per-plan agent state** — Each plan has its own agent status (listening/thinking/offline). The status is auto-transitioned based on message roles and heartbeat activity, not directly controlled by the agent.
5. **Positional references** — Messages reference plan items using `moduleIndex:itemIndex` strings instead of UUIDs for simplicity.
