---
name: grilling
description: Grill the user relentlessly about a plan, decision, or idea. Use when the user wants to stress-test their thinking or mentions grilling.
compatibility: opencode
---

Interview me relentlessly about every aspect of this until we reach a shared understanding. Walk down each branch of the decision tree, resolving dependencies between decisions one-by-one. For each question, provide your recommended answer. Always include a `recommended` index (1-based) in every prompt to highlight your recommended choice in the wizard UI.

Ask the questions one at a time, waiting for feedback on each question before continuing.

If a *fact* can be found by exploring the environment (filesystem, tools, etc.), look it up rather than asking me. The *decisions*, though, are mine — put each one to me and wait for my answer.

Do not act on it until I confirm we have reached a shared understanding.

Follow the wizard flow described below for how to interact with me.

## When to Skip Grilling

Do not grill when:
- There are fewer than 2 decisions to make (trivial changes)
- The user explicitly declines ("no grilling", "just do it", "skip questions")
- All open questions are factual (can be answered by research), not decisional

When in doubt, start grilling — it is cheaper to skip one question than to miss a decision. 

## Grilling Wizard Flow


### 1. Start the Maestro server

Ensure the Maestro server is running. Start it if needed.
The server should be available at `http://localhost:8080`.

### 2. Create a plan via API

Create a new plan by POSTing JSON to `/api/plans`. Pick a descriptive kebab-case ID.

POST JSON to `/api/plans`:

```json
{
  "id": "{session-id}",
  "title": "Grilling Session — {topic}",
  "summary": "Interactive grilling session about {topic}"
}
```

The server creates the plan and returns it as JSON. No file writes needed.

### 3. Open the wizard page

Open the browser at `http://localhost:8080/grill/{id}`.
The wizard page will show "Waiting for the agent to ask the first question…" until the first prompt arrives.

### 4. Post questions as messages with prompt payload

For each question, POST a message to `/api/plan/{id}/messages` with an agent role and a `prompt` payload:

```json
{
  "role": "agent",
  "text": "Should we use a managed database service like AWS RDS or self-host PostgreSQL?",
  "prompt": {
    "question_key": "1",
    "options": ["AWS RDS", "Self-hosted PostgreSQL", "CockroachDB"],
    "allow_custom": true,
    "total_questions": 5,
    "answered": false,
    "recommended": 1
  }
}
```

**Prompt fields:**
- `question_key` (string, required) — unique identifier for this question (e.g. "1", "db-choice")
- `options` ([]string, required) — clickable choices displayed as buttons
- `allow_custom` (bool, required) — show a free-text "Other" field below options
- `total_questions` (int, optional) — display progress badge like "Q 1/5"
- `answered` (bool, **ALWAYS set to `false`** when posting a new question)
- `answer` (string, omit when posting, only set when marking as answered)
- `recommended` (int, required) — 1-based index into `options` of the agent's recommended choice. 0 means no recommendation (should not be used per skill mandate).

The wizard will automatically detect unanswered prompts (scanning newest-first) and render the current question as a centered card with clickable option buttons.

### 5. Wait for human response (feedback loop)

After posting a question, enter the standard Maestro feedback loop:
- Send heartbeats to `POST /api/agent/{id}/heartbeat` periodically (every 30s)
- Poll the plan state or listen for WebSocket updates
- The wizard page handles message posting — when the user clicks an option or submits custom text, it POSTs `{role: "human", text: "selected option"}`

When a human response arrives (check the plan's messages for a new `human` message after your question), proceed to the next step.

### 6. Mark previous prompt as answered and post next question

When you detect the human has responded to your previous question:

**First**, update the previous prompt by posting an agent message that re-sends the previous prompt but with `answered: true` and `answer` set to what the user chose:

```json
{
  "role": "agent",
  "text": "previous question text",
  "prompt": {
    "question_key": "1",
    "options": ["AWS RDS", "Self-hosted PostgreSQL", "CockroachDB"],
    "allow_custom": true,
    "total_questions": 5,
    "answered": true,
    "answer": "AWS RDS"
  }
}
```

**Note**: You can also update the existing message by directly modifying the plan file and triggering a reload (POST `/api/admin/reload`), or by using the WebSocket broadcast approach.

**Then**, post the next question as a new agent message with a fresh prompt (answered: false).

### 7. Confirm shared understanding

Before finalizing, ask the user: **"Do you feel all branches are exhausted? Do we have a shared understanding?"**
Wait for explicit confirmation before proceeding.

### 8. When all questions resolved

When the grilling session is complete and all decisions have been made:
1. Write the updated plan file (JSON) with populated modules:
   - `decision` modules — each resolved question becomes a decision (text, options, rationale)
   - `steps` — implementation steps derived from the decisions
   - `risks` — risks identified during the Q&A
   - `criteria` — acceptance criteria based on what was decided
2. Optionally, update the `questions` module items to mark things as resolved
3. Set the plan state to `draft` via `POST /api/plan/{id}/state` (do NOT approve)
4. Force reload via `POST /api/admin/reload` so the server picks up the module changes

The wizard page detects the changed state and redirects to `/plan/{id}` so the user can review the generated decision modules.

### 9. User reviews at /plan/{id}

The user reviews the generated plan with decision modules, discussion history, and structured plan content.
The user approves when satisfied by clicking "Approve Plan" on the plan page.

For a quick-reference summary of all API calls used in this flow, see [API_REFERENCE.md](API_REFERENCE.md).
