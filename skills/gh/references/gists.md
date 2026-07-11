## Gists

```bash
# Create gist
gh gist create file.ts                        # public (default)
gh gist create file.ts --public
gh gist create *.ts                           # multi-file gist

# Create gist from stdin
echo "const x = 1;" | gh gist create

# List gists
gh gist list --limit 50

# View, edit, clone
gh gist view 1234
gh gist view 1234 --filename file.ts
gh gist edit 1234 file.ts --add "new content"

# Fork a gist
gh gist fork 1234

# Delete a gist
gh gist delete 1234
```
