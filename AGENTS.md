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

## Maestro — Interactive Planning Server

When you produce any substantive plan (architecture, design, implementation, refactor, or investigation), you MUST serve it via the Maestro web UI.

First, decide whether you need to ask the user any questions — any grilling happens in this agent session BEFORE maestro is invoked:

- **If you have unresolved decisions or questions** → run a grilling interview FIRST via the `grilling` skill in the coding-agent session (interactive question tool, one question at a time), recording each resolved decision as a `decision` module entry.
- **If you have no questions** (all decisions are clear, or the user explicitly declined grilling) → create the plan directly.

Both branches then converge: author the finalized plan document, then invoke maestro solely to render and review it:

1. **Format the finalized Maestro plan document** — use the `maestro` skill's JSON plan format and populate it with all modules (e.g., `steps` for sequential work, `risks` for trade-offs, `notes` for design rationale, `criteria` for acceptance criteria, `questions` for open decisions).
2. **Serve it via the Maestro web UI** — start the server, write the plan JSON file to `$MAESTRO_PLANS_DIR`, and open the browser URL. If `$MAESTRO_PLANS_DIR` is unset, the server defaults to `plans` relative to the current working directory; the `setup` script exports `$MAESTRO_PLANS_DIR`, so ensure it is set before writing the file.
3. **Enter the feedback loop ONLY for rendering and reviewing** — collect item-level comments and secure approval; interviewing does not happen here. See the `maestro` skill for the heartbeat cadence, listen endpoint, and exit conditions.

This does NOT apply to: trivial 1-3 line responses, commit messages, or inline code comments. When in doubt, use Maestro.

Direct text output of plan content is the wrong path — the listener should see it rendered in the browser with structured modules, discussion threading, and item-level commenting.

## Symphony Skill Suite — Two-Stage Architecture

The symphony skill suite separates planning from implementation into two distinct stages.
They are **never** chained automatically — the user must explicitly invoke each stage.

### Composer Stage (Planning)

The composer stage is for creating, stress-testing, and approving plans.
It **ends** when the plan is approved and a work ticket is exported.
It does **not** proceed to implementation.

Flow: **research → in-session grilling interview (via the `grilling` skill, before the plan exists) → plan population → final review of the authored plan served via maestro → approval → export work ticket → stop**

Key skills:
- `research` — gather facts and context for the plan
- `maestro` — author the finalized structured plan from the resolved decisions, serve it for review, run the feedback loop, handle approval, and trigger the export
- `grilling` — interviews the user relentlessly in the coding-agent session BEFORE a plan exists (one question at a time), resolving decisions one-by-one and handing them to plan authoring
- `maestro-export` — convert the approved maestro plan JSON to a standardized Markdown work ticket

Output: one Markdown work ticket per stage at `~/.config/symphony/work_tickets/{plan-id}-stage-{N}.md`.
For legacy single-stage plans, output the fallback ticket at `~/.config/symphony/work_tickets/{plan-id}.md`.

### Performance Stage (Implementation)

The performance stage is invoked **explicitly** by the user via "pip it" / "implement the plan" / "execute the plan".
It reads the work ticket and enters an autonomous implementer↔reviewer loop.

Flow: **read work ticket → implementer↔reviewer loop (max 3 iterations) → publish-it draft PR → stop**

Key skills:
- `plan-implementation-procedure` — orchestrates the implementer↔reviewer subagent loop against the work ticket
- `publish-it` — creates a branch, commits, pushes, and opens a **draft** PR

Input: the lowest-numbered pending stage ticket from `~/.config/symphony/work_tickets/{plan-id}-stage-{N}.md`.
For legacy single-stage plans, use `~/.config/symphony/work_tickets/{plan-id}.md` (fallback: maestro plan JSON).

### Work Ticket Storage

Work tickets are stored one per stage at `~/.config/symphony/work_tickets/{plan-id}-stage-{N}.md`.
The legacy single-stage fallback is `~/.config/symphony/work_tickets/{plan-id}.md`.
The config directory is user-global (not per-project), so work tickets survive across projects and sessions.
Each ticket is a clean, self-contained Markdown file with acceptance criteria, implementation steps, decisions, risks, and assumptions — suitable for copy-paste into Linear, Jira, or GitHub Issues.
