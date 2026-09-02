---
name: bb
description: Expert guidance for using the Bitbucket CLI (bb) tool for repository management, pull requests, pipelines, issues, and more.
compatibility: opencode
---

## Purpose

This skill provides guidance on using the Bitbucket CLI (`bb`) for repositories hosted on Bitbucket (bitbucket.org), as an equivalent to the `gh` skill for GitHub. Check the git remote (`git remote -v`) before reaching for `gh` — a `bitbucket.org` remote means `bb` is the right tool, not `gh`.

## Setup & Authentication

```bash
bb profile which              # current default profile name
bb profile list -o table      # all configured profiles, which is default
bb profile create -n <name> --default-workspace <ws> ...   # new profile (interactive prompts for auth if no token flags given)
bb profile authorize          # complete an Authorization Code Grant login
bb profile use <name>         # switch default profile
```

Credentials are stored in the OS credential vault by default (`--no-vault` stores them in plaintext config — avoid this). Repository and workspace are auto-detected from the current directory's git config, so most commands need no `--repository`/`--workspace` flags when run inside the repo.

## Global Flags (apply to every subcommand)

| Flag | Purpose |
|---|---|
| `-o, --output` | `json`, `yaml`, or `table` — use `json` for scripting/parsing |
| `-p, --profile` | override the default profile |
| `--workspace`, `--repository` | override the git-derived workspace/repo |
| `--dry-run` / `--whatif` / `--noop` | preview without making changes |
| `--stop-on-error` / `--ignore-errors` / `--warn-on-error` | error handling for multi-item commands |
| `--debug`, `-v` | verbose/debug logging when a command fails unexpectedly |

## Quick Reference

| Topic | Command group | Key subcommands |
|---|---|---|
| Pull Requests | `bb pullrequest` (alias `bb pr`) | `create`, `list`, `get`, `update`, `approve`/`unapprove`, `request-changes`, `merge`, `decline`, `diff`, `patch`, `commits`, `comment`, `task`, `activities` |
| Repos | `bb repo` (alias `bb repository`) | `create`, `clone`, `get`, `list`, `update`, `delete`, `fork` |
| Branches | `bb branch` | `list` |
| Pipelines | `bb pipeline` (aliases `pipe`, `pp`) | `trigger`, `list`, `get`, `stop`, `step` |
| Issues | `bb issue` | `create`, `list`, `get`, `update`, `delete`, `comment`, `attachment`, `watch`/`unwatch`, `vote`/`unvote` |
| Projects | `bb project` | `create`, `list`, `get`, `update`, `delete`, `reviewer` |
| Workspaces | `bb workspace` | `get`, `list`, `permission` |
| Commits | `bb commit` | `get`, `list`, `diff`, `patch`, `ancestor` |
| Tags | `bb tag` | `create`, `get`, `list`, `delete` |
| Profiles | `bb profile` | `create`, `authorize`, `use`, `which`, `list`, `update`, `delete` |
| Artifacts | `bb artifact` | `upload`, `download`, `list`, `delete` |
| Components | `bb component` | `get`, `list` |

Run `bb <group> --help` or `bb <group> <command> --help` for the full flag list of any command — help output is authoritative and complete for this CLI.

## Pull Requests (most common workflow)

```bash
# Create a draft PR (equivalent to gh pr create --draft)
bb pullrequest create \
  --source <branch> \
  --destination development \
  --title "fix(scope): description" \
  --description "$(cat <<'EOF'
## Summary
...
EOF
)" \
  --draft \
  -o json

# List open PRs (default state is "open")
bb pullrequest list --state open -o json
bb pullrequest list --state merged -o json

# Get/update a specific PR
bb pullrequest get <id> -o json
bb pullrequest update <id> --title "..." --description "..." --add-reviewer <account-id>

# Review actions
bb pullrequest approve <id>
bb pullrequest request-changes <id>
bb pullrequest comment create --pullrequest <id> --comment "..." [--file <path> --line <n>]

# Merge (strategies: merge_commit | squash | fast_forward)
bb pullrequest merge <id> --merge-strategy squash --close-source-branch
```

If `<pullrequest-id>` is omitted from `get`/`approve`/`merge`/`decline`/`request-changes`, `bb` tries to resolve the single open PR on the current branch — pass the id explicitly when there's any ambiguity.

`bb pullrequest create` does not auto-assign reviewers unless the repo/project has default reviewers configured — those are added automatically (as seen when creating without `--reviewer`).

## Pipelines (CI/CD)

```bash
bb pipeline trigger --branch <branch>              # run default pipeline for a branch
bb pipeline trigger --pattern deploy-to-prod --branch <branch>
bb pipeline trigger --variable KEY=value --variable KEY2=value2
bb pipeline list -o json
bb pipeline get <uuid-or-build-number> -o json
bb pipeline step list <pipeline-uuid>
bb pipeline step logs <pipeline-uuid> <step-uuid>
bb pipeline stop <uuid-or-build-number>
```

## Common Gotchas

- **This is Bitbucket, not GitHub** — always confirm with `git remote -v` before assuming `gh` applies; the two CLIs are not interchangeable and target different platforms/APIs.
- **`gh`-style flags don't map 1:1** — e.g. there is no `bb pr checkout`; use `git fetch origin <branch> && git checkout <branch>` directly. Confirm a flag/subcommand exists with `--help` rather than assuming `gh` naming carries over.
- **`--output json`** is the reliable way to get IDs (PR id, pipeline UUID) for chaining into a follow-up command — table output is for humans.
- **Merge strategy defaults to `merge_commit`** — pass `--merge-strategy squash` or `--merge-strategy fast_forward` explicitly for other strategies (mirrors `gh pr merge`'s default).
- **Workspace/repo auto-detection** relies on the git remote pointing at `bitbucket.org` — commands run outside a git repo (or against a non-Bitbucket remote) need explicit `--workspace`/`--repository`.
- **`--dry-run`/`--whatif`/`--noop`** are three names for the same flag — use before any destructive command (`repo delete`, `tag delete`, `issue delete`).
