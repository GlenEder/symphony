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

First, decide whether you need to ask the user any questions:

- **If you have unresolved decisions or questions** → start a grilling interview via the `grilling` skill. The grilling session creates the plan and resolves decisions one-by-one. After all questions are answered, populate the plan with modules (decisions, steps, risks, criteria, etc.) and proceed to the feedback loop.
- **If you have no questions** (all decisions are clear, or the user explicitly declined grilling) → create the plan directly:
  1. **Author the plan** — use `maestro-author` for the JSON format, module types, and stage decomposition rules.
  2. **Serve and review it** — hand the draft to `maestro-session`, which owns server reuse, the browser session, heartbeat polling, and approval.

This does NOT apply to: trivial 1-3 line responses, commit messages, or inline code comments. When in doubt, use Maestro.

Direct text output of plan content is the wrong path — the listener should see it rendered in the browser with structured modules, discussion threading, and item-level commenting.

## Symphony Skill Suite — Two-Stage Architecture

The symphony skill suite separates planning from implementation into two distinct stages.
They are **never** chained automatically — the user must explicitly invoke each stage.

### Composer Stage (Planning)

The composer stage is for creating, stress-testing, and approving plans.
It **ends** when the plan is approved and a work ticket is exported.
It does **not** proceed to implementation.

Flow: **research → grilling → maestro-author (plan) → maestro-session (review → approval) → maestro-export (work ticket) → stop**

### Performance Stage (Implementation)

The performance stage is invoked **explicitly** by the user via "pip it" / "implement the plan" / "execute the plan".
It reads the work ticket and enters an autonomous implementer↔reviewer loop.

Flow: **read work ticket → plan-implementation-procedure (implementer↔reviewer loop, max 3) → publish-it (draft PR) → stop**
