# Maestro server lifecycle reference

Use this reference when serving or authoring a plan through the HTTP API.

## Discover or start

Probe ports 8080–8089 and reuse the first server whose `GET /api/plans` response begins with `[`: 

```bash
port=""
for p in $(seq 8080 8089); do
  if curl -s --max-time 0.3 "http://localhost:$p/api/plans" 2>/dev/null | grep -q '^\['; then
    port=$p
    break
  fi
done
```

If no port is found, start `maestro` in the background, wait for `/api/plans`, and record `started_server=true`.
Reuse means `started_server=false`; never stop a server you did not start.

## Session commands

Create or replace a plan with `POST /api/plan/{id}` and confirm it with `GET /api/plan/{id}`.
Open `http://localhost:$port/plan/{id}` for human review.

Each poll iteration is one fast call: POST `/api/agent/{id}/heartbeat`, then GET `/api/plan/{id}`.
Compare the returned JSON with the prior response.

For new human messages, resolve `item_ref` as `moduleIndex:itemIndex`, PATCH the plan if its structure changes, then POST an agent message.
Messages are preserved by plan updates.

On approval, POST `{"status":"offline"}` to `/api/agent/{id}/status`, acknowledge the approval, and export tickets.

## Helper scripts

The bundled `scripts/maestro-discover.sh` probes the standard port range.
`scripts/maestro-heartbeat.sh` is available for integrations that need a background heartbeat.
`scripts/maestro-listen.sh` watches a plan file and fetches its JSON after a change.
The live session procedure remains reasoning-level polling so message handling and approval are explicit.

## Edge cases

| Condition | Action |
|---|---|
| Server already running | Reuse it and do not stop it at session end. |
| Poll failure or interruption | Ask whether to resume, discard, or proceed anyway. |
| Discard after starting a server | Stop only the server started by this session. |
| Approval revoked | Treat `draft` as active again and continue listening. |
| Rapid messages | Process all unseen human messages in one iteration. |
| Direct JSON edit | Fetch the next GET result; use `/api/admin/reload` after bulk edits if needed. |
| Idle review for 30 minutes | Ask whether the user is still reviewing. |
