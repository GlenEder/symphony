---
name: maestro
description: Route an ambiguous Maestro planning request to exactly one focused skill and hand off with the plan ID. Use only when the user says "Maestro" without clearly choosing a phase. NOT for authoring or decomposing a plan (maestro-author), serving or approving a plan (maestro-session), exporting work tickets (maestro-export), interviewing decisions (grilling), or implementing a work ticket (plan-implementation-procedure).
compatibility: opencode
---

# Maestro router

- New plan, plan JSON, or stage decomposition → [`maestro-author`](../maestro-author/SKILL.md).
- Existing plan review, live feedback, polling, approval, or export → [`maestro-session`](../maestro-session/SKILL.md).
- Unresolved decisions during authoring → [`grilling`](../grilling/SKILL.md).
- Approved-plan ticket generation → [`maestro-export`](../maestro-export/SKILL.md).

Choose exactly one phase and hand off explicitly with the plan ID.
The focused skill links the shared API, glossary, and server-lifecycle references under `references/`.
