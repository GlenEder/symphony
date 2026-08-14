1: ---
2: name: maestro-author
3: description: Author a NEW Maestro plan JSON and decompose implementation work into sequential stages. Use when the user asks to create, draft, compose, or decompose a plan. NOT for serving an existing plan, live review, approval, or work-ticket export.
4: compatibility: opencode
5: ---
6: 
7: # Author a Maestro plan
8: 
9: Use this skill for the authoring phase only.
10: When the plan is ready for human review, hand it explicitly to [`maestro-session`](../maestro-session/SKILL.md).
11: 
12: ## Authoring workflow
13: 
14: 1. Read the repository instructions and relevant OpenWiki pages before making planning decisions.
15: 2. Resolve open decisions before finalizing the plan.
16: If decisions remain, use the [`grilling`](../grilling/SKILL.md) skill and ask one question at a time.
17: 3. Build a plan object with `title`, `summary`, `state: "draft"`, and typed `modules`.
18: 4. Validate the plan data model and stage rules below.
19: 5. Write the plan to `$MAESTRO_PLANS_DIR/<plan-id>.json`, or POST it to a running Maestro server when handing off immediately.
20: 6. Confirm the plan can be read back, then invoke `maestro-session` for serving and review.
21: 
22: Completion means the draft is valid, every implementation stage is independently verifiable, and the next skill has an explicit plan ID.
23: 
24: ## Plan data model
25: 
26: The required top-level fields are `title`, `summary`, `state`, and `modules`.
27: Set `state` to `draft` until the human approves the plan.
28: Each module has a `type`, `heading`, and `items` array.
29: Each item has required `text` and type-specific fields.
30: 
31: Read [`references/API.md`](../maestro/references/API.md) for the complete HTTP contract.
32: Read [`references/GLOSSARY.md`](../maestro/references/GLOSSARY.md) for module fields and examples.
Read [`references/server-lifecycle.md`](../maestro/references/server-lifecycle.md) when handing a plan to the HTTP server or checking server requirements.
33: 
34: ## Stage decomposition rules
35: 
36: Always create stages, including for trivial work (exactly one stage is the degenerate case).
37: Use one `criteria`, `steps`, and `risks` module group per stage.
38: Prefix each of those headings exactly `Stage N: <name>`.
39: Use contiguous positive stage numbers with no duplicates.
40: Keep `decision`, `assumptions`, and `notes` modules global and unprefixed.
41: Order stages by dependency so a stage depends only on earlier stages.
42: Make each stage a small coherent slice that can be implemented and verified independently.
43: 
44: ## Handoff
45: 
46: Pass the exact plan ID and server requirements to `maestro-session`.
47: The session skill owns server reuse/startup, browser opening, heartbeat polling, responses, approval, and export.
48: Use [`maestro-export`](../maestro-export/SKILL.md) only through that explicit approval handoff.
