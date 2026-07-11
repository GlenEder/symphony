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
