# Skillgrade Evaluation Template

Use this template to set up automated evaluation for a skill using [mgechev/skillgrade](https://github.com/mgechev/skillgrade).

## File: `eval.yaml` (place in the skill root directory)

```yaml
# eval.yaml — Skillgrade evaluation configuration
name: <skill-name>
description: "<trigger-optimized description>"

tasks:
  - name: "Basic task"
    prompt: "A user prompt that should trigger this skill"
    graders:
      - name: "task-completion"
        type: llm
        rubric: |
          Did the agent complete the task correctly?
          Score 0-10:
          - 10: Perfect execution with all requirements met
          - 5: Partial execution, some steps missed
          - 0: Failed to complete or hallucinated

  - name: "Edge case"
    prompt: "An edge case input from the skill's Edge Cases section"
    graders:
      - name: "edge-handling"
        type: llm
        rubric: |
          Did the agent handle the edge case appropriately?

  - name: "Deterministic check"
    prompt: "A prompt producing deterministic output"
    graders:
      - name: "output-format"
        type: code
        code: |
          const result = agent.output;
          return result.includes('EXPECTED_VALUE') ? 1.0 : 0.0;
```

## Run commands

```bash
# Quick check during development
npx skillgrade eval.yaml --smoke

# Pre-commit gate
npx skillgrade eval.yaml --reliable

# Full test suite (CI)
skillgrade eval.yaml --ci --threshold 0.8

# Iterate: adjust rubric, fix skill, re-run until passing
```

## Grading strategies

| Grading Type | When to use |
|--------------|-------------|
| `type: llm` with rubric | Evaluating reasoning, completeness, edge case handling |
| `type: code` with assertions | Checking deterministic output (format, specific values, file structure) |
| Combined | Best coverage — separate tasks for different aspects |

## Best practices

- Keep eval tasks focused on one thing each (single-responsibility tasks)
- Include tasks for: happy path, each edge case, security concerns
- Grade outcomes, not steps — the rubric should score the result, not the process
- Test with different model tiers (weak, medium, strong) to find fragility
- Never use production data, live APIs, or real credentials in eval tasks
- Iterate on the rubric as you discover what "good" looks like
