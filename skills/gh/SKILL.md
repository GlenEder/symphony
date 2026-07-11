---
name: gh
description: Expert guidance for using the GitHub CLI (gh) tool for repository management, pull requests, issues, actions, and more.
compatibility: opencode
---

## Purpose

This skill provides guidance on using the GitHub CLI (`gh`) for efficient GitHub operations. Use these commands and patterns when interacting with GitHub through the CLI.

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

## Pull Requests

### Creating PRs

```bash
# Create a PR with default params
gh pr create

# Draft PR
gh pr create --draft

# PR with title and body
gh pr create --title "feat(scope): description" --body "## Summary\n\nWhat this does..."

# PR targeting a specific base branch
gh pr create --base main --head my-branch

# Fill body from a template file
gh pr create --body-file .github/PULL_REQUEST_TEMPLATE.md

# PR with assignee, labels, reviewers
gh pr create --assignee @me --label bug,frontend --reviewer username

# PR with web browser fallback
gh pr create --web
```

### Reviewing & Managing PRs

```bash
# List PRs with filters
gh pr list                   # open PRs (default)
gh pr list --state all       # all states
gh pr list --state merged    # merged only
gh pr list --author @me      # your PRs
gh pr list --label bug       # filter by label
gh pr list --limit 50        # pagination limit

# View PR details
gh pr view 42                # view PR #42
gh pr view 42 --comments     # with comments
gh pr view 42 --json title,body,additions,deletions,files

# Checkout a PR locally
gh pr checkout 42

# Merge a PR
gh pr merge 42               # creates merge commit (default)
gh pr merge 42 --squash      # squash merge
gh pr merge 42 --rebase      # rebase merge
gh pr merge 42 --auto        # enable auto-merge

# Close or reopen a PR
gh pr close 42
gh pr reopen 42

# Review a PR
gh pr review 42 --approve
gh pr review 42 --comment --body "Looks good, but see inline notes."
gh pr review 42 --request-changes --body "Need to fix typing."

# Diff of a PR
gh pr diff 42

# Check CI status on a PR
gh pr checks 42

# Add PR to a project (v2) — see also gh project item-add below
gh pr edit 42 --add-project "Project Name"
```

## Issues

```bash
# Create an issue
gh issue create --title "Bug: login fails on Safari" --body "Steps to reproduce..."

# List issues
gh issue list
gh issue list --state all --label bug
gh issue list --search "performance in:title"

# View, close, reopen
gh issue view 123
gh issue close 123
gh issue reopen 123

# Edit issue
gh issue edit 123 --add-label frontend --remove-label needs-triage
gh issue edit 123 --add-assignee @me

# Comment on issue
gh issue comment 123 --body "I'll take a look at this."

# Transfer issue to another repo
gh issue transfer 123 owner/repo

# Pin/unpin issue
gh issue pin 123
gh issue unpin 123
```

## Repos

```bash
# Create a new repo
gh repo create my-project              # interactive
gh repo create my-project --public     # explicit visibility
gh repo create my-project --private --clone

# Create repo from template
gh repo create my-project --template owner/template-repo --clone

# Fork a repo
gh repo fork owner/repo --clone

# Clone a repo
gh repo clone owner/repo

# View repo info
gh repo view owner/repo
gh repo view owner/repo --web

# Open repo in browser
gh repo view --web          # current dir's repo
gh browse                    # shorthand

# List repos
gh repo list owner --limit 100

# Rename repo
gh repo rename new-name

# Archive/unarchive
gh repo archive
gh repo unarchive

# Delete repo (irreversible!)
gh repo delete owner/repo

# Manage repo settings
gh repo edit --description "New description" --homepage "https://..."
gh repo edit --default-branch main
gh repo edit --enable-issues=false
gh repo edit --visibility private
```

## Releases

```bash
# List releases
gh release list
gh release list --limit 30

# View release
gh release view v1.0.0
gh release view v1.0.0 --json name,tagName,createdAt,publishedAt

# Create release
gh release create v1.0.0 --title "v1.0.0" --notes "Release notes here"
gh release create v1.0.0 --generate-notes   # auto-generated notes
gh release create v1.0.0 --notes-file CHANGELOG.md
gh release create v1.0.0 --prerelease        # mark as pre-release
gh release create v1.0.0 --draft             # draft only

# Upload assets to release
gh release upload v1.0.0 ./dist/app.tar.gz

# Download assets from release
gh release download v1.0.0 --pattern "*.tar.gz"

# Delete release
gh release delete v1.0.0
```

## Actions (CI/CD)

```bash
# List workflows
gh workflow list
gh workflow list --all

# Run a workflow
gh workflow run build.yml
gh workflow run deploy.yml --ref main --field env=production

# List runs
gh run list
gh run list --workflow=build.yml --limit 20

# View run details
gh run view 1234
gh run view 1234 --log           # view full log
gh run view 1234 --log-failed    # only failed steps

# Watch run in real-time
gh run watch 1234

# Rerun a failed run
gh run rerun 1234
gh run rerun 1234 --failed       # only failed jobs

# Download logs
gh run download 1234

# Cancel a run
gh run cancel 1234

# Delete a run
gh run delete 1234

# View workflow YAML
gh workflow view build.yml
```

## Projects (v2)

```bash
# List projects for current user
gh project list

# List projects for an org
gh project list --owner org-name

# List projects including closed
gh project list --closed

# View a project
gh project view 1
gh project view 1 --owner monalisa --web
gh project view 1 --format json

# Create a project
gh project create --owner @me --title "Roadmap"

# Edit a project
gh project edit 1 --title "New Title" --description "Updated description"

# Copy a project
gh project copy 1 --title "Q4 Copy"

# Close / delete a project
gh project close 1
gh project delete 1

# Mark as template / unlink
gh project mark-template 1
gh project unlink 1 --repo owner/repo

# Link project to repo or team
gh project link 1 --repo owner/repo
gh project link 1 --team my-team

# List fields in a project
gh project field-list 1 --owner @me
gh project field-list 1 --owner @me --format json

# Create / delete a custom field
gh project field-create 1 --name "Sprint" --type text
gh project field-delete 1 --field-id <field-id>

# List items in a project
gh project item-list 1 --owner @me
gh project item-list 1 --owner @me --query "assignee:monalisa -status:Done"
gh project item-list 1 --limit 100

# Add an issue or PR to a project
gh project item-add 1 --url https://github.com/owner/repo/issues/42

# Create a draft issue in a project
gh project item-create 1 --title "Draft task" --body "Notes..."

# Edit an item in a project
gh project item-edit 1 --item-id <item-id> --field "Status" --value "In Progress"

# Archive / delete an item
gh project item-archive 1 --item-id <item-id>
gh project item-delete 1 --item-id <item-id>
```

**Scope requirement**: Token needs the `project` scope. Add it with:

```bash
gh auth refresh -s project
```

## Cache (Actions)

```bash
# List caches for current repo
gh cache list

# List for a specific repo
gh cache list --repo cli/cli

# List sorted by size (newest first)
gh cache list --sort size_in_bytes --order desc

# Filter by key prefix
gh cache list --key setup-

# Filter by branch ref
gh cache list --ref refs/heads/main
gh cache list --ref refs/pull/42/merge

# Output as JSON with selected fields
gh cache list --json id,key,sizeInBytes,ref,createdAt

# Delete a cache by ID
gh cache delete 1234

# Delete a cache by key
gh cache delete setup-node-abc123

# Delete cache by key on a specific branch
gh cache delete cache-key --ref refs/heads/feature-branch

# Delete all caches
gh cache delete --all

# Delete all caches for a specific ref (exit 0 on no caches)
gh cache delete --all --ref refs/pull/42/merge --succeed-on-no-caches
```

**Scope requirement**: Cache deletion requires the `repo` scope.

## Gists

```bash
# Create gist
gh gist create file.ts                        # public (default)
gh gist create file.ts --public
gh gist create *.ts                           # multi-file gist

# Create gist from stdin
echo "const x = 1;" | gh gist create

# List gists
gh gist list --limit 50

# View, edit, clone
gh gist view 1234
gh gist view 1234 --filename file.ts
gh gist edit 1234 file.ts --add "new content"

# Fork a gist
gh gist fork 1234

# Delete a gist
gh gist delete 1234
```

## API

```bash
# Make authenticated REST API calls
gh api /repos/owner/repo
gh api /repos/owner/repo/issues --field state=open --method GET

# Paginated results (all pages)
gh api /repos/owner/repo/issues --paginate

# POST/PUT/DELETE
gh api /repos/owner/repo/issues/42/comments --method POST --field body="Comment text"

# GraphQL queries
gh api graphql -f query='
  query {
    repository(owner: "owner", name: "repo") {
      pullRequests(first: 10, states: OPEN) {
        nodes { number title }
      }
    }
  }
'

# Use jq for filtering JSON output
gh api /repos/owner/repo/pulls | jq '.[] | {number: .number, title: .title}'

# Headers
gh api -H "Accept: application/vnd.github.v3.raw" /repos/owner/repo/contents/README.md
```

## Secrets & Variables

```bash
# List secrets
gh secret list
gh secret list --repo owner/repo
gh secret list --org my-org

# Set secrets
gh secret set MY_SECRET --body "value"
gh secret set MY_SECRET --body "$(cat secret.txt)"

# Remove secrets
gh secret remove MY_SECRET

# List variables (non-secret)
gh variable list
gh set MY_VAR --body "value"
gh variable delete MY_VAR
```

## Notifications

```bash
# List notifications
gh notification list
gh notification list --since 2024-01-01

# Mark as read
gh notification read --id 123
gh notification read --all
```

## Search

```bash
# Search code
gh search code "import React" --owner owner --language tsx

# Search issues/PRs
gh search issues "bug in login" --label bug --state open
gh search prs "feat" --author @me --state merged

# Search repos
gh search repos "topic:react stars:>1000"
```

## Rulesets

```bash
# List rulesets for current repository
gh ruleset list

# List rulesets for another repo (including inherited)
gh ruleset list --repo owner/repo --parents

# List organization-wide rulesets
gh ruleset list --org my-org

# Open ruleset list in browser
gh ruleset list --web

# Interactively select ruleset to view
gh ruleset view

# View a specific ruleset by ID
gh ruleset view 43

# View ruleset from another repo
gh ruleset view 23 --repo owner/repo

# View org-level ruleset
gh ruleset view 23 --org my-org

# Open ruleset in browser
gh ruleset view --web

# Check rules that apply to current branch
gh ruleset check

# Check rules for a specific branch
gh ruleset check my-branch

# Check rules for the default branch
gh ruleset check --default
```

**Alias**: `gh rs` can be used instead of `gh ruleset`.

**Scope requirement**: Listing org rulesets needs the `admin:org` scope (`gh auth refresh -s admin:org`).

## Configuration

```bash
# Check gh config
gh config list

# Set editor
gh config set editor code

# Set git protocol
gh config set git_protocol ssh   # default: https

# Set pager
gh config set pager less

# Set host
gh config set host example.com
```

## Copilot

> **Status**: GitHub Copilot CLI is currently in preview and subject to change.

```bash
# Run Copilot interactively
gh copilot

# Run with a prompt (non-interactive)
gh copilot -p "Summarize this week's commits"

# Allow all tools automatically (for scripting)
gh copilot -p "Fix the bug in main.js" --allow-all-tools

# Enable all permissions (tools, paths, URLs)
gh copilot -p "Fix the bug in main.js" --allow-all
gh copilot -p "Fix the bug in main.js" --yolo

# Use a specific model
gh copilot --model gpt-5.2

# Resume the most recent session
gh copilot --continue

# Resume a specific session
gh copilot --resume=<session-id>

# Grant specific tool permissions
gh copilot -p "Update package.json" --allow-tool='write'

# Grant shell with restrictions
gh copilot --allow-tool='shell(git:*)' --deny-tool='shell(git push)'

# Allow access to directories outside cwd
gh copilot --add-dir /home/user/projects

# Configure additional MCP servers
gh copilot --additional-mcp-config '{"my-server":{"command":"node","args":["server.js"]}}'

# Share session to markdown file on completion
gh copilot -p "Refactor auth" --share

# Share session to secret gist
gh copilot -p "Refactor auth" --share-gist

# Initialize Copilot instructions for a repo
gh copilot init

# Remove the Copilot CLI (if installed via gh)
gh copilot --remove

# Pass through flags to Copilot (use -- before them)
gh copilot -- --help
```

**Key flags reference**:

| Flag | Purpose |
|------|---------|
| `-p`, `--prompt` | Execute a prompt in non-interactive mode |
| `--model` | Set the AI model (`gpt-5.2`, `claude-4`, etc.) |
| `--allow-all`, `--yolo` | Allow all tools, paths, and URLs |
| `--allow-all-tools` | Auto-approve all tool calls (for CI/automation) |
| `--allow-tool` | Allow specific tool(s) without prompting |
| `--deny-tool` | Deny specific tool(s) |
| `--add-dir` | Add directory to allowed file access paths |
| `--resume` | Resume a previous session |
| `--continue` | Resume the most recent session |
| `--share` / `--share-gist` | Export session to file/gist on completion |
| `--additional-mcp-config` | Configure extra MCP servers for the session |
| `--reasoning-effort` | Set reasoning level (`low`, `medium`, `high`, `xhigh`) |
| `--silent` | Output only the agent response (for scripting) |
| `--remove` | Remove the Copilot CLI downloaded by gh |

## Useful Flags & Patterns

```bash
# JSON output with specific fields
gh pr list --json number,title,author,headRefName,createdAt

# Custom template output (Go templates)
gh pr list --template '{{range .}}{{.title}} (#{{.number}}) by {{.author.login}}{{"\n"}}{{end}}'

# Combine with jq for advanced filtering
gh pr list --json number,title,labels --jq '.[] | select(.labels | length == 0)'

# Filter by dates
gh pr list --search "created:>2024-01-01"
gh issue list --search "updated:<=2024-06-01"

# Output format
gh pr view 42 --json number,title  # JSON
gh pr view 42 --template '{{.title}}'  # Go template

# Browser shortcuts
gh browse                    # open current repo in browser
gh browse --settings         # repo settings page
gh browse --wiki             # repo wiki
gh browse --projects         # repo projects
gh browse -b main -- src/    # open file at specific branch/path

# Use pagination with all items
gh search prs "state:merged" --limit 1000 --json number --paginate | jq 'length'

# Get repo home directory info
gh repo set-default owner/repo  # set default repo for current dir
```

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
