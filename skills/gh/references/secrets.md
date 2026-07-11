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
