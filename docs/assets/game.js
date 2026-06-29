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
const SURVIVAL_STAGE_WAVES = 3;

const G = {
  state: "ready",          // ready | playing | won | lost
  units: [], enemies: [], nodes: [], log: [], mcps: [], projectiles: [],
  res: { features: 0, prompts: 0, hardening: 0, logs: 0, trust: 60 },
  skills: { attack: 0, gather: 0 },   // in-run global upgrades
  phase: 0,
  castle: { tx: 5, ty: 4, hp: 100, maxhp: 100 },
  mine: { tx: 3, ty: 9 },
  selected: null,
  clock: 0, demoIdx: 0, done: false, grace: 9,
  warlordSeen: false, warlordDead: false,
  mode: "demo",            // demo | survival
  wave: 0, waveGap: 0, wavesSurvived: 0, recorded: false,
  playerLevel: 1, unlockedStage: 1, stage: 1, startWave: 1,
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
    else if (t === "r") { tileBlock(tx, ty, PAL.roadTop, PAL.roadL, PAL.roadR); drawBricks(tx, ty); }
    else tileBlock(tx, ty, PAL.grassTop, PAL.grassL, PAL.grassR);
  }
}
function drawBricks(tx, ty) {
  const cx = isoX(tx, ty), cy = isoY(tx, ty);
  for (let r = -1; r <= 1; r++) { px(cx - 9, cy + r * 5, 7, 1, "#6e5230"); px(cx + 2, cy + r * 5 + 2, 6, 1, "#6e5230"); }
}
function drawTree(tx, ty) {
  const cx = isoX(tx, ty), cy = isoY(tx, ty);
  const t = (ox, oy, r) => {
    px(cx + ox - 1, cy + oy - 2, 3, 9, "#5a3a1c");
    ctx.fillStyle = "#244a1e"; ctx.beginPath(); ctx.ellipse(cx + ox, cy + oy - 6, r, r - 1, 0, 0, 7); ctx.fill();
    ctx.fillStyle = "#2f5a26"; ctx.beginPath(); ctx.ellipse(cx + ox - 1, cy + oy - 8, r - 3, r - 4, 0, 0, 7); ctx.fill();
  };
  t(-5, 3, 6); t(5, 1, 7); t(0, -3, 8);   // a little cluster
}
function drawCastle() {
  const cx = isoX(G.castle.tx, G.castle.ty), cy = isoY(G.castle.tx, G.castle.ty);
  const lvl = G.phase; // taller with progress
  px(cx - 16, cy - 26 - lvl * 2, 32, 26 + lvl * 2, PAL.stoneDark);
  px(cx - 14, cy - 24 - lvl * 2, 28, 22 + lvl * 2, PAL.stone);
  for (let i = -14; i <= 11; i += 6) px(cx + i, cy - 30 - lvl * 2, 4, 4, PAL.stoneDark); // crenellations
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
  if (a.drift > 0) { ctx.fillStyle = "#c79bff"; ctx.font = "10px monospace"; ctx.textAlign = "center"; ctx.fillText("≈", cx, cy - 36); px(cx - 6, cy - 28, 12, 1, "#8e6fc4"); }
  if (G.selected === a) { ctx.strokeStyle = "#7bff5a"; ctx.lineWidth = 1.5; ctx.strokeRect(cx - 9, cy - 26, 18, 30); }
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
  for (const m of G.mcps) sp.push({ d: m.tx + m.ty, f: () => drawMCP(m) });
  for (const a of G.units) sp.push({ d: a.tx + a.ty + 0.5, y: a.py, f: () => drawHuman(a) });
  for (const e of G.enemies) sp.push({ d: e.tx + e.ty + 0.5, y: e.py, f: () => drawOrc(e) });
  sp.sort((p, q) => (p.d - q.d) || ((p.y || 0) - (q.y || 0)));
  for (const s of sp) s.f();
  drawProjectiles();
  drawMinimap();
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
  const u = { type, ...d, hp: 120, maxhp: 120, tx, ty, px: isoX(tx, ty), py: isoY(tx, ty), target: null };
  G.units.push(u);
  return u;
}
function spawnEnemy(type) {
  const d = ENEMY_TYPES[type]; if (!d) return;
  const edges = [{ tx: 0, ty: 0 }, { tx: GW - 1, ty: 0 }, { tx: 0, ty: GH - 1 }, { tx: GW - 1, ty: GH - 1 }];
  const e0 = edges[(G.enemies.length + (d.boss ? 2 : 0)) % edges.length];
  const stats = scaledEnemyStats(d);
  const e = { type, ...d, hp: stats.hp, maxhp: stats.hp, dmg: stats.dmg, speed: stats.speed, tx: e0.tx, ty: e0.ty, px: isoX(e0.tx, e0.ty), py: isoY(e0.tx, e0.ty) };
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
/* ---- ranged combat (projectiles) ------------------------- */
function nearestUnit(e) { let best = null, bd = 150; for (const u of G.units) { const d = Math.hypot(u.px - e.px, u.py - e.py); if (d < bd) { bd = d; best = u; } } return best; }
function shoot(from, target, dmg, color, fromEnemy) {
  G.projectiles.push({ px: from.px, py: from.py - 10, target, dmg, color, fromEnemy });
  sfx("command");
}
function projAim(p) {
  if (p.target && p.target.castle) return [isoX(G.castle.tx, G.castle.ty), isoY(G.castle.tx, G.castle.ty) - 16];
  if (p.target) return [p.target.px, p.target.py - 10];
  return null;
}
function projAlive(p) {
  if (p.target && p.target.castle) return true;
  return p.fromEnemy ? G.units.includes(p.target) : G.enemies.includes(p.target);
}
function updateProjectiles(dt) {
  const keep = [];
  for (const p of G.projectiles) {
    if (!projAlive(p)) continue;                 // target gone → fizzle
    const aim = projAim(p), dx = aim[0] - p.px, dy = aim[1] - p.py, d = Math.hypot(dx, dy);
    if (d < 8) {
      if (p.target.castle) { G.castle.hp -= p.dmg * 0.6; G.res.trust = clamp(G.res.trust - p.dmg * 0.1, 0, 100); }
      else if (p.fromEnemy) { p.target.hp -= p.dmg; }
      else { p.target.hp -= p.dmg; if (p.target.hp <= 0) killEnemy(p.target); }
      continue;
    }
    const v = 270 * dt; p.px += dx / d * v; p.py += dy / d * v; keep.push(p);
  }
  G.projectiles = keep;
}
function drawProjectiles() { for (const p of G.projectiles) { px(p.px - 1, p.py - 1, 3, 3, "#14110d"); px(p.px - 1, p.py - 1, 2, 2, p.color); } }

/* Model Drift: a CVE Shaman hexes the nearest agents so they miss for 3s. */
function castDrift(e) {
  const sorted = G.units.slice().sort((a, b) => Math.hypot(a.px - e.px, a.py - e.py) - Math.hypot(b.px - e.px, b.py - e.py));
  const n = Math.min(3, sorted.length);
  for (let i = 0; i < n; i++) sorted[i].drift = 3;
  if (n) logMsg("CVE Shaman casts Model Drift — " + n + " agent" + (n === 1 ? "" : "s") + " miss for 3s.", "high");
  sfx("kill");
}
/* Clear Context: the player's counter-spell — purge drift from all agents. */
function clearContext() {
  let n = 0; for (const a of G.units) if (a.drift > 0) { a.drift = 0; n++; }
  logMsg("Clear Context — drift purged from " + n + " agent" + (n === 1 ? "" : "s") + ".");
  sfx("win");
}

function update(dt) {
  G.dt = dt;
  if (G.state !== "playing") return;
  G.clock += dt;

  if (G.mode === "demo") {
    // play demo logs by timestamp
    while (G.demoIdx < DEMO_LOGS.length && DEMO_LOGS[G.demoIdx].t <= G.clock) applyLog(DEMO_LOGS[G.demoIdx++]);
    if (G.demoIdx >= DEMO_LOGS.length) G.done = true;
  } else {
    // survival: endless escalating waves; clear one to face the next
    G.waveGap -= dt;
    if (!G.enemies.length && G.waveGap <= 0) spawnSurvivalWave();
  }

  // units behave by job; everyone defends if a threat is close
  for (const a of G.units) {
    if (a.hp <= 0) continue;
    if (a.drift > 0) a.drift -= dt;             // Model Drift wears off
    const drifting = a.drift > 0;               // while drifted, attacks miss
    const threat = pickThreat(a, a.job === "defend" ? 1e9 : 260);
    if (threat) {
      const dist = Math.hypot(threat.px - a.px, threat.py - a.py);
      if (a.ranged && dist <= (a.range || 130)) {
        a.cd = (a.cd || 0) - dt;
        if (a.cd <= 0) { if (!drifting) shoot(a, threat, (a.dmg || 20) + G.skills.attack * 8, "#eaf3ff", false); a.cd = 0.8; }
      } else if (moveToward(a, threat.tx, threat.ty, 1.8)) {
        if (!drifting) { threat.hp -= (36 + G.skills.attack * 12) * dt; if (threat.hp <= 0) killEnemy(threat); }
      }
      continue;
    }
    if (NODE_RES[a.job]) {
      const n = nodeOf(a.job);
      if (n && moveToward(a, n.tx, n.ty, 1.4)) G.res[NODE_RES[a.job]] += (6 + G.skills.gather * 3) * dt;
    } else if (a.job === "build") {
      if (moveToward(a, G.castle.tx, G.castle.ty + 1, 1.3)) G.castle.hp = Math.min(G.castle.maxhp, G.castle.hp + 4 * dt);
    } else { // release / idle → rally near castle
      moveToward(a, G.castle.tx + 1, G.castle.ty + 1, 1.0);
    }
  }
  G.units = G.units.filter((a) => a.hp > 0);

  // MCP relays: passive observability — signal trickle + light castle repair
  for (const m of G.mcps) { G.res.logs += 1.2 * dt; G.castle.hp = Math.min(G.castle.maxhp, G.castle.hp + 0.6 * dt); }

  // enemies march on the castle (ranged ones stop and pelt it / nearby units)
  for (const e of G.enemies) {
    if (e.cve && !e.boss) {                      // CVE Shaman — caster of Model Drift
      e.castCd = (e.castCd == null ? 4 : e.castCd) - dt;
      if (e.castCd <= 0 && G.units.length) { castDrift(e); e.castCd = 7; }
      moveToward(e, G.castle.tx, G.castle.ty, e.speed * 0.7);
      continue;
    }
    if (e.ranged) {
      const tgt = nearestUnit(e);
      const aimX = tgt ? tgt.px : isoX(G.castle.tx, G.castle.ty);
      const aimY = tgt ? tgt.py : isoY(G.castle.tx, G.castle.ty);
      if (Math.hypot(aimX - e.px, aimY - e.py) <= (e.range || 120)) {
        e.cd = (e.cd || 0) - dt;
        if (e.cd <= 0) { shoot(e, tgt || { castle: true }, e.dmg, "#caa05a", true); e.cd = 1.1; }
      } else moveToward(e, G.castle.tx, G.castle.ty, e.speed);
      continue;
    }
    if (moveToward(e, G.castle.tx, G.castle.ty, e.speed)) {
      // reaching the wall hurts the citadel; the boss hits hardest, but the
      // outcome is decided by the battle (castle HP / trust), not an instakill.
      G.castle.hp -= e.dmg * dt * (e.boss ? 0.9 : 0.5);
      G.res.trust = clamp(G.res.trust - e.dmg * dt * 0.12, 0, 100);
    }
  }

  // passive trust from delivery
  G.res.trust = clamp(G.res.trust + (G.units.length && !G.enemies.length ? 2 * dt : 0), 0, 100);

  // survival base income — keeps the Command panel usable even mid-fight
  if (G.mode === "survival") {
    const k = 1 + G.phase * 0.25;
    G.res.features += 1.6 * k * dt; G.res.prompts += 1.2 * k * dt;
    G.res.hardening += 1.2 * k * dt; G.res.logs += 1.0 * k * dt;
  }
  updateProjectiles(dt);

  evaluateEnd(dt);
  syncPanels();
}
function killEnemy(e) {
  G.enemies = G.enemies.filter((x) => x !== e);
  G.res.trust = clamp(G.res.trust + e.trust, 0, 100);
  if (e.boss) G.warlordDead = true;
  logMsg("Slain: " + e.label + " (+%" + e.trust + " trust)");
  sfx("kill");
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
/* ---- survival mode --------------------------------------- */
const sfx = (n) => { try { if (window.AUDIO) window.AUDIO.play(n); } catch (_) {} };
function stageForWave(wave) {
  return Math.max(1, Math.floor((Math.max(1, wave) - 1) / SURVIVAL_STAGE_WAVES) + 1);
}
function firstWaveForStage(stage) {
  return (Math.max(1, stage) - 1) * SURVIVAL_STAGE_WAVES + 1;
}
function setSurvivalStageForWave(wave) {
  G.stage = stageForWave(wave);
  G.phase = Math.min(PHASES.length - 1, G.stage - 1);
}
function survivalDifficulty() {
  const stage = Math.max(1, G.stage || 1), level = Math.max(1, G.playerLevel || 1);
  // Difficulty scales from both current stage and persistent player level:
  // stage controls spawn density/unlocks/boss cadence, while level gives a
  // smaller enemy power multiplier so skipping ahead is progress, not free score.
  return {
    density: (stage - 1) * 0.9 + (level - 1) * 0.45,
    hp: 1 + (stage - 1) * 0.18 + (level - 1) * 0.08,
    dmg: 1 + (stage - 1) * 0.13 + (level - 1) * 0.06,
    speed: Math.min(1.45, 1 + (stage - 1) * 0.04 + (level - 1) * 0.02),
    bossEvery: Math.max(2, 4 - Math.min(2, Math.floor(Math.max(0, stage - 3) / 2))),
  };
}
function scaledEnemyStats(d) {
  if (G.mode !== "survival") return { hp: d.hp, dmg: d.dmg, speed: d.speed };
  const scale = survivalDifficulty();
  return {
    hp: Math.round(d.hp * scale.hp),
    dmg: Math.round(d.dmg * scale.dmg),
    speed: Number((d.speed * scale.speed).toFixed(2)),
  };
}
function applySurvivalStartStage() {
  G.unlockedStage = Math.max(1, G.playerLevel || 1);
  G.startWave = firstWaveForStage(G.unlockedStage);
  G.wave = G.startWave - 1;
  G.wavesSurvived = Math.max(0, G.wave);
  setSurvivalStageForWave(G.startWave);
  if (G.unlockedStage > 1) {
    logMsg("Player level " + G.playerLevel + " unlocks Stage " + G.unlockedStage + " — starting at wave " + G.startWave + ".", "high");
  }
}
function spawnSurvivalWave() {
  if (G.wave > 0) G.wavesSurvived = G.wave;        // previous wave cleared
  G.wave++;
  const w = G.wave;
  setSurvivalStageForWave(w);
  const scale = survivalDifficulty();
  const power = G.units.length;                       // your army size drives the threat
  const grunts = 1 + Math.floor(w * 0.55 + power * 0.4 + scale.density);
  for (let i = 0; i < grunts; i++) spawnEnemy(i % 4 === 3 ? "regression_raider" : "bug_grunt");
  if (w >= 2 && w % 2 === 0) spawnEnemy("spear_hurler");
  if (w >= 3) spawnEnemy("cve_shaman");
  if (w >= 4 && w % 2 === 0) spawnEnemy("incident_ogre");
  if (w >= 6 && w % 3 === 0) spawnEnemy("tech_debt_troll");
  if (w >= 8 && w % scale.bossEvery === 0) spawnEnemy("deadline_warlord");
  // a small free reinforcement keeps passive play viable; real growth = recruiting
  if (w % 2 === 1) spawnUnit(["incident_defender", "sentinel_archer"][(w >> 1) % 2]);
  G.waveGap = 2.5;
  logMsg("Stage " + G.stage + " · Wave " + w + " advances — army " + power + " vs " + grunts + "+ orcs.", w % 5 === 0 ? "high" : undefined);
  sfx("wave");
}

/* ---- leaderboard ----------------------------------------- */
// Official top-3 is git-tracked in docs/leaderboard.json and shown on launch.
// Personal bests live in localStorage. A new record that beats the official
// board is flagged for a maintainer/CI to commit into leaderboard.json.
const LB_KEY = "bqa-citadel-scores";
const OFFICIAL_DEFAULT = [
  { name: "Sir Buildsworth", waves: 12 }, { name: "Lady Hardena", waves: 9 }, { name: "Prompt-Smith Vee", waves: 7 },
];
G.official = OFFICIAL_DEFAULT.slice();
function loadOfficial() {
  try {
    fetch("leaderboard.json").then((r) => r.json()).then((d) => {
      if (d && Array.isArray(d.top)) { G.official = d.top.slice(0, 3); renderLeaderboard(document.getElementById("lb-start")); }
    }).catch(() => {});
  } catch (_) {}
}
function getScores() { try { return JSON.parse(localStorage.getItem(LB_KEY) || "[]"); } catch (_) { return []; } }
function recordScore(waves) {
  let name = "You";
  const beatsOfficial = waves > 0 && (G.official.length < 3 || waves > G.official[G.official.length - 1].waves);
  try {
    if (typeof window !== "undefined" && window.prompt) {
      const n = window.prompt((beatsOfficial ? "New top-3! " : "") + "You survived " + waves + " waves. Enter your name:", "");
      if (n) name = n.slice(0, 24);
    }
  } catch (_) {}
  try {
    const s = getScores();
    s.push({ name, waves, ts: new Date().toISOString().slice(0, 10) });
    s.sort((a, b) => b.waves - a.waves);
    localStorage.setItem(LB_KEY, JSON.stringify(s.slice(0, 10)));
  } catch (_) {}
  G.lastRecord = { name, waves, beatsOfficial };
}
function renderLeaderboard(box) {
  if (!box) return;
  box.replaceChildren();
  (G.official || []).slice(0, 3).forEach((r, i) => {
    const row = el("div", "lb-row");
    row.appendChild(el("span", "lb-rank", ["🥇", "🥈", "🥉"][i] || ("#" + (i + 1))));
    row.appendChild(el("span", "lb-waves", (r.name || "—") + " — " + r.waves + " waves"));
    box.appendChild(row);
  });
  const mine = getScores()[0];
  if (mine) { const r = el("div", "lb-row"); r.appendChild(el("span", "lb-rank", "you")); r.appendChild(el("span", "lb-waves", "Your best: " + mine.waves + " waves")); box.appendChild(r); }
}

/* ---- minimap --------------------------------------------- */
function drawMinimap() {
  const mc = document.getElementById("minimap"); if (!mc) return;
  const m = mc.getContext("2d"); const W = mc.width, H = mc.height;
  const cw = W / GW, ch = H / GH;
  for (let ty = 0; ty < GH; ty++) for (let tx = 0; tx < GW; tx++) {
    const t = terrainAt(tx, ty);
    m.fillStyle = t === "w" ? "#2f6090" : t === "f" ? "#244a1e" : t === "r" ? "#8a6a3a" : "#4a7a3a";
    m.fillRect(tx * cw, ty * ch, Math.ceil(cw), Math.ceil(ch));
  }
  m.fillStyle = "#f2c14e"; m.fillRect(G.castle.tx * cw - 1, G.castle.ty * ch - 1, cw + 2, ch + 2);
  m.fillStyle = "#7bc45a"; for (const a of G.units) m.fillRect(a.tx * cw, a.ty * ch, cw, ch);
  m.fillStyle = "#c0392b"; for (const e of G.enemies) m.fillRect(e.tx * cw, e.ty * ch, cw, ch);
}

/* ---- player control: recruit / MCP / skills -------------- */
const ECONOMY = {
  recruit: {
    feature_worker: { features: 8 }, prompt_smith: { prompts: 8 },
    hardening_engineer: { hardening: 8 }, incident_defender: { features: 6, hardening: 6 },
    context_logger: { logs: 6 }, sentinel_archer: { features: 7, hardening: 7 },
  },
  mcp: { logs: 12, hardening: 8 },
  skill: { attack: { features: 10, hardening: 10 }, gather: { prompts: 10, logs: 6 } },
  spell: { clear_context: { logs: 5, prompts: 5 } },
};
const RES_ICON = { features: "★", prompts: "✦", hardening: "⛨", logs: "≋" };
const COMMANDS = [
  { id: "feature_worker", label: "Feature Worker", glyph: "★", kind: "recruit" },
  { id: "prompt_smith", label: "Prompt Smith", glyph: "✦", kind: "recruit" },
  { id: "hardening_engineer", label: "Hardening Eng.", glyph: "⛨", kind: "recruit" },
  { id: "incident_defender", label: "Defender", glyph: "⚔", kind: "recruit" },
  { id: "sentinel_archer", label: "Archer (ranged)", glyph: "➹", kind: "recruit" },
  { id: "mcp", label: "MCP relay", glyph: "⛁", kind: "mcp" },
  { id: "attack", label: "+Attack", glyph: "⚒", kind: "skill" },
  { id: "gather", label: "+Gather", glyph: "⛏", kind: "skill" },
  { id: "clear_context", label: "Clear Context", glyph: "✺", kind: "spell" },
];
// Costs rise the more you already have, so growth is steady, never trivial,
// and the challenge scales proportionally with your army.
function costFor(c) {
  if (c.kind === "recruit") {
    const base = ECONOMY.recruit[c.id], n = G.units.filter((u) => u.type === c.id).length, out = {};
    for (const k in base) out[k] = Math.round(base[k] * (1 + 0.2 * n));
    return out;
  }
  if (c.kind === "mcp") { const n = G.mcps.length, out = {}; for (const k in ECONOMY.mcp) out[k] = Math.round(ECONOMY.mcp[k] * (1 + 0.35 * n)); return out; }
  if (c.kind === "spell") return ECONOMY.spell[c.id];
  const base = ECONOMY.skill[c.id], lvl = G.skills[c.id], out = {};
  for (const k in base) out[k] = base[k] * (lvl + 1);
  return out;
}
function costStr(cost) { return Object.keys(cost).map((k) => Math.round(cost[k]) + (RES_ICON[k] || k)).join(" "); }
function canAfford(cost) { return Object.keys(cost).every((k) => G.res[k] >= cost[k]); }
function spend(cost) { for (const k in cost) G.res[k] -= cost[k]; }
function addMCP() {
  const around = [[1, 0], [-1, 0], [2, 0], [-1, 1], [2, 1], [0, 2]];
  const used = new Set(G.mcps.map((m) => m.tx + "," + m.ty));
  for (const o of around) { const tx = G.castle.tx + o[0], ty = G.castle.ty + o[1] + 1; if (walkable(tx, ty) && !used.has(tx + "," + ty)) { G.mcps.push({ tx, ty }); break; } }
  logMsg("MCP relay online — observability up.");
}
function doCommand(c) {
  if (G.state !== "playing") return;
  const cost = costFor(c);
  if (!canAfford(cost)) { logMsg("Not enough resources for " + c.label + "."); return; }
  spend(cost); sfx("command");
  if (c.kind === "recruit") { spawnUnit(c.id); logMsg("Recruited " + UNIT_TYPES[c.id].label + "."); }
  else if (c.kind === "mcp") addMCP();
  else if (c.kind === "spell") { if (c.id === "clear_context") clearContext(); }
  else { G.skills[c.id]++; logMsg("Upgraded " + c.label + " to L" + G.skills[c.id] + "."); }
}
function buildCommands() {
  const box = document.getElementById("commands"); if (!box) return;
  box.replaceChildren();
  for (const c of COMMANDS) {
    const b = el("button", "cmd"); b.dataset.id = c.id;
    b.appendChild(el("span", "cmd-g", c.glyph));
    b.appendChild(el("small", "cmd-l", c.label));
    b.appendChild(el("span", "cmd-cost", costStr(costFor(c))));
    b.addEventListener("click", () => doCommand(c));
    box.appendChild(b);
  }
}
function updateCommands() {
  const box = document.getElementById("commands"); if (!box) return;
  for (const c of COMMANDS) {
    const b = box.querySelector('[data-id="' + c.id + '"]'); if (!b) continue;
    const cost = costFor(c);
    b.disabled = G.state !== "playing" || !canAfford(cost);
    const cs = b.querySelector(".cmd-cost"); if (cs) cs.textContent = costStr(cost);
  }
}
function drawMCP(m) {
  const cx = isoX(m.tx, m.ty), cy = isoY(m.tx, m.ty);
  px(cx - 5, cy - 18, 10, 18, "#33405e"); px(cx - 5, cy - 18, 10, 2, "#6fb3e0");
  px(cx - 3, cy - 12, 6, 4, "#8fd0ff"); px(cx - 1, cy - 24, 2, 7, "#6fb3e0"); px(cx - 4, cy - 25, 8, 2, "#8fd0ff");
}
function toggleFullscreen() {
  const elx = document.getElementById("stage"); if (!elx) return;
  if (!document.fullscreenElement) { (elx.requestFullscreen || elx.webkitRequestFullscreen || (() => {})).call(elx); }
  else if (document.exitFullscreen) document.exitFullscreen();
  setTimeout(() => { resize(); render(); }, 120);
}

function loop(ts) {
  if (!G.last) G.last = ts;
  const dt = Math.min(0.05, (ts - G.last) / 1000); G.last = ts;
  update(dt); render();
  if (G.state === "won" || G.state === "lost") {
    if (G.mode === "survival" && !G.recorded) { if (G.state === "lost") recordScore(G.wavesSurvived); G.recorded = true; }
    if (!G.awarded) { awardPoints(); G.awarded = true; }
    sfx(G.state === "won" ? "win" : "lose");
    return showModal();
  }
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
  const lvl = document.getElementById("r-level"); if (lvl) lvl.textContent = String(G.playerLevel || 1);
  const st = document.getElementById("r-stage"); if (st) st.textContent = G.mode === "survival" ? String(G.stage || 1) : "—";
  const wv = document.getElementById("r-wave"); if (wv) wv.textContent = G.mode === "survival" ? String(G.wave) : "—";
  updateCommands();
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
  const survival = G.mode === "survival";
  card.appendChild(el("h2", null, survival ? "THE CITADEL FALLS" : (G.state === "won" ? "RELEASE SHIPPED" : "PROJECT LOST")));
  card.appendChild(el("p", null, survival ? ("You outlasted " + G.wavesSurvived + " wave" + (G.wavesSurvived === 1 ? "" : "s") + ".") : (G.result || "")));
  const stats = el("ul", "stats");
  if (survival) {
    stats.appendChild(el("li", null, "Player level: " + (G.playerLevel || 1)));
    stats.appendChild(el("li", null, "Stage reached: " + (G.stage || 1)));
    stats.appendChild(el("li", null, "Waves survived: " + G.wavesSurvived));
    stats.appendChild(el("li", null, "Reached wave: " + G.wave));
    stats.appendChild(el("li", null, "Final Trust: " + Math.round(G.res.trust) + "%"));
  } else {
    stats.appendChild(el("li", null, "Features delivered: " + Math.floor(G.res.features)));
    stats.appendChild(el("li", null, "Prompts hardened: " + Math.floor(G.res.prompts)));
    stats.appendChild(el("li", null, "Hardening laid: " + Math.floor(G.res.hardening)));
    stats.appendChild(el("li", null, "Final Trust: " + Math.round(G.res.trust) + "%  ·  Citadel " + Math.max(0, Math.round(G.castle.hp)) + "%"));
  }
  card.appendChild(stats);
  card.appendChild(el("p", "newtop", "★ Earned " + (G.earnedPoints || 0) + " points" + (G.broughtArchive ? " (incl. +25% for your own agents)" : "") + "."));
  if (survival) {
    if (G.lastRecord && G.lastRecord.beatsOfficial) card.appendChild(el("p", "newtop", "🏆 New official top-3 — " + G.lastRecord.name + ", " + G.lastRecord.waves + " waves! It will be committed to the git leaderboard."));
    card.appendChild(el("div", "panel-title", "★ Leaderboard")); const lb = el("div", "lb"); renderLeaderboard(lb); card.appendChild(lb);
  }
  card.appendChild(el("div", "panel-title", "⚒ Armory — spend points to upgrade")); const arm = el("div", "armory"); arm.id = "upgrades-modal"; renderArmory(arm); card.appendChild(arm);
  const btn = el("button", "btn", "⟲ Play again"); btn.addEventListener("click", () => location.reload());
  card.appendChild(btn); m.appendChild(card);
}

/* ---- input ----------------------------------------------- */
function onClick(e) {
  const rect = canvas.getBoundingClientRect();
  const sx = (e.clientX - rect.left) / rect.width * canvas.width, sy = (e.clientY - rect.top) / rect.height * canvas.height;
  let hit = null, hd = 18;
  for (const a of G.units) { const d = Math.hypot(a.px - sx, a.py - sy); if (d < hd) { hd = d; hit = a; } }
  G.selected = hit;
  if (hit) sfx("select:" + hit.type);
  const ins = $("#inspect"); ins.replaceChildren(el("div", "panel-title", "⚑ Inspect"));
  if (hit) {
    const b = el("p"); const n = el("b", null, hit.label); n.style.color = hit.color; b.appendChild(n); ins.appendChild(b);
    ins.appendChild(el("p", "muted", "SDLC role")); ins.appendChild(el("p", null, hit.sdlc));
    if (hit.strong) { const s = el("p", "str"); s.appendChild(el("b", null, "▲ Strength: ")); s.appendChild(document.createTextNode(hit.strong)); ins.appendChild(s); }
    if (hit.weak) { const w = el("p", "wk"); w.appendChild(el("b", null, "▼ Weakness: ")); w.appendChild(document.createTextNode(hit.weak)); ins.appendChild(w); }
    if (hit.drift > 0) ins.appendChild(el("p", "wk", "≈ Model Drift active — missing! Cast Clear Context."));
    ins.appendChild(el("p", "hint", "Job: " + hit.job + (NODE_RES[hit.job] ? " → " + NODE_RES[hit.job] : "")));
  } else ins.appendChild(el("p", "muted", "Click an agent to inspect."));
}

/* ---- meta-progression: points + persistent upgrades ------ */
const PT_KEY = "bqa-citadel-points", UP_KEY = "bqa-citadel-upgrades";
const UPGRADES = [
  { key: "level", label: "Player level", glyph: "⚑", desc: "+1 unlocked Survival stage / level" },
  { key: "prompt", label: "Prompt hardening", glyph: "✦", desc: "+8% starting Trust / level" },
  { key: "mcp", label: "MCP relays", glyph: "⛁", desc: "Start with +1 MCP / level" },
  { key: "agents", label: "Standing agents", glyph: "♛", desc: "+1 starting defender / level" },
  { key: "workflow", label: "Workflow tuning", glyph: "⛏", desc: "+1 base Gather & Attack / 2 levels" },
];
function getPoints() { try { return parseInt(localStorage.getItem(PT_KEY) || "0", 10) || 0; } catch (_) { return 0; } }
function setPoints(n) { try { localStorage.setItem(PT_KEY, String(Math.max(0, Math.floor(n)))); } catch (_) {} }
function getUpgrades() { try { return Object.assign({ level: 0, prompt: 0, mcp: 0, agents: 0, workflow: 0 }, JSON.parse(localStorage.getItem(UP_KEY) || "{}")); } catch (_) { return { level: 0, prompt: 0, mcp: 0, agents: 0, workflow: 0 }; } }
function setUpgrades(u) { try { localStorage.setItem(UP_KEY, JSON.stringify(u)); } catch (_) {} }
function upCost(lvl) { return 60 + lvl * 70; }   // proportional cost growth
function buyUpgrade(key) {
  const u = getUpgrades(), cost = upCost(u[key]);
  if (getPoints() < cost) return;
  setPoints(getPoints() - cost); u[key]++; setUpgrades(u);
  renderArmory(document.getElementById("upgrades-start"));
  renderArmory(document.getElementById("upgrades-modal"));
}
function applyUpgrades() {
  const u = getUpgrades();
  G.playerLevel = 1 + Math.max(0, u.level || 0);
  G.unlockedStage = G.playerLevel;
  G.res.trust = clamp(60 + u.prompt * 8, 0, 100);
  for (let i = 0; i < u.mcp; i++) addMCP();
  for (let i = 0; i < u.agents; i++) spawnUnit("incident_defender");
  G.skills.gather += Math.floor(u.workflow / 2); G.skills.attack += Math.floor(u.workflow / 2);
}
function awardPoints() {
  let pts = G.wavesSurvived * 6 + Math.floor(G.res.features + G.res.prompts + G.res.hardening) + (G.state === "won" ? 60 : 0);
  if (G.broughtArchive) pts = Math.round(pts * 1.25);   // reward bringing your own agents
  G.earnedPoints = pts; setPoints(getPoints() + pts);
}
function renderArmory(box) {
  if (!box) return; box.replaceChildren();
  box.appendChild(el("p", "arm-pts", "★ Points: " + getPoints()));
  const u = getUpgrades();
  for (const up of UPGRADES) {
    const cost = upCost(u[up.key]);
    const row = el("button", "arm-row"); row.disabled = getPoints() < cost;
    row.appendChild(el("span", "arm-g", up.glyph));
    const body = el("div", "arm-body"); body.appendChild(el("b", null, up.label + " · L" + u[up.key])); body.appendChild(el("small", null, up.desc));
    row.appendChild(body); row.appendChild(el("span", "arm-cost", cost + "pt"));
    row.addEventListener("click", () => buyUpgrade(up.key));
    box.appendChild(row);
  }
}

/* ---- boot ------------------------------------------------ */
function resize() { const w = canvas.clientWidth; canvas.width = w; canvas.height = Math.round(w * 0.6); ORIGIN_X = canvas.width / 2; ORIGIN_Y = 56; }
function seedAgents(seedArchive) {
  spawnUnit("builder");
  G.broughtArchive = false;
  if (seedArchive && Array.isArray(seedArchive.sessions) && seedArchive.sessions.length) {
    G.broughtArchive = true;
    const map = { etl: "feature_worker", api: "hardening_engineer", graphql: "prompt_smith", data_quality: "hardening_engineer", bugs: "incident_defender", prompts: "prompt_smith" };
    const seen = new Set();
    for (const s of seedArchive.sessions) { const u = map[String(s.domain || "").toLowerCase()]; if (u && !seen.has(u)) { seen.add(u); spawnUnit(u); } }
    logMsg("Your uploaded agents answer the call (+25% points).");
  }
}
function begin() {
  applyUpgrades();
  if (G.mode === "survival") applySurvivalStartStage();
  G.state = "playing"; G.clock = 0; G.demoIdx = 0; G.last = 0; G.awarded = false;
  $("#start").style.display = "none";
  try { if (window.AUDIO) window.AUDIO.music(true); } catch (_) {}
  syncPanels(); requestAnimationFrame(loop);
}
function startDemo(seedArchive) {
  G.mode = "demo"; placeNodes(); seedAgents(seedArchive);
  logMsg("Builder Agent raised the first wall of architecture.");
  begin();
}
function startSurvival(seedArchive) {
  G.mode = "survival"; placeNodes();
  seedAgents(seedArchive);
  spawnUnit("incident_defender"); spawnUnit("incident_defender");
  G.wave = 0; G.waveGap = 1.5; G.wavesSurvived = 0; G.recorded = false;
  logMsg("Survival: hold the citadel — how many waves can you outlast?");
  begin();
}

document.addEventListener("DOMContentLoaded", () => {
  canvas = $("#field"); ctx = canvas.getContext("2d"); ctx.imageSmoothingEnabled = false;
  resize(); window.addEventListener("resize", () => { resize(); if (G.state !== "playing") render(); });
  canvas.addEventListener("click", onClick);
  placeNodes(); render(); syncPanels();
  let archive = null; try { const s = sessionStorage.getItem("bqa-archive"); if (s) archive = JSON.parse(s); } catch (_) {}
  $("#start-demo").addEventListener("click", () => startDemo(archive));
  const sv = document.getElementById("start-survival"); if (sv) sv.addEventListener("click", () => startSurvival(archive));
  const mute = document.getElementById("mute");
  if (mute) mute.addEventListener("click", () => { const on = window.AUDIO && window.AUDIO.toggle(); mute.textContent = on ? "🔊" : "🔇"; });
  renderLeaderboard(document.getElementById("lb-start"));
  loadOfficial();
  buildCommands();
  renderArmory(document.getElementById("upgrades-start"));
  const fs = document.getElementById("fs"); if (fs) fs.addEventListener("click", toggleFullscreen);
  if (archive) { const note = document.getElementById("archive-note"); if (note) note.style.display = ""; }
});
