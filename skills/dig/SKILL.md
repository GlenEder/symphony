---
name: dig
description: Kick off a specialized debugging subagent that digs into a bug or error the user presents and reports back the root cause (RCA) with evidence. Use when the user presents a bug, stack trace, or error and wants the root cause found, or says "dig", "find the root cause", "why is this broken", or "root cause this". The digger never makes the fix.
compatibility: opencode
---

## Purpose

Dispatch a **strong-tier digger subagent** that produces an evidence-backed root cause analysis (RCA).
The digger investigates and reports — it never fixes the bug.
Your job is to scope the case, dispatch the digger, and hand the finished RCA back to the user for fix-planning.

## Dispatch workflow

### 1. Frame the case

State the bug in 1-2 sentences, using only what the user presented.
If essential facts are missing — how it was triggered, the environment, an artifact — ask at most 3 short questions before dispatching.
Do not ask more than 3; dispatch with what you have rather than stall on a fourth question.

Done when: you can state the bug and have a concrete artifact (error message, failing test, repro command) or an explicit note that no artifact exists.

### 2. Pick the save path

Save the RCA to one Markdown file:

- Inside a repo: `.research/bug-<slug>.md` at the repo root.
- Outside a repo: `.research/bug-<slug>.md` in the current working directory.

Create `.research/` if it does not exist; it is gitignored scratch, the same convention the `research` skill uses.
Name the file for the bug, kebab-case.

Done when: you have a concrete file path.

### 3. Dispatch the digger

Read `skills/dig/DIGGER.md` in full and inline its content into the subagent prompt.
Dispatch a **blocking**, **strong-tier**, **general** subagent — this is not a background dispatch; you wait for it to finish.
In Cursor: the Task tool with `run_in_background: false` (or omitted) and `subagent_type: generalPurpose`, on the strongest available model tier.
In Maki: a foreground agent, not a background agent, on the strong-tier model.
Hand it:

- The case: the bug as presented, plus any artifact and answers from step 1.
- The save path from step 2.
- The full `DIGGER.md` content, inlined.
- The hard rules below, restated explicitly.

Done when: the subagent is running with the case, the methodology, and the hard rules.

### 4. Report back

When the digger returns, read the saved RCA file and present it **in full** in chat:

- The save path.
- The verdict, with its confidence label repeated.
- The root cause.
- What was ruled out.
- The fix direction.

Then stop.
Do not implement the fix.
Do not chain into the planning flow — maestro is a separate, explicitly invoked stage.

Done when: the RCA has been presented in chat and you have stopped without editing code or opening a plan.

## Hard rules

- The digger never modifies code.
  Running code — builds, tests, repros — is investigation, not a fix.
- No commits, no branch creation, no installs.
- Scratch instrumentation (temp log lines, temp test files) is allowed during investigation, but every scratch artifact must be deleted before the digger finishes, leaving the tree as found.
- Every claim in the RCA cites evidence: file:line, a command plus its output, or a log line.
- The verdict carries a confidence label — confirmed, likely, or speculative — and the chat presentation repeats that label.

## Full methodology

The complete digger method, confidence ladder, and RCA document contract live in `skills/dig/DIGGER.md`.
Read it only at dispatch time, so this file's context load stays light.
