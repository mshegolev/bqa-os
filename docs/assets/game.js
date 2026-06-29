/* ============================================================
   CITADEL: The Release War — engine.
   A light, log-driven "inspired by 90s fantasy RTS" demo that
   visualizes an SDLC/security workflow as Humans (agents) vs
   Orcs (deadlines/bugs/CVEs/incidents). Demo logs play in
   sequence and spawn units/threats; resources & a Trust Score
   move; the run ends in victory or defeat.
   Vanilla canvas + DOM. Depends on game-logs.js. No assets.
   ============================================================ */

"use strict";

/* ---- isometric grid -------------------------------------- */
const GW = 12, GH = 12, TW = 56, TH = 28, TZ = 12;
let ORIGIN_X = 0, ORIGIN_Y = 60;

const PAL = {
  grassTop: "#4a7a3a", grassL: "#3c6630", grassR: "#2f5226",
  roadTop: "#a3814a", roadL: "#8a6a3a", roadR: "#6e5230",
  waterTop: "#2f6090", waterL: "#274f78", waterR: "#1e3d5e",
  forestTop: "#3c6e30", forestL: "#2f5a26", forestR: "#244a1e",
  gold: "#f2c14e", bone: "#f4eddc", blood: "#c0392b", forestBright: "#7bc45a",
  stone: "#b9a37a", stoneDark: "#7c6a47",
};
const MAP = [
  "ffffffffffff", "fggggggggwwf", "fggggggggwwf", "fgggggggggwf",
  "fggggCgggggf", "fgggggggggggf".slice(0, 12), "fgggggrggggf", "fgggggrggggf",
  "fgggggrggggf", "fggMggrggggf", "ffggggggggff", "ffffffrfffff",
];

/* resource nodes: kind -> resource key */
const NODE_RES = { gold: "features", wood: "prompts", stone: "hardening", mana: "logs" };

const G = {
  state: "ready",          // ready | playing | won | lost
  units: [], enemies: [], nodes: [], log: [],
  res: { features: 0, prompts: 0, hardening: 0, logs: 0, trust: 60 },
  phase: 0,
  castle: { tx: 5, ty: 4, hp: 100, maxhp: 100 },
  mine: { tx: 3, ty: 9 },
  selected: null,
  clock: 0, demoIdx: 0, done: false, grace: 9,
  warlordSeen: false, warlordDead: false,
  dt: 0, last: 0,
};

let canvas, ctx;
const $ = (s) => document.querySelector(s);
const el = (t, c, x) => { const e = document.createElement(t); if (c) e.className = c; if (x != null) e.textContent = String(x); return e; };
const clamp = (v, a, b) => Math.max(a, Math.min(b, v));

/* ---- iso math -------------------------------------------- */
const isoX = (tx, ty) => ORIGIN_X + (tx - ty) * (TW / 2);
const isoY = (tx, ty) => ORIGIN_Y + (tx + ty) * (TH / 2);
function screenToTile(sx, sy) {
  const a = (sx - ORIGIN_X) / (TW / 2), b = (sy - ORIGIN_Y) / (TH / 2);
  return { tx: Math.round((a + b) / 2 - 0.5), ty: Math.round((b - a) / 2 - 0.5) };
}
const terrainAt = (tx, ty) => (ty < 0 || ty >= GH || tx < 0 || tx >= GW) ? "f" : (MAP[ty][tx] || "g");
const walkable = (tx, ty) => { const t = terrainAt(tx, ty); return t === "g" || t === "r"; };

/* ---- drawing primitives ---------------------------------- */
function px(x, y, w, h, c) { ctx.fillStyle = c; ctx.fillRect(x | 0, y | 0, w | 0, h | 0); }
function diamond(cx, cy, top) {
  ctx.fillStyle = top; ctx.beginPath();
  ctx.moveTo(cx, cy - TH / 2); ctx.lineTo(cx + TW / 2, cy); ctx.lineTo(cx, cy + TH / 2); ctx.lineTo(cx - TW / 2, cy);
  ctx.closePath(); ctx.fill();
}
function tileBlock(tx, ty, top, left, right) {
  const cx = isoX(tx, ty), cy = isoY(tx, ty);
  ctx.fillStyle = left; ctx.beginPath(); ctx.moveTo(cx - TW / 2, cy); ctx.lineTo(cx, cy + TH / 2); ctx.lineTo(cx, cy + TH / 2 + TZ); ctx.lineTo(cx - TW / 2, cy + TZ); ctx.closePath(); ctx.fill();
  ctx.fillStyle = right; ctx.beginPath(); ctx.moveTo(cx + TW / 2, cy); ctx.lineTo(cx, cy + TH / 2); ctx.lineTo(cx, cy + TH / 2 + TZ); ctx.lineTo(cx + TW / 2, cy + TZ); ctx.closePath(); ctx.fill();
  diamond(cx, cy, top);
}
function bar(x, y, w, frac, color) { px(x, y, w, 5, "#14110d"); px(x + 1, y + 1, (w - 2) * clamp(frac, 0, 1), 3, color); }

function drawTerrain() {
  for (let ty = 0; ty < GH; ty++) for (let tx = 0; tx < GW; tx++) {
    const t = terrainAt(tx, ty);
    if (t === "w") tileBlock(tx, ty, PAL.waterTop, PAL.waterL, PAL.waterR);
    else if (t === "f") { tileBlock(tx, ty, PAL.forestTop, PAL.forestL, PAL.forestR); drawTree(tx, ty); }
    else if (t === "r") tileBlock(tx, ty, PAL.roadTop, PAL.roadL, PAL.roadR);
    else tileBlock(tx, ty, PAL.grassTop, PAL.grassL, PAL.grassR);
  }
}
function drawTree(tx, ty) {
  const cx = isoX(tx, ty), cy = isoY(tx, ty);
  px(cx - 2, cy - 16, 4, 12, "#5a3a1c");
  ctx.fillStyle = "#244a1e"; ctx.beginPath(); ctx.ellipse(cx, cy - 18, 10, 9, 0, 0, 7); ctx.fill();
  ctx.fillStyle = "#2f5a26"; ctx.beginPath(); ctx.ellipse(cx - 2, cy - 20, 6, 5, 0, 0, 7); ctx.fill();
}
function drawCastle() {
  const cx = isoX(G.castle.tx, G.castle.ty), cy = isoY(G.castle.tx, G.castle.ty);
  const lvl = G.phase; // taller with progress
  px(cx - 16, cy - 26 - lvl * 2, 32, 26 + lvl * 2, PAL.stoneDark);
  px(cx - 14, cy - 24 - lvl * 2, 28, 22 + lvl * 2, PAL.stone);
  px(cx - 18, cy - 34 - lvl * 2, 36, 10, PAL.blood);
  px(cx - 4, cy - 12, 8, 12, "#3a2a18");
  px(cx + 12, cy - 42 - lvl * 2, 2, 12, "#5a3a1c"); px(cx + 6, cy - 42 - lvl * 2, 6, 4, PAL.gold);
  bar(cx - 16, cy - 40 - lvl * 2, 32, G.castle.hp / G.castle.maxhp, PAL.forestBright);
}
function drawNode(n) {
  const cx = isoX(n.tx, n.ty), cy = isoY(n.tx, n.ty);
  if (n.kind === "gold") { px(cx - 8, cy - 6, 16, 6, "#6e5230"); px(cx - 5, cy - 9, 4, 4, PAL.gold); px(cx + 1, cy - 9, 4, 4, PAL.gold); }
  else if (n.kind === "stone") { px(cx - 9, cy - 10, 18, 10, "#7c8a8f"); px(cx - 7, cy - 14, 12, 6, "#aab6ba"); }
  else if (n.kind === "mana") { ctx.fillStyle = "#6fb3e0"; ctx.beginPath(); ctx.ellipse(cx, cy - 8, 8, 5, 0, 0, 7); ctx.fill(); px(cx - 1, cy - 16, 2, 8, "#8fd0ff"); }
  else { px(cx - 3, cy - 14, 6, 12, "#5a3a1c"); ctx.fillStyle = PAL.forestBright; ctx.beginPath(); ctx.ellipse(cx, cy - 16, 9, 8, 0, 0, 7); ctx.fill(); }
}
// Human agent — armoured footman in role colour.
function drawHuman(a) {
  const cx = a.px, cy = a.py;
  ctx.fillStyle = "rgba(0,0,0,.35)"; ctx.beginPath(); ctx.ellipse(cx, cy + 2, 9, 4, 0, 0, 7); ctx.fill();
  px(cx - 4, cy - 6, 3, 6, "#3a2a18"); px(cx + 1, cy - 6, 3, 6, "#3a2a18");
  px(cx - 4, cy - 1, 3, 2, "#241a10"); px(cx + 1, cy - 1, 3, 2, "#241a10");
  px(cx - 5, cy - 15, 10, 9, a.color); px(cx - 5, cy - 15, 10, 2, "#ffffff55");
  px(cx - 5, cy - 7, 10, 2, "#5a3a1c");
  px(cx - 7, cy - 14, 2, 7, "#e0a878"); px(cx + 5, cy - 14, 2, 7, "#e0a878");
  px(cx - 3, cy - 21, 6, 6, "#e0a878");
  px(cx - 4, cy - 23, 8, 3, "#c9d2dc"); px(cx - 4, cy - 21, 1, 2, "#c9d2dc"); px(cx + 3, cy - 21, 1, 2, "#c9d2dc");
  px(cx - 2, cy - 19, 1, 1, "#2a1a12"); px(cx + 1, cy - 19, 1, 1, "#2a1a12");
  px(cx + 6, cy - 21, 1, 10, "#dfe6ee"); px(cx + 5, cy - 12, 3, 1, "#8a6a3a");
  ctx.fillStyle = a.color; ctx.font = "8px monospace"; ctx.textAlign = "center"; ctx.fillText(a.glyph, cx, cy - 25);
  bar(cx - 10, cy - 33, 20, a.hp / a.maxhp, PAL.forestBright);
  if (G.selected === a) { ctx.strokeStyle = PAL.forestBright; ctx.lineWidth = 2; ctx.beginPath(); ctx.ellipse(cx, cy + 2, 13, 6, 0, 0, 7); ctx.stroke(); }
}
// Orc threat — green, scaled by type size; boss = warlord.
function drawOrc(e) {
  const cx = e.px, cy = e.py, s = e.size, green = e.color, dark = "#243f18", cve = e.cve;
  ctx.fillStyle = "rgba(0,0,0,.4)"; ctx.beginPath(); ctx.ellipse(cx, cy + 2, 10 * s, 5 * s, 0, 0, 7); ctx.fill();
  const w = 12 * s, hh = 10 * s;
  px(cx - 4 * s, cy - 6 * s, 3 * s, 6 * s, dark); px(cx + 1 * s, cy - 6 * s, 3 * s, 6 * s, dark);
  px(cx - w / 2, cy - 15 * s, w, hh, green); px(cx - w / 2, cy - 15 * s, w, 2, "#86c25a55");
  px(cx - w / 2, cy - 7 * s, w, 2 * s, cve ? "#7a1f15" : "#3a2a18");
  px(cx - (w / 2 + 2), cy - 14 * s, 3 * s, 8 * s, green); px(cx + (w / 2 - 1), cy - 14 * s, 3 * s, 8 * s, green);
  const hw = 8 * s;
  px(cx - hw / 2, cy - 22 * s, hw, 7 * s, green); px(cx - hw / 2, cy - 22 * s, hw, 1, dark);
  px(cx - 3 * s, cy - 20 * s, 1, 1, "#ff4422"); px(cx + 2 * s, cy - 20 * s, 1, 1, "#ff4422");
  px(cx - 3 * s, cy - 16 * s, 1, 2 * s, "#f4eddc"); px(cx + 2 * s, cy - 16 * s, 1, 2 * s, "#f4eddc");
  if (cve) px(cx - hw / 2, cy - 19 * s, hw, 1, "#c0392b");
  px(cx + (w / 2 + 1), cy - 18 * s, 2 * s, 9 * s, "#5a3a1c"); px(cx + (w / 2), cy - 19 * s, 4 * s, 3 * s, "#7c6a47");
  bar(cx - 12 * s, cy - 30 * s, 24 * s, e.hp / e.maxhp, PAL.blood);
}

function render() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  drawTerrain();
  const sp = [];
  sp.push({ d: G.castle.tx + G.castle.ty, f: drawCastle });
  sp.push({ d: G.mine.tx + G.mine.ty, f: () => { const cx = isoX(G.mine.tx, G.mine.ty), cy = isoY(G.mine.tx, G.mine.ty); px(cx - 14, cy - 18, 28, 18, "#4a4a4a"); px(cx - 12, cy - 16, 24, 14, "#6b6b6b"); px(cx - 5, cy - 10, 10, 10, "#14110d"); } });
  for (const n of G.nodes) sp.push({ d: n.tx + n.ty, f: () => drawNode(n) });
  for (const a of G.units) sp.push({ d: a.tx + a.ty + 0.5, y: a.py, f: () => drawHuman(a) });
  for (const e of G.enemies) sp.push({ d: e.tx + e.ty + 0.5, y: e.py, f: () => drawOrc(e) });
  sp.sort((p, q) => (p.d - q.d) || ((p.y || 0) - (q.y || 0)));
  for (const s of sp) s.f();
}

/* ---- spawning -------------------------------------------- */
function placeNodes() {
  G.nodes = [
    { kind: "gold", tx: 9, ty: 3 }, { kind: "wood", tx: 2, ty: 2 },
    { kind: "stone", tx: 8, ty: 8 }, { kind: "mana", tx: 3, ty: 9 },
  ];
}
function nodeOf(kind) { return G.nodes.find((n) => n.kind === kind); }
function spawnUnit(type) {
  const d = UNIT_TYPES[type]; if (!d) return;
  const tx = clamp(G.castle.tx + (G.units.length % 3) - 1, 1, GW - 2);
  const ty = clamp(G.castle.ty + 1 + (G.units.length % 2), 1, GH - 2);
  const u = { type, ...d, hp: 100, maxhp: 100, tx, ty, px: isoX(tx, ty), py: isoY(tx, ty), target: null };
  G.units.push(u);
  return u;
}
function spawnEnemy(type) {
  const d = ENEMY_TYPES[type]; if (!d) return;
  const edges = [{ tx: 0, ty: 0 }, { tx: GW - 1, ty: 0 }, { tx: 0, ty: GH - 1 }, { tx: GW - 1, ty: GH - 1 }];
  const e0 = edges[(G.enemies.length + (d.boss ? 2 : 0)) % edges.length];
  const e = { type, ...d, maxhp: d.hp, tx: e0.tx, ty: e0.ty, px: isoX(e0.tx, e0.ty), py: isoY(e0.tx, e0.ty) };
  G.enemies.push(e);
  if (d.boss) G.warlordSeen = true;
  return e;
}

/* ---- log -> visual transform ----------------------------- */
function logMsg(msg, sev) { G.log.unshift({ msg, sev }); if (G.log.length > 40) G.log.pop(); renderEventLog(); }
function applyLog(ev) {
  G.phase = Math.max(G.phase, ev.phase || 0);
  const spawn = LOG_SPAWN[ev.type];
  if (spawn === "horde") { for (let i = 0; i < 4; i++) spawnEnemy("bug_grunt"); spawnEnemy("incident_ogre"); }
  else if (spawn && spawn.startsWith("unit:")) spawnUnit(spawn.slice(5));
  else if (spawn && spawn.startsWith("enemy:")) spawnEnemy(spawn.slice(6));
  logMsg(ev.msg, ev.severity);
}

/* ---- simulation ------------------------------------------ */
function moveToward(a, tx, ty, speed) {
  const dx = isoX(tx, ty) - a.px, dy = isoY(tx, ty) - a.py, dist = Math.hypot(dx, dy);
  if (dist < 3) { a.px = isoX(tx, ty); a.py = isoY(tx, ty); a.tx = tx; a.ty = ty; return true; }
  const v = speed * 34 * G.dt; a.px += dx / dist * v; a.py += dy / dist * v;
  return false;
}
// Pick a threat within range, focus-firing bosses then CVEs then nearest.
function pickThreat(a, range) {
  let best = null, bs = -1;
  for (const e of G.enemies) {
    const d = Math.hypot(e.px - a.px, e.py - a.py);
    if (d > range) continue;
    const pr = (e.boss ? 3 : e.cve ? 2 : 1) - d / 1000;
    if (pr > bs) { bs = pr; best = e; }
  }
  return best;
}
function update(dt) {
  G.dt = dt;
  if (G.state !== "playing") return;
  G.clock += dt;

  // play demo logs by timestamp
  while (G.demoIdx < DEMO_LOGS.length && DEMO_LOGS[G.demoIdx].t <= G.clock) applyLog(DEMO_LOGS[G.demoIdx++]);
  if (G.demoIdx >= DEMO_LOGS.length) G.done = true;

  // units behave by job; everyone defends if a threat is close
  for (const a of G.units) {
    if (a.hp <= 0) continue;
    const threat = pickThreat(a, a.job === "defend" ? 1e9 : 260);
    if (threat) {
      if (moveToward(a, threat.tx, threat.ty, 1.8)) {
        threat.hp -= 36 * dt;
        if (threat.hp <= 0) killEnemy(threat);
      }
      continue;
    }
    if (NODE_RES[a.job]) {
      const n = nodeOf(a.job);
      if (n && moveToward(a, n.tx, n.ty, 1.4)) G.res[NODE_RES[a.job]] += 6 * dt;
    } else if (a.job === "build") {
      if (moveToward(a, G.castle.tx, G.castle.ty + 1, 1.3)) G.castle.hp = Math.min(G.castle.maxhp, G.castle.hp + 4 * dt);
    } else { // release / idle → rally near castle
      moveToward(a, G.castle.tx + 1, G.castle.ty + 1, 1.0);
    }
  }
  G.units = G.units.filter((a) => a.hp > 0);

  // enemies march on the castle
  for (const e of G.enemies) {
    if (moveToward(e, G.castle.tx, G.castle.ty, e.speed)) {
      // reaching the wall hurts the citadel; the boss hits hardest, but the
      // outcome is decided by the battle (castle HP / trust), not an instakill.
      G.castle.hp -= e.dmg * dt * (e.boss ? 0.9 : 0.5);
      G.res.trust = clamp(G.res.trust - e.dmg * dt * 0.12, 0, 100);
    }
  }

  // passive trust from delivery
  G.res.trust = clamp(G.res.trust + (G.units.length && !G.enemies.length ? 2 * dt : 0), 0, 100);

  evaluateEnd(dt);
  syncPanels();
}
function killEnemy(e) {
  G.enemies = G.enemies.filter((x) => x !== e);
  G.res.trust = clamp(G.res.trust + e.trust, 0, 100);
  if (e.boss) G.warlordDead = true;
  logMsg("Slain: " + e.label + " (+%" + e.trust + " trust)");
}
function evaluateEnd(dt) {
  if (G.state !== "playing") return;
  if (G.castle.hp <= 0) { G.state = "lost"; G.result = "Tech debt overran the citadel."; return; }
  if (G.res.trust < 30) { G.state = "lost"; G.result = "Trust collapsed below 30%."; return; }
  if (G.done) {
    G.grace -= dt;
    const settled = G.enemies.filter((e) => e.cve || e.boss).length === 0;
    if ((G.warlordSeen ? G.warlordDead : true) && settled && G.grace < 7) {
      if (G.res.trust >= 70 && G.castle.hp > 0) { G.state = "won"; G.result = "Release shipped. Trust " + Math.round(G.res.trust) + "%, citadel " + Math.round(G.castle.hp) + "%."; }
    }
    if (G.grace <= 0 && G.state === "playing") {
      G.state = G.res.trust >= 70 && G.castle.hp > 0 ? "won" : "lost";
      G.result = G.state === "won" ? "Release shipped at the buzzer." : "The release slipped — trust too low.";
    }
  }
}
function loop(ts) {
  if (!G.last) G.last = ts;
  const dt = Math.min(0.05, (ts - G.last) / 1000); G.last = ts;
  update(dt); render();
  if (G.state === "won" || G.state === "lost") return showModal();
  requestAnimationFrame(loop);
}

/* ---- panels ---------------------------------------------- */
function syncPanels() {
  $("#r-features").textContent = Math.floor(G.res.features);
  $("#r-prompts").textContent = Math.floor(G.res.prompts);
  $("#r-hardening").textContent = Math.floor(G.res.hardening);
  $("#r-logs").textContent = Math.floor(G.res.logs);
  const tr = $("#r-trust"); tr.textContent = Math.round(G.res.trust) + "%";
  tr.style.color = G.res.trust >= 70 ? PAL.forestBright : G.res.trust < 40 ? PAL.blood : PAL.gold;
  $("#c-hp").textContent = Math.max(0, Math.round(G.castle.hp)) + "%";
  // timeline
  const tl = $("#timeline"); tl.replaceChildren();
  PHASES.forEach((p, i) => tl.appendChild(el("span", "phase" + (i <= G.phase ? " on" : ""), p.key)));
  // unit panel (counts by type)
  const up = $("#unitpanel"); up.replaceChildren();
  const byType = {}; for (const u of G.units) byType[u.type] = (byType[u.type] || 0) + 1;
  for (const t of Object.keys(byType)) {
    const d = UNIT_TYPES[t]; const row = el("div", "mini");
    const sw = el("span", "sw", d.glyph); sw.style.background = d.color; row.appendChild(sw);
    row.appendChild(el("small", null, d.label + " ×" + byType[t] + " — " + d.sdlc)); up.appendChild(row);
  }
  if (!G.units.length) up.appendChild(el("p", "muted", "No agents yet."));
  // threat panel
  const tp = $("#threatpanel"); tp.replaceChildren();
  const byT = {}; for (const e of G.enemies) byT[e.type] = (byT[e.type] || 0) + 1;
  for (const t of Object.keys(byT)) { const d = ENEMY_TYPES[t]; const row = el("div", "mini"); const sw = el("span", "sw", "☠"); sw.style.background = d.color; row.appendChild(sw); row.appendChild(el("small", null, d.label + " ×" + byT[t])); tp.appendChild(row); }
  if (!G.enemies.length) tp.appendChild(el("p", "muted", "Perimeter clear."));
}
function renderEventLog() {
  const box = $("#eventlog"); box.replaceChildren();
  for (const e of G.log) { const li = el("div", "ev" + (e.sev ? " sev-" + e.sev : "")); li.textContent = "» " + e.msg; box.appendChild(li); }
}
function showModal() {
  render();
  const m = $("#modal"); m.style.display = "grid"; m.replaceChildren();
  const card = el("div", "modal-card");
  card.appendChild(el("h2", null, G.state === "won" ? "RELEASE SHIPPED" : "PROJECT LOST"));
  card.appendChild(el("p", null, G.result || ""));
  const stats = el("ul", "stats");
  stats.appendChild(el("li", null, "Features delivered: " + Math.floor(G.res.features)));
  stats.appendChild(el("li", null, "Prompts hardened: " + Math.floor(G.res.prompts)));
  stats.appendChild(el("li", null, "Hardening laid: " + Math.floor(G.res.hardening)));
  stats.appendChild(el("li", null, "Threats remaining: " + G.enemies.length));
  stats.appendChild(el("li", null, "Final Trust: " + Math.round(G.res.trust) + "%  ·  Citadel " + Math.max(0, Math.round(G.castle.hp)) + "%"));
  card.appendChild(stats);
  const btn = el("button", "btn", "⟲ Run again"); btn.addEventListener("click", () => location.reload());
  card.appendChild(btn); m.appendChild(card);
}

/* ---- input ----------------------------------------------- */
function onClick(e) {
  const rect = canvas.getBoundingClientRect();
  const sx = (e.clientX - rect.left) / rect.width * canvas.width, sy = (e.clientY - rect.top) / rect.height * canvas.height;
  let hit = null, hd = 18;
  for (const a of G.units) { const d = Math.hypot(a.px - sx, a.py - sy); if (d < hd) { hd = d; hit = a; } }
  G.selected = hit;
  const ins = $("#inspect"); ins.replaceChildren(el("div", "panel-title", "⚑ Inspect"));
  if (hit) {
    const b = el("p"); const n = el("b", null, hit.label); n.style.color = hit.color; b.appendChild(n); ins.appendChild(b);
    ins.appendChild(el("p", "muted", "SDLC role")); ins.appendChild(el("p", null, hit.sdlc));
    ins.appendChild(el("p", "hint", "Job: " + hit.job + (NODE_RES[hit.job] ? " → " + NODE_RES[hit.job] : "")));
  } else ins.appendChild(el("p", "muted", "Click an agent to inspect."));
}

/* ---- boot ------------------------------------------------ */
function resize() { const w = canvas.clientWidth; canvas.width = w; canvas.height = Math.round(w * 0.6); ORIGIN_X = canvas.width / 2; ORIGIN_Y = 56; }
function startDemo(seedArchive) {
  placeNodes();
  // seed a builder + any agents implied by a decoded archive's domains
  spawnUnit("builder");
  if (seedArchive && Array.isArray(seedArchive.sessions)) {
    const map = { etl: "feature_worker", api: "hardening_engineer", graphql: "prompt_smith", data_quality: "hardening_engineer", bugs: "incident_defender", prompts: "prompt_smith" };
    const seen = new Set();
    for (const s of seedArchive.sessions) { const u = map[String(s.domain || "").toLowerCase()]; if (u && !seen.has(u)) { seen.add(u); spawnUnit(u); } }
  }
  G.state = "playing"; G.clock = 0; G.demoIdx = 0; G.last = 0;
  $("#start").style.display = "none";
  logMsg("Builder Agent raised the first wall of architecture.");
  syncPanels(); requestAnimationFrame(loop);
}

document.addEventListener("DOMContentLoaded", () => {
  canvas = $("#field"); ctx = canvas.getContext("2d"); ctx.imageSmoothingEnabled = false;
  resize(); window.addEventListener("resize", () => { resize(); if (G.state !== "playing") render(); });
  canvas.addEventListener("click", onClick);
  placeNodes(); render(); syncPanels();
  let archive = null; try { const s = sessionStorage.getItem("bqa-archive"); if (s) archive = JSON.parse(s); } catch (_) {}
  $("#start-demo").addEventListener("click", () => startDemo(archive));
});
