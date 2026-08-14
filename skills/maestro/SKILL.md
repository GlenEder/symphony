---
name: maestro
description: Route Maestro planning requests to the focused authoring or feedback-session skill. Use when the request says Maestro without clearly choosing a phase.
compatibility: opencode
---

# Maestro router

- New plan, plan JSON, or stage decomposition → [`maestro-author`](../maestro-author/SKILL.md).
- Existing plan review, live feedback, polling, approval, or export → [`maestro-session`](../maestro-session/SKILL.md).
- Unresolved decisions during authoring → [`grilling`](../grilling/SKILL.md).
- Approved-plan ticket generation → [`maestro-export`](../maestro-export/SKILL.md).

Choose exactly one phase and hand off explicitly with the plan ID.
The focused skill links the shared API, glossary, and server-lifecycle references under `references/`.
