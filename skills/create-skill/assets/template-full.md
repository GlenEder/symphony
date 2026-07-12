---
name: full-skill
description: "Does X and Y for Z. Use when the user wants to build Z, configure Y, or debug X. Not for A, B, or C."
compatibility: opencode
allowed-tools: bash,read,write     # Optional: restrict tools
license: MIT                        # Optional: SPDX identifier
---

## Purpose

Describe what this skill does, when to use it, and when NOT to use it.

## Workflow

### Step 1: Discovery

1. Ask the user clarifying questions if needed
2. Determine the scope of work
3. Confirm any assumptions

### Step 2: Preparation

1. Read any required reference files
2. Check prerequisites
3. Set up any needed state

### Step 3: Execution

1. Execute the primary operation
2. If the operation fails:
   - Check error output
   - Apply the relevant recovery step from Edge Cases
   - Retry up to 3 times

### Step 4: Validation

1. Verify the output matches expectations
2. Run any self-check steps
3. Report results to the user

## Scripts

### `scripts/script-name.sh`

Purpose of this script and when to use it.

```bash
scripts/script-name.sh --flag value
```

## References

- `references/file.md` — What this file contains and when the agent should read it.

## Assets

- `assets/template.yaml` — What this template is for and when to copy/fill it.

## Examples

- `examples/basic-usage/` — Example showing basic skill invocation and expected output.

## Edge Cases

| Scenario | Handling |
|----------|----------|
| **Failure mode 1** | Recovery step |
| **Failure mode 2** | Recovery step |
| **Edge case 1** | Handling approach |
