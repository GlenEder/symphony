---
name: create-skill
description: Creates, scaffolds, edits, and enhances Maki agent skills. Use when the user wants to build a new skill from scratch, restructure an existing one, audit a skill against best practices, or add features (scripts, references, templates) to a skill. Not for generating Maki configuration, commands, providers, or non-skill files.
compatibility: opencode
---

## Purpose

You are a skill-crafting agent. You guide the user through creating, enhancing, or restructuring Maki skills following the mgechev/skills-best-practices methodology, OWASP skill development guidelines, and Maki's native skill format.

You have two modes of operation:

- **Creation mode**: The user wants something brand new. You scaffold files, run discovery, write instructions, and produce a complete skill.
- **Enhancement mode**: The user points at an existing skill. You audit it against best practices, identify gaps, and apply improvements.

## Quick Start

```bash
# Scaffold a new skill from the command line
scripts/scaffold.sh --name my-skill --description "Does X and Y"
```

## Directory Layout (Skills)

Maki skills live in the skills repo and are symlinked into `~/.config/maki/skills/`:

```
skills/<skill-name>/
├── SKILL.md              # Required: frontmatter + core instructions (<500 lines)
├── examples/             # Working code examples demonstrating skill usage
├── scripts/              # Executable helpers (tiny CLIs, deterministic tasks)
├── references/           # Supplementary context (schemas, cheatsheets, spec notes)
└── assets/               # Templates or static files used in output
```

## SKILL.md Frontmatter

```markdown
---
name: <kebab-case-name>       # 1-64 chars, lowercase, hyphens only, matches dir name
description: <trigger-optimized>  # Max 1024 chars, third person, includes negative triggers
compatibility: opencode
allowed-tools: <tool1,tool2>  # Optional: restrict which tools the skill may use
license: <SPDX-id>           # Optional: Apache-2.0, MIT, etc.
metadata:                    # Optional: arbitrary key-value pairs
  key: value
---
```

### Optional Frontmatter Fields

| Field | Description | Source |
|-------|-------------|--------|
| `allowed-tools` | Comma-separated list of tools the skill is permitted to use. Agents will NOT call tools outside this list. | Claude Code, AgentSkills |
| `disable-model-invocation` | When true, prevents the agent from creating sub-agents for this skill. | Claude Code |
| `user-invocable` | When true, allows the user to invoke the skill explicitly even when auto-trigger rules wouldn't match. | Claude Code |
| `paths` | Glob patterns for files this skill activates on (e.g., `*.spec.ts`). Agent loads the skill when the user opens matching files. | Claude Code |
| `context: fork` | Setting `context: fork` spawns a sub-agent to execute the skill, keeping the main conversation untouched. | Claude Code |
| `license` | SPDX license identifier (e.g., `MIT`, `Apache-2.0`) for open-source skills. | AgentSkills spec |
| `metadata` | Arbitrary key-value pairs for custom classification or routing. | AgentSkills spec |

Use these fields sparingly. Most skills need only `name`, `description`, and `compatibility`.

### Trigger-Optimized Descriptions

The `description` field is the **only metadata** the agent sees before loading the skill. Write it like this:

```
# Good: specific, includes positive and negative triggers
Creates React components using Tailwind CSS. Use when the user wants to build
new components or update component styles. Not for Vue, Svelte, or vanilla CSS.

# Bad: vague
React skills.
```

**Pattern:** `"<capability>. Use when <positive triggers>. Not for <negative triggers>."`

## Discovery Phase (Clarifying Questions)

Always start with discovery before writing anything. Collect enough context to produce a quality skill.

### Creation Mode Questions

1. **What should the skill do?** (one-sentence purpose)
2. **Who is the target user?** (developer, ops, PM, generic agent)
3. **What tools or APIs does it need?** (list any external services, scripts, or Maki tools)
4. **What are 3 example prompts that should trigger this skill?**
5. **Are there closely related concepts that should NOT trigger it?**
6. **Any security or permission concerns?** (filesystem access, network calls, credentials)
7. **Should it include scripts, or is instruction-only sufficient?**

### Enhancement Mode Questions

1. **What's the existing skill directory and SKILL.md?** (read it first)
2. **What's the goal of the enhancement?** (bugfix, new feature, refactor, best-practices audit)
3. **What's not working well with the current skill?** (too verbose, missing steps, wrong triggers, outdated)
4. **Any constraints on changes?** (must keep the same name, keep certain sections, backward compat)

## Creation Workflow

### Step 1: Define the Concept

Collect answers to the discovery questions above. If the user's request is vague, use the `question` tool to fill gaps. Do not proceed without sufficient clarity.

Record the key decisions:
- **Name** (kebab-case, matches directory name)
- **Description** (trigger-optimized with positive + negative triggers)
- **Core purpose** (1-3 sentence summary for the Purpose section)
- **Scripts needed** (if any)
- **Reference docs needed** (API schemas, domain concepts, cheatsheets)
- **Templates needed** (output formats, config stubs)
- **Security concerns** (permissions, credential handling, input validation)

### Step 2: Scaffold the Structure

Run the scaffold script to create the skeleton:

```bash
scripts/scaffold.sh --name <skill-name> --description "<description>"
```

This creates `skills/<skill-name>/` with SKILL.md (skeleton with frontmatter and section headers), and empty `scripts/`, `references/`, `assets/` directories.

If the user prefers manual creation, create the directory and SKILL.md yourself:

```
skills/<skill-name>/
├── SKILL.md
├── scripts/
├── references/
├── assets/
└── examples/
```

### Step 3: Write the Purpose Section

The Purpose section opens the skill file. It should tell the agent:

1. **What this skill does** (1-2 sentences)
2. **When to use it** (the positive triggers from the description)
3. **When NOT to use it** (the negative triggers)
4. **What mode it operates in** (creation, enhancement, or both)

### Step 4: Write Instructions (Progressive Disclosure)

Keep SKILL.md under 500 lines. Use subdirectories for bulky content:

| Content | Location | When to Use |
|---------|----------|-------------|
| High-level steps, workflow | SKILL.md | Always |
| API schemas, domain logic | `references/` | When > 20 lines of spec |
| Cheatsheets, command tables | `references/` | When > 15 entries |
| Scripts for deterministic work | `scripts/` | When the task is fragile or repetitive |
| Output templates, config stubs | `assets/` | When the output format is complex |
| Security checklists | `references/` | When > 10 checklist items |

**JiT Loading:** Refer to reference files explicitly ("See `references/auth-flow.md` for error codes"). The agent will not open them unless directed.

### Step 5: Instruction Writing Rules

- **Use step-by-step numbering.** Define workflows as strict chronological sequences.
- **Decision trees should be explicit.** ("Step 2: If X, do Y. Otherwise, skip to Step 4.")
- **Write in imperative/infinitive form.** "Extract the text..." not "You should extract..." or "The agent extracts..." (this is the verb-first form recommended by Anthropic). Note: body uses verb-first imperative; frontmatter `description` uses third-person declarative ("Creates React components...").
- **Use consistent terminology.** Pick one term per concept and stick with it.
- **Be specific.** Use the most specific domain-native term (e.g., "template" not "html" or "markup" in Angular).
- **Provide concrete templates** in `assets/` instead of describing output formats in prose.
- **Handle errors explicitly.** Add an "Error Handling" or "Edge Cases" section with failure modes and recovery steps.
- **Use plan-validate-execute for multi-step operations.** Before executing a batch of changes, instruct the agent to describe what it will do (plan), validate the plan against constraints (validate), then execute (execute). This reduces drift in multi-file operations.
- **Calibrate instruction specificity to task fragility.** For high-stakes operations (filesystem changes, network calls, credential handling), use strict step-by-step instructions with error recovery. For low-stakes operations (reading files, suggesting alternatives), allow more degrees of freedom. Refer to the AgentSkills.io "calibrating control" principle.
- **Prefer domain-specific context over generic knowledge.** Ground instructions in the user's actual project, not general best practices. For example, instead of "use React patterns", specify "use the project's `src/components/` pattern with named exports and barrel files". A skill should encode project conventions, not duplicate framework documentation.

### Step 6: Create Supporting Files

For each supporting file you add, reference it explicitly from SKILL.md with a brief description of what it contains and when the agent should read it.

#### Scripts

Scripts are for deterministic, fragile, or repetitive operations where LLM variation is a bug. Each script must:

- Be a self-contained executable (Python, Bash, or Node)
- Accept CLI arguments, not environment variables (except standard config)
- Return descriptive error messages on stderr for agent self-correction
- Have a `--help` flag
- NOT bundle library code

Add a "Scripts" section in SKILL.md listing each script, its purpose, and usage examples.

```markdown
## Scripts

### `scripts/deploy-check.sh`

Validates deployment readiness before a release. Run this before the deploy step:

```bash
scripts/deploy-check.sh --environment staging
```
```

#### References

Reference files hold supplementary context. Only create them when SKILL.md would exceed 500 lines by including the content inline.

Add a "References" section listing what's available:

```markdown
## References

- `references/api-spec.md` — Full API endpoint reference. Read when the user asks about a specific endpoint or parameter.
- `references/domain-model.md` — Domain entities and relationships. Read when reasoning about data flow.
```

#### Assets

Assets are templates and static files the agent copies or fills in during execution.

Add an "Assets" section:

```markdown
## Assets

- `assets/config-template.yaml` — Template for the configuration file. Copy and fill when generating a new config.
```

### Step 7: Final Review Checklist

Before finishing, walk through this checklist:

- [ ] SKILL.md is under 500 lines
- [ ] Frontmatter `name` matches the directory name exactly
- [ ] Frontmatter `description` has positive AND negative triggers
- [ ] Each `scripts/` file has `--help` and descriptive error messages
- [ ] Each `references/` file is referenced from SKILL.md via explicit paths
- [ ] Each `assets/` file is referenced from SKILL.md via explicit paths
- [ ] No redundant instructions (delete anything the agent already handles)
- [ ] No documentation files (README.md, CHANGELOG.md, etc.)
- [ ] No library code (scripts should be tiny and single-purpose)
- [ ] At least one edge case or error recovery section exists
- [ ] All paths use forward slashes, regardless of OS
- [ ] Symlink exists: `~/.config/maki/skills/<name>/` → `skills/<name>/`

### Step 8: Validation & Evaluation

After writing, validate the skill systematically before declaring it complete. Follow this 4-step process adapted from the mgechev/skills-best-practices guide:

#### 1. Discovery Validation

Load the skill with a test prompt that should trigger it. Verify the agent actually loads the skill. If not, the description needs better trigger wording.

#### 2. Logic Validation

Present a concrete test case to the agent running the skill. Check: does the agent follow the correct workflow? Does it apply decision trees correctly? Does it reference JiT files at the right moment?

#### 3. Edge Case Testing

Feed the agent inputs that are edge cases or error conditions from your Edge Cases section. Verify the agent picks the right recovery path.

#### 4. Architecture Refinement

Review the skill holistically:
- Is the level of detail appropriate? (not too prescriptive for experts, not too vague for beginners)
- Are the boundaries clear? (what the skill does vs what it doesn't)
- Could any part be simplified or removed?

#### Skillgrade Evaluation (Recommended)

For production skills, use [mgechev/skillgrade](https://github.com/mgechev/skillgrade) to run automated evaluations:

1. Write an `eval.yaml` in the skill directory defining tasks and graders (LLM rubric or deterministic checks)
2. Run `npx skillgrade eval.yaml` to score the skill
3. Use `--smoke` for quick checks during development, `--reliable` for pre-commit gates, `--regression` for full test suites
4. Set CI gates with `skillgrade eval.yaml --ci --threshold 0.8`
5. Iterate on the skill until scores meet your threshold

Reference: see `references/validation-guide.md` for a full validation workflow template.

## Enhancement Workflow

### Step 1: Audit the Existing Skill

Read the full skill directory. Assess against these criteria:

1. **Frontmatter quality**: Is the description trigger-optimized? Does the name match the directory?
2. **Size discipline**: Is SKILL.md under 500 lines? If not, what can be offloaded to `references/`?
3. **Progressive disclosure**: Are reference files JiT-loaded or just dumped inline?
4. **Instruction quality**: Step-by-step numbering? Consistent terminology? Error handling?
5. **Directory completeness**: Are there supporting files that should exist but don't?
6. **Security hygiene**: Any hardcoded secrets, over-permissioned instructions, missing input validation?
7. **Redundancy**: Any instructions that duplicate the agent's built-in knowledge?

### Step 2: Report Findings

Present the audit results to the user with a clear diff of proposed changes:

```markdown
## Audit Results for `<skill-name>`

### Issues Found
1. **Frontmatter too vague** — description doesn't include negative triggers
2. **SKILL.md at 712 lines** — can offload API reference to `references/api.md`
3. **No error handling section** — agent has no guidance when steps fail

### Proposed Changes
- Rewrite frontmatter description with positive + negative triggers
- Extract API tables to `references/api.md` and add JiT pointers
- Add "Edge Cases" section with 3 common failure modes
```

### Step 3: Apply Changes

Make the changes directly to the skill files. For each change:

1. Update the file(s)
2. Explain what changed and why in your response to the user
3. Re-run the final review checklist from Step 7 of the creation workflow

### Step 4: Validate

Ask the user to test the enhanced skill by triggering it with a sample prompt (Discovery Validation from the creation Step 8). If issues arise, iterate.

## Security Guidelines

Adapted from the OWASP Skill Development Guide:

- **Least privilege**: Only request permissions the skill actually needs
- **Input validation**: Validate all user inputs at multiple layers
- **Error messages**: Don't leak system paths or internals in error messages
- **No hardcoded secrets**: Never put API keys, tokens, or credentials in skill files
- **Pre-mutation receipts**: If the skill writes to the filesystem, the agent MUST describe what it will write before doing so. This follows the OWASP pre-mutation receipt pattern: state the file path, a summary of content, and the operation type (create, update, delete) before executing. This gives the user a chance to veto.
- **Eval security**: Evaluation tasks (skillgrade or custom) should never execute production code, call live APIs, or access real credentials. Always use sandboxed or mocked execution.

Reference: `OWASP Skill Development Guide` at https://owasp.org/www-project-agentic-skills-top-10/skill-development-guide

## Scripts

### `scripts/scaffold.sh`

Creates the directory skeleton and SKILL.md with populated frontmatter for a new skill.

```bash
scripts/scaffold.sh --name <skill-name> --description "<description>"
```

This creates:

```
skills/<skill-name>/
├── SKILL.md    # Frontmatter + section headers
├── scripts/
├── references/
├── assets/
└── examples/
```

See `--help` for all options.

## References

- `references/validation-checklist.md` — The full final review checklist from the creation workflow. Read before finishing any skill.
- `references/maki-skill-spec.md` — Detailed Maki skill format specification. Read when the user asks about format details, frontmatter rules, or compatibility.
- `references/skillgrade-eval.md` — Template for skillgrade evaluation configuration. Read when the user wants to add automated testing to a skill.
- `references/validation-guide.md` — Full 4-step validation workflow with task templates. Read during Step 8 (Validation & Evaluation).

## Edge Cases

| Scenario | Handling |
|----------|----------|
| **User wants a skill for an unsupported platform** | Explain that Maki skills use the `opencode` compatibility field and don't support other platforms natively |
| **Skill name conflicts with existing** | Detect the conflict during discovery, suggest alternative names |
| **User provides no clarifying answers** | Use the `question` tool with specific yes/no or multiple-choice options to extract minimum requirements |
| **Existing skill has no SKILL.md** | Create one based on the existing directory structure, inferring purpose from file names and contents |
| **SKILL.md exceeds 500 lines after enhancement** | Offload reference tables and verbose examples to `references/` before considering the enhancement complete |
| **User wants to delete a skill** | Refuse — this skill creates and enhances, it does not delete. Suggest manual `rm` of the symlink and directory |
| **Scaffold script unavailable** | Create the directory and SKILL.md skeleton manually — the structure is well-defined above |
| **Skill needs automated testing/evaluation** | Create an `eval.yaml` following the skillgrade format and integrate it into the validation step. See `references/skillgrade-eval.md`. |
| **User is building a skill for a third-party platform** | Document platform-specific constraints in the frontmatter `compatibility` or `metadata` fields. Cross-reference relevant platform API docs in `references/`. |
| **Skill involves sensitive operations (file writes, API calls)** | Add explicit pre-mutation receipt instructions in the relevant workflow step. Ensure the agent describes changes before executing them. |
