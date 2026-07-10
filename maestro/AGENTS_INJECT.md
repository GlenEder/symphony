## Maestro -- Interative Planning Server

When you produce any substantive plan (architecture, design, implementation, refactor, or investigation), you MUST:

1. **Format it as a Maestro plan document** — use the `maestro` skill's `.toon` format with appropriate module types (e.g., `steps` for sequential work, `risks` for trade-offs, `notes` for design rationale, `criteria` for acceptance criteria, `questions` for open decisions).
2. **Serve it via the Maestro web UI** — start the server, write the `.toon` file to `maestro/plans/`, open the browser URL, and enter the listen loop for feedback.

This does NOT apply to: trivial 1-3 line responses, commit messages, or inline code comments. When in doubt, use Maestro.

Direct text output of plan content is the wrong path — the listener should see it rendered in the browser with structured modules, discussion threading, and item-level commenting.
