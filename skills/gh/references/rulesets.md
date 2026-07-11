## Rulesets

**Alias**: `gh rs` can be used instead of `gh ruleset`.

**Scope requirement**: Listing org rulesets needs the `admin:org` scope (`gh auth refresh -s admin:org`).

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
