# Validation Guide

Full 4-step validation workflow for skill testing, adapted from [mgechev/skills-best-practices](https://github.com/mgechev/skills-best-practices).

## Overview

Validate any skill systematically before declaring it complete. Run these four steps in order.

---

## Step 1: Discovery Validation

**Goal**: Confirm the skill triggers correctly from its description.

### Procedure

1. Write 3 test prompts that SHOULD trigger the skill
2. Write 2 test prompts that SHOULD NOT trigger the skill (close neighbors)
3. Load the skill in the agent with each prompt
4. Verify the agent loads the skill for the positive cases and doesn't for the negative

### Failure modes

| Symptom | Fix |
|---------|-----|
| Skill triggers on wrong prompts | Add negative triggers to description |
| Skill doesn't trigger on right prompts | Add missing positive triggers |
| Description too vague | Follow the pattern: `<capability>. Use when <positive>. Not for <negative>.` |

### Template

```markdown
## Discovery Validation

### Should trigger
1. "(prompt 1)" → Agent loaded skill? [Yes/No]
2. "(prompt 2)" → Agent loaded skill? [Yes/No]
3. "(prompt 3)" → Agent loaded skill? [Yes/No]

### Should NOT trigger
1. "(prompt 4)" → Agent loaded skill? [Yes/No / Expected: No]
2. "(prompt 5)" → Agent loaded skill? [Yes/No / Expected: No]
```

---

## Step 2: Logic Validation

**Goal**: Confirm the agent follows the correct workflow when executing the skill.

### Procedure

1. Give the agent a concrete task the skill is designed to handle
2. Observe: does the agent follow the steps in order?
3. Check: does the agent apply decision trees correctly?
4. Check: does the agent reference JiT files at the right moment?
5. Check: does the agent use any scripts or templates correctly?

### What to look for

- Skipped steps (agent jumped ahead)
- Wrong order (agent did Step 3 before Step 1)
- Incorrect decisions (agent took the wrong branch in a decision tree)
- Missing JiT loads (agent should have read a reference file but didn't)
- Script misuse (wrong arguments, wrong script)

---

## Step 3: Edge Case Testing

**Goal**: Verify error handling and edge case recovery.

### Procedure

1. Feed the agent inputs from the Edge Cases section of the skill
2. Verify the agent picks the correct recovery path
3. Test boundaries: empty inputs, extreme values, missing dependencies

### Edge case test template

```markdown
| Edge case | Expected handling | Actual |
|-----------|-------------------|--------|
| (edge case 1) | (recovery step from SKILL.md) | Pass/Fail |
| (edge case 2) | (recovery step from SKILL.md) | Pass/Fail |
```

---

## Step 4: Architecture Refinement

**Goal**: Review the skill holistically for quality and maintainability.

### Questions to answer

1. **Level of detail**: Is the skill too prescriptive for experts? Too vague for beginners?
2. **Boundaries**: Is it clear what the skill does vs what it doesn't?
3. **Simplicity**: Could any part be simplified, merged, or removed?
4. **Redundancy**: Does any instruction duplicate the agent's built-in knowledge?
5. **Progressive disclosure**: Are bulky details offloaded to `references/`?
6. **Consistency**: Are terminology, style, and format consistent throughout?

### Scoring rubric

| Aspect | Score (1-5) | Notes |
|--------|-------------|-------|
| Appropriate detail | | |
| Clear boundaries | | |
| No redundancy | | |
| Progressive disclosure | | |
| Consistency | | |
| **Total** | **/25** | |
