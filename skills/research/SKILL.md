---
name: research
description: Research one well-scoped question against primary sources by dispatching a background subagent, so you keep working while it reads. Use when the user wants a topic researched, docs or API facts gathered, a plan item looked up, or reading legwork delegated so the main agent can continue.
compatibility: opencode
---

## Purpose

Delegate reading **legwork** to a **background subagent** so you keep working on the plan instead of blocking on research.
The subagent investigates one question against **primary sources**, writes a cited Markdown file under `.research/`, and saves it.
Dispatch, state the save path, and continue.

## Dispatch workflow

### 1. Scope the question

Narrow the item to **one tight question**.
If it is broad, split it and dispatch one subagent per part.
A vague question buys vague research; **scope** is the lever that buys focus.

Done when: the question is a single sentence a subagent could answer by reading sources, not a topic area.

### 2. Pick the save path

Save findings to one Markdown file:

- Inside a repo: `.research/<topic>.md` at the repo root.
- Outside a repo: `.research/<topic>.md` in the current working directory.

Create `.research/` if it does not exist; it is gitignored scratch, never committed.
Name the file for the topic, kebab-case.

Done when: you have a concrete file path to hand the subagent.

### 3. Dispatch a background subagent

Spin up a **non-blocking** subagent — in Cursor, the Task tool with `run_in_background: true` and `subagent_type: generalPurpose`; in Maki, a background agent.
Run it at **medium tier**: research verifies facts against primary sources, and a cheap model's misreadings become plan defects downstream.
Hand it the question, the save path, and the primary-sources requirement below.
Tell it to write findings to that path and cite each claim.

Done when: a background subagent is running with the question, the save path, and the primary-sources requirement — and you have moved on to other work.

### 4. Continue

Return to the plan.
State the save path so it can be retrieved later.
You will be notified when the subagent finishes; read the findings then, or when you next need them — whichever comes first.

Done when: you have resumed other plan work and the save path is recorded.

## The dispatch prompt

Give the subagent:

- The **question**, scoped to one sentence.
- The **save path** (`.research/<topic>.md`).
- The **primary-sources requirement**: investigate against primary sources — official docs, source code, specs, first-party APIs — the things that own the claim.
  Follow every claim back to the source that owns it.
- The **citation rule**: each claim in the findings cites its source.
- The **output**: one Markdown file at the save path, with a short summary and cited findings.

## Primary sources

Primary sources are the things that own the claim:

- Official documentation — vendor docs, language specs, RFCs.
- Source code and tests in the repo or the library.
- First-party APIs and their responses.
- Error messages and runtime behavior you reproduce.

A blog post about a library is secondary; the library's docs and source are primary.
An AI summary is secondary; the spec it summarizes is primary.
When a source is itself summarizing, follow it back to what it summarizes.

## Save convention

`.research/` is gitignored scratch — research notes are never committed.
One file per question, named by topic.
The skill states the save path on dispatch; reading findings back is your job.
