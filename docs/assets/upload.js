/* ============================================================
   BQA-OS — Archive Decoder (issue #31, static upload flow)
   Local-first: parse an uploaded session archive in the browser,
   extract agents / workflows / specs / recommendations, and let
   the user download a generated output .zip. No backend.
   Synthetic data only. DOM built with safe methods (no innerHTML).
   ============================================================ */

"use strict";

/* Inlined synthetic archive so the "Load demo" button works even
   when the page is opened directly from file:// (no fetch). It is
   identical to docs/fixtures/demo-archive.json. */
const DEMO_ARCHIVE = {
  archive: "bqa-demo-archive",
  version: 1,
  sessions: [
    { id: "s-0001", tool: "claude", domain: "etl", title: "Airflow DAG reconciliation review", text: "airflow DAG failure spark row count hive parquet reconciliation etl_logs" },
    { id: "s-0002", tool: "codex", domain: "graphql", title: "GraphQL contract regression", text: "graphql query resolver null contract test graphql schema mutation payload" },
    { id: "s-0003", tool: "claude", domain: "api", title: "REST endpoint status codes", text: "rest api endpoint http status code contract test openapi request payload" },
    { id: "s-0004", tool: "opencode", domain: "data_quality", title: "Schema drift + null checks", text: "data quality schema drift null check duplicate check checksum row count" },
    { id: "s-0005", tool: "claude", domain: "bugs", title: "Flaky pipeline traceback", text: "traceback exception flaky regression failure panic retry" },
    { id: "s-0006", tool: "codex", domain: "prompts", title: "Reusable QA task prompt", text: "your task analyze this repository act as qa implement read .bqa" }
  ]
};

/* Knowledge artifact schema version emitted by this web mirror. Must match
   internal/core/knowledge/contract.go (SchemaVersion) so the demo output is the
   same v1 shape a real `bqa build` writes. */
const KNOWLEDGE_SCHEMA_VERSION = 1;
/* Provenance stamp. This is the browser build, so it is "bqa web" (the Go CLI
   stamps "bqa dev" in tests / "bqa vX.Y.Z" in a release). */
const GENERATED_BY = "bqa web";

/* Domain -> generated unit metadata (RTS flavour) -------------------- */
const DOMAIN_DEFS = {
  etl:          { agent: "etl-qa-agent",           workflow: "etl_validation_workflow",   spec: "etl_patterns.yaml",          role: "Siege Engineer", color: "var(--forest-bright)", glyph: "⛏" },
  graphql:      { agent: "graphql-contract-agent", workflow: "graphql_contract_workflow", spec: "graphql_patterns.yaml",      role: "Arcane Scout",   color: "var(--mana)",          glyph: "❖" },
  api:          { agent: "api-contract-agent",     workflow: "api_contract_workflow",     spec: "api_patterns.yaml",          role: "Gate Warden",    color: "var(--sky-bright)",    glyph: "⚑" },
  data_quality: { agent: "dq-sentinel-agent",      workflow: "data_quality_workflow",     spec: "data_quality_patterns.yaml", role: "Sentinel",       color: "var(--gold)",          glyph: "◆" },
  bugs:         { agent: "regression-hunter-agent",workflow: "regression_triage_workflow",spec: "common_bugs.yaml",           role: "Bug Hunter",     color: "var(--blood)",         glyph: "☠" },
  prompts:      { agent: "prompt-librarian-agent", workflow: "prompt_library_workflow",   spec: "successful_prompts.yaml",    role: "Loremaster",     color: "var(--parchment)",     glyph: "✦" }
};

/* Per-domain metadata used to render v1 knowledge findings. Mirrors the fixed
   finding names, per-domain keyword sets (for confidence), and reusable_check
   templates in internal/core/knowledge (usecase.go / schema.go). */
const KNOWLEDGE_DOMAINS = {
  etl:          { spec: "etl_patterns.yaml",          findingName: "etl_validation",             reusableCheck: "compare source vs target row counts for the window",       keywords: ["airflow", "spark", "hive", "oozie", "dag", "reconciliation", "source table", "target table", "row count", "parquet", "pipeline", "partition", "schedule"] },
  graphql:      { spec: "graphql_patterns.yaml",      findingName: "graphql_functional_testing", reusableCheck: "assert query/mutation response shape and error handling",   keywords: ["graphql", "query", "mutation", "schema", "resolver", "variables", "pagination", "auth", "error"] },
  api:          { spec: "api_patterns.yaml",          findingName: "api_contract_testing",        reusableCheck: "assert endpoint status code and response contract",        keywords: ["rest api", "http status", "status code", "endpoint", "contract", "openapi", "swagger", "request", "response"] },
  data_quality: { spec: "data_quality_patterns.yaml", findingName: "data_quality_validation",     reusableCheck: "assert null / duplicate / schema-drift rules pass",        keywords: ["data quality", "schema drift", "null", "duplicate", "row count", "checksum", "unique", "validation"] },
  bugs:         { spec: "common_bugs.yaml",           findingName: "common_failure_signal",       reusableCheck: "add a regression check reproducing the failure signal",    keywords: ["failed", "failure", "error", "panic", "regression", "flaky", "stack trace", "exception", "traceback"] },
  prompts:      { spec: "successful_prompts.yaml",    findingName: "successful_prompt_candidate", reusableCheck: null /* prompts echo the evidence itself */,               keywords: ["task", "goal", "acceptance", "implement", "verify", "context"] },
};

/* fnv1a32 is a tiny deterministic hash (no crypto dependency, works on file://).
   The Go side uses sha256; here we only need a stable 8-hex id per finding so the
   demo artifacts look and diff like real ones. */
function fnv1a32(str) {
  let h = 0x811c9dc5;
  for (let i = 0; i < str.length; i++) {
    h ^= str.charCodeAt(i);
    h = (h + ((h << 1) + (h << 4) + (h << 7) + (h << 8) + (h << 24))) >>> 0;
  }
  return ("0000000" + h.toString(16)).slice(-8);
}

/* findingId returns "<domain>-<8 hex>", stable in domain|name|source (mirrors
   findingID in internal/core/knowledge/schema.go). */
function findingId(f) {
  return f.domain + "-" + fnv1a32(f.domain + "|" + f.name + "|" + f.source);
}

/* findingConfidence: low/medium/high by distinct domain keywords in evidence
   (high >= 3, medium 2, low otherwise) — mirrors findingConfidence in schema.go. */
function findingConfidence(f) {
  const lower = String(f.evidence || "").toLowerCase();
  const kws = (KNOWLEDGE_DOMAINS[f.domain] || {}).keywords || [];
  let n = 0;
  for (const kw of kws) if (lower.includes(kw)) n++;
  return n >= 3 ? "high" : n === 2 ? "medium" : "low";
}

/* reusableCheck per domain; prompts return the evidence text itself (mirrors
   reusableCheck in schema.go). */
function reusableCheck(f) {
  const def = KNOWLEDGE_DOMAINS[f.domain];
  if (!def) return "add a check that reproduces this signal";
  if (f.domain === "prompts") return f.evidence;
  return def.reusableCheck;
}

/* quoteYaml double-quotes a scalar for YAML, escaping backslashes and quotes so
   evidence with punctuation stays valid (mirrors textutil.QuoteYAML closely
   enough for the synthetic demo). */
function quoteYaml(s) {
  return '"' + String(s).replace(/\\/g, "\\\\").replace(/"/g, '\\"') + '"';
}

let LAST_OUTPUT = null; // { files: {name: content}, result: {...} }

/* --- Extraction (the "local engine") ------------------------------- */
function extract(archive) {
  if (!archive || !Array.isArray(archive.sessions)) {
    throw new Error("archive has no 'sessions' array");
  }
  const counts = {};
  for (const s of archive.sessions) {
    const d = String(s.domain || "").toLowerCase();
    if (DOMAIN_DEFS[d]) counts[d] = (counts[d] || 0) + 1;
  }
  const domains = Object.keys(counts).sort((a, b) => counts[b] - counts[a]);

  const agents = domains.map((d) => ({
    name: DOMAIN_DEFS[d].agent, role: DOMAIN_DEFS[d].role, domain: d,
    color: DOMAIN_DEFS[d].color, glyph: DOMAIN_DEFS[d].glyph,
    level: Math.min(5, counts[d] + 1), power: Math.min(100, 35 + counts[d] * 22)
  }));
  const workflows = domains.map((d) => ({ name: DOMAIN_DEFS[d].workflow, domain: d, steps: counts[d] + 2 }));

  // Specs always emit the MVP artifact set (matches the Go extractor contract).
  const specOrder = ["etl_patterns.yaml", "graphql_patterns.yaml", "api_patterns.yaml",
    "data_quality_patterns.yaml", "common_bugs.yaml", "successful_prompts.yaml", "project_profile.yaml"];
  const specs = specOrder.map((name) => {
    const d = Object.keys(DOMAIN_DEFS).find((k) => DOMAIN_DEFS[k].spec === name);
    // Only domain specs carry real finding counts. project_profile.yaml (no
    // matching domain) is a profile summary, not a finding — report 0 rather
    // than echoing the session count as a fabricated finding total.
    const found = d ? (counts[d] || 0) : 0;
    return { name, findings: found };
  });

  const recommendations = buildRecommendations(domains, counts, archive.sessions.length);
  const profile = { sessions: archive.sessions.length, signals: counts, maturity: domains.length >= 4 ? "established" : "initial" };

  // Collect real per-domain findings from the sessions so the emitted
  // knowledge/*.yaml carries evidence + source, not just a count.
  const findings = collectFindings(archive.sessions);

  return { agents, workflows, specs, recommendations, profile, findings };
}

/* collectFindings groups sessions by domain into v1 findings. Evidence is the
   session's sanitized text (bounded), source is a synthetic normalized path.
   Only domains known to KNOWLEDGE_DOMAINS produce findings. */
function collectFindings(sessions) {
  const byDomain = {};
  for (const s of sessions) {
    const d = String(s.domain || "").toLowerCase();
    const def = KNOWLEDGE_DOMAINS[d];
    if (!def) continue;
    const tool = String(s.tool || "session").toLowerCase();
    const evidence = String(s.text || s.title || "").replace(/\s+/g, " ").trim().slice(0, 480);
    const finding = {
      name: def.findingName,
      domain: d,
      evidence,
      source: "normalized/" + tool + "/" + (s.id || "session") + ".txt",
    };
    (byDomain[d] = byDomain[d] || []).push(finding);
  }
  // Deterministic order by source (mirrors uniqueFindings sort in schema.go).
  for (const d of Object.keys(byDomain)) {
    byDomain[d].sort((a, b) => (a.source < b.source ? -1 : a.source > b.source ? 1 : 0));
  }
  return byDomain;
}

function buildRecommendations(domains, counts, total) {
  const recs = [];
  if (counts.etl) recs.push("Promote ETL reconciliation patterns into a reusable validation workflow.");
  if (counts.graphql) recs.push("Add GraphQL contract tests to the regression gate before deploy.");
  if (counts.bugs) recs.push("Triage flaky failures separately from hard regressions.");
  if (counts.prompts) recs.push("Save high-signal prompts to the prompt library for onboarding.");
  if (counts.data_quality) recs.push("Schedule schema-drift and null-check sentinels on critical tables.");
  if (!recs.length) recs.push("No strong signals found - ingest more sanitized sessions to level up agents.");
  recs.push(total + " sessions analyzed across " + domains.length + " QA domains.");
  return recs;
}

/* --- YAML + output assembly (v1 knowledge schema) ------------------- */
// Emits the same v1 envelope as internal/core/knowledge (schema_version / kind /
// generated_by), with a `patterns:` list for domain specs and a `profile:` block
// for project_profile.yaml. Keep in sync with docs/knowledge-artifacts.md.
function artifactHeader(kind) {
  return "schema_version: " + KNOWLEDGE_SCHEMA_VERSION + "\n" +
    "kind: " + kind + "\n" +
    "generated_by: " + GENERATED_BY + "\n";
}

function renderPatternsYaml(kind, findings) {
  let out = artifactHeader(kind);
  if (!findings || !findings.length) return out + "patterns: []\n";
  out += "patterns:\n";
  for (const f of findings) {
    out += "  - id: " + quoteYaml(findingId(f)) + "\n";
    out += "    name: " + quoteYaml(f.name) + "\n";
    out += "    domain: " + quoteYaml(f.domain) + "\n";
    out += "    evidence: " + quoteYaml(f.evidence) + "\n";
    out += "    source: " + quoteYaml(f.source) + "\n";
    out += "    reusable_check: " + quoteYaml(reusableCheck(f)) + "\n";
    out += "    confidence: " + findingConfidence(f) + "\n";
  }
  return out;
}

function renderProfileYaml(result) {
  const signals = {
    etl: 0, graphql: 0, api: 0, data_quality: 0, droid: 0, runtime: 0,
    ...(result.profile.signals || {}),
  };
  // Domains with signals > 0, ordered by count desc then name asc (deterministic).
  const detected = ["etl", "graphql", "api", "data_quality", "droid", "runtime"]
    .filter((d) => signals[d] > 0)
    .sort((a, b) => (signals[b] - signals[a]) || (a < b ? -1 : 1));

  let out = artifactHeader("project_profile");
  out += "profile:\n";
  out += "  sessions_analyzed: " + result.profile.sessions + "\n";
  out += "  domains_detected: [" + detected.join(", ") + "]\n";
  out += "  signals:\n";
  for (const d of ["etl", "graphql", "api", "data_quality", "droid", "runtime"]) {
    out += "    " + d + ": " + signals[d] + "\n";
  }
  out += "  suggested_next_reviews:\n";
  if (!detected.length) {
    out += "    []\n";
  } else {
    for (const d of detected) {
      out += "    - " + quoteYaml("Review " + d + " coverage (" + signals[d] + " signals).") + "\n";
    }
  }
  out += "  maturity: initial\n";
  return out;
}

function renderSpecYaml(spec, result) {
  if (spec.name === "project_profile.yaml") return renderProfileYaml(result);
  const kind = spec.name.replace(".yaml", "");
  const domain = Object.keys(KNOWLEDGE_DOMAINS).find((d) => KNOWLEDGE_DOMAINS[d].spec === spec.name);
  const findings = domain ? ((result.findings || {})[domain] || []) : [];
  return renderPatternsYaml(kind, findings);
}
function buildOutputFiles(result) {
  const files = {};
  for (const spec of result.specs) files["knowledge/" + spec.name] = renderSpecYaml(spec, result);
  files["agents/agents.md"] = "# Generated agents (synthetic)\n\n" +
    result.agents.map((a) => "- **" + a.name + "** (" + a.role + ") - lvl " + a.level + ", domain " + a.domain).join("\n") + "\n";
  files["workflows/workflows.md"] = "# Generated workflows (synthetic)\n\n" +
    result.workflows.map((w) => "- " + w.name + " - " + w.steps + " steps").join("\n") + "\n";
  files["recommendations.md"] = "# Recommendations\n\n" + result.recommendations.map((r) => "- " + r).join("\n") + "\n";
  files["result.json"] = JSON.stringify(result, null, 2) + "\n";
  // QA Memory Score scorecard (issue #46) — only when the scorer is loaded.
  if (typeof BQA_SCORE !== "undefined") {
    const card = BQA_SCORE.scoreResult(result);
    files[".bqa/scorecard.yaml"] = BQA_SCORE.scorecardYaml(card);
  }
  return files;
}

/* --- Minimal store-only ZIP writer (no dependencies) --------------- */
const CRC_TABLE = (() => {
  const t = new Uint32Array(256);
  for (let n = 0; n < 256; n++) {
    let c = n;
    for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    t[n] = c >>> 0;
  }
  return t;
})();
function crc32(bytes) {
  let c = 0xffffffff;
  for (let i = 0; i < bytes.length; i++) c = CRC_TABLE[(c ^ bytes[i]) & 0xff] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}
function buildZip(files) {
  const enc = new TextEncoder();
  const chunks = [];
  const central = [];
  let offset = 0;
  const u16 = (n) => [n & 0xff, (n >>> 8) & 0xff];
  const u32 = (n) => [n & 0xff, (n >>> 8) & 0xff, (n >>> 16) & 0xff, (n >>> 24) & 0xff];

  for (const name of Object.keys(files)) {
    const nameBytes = enc.encode(name);
    const data = enc.encode(files[name]);
    const crc = crc32(data);
    const local = [].concat(
      u32(0x04034b50), u16(20), u16(0), u16(0), u16(0), u16(0),
      u32(crc), u32(data.length), u32(data.length), u16(nameBytes.length), u16(0)
    );
    chunks.push(new Uint8Array(local), nameBytes, data);
    central.push({ name: nameBytes, crc, size: data.length, offset });
    offset += local.length + nameBytes.length + data.length;
  }

  const cdir = [];
  let cdirSize = 0;
  for (const e of central) {
    const rec = [].concat(
      u32(0x02014b50), u16(20), u16(20), u16(0), u16(0), u16(0), u16(0),
      u32(e.crc), u32(e.size), u32(e.size), u16(e.name.length),
      u16(0), u16(0), u16(0), u16(0), u32(0), u32(e.offset)
    );
    cdir.push(new Uint8Array(rec), e.name);
    cdirSize += rec.length + e.name.length;
  }
  const eocd = [].concat(
    u32(0x06054b50), u16(0), u16(0), u16(central.length), u16(central.length),
    u32(cdirSize), u32(offset), u16(0)
  );
  return new Blob([...chunks, ...cdir, new Uint8Array(eocd)], { type: "application/zip" });
}

/* --- Safe DOM helpers (no innerHTML) ------------------------------- */
const $ = (sel) => document.querySelector(sel);
function el(tag, cls, text) {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (text != null) e.textContent = String(text);
  return e;
}

/* --- Rendering ----------------------------------------------------- */
function renderResult(result, sourceName) {
  $("#empty").style.display = "none";
  $("#results").style.display = "block";
  $("#source-name").textContent = sourceName;

  const agentGrid = $("#agents"); agentGrid.replaceChildren();
  for (const a of result.agents) {
    const card = el("div", "unit");
    const portrait = el("div", "unit-portrait");
    portrait.style.setProperty("--unit", a.color);
    portrait.appendChild(el("span", null, a.glyph));
    const body = el("div", "unit-body");
    body.appendChild(el("b", null, a.name));
    body.appendChild(el("div", "unit-role", a.role + " · LVL " + a.level));
    const bar = el("div", "bar");
    const fill = el("span");
    fill.style.width = a.power + "%";
    bar.appendChild(fill);
    body.appendChild(bar);
    card.appendChild(portrait);
    card.appendChild(body);
    agentGrid.appendChild(card);
  }

  fillList($("#workflows"), result.workflows, (w) => [w.steps + " steps", w.name]);
  fillList($("#specs"), result.specs, (s) => [String(s.findings), s.name]);

  const recList = $("#recs"); recList.replaceChildren();
  for (const r of result.recommendations) recList.appendChild(el("li", null, r));

  runPipeline(result);                 // #43 staged processing viewer
  const card = renderScorecard(result); // #46 QA Memory Score
  renderCampaign(result, card);         // #44 campaign scenarios

  LAST_OUTPUT = { files: buildOutputFiles(result), result };
  $("#download").disabled = false;
  setStatus("Decoded " + result.profile.sessions + " sessions -> " + result.agents.length +
    " agents, " + result.specs.length + " specs.", "ok");
}

/* --- Processing pipeline viewer (issue #43) ------------------------ */
// Six original BQA-OS forge stages: raw QA chaos -> downloadable QA system.
const PIPELINE_STAGES = [
  { glyph: "⛏", label: "Session Mine", note: "Raw QA notes mined as resources" },
  { glyph: "🜄", label: "Sanitizer Gate", note: "Dangerous/private data removed" },
  { glyph: "⚒", label: "Knowledge Forge", note: "Reusable patterns extracted" },
  { glyph: "⛨", label: "Skill Barracks", note: "Specialized skills unlocked" },
  { glyph: "♛", label: "Agent Hall", note: "QA agents recruited" },
  { glyph: "⚑", label: "Workflow Command", note: "QA system exported" },
];
let PIPELINE_TIMER = null;
function reducedMotion() {
  try { return window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches; }
  catch (_) { return false; }
}
function renderPipelineStages() {
  const box = $("#pipeline"); if (!box) return [];
  box.replaceChildren();
  return PIPELINE_STAGES.map((s) => {
    const cell = el("div", "pstage");
    cell.appendChild(el("span", "pg", s.glyph));
    cell.appendChild(el("b", null, s.label));
    cell.appendChild(el("small", null, s.note));
    box.appendChild(cell);
    return cell;
  });
}
function runPipeline(result) {
  const cells = renderPipelineStages();
  if (!cells.length) return;
  if (PIPELINE_TIMER) { clearInterval(PIPELINE_TIMER); PIPELINE_TIMER = null; }
  // Reduced-motion: mark every stage done immediately, no animation.
  if (reducedMotion() || typeof setInterval === "undefined") {
    cells.forEach((c) => { c.className = "pstage done"; });
    return;
  }
  let i = 0;
  const advance = () => {
    if (i > 0) cells[i - 1].className = "pstage done";
    if (i >= cells.length) { clearInterval(PIPELINE_TIMER); PIPELINE_TIMER = null; return; }
    cells[i].className = "pstage active";
    i++;
  };
  advance();
  PIPELINE_TIMER = setInterval(advance, 380);
}

/* --- QA Memory Score (issue #46) ----------------------------------- */
function renderScorecard(result) {
  const box = $("#scorecard"); if (!box || typeof BQA_SCORE === "undefined") return null;
  const card = BQA_SCORE.scoreResult(result);
  box.replaceChildren();

  const dial = el("div", "score-dial");
  const num = el("div", "score-num"); num.appendChild(document.createTextNode(String(card.score)));
  num.appendChild(el("small", null, " / 100"));
  dial.appendChild(num);
  const band = el("div", "score-band " + (card.pilotReady ? "ready" : "notready"), card.band);
  dial.appendChild(band);
  dial.appendChild(el("div", "score-bandlabel", card.bandLabel));
  dial.appendChild(el("span", "chip pilot-chip " + (card.pilotReady ? "ready" : "notready"),
    card.pilotReady ? "✓ Pilot-ready" : "Pilot-ready ≥ " + card.pilotReadyAt));
  box.appendChild(dial);

  const detail = el("div");
  for (const d of card.dimensions) {
    const row = el("div", "dim-row");
    row.appendChild(el("span", "dim-label", d.label));
    const bar = el("div", "dim-bar"); const fill = el("span"); fill.style.width = d.score + "%"; bar.appendChild(fill);
    row.appendChild(bar);
    row.appendChild(el("span", "dim-val", String(d.score)));
    detail.appendChild(row);
  }
  const meta = el("div", "score-meta");
  const best = el("span"); best.appendChild(document.createTextNode("Best: ")); best.appendChild(el("b", null, card.bestDomain)); meta.appendChild(best);
  const needs = el("span"); needs.appendChild(document.createTextNode("Needs work: ")); needs.appendChild(el("b", null, card.needsWork)); meta.appendChild(needs);
  detail.appendChild(meta);
  if (card.weakSpots.length) {
    const weak = el("ul", "weak");
    for (const w of card.weakSpots) weak.appendChild(el("li", null, w));
    detail.appendChild(weak);
  }
  box.appendChild(detail);

  const na = $("#next-action");
  if (na) { na.replaceChildren(el("b", null, "Recommended next action: "), document.createTextNode(card.nextAction)); }
  return card;
}

/* --- Campaign scenario map (issue #44) ----------------------------- */
function renderCampaign(result, card) {
  const box = $("#campaign"); if (!box || typeof BQA_CAMPAIGN === "undefined") return;
  box.replaceChildren();
  for (const node of BQA_CAMPAIGN.unlockedNodes(result, card)) {
    const cell = el("div", "node" + (node.unlocked ? " on" : ""));
    const head = el("div", "node-head");
    head.appendChild(el("span", "ng", node.glyph));
    head.appendChild(el("b", null, node.title));
    cell.appendChild(head);
    cell.appendChild(el("div", "node-maps", node.maps));
    cell.appendChild(el("div", "node-blurb", node.blurb));
    const unlocks = el("div", "node-unlocks");
    for (const u of node.unlocks) unlocks.appendChild(el("span", "chip", u));
    cell.appendChild(unlocks);
    cell.appendChild(el("div", "lock", node.unlocked ? "✓ unlocked" : "🔒 locked"));
    box.appendChild(cell);
  }
}

function fillList(ul, items, parts) {
  ul.replaceChildren();
  for (const it of items) {
    const [chip, label] = parts(it);
    const li = el("li");
    li.appendChild(el("span", "chip", chip));
    li.appendChild(document.createTextNode(" " + label));
    ul.appendChild(li);
  }
}

function setStatus(msg, kind) {
  const s = $("#status");
  s.textContent = "> " + msg;
  s.style.color = kind === "err" ? "var(--blood)" : kind === "ok" ? "var(--forest-bright)" : "var(--parchment-dim)";
}

/* --- Input handling ------------------------------------------------ */
function handleArchive(obj, name) {
  try {
    renderResult(extract(obj), name);
    // Hand the decoded archive to the battle (game.html reads this).
    try { sessionStorage.setItem("bqa-archive", JSON.stringify(obj)); } catch (_) {}
    const field = document.getElementById("to-field");
    if (field) field.style.display = "";
  } catch (e) { setStatus("decode failed: " + e.message, "err"); }
}
/* --- ZIP reading (store + deflate) --------------------------------- */
async function inflateRaw(bytes) {
  if (typeof DecompressionStream === "undefined") throw new Error("this browser can't inflate; use a stored zip or json");
  const ds = new DecompressionStream("deflate-raw");
  const ab = await new Response(new Blob([bytes]).stream().pipeThrough(ds)).arrayBuffer();
  return new Uint8Array(ab);
}
async function readZip(buffer) {
  const u8 = new Uint8Array(buffer), dv = new DataView(buffer), dec = new TextDecoder();
  const files = {}; let off = 0;
  while (off + 4 <= u8.length && dv.getUint32(off, true) === 0x04034b50) {
    const method = dv.getUint16(off + 8, true);
    const compSize = dv.getUint32(off + 18, true);
    const nameLen = dv.getUint16(off + 26, true), extraLen = dv.getUint16(off + 28, true);
    const name = dec.decode(u8.subarray(off + 30, off + 30 + nameLen));
    const dataStart = off + 30 + nameLen + extraLen;
    const raw = u8.subarray(dataStart, dataStart + compSize);
    let out = raw;
    if (method === 8) out = await inflateRaw(raw);
    else if (method !== 0) { off = dataStart + compSize; continue; }
    files[name] = dec.decode(out);
    off = dataStart + compSize;
  }
  return files;
}
// pick the most archive-like JSON entry from a zip's text files
function archiveFromZip(files) {
  const jsons = Object.keys(files).filter((n) => n.toLowerCase().endsWith(".json"));
  let best = null;
  for (const n of jsons) {
    try { const o = JSON.parse(files[n]); if (o && (Array.isArray(o.sessions) || Array.isArray(o.agents))) return o; if (!best) best = o; } catch (_) {}
  }
  if (best) return best;
  throw new Error("no archive JSON (with sessions/agents) found in the zip");
}

function readFile(file) {
  const name = (file.name || "").toLowerCase();
  setStatus("reading " + file.name + " ...");
  const reader = new FileReader();
  reader.onerror = () => setStatus("could not read file", "err");
  if (name.endsWith(".zip")) {
    reader.onload = () => {
      readZip(reader.result).then((files) => handleArchive(archiveFromZip(files), file.name))
        .catch((e) => setStatus("zip read failed: " + e.message, "err"));
    };
    reader.readAsArrayBuffer(file);
  } else {
    reader.onload = () => {
      try { handleArchive(JSON.parse(reader.result), file.name); }
      catch (e) { setStatus("not valid JSON: " + e.message, "err"); }
    };
    reader.readAsText(file);
  }
}

function downloadZip() {
  if (!LAST_OUTPUT) return;
  const blob = buildZip(LAST_OUTPUT.files);
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url; a.download = "bqa-os-output.zip";
  document.body.appendChild(a); a.click(); a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
  setStatus("output zip generated (" + Object.keys(LAST_OUTPUT.files).length + " files).", "ok");
}

document.addEventListener("DOMContentLoaded", () => {
  const drop = $("#drop");
  const input = $("#file");

  $("#pick").addEventListener("click", () => input.click());
  input.addEventListener("change", (e) => { if (e.target.files[0]) readFile(e.target.files[0]); });
  const loadDemo = () => handleArchive(DEMO_ARCHIVE, "demo-archive.json");
  $("#demo").addEventListener("click", loadDemo);
  const demo2 = $("#demo2");
  if (demo2) demo2.addEventListener("click", () => {
    loadDemo();
    const r = document.getElementById("results");
    if (r) r.scrollIntoView({ behavior: "smooth", block: "start" });
  });
  $("#download").addEventListener("click", downloadZip);

  ["dragenter", "dragover"].forEach((ev) => drop.addEventListener(ev, (e) => { e.preventDefault(); drop.classList.add("over"); }));
  ["dragleave", "drop"].forEach((ev) => drop.addEventListener(ev, (e) => { e.preventDefault(); drop.classList.remove("over"); }));
  drop.addEventListener("drop", (e) => { const f = e.dataTransfer.files[0]; if (f) readFile(f); });

  // Preview the campaign map (all nodes locked except always-on) before any decode.
  renderCampaign(null, null);

  setStatus("awaiting archive. drop a .json or load the demo.");
});
