#!/usr/bin/env node
/**
 * Deterministic static proxy for planning-skill evaluation.
 *
 * This proxy derives routing candidates from fixture prompts and explicit routing
 * triggers, but only accepts evidence supported by the target skill's frontmatter or
 * body. It detects stale metadata, trigger/body corruption, and ambiguity without
 * pretending to be a model router.

 *
 * Run: node __tests__/skill_eval_harness.js
 */

const fs = require("fs");
const path = require("path");

const root = path.join(__dirname, "..");
const fixtures = JSON.parse(fs.readFileSync(path.join(__dirname, "skill_eval_fixtures.json"), "utf8"));
const planningSkillNames = new Set([
  "maestro", "maestro-author", "maestro-session", "maestro-export",
  "grilling", "plan-implementation-procedure"
]);

function availablePlanningSkills() {
  return fs.readdirSync(path.join(root, "skills"), { withFileTypes: true })
    .filter((entry) => entry.isDirectory() && planningSkillNames.has(entry.name))
    .map((entry) => entry.name)
    .sort();
}

function bodyFor(skill) {
  return fs.readFileSync(path.join(root, "skills", skill, "SKILL.md"), "utf8");
}

function skillMetadata(skill) {
  const body = bodyFor(skill);
  const match = body.match(/^---\n([\s\S]*?)\n---/);
  const frontmatter = match ? match[1] : "";
  const description = (frontmatter.match(/^description:\s*(.+)$/m) || ["", ""])[1].trim();
  return { body, description, searchable: `${description}\n${body}`.toLowerCase() };
}

function ruleTerms(rule) {
  return rule.toLowerCase().split(/[^a-z0-9]+/).filter((term) => term.length > 2)
    .map((term) => term.endsWith("ing") && term.length > 5 ? term.slice(0, -3) : term);
}

function supportedTrigger(trigger, searchable) {
  return searchable.includes(trigger.toLowerCase());
}

function routingTriggerChecks(fixture, skill) {
  const metadata = skillMetadata(skill);
  const triggers = fixture.routingTriggers[skill] || [];
  return triggers.map((trigger) => ({
    trigger,
    found: supportedTrigger(trigger, metadata.searchable)
  }));
}

function candidateScore(fixture, skill) {
  const metadata = skillMetadata(skill);
  const promptTerms = [...new Set(ruleTerms(fixture.prompt))];
  const promptMatches = promptTerms.filter((term) => metadata.searchable.includes(term));
  const triggers = fixture.routingTriggers[skill] || [];
  const triggerMatches = triggers.filter((trigger) => supportedTrigger(trigger, metadata.searchable));
  return {
    skill,
    score: promptMatches.length + (triggerMatches.length * 10),
    promptMatches,
    triggerMatches,
    triggerCount: triggerMatches.length
  };
}

function countTokens(body) {
  return body.trim() ? body.trim().split(/\s+/).length : 0;
}

function evaluate(fixture) {
  const candidates = availablePlanningSkills().map((skill) => candidateScore(fixture, skill));
  const routingChecks = fixture.expectedSkills.flatMap((skill) =>
    routingTriggerChecks(fixture, skill).map((check) => ({ skill, ...check })));
  const routingTriggersValid = routingChecks.every(({ found }) => found);
  const positiveCandidates = candidates.filter(({ triggerCount }) => triggerCount > 0);
  const selectedSkills = positiveCandidates.map(({ skill }) => skill);
  const unexpectedMatches = selectedSkills.filter((skill) => !fixture.expectedSkills.includes(skill));
  const maxScore = Math.max(0, ...positiveCandidates.map(({ score }) => score));
  const topCandidates = positiveCandidates.filter(({ score }) => score === maxScore).map(({ skill }) => skill);
  const ambiguous = unexpectedMatches.length > 0 || topCandidates.length > 1 ||
    topCandidates.some((skill) => !fixture.expectedSkills.includes(skill));
  const loaded = fixture.expectedSkills.map((skill) => {
    const body = bodyFor(skill);
    return { skill, path: `skills/${skill}/SKILL.md`, lines: body.split("\n").length, tokens: countTokens(body) };
  });
  const completionChecks = fixture.completionRules.map((rule) => {
    const skillsBody = loaded.map(({ skill }) => bodyFor(skill)).join("\n");
    return { rule, found: skillsBody.toLowerCase().includes(rule.toLowerCase()) };
  });
  return {
    id: fixture.id, prompt: fixture.prompt, expectedSkills: fixture.expectedSkills,
    routingTriggerChecks: routingChecks, routingTriggersValid, candidateScores: candidates,
    selectedSkills, topCandidates, unexpectedMatches, ambiguous,
    selectionCorrect: routingTriggersValid && !ambiguous &&
      selectedSkills.length === fixture.expectedSkills.length &&
      selectedSkills.every((skill) => fixture.expectedSkills.includes(skill)),
    workflowCompleted: completionChecks.every(({ found }) => found), completionChecks, loadedSkills: loaded,
    totalLines: loaded.reduce((sum, item) => sum + item.lines, 0),
    totalTokens: loaded.reduce((sum, item) => sum + item.tokens, 0)
  };
}

const scenarios = fixtures.map(evaluate);
const result = {
  schemaVersion: 2,
  generatedBy: "__tests__/skill_eval_harness.js",
  mode: "static-analysis-proxy",
  limitations: [
    "No headless agent runtime is available: this deterministic proxy scores declared prompt terms and fixture routing triggers against actual skill frontmatter/body content; it cannot model semantic routing.",
    "Candidate coverage is limited to the relevant planning skill names present in this repository, and selected skills require fixture trigger evidence; it cannot discover undeclared capabilities.",
    "Token counts use whitespace-separated words rather than a model tokenizer.",
    "Loaded bodies are the fixture-declared current SKILL.md files, not observed runtime loads."
  ],
  scenarioCount: scenarios.length,
  completedScenarioCount: scenarios.filter(({ workflowCompleted }) => workflowCompleted).length,
  selectionAccuracy: scenarios.filter(({ selectionCorrect }) => selectionCorrect).length / scenarios.length,
  totalLines: scenarios.reduce((sum, scenario) => sum + scenario.totalLines, 0),
  totalTokens: scenarios.reduce((sum, scenario) => sum + scenario.totalTokens, 0),
  scenarios
};

const outputPath = process.argv[2];
if (outputPath) fs.writeFileSync(path.resolve(outputPath), `${JSON.stringify(result, null, 2)}\n`);
console.log(JSON.stringify(result, null, 2));

