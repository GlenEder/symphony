---
name: git-worktree
description: Own the full git worktree lifecycle for isolated plan work — establish an isolated worktree at ~/.worktrees/<repo>/<plan-id> on the plan branch, bootstrap its dependencies from the detected lockfile, and reap (remove) worktrees whose plan PR has merged. Use when the user says 'set up a worktree', 'establish the worktree', 'worktree it', 'reap the worktrees', or 'clean up merged worktrees', or when plan-implementation-procedure needs plan work isolated so the main checkout stays free for other agents.
compatibility: opencode
---

## Purpose

Run plan work in an **isolated git worktree** so the main checkout of the repo stays free for other agents, sessions, and the user.

This skill owns three operations:

- **Establish** — create or reuse the worktree at `~/.worktrees/<repo-name>/<plan-id>` on the plan branch, including migrating a legacy plan branch that lives in the main checkout.
- **Bootstrap** — install dependencies inside the fresh worktree, detected from its lockfile.
- **Reap** — remove worktrees whose plan's PR has merged, and delete the merged local branch.

`<repo-name>` is the basename of the repo root.
`<plan-id>` is the plan's identifier — the same id as the work ticket filename (`~/.config/symphony/work_tickets/{plan-id}-stage-N.md`).

## Paths and invariants

```bash
REPO_ROOT=$(git rev-parse --show-toplevel)
REPO_NAME=$(basename "$REPO_ROOT")
WORKTREE_HOME="$HOME/.worktrees/$REPO_NAME"
WORKTREE_PATH="$WORKTREE_HOME/$PLAN_ID"
DEFAULT_BRANCH=$(git rev-parse --abbrev-ref origin/HEAD 2>/dev/null | sed 's|origin/||')
if [ -z "$DEFAULT_BRANCH" ]; then
  DEFAULT_BRANCH=$(git remote show origin 2>/dev/null | sed -n 's/.*HEAD branch: //p')
fi
DEFAULT_BRANCH=${DEFAULT_BRANCH:-main}
git rev-parse --verify --quiet "origin/$DEFAULT_BRANCH" >/dev/null || { echo "no default branch ref on origin" >&2; exit 1; }
```

If `origin/HEAD` is unset (common on fresh clones), fall back to `git remote show origin` (the `HEAD branch:` line) — or `git symbolic-ref refs/remotes/origin/HEAD` — before defaulting to `main`.
Then verify the **final** value once, after all fallbacks: `git rev-parse --verify origin/<branch>` — and **halt** when no remote default ref exists, rather than assume one into `git worktree add`.

Invariant: apart from the clean-tree legacy migration in step 2, this skill **never** checks out, commits, or switches branches in the main checkout.
All branch switching happens inside the worktree, or on the main checkout only in that one migration case.

## Workflow

### 1. Resolve inputs and detect current state

Collect, before touching anything:

- The **repo root** (`git rev-parse --show-toplevel`), the **plan-id**, and the **plan branch name** (from the plan's work ticket or the plan's open PR head).
- The **default branch**: resolve it as shown in **Paths and invariants** — `origin/HEAD`, then `git remote show origin`, then `main`, verifying the resolved ref exists.
- Whether `~/.worktrees/<repo-name>/<plan-id>` already exists, and whether it is a **registered** worktree: `git worktree list` — the path appears with its branch, or it does not.
- Whether the plan branch exists locally: `git branch --list <branch>`.
- Whether the plan branch is **checked out in the main checkout**: `git branch --show-current` in the repo root.
- Whether the main checkout is **clean**: `git status --porcelain` — empty output means clean.
- Whether you are already running **inside** a worktree: `git rev-parse --git-common-dir` and `git rev-parse --git-dir` differ when inside a worktree (the common dir points at the main repo's `.git`).

```bash
git worktree list
git rev-parse --git-common-dir
git rev-parse --git-dir
git status --porcelain
```

**Done when** the worktree path's existence and registration, the plan branch's existence and checkout location, the main tree's cleanliness, and the default branch are all known.

### 2. Establish — create or reuse the worktree

Handle exactly one case, in this order:

- **Reuse (resume)** — the worktree path exists **and** appears in `git worktree list`: use it as-is and continue on the plan branch there.
  Do not re-create, re-branch, or re-bootstrap an existing worktree.
- **Stale directory** — the path exists but is **not** a registered worktree (reap interrupted, manual `rm`, or a pruned registration): remove the leftover directory and re-run the applicable case below.
  This case must be checked **before** any case that runs `git worktree add` — an unregistered-but-existing path makes that command fail with `already exists`:

  ```bash
  rm -rf ~/.worktrees/<repo-name>/<plan-id>
  ```

  Only `rm` the directory after confirming `git worktree list` does not reference it and `git worktree prune` has run — the path then holds nothing git tracks.
- **Legacy migration** — the plan branch is checked out in the main checkout:
  - Main tree **clean**: move the branch out of the main checkout, then add it as the worktree:

    ```bash
    git checkout <default-branch>
    mkdir -p ~/.worktrees/<repo-name>
    git worktree add ~/.worktrees/<repo-name>/<plan-id> <branch>
    ```

  - Main tree **dirty**: **halt** and ask the user — never stash, discard, or commit their uncommitted work to make room for the migration.
  This is the only case that switches the main checkout's branch.
- **Branch exists, not checked out** — the worktree path does not exist (or was cleared per the stale-directory case), and the plan branch exists locally but is not checked out anywhere relevant: add the worktree on it:

  ```bash
  git worktree add ~/.worktrees/<repo-name>/<plan-id> <branch>
  ```

- **New branch** — the worktree path does not exist (or was cleared per the stale-directory case), and the plan branch does not exist: create it inside a new worktree, **always cut from the remote tip of the default branch** (`origin/<default-branch>`) — never from the local default branch, which may be stale or absent, and never from whatever the main checkout happens to have checked out:

  ```bash
  git worktree add -b <branch> ~/.worktrees/<repo-name>/<plan-id> origin/<default-branch>
  ```

**Done when** `git worktree list` shows `~/.worktrees/<repo-name>/<plan-id>` on the plan branch, the main checkout is untouched (or migrated per the clean-tree case), and you are operating inside the worktree path.

### 3. Bootstrap — install dependencies

Run **once**, only after the worktree was newly created — never on resume.
Inside the worktree, detect the lockfile and run its matching install:

| Lockfile present | Command |
|---|---|
| `package-lock.json` | `npm ci` |
| `pnpm-lock.yaml` | `pnpm install` |
| `yarn.lock` | `yarn install` |
| `poetry.lock` | `poetry install` |
| `uv.lock` | `uv sync` |
| `go.mod` or `go.sum` | `go mod download` |
| `Cargo.lock` | `cargo fetch` |
| none of the above | skip install, **note it** to the user |

Install for **every** recognized lockfile present — a polyglot repo gets each recognized install run once.

Never copy `.env`, `.env.local`, or any other `.env`-style file from the main checkout into the worktree — secret exposure.
If the worktree's tooling needs local config that is gitignored (`.env` and friends), report the missing local config to the user and let them provision it; do not silently work around it.

**Done when** the detected install command(s) have run successfully — or the no-lockfile / unrecognizable-lockfile case is noted — and any missing local config has been reported to the user.

### 4. Reap — remove worktrees with merged PRs

Operates over the whole worktree home, not just one plan:

```bash
for worktree in ~/.worktrees/<repo-name>/*/; do ...; done
git worktree list
```

For each worktree under `~/.worktrees/<repo-name>/`:

1. Determine its branch from `git worktree list`.
2. Find the plan's PR and check whether it is **merged**:

   ```bash
   gh pr list --head <branch> --state merged --json number,url
   ```

   If `gh pr list` returns empty, confirm with `gh pr view <branch> --json state` when a PR number is known from context.
3. **Merged**: remove the worktree and delete the merged local branch:

   ```bash
   git worktree remove ~/.worktrees/<repo-name>/<plan-id>
   git branch -D <branch>
   ```

   Use `-D`, not `-d`: the merge check in step 2 is the real safety gate.
   `-d` refuses here because the local default branch is often stale (not fetched since the merge), and squash-merged PRs are never ancestors of the default branch even when it is current — so a merged branch would fail the ancestry check and be left behind.

   Never pass `--force` to `git worktree remove` — its dirty-tree refusal is the safety net.
4. **Not merged / open PR**: skip it and report that it stays.
5. **Dirty worktree**: skip it and report the dirty paths — never `--force`, never discard.
6. **`gh` failure** (network, auth, not installed): skip the worktree and report the failure — never guess merge state.
7. **Unregistered directory** — the path exists under `~/.worktrees/<repo-name>/` but `git worktree list` does not reference it: confirm it is unreferenced, run `git worktree prune`, then remove the leftover directory and report that it was cleaned:

   ```bash
   ! git worktree list --porcelain | grep -F "<path>"
   git worktree prune
   rm -rf ~/.worktrees/<repo-name>/<dir>
   ```

   The first command exits 0 only when the path is **not** referenced — proceed only then; if it is referenced, reap it through `git worktree remove` instead.

   Never `rm` a directory still referenced by `git worktree list` — reap it through `git worktree remove` instead.

Never delete remote branches — `git branch -D` is local-only, and no `git push --delete` is ever run by this skill.

**Done when** every worktree under `~/.worktrees/<repo-name>/` has been either reaped (merged PR, clean tree) or reported with the reason it was skipped — and every reaped one had its merged local branch deleted.

## Edge cases

| Scenario | Handling |
|---|---|
| **Existing worktree for the plan** (step 2) | Reuse it as-is — never re-create, re-branch, or re-bootstrap; resume the run inside it. |
| **No recognizable lockfile** (step 3) | Skip the install and note it to the user; do not guess a package manager. |
| **`gh` unavailable or fails** (step 4) | Skip the worktree and report the failure; never guess PR merge state — an unknown state is never treated as merged. |
| **Dirty worktree on reap** (step 4) | Skip it and report the dirty paths; never `--force` — rely on `git worktree remove`'s dirty-tree refusal. |
| **Plan branch checked out in the main checkout** (step 2) | Clean tree: migrate — main checkout back to the default branch, then `git worktree add` the branch into the worktree home. Dirty tree: halt and ask the user; never risk their uncommitted work. |
| **Stale worktree directory not registered in git** (steps 2 and 4) | Confirm `git worktree list` does not reference it, run `git worktree prune`, `rm -rf` the leftover directory, and re-run the applicable establish case (step 2) or report the cleanup (reap, step 4, unregistered-directory branch). |
| **Merged branch fails `git branch -d`** (step 4) | Expected — the local default branch may be stale and squash-merged PRs share no ancestry, so the merged-PR check is the safety gate: delete with `git branch -D`. |

## Important notes

- Worktrees live at `~/.worktrees/<repo-name>/<plan-id>` — keyed by repo basename and plan-id, so multiple plans in one repo never collide and other repos' worktrees are untouched.
- The plan branch is **always** cut from the remote tip of the default branch (`origin/<default-branch>`, resolved per **Paths and invariants**) — never from the local default branch, which may be stale or absent, and never from whatever the main checkout has checked out.
- **Never copy `.env`-style files** into a worktree; report missing local config instead of working around it.
- Worktrees are **kept** after publish and after halts — removal happens only via reap, and only once the plan's PR has merged.
- Apart from the clean-tree legacy migration, this skill never checks out, commits, or switches branches in the main checkout — that invariant is the whole point of the isolation.
- Never pass `--force` to `git worktree remove`, and never delete remote branches.
- Bootstrap runs once per worktree creation; resume runs skip it.
