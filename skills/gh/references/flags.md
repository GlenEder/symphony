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
