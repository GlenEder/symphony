---
name: grilling
description: Interview the user relentlessly to stress-test a plan or resolve open decisions before planning. Use when the user wants their thinking challenged, has unresolved decisions to settle before a plan is authored, or mentions grilling.
compatibility: opencode
---

Interview me relentlessly about every aspect of this until we reach a shared understanding.
Walk down each branch of the decision tree, resolving dependencies between decisions one-by-one.
For each question, provide your recommended answer, with the recommended option listed first and marked.

Ask the questions one at a time, waiting for feedback on each question before continuing.

If a *fact* can be found by exploring the environment (filesystem, tools, code), look it up rather than asking me.
The *decisions*, though, are mine — put each one to me and wait for my answer.

Do not act on it until I confirm we have reached a shared understanding.

Follow the terminal interview flow described below for how to interact with me.

## When to Skip Grilling

Do not grill when:
- There are fewer than 2 decisions to make (trivial changes)
- The user explicitly declines ("no grilling", "just do it", "skip questions")
- All open questions are factual (can be answered by research), not decisional

When in doubt, start grilling — it is cheaper to skip one question than to miss a decision.

## Terminal Interview Flow

### 1. Research facts first

Before asking anything, explore the environment: the filesystem, the available tools, and the code itself.
Never ask the user about something that can be looked up.
Note what you found so each question can cite concrete evidence instead of asking for it.

### 2. Enumerate the open decisions

List the decisions that only the user can make: scope, trade-offs, architecture, priorities, risk tolerance.
Order them by dependency so that earlier answers unlock, retire, or reshape later questions.

### 3. Ask via the interactive question tool

Ask the questions one at a time using the interactive question tool, and wait for the answer before asking the next one.
For each question:
- Offer discrete clickable options whenever the decision space has enumerable choices.
- Put the recommended option first and mark it as the recommended choice.
- Allow a custom free-text answer for choices outside the listed options.
- Keep the question self-contained: state the context, the options, and the recommendation.

### 4. Record each resolved decision

As each question is answered, record the decision, the alternatives considered, and the rationale.
Do not re-ask branches an answer has closed; do append any new branches it opens to the queue.

### 5. Confirm shared understanding

Before finishing, ask: **"Do you feel all branches are exhausted? Do we have a shared understanding?"**
Wait for explicit confirmation before proceeding to any action.

## Plain-Chat Fallback

If no interactive question tool is available, run the same interview in plain chat.
Ask one question at a time, list the options as plain text, and still mark the recommended choice.
Everything else is unchanged: research facts first, wait for each answer, and confirm a shared understanding before finishing.

## Output Contract

Each resolved question becomes a `decision` module entry handed off to plan authoring:
- `text` — the decision, stated plainly.
- `options` — the alternatives considered, with the chosen one listed first or explicitly named.
- `rationale` — why this choice was made.

The maestro skill renders these entries when it authors the plan.
maestro-export turns them into the work ticket's Key Decisions section.
Grilling itself writes no plan file — it produces resolved decisions and nothing else.
