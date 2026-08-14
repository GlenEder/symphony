# Planning Skill Eval — Pre-change vs Post-change Comparison

Stage 5 artifact for the `maestro-skill-decomposition` plan.
Compares the pre-change (monolithic `maestro`) baseline against the post-change (split `maestro-author` + `maestro-session`) harness output.
The harness is a deterministic static-analysis proxy at `__tests__/skill_eval_harness.js` that scores fixture prompt terms and routing triggers against actual `skills/*/SKILL.md` bodies, using `__tests__/skill_eval_fixtures.json`.

## Baseline recovery

The committed `__tests__/skill_eval_baseline.json` was refreshed during Stage 3 and represents a mid-change (post-split) snapshot, not the true pre-change baseline.
The true pre-change baseline was recovered from git history at the Stage 1 merge commit `4b6e055` ("test(skills): add planning skill evaluation harness (#45)"), the monolithic-`maestro` era, via `git show 4b6e055:__tests__/skill_eval_baseline.json`.
That recovered file is the "before"; the current harness run is the "after".
Both artifacts use `schemaVersion: 2`, so the metrics are directly comparable.
The pre-change fixtures expected the monolithic `maestro` for the author-plan and session-loop scenarios, whereas the current fixtures expect `maestro-author` / `maestro-session`, so the contrast measures exactly the skill split.

## Top-level metrics

| metric | before (4b6e055) | after (current run) | delta |
| --- | --- | --- | --- |
| scenarioCount | 4 | 4 | 0 |
| completedScenarioCount | 4 | 4 | 0 |
| selectionAccuracy | 1.00 (4/4) | 1.00 (4/4) | 0 |
| totalLines | 1299 | 590 | -709 (-54.6%) |
| totalTokens | 8189 | 3686 | -4503 (-55.0%) |

## Per-scenario results

All four scenarios hold `selectionCorrect=true`, `ambiguous=false`, and `unexpectedMatches=[]` in both before and after.
No scenario gained ambiguity or an unexpected match.

| scenario | before selected | after selected | before lines / tokens | after lines / tokens | regression |
| --- | --- | --- | --- | --- | --- |
| author-plan | grilling, maestro | grilling, maestro-author | 545 / 3589 | 194 / 1341 | none |
| session-loop | maestro | maestro-session | 401 / 2662 | 43 / 350 | none |
| export-approved-plan | maestro-export | maestro-export | 209 / 1011 | 209 / 1038 | none |
| grill-decision | grilling | grilling | 144 / 927 | 144 / 957 | none |

The author-plan and session-loop scenarios account for the large token reduction.
The monolithic `maestro` skill (401 lines / 2662 tokens) was loaded wholesale for both; after the split, only the relevant focused skill is loaded — `maestro-author` (50 lines / 384 tokens) for authoring, `maestro-session` (43 / 350) for the live session.
The export-approved-plan and grill-decision scenarios are unchanged in lines and rose ~27-30 tokens, attributable to Stage 3 description differentiation adding differentiating words to the `maestro-export` and `grilling` bodies without changing their line counts.
These upticks are not selection regressions: accuracy, ambiguity, and unexpected matches are identical before and after, and the net loaded-body token cost still falls 55%.

## Regression status

No regression found.
`selectionAccuracy` is maintained at 1.00 (4 of 4 scenarios correct).
No new ambiguity and no new unexpected matches in any scenario.
`workflowCompleted` holds at 4 of 4.
No H2 description-tightening loop was required, and no skill `SKILL.md` content was changed in this stage.

## Static-proxy limitation

This comparison validates trigger/body integrity and loaded-body token cost, not semantic routing quality.
The proxy scores declared prompt terms and fixture routing triggers against actual skill frontmatter and body content; it does not run a headless agent or model a router.
A clean result here means the split did not corrupt routing triggers or completion-rule coverage and reduced loaded token cost; it does not by itself prove real-world selection improved.
Token counts use whitespace-separated words rather than a model tokenizer, and loaded bodies are the fixture-declared current `SKILL.md` files, not observed runtime loads.
