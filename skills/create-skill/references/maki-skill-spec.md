# Maki Skill Format Specification

## Overview

Maki skills are reusable instruction sets stored as Markdown files with YAML frontmatter. The agent loads a skill on demand via the `skill()` tool, injecting the entire file content into the system prompt.

## Directory

```
~/.config/maki/skills/<skill-name>/SKILL.md
```

Skills are symlinked from the source repository:

```
~/.config/maki/skills/<skill-name>/  →  skills/<skill-name>/
```

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Skill name used with `skill()` tool. Kebab-case, 1-64 chars, lowercase, hyphens only, no consecutive hyphens. Must match parent directory name. |
| `description` | Yes | Shown in available skills list. Max 1024 chars. Should include positive and negative triggers for agent routing. |
| `compatibility` | Yes | Set to `opencode` (the only supported value). |

### Frontmatter Example

```markdown
---
name: my-skill
description: Does X and Y for Z. Use when the user wants to... Not for...
compatibility: opencode
allowed-tools: bash,read,write
license: MIT
---
```

## Agent Loading Behavior

- When `skill("<name>")` is called, the entire SKILL.md content is injected into the agent's system prompt
- The agent sees only the `name` and `description` from frontmatter when listing available skills
- The agent decides whether to load a skill based entirely on the frontmatter description
- There is no mechanism for conditional or partial loading — the full file is always injected

## Best Practices

- SKILL.md should be under 500 lines (the mgechev guideline)
- Use progressive disclosure: keep high-level logic in SKILL.md, offload details to `references/`, `scripts/`, `assets/`
- Reference files explicitly with JiT (Just-in-Time) loading instructions: "See `references/auth-flow.md` for error codes"
- Use relative paths with forward slashes regardless of OS
- No README.md, CHANGELOG.md, or other documentation files in skill directories
- No library code — scripts should be tiny, single-purpose executables

## Known Limitations

- No skill dependency mechanism (cannot declare that one skill requires another)
- No versioning or migration support
- No access control or permission scoping per skill
- No built-in validation or testing framework
ts/`, `assets/`
- Reference files explicitly with JiT (Just-in-Time) loading instructions: "See `references/auth-flow.md` for error codes"
- Use relative paths with forward slashes regardless of OS
- No README.md, CHANGELOG.md, or other documentation files in skill directories
- No library code — scripts should be tiny, single-purpose executables

## Known Limitations

- No skill dependency mechanism (cannot declare that one skill requires another)
- No versioning or migration support
- No access control or permission scoping per skill
- No built-in validation or testing framework
