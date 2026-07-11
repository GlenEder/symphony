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
