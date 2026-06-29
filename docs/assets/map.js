/* ============================================================
   BQA-OS — Agent War Map (issue #26)
   A lightweight, framework-free pixel "RTS map" that shows the
   QA team as units advancing tasks through workflow territories.
   Synthetic data only. No backend. Safe DOM (no innerHTML).
   ============================================================ */

"use strict";

const STAGES = [
  { key: "idea",     label: "Idea",        glyph: "✦" },
  { key: "arch",     label: "Arch Review", glyph: "❖" },
  { key: "ready",    label: "Ready Dev",   glyph: "⚑" },
  { key: "dev",      label: "In Dev",      glyph: "⛏" },
  { key: "qa",       label: "QA",          glyph: "◆" },
  { key: "bug",      label: "Bug Found",   glyph: "☠" },
  { key: "accept",   label: "Biz Accept",  glyph: "⚖" },
  { key: "done",     label: "Done",        glyph: "★" }
];

const ROLES = {
  owner:     { label: "Business Owner", color: "var(--gold)",          glyph: "♛" },
  architect: { label: "Architect",      color: "var(--sky-bright)",    glyph: "❖" },
  developer: { label: "Developer",      color: "var(--forest-bright)", glyph: "⛏" },
  qa:        { label: "QA",             color: "var(--mana)",          glyph: "◆" },
  reviewer:  { label: "Reviewer",       color: "var(--blood)",         glyph: "⚔" }
};

// Which role "carries" a task at each stage.
const STAGE_ROLE = {
  idea: "owner", arch: "architect", ready: "developer", dev: "developer",
  qa: "qa", bug: "developer", accept: "reviewer", done: "owner"
};

// Synthetic backlog (no real data).
let TASKS = [
  { id: 25, title: "Static BQA Web App MVP", stage: "dev" },
  { id: 26, title: "Agent War Map MVP", stage: "dev" },
  { id: 31, title: "Static upload flow", stage: "qa" },
  { id: 28, title: "Sales pilot landing page", stage: "accept" },
  { id: 10, title: "Knowledge domain models", stage: "qa" },
  { id: 11, title: "Harden filesystem store", stage: "qa" },
  { id: 12, title: "Align build CLI to MVP", stage: "ready" },
  { id: 27, title: "Codex Team Pipeline MVP", stage: "arch" },
  { id: 64, title: "ETL QA Agent Pack MVP", stage: "ready" },
  { id: 41, title: "Portable Brain install", stage: "idea" },
  { id: 5,  title: "Knowledge docs + demo", stage: "done" },
  { id: 29, title: "Demo archive generator", stage: "done" }
];

let turn = 1;
const NEXT = { idea: "arch", arch: "ready", ready: "dev", dev: "qa", qa: "accept", bug: "dev", accept: "done", done: "done" };

const $ = (s) => document.querySelector(s);
function el(tag, cls, text) {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (text != null) e.textContent = String(text);
  return e;
}

function stageCounts() {
  const c = {};
  for (const s of STAGES) c[s.key] = 0;
  for (const t of TASKS) c[t.stage] = (c[t.stage] || 0) + 1;
  return c;
}

function renderBoard() {
  const board = $("#board");
  board.replaceChildren();
  const counts = stageCounts();

  for (const stage of STAGES) {
    const col = el("section", "territory");
    col.dataset.stage = stage.key;

    const head = el("header", "terr-head");
    head.appendChild(el("span", "terr-glyph", stage.glyph));
    head.appendChild(el("span", "terr-name", stage.label));
    head.appendChild(el("span", "terr-count", counts[stage.key]));
    col.appendChild(head);

    const yard = el("div", "yard");
    for (const t of TASKS.filter((t) => t.stage === stage.key)) {
      const role = ROLES[STAGE_ROLE[t.stage]];
      const unit = el("button", "token");
      unit.setAttribute("aria-label", "#" + t.id + " " + t.title + " — " + role.label);
      const port = el("span", "token-port", role.glyph);
      port.style.setProperty("--unit", role.color);
      const meta = el("span", "token-meta");
      meta.appendChild(el("b", null, "#" + t.id));
      meta.appendChild(el("small", null, t.title));
      unit.appendChild(port);
      unit.appendChild(meta);
      if (stage.key === "done") unit.classList.add("is-done");
      if (stage.key === "bug") unit.classList.add("is-bug");
      unit.addEventListener("click", () => selectTask(t));
      yard.appendChild(unit);
    }
    if (!counts[stage.key]) yard.appendChild(el("p", "terr-empty", "— empty —"));
    col.appendChild(yard);
    board.appendChild(col);
  }
  $("#turn").textContent = turn;
  renderRoster();
}

function renderRoster() {
  const roster = $("#roster");
  roster.replaceChildren();
  const load = {};
  for (const k of Object.keys(ROLES)) load[k] = 0;
  for (const t of TASKS) if (t.stage !== "done") load[STAGE_ROLE[t.stage]]++;

  for (const k of Object.keys(ROLES)) {
    const r = ROLES[k];
    const row = el("div", "rost-row");
    const port = el("span", "rost-port", r.glyph);
    port.style.setProperty("--unit", r.color);
    const body = el("div", "rost-body");
    body.appendChild(el("b", null, r.label));
    const bar = el("div", "bar");
    const fill = el("span");
    fill.style.width = Math.min(100, load[k] * 25) + "%";
    bar.appendChild(fill);
    body.appendChild(bar);
    row.appendChild(port);
    row.appendChild(body);
    row.appendChild(el("span", "chip", load[k] + " active"));
    roster.appendChild(row);
  }
}

function selectTask(t) {
  const role = ROLES[STAGE_ROLE[t.stage]];
  const stage = STAGES.find((s) => s.key === t.stage);
  const box = $("#detail");
  box.replaceChildren();
  box.appendChild(el("div", "panel-title", "⚑ Order #" + t.id));
  box.appendChild(el("p", null, t.title));
  const line = el("p");
  line.appendChild(el("span", "chip", stage.label));
  line.appendChild(document.createTextNode(" carried by "));
  const who = el("b", null, role.label);
  who.style.color = role.color;
  line.appendChild(who);
  box.appendChild(line);
}

/* Advance the simulation one turn: push a few tasks forward, with a
   small chance a QA task uncovers a bug and retreats to dev. */
function advance() {
  let moved = 0;
  for (const t of TASKS) {
    if (t.stage === "done") continue;
    if (t.stage === "qa" && (t.id % 5 === 1)) { t.stage = "bug"; moved++; continue; }
    if (moved >= 4) break;
    t.stage = NEXT[t.stage];
    moved++;
  }
  // bugs get re-fixed next turn
  for (const t of TASKS) if (t.stage === "bug" && turn % 2 === 0) t.stage = "dev";
  turn++;
  renderBoard();
}

document.addEventListener("DOMContentLoaded", () => {
  $("#advance").addEventListener("click", advance);
  $("#reset").addEventListener("click", () => location.reload());
  renderBoard();
});
