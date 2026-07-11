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
