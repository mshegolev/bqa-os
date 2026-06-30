"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");

const SCORE = require("./scorecard.js");
const CAMPAIGN = require("./campaign.js");

// A rich decoded archive: 5 domains, leveled agents, deep workflows.
function fullResult() {
  return {
    agents: [
      { name: "etl-qa-agent", level: 4 },
      { name: "graphql-contract-agent", level: 4 },
      { name: "api-contract-agent", level: 3 },
      { name: "dq-sentinel-agent", level: 3 },
      { name: "regression-hunter-agent", level: 3 },
    ],
    workflows: [
      { name: "etl_validation_workflow", steps: 6 },
      { name: "graphql_contract_workflow", steps: 5 },
      { name: "api_contract_workflow", steps: 5 },
      { name: "data_quality_workflow", steps: 4 },
      { name: "regression_triage_workflow", steps: 4 },
    ],
    specs: SCORE.KNOWLEDGE_SPECS.map((name) => ({ name, findings: 3 })),
    profile: { sessions: 12, signals: { etl: 3, graphql: 3, api: 2, data_quality: 2, bugs: 2 } },
  };
}

test("scoreResult returns a clamped 0-100 score with a band", () => {
  const card = SCORE.scoreResult(fullResult());
  assert.ok(card.score >= 0 && card.score <= 100, "score is within 0..100");
  assert.ok(["S", "A", "B", "C", "D", "E"].includes(card.band), "band is a known letter");
  // A broad, deep archive should be pilot-ready (>= 70).
  assert.ok(card.score >= 70, "full coverage scores pilot-ready, got " + card.score);
  assert.equal(card.pilotReady, true);
});

test("empty result floors every dimension and is not pilot-ready", () => {
  const card = SCORE.scoreResult({});
  assert.equal(card.score, 0);
  assert.equal(card.pilotReady, false);
  assert.equal(card.band, "E");
  for (const d of card.dimensions) assert.equal(d.score, 0);
});

test("more coverage scores strictly higher than less coverage", () => {
  const rich = SCORE.scoreResult(fullResult());
  const thin = SCORE.scoreResult({
    agents: [{ name: "api-contract-agent", level: 1 }],
    workflows: [{ name: "api_contract_workflow", steps: 2 }],
    specs: [{ name: "api_patterns.yaml", findings: 1 }],
    profile: { sessions: 1, signals: { api: 1 } },
  });
  assert.ok(rich.score > thin.score, "rich (" + rich.score + ") > thin (" + thin.score + ")");
});

test("dimension weights sum to 100 and each dimension is reported", () => {
  const totalWeight = SCORE.DIMENSIONS.reduce((n, d) => n + d.weight, 0);
  assert.equal(totalWeight, 100);
  const card = SCORE.scoreResult(fullResult());
  assert.equal(card.dimensions.length, SCORE.DIMENSIONS.length);
  for (const d of card.dimensions) assert.ok(d.score >= 0 && d.score <= 100);
});

test("weak spots name every missing domain and produce a next action", () => {
  // Only API + GraphQL present -> ETL, DQ, bugs are weak spots.
  const card = SCORE.scoreResult({
    agents: [{ name: "api-contract-agent", level: 2 }, { name: "graphql-contract-agent", level: 2 }],
    workflows: [{ name: "api_contract_workflow", steps: 3 }],
    specs: [{ name: "api_patterns.yaml", findings: 1 }, { name: "graphql_patterns.yaml", findings: 1 }],
    profile: { sessions: 3, signals: { api: 1, graphql: 1 } },
  });
  assert.ok(card.weakSpots.some((w) => /ETL/i.test(w)), "flags missing ETL");
  assert.ok(card.weakSpots.some((w) => /data-quality/i.test(w)), "flags missing data quality");
  assert.ok(card.nextAction && card.nextAction.length > 0, "gives a concrete next action");
});

test("bands map score thresholds to the documented letters", () => {
  assert.equal(SCORE.bandFor(95).band, "S");
  assert.equal(SCORE.bandFor(85).band, "A");
  assert.equal(SCORE.bandFor(72).band, "B");
  assert.equal(SCORE.bandFor(60).band, "C");
  assert.equal(SCORE.bandFor(40).band, "D");
  assert.equal(SCORE.bandFor(10).band, "E");
});

test("example scorecard is stable and pilot-ready threshold is 70", () => {
  assert.equal(SCORE.PILOT_READY_AT, 70);
  const a = SCORE.exampleScorecard();
  const b = SCORE.exampleScorecard();
  assert.deepEqual(a, b, "exampleScorecard is deterministic");
  assert.equal(a.bestDomain, "GraphQL QA", "GraphQL is the strongest signal in the example");
});

test("scorecardYaml renders the score, band and weak spots", () => {
  const yaml = SCORE.scorecardYaml(SCORE.exampleScorecard());
  assert.match(yaml, /qa_memory_score:\s*\d+/);
  assert.match(yaml, /band:\s*[SABCDE]/);
  assert.match(yaml, /next_action:/);
});

/* --- campaign unlock resolver (issue #44) -------------------------- */

test("campaign starts with only always-on nodes unlocked", () => {
  const nodes = CAMPAIGN.unlockedNodes(null, null);
  const states = Object.fromEntries(nodes.map((n) => [n.id, n.unlocked]));
  assert.equal(states["interview-guild"], true, "discovery interview is always available");
  assert.equal(states["scout-ruins"], false);
  assert.equal(states["forge-brain"], false);
  assert.equal(states["pilot"], false);
});

test("decoded archive unlocks discover/ingest/build/domain/agent nodes", () => {
  const result = fullResult();
  const card = SCORE.scoreResult(result);
  const nodes = CAMPAIGN.unlockedNodes(result, card);
  const states = Object.fromEntries(nodes.map((n) => [n.id, n.unlocked]));
  assert.equal(states["scout-ruins"], true, "discover unlocks with sessions");
  assert.equal(states["cleanse-stream"], true, "ingest unlocks with sessions");
  assert.equal(states["forge-brain"], true, "build unlocks with specs");
  assert.equal(states["domain-nodes"], true, "domain node unlocks with any domain signal");
  assert.equal(states["agent-progression"], true, "agent progression unlocks with agents");
  assert.equal(states["pilot"], true, "pilot unlocks when scorecard is pilot-ready");
});

test("pilot node stays locked until the scorecard is pilot-ready", () => {
  const thin = {
    agents: [{ name: "api-contract-agent", level: 1 }],
    workflows: [{ name: "api_contract_workflow", steps: 2 }],
    specs: [{ name: "api_patterns.yaml", findings: 1 }],
    profile: { sessions: 1, signals: { api: 1 } },
  };
  const card = SCORE.scoreResult(thin);
  assert.equal(card.pilotReady, false, "thin archive is not pilot-ready");
  const nodes = CAMPAIGN.unlockedNodes(thin, card);
  const pilot = nodes.find((n) => n.id === "pilot");
  assert.equal(pilot.unlocked, false);
});

test("activeDomains reflects the decoded signals", () => {
  assert.deepEqual(CAMPAIGN.activeDomains({ profile: { signals: { etl: 2, api: 1 } } }).sort(), ["api", "etl"]);
  assert.deepEqual(CAMPAIGN.activeDomains(null), []);
});
