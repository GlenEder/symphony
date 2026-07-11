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

# Add PR to a project (v2) — see also `gh project item-add`
gh pr edit 42 --add-project "Project Name"
```
