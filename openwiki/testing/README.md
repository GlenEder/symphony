# Testing

The repository has automated tests across three layers: Go unit tests for the Maestro server, Node tests for the browser-side message handling, and a shell test for the Maestro helper scripts.
There is also a static-analysis eval harness for the planning skills.
There is no CI configuration; run the suites manually from the commands below.

## How to Run

### Go tests (Maestro server)

From the `maestro/` directory:

```bash
go test ./...
```

Tests live alongside the code they cover:

| File | Coverage |
|------|----------|
| `maestro/handler_test.go` | HTTP message endpoints — single, batch, prompt, and invalid-`recommended` handling |
| `maestro/model_test.go` | `decodePlan`, JSON round-trip, `AddMessage`, and `FlatPlan` field preservation (decision modules, prompt messages) |
| `maestro/watcher_test.go` | The file poller — new-file detection, change detection, deletion, self-write skipping, configurable interval, mtime recording, empty and non-existent directories |

Twenty `func Test...` cases run in total.
Use `go test ./... -v` to see each case.

### Node tests (browser-side message handling)

Vanilla Node, no test framework or `package.json` — each file is self-contained with `assert`/`assertEqual` helpers:

```bash
node __tests__/maestro-duplicate-messages.test.js
node __tests__/maestro-script-utils.test.js
node __tests__/maestro-send-responses.test.js
```

| File | Coverage |
|------|----------|
| `maestro-duplicate-messages.test.js` | Thread construction and duplicate-message suppression in the rendered DOM |
| `maestro-script-utils.test.js` | Shared script utilities |
| `maestro-send-responses.test.js` | The grilling response UI — send-bar state, answered counts, button enable/disable, and `escapeHtml` |

### Shell test (Maestro helper scripts)

```bash
bash __tests__/maestro-scripts.test.sh
```

Validates `--plan-id` flag parsing and help output for `skills/maestro/scripts/maestro-heartbeat.sh` and `skills/maestro/scripts/maestro-listen.sh`.

### Skill eval harness (static-analysis proxy)

```bash
node __tests__/skill_eval_harness.js
```

A deterministic static proxy that scores planning-skill selection accuracy, workflow completion, and loaded-body token cost against `__tests__/skill_eval_fixtures.json`.
It validates trigger/body integrity and token cost, not semantic routing quality (no headless agent runtime).
Artifacts: `skill_eval_baseline.json` (pre-change baseline) and `skill_eval_comparison.md` (before/after summary).

## Testing Approach for Manual Verification

For end-to-end validation the automated suites do not cover:

1. **Build** the Maestro binary: `cd maestro && go build -o maestro .`
2. **Start the server** with a test plans directory: `MAESTRO_PLANS_DIR=/tmp/test-plans ./maestro`
3. **Place test `.json` files** and verify they load via the API
4. **Exercise endpoints** with `curl` or the browser UI
5. **Test WebSocket** with a tool like `websocat` or a browser page
6. **Test scripts** by running them against the test server

## Adding Tests

When adding tests, follow these conventions:

- Go tests go in the same package as the code they test (standard `_test.go` pattern), as `handler_test.go`, `model_test.go`, and `watcher_test.go` already do.
- HTTP handler tests use `net/http/httptest` with a temporary `PlanStore` and `Hub`.
- Use a temp directory (`t.TempDir()`) for plan file operations.
- Node tests stay framework-free: add a new `__tests__/*.test.js` file with `assert`/`assertEqual` helpers, runnable with `node`.
