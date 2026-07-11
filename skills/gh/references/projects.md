## Projects (v2)

**Scope requirement**: Token needs the `project` scope. Add it with:

```bash
gh auth refresh -s project
```

```bash
# List projects for current user
gh project list

# List projects for an org
gh project list --owner org-name

# List projects including closed
gh project list --closed

# View a project
gh project view 1
gh project view 1 --owner monalisa --web
gh project view 1 --format json

# Create a project
gh project create --owner @me --title "Roadmap"

# Edit a project
gh project edit 1 --title "New Title" --description "Updated description"

# Copy a project
gh project copy 1 --title "Q4 Copy"

# Close / delete a project
gh project close 1
gh project delete 1

# Mark as template / unlink
gh project mark-template 1
gh project unlink 1 --repo owner/repo

# Link project to repo or team
gh project link 1 --repo owner/repo
gh project link 1 --team my-team

# List fields in a project
gh project field-list 1 --owner @me
gh project field-list 1 --owner @me --format json

# Create / delete a custom field
gh project field-create 1 --name "Sprint" --type text
gh project field-delete 1 --field-id <field-id>

# List items in a project
gh project item-list 1 --owner @me
gh project item-list 1 --owner @me --query "assignee:monalisa -status:Done"
gh project item-list 1 --limit 100

# Add an issue or PR to a project
gh project item-add 1 --url https://github.com/owner/repo/issues/42

# Create a draft issue in a project
gh project item-create 1 --title "Draft task" --body "Notes..."

# Edit an item in a project
gh project item-edit 1 --item-id <item-id> --field "Status" --value "In Progress"

# Archive / delete an item
gh project item-archive 1 --item-id <item-id>
gh project item-delete 1 --item-id <item-id>
```
