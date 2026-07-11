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
