## Cache (Actions)

**Scope requirement**: Cache deletion requires the `repo` scope.

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
