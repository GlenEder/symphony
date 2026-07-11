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
