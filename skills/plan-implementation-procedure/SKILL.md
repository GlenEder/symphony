---
name: plan-implementation-procedure
description: Orchestrate a plan-level loop over an approved plan's stage work tickets — work through ALL not-done stages in one invocation (a qualifier like 'pip stage 2' runs only that stage), running each stage through an implementer↔reviewer subagent loop (max 3 iterations), committing cleanly-completed stages on a plan branch, and invoking publish-it once to open a single draft PR. Triggered by 'pip it', 'implement the plan', or 'execute the plan'.
compatibility: opencode
---

## Purpose

When the user says "pip it" or "implement the plan" against an approved plan, orchestrate a **plan-level loop** over its stage work tickets.

A bare `pip it` processes **all not-done stages** in one invocation, stopping only at a halt condition or when no not-done stage remains.
A qualifier like `pip stage 2` runs only that single stage, then stops — publishing only if the user explicitly asked to publish.

For each stage, run an autonomous loop of two subagents:

- An **implementer** subagent (medium-tier) edits the working tree to satisfy the stage and writes tests.
- A **reviewer** subagent (strong-tier) reviews the implementation + tests against the stage and best practices, emitting structured findings and a verdict.

The per-stage loop runs up to **3 iterations** (implementer → reviewer per iteration) and stops early when the reviewer is satisfied (no blocker/major findings).

The orchestrator owns all git:

- It creates the plan branch once at plan start, so stage commits never land on the default branch.
- It commits each cleanly-completed stage on the plan branch.
- It invokes **publish-it** exactly once — at plan completion or at a halt — to open a single draft PR for the whole plan.

Subagents never commit and never branch.

## Example Triggers
- pip it
- pip this
- pip stage 2
- implement the plan
- execute the plan

## Workflow

### 0. Load the next not-done stage ticket (re-globbed every cycle)

Locate the target plan's tickets and select the stage to run:

- **Primary staged path**: glob `~/.config/symphony/work_tickets/{plan-id}-stage-*.md` and select the **lowest-numbered stage whose Status is `pending` OR `in-progress`**.
- An `in-progress` ticket is an abandoned run — resume it, do not skip past it.
- Re-glob on **every cycle** of the plan-level loop (step 2); the lowest not-done stage changes as stages complete.
- **Single-stage qualifier**: when the trigger names a stage (`pip stage 2`), run only that stage — it must be `pending` or `in-progress`; if it is already `done`, report and exit.
- **Legacy fallback**: use `~/.config/symphony/work_tickets/{plan-id}.md` for single-stage plans.
- **Plan JSON fallback**: use `$MAESTRO_PLANS_DIR/{plan-id}.json` (parse its `modules` — especially `criteria` and `steps`).
- For a standalone `pip it` (no composer stage): use the plan content already in context, or the plan file the user points to.

Ticket status transitions:

- `pending` → `in-progress` at stage entry, before iteration 1.
- `in-progress` → `done` on clean loop completion, before the stage commit (step 2g).
- Unchanged (`in-progress`) at a halt (step 3).

Extract the acceptance criteria and implementation steps for the selected stage; these are what the implementer works to and the reviewer checks against.
If the ticket carries an `## Outstanding Review Findings` section from a halted run, those findings become iteration 1's `latest_findings`.

Done when: the selected stage ticket's criteria + steps are in hand to pass to subagents, and its status is `in-progress`.

### 1. Create the plan branch once, up front

The orchestrator — never a subagent — creates the feature branch **once at plan start**, before the first stage runs.
Stage commits must never land on the default branch.

```bash
CURRENT_BRANCH=$(git branch --show-current)
DEFAULT_BRANCH=$(git rev-parse --abbrev-ref origin/HEAD 2>/dev/null | sed 's|origin/||')
DEFAULT_BRANCH=${DEFAULT_BRANCH:-main}
```

- **Already on this plan's branch** (a prior run created it and left it checked out): stay on it.
- **Plan branch exists but isn't checked out** (e.g. a fresh session on the default branch): discover it before creating anything — `git branch --list` for a branch matching this plan (its short description or Linear issue key), or the head branch of this plan's open PR — then `git checkout <branch>`.
- Never `checkout -b` a second branch for the same plan; that forks history.
- Otherwise, name the branch with publish-it's conventions: if a Linear issue key is in context, `<linear-issue-key>/<short-description>`; otherwise `<type>/<scope>-<short-description>` per [Conventional Commits](../_shared/conventional-commits.md).

```bash
git checkout -b <branch-name>
```

Done when: the orchestrator is on the plan branch — created now or resumed — and no subagent has touched git.

### 2. The plan-level loop — work through all not-done stages

Outer loop: while a not-done stage exists and no halt has fired, run one stage through the inner implementer↔reviewer loop (2a–2f) and close it out (2g).
A stage qualifier overrides the loop: it scopes this invocation to the named stage, and the run ends when that stage closes out (2g) — never continue into later stages.

Per stage, initialize `iteration=0`, `latest_findings=""` (empty on iteration 1 — unless resuming from `## Outstanding Review Findings`), `no_progress_streak=0`.

Loop the inner loop while `iteration < 3`:

```bash
iteration=$((iteration + 1))
```

#### 2a. Dispatch the implementer subagent

Spawn a **fresh** implementer subagent (foreground/blocking).
In Cursor, the Task tool with `subagent_type: generalPurpose` and a **medium-tier** model; in Maki, the `task` tool (general) at medium tier.

Pass it the **implementer dispatch contract** (below): the stage plan, the latest reviewer findings (empty on iteration 1, unless resuming from `## Outstanding Review Findings`), and the constraints (read working tree first, do not commit or branch, run tests + linters before reporting done).

Receive the implementer's summary: what it changed, whether tests/lint pass, and whether it made progress.

Done when: the implementer subagent has returned a change summary.

#### 2b. Post the implementer summary to the user

Post a short summary of what the implementer changed (files, behavior, test/lint status) and the iteration count.

Done when: the user has been informed of the implementer pass.

#### 2c. Handle no-progress

If the implementer made **no changes** (or errored out):

- Increment `no_progress_streak`.
- If `no_progress_streak >= 2`: **halt** — go to step 3 (halt handling).
- Else: still run the reviewer on the current state (it may find the state is actually fine, or surface what's blocking progress).

If the implementer made progress, reset `no_progress_streak=0`.

Done when: the no-progress streak is updated and the halt decision is made.

#### 2d. Dispatch the reviewer subagent

Spawn a **fresh** reviewer subagent (foreground/blocking).
In Cursor, the Task tool with `subagent_type: generalPurpose` and a **strong-tier** model; in Maki, the `task` tool (general) at strong tier.

Pass it the **reviewer dispatch contract** (below): the stage plan, the rubric (plan adherence, tests, best practices), instruction to run tests + linters, and the requirement to emit a structured TOON block (findings + `satisfied` verdict) via the `toon-output` skill.

Receive the reviewer's TOON block.

Done when: the reviewer subagent has returned a TOON block.

#### 2e. Post the reviewer summary to the user

Post a short summary: the findings (by severity), the `satisfied` verdict, and the iteration count.

Done when: the user has been informed of the reviewer pass.

#### 2f. Parse the verdict and decide

Parse the reviewer's TOON block (use the `toon` skill's parsing rules):

- If `satisfied == true` **or** there are no `blocker`/`major` findings → **stage loop done** (go to step 2g).
- Else: extract the `blocker`/`major` findings → set `latest_findings` to them, and loop again (2a).
- `minor`/`nit` findings are reported but do not extend the loop.

If the inner loop exits at **max iterations (3)**, classify the final verdict:

- No unresolved `blocker`/`major` findings → clean completion (go to step 2g).
- Unresolved `blocker`/`major` findings remain → **halt** (go to step 3).
- No parseable final verdict → **halt** (go to step 3) — an unparseable verdict is never clean completion.

If no parseable verdict is found and iterations remain, treat it as a failed reviewer pass: surface the raw reviewer output to the user, and loop again (counts as one of the 3).

Done when: the verdict is parsed and the continue/clean/halt decision is made.

#### 2g. Stage closeout — mark done, commit, summarize

When a stage's loop completes cleanly:

- Mark the ticket `Status: done` (before the commit).
- Commit the stage's changes on the plan branch as a stage-scoped conventional commit per [Conventional Commits](../_shared/conventional-commits.md), e.g. `feat(<plan-scope>): stage N — <short description>`, where `<plan-scope>` derives from the plan's actual scope — never hardcoded `pip`.
- Post a **stage-completion summary** to the user: stage number, what was implemented, iteration count, commit hash.
- Bare `pip it`: re-glob (step 0) for the lowest-numbered not-done stage and loop; if no not-done stage remains, the plan is complete → step 4.
- **Qualified run (`pip stage N`)**: stop after this stage — do not re-glob into later stages.
  - Publish only if the user explicitly asked this run to publish (invoke publish-it as in step 4); otherwise report the stage completion and stop, leaving the working tree on the plan branch for the next run.

Done when: the stage is committed on the plan branch, the summary is posted, and the next stage, plan completion, or qualified-run stop is determined.

### 3. Halt handling

Two halt conditions stop the whole plan-level loop:

1. A stage's reviewer loop hits **max 3 iterations** with unresolved `blocker`/`major` findings **or an unparseable final verdict** (step 2f).
2. **2 consecutive no-progress implementer iterations** (step 2c).

On a halt:

- Do **not** mark the flagged stage `done` — leave its status `in-progress`.
- Persist the **halt reason** (max-iter with unresolved findings vs two consecutive no-progress iterations) and the stage's outstanding `blocker`/`major` findings (if any) into the ticket file under a clearly marked `## Outstanding Review Findings` section, so the next run's iteration 1 isn't blind to why it stopped.
- Invoke **publish-it once** to publish the plan branch itself as the single plan PR.
  - publish-it must push the plan branch and open the draft PR if none exists yet — including a first halt, where the branch holds only prior stage commits plus the flagged stage's working-tree changes — or append to the plan's existing open draft PR; never stack a second PR.
- On a qualified run, publish only if the user explicitly asked to publish — otherwise skip this step and go straight to asking the user.
- Ask the user how to proceed.
- Do **not** continue to later stages after a halt.

Done when: the halt reason and findings are persisted in the flagged ticket, publish-it has run (or been skipped on a qualified run that isn't publishing), and the user has been asked how to proceed.

### 4. Publish once and stop

When no not-done stage remains, the plan is complete — the bare-`pip it` completion path; a qualified run already stopped at step 2g.

- Invoke **publish-it once** on the plan branch — one draft PR for the whole plan, never per stage.
- publish-it must publish the plan branch itself as that single plan PR: push it and open the draft PR if none exists, or append to the plan's existing open draft PR (e.g. from a prior halt) — never stack a second PR.
- If invoked when every stage is already `done`, publish only when no draft PR exists yet; otherwise just report it.
- Report to the user: the stages completed, the PR URL, and any outstanding findings.
- Stop at the PR.
- Do not write back to Maestro or Linear.
- The Maestro approval-time closeout (agent offline, heartbeat stop, final ack) already happened before the loop started.

Done when: the single draft PR URL is reported with the stage list, and no further writeback is attempted.

## Implementer dispatch contract

Give each fresh implementer subagent:

- The **approved stage plan** (criteria + steps) for the current stage, verbatim or summarized.
- The **latest reviewer findings** (`blocker`/`major` only): the previous iteration's TOON block, or — on iteration 1 of a resumed stage — the ticket's `## Outstanding Review Findings` section.
- **Read the working tree first**: prior iterations' and prior stages' edits live in the working tree — read the current state of the files you will touch before editing. Do not assume a fresh checkout.
- **Do not commit, do not create or switch branches**: edit the working tree only. The orchestrator owns all git — the plan branch, the stage-boundary commits, and the final publish.
- **Write tests** for the work, following the repo's existing test patterns and directory structure.
- **Run tests + linters before reporting done**: report pass/fail. If the repo has no test infrastructure, skip tests and note that.
- **Report back**: a concise summary of what changed, test/lint status, and whether progress was made.

## Reviewer dispatch contract

Give each fresh reviewer subagent:

- The **approved stage plan** (criteria + steps) — to check adherence.
- The **rubric**:
  1. **Plan adherence** — does the implementation satisfy each acceptance criterion?
  2. **Tests** — do they exist, do they pass, do they cover the plan's behavior?
  3. **Best practices** — lint/format, security, error handling, style.
- **Do not commit, do not create or switch branches**: review the working tree only — the orchestrator owns all git.
- **Run tests + linters yourself** to verify, not just read them. Treat genuine failures as findings; treat flakes as non-blockers.
- **Emit a structured TOON block** (via the `toon-output` skill) as your final output, with:
  - A `findings` list — each row: `severity` (`blocker` | `major` | `minor` | `nit`), `location` (file:area), `description`, `suggested_fix`.
  - A `satisfied` boolean verdict — `true` when no `blocker`/`major` findings remain.
- **No parseable TOON block = failed pass**: always end with the TOON block.

## Edge cases

| Scenario | Handling |
|---|---|
| **No test infrastructure** | Best-effort: implementer skips tests and notes it; reviewer waives the tests criterion as a non-blocker finding; loop proceeds on plan + best practices. |
| **No-progress implementer iteration** | Reviewer still runs on current state; halt after 2 consecutive no-progress iterations (step 3). |
| **Unparseable reviewer verdict** | Treat as a failed reviewer pass; surface raw output to the user; counts as one of the 3 iterations. If it is the final iteration, halt (step 3) — never clean completion. |
| **Max iterations, unresolved blocker/major findings** | Halt (step 3): keep the flagged stage `in-progress`, persist its findings and halt reason in the ticket, publish the accumulated work, ask the user. Never run later stages after a halt. |
| **No not-done stage at invocation** | The plan is already complete: report the open draft PR if one exists; otherwise publish via publish-it and stop. |
| **Stage qualifier (`pip stage 2`)** | Run only the named stage; if it is already `done`, report and exit. When it closes out, stop after the stage commit — publish only if the user asked. |
| **Plan branch exists but isn't checked out** | Discover it (`git branch --list`, or the plan's open PR head) and check it out — never `checkout -b` a second branch. |
| **Already on the plan's branch (resume)** | Stay on it; re-glob picks up the `in-progress` stage; grow the existing draft PR instead of stacking. |
| **First iteration** | Implementer gets the stage plan only (no prior findings) — unless the ticket carries `## Outstanding Review Findings` from a halted run. |

## Important notes

- The orchestrator owns **all** git: it creates the plan branch up front (step 1), commits at each stage boundary (step 2g), and invokes publish-it once at completion or halt (steps 3–4). Subagents never commit or branch.
- Bare `pip it` processes all not-done stages in one invocation; a stage qualifier runs a single stage and stops after its commit, publishing only on explicit request.
- The inner loop continues only on `blocker`/`major` findings; `minor`/`nit` are reported but don't extend it.
- Max 3 iterations per stage; early-stop on `satisfied`.
- A halt stops the entire plan: persist the halt reason and findings, keep the flagged stage `in-progress`, publish the accumulated work (qualified runs publish only on explicit request), ask the user.
- Each iteration spawns **fresh** subagents — continuity is the working tree (filesystem), not agent memory.
- Post a short summary to the user after each implementer pass and each reviewer pass, plus a stage-completion summary after each stage.
- One draft PR per plan — never per stage: publish-it publishes the plan branch itself, opening the PR if none exists or appending to it if one does.
- After publish, stop at the PR — no Maestro/Linear writeback.
