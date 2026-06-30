/* ============================================================
   BQA-OS — QA Memory Score (issue #46)
   A small, explainable scoring/rating layer for the demo.
   Pure logic: scores a decoded archive `result` (the shape the
   Forge / upload.js produces: { agents, workflows, specs,
   recommendations, profile }) into a 0–100 QA Memory Score with
   five weighted dimensions, a D/C/B/A/S band, weak spots and a
   recommended next action.

   No DOM, no I/O — safe to require() in node tests and to load as
   a <script> in the browser (exports onto window.BQA_SCORE).
   Synthetic data only.
   ============================================================ */

"use strict";

(function (root, factory) {
  const api = factory();
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  if (root) root.BQA_SCORE = api;
})(typeof globalThis !== "undefined" ? globalThis : this, function () {
  // The five QA domains we reward coverage across.
  const DOMAINS = ["etl", "graphql", "api", "data_quality", "bugs"];

  // The MVP knowledge artifact set the Forge always emits.
  const KNOWLEDGE_SPECS = [
    "etl_patterns.yaml", "graphql_patterns.yaml", "api_patterns.yaml",
    "data_quality_patterns.yaml", "common_bugs.yaml",
    "successful_prompts.yaml", "project_profile.yaml",
  ];

  // Dimension weights (sum = 100). The score is explainable: each
  // dimension is scored 0–100 then folded by its weight.
  const DIMENSIONS = [
    { key: "knowledge", label: "Knowledge Coverage", weight: 26 },
    { key: "agents", label: "Agent Readiness", weight: 22 },
    { key: "workflow", label: "Workflow Reuse", weight: 20 },
    { key: "evidence", label: "Evidence", weight: 18 },
    { key: "pilot", label: "Pilot Readiness", weight: 14 },
  ];

  // Letter bands. Score >= 70 is pilot-ready (per issue #46).
  const BANDS = [
    { min: 90, band: "S", label: "Pilot-ready, broad coverage" },
    { min: 80, band: "A", label: "Strong — pilot-ready" },
    { min: 70, band: "B", label: "Solid — pilot-ready" },
    { min: 55, band: "C", label: "Promising — needs depth" },
    { min: 35, band: "D", label: "Early — keep ingesting" },
    { min: 0, band: "E", label: "Sparse — feed more sessions" },
  ];

  const PILOT_READY_AT = 70;

  function clamp(v, lo, hi) { return Math.max(lo, Math.min(hi, v)); }
  function round(v) { return Math.round(v); }

  function bandFor(score) {
    for (const b of BANDS) if (score >= b.min) return b;
    return BANDS[BANDS.length - 1];
  }

  // Count how many of the rewarded domains have any signal.
  function coveredDomains(signals) {
    if (!signals || typeof signals !== "object") return [];
    return DOMAINS.filter((d) => (signals[d] || 0) > 0);
  }

  /* ---- dimension scorers (each returns 0..100) ---------------- */

  // Knowledge Coverage: domains covered + reusable patterns extracted.
  function scoreKnowledge(result) {
    const signals = (result.profile && result.profile.signals) || {};
    const covered = coveredDomains(signals).length;
    const domainScore = (covered / DOMAINS.length) * 70;          // up to 70 for breadth
    const findings = (result.specs || []).reduce((n, s) => n + (s.findings || 0), 0);
    const depthScore = clamp(findings / 12, 0, 1) * 30;           // up to 30 for depth
    return clamp(domainScore + depthScore, 0, 100);
  }

  // Agent Readiness: do we have agents, and are they leveled (role/power)?
  function scoreAgents(result) {
    const agents = result.agents || [];
    if (!agents.length) return 0;
    const countScore = clamp(agents.length / 5, 0, 1) * 60;      // up to 60 for breadth
    const avgLevel = agents.reduce((n, a) => n + (a.level || 1), 0) / agents.length;
    const levelScore = clamp((avgLevel - 1) / 4, 0, 1) * 40;     // up to 40 for maturity (lvl 1..5)
    return clamp(countScore + levelScore, 0, 100);
  }

  // Workflow Reuse: named, repeatable workflows with real steps.
  function scoreWorkflow(result) {
    const workflows = result.workflows || [];
    if (!workflows.length) return 0;
    const countScore = clamp(workflows.length / 5, 0, 1) * 60;
    const avgSteps = workflows.reduce((n, w) => n + (w.steps || 0), 0) / workflows.length;
    const depthScore = clamp((avgSteps - 2) / 4, 0, 1) * 40;     // 2 steps = floor, 6+ = full
    return clamp(countScore + depthScore, 0, 100);
  }

  // Evidence: artifacts grounded in supplied sessions, not generic text.
  function scoreEvidence(result) {
    const sessions = (result.profile && result.profile.sessions) || 0;
    if (!sessions) return 0;
    const volumeScore = clamp(sessions / 10, 0, 1) * 60;         // ~10 sessions = full volume
    const grounded = (result.specs || []).filter((s) => (s.findings || 0) > 0).length;
    const groundedScore = clamp(grounded / KNOWLEDGE_SPECS.length, 0, 1) * 40;
    return clamp(volumeScore + groundedScore, 0, 100);
  }

  // Pilot Readiness: blended signal — enough domains + agents + workflows
  // to run a QA Lead review session.
  function scorePilot(result) {
    const signals = (result.profile && result.profile.signals) || {};
    const covered = coveredDomains(signals).length;
    const agents = (result.agents || []).length;
    const workflows = (result.workflows || []).length;
    const breadth = clamp(covered / 4, 0, 1) * 50;               // 4+ domains = full breadth
    const muster = clamp((agents + workflows) / 8, 0, 1) * 50;   // 8 deliverables = full muster
    return clamp(breadth + muster, 0, 100);
  }

  const SCORERS = {
    knowledge: scoreKnowledge,
    agents: scoreAgents,
    workflow: scoreWorkflow,
    evidence: scoreEvidence,
    pilot: scorePilot,
  };

  /* ---- weak spots & next action ------------------------------ */

  function weakSpots(result, dims) {
    const signals = (result.profile && result.profile.signals) || {};
    const out = [];
    // Missing-domain gaps (concrete, actionable).
    const missing = DOMAINS.filter((d) => !((signals[d] || 0) > 0));
    const LABELS = {
      etl: "ETL validation patterns are missing",
      graphql: "GraphQL contract coverage is missing",
      api: "API regression patterns are missing",
      data_quality: "Data-quality rules are missing",
      bugs: "Recurring bug taxonomy is thin",
    };
    for (const d of missing) out.push(LABELS[d]);
    // Weak dimensions (below the pilot-ready bar).
    for (const dim of DIMENSIONS) {
      if (dims[dim.key] < PILOT_READY_AT) {
        if (dim.key === "agents") out.push("Agent guardrails need improvement");
        else if (dim.key === "workflow") out.push("No repeatable contract workflow detected");
        else if (dim.key === "evidence") out.push("Artifacts need more grounding sessions");
      }
    }
    // De-dup, keep order, cap to keep the panel scannable.
    return [...new Set(out)].slice(0, 5);
  }

  function nextAction(result) {
    const signals = (result.profile && result.profile.signals) || {};
    const missing = DOMAINS.filter((d) => !((signals[d] || 0) > 0));
    const HINTS = {
      etl: "Upload 5–10 sanitized ETL validation checklists.",
      data_quality: "Add a data-quality rule library (completeness, uniqueness, freshness).",
      graphql: "Add GraphQL contract/regression notes for resolver coverage.",
      api: "Upload 5–10 sanitized API regression notes.",
      bugs: "Attach recurring bug reports to build a common-bugs taxonomy.",
    };
    if (missing.length) return HINTS[missing[0]];
    const sessions = (result.profile && result.profile.sessions) || 0;
    if (sessions < 10) return "Ingest more sanitized sessions to deepen each pattern set.";
    return "Schedule a QA Lead review session and lock the pilot scope.";
  }

  /* ---- best / needs-work domain (the B+ / GraphQL example) ---- */

  function domainBreakdown(result) {
    const signals = (result.profile && result.profile.signals) || {};
    const ranked = DOMAINS.map((d) => ({ domain: d, signal: signals[d] || 0 }))
      .sort((a, b) => b.signal - a.signal);
    const PRETTY = {
      etl: "ETL QA", graphql: "GraphQL QA", api: "API QA",
      data_quality: "Data Quality Validation", bugs: "Bug Reporting",
    };
    const best = ranked[0] && ranked[0].signal > 0 ? PRETTY[ranked[0].domain] : "—";
    const worst = ranked[ranked.length - 1];
    return { best, needsWork: PRETTY[worst.domain] };
  }

  /* ---- public API -------------------------------------------- */

  // scoreResult(result) -> a full, explainable scorecard object.
  function scoreResult(result) {
    result = result || {};
    const dims = {};
    let total = 0;
    const breakdown = DIMENSIONS.map((dim) => {
      const raw = clamp(SCORERS[dim.key](result), 0, 100);
      dims[dim.key] = raw;
      total += raw * (dim.weight / 100);
      return { key: dim.key, label: dim.label, weight: dim.weight, score: round(raw) };
    });
    const score = clamp(round(total), 0, 100);
    const band = bandFor(score);
    const { best, needsWork } = domainBreakdown(result);
    const unlocked = (result.agents || []).map((a) => a.name);
    return {
      version: 1,
      score,
      band: band.band,
      bandLabel: band.label,
      pilotReady: score >= PILOT_READY_AT,
      pilotReadyAt: PILOT_READY_AT,
      dimensions: breakdown,
      bestDomain: best,
      needsWork,
      unlocked,
      weakSpots: weakSpots(result, dims),
      nextAction: nextAction(result),
    };
  }

  // A stable synthetic scorecard for the demo / docs / examples.
  function exampleScorecard() {
    return scoreResult({
      agents: [
        { name: "api-contract-agent", level: 3 },
        { name: "graphql-contract-agent", level: 4 },
        { name: "regression-hunter-agent", level: 2 },
      ],
      workflows: [
        { name: "api_contract_workflow", steps: 4 },
        { name: "graphql_contract_workflow", steps: 5 },
        { name: "regression_triage_workflow", steps: 3 },
      ],
      specs: KNOWLEDGE_SPECS.map((name) => ({ name, findings: 2 })),
      profile: { sessions: 6, signals: { api: 2, graphql: 2, bugs: 1 } },
    });
  }

  // Render the scorecard as `.bqa/scorecard.yaml` text (synthetic).
  function scorecardYaml(card) {
    const lines = [
      "# .bqa/scorecard.yaml — generated by BQA-OS (synthetic)",
      "qa_memory_score: " + card.score,
      "band: " + card.band + "   # " + card.bandLabel,
      "pilot_ready: " + card.pilotReady + "   # threshold: " + card.pilotReadyAt,
      "best_domain: " + JSON.stringify(card.bestDomain),
      "needs_work: " + JSON.stringify(card.needsWork),
      "dimensions:",
    ];
    for (const d of card.dimensions) lines.push("  " + d.key + ": " + d.score + "   # " + d.label + " (w" + d.weight + ")");
    lines.push("unlocked:");
    for (const u of card.unlocked) lines.push("  - " + u);
    lines.push("weak_spots:");
    for (const w of card.weakSpots) lines.push("  - " + JSON.stringify(w));
    lines.push("next_action: " + JSON.stringify(card.nextAction));
    return lines.join("\n") + "\n";
  }

  return {
    DOMAINS, DIMENSIONS, BANDS, KNOWLEDGE_SPECS, PILOT_READY_AT,
    scoreResult, exampleScorecard, scorecardYaml, bandFor,
  };
});
