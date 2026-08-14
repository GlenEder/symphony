---
name: maestro-session
description: Serve an existing Maestro plan and run its live feedback session through approval. Use when a plan already exists and the user wants review, comments, polling, or approval. NOT for authoring or decomposing a new plan (maestro-author), exporting work tickets (maestro-export), interviewing decisions (grilling), implementing a work ticket (plan-implementation-procedure), or routing an ambiguous request (maestro).
compatibility: opencode
---

# Run a Maestro feedback session

Use this skill after [`maestro-author`](../maestro-author/SKILL.md) has produced a draft plan.
This skill owns the live session and stops after the approved plan is handed to [`maestro-export`](../maestro-export/SKILL.md).

## Session workflow

1. Discover a running Maestro server on ports 8080–8089 by checking `GET /api/plans`.
Reuse it when present; otherwise start one and record whether you started it.
2. POST the draft plan to `/api/plan/{plan-id}`, confirm it with GET, and open `/plan/{plan-id}` in the browser.
Tell the user where to review it.
3. Initialize the set of already-seen message IDs.
4. Poll at the reasoning level.
Each iteration sends one heartbeat to `/api/agent/{plan-id}/heartbeat`, fetches `/api/plan/{plan-id}`, and compares messages and state with the previous response.
5. For every new human message, resolve `item_ref` against the plan, update the plan with PATCH when needed, then post one agent response.
Track the message ID immediately.
6. When state is `approved`, set agent status to `offline`, post final acknowledgment, and invoke `maestro-export` to write the stage work tickets.
Report the ticket directory and stop.

Completion means every new human message has a response, the approved plan is exported, and no implementation work has started.

Read [`references/server-lifecycle.md`](../maestro/references/server-lifecycle.md) for the probe, startup, heartbeat, polling, and interruption commands.
Read [`references/API.md`](../maestro/references/API.md) for endpoint details and [`references/GLOSSARY.md`](../maestro/references/GLOSSARY.md) when resolving module items.

## Interruption handling

If polling fails or is interrupted, ask whether to resume, discard, or proceed anyway.
Resume by restarting the loop.
Discard by stopping only a server started by this session; leave reused servers running.
For an idle review, ask whether the user is still reviewing after 30 minutes.

## Explicit handoffs

- New plan or stage decomposition → [`maestro-author`](../maestro-author/SKILL.md).
- Stress-testing unresolved decisions → [`grilling`](../grilling/SKILL.md).
- Approved plan export → [`maestro-export`](../maestro-export/SKILL.md).
