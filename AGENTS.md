---
description: Global baseline instructions for all agent sessions
alwaysApply: true
---

# Agent Instructions
These instructions are for all scenarios across all agents.

## OpenWiki

Repositories may have documentation located in the /openwiki directory.

ALWAYS look for this and if found start here:
- [OpenWiki quickstart](openwiki/quickstart.md)

OpenWiki includes repository overview, architecture notes, workflows, domain concepts, operations, integrations, testing guidance, and source maps.

When working in a repository with openwiki, read the OpenWiki quickstart first, then follow its links to the relevant architecture, workflow, domain, operation, and testing notes.

## General Guidelines
- When writing commit messages, never auto-add your agent name as co-author.
- Never manually modify changelog files, or any files that are auto-generated.
- When writing Markdown files, ensure every sentence is on its own line.
  Preserve normal Markdown structure, and avoid putting multiple sentences on one line.
- When planning work, do **not** implement anything.
  Only investigate, explore, and produce the plan.
  Wait for explicit approval or a prompt like "implement" to begin coding.
  During execution, perform the planned work, including any git branch, commit, push, and PR steps the plan calls for.
- When doing bug fixes, always start with reproducing the bug in an E2E setting as closely aligned with how an end user would experience it.
  This helps ensure you find the real problem so your fix solves it.
- When end-to-end testing, be picky about the UI you see and be obsessed with pixel perfection.
  If something clearly looks off, even if it is not directly related to what you are doing, try to fix it along the way.
- Apply the same high standard to engineering excellence: lint failures, test failures, and test flakiness.
  If you see one, even if it is not caused by what you are working on right now, still fix it.
- For planning discussions, stay in Agent mode by default.

## Symphony Skill Suite — Two-Stage Architecture

The symphony skill suite separates planning from implementation into two distinct stages.
They are **never** chained automatically — the user must explicitly invoke each stage.

### Composer Stage (Planning)

The composer stage is for creating, stress-testing, and approving plans.
It **ends** when the plan is approved and its work tickets are written.
It does **not** proceed to implementation.

Flow: **research → in-session grilling interview (via the `grilling` skill, before the plan exists) → plan authoring → final review of the authored plan with the user → approval → write work tickets → stop**

Key skills:
- `research` — gather facts and context for the plan
- `grilling` — interviews the user relentlessly in the coding-agent session BEFORE a plan exists (one question at a time), resolving decisions one-by-one and handing them to plan authoring

Plan authoring happens in the agent session, not in a dedicated tool.
When the plan decomposes into implementation stages, author one work ticket per stage directly in the Work Ticket Storage format; single-stage plans get a single ticket.
Present the authored plan to the user for review and secure explicit approval before writing the tickets — never write tickets for an unapproved plan.

Structured plan rendering and item-level review — previously provided by the Maestro planning server — will be reintegrated via the plannotator project.
Until that integration lands, plan review happens in the agent session.

Output: one Markdown work ticket per stage at `~/.config/symphony/work_tickets/{plan-id}-stage-{N}.md`.
For single-stage plans, output the ticket at `~/.config/symphony/work_tickets/{plan-id}.md`.

### Performance Stage (Implementation)

The performance stage is invoked **explicitly** by the user via "pip it" / "implement the plan" / "execute the plan".
It works through **all not-done stage tickets** of the plan, lowest pending-or-in-progress stage first, running each stage through the implementer↔reviewer loop and committing it on the plan branch.
A qualifier like `pip stage N` runs only that single stage.

Flow: **create plan branch → per not-done stage (lowest first): implementer↔reviewer loop (max 3 iterations) → stage commit → next stage → publish-it draft PR for the whole plan → stop**

A run halts when a stage is flagged: max 3 iterations with unresolved blocker/major findings, or 2 consecutive no-progress iterations.
On a halt it publishes what's done as the plan's single draft PR, persists the findings into the halted ticket, and asks the user how to proceed.

Key skills:
- `plan-implementation-procedure` — orchestrates the multi-stage implementer↔reviewer subagent loop over the plan's stage work tickets
- `publish-it` — creates a branch, commits, pushes, and opens a **draft** PR

Input: the plan's not-done stage tickets at `~/.config/symphony/work_tickets/{plan-id}-stage-{N}.md`, lowest pending-or-in-progress first.
For single-stage plans, use `~/.config/symphony/work_tickets/{plan-id}.md`.

### Work Ticket Storage

Work tickets are stored one per stage at `~/.config/symphony/work_tickets/{plan-id}-stage-{N}.md`.
The single-stage fallback is `~/.config/symphony/work_tickets/{plan-id}.md`.
The config directory is user-global (not per-project), so work tickets survive across projects and sessions.

Each ticket is a clean, self-contained Markdown file with acceptance criteria, implementation steps, decisions, risks, and assumptions — suitable for copy-paste into Linear, Jira, or GitHub Issues.

A staged ticket contains, in order:

1. A header block — `# {plan title} — Stage N: {stage name}`, `**Status**: pending`, and `**Stage**: N of M of the approved {plan-id} plan`.
2. The plan summary, when present.
3. Global **Key Decisions** — one subheading per decision with the chosen alternative, the rationale, and the alternatives considered — and **Assumptions**.
4. The stage's **Acceptance Criteria** as a checkbox list, **Implementation Steps** as a numbered list, and **Risks** as severity-tagged bullets.

A single-stage ticket omits the stage header block and carries all sections under the plain plan title.
Omit any section the plan has no content for.
