---
name: plan-implementation-procedure
description: Orchestrate an implementer↔reviewer subagent loop on a work ticket (max 3 iterations) until the reviewer is satisfied, then invoke publish-it to open a draft PR. Use when the user says "pip it", "implement the plan", or "execute the plan". NOT for authoring or decomposing a plan (maestro-author), serving or approving a plan (maestro-session), exporting work tickets (maestro-export), interviewing decisions (grilling), or routing an ambiguous request (maestro).
compatibility: opencode
---

## Purpose

When the user says "pip it" or "implement the plan" against an approved work ticket, orchestrate an autonomous loop of two subagents:

- An **implementer** subagent (medium-tier) edits the working tree to satisfy the plan and writes tests.
- A **reviewer** subagent (strong-tier) reviews the implementation + tests against the plan and best practices, emitting structured findings and a verdict.

The loop runs up to **3 iterations** (implementer → reviewer per iteration). It stops early when the reviewer is satisfied (no blocker/major findings). When the loop ends, invoke the **publish-it** skill to open a draft PR. The loop never commits and never creates a branch — all git is deferred to publish-it.

## Example Triggers
- pip it
- pip this
- implement the plan
- execute the plan

## Workflow

### 0. Load the approved work ticket

Read the canonical work ticket:

- **Primary staged path**: glob `~/.config/symphony/work_tickets/{plan-id}-stage-*.md`, select the lowest N whose header contains `Status: pending`, and implement only that stage.
- At loop start, change the selected ticket's status from `pending` to `in-progress`.
- After `publish-it` opens the draft PR, change that ticket's status from `in-progress` to `done`.
- **Legacy fallback**: use `~/.config/symphony/work_tickets/{plan-id}.md` for single-stage plans.
- **Plan JSON fallback**: use `$MAESTRO_PLANS_DIR/{plan-id}.json` (parse its `modules` — especially `criteria` and `steps`).
- For a standalone `pip it` (no composer stage): use the plan content already in context, or the plan file the user points to.

Extract the acceptance criteria and implementation steps for the selected stage; these are what the implementer works to and the reviewer checks against.

Done when: the selected stage ticket's criteria + steps are in hand to pass to subagents, and its status is `in-progress`.

### 1. Branch check — do NOT branch

The loop works in the **working tree on the current branch** and never commits or creates a branch. `publish-it` creates the branch and commits at the end.

- If on the default branch: stay there. publish-it will branch from default.
- If already on a feature branch: stay there. publish-it will open a stacked PR onto that branch.

```bash
CURRENT_BRANCH=$(git branch --show-current)
DEFAULT_BRANCH=$(git rev-parse --abbrev-ref origin/HEAD 2>/dev/null | sed 's|origin/||')
DEFAULT_BRANCH=${DEFAULT_BRANCH:-main}
```

Done when: confirmed the loop will not create or switch branches.

### 2. The orchestration loop (max 3 iterations)

Initialize `iteration=0`, `latest_findings=""` (empty on iteration 1), `no_progress_streak=0`.

Loop while `iteration < 3`:

```bash
iteration=$((iteration + 1))
```

#### 2a. Dispatch the implementer subagent

Spawn a **fresh** implementer subagent (foreground/blocking). In Cursor, the Task tool with `subagent_type: generalPurpose` and a **medium-tier** model; in Maki, the `task` tool (general) at medium tier.

Pass it the **implementer dispatch contract** (below): the plan, the latest reviewer findings (empty on iteration 1), and the constraints (read working tree first, do not commit or branch, run tests + linters before reporting done).

Receive the implementer's summary: what it changed, whether tests/lint pass, and whether it made progress.

Done when: the implementer subagent has returned a change summary.

#### 2b. Post the implementer summary to the user

Post a short summary of what the implementer changed (files, behavior, test/lint status) and the iteration count.

Done when: the user has been informed of the implementer pass.

#### 2c. Handle no-progress

If the implementer made **no changes** (or errored out):

- Increment `no_progress_streak`.
- If `no_progress_streak >= 2`: **halt and ask the user** whether to continue, adjust, or abort. Do not publish.
- Else: still run the reviewer on the current state (it may find the state is actually fine, or surface what's blocking progress).

If the implementer made progress, reset `no_progress_streak=0`.

Done when: the no-progress streak is updated and the halt decision is made.

#### 2d. Dispatch the reviewer subagent

Spawn a **fresh** reviewer subagent (foreground/blocking). In Cursor, the Task tool with `subagent_type: generalPurpose` and a **strong-tier** model; in Maki, the `task` tool (general) at strong tier.

Pass it the **reviewer dispatch contract** (below): the plan, the rubric (plan adherence, tests, best practices), instruction to run tests + linters, and the requirement to emit a structured TOON block (findings + `satisfied` verdict) via the `toon-output` skill.

Receive the reviewer's TOON block.

Done when: the reviewer subagent has returned a TOON block.

#### 2e. Post the reviewer summary to the user

Post a short summary: the findings (by severity), the `satisfied` verdict, and the iteration count.

Done when: the user has been informed of the reviewer pass.

#### 2f. Parse the verdict and decide

Parse the reviewer's TOON block (use the `toon` skill's parsing rules):

- If `satisfied == true` **or** there are no `blocker`/`major` findings → **loop done** (go to step 3).
- Else: extract the `blocker`/`major` findings → set `latest_findings` to them, and loop again (2a). `minor`/`nit` findings are reported but do not extend the loop.

If no parseable verdict is found, treat it as a failed reviewer pass: surface the raw reviewer output to the user, and loop again (counts as one of the 3).

Done when: the verdict is parsed and the continue/stop decision is made.

### 3. Publish — invoke publish-it

When the loop is done (satisfied, or 3 iterations reached), invoke the **publish-it** skill on the accumulated working-tree changes:

- publish-it creates the branch (Linear-key convention if a key is in context, else `<type>/<scope>-<desc>`), commits the changes, pushes, and opens a **draft** PR.
- The loop's only git action is this publish-it call.

#### 3a. Max-iter transparency

If the loop ended at **max iterations (3) with unresolved findings**:

- Surface the outstanding `blocker`/`major` findings and any failing tests to the user.
- Include them in the draft PR description under an **"Outstanding Review Findings"** section so human reviewers see what's known-broken.

Done when: publish-it has opened the draft PR and its URL is reported to the user.

### 4. Stop at the PR

After publish-it opens the draft PR, mark the selected stage ticket `Status: done`, then stop. Report the implemented stage and PR URL to the user. Do not write back to Maestro or Linear. (The Maestro approval-time closeout — agent offline, heartbeat stop, final ack — already happened before the loop started.)

Done when: the draft PR URL is reported and no further writeback is attempted.

## Implementer dispatch contract

Give each fresh implementer subagent:

- The **approved plan** (criteria + steps), verbatim or summarized.
- The **latest reviewer findings** (from the previous iteration's TOON block, `blocker`/`major` only). Empty on iteration 1.
- **Read the working tree first**: prior iterations' edits live in the working tree — read the current state of the files you will touch before editing. Do not assume a fresh checkout.
- **Do not commit, do not create or switch branches**: edit the working tree only. The orchestrator handles git at the end via publish-it.
- **Write tests** for the work, following the repo's existing test patterns and directory structure.
- **Run tests + linters before reporting done**: report pass/fail. If the repo has no test infrastructure, skip tests and note that.
- **Report back**: a concise summary of what changed, test/lint status, and whether progress was made.

## Reviewer dispatch contract

Give each fresh reviewer subagent:

- The **approved plan** (criteria + steps) — to check adherence.
- The **rubric**:
  1. **Plan adherence** — does the implementation satisfy each acceptance criterion?
  2. **Tests** — do they exist, do they pass, do they cover the plan's behavior?
  3. **Best practices** — lint/format, security, error handling, style.
- **Run tests + linters yourself** to verify, not just read them. Treat genuine failures as findings; treat flakes as non-blockers.
- **Emit a structured TOON block** (via the `toon-output` skill) as your final output, with:
  - A `findings` list — each row: `severity` (`blocker` | `major` | `minor` | `nit`), `location` (file:area), `description`, `suggested_fix`.
  - A `satisfied` boolean verdict — `true` when no `blocker`/`major` findings remain.
- **No parseable TOON block = failed pass**: always end with the TOON block.

## Edge cases

| Scenario | Handling |
|---|---|
| **No test infrastructure** | Best-effort: implementer skips tests and notes it; reviewer waives the tests criterion as a non-blocker finding; loop proceeds on plan + best practices. |
| **No-progress implementer iteration** | Reviewer still runs on current state; halt and ask the user after 2 consecutive no-progress iterations. |
| **Unparseable reviewer verdict** | Treat as a failed reviewer pass; surface raw output to the user; counts as one of the 3 iterations. |
| **Max iterations, still unsatisfied** | Publish anyway via publish-it; surface outstanding blocker/major findings + failing tests to the user and in the PR's "Outstanding Review Findings" section. |
| **Already on a feature branch** | Stay; publish-it opens a stacked PR onto that branch. |
| **First iteration** | Implementer gets the plan only (no prior findings). |

## Important notes

- The loop never commits and never branches; publish-it is the only git action.
- Each iteration spawns **fresh** subagents — continuity is the working tree (filesystem), not agent memory.
- The loop continues only on `blocker`/`major` findings; `minor`/`nit` are reported but don't extend it.
- Max 3 iterations; early-stop on `satisfied`.
- Post a short summary to the user after each implementer pass and each reviewer pass.
- After publish, stop at the PR — no Maestro/Linear writeback.
