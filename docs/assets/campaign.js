/* ============================================================
   BQA-OS — QA Agent Campaign (issue #44)
   The campaign scenario taxonomy: 7 original BQA-OS nodes that map
   the `bqa discover -> ingest2 -> build` flow plus discovery,
   pilot, domain, and agent-progression scenarios. Pure data + a
   small unlock resolver so the Forge can light nodes up as a
   decoded archive is processed.

   No DOM, no I/O — require()-able in node tests and loadable as a
   <script> in the browser (exports onto window.BQA_CAMPAIGN).
   Synthetic data only; original BQA-OS names — no third-party game
   references.
   ============================================================ */

"use strict";

(function (root, factory) {
  const api = factory();
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  if (root) root.BQA_CAMPAIGN = api;
})(typeof globalThis !== "undefined" ? globalThis : this, function () {
  // Seven campaign nodes. `unlock` is a predicate-key resolved by
  // unlockedNodes() against a decoded archive result.
  const SCENARIOS = [
    {
      id: "scout-ruins",
      title: "Scout the QA Ruins",
      maps: "bqa discover",
      glyph: "🔭",
      blurb: "Discovery Scout locates candidate session files; unsafe-looking examples are blocked, never shown.",
      unlocks: ["Session Manifest", "Runtime Detector skill", "milestone: Raw Memory Located"],
      unlock: "hasSessions",
    },
    {
      id: "cleanse-stream",
      title: "Cleanse the Memory Stream",
      maps: "bqa ingest2",
      glyph: "🜄",
      blurb: "Session Curator normalizes raw sessions into .bqa/input/sessions/normalized/ and builds index.json.",
      unlocks: ["Normalized Sessions", "Index Builder skill", "Session Curator agent"],
      unlock: "hasSessions",
    },
    {
      id: "forge-brain",
      title: "Forge the QA Brain",
      maps: "bqa build",
      glyph: "⚒",
      blurb: "Knowledge Extractor analyzes normalized sessions into .bqa/knowledge/*.yaml artifact cards.",
      unlocks: ["Knowledge artifacts", "processed-session count", "generated-artifact count"],
      unlock: "hasSpecs",
    },
    {
      id: "interview-guild",
      title: "Interview the Guild",
      maps: "Customer discovery",
      glyph: "🗣",
      blurb: "Interview personas without pitching — good questions reveal pain, trigger, authority, urgency and next step.",
      unlocks: ["Validated Pain Pattern", "Buyer Map", "Pilot Qualified Opportunity", "Exact Customer Quote"],
      unlock: "always",
    },
    {
      id: "pilot",
      title: "Two-Week QA Memory Pilot",
      maps: "Pilot delivery",
      glyph: "🛡",
      blurb: "Import, sanitize, ingest and build 10–30 synthetic artifacts into 3–5 reusable workflows, then review.",
      unlocks: ["Pilot Scope", "Success Criteria", "Review Session", "Renewal Signal"],
      unlock: "pilotReady",
    },
    {
      id: "domain-nodes",
      title: "QA Domain Campaign Nodes",
      maps: "Domain mastery",
      glyph: "🗺",
      blurb: "ETL, GraphQL, API and Data-Quality nodes — each turns concrete QA challenges into BQA-OS knowledge artifacts.",
      unlocks: ["ETL Reconciliation Heuristic", "GraphQL Query Coverage", "API Regression Workflow", "DQ Rule Library"],
      unlock: "hasAnyDomain",
    },
    {
      id: "agent-progression",
      title: "Agent Progression",
      maps: "Roster growth",
      glyph: "♛",
      blurb: "Unlock Discovery Scout, Session Curator, Sanitizer Guardian, Knowledge Extractor, domain specialists and Pilot Manager.",
      unlocks: ["Discovery Scout", "Knowledge Extractor", "GraphQL/ETL Specialists", "Pilot Manager"],
      unlock: "hasAgents",
    },
  ];

  // Domain challenge detail for Scenario 6 (concrete QA language).
  const DOMAIN_NODES = {
    etl: {
      label: "Big Data / ETL QA",
      challenges: ["schema drift", "row count mismatch", "duplicate records", "null spike", "late arriving data", "source-to-target mismatch"],
      unlocks: ["ETL Reconciliation Heuristic", "Data Quality Checklist", "Pipeline Regression Workflow"],
    },
    graphql: {
      label: "GraphQL Functional QA",
      challenges: ["missing field validation", "resolver regression", "authorization mismatch", "query shape drift", "nested response inconsistency"],
      unlocks: ["GraphQL Query Coverage Heuristic", "Resolver Regression Checklist", "Schema Contract Workflow"],
    },
    api: {
      label: "API Testing",
      challenges: ["status code mismatch", "response contract break", "auth/session bug", "pagination inconsistency", "idempotency issue"],
      unlocks: ["API Regression Workflow", "Contract Check Skill"],
    },
    data_quality: {
      label: "Data Quality Validation",
      challenges: ["completeness", "validity", "uniqueness", "consistency", "freshness"],
      unlocks: ["DQ Rule Library", "Anomaly Review Workflow"],
    },
  };

  // Agent + skill roster (Scenario 7) — original BQA-OS concepts.
  const ROSTER = {
    agents: [
      "Discovery Scout", "Session Curator", "Sanitizer Guardian", "Knowledge Extractor",
      "GraphQL Specialist", "ETL Specialist", "Prompt Librarian", "Pilot Manager",
    ],
    skills: [
      "Detect Runtime", "Normalize Session", "Sanitize Memory", "Extract Pattern",
      "Find Common Bug", "Create Workflow", "Prepare Review",
    ],
  };

  // Good/bad discovery questions for Scenario 4 (rewards discovery, not pitching).
  const INTERVIEW = {
    good: [
      "Tell me about the last release where QA became a bottleneck.",
      "What repeated checks live only in people's heads today?",
      "What happened the last time a regression escaped?",
      "What would make a 2-week pilot worth paying for?",
    ],
    bad: [
      "Would you like an AI QA tool?",
      "Can we build a fully autonomous QA agent for you?",
      "Do you want unlimited free pilot support?",
    ],
  };

  // Predicates over a decoded archive result + optional scorecard.
  const PREDICATES = {
    always: () => true,
    hasSessions: (r) => ((r.profile && r.profile.sessions) || 0) > 0,
    hasSpecs: (r) => (r.specs || []).some((s) => (s.findings || 0) > 0),
    hasAgents: (r) => (r.agents || []).length > 0,
    hasAnyDomain: (r) => {
      const sig = (r.profile && r.profile.signals) || {};
      return Object.keys(DOMAIN_NODES).some((d) => (sig[d] || 0) > 0);
    },
    pilotReady: (r, card) => !!(card && card.pilotReady),
  };

  // Return scenarios annotated with `unlocked` for the given result.
  // `result` null => everything except `always` is locked (preview).
  function unlockedNodes(result, card) {
    const r = result || {};
    return SCENARIOS.map((s) => ({
      ...s,
      unlocked: !!(PREDICATES[s.unlock] && PREDICATES[s.unlock](r, card)),
    }));
  }

  // Which concrete domain nodes are present in the decoded archive.
  function activeDomains(result) {
    const sig = (result && result.profile && result.profile.signals) || {};
    return Object.keys(DOMAIN_NODES).filter((d) => (sig[d] || 0) > 0);
  }

  return { SCENARIOS, DOMAIN_NODES, ROSTER, INTERVIEW, unlockedNodes, activeDomains };
});
