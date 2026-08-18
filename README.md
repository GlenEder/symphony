# Symphony Maestro

Symphony Maestro is a standalone planning application.
It opens a browser, researches a local codebase through an OpenAI-compatible API, guides a grilling conversation, renders a reviewable plan, and exports approved work tickets.
Maki is not required to run Maestro.
The legacy implementation remains in `legacy-maestro/` for compatibility.

## Standalone setup

Requirements: Go 1.26 or newer and an OpenAI-compatible API provider.

```bash
git clone https://github.com/gleneder/symphony.git
cd symphony
go build -o maestro ./cmd/maestro
./maestro serve /path/to/codebase
```

The server binds to `127.0.0.1:8080` by default, so it is not exposed to the network. Set `MAESTRO_ADDRESS` explicitly when remote access is required, and use the configured address and actual port shown in the startup log as the browser URL.
The HTTP server, file watcher, and active conversation workers terminate cleanly on SIGINT or SIGTERM.
The command must be run from the repository checkout, or set `MAESTRO_DIR` to the checkout path so templates and static assets can be found.

## Configuration

Copy `.env.example` to `.env`, or set environment variables directly.
Existing environment variables take precedence over `.env`.

| Variable | Default | Purpose |
| --- | --- | --- |
| `MAESTRO_ADDRESS` | `127.0.0.1` | HTTP bind address; keep loopback unless remote access is explicitly required |
| `PORT` | `8080` | HTTP listening port |
| `CODEBASE_PATH` | `.` | Codebase to research; the `serve` argument sets this |
| `LLM_BASE_URL` | `https://api.openai.com/v1` | OpenAI-compatible API base URL |
| `LLM_API_KEY` | empty | API credential |
| `LLM_MODEL` | `gpt-4o-mini` | Chat completion model |
| `MAESTRO_PLANS_DIR` | `plans` | Persistent plan files |
| `MAESTRO_TRACEABILITY_URL` | empty | Optional ticket traceability link |
| `MAESTRO_DIR` | executable directory or `.` | Templates and static assets |

Any provider implementing `POST /chat/completions` with OpenAI-compatible request and response JSON can be used.
Tests use a local mock provider and never make live API calls.

## Workflow and output

Open the welcome page, submit a prompt, let Maestro research the codebase, answer the grilling questions, review the generated plan, and approve it.
Approval exports Markdown work tickets to `~/.config/symphony/work_tickets/` by default.
Staged plans are written as `{plan-id}-stage-{N}.md`.
Legacy single-stage plans are written as `{plan-id}.md`.
Plans are persisted in `MAESTRO_PLANS_DIR` and can be reopened after restart.

Maestro logs HTTP requests, opens the browser cross-platform on macOS, Windows, and Linux, and shuts down the HTTP server, watcher, and active conversations gracefully on SIGINT or SIGTERM.

## Development

```bash
gofmt -l .
go vet ./...
go test ./...
go test -race ./...
go build -o /tmp/symphony-maestro ./cmd/maestro
```

The automated suite includes unit, HTTP handler, local-provider end-to-end workflow, browser-fallback, and deterministic SIGINT/SIGTERM lifecycle coverage. The integration workflow uses a temporary codebase and local OpenAI-compatible mock server, and verifies ticket export without live API calls.
