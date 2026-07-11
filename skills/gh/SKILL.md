---
name: gh
description: Expert guidance for using the GitHub CLI (gh) tool for repository management, pull requests, issues, actions, and more.
compatibility: opencode
---

## Purpose

This skill provides guidance on using the GitHub CLI (`gh`) for efficient GitHub operations. Individual command groups are in `references/` — load the ones you need.

## Common Setup & Authentication

```bash
# Check auth status
gh auth status

# Login (interactive or token-based)
gh auth login
gh auth login --with-token < ~/path/to/token.txt  # scripted

# List authenticated accounts
gh auth status --show-token  # shows token (use carefully)
```

## Quick Reference

Load the relevant reference file for detailed commands on each topic:

| Topic | File | Key commands |
|---|---|---|
| Pull Requests | `read references/pull-requests.md` | create, list, view, checkout, merge, review, close, reopen |
| Issues | `read references/issues.md` | create, list, view, edit, comment, transfer, pin |
| Repos | `read references/repos.md` | create, fork, clone, view, list, rename, archive, delete, edit |
| Releases | `read references/releases.md` | list, view, create, upload, download, delete |
| Actions (CI/CD) | `read references/actions.md` | workflow list/run, run list/view/watch/rerun/cancel/delete |
| Cache (Actions) | `read references/cache.md` | list (sort/filter/key/ref), delete (by id/key/all/ref) |
| Projects (v2) | `read references/projects.md` | list, view, create, edit, copy, close, delete, field/item management |
| Gists | `read references/gists.md` | create, list, view, edit, fork, delete |
| API | `read references/api.md` | REST calls, pagination, GraphQL, jq filtering |
| Secrets & Variables | `read references/secrets.md` | list, set, remove (repo + org scope) |
| Notifications | `read references/notifications.md` | list, mark read |
| Search | `read references/search.md` | code, issues/PRs, repos |
| Rulesets | `read references/rulesets.md` | list, view, check (org/repo scope, alias `gh rs`) |
| Configuration | `read references/configuration.md` | config list/set (editor, protocol, pager, host) |
| Copilot | `read references/copilot.md` | interactive, prompt, permissions, MCP, model, resume, share |
| Flags & Patterns | `read references/flags.md` | JSON output, Go templates, jq, browser shortcuts, pagination |

## Best Practices

1. **Always use `--json` with explicit fields** for scriptable output — avoids parsing human-readable text
2. **Prefer `--jq` for simple filtering** over piping to `jq` when possible (one process)
3. **Use `--paginate` for large result sets** — gh automatically fetches all pages
4. **Set a default repo** with `gh repo set-default` to avoid specifying `owner/repo` on every command
5. **Quote or heredoc long GraphQL queries** — use `-f query='...'` with single quotes
6. **Use `--draft` for work-in-progress PRs** created by automation/agents
7. **Use `gh run watch`** to monitor CI in terminal without switching to browser
8. **Check `gh auth status` first** if commands fail unexpectedly
9. **Use `gh browse`** to quickly open the browser at the right page (supports file paths and branches)
10. **Avoid `gh repo delete` in scripts** — it's irreversible with no undo
11. **Add the `project` scope** via `gh auth refresh -s project` before using `gh project` commands
12. **Use `gh cache list` before `gh cache delete`** to confirm cache IDs — deletion is irreversible
13. **Use `gh copilot --allow-all-tools -p` for agent automation** — the `-p` flag enables non-interactive scripting
14. **Use `gh ruleset check <branch>` before pushing** to see what rules apply to a branch
15. **Prefer `gh rs` as a shorthand** for `gh ruleset` in interactive use

## Common Gotchas

- **Large output**: `gh` may truncate unless you use `--limit` or `--paginate`
- **SSH vs HTTPS**: `gh auth login` handles both; configure with `gh config set git_protocol ssh`
- **Org secrets**: need `--org` flag and org membership with appropriate permissions
- **GitHub App tokens**: use `--with-token` for non-user authentication
- **Rate limiting**: unauthenticated requests get 60/hr; authenticated gets 5000/hr
- **`gh pr merge` defaults to merge commit** — use `--squash` or `--rebase` explicitly for other strategies
- **`gh project` requires the `project` scope** — `gh auth refresh -s project` before use
- **`gh cache delete` is irreversible** — confirm cache IDs with `gh cache list` first
- **`gh copilot` is in preview** — flags may change between gh CLI versions
- **`gh ruleset` defaults to `--parents true`** — use `--no-parents` to see only repo-level rulesets
