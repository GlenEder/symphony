# Skills — Agent Instruction Files

Skills are Markdown instruction files (with optional shell scripts) that teach coding agents specialized workflows. They are the primary way Symphony extends an agent's capabilities.

## Structure

Each skill lives in a subdirectory under `skills/` and contains:

```
skills/<name>/
├── SKILL.md        # Agent instructions (required)
├── scripts/        # Optional shell scripts the skill references
└── ...             # Other supporting files
```

The `SKILL.md` file has a frontmatter header:

```yaml
---
name: <skill-name>
description: <one-line description>
compatibility: opencode  # or other agent systems
---
```

## Skills Included

| Skill | Directory | Purpose |
|-------|-----------|---------|
| **maestro** | `skills/maestro/` | Thin router for Maestro authoring and feedback skills |
| **maestro-author** | `skills/maestro-author/` | Compose new plan JSON and decompose implementation stages |
| **maestro-session** | `skills/maestro-session/` | Serve plans, run live feedback, approve, and hand off export |
| **toon** | `skills/toon/` | Token-Oriented Object Notation — encode, decode, and validate TOON data |
| **gh** | `skills/gh/` | GitHub CLI operations — PRs, issues, releases, Actions |
| **maki-agent** | `skills/maki-agent/` | Maki agent configuration and usage reference |
| **plan-implementation-procedure** | `skills/plan-implementation-procedure/` | Full plan-to-PR workflow |
| **pm** | `skills/pm/` | Project management ticket generation |
| **publish-it** | `skills/publish-it/` | Quick one-shot PR publishing |
| **create-bash-script** | `skills/create-bash-script/` | Bash script scaffolding |
| **research** | `skills/research/` | Delegate reading legwork to a background subagent against primary sources |

## Installation

The `setup.sh` script at the repo root symlinks each skill directory into the agent's config directory:

```
~/.config/maki/skills/<name>/ → .../symphony/skills/<name>/
```

Because they are symlinks, edits to the repo are immediately available to the agent.

## How Skills Work

When an agent loads a skill (e.g., via a "use the maestro-session skill" instruction), it reads the `SKILL.md` file which provides:

1. **Purpose** — What the skill is for and when to use it
2. **Quick Start** — Minimal working example
3. **Reference** — Commands, APIs, formats, and patterns
4. **Examples** — Common use cases with complete workflows
5. **Gotchas** — Known pitfalls and edge cases

The skills are designed to be self-contained — an agent should be able to follow a skill without additional context.

## The Maestro Skills

Maestro planning is split by mutually exclusive trigger:

- `maestro-author/SKILL.md` handles new plan JSON and stage decomposition.
- `maestro-session/SKILL.md` handles serving, live feedback, approval, and export handoff.
- `maestro/SKILL.md` routes ambiguous Maestro requests without duplicating workflow prose.

Shared API, module glossary, and server lifecycle details are bundled under `skills/maestro/references/`.
The session helpers remain under `skills/maestro/scripts/`:
- `scripts/maestro-discover.sh` — find an existing server
- `scripts/maestro-heartbeat.sh` — background heartbeat for integrations
- `scripts/maestro-listen.sh` — watch plan files and output JSON

See the [Maestro section](../maestro/README.md) for the server's capabilities, and the [Operations section](../operations/README.md) for heartbeat/listen script usage.

## The TOON Skill

The `skills/toon/SKILL.md` (15KB) teaches agents the TOON format:

- Syntax: objects, primitive arrays, tabular arrays, mixed arrays
- Quoting rules and escape sequences
- Key folding for deeply nested data
- CLI usage (`npx @toon-format/cli`)
- Token efficiency strategies
- Streaming large outputs

See the [TOON Format section](../toon/README.md) for a format overview.

## Other Skills

The remaining skills (`gh`, `maki-agent`, `plan-implementation-procedure`, `pm`, `publish-it`, `create-bash-script`) are standard Markdown instruction files for agent workflows. They do not have associated server-side code or scripts beyond what they describe.

## Global Agent Instructions

The `AGENTS.md` file at the repo root (alwaysApply: true) provides baseline instructions for all agent sessions:

- Line-per-sentence Markdown formatting
- Plan-before-implement workflow
- Bug reproduction discipline
- Pixel-perfection UI standards
- **Plan Display** — requires agents to use the Maestro format for substantive plans, serving them via the Maestro web UI for structured feedback

## Important Notes

- Skills are loaded by the Maki agent at startup when symlinked to `~/.config/maki/skills/`
- Other agent systems may require different paths or naming conventions
- The `setup.sh` script only handles symlinks — it does not validate skills or check for conflicts
- If a skill references a script, the agent must know the path to that script (the skill file handles this with relative paths)
