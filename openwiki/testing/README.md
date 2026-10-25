# Testing

Tests live in two places: `__tests__/` (Node.js — the skill-eval harness and Maestro UI static-asset tests) and `maestro/` (Go tests for the legacy planning server).
There is no CI configuration — run the suites locally.

## Skill Evaluation Suite

`node __tests__/skill_eval_harness.js [output-path]` is a deterministic static proxy for planning-skill evaluation.
It reads scenarios from `__tests__/skill_eval_fixtures.json`, scores declared prompt terms and routing triggers against actual skill frontmatter/body content, and reports selection accuracy and completion checks.

- **Fixtures** (`skill_eval_fixtures.json`) — scenario prompts, expected skills, routing triggers, and completion rules; edit these to change coverage.
- **Baseline** (`skill_eval_baseline.json`) — generated output; regenerate with `node __tests__/skill_eval_harness.js __tests__/skill_eval_baseline.json` — do not hand-edit.

## Maestro Server Tests (Legacy)

The legacy Maestro planning server has both Go tests and Node.js UI tests.

Go tests (run with `cd maestro && go test ./...`) cover:

- `handler_test.go` — POST message endpoints (single/batch, prompts, invalid recommended options)
- `model_test.go` — plan decoding, JSON round-trips, FlatPlan field preservation, store self-write detection and persistence
- `watcher_test.go` — the file poller (new/changed/deleted files, self-write skip, configurable interval)

Node.js tests (run with `node __tests__/<file>`) exercise the DOM logic in `maestro/static/script.js` with a minimal DOM mock — no browser required:

- `maestro-script-utils.test.js` — shared utilities
- `maestro-send-responses.test.js` — send-responses DOM patterns
- `maestro-duplicate-messages.test.js` — discussion-panel duplicate message prevention

## What Remains Untested

- `setup` and `setup_cmds/*.sh` have no automated tests — exercise them with `--dry-run` previews.
- Skills are Markdown instructions; routing is proxied by the eval harness, not a real model router.

## Testing the Legacy Maestro Server Manually

1. **Build** the binary: `cd maestro && go build -o maestro .`
2. **Start the server** with a test plans directory: `MAESTRO_PLANS_DIR=/tmp/test-plans ./maestro`
3. **Place test `.json` plan files** there and verify they load via `GET /api/plans`
4. **Exercise the endpoints** with `curl` or the browser UI
5. **Test WebSocket** with a tool like `websocat` or a browser page

## Adding Tests

- Go tests go in the same package as the code they test (standard `_test.go` pattern), using `httptest.NewServer`/`NewRecorder` and `t.TempDir()` for plan file operations.
- New skill-eval coverage is added via `skill_eval_fixtures.json`; regenerate the baseline afterward.
- Node.js UI tests follow the existing minimal-DOM-mock pattern in `__tests__/`.
