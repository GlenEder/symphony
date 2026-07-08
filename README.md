# Symphony — Coding Agent Skill Suite

A collection of skills, prompts, and configurations designed for the [Maki](https://github.com/gleneder/maki) coding agent. These skills encode reusable workflows and agent instructions that Maki loads at runtime to perform complex software engineering tasks.

While tailored for Maki, the individual skills are plain markdown and shell files — they should transpose to other coding agent systems (e.g., Claude Code, Cline, Aider) with minimal adaptation.

## What's Included

- **Skills** (`.config/maki/skills/`) — Markdown instruction files that teach the agent specialized workflows, such as:
  - `gh` — GitHub CLI operations (PRs, issues, releases, Actions)
  - `maki-agent` — Maki configuration and usage reference
  - `plan-implementation-procedure` — Full plan-to-PR workflow
  - `pm` — Project management ticket generation
  - `publish-it` — Quick one-shot PR publishing
  - `create-bash-script` — Bash script scaffolding
- **Setup script** — Symlinks everything into the correct Maki config directory

## Prerequisites

- [Maki](https://github.com/gleneder/maki) coding agent installed and configured
- Bash
- [fswatch](https://emcrisostomo.github.io/fswatch/) (recommended, not required) — enables zero-token-wait feedback sessions in the Maestro skill. Falls back to `stat` polling if unavailable.

## Setup

Run the setup script from the repo root:

```bash
./setup.sh
```

This will symlink all skills and config files into `~/.config/maki/` (or `$XDG_CONFIG_HOME/maki/` if set).

To preview what would be linked without making changes:

```bash
./setup.sh --dry-run
```

The setup is safe to re-run — it updates symlinks incrementally and skips any sources that don't exist.

## Project Structure

```
symphony/
├── setup.sh          # Installation script (symlinks into ~/.config/maki/)
├── AGENTS.md         # Global agent instructions (optional)
├── init.lua          # Maki init configuration (optional)
├── skills/           # Skill markdown files (subdirectories)
├── commands/         # Custom commands (optional)
├── providers/        # Provider scripts (optional)
└── .plans/           # Planning documents
```

## Usage

Once setup is complete, skills are available to the Maki agent on demand. For example, when working with GitHub, the agent can load the `gh` skill for guided PR/release workflows.

## Transposing to Other Agents

Each skill is a standalone markdown file with a consistent structure (purpose, instructions, examples, gotchas). To adapt for another agent system:

1. Port the markdown skills to your agent's instruction format (many agents accept markdown prompts directly)
2. Port shell-based commands as needed
3. The `setup.sh` script is Maki-specific; for other agents, simply copy the skill files to the equivalent config directory

## License

MIT
