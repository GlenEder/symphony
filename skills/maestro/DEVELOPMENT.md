# Maestro Development Reference

Disclosed reference for [`maestro`](SKILL.md) — only needed when modifying the Maestro server itself, not when running plans.

## Code Layout

```
maestro/
├── main.go              # Entry point, env config, server setup
├── handler.go           # HTTP route handlers (UI + JSON API)
├── model.go             # Data model types (Plan, Module, Item)
├── store.go             # PlanStore — loads, caches, saves plans to disk
├── watcher.go           # File poller (stat-based, no fsnotify) for live reload
├── watcher_test.go      # Tests for polling, self-write tracking, file detection
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
    └── demo.json        # Example plan
```

## Dependencies

- Go 1.26+
- `github.com/gorilla/websocket` — WebSocket support
- `github.com/sstraus/toon_go/toon` — TOON format parser
