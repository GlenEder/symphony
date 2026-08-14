---
name: maestro-author
description: Author a NEW Maestro plan JSON and decompose implementation work into sequential stages. Use when the user asks to create, draft, compose, or decompose a plan. NOT for serving or approving an existing plan (maestro-session), exporting work tickets (maestro-export), interviewing decisions (grilling), implementing a work ticket (plan-implementation-procedure), or routing an ambiguous request (maestro).
compatibility: opencode
---

# Author a Maestro plan

Use this skill for the authoring phase only.
When the plan is ready for human review, hand it explicitly to [`maestro-session`](../maestro-session/SKILL.md).

## Authoring workflow

1. Read the repository instructions and relevant OpenWiki pages before making planning decisions.
2. Resolve open decisions before finalizing the plan.
If decisions remain, use the [`grilling`](../grilling/SKILL.md) skill and ask one question at a time.
3. Build a plan object with `title`, `summary`, `state: "draft"`, and typed `modules`.
4. Validate the plan data model and stage rules below.
5. Write the plan to `$MAESTRO_PLANS_DIR/<plan-id>.json`, or POST it to a running Maestro server when handing off immediately.
6. Confirm the plan can be read back, then invoke `maestro-session` for serving and review.

Completion means the draft is valid, every implementation stage is independently verifiable, and the next skill has an explicit plan ID.

## Plan data model

The required top-level fields are `title`, `summary`, `state`, and `modules`.
Set `state` to `draft` until the human approves the plan.
Each module has a `type`, `heading`, and `items` array.
Each item has required `text` and type-specific fields.

Read [`references/API.md`](../maestro/references/API.md) for the complete HTTP contract.
Read [`references/GLOSSARY.md`](../maestro/references/GLOSSARY.md) for module fields and examples.
Read [`references/server-lifecycle.md`](../maestro/references/server-lifecycle.md) when handing a plan to the HTTP server or checking server requirements.

## Stage decomposition rules

Always create stages, including for trivial work (exactly one stage is the degenerate case).
Use one `criteria`, `steps`, and `risks` module group per stage.
Prefix each of those headings exactly `Stage N: <name>`.
Use contiguous positive stage numbers with no duplicates.
Keep `decision`, `assumptions`, and `notes` modules global and unprefixed.
Order stages by dependency so a stage depends only on earlier stages.
Make each stage a small coherent slice that can be implemented and verified independently.

## Handoff

Pass the exact plan ID and server requirements to `maestro-session`.
The session skill owns server reuse/startup, browser opening, heartbeat polling, responses, approval, and export.
Use [`maestro-export`](../maestro-export/SKILL.md) only through that explicit approval handoff.
