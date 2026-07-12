# Validation Checklist

Use this checklist before finishing any skill (creation or enhancement).

## Frontmatter

- [ ] `name` is 1-64 characters, lowercase, hyphens only, no consecutive hyphens
- [ ] `name` matches the parent directory name exactly
- [ ] `description` includes both positive triggers ("Use when...") and negative triggers ("Not for...")
- [ ] `description` is under 1024 characters
- [ ] `compatibility` is set to `opencode`
- [ ] Optional fields (`allowed-tools`, `license`, `metadata`) are only present if actually needed
- [ ] `allowed-tools` doesn't list tools the skill doesn't use

## Size Discipline

- [ ] SKILL.md is under 500 lines
- [ ] Bulky content is offloaded to `references/` directory
- [ ] Reference files are exactly one level deep (no `references/subdir/file.md`)

## Instruction Quality

- [ ] Steps are numbered in strict chronological sequence
- [ ] Decision trees are explicit ("If X, do Y. Otherwise, skip to Z.")
- [ ] Instructions use imperative/infinitive form ("Extract the text...") — not "I will extract..." or "The agent extracts..."
- [ ] Terminology is consistent (one term per concept)
- [ ] Most specific domain-native terms are used
- [ ] No prose descriptions where templates would be clearer
- [ ] Plan-validate-execute pattern is used for multi-step operations
- [ ] Instruction specificity is calibrated to task fragility (stricter for high-stakes ops)
- [ ] Instructions encode domain/project-specific context, not generic knowledge

## Supporting Files

- [ ] `scripts/` only contains executable files (no library code)
- [ ] Each script has a `--help` flag and descriptive error messages
- [ ] Each `references/` file is explicitly referenced from SKILL.md
- [ ] Each `assets/` file is explicitly referenced from SKILL.md
- [ ] `examples/` files are referenced from SKILL.md where applicable
- [ ] All paths use forward slashes regardless of OS

## Redundancy Check

- [ ] No instructions that duplicate the agent's built-in knowledge
- [ ] No README.md, CHANGELOG.md, or other documentation files
- [ ] No library code — scripts are tiny and single-purpose

## Error Handling

- [ ] At least one "Edge Cases" or "Error Handling" section exists
- [ ] Common failure modes are documented with recovery steps
- [ ] Error messages don't leak system paths, internals, or secrets

## Security

- [ ] No hardcoded credentials, API keys, or tokens
- [ ] Input validation is described where user input is processed
- [ ] Permissions follow least privilege principle
- [ ] Sensitive data storage is NOT described as plaintext
- [ ] Pre-mutation receipt pattern is used for filesystem writes
- [ ] Eval tasks use sandboxed/mocked execution, not production code

## Setup
