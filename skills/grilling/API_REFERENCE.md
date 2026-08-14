# Grilling API Reference

Quick-reference summary of all API calls used in the Grilling Wizard Flow.

| Step | Method | Endpoint | Purpose |
|------|--------|----------|---------|
| 1 | — | Start server | Ensure Maestro is running |
| 2 | POST | `/api/plans` | Create plan (JSON) |
| 3 | Open browser | `/grill/{id}` | Show wizard |
| 4 | POST | `/api/plan/{id}/messages` | Post question with prompt |
| 5 | POST | `/api/agent/{id}/heartbeat` | Keep agent alive |
| 6 | POST | `/api/plan/{id}/messages` | Mark answered + next question |
| 7 | POST | `/api/plan/{id}/messages` | Confirm shared understanding |
| 8 | Write file | `$MAESTRO_PLANS_DIR/{id}.json` | Populate decisions |
| 8 | POST | `/api/plan/{id}/state` | Set state to draft |
| 8 | POST | `/api/admin/reload` | Force server reload |
| 9 | Open browser | `/plan/{id}` | User reviews plan |
