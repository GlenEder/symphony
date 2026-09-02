---
name: publish-it
description: Publish uncommitted working-tree changes to a draft PR when the user says 'publish it', 'push it up', 'pr it', or wants small fixes shipped without a planning pass. Opens a new draft PR, or appends to the existing draft PR when the current branch already has one — a resumed plan run grows its single plan PR, never a stacked one. Worktree-aware: when the cwd is a git worktree (typically a pip plan run) it detects this, publishes exactly as on the plan branch, and leaves worktree teardown to git-worktree. Use plan-implementation-procedure ('pip it') when the work needs implementing or tests; publish-it acts only on changes already in the working tree.
compatibility: opencode
---

## Workflow

### 1. Assess Current State

Run in parallel:

- `git branch --show-current` — the **current branch**
- `git rev-parse --abbrev-ref origin/HEAD 2>/dev/null | sed 's|origin/||'` (fallback `main`) — the **default branch**
- `git status` — staged / unstaged / untracked changes
- `git diff` — the changes that will ship
- `gh pr list --head <current-branch> --state open --json number,url,title` — any **open PR** on the current branch
- `git log <default-branch>..HEAD --oneline` — commits already on the current branch; the plan's stage commits (e.g. `feat(<scope>): stage N — …`) mark it as the **plan branch**
- `git rev-parse --git-common-dir` and `git rev-parse --git-dir` — they **differ** when cwd is inside a git worktree (the common dir points at the main repo's `.git`); a cwd under `~/.worktrees/` is also a worktree

Note whether **plan-implementation-procedure invoked you**: then the current branch is its **plan branch** and must become the single plan PR.

If the two `git rev-parse` calls differ, or the cwd is under `~/.worktrees/`, state to the user that you are running **inside a worktree** (typically a pip plan run): plan-branch mode may apply, per the caller's context exactly as today — detection is informational and never selects a mode by itself.

If the `gh` call fails (network/auth), say so to the user and treat the open-PR state as unknown — the fallback is the new-PR flow in step 3, exactly the pre-append behavior (the plan-branch case is unaffected: it is detected from the caller and the log, not from `gh`).

**Done when** every command has run and its output is accounted for: you know the current branch, the default branch, the changes to ship, the current branch's open PR (or that detection failed), whether the current branch is a plan branch, and whether the cwd is a worktree.

### 2. Generate Branch Name

From `git diff` and `git status`, derive `type`, `scope`, and a short `description` per [Conventional Commits](../_shared/conventional-commits.md).

Default branch name: `<type>/<scope>-<short-description>`. If scope is unclear, use `<type>/<short-description>`.

If a Linear issue key is in context (the user mentioned it, or it appears in the current branch, a commit, or a linked PR), follow the workspace convention instead: `<linear-issue-key>/<shortWorkDescription>` (e.g. `cops-308/fixLoginTimeout`). Do not invent a key; if none is present, use the default form.

Only new-PR mode (step 3) creates a branch; in the other modes reuse the current branch and use this derivation only for the PR title (step 6), taking it from the plan's scope or the branch name when the tree is clean.

**Done when** the branch name matches the chosen convention and reflects an actual change present in `git diff`, or — when no branch is created — a PR title is derived for step 6.

### 3. Create or Reuse Branch

State the mode to the user, then:

- **Append mode** — the current branch has an open PR (step 1): stay on it and skip branch creation; the push in step 5 updates that PR (a resumed plan run grows its single plan PR this way).
  Never on the default branch — a PR headed there is not this skill's to grow.
- **Plan-branch mode** — the current branch is the plan branch (plan-implementation-procedure invoked you, or step 1's log shows the plan's stage commits) with no open PR yet, e.g. a first halt: stay on it and publish it directly, never a second branch or stacked PR for the plan.
  When the cwd is a worktree (step 1), plan-branch mode behaves identically: `git push` and `gh` work normally from a worktree, so stay on the plan branch, push it, and open or append the single plan PR — no special handling.
  publish-it never removes or prunes a worktree: teardown belongs to the calling orchestrator via [git-worktree](../git-worktree/SKILL.md)'s reap.
- **New-PR mode** — no open PR on the current branch and not a plan-branch scenario: create the new branch from current HEAD:

```bash
git checkout -b <branch-name>
```

In new-PR mode the PR (step 6) targets the **current branch from step 1**. State which case you are in to the user:

- Current branch is the default branch → a normal PR against default.
- Current branch is a feature branch → a **stacked PR** onto that feature branch (the new branch carries only the uncommitted changes; the feature branch's existing commits are its base, not part of this PR's diff).

**Done when** the mode is stated to the user and `git branch --show-current` returns the branch for that mode — the new branch in new-PR mode, the unchanged current branch otherwise.

### 4. Commit

Stage every change identified in step 1 and commit per [Conventional Commits](../_shared/conventional-commits.md):

```bash
git add <files>
git commit -m "<type>(<scope>): <description>"
```

If step 1 found no changes — a clean tree, e.g. plan-branch mode after the orchestrator committed every stage — skip the commit.

**Done when** `git status` is clean — every change from step 1 is committed, or there was nothing to commit.

### 5. Push

```bash
git push -u origin <branch-name>
```

`<branch-name>` is the branch chosen in step 3: the newly created one in new-PR mode, the current branch otherwise.

**Done when** the branch is pushed and remote tracking is set.

### 6. Open or Update Draft PR

- **Append mode**: do not create a PR — the push in step 5 already updated the open PR from step 1; report its URL.
- **Otherwise**: open a draft PR.

```bash
gh pr create --draft --base <base> --title "<type>(<scope>): <description>" --body "$(cat <<'EOF'
## Summary
<Brief summary of changes>

## Changes Made
- <List of key changes>

## Testing
<How to test these changes>

## Notes
<Any additional notes>
EOF
)"
```

`<base>` is the **current branch from step 1** in new-PR mode, or the **default branch** in plan-branch mode — the single plan PR for the whole plan.

**Done when** the draft PR is opened against `<base>` and its URL is returned to the user, or — in append mode — the updated PR's URL is returned.

### 7. Follow-up

Only when the user asks for more changes after the draft PR exists: make the changes, commit per [Conventional Commits](../_shared/conventional-commits.md), and push to the branch carrying the open PR — never a new branch, never a stacked PR.

```bash
git add <files>
git commit -m "<type>(<scope>): <description>"
git push
```

**Done when** the new changes are committed and pushed to that branch and the draft PR is updated.

publish-it never removes or prunes a worktree, including the one it ran in: teardown belongs to the calling orchestrator via [git-worktree](../git-worktree/SKILL.md)'s reap, which removes worktrees whose plan PR has merged.

## Edge cases

| Scenario | Handling |
|---|---|
| **cwd is a plan worktree with no open PR** (e.g. a first halt) | Publish the plan branch directly (plan-branch mode, step 3) and leave the worktree in place — teardown is the orchestrator's, via [git-worktree](../git-worktree/SKILL.md). |
| **cwd is a plan worktree with an open PR** | Append to the single plan PR (append mode, step 3) and leave the worktree in place — teardown is the orchestrator's, via [git-worktree](../git-worktree/SKILL.md). |
