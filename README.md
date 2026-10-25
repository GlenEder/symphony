# Symphony — Coding Agent Skill Suite

A collection of skills, prompts, and configurations designed for the [Maki](https://github.com/gleneder/maki) coding agent. These skills encode reusable workflows and agent instructions that Maki loads at runtime to perform complex software engineering tasks.

While tailored for Maki, the individual skills are plain markdown and shell files — they should transpose to other coding agent systems (e.g., Claude Code, Cline, Aider) with minimal adaptation.

The skills and systems in this repo are heavily influenced by outside sources like:
- [Matt Pocock's Skills](https://github.com/mattpocock/skills)
- [Lavish-Axi](https://github.com/kunchenguid/lavish-axi)


## What's Included

- **Skills** (`.config/maki/skills/`) — Markdown instruction files that teach the agent specialized workflows, such as:
  - `gh` — GitHub CLI operations (PRs, issues, releases, Actions)
  - `maki-agent` — Maki configuration and usage reference
  - `plan-implementation-procedure` — Multi-stage plan orchestration: works through all not-done stage tickets (per-stage implementer↔reviewer loop, max 3 iterations, per-stage commits on a plan branch), halting on flagged stages, then one draft PR via publish-it
  - `publish-it` — Quick one-shot PR publishing
  - `create-bash-script` — Bash script scaffolding
  - `research` — Delegate reading legwork to a background subagent against primary sources
  - `grilling` — Interview the user relentlessly to stress-test a plan and resolve open decisions before plan authoring
- **Setup script** — Symlinks everything into the correct Maki config directory

## Prerequisites

- [Maki](https://github.com/gleneder/maki) coding agent installed and configured
- Bash

## Setup

Run the setup script from the repo root:

```bash
./setup
```

This will symlink all skills and config files into `~/.config/maki/` (or `$XDG_CONFIG_HOME/maki/` if set).

To preview what would be linked without making changes:

```bash
./setup --dry-run
```

The setup is safe to re-run — it updates symlinks incrementally and skips any sources that don't exist.

## Project Structure

```
symphony/
├── setup             # Installation dispatcher (symlinks into ~/.config/maki/)
├── setup_cmds/       # Setup subcommand scripts
├── AGENTS.md         # Global agent instructions
├── maestro/          # Legacy planning server (pending deprecation)
├── skills/           # Skill markdown files (subdirectories)
├── openwiki/         # Repository documentation
└── __tests__/        # Test suites
```

## Usage

Once setup is complete, skills are available to the Maki agent on demand. For example, when working with GitHub, the agent can load the `gh` skill for guided PR/release workflows.

## Transposing to Other Agents

Each skill is a standalone markdown file with a consistent structure (purpose, instructions, examples, gotchas). To adapt for another agent system:

1. Port the markdown skills to your agent's instruction format (many agents accept markdown prompts directly)
2. Port shell-based commands as needed
3. The `setup` script is Maki-specific; for other agents, simply copy the skill files to the equivalent config directory

## License

MIT
