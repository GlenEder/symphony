## Copilot

> **Status**: GitHub Copilot CLI is currently in preview and subject to change.

```bash
# Run Copilot interactively
gh copilot

# Run with a prompt (non-interactive)
gh copilot -p "Summarize this week's commits"

# Allow all tools automatically (for scripting)
gh copilot -p "Fix the bug in main.js" --allow-all-tools

# Enable all permissions (tools, paths, URLs)
gh copilot -p "Fix the bug in main.js" --allow-all
gh copilot -p "Fix the bug in main.js" --yolo

# Use a specific model
gh copilot --model gpt-5.2

# Resume the most recent session
gh copilot --continue

# Resume a specific session
gh copilot --resume=<session-id>

# Grant specific tool permissions
gh copilot -p "Update package.json" --allow-tool='write'

# Grant shell with restrictions
gh copilot --allow-tool='shell(git:*)' --deny-tool='shell(git push)'

# Allow access to directories outside cwd
gh copilot --add-dir /home/user/projects

# Configure additional MCP servers
gh copilot --additional-mcp-config '{"my-server":{"command":"node","args":["server.js"]}}'

# Share session to markdown file on completion
gh copilot -p "Refactor auth" --share

# Share session to secret gist
gh copilot -p "Refactor auth" --share-gist

# Initialize Copilot instructions for a repo
gh copilot init

# Remove the Copilot CLI (if installed via gh)
gh copilot --remove

# Pass through flags to Copilot (use -- before them)
gh copilot -- --help
```

**Key flags reference**:

| Flag | Purpose |
|------|---------|
| `-p`, `--prompt` | Execute a prompt in non-interactive mode |
| `--model` | Set the AI model (`gpt-5.2`, `claude-4`, etc.) |
| `--allow-all`, `--yolo` | Allow all tools, paths, and URLs |
| `--allow-all-tools` | Auto-approve all tool calls (for CI/automation) |
| `--allow-tool` | Allow specific tool(s) without prompting |
| `--deny-tool` | Deny specific tool(s) |
| `--add-dir` | Add directory to allowed file access paths |
| `--resume` | Resume a previous session |
| `--continue` | Resume the most recent session |
| `--share` / `--share-gist` | Export session to file/gist on completion |
| `--additional-mcp-config` | Configure extra MCP servers for the session |
| `--reasoning-effort` | Set reasoning level (`low`, `medium`, `high`, `xhigh`) |
| `--silent` | Output only the agent response (for scripting) |
| `--remove` | Remove the Copilot CLI downloaded by gh |
