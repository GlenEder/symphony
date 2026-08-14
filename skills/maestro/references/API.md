1: # Maestro API Reference
2: 
3: Disclosed reference for [`maestro`](SKILL.md).
4: All routes return JSON.
5: 
6: ## List Plans
7: 
8: ```
9: GET /api/plans
10: ```
11: 
12: Response:
13: 
14: ```json
15: [
16:   {"id": "demo", "title": "Database Migration Plan", "summary": "Migrate from PostgreSQL 12 to PostgreSQL 15..."}
17: ]
18: ```
19: 
20: ## Get Plan
21: 
22: ```
23: GET /api/plan/{id}
24: ```
25: 
26: Response is a flat JSON structure:
27: 
28: ```json
29: {
30:   "title": "Database Migration Plan",
31:   "summary": "...",
32:   "state": "draft",
33:   "messages": [],
34:   "modules": [
35:     {
36:       "type": "criteria",
37:       "heading": "Acceptance Criteria",
38:       "items": [
39:         {"text": "All existing data is preserved after migration"}
40:       ]
41:     }
42:   ]
43: }
44: ```
45: 
46: ## Add a Message
47: 
48: ```
49: POST /api/plan/{id}/messages
50: Content-Type: application/json
51: 
52: {"role": "human", "text": "Your feedback", "item_ref": "2:1"}
53: ```
54: 
55: - `role`: `"agent"` or `"human"`
56: - `text`: message body (required)
57: - `item_ref`: optional positional reference `"moduleIndex:itemIndex"` (e.g. `"2:1"` = module 2, item 1)
58: 
59: Returns the created message:
60: 
61: ```json
62: {
63:   "id": "msg_18bfc3e196bafae0",
64:   "role": "human",
65:   "text": "Your feedback",
66:   "item_ref": "2:1",
67:   "created_at": "2026-07-06T17:35:51Z"
68: }
69: ```
70: 
71: The message is appended to the plan's conversation thread and the plan is persisted.
72: 
73: ## Set Plan State
74: 
75: ```
76: POST /api/plan/{id}/state
77: Content-Type: application/json
78: 
79: {"state": "approved"}
80: ```
81: 
82: Valid states: `"draft"`, `"approved"`.
83: Returns the full updated flat JSON plan.
84: 
85: ## Set Agent Status
86: 
87: ```
88: POST /api/agent/{id}/status
89: Content-Type: application/json
90: 
91: {"status": "offline"}
92: ```
93: 
94: Used to set the agent dot to `offline` explicitly (e.g. when the plan is approved).
95: The `listening` and `thinking` states are driven automatically by the server based on message roles — see the feedback loop in [`SKILL.md`](SKILL.md).
96: 
97: ## Reload Plans (Admin)
98: 
99: Trigger an immediate full directory rescan.
100: Useful when plans are modified externally and you don't want to wait for the next poll cycle.
101: 
102: ```
103: POST /api/admin/reload
104: ```
105: 
106: Response:
107: 
108: ```json
109: {"status": "ok"}
110: ```
111: 
112: ## WebSocket (Live Updates)
113: 
114: ```
115: ws://host/ws/plan/{id}
116: ```
117: 
118: When the plan file is modified, the server sends the full flat JSON plan over the WebSocket.
119: The client can then reload or patch the view.
120: 
121: ## Web UI Routes
122: 
123: | Route | Description |
124: |---|---|
125: | `/` | Redirects to `/plans` |
126: | `/plans` | Plan listing page |
127: | `/plan/{id}` | Plan detail page with modules, sidebar, messages |
128: | `POST /api/admin/reload` | Trigger full directory rescan (JSON) |