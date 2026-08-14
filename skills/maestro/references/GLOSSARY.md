1: # Glossary — Maestro Module Types
2: 
3: The catalog of typed modules a Maestro plan is built from. Each entry shows the module's purpose, its fields, and a worked example in JSON. This is the disclosed reference for [`maestro`](SKILL.md); when authoring a module, reach for the type whose example matches what you want to express.
4: 
5: All module types share one required field — **text**, the primary description. Field names in **bold** below recur across types. For the full plan shape these modules sit inside, see `examples/demo.json` and `examples/regression-suite.json`.
6: 
7: ## criteria
8: 
9: Acceptance criteria — the checkbox list the plan must satisfy to be done. One criterion per item, phrased as a checkable outcome ("All existing data is preserved", not "handle data"). The plan is approved against these; make them exhaustive so nothing slips past approval.
10: 
11: Fields: `text`.
12: 
13: ```json
14: {
15:   "type": "criteria",
16:   "heading": "Acceptance Criteria",
17:   "items": [
18:     {"text": "All existing data is preserved after migration"},
19:     {"text": "Read replicas sync within 5 seconds of primary"}
20:   ]
21: }
22: ```
23: 
24: _Avoid_: requirements, goals, exit criteria
25: 
26: ## steps
27: 
28: Implementation steps — the numbered, owned, tracked list of work. Each step ends on an implicit completion criterion; pair **status** with an **owner** so responsibility is visible. **status** is one of `pending`, `in-progress`, `done`, `blocked`.
29: 
30: Fields: `text`, `owner`, `status`.
31: 
32: ```json
33: {
34:   "type": "steps",
35:   "heading": "Implementation Steps",
36:   "items": [
37:     {"text": "Provision PostgreSQL 15 instance in staging", "owner": "infra-team", "status": "done"},
38:     {"text": "Run schema compatibility checks on all databases", "owner": "app-team", "status": "in-progress"},
39:     {"text": "Switch write traffic during maintenance window", "owner": "both", "status": "blocked"}
40:   ]
41: }
42: ```
43: 
44: _Avoid_: tasks, actions, todo
45: 
46: ## risks
47: 
48: Risk items — each a threat with its **severity**, **impact**, and **mitigation**. **severity** is `high`, `medium`, or `low`. Put the threat in `text`; the consequence in `impact`; the action in `mitigation`.
49: 
50: Fields: `text`, `severity`, `impact`, `mitigation`.
51: 
52: ```json
53: {
54:   "type": "risks",
55:   "heading": "Risks",
56:   "items": [
57:     {
58:       "text": "Application connection strings need updates across all services",
59:       "severity": "medium",
60:       "impact": "Services unable to connect to new database",
61:       "mitigation": "Use a DNS alias so the connection string remains unchanged"
62:     },
63:     {
64:       "text": "Minor PostgreSQL extension version mismatch",
65:       "severity": "low",
66:       "impact": "Some advanced features may be temporarily unavailable",
67:       "mitigation": "Verify all extensions are compatible with PG15 ahead of time"
68:     }
69:   ]
70: }
71: ```
72: 
73: _Avoid_: issues, concerns, threats
74: 
75: ## decision
76: 
77: Decisions — each a fork-in-the-road that was resolved, recorded with the alternatives considered and the rationale for the winner.
78: Put the chosen decision in `text`; the rejected alternatives in **options**; the reasoning in **rationale**.
79: Use for the output of a grilling session or any plan whose primary content is decisions rather than steps.
80: `criteria` and `risks` belong in their own sibling modules — `decision` does not duplicate them.
81: 
82: Fields: `text`, `options`, `rationale`.
83: 
84: ```json
85: {
86:   "type": "decision",
87:   "heading": "Key Decisions",
88:   "items": [
89:     {
90:       "text": "Use library X for the search layer",
91:       "options": "library Y — too heavy; library Z — unmaintained",
92:       "rationale": "X wins on speed and maintenance; Y's feature set is not needed here"
93:     },
94:     {
95:       "text": "Build the indexing pipeline in-house",
96:       "options": "build in-house; buy SaaS",
97:       "rationale": "build cost is justified by tight latency requirements and existing team expertise"
98:     }
99:   ]
100: }
101: ```
102: 
103: _Avoid_: conclusions, verdicts (use `notes` for those)
104: 
105: ## assumptions
106: 
107: Assumptions being made — the premises the plan rests on, named so they can be challenged. One assumption per item; if one proves false, promote it to a **risk** or a **question**.
108: 
109: Fields: `text`.
110: 
111: ```json
112: {
113:   "type": "assumptions",
114:   "heading": "Assumptions",
115:   "items": [
116:     {"text": "Application uses connection pooling via PgBouncer"},
117:     {"text": "Network latency between old and new instances is under 1ms"}
118:   ]
119: }
120: ```
121: 
122: _Avoid_: premises, givens, prerequisites
123: 
124: ## changes
125: 
126: Files or resources that change — the footprint of the plan. `text` is the path or resource; **type** tags its kind (`terraform`, `config`, `docs`, …). Use for things the plan touches, not for concepts.
127: 
128: Fields: `text`, `changeType` (`type` in JSON).
129: 
130: ```json
131: {
132:   "type": "changes",
133:   "heading": "Changes Required",
134:   "items": [
135:     {"text": "infra/terraform/database.tf", "type": "terraform"},
136:     {"text": "config/deploy.yml", "type": "config"},
137:     {"text": "docs/runbooks/migration.md", "type": "docs"}
138:   ]
139: }
140: ```
141: 
142: _Avoid_: files, artifacts, deliverables
143: 
144: ## notes
145: 
146: Freeform notes — anything that does not fit a typed list. The catch-all; reach for it last, after the typed modules have absorbed what they can. Keep prose here, not actions — actionable work belongs in **steps**.
147: 
148: Fields: `text`.
149: 
150: ```json
151: {
152:   "type": "notes",
153:   "heading": "Notes",
154:   "items": [
155:     {"text": "Coordinate with DevOps to schedule the maintenance window. Suggested: Saturday 02:00–04:00 UTC."},
156:     {"text": "Run the migration script with --dry-run first to verify all steps."}
157:   ]
158: }
159: ```
160: 
161: _Avoid_: comments, remarks, misc
162: 
163: ## questions
164: 
165: Open questions — each an unresolved decision with an **answered** flag and, when answered, an **answer**. `answered` is `true` or `false`; omit `answer` when `answered` is false. When a question resolves, flip `answered` to `true` and fill `answer`.
166: 
167: Fields: `text`, `answered`, `answer`.
168: 
169: ```json
170: {
171:   "type": "questions",
172:   "heading": "Open Questions",
173:   "items": [
174:     {
175:       "text": "Should we keep the old PG12 instance running for 30 days as a fallback?",
176:       "answered": true,
177:       "answer": "Yes — keep for 30 days at reduced cost."
178:     },
179:     {
180:       "text": "What is the acceptable replication lag threshold for cutover?",
181:       "answered": true,
182:       "answer": "Maximum 5 seconds lag before we abort the cutover."
183:     },
184:     {
185:       "text": "Do we need to update any monitoring dashboards or alerts?",
186:       "answered": false
187:     }
188:   ]
189: }
190: ```
191: 
192: _Avoid_: unknowns, todos