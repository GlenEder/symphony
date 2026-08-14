# Testing

**There are currently no automated tests in this repository.**

No `*_test.go` files exist in the Maestro Go server or anywhere else in the project. There are no test scripts, CI configurations, or test harnesses.

## This Means

- The Maestro HTTP handlers, PlanStore, WebSocket hub, and AgentState are untested.
- The shell scripts (`setup`, `maestro-heartbeat.sh`, `maestro-listen.sh`) are not tested.
- Changes rely entirely on manual verification through the browser UI and API calls.

## What Should Be Tested

### Maestro Server

**Unit tests** for core logic:
- `maestro/store.go` — `decodePlan`, `AddMessage`, `DeleteMessage`, `SetState`, `List` (sorting), `persistPlan`
- `maestro/model.go` — `toFlatPlan` conversion
- `maestro/ws.go` — `AgentState` transitions (heartbeat, GC, SetThinking/SetListening/SetOffline)

**Integration tests** for HTTP handlers:
- `maestro/handler.go` — all API endpoints, WebSocket upgrade, template rendering
- Test with a temp directory and sample `.json` files

**Test HTTP handler patterns:** Use `net/http/httptest` with a temporary `PlanStore` and `Hub`.

### Shell Scripts

Test `setup` with a temp directory. Test `maestro-heartbeat.sh` and `maestro-listen.sh` with a running test instance.

## Testing Approach

Until tests are added, the recommended approach for validating changes:

1. **Build** the Maestro binary: `cd maestro && go build -o maestro .`
2. **Start the server** with a test plans directory: `MAESTRO_PLANS_DIR=/tmp/test-plans ./maestro`
3. **Place test `.json` files** and verify they load via the API
4. **Exercise endpoints** with `curl` or the browser UI
5. **Test WebSocket** with a tool like `websocat` or a browser page
6. **Test scripts** by running them against the test server

## Adding Tests

When adding tests, follow these conventions:

- Tests should go in the same package as the code they test (standard Go `_test.go` pattern)
- HTTP handler tests should use `httptest.NewServer` or `httptest.NewRecorder`
- Use a temp directory (`t.TempDir()`) for plan file operations
- Test the TOON round-trip for all edge cases (empty arrays, special characters, deep nesting)
