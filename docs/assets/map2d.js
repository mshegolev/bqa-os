/* ============================================================
   BQA-OS — Mission 1: a Warcraft-1-style 2D tile map.
   Top-down pixel terrain rendered on a canvas: grass, forest,
   water, a dirt road, a Town Hall and a gold mine, with the QA
   team standing as units. Click a unit to read its order;
   "Advance" marches the team down the road (ships tasks).
   Pure canvas, no assets, synthetic data only.
   ============================================================ */

"use strict";

const TILE = 16;           // logical pixels per tile
const MAP = [
  "FFFFFFFFFFFFFFFFFFFF",
  "FFGGGGGGGGGGGFFWWWFF",
  "FFGGGGGGGGGGGFFWWWFF",
  "FFGGGGGGGGGGGGFWWWFF",
  "FGGGGGGGGGGGGGGGGGFF",
  "FGGGGGGGGGGGGGGGGGFF",
  "FGGGGGGGGGGGGGGGGGFF",
  "FGGGGGGGGGGGGGGGGGFF",
  "FFGGGGGGGDGGGGGGGGFF",
  "FFGGGGGGGDGGGGGGGGFF",
  "FFFGGGGGGDGGGGGGGFFF",
  "FFFFFGGGGDGGGGFFFFFF",
  "FFFFFFFFFDGFFFFFFFFF",
  "FFFFFFFFFDFFFFFFFFFF"
];
const COLS = MAP[0].length;
const ROWS = MAP.length;
const ROAD_COL = 9;

const BUILDINGS = [
  { type: "hall", tx: 6, ty: 5, w: 2, h: 2 },
  { type: "mine", tx: 14, ty: 4, w: 2, h: 2 }
];

// QA team as units. role drives colour + the order they carry.
const ROLES = {
  owner:     { label: "Business Owner", color: "#f2c14e", task: "Hold the Citadel & accept work" },
  architect: { label: "Architect",      color: "#6fb3e0", task: "Approve the design for #27" },
  developer: { label: "Developer",      color: "#7bc45a", task: "Implement ETL QA pack #37" },
  qa:        { label: "QA",             color: "#8e6fc4", task: "Verify the upload flow #31" },
  reviewer:  { label: "Reviewer",       color: "#c0392b", task: "Final acceptance of #25" }
};
const HALL_DOOR = { tx: 7, ty: 7 };
const UNITS = [
  { role: "owner",     tx: 6,  ty: 6, guard: true },
  { role: "architect", tx: 5,  ty: 5 },
  { role: "developer", tx: 11, ty: 6 },
  { role: "qa",        tx: 6,  ty: 8 },
  { role: "reviewer",  tx: 11, ty: 8 }
];

let turn = 1, shipped = 0, selected = null;
let ctx = null, canvas = null;

/* deterministic per-tile pseudo-random so terrain doesn't flicker */
function hash(x, y, s) {
  let h = (x * 374761393 + y * 668265263 + s * 2147483647) >>> 0;
  h = (h ^ (h >>> 13)) * 1274126177 >>> 0;
  return ((h ^ (h >>> 16)) >>> 0) / 4294967295;
}
function px(x, y, w, h, c) { ctx.fillStyle = c; ctx.fillRect(x | 0, y | 0, w | 0, h | 0); }

/* ---- terrain tiles ---------------------------------------- */
function drawGrass(x, y, tx, ty) {
  px(x, y, TILE, TILE, "#4a7a3a");
  for (let i = 0; i < 6; i++) {
    const r = hash(tx, ty, i + 1);
    const dx = (hash(tx, ty, i + 11) * 14) | 0, dy = (hash(tx, ty, i + 21) * 14) | 0;
    px(x + dx, y + dy, r > 0.5 ? 2 : 1, 1, r > 0.7 ? "#5c9248" : "#3c6630");
  }
}
function drawDirt(x, y, tx, ty) {
  px(x, y, TILE, TILE, "#a3814a");
  for (let i = 0; i < 7; i++) {
    const dx = (hash(tx, ty, i + 3) * 15) | 0, dy = (hash(tx, ty, i + 7) * 15) | 0;
    px(x + dx, y + dy, 1, 1, hash(tx, ty, i) > 0.5 ? "#8a6a3a" : "#6e5230");
  }
}
function drawWater(x, y, tx, ty) {
  px(x, y, TILE, TILE, "#2f6090");
  for (let row = 2; row < TILE; row += 5) {
    const off = ((ty + row) % 2) * 4;
    px(x + off, y + row, 6, 1, "#4a86c0");
    px(x + off + 8, y + row + 2, 4, 1, "#8fc4e8");
  }
}
function drawForest(x, y, tx, ty) {
  px(x, y, TILE, TILE, "#3c6630");
  // trunk
  px(x + 7, y + 10, 2, 4, "#5a3a1c");
  // canopy
  px(x + 4, y + 3, 8, 7, "#2f5a26");
  px(x + 3, y + 5, 10, 4, "#2f5a26");
  px(x + 5, y + 2, 6, 2, "#3c6e30");
  px(x + 6, y + 4, 3, 2, "#4a7a3a"); // highlight
  px(x + 9, y + 7, 2, 2, "#244a1e"); // shadow
}

function drawTile(tx, ty) {
  const x = tx * TILE, y = ty * TILE, t = MAP[ty][tx];
  if (t === "W") drawWater(x, y, tx, ty);
  else if (t === "F") drawForest(x, y, tx, ty);
  else if (t === "D") drawDirt(x, y, tx, ty);
  else drawGrass(x, y, tx, ty);
}

/* ---- buildings -------------------------------------------- */
function drawHall(b) {
  const x = b.tx * TILE, y = b.ty * TILE, w = b.w * TILE, h = b.h * TILE;
  // ground shadow
  px(x + 2, y + h - 3, w - 2, 4, "#244a1e");
  // stone base
  px(x + 2, y + 8, w - 4, h - 10, "#7c6a47");
  px(x + 3, y + 9, w - 6, h - 12, "#b9a37a");
  // stone blocks
  for (let by = y + 11; by < y + h - 4; by += 5)
    for (let bx = x + 4; bx < x + w - 5; bx += 7) px(bx, by, 6, 1, "#7c6a47");
  // roof
  px(x + 1, y + 6, w - 2, 4, "#6e281c");
  px(x + 3, y + 2, w - 6, 6, "#9c3a2a");
  px(x + 6, y, w - 12, 4, "#9c3a2a");
  px(x + 5, y + 3, w - 10, 1, "#c0503a");
  // door
  px(x + w / 2 - 2, y + h - 9, 4, 8, "#3a2a18");
  // banner
  px(x + w - 4, y - 3, 1, 8, "#5a3a1c");
  px(x + w - 8, y - 3, 4, 3, "#f2c14e");
}
function drawMine(b) {
  const x = b.tx * TILE, y = b.ty * TILE, w = b.w * TILE, h = b.h * TILE;
  px(x + 2, y + 6, w - 4, h - 7, "#4a4a4a");
  px(x + 3, y + 7, w - 6, h - 9, "#6b6b6b");
  // cave mouth
  px(x + w / 2 - 4, y + h - 10, 8, 9, "#14110d");
  px(x + w / 2 - 5, y + h - 11, 10, 2, "#3a2a18");
  // gold flecks
  for (let i = 0; i < 8; i++) {
    const dx = (hash(b.tx, b.ty, i) * (w - 6) + 3) | 0, dy = (hash(b.tx, b.ty, i + 5) * (h - 10) + 7) | 0;
    px(x + dx, y + dy, 1, 1, "#f2c14e");
  }
}
function drawBuildings() {
  for (const b of BUILDINGS) (b.type === "hall" ? drawHall : drawMine)(b);
}

/* ---- units ------------------------------------------------ */
function drawUnit(u) {
  const x = u.tx * TILE, y = u.ty * TILE, c = ROLES[u.role].color;
  const cx = x + 8, top = y + 3;
  // shadow
  px(x + 4, y + 14, 8, 2, "rgba(0,0,0,0.35)");
  // legs
  px(cx - 2, top + 8, 2, 3, "#3a2a18");
  px(cx, top + 8, 2, 3, "#3a2a18");
  // body (role colour)
  px(cx - 3, top + 3, 6, 6, c);
  px(cx - 3, top + 3, 6, 1, "#ffffff33");
  // head
  px(cx - 2, top, 4, 4, "#e0a878");
  // helmet
  px(cx - 2, top - 1, 4, 2, "#c9d2dc");
  // selection brackets
  if (selected === u) {
    ctx.fillStyle = "#7bc45a";
    const b = 3;
    px(x, y, b, 1); px(x, y, 1, b);
    px(x + TILE - b, y, b, 1); px(x + TILE - 1, y, 1, b);
    px(x, y + TILE - 1, b, 1); px(x, y + TILE - b, 1, b);
    px(x + TILE - b, y + TILE - 1, b, 1); px(x + TILE - 1, y + TILE - b, 1, b);
  }
}

/* ---- fog / vignette --------------------------------------- */
function drawFog() {
  const g = ctx.createRadialGradient(
    ROAD_COL * TILE, 6 * TILE, TILE * 3,
    ROAD_COL * TILE, 6 * TILE, TILE * 13);
  g.addColorStop(0, "rgba(0,0,0,0)");
  g.addColorStop(1, "rgba(8,8,12,0.5)");
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, canvas.width, canvas.height);
}

function render() {
  for (let ty = 0; ty < ROWS; ty++) for (let tx = 0; tx < COLS; tx++) drawTile(tx, ty);
  drawBuildings();
  for (const u of UNITS) drawUnit(u);
  drawFog();
  document.getElementById("turn").textContent = turn;
  document.getElementById("shipped").textContent = shipped;
  renderRoster();
}

/* ---- side panel ------------------------------------------- */
const el = (tag, cls, text) => { const e = document.createElement(tag); if (cls) e.className = cls; if (text != null) e.textContent = String(text); return e; };

function renderRoster() {
  const box = document.getElementById("roster");
  box.replaceChildren();
  for (const u of UNITS) {
    const r = ROLES[u.role];
    const row = el("button", "rost-row" + (selected === u ? " sel" : ""));
    const sw = el("span", "rost-sw");
    sw.style.background = r.color;
    const body = el("div", "rost-body");
    body.appendChild(el("b", null, r.label));
    body.appendChild(el("small", null, r.task));
    row.appendChild(sw); row.appendChild(body);
    row.addEventListener("click", () => { selected = u; render(); showOrder(u); });
    box.appendChild(row);
  }
}
function showOrder(u) {
  const r = ROLES[u.role];
  const box = document.getElementById("order");
  box.replaceChildren();
  box.appendChild(el("div", "panel-title", "⚑ Orders"));
  const name = el("p"); const b = el("b", null, r.label); b.style.color = r.color;
  name.appendChild(b); box.appendChild(name);
  box.appendChild(el("p", null, r.task));
  box.appendChild(el("p", "muted", "Tile " + u.tx + "," + u.ty + (u.guard ? " · guarding the hall" : " · ready to march")));
}

/* ---- interaction ------------------------------------------ */
function tileFromEvent(e) {
  const rect = canvas.getBoundingClientRect();
  const lx = (e.clientX - rect.left) / rect.width * canvas.width;
  const ly = (e.clientY - rect.top) / rect.height * canvas.height;
  return { tx: Math.floor(lx / TILE), ty: Math.floor(ly / TILE) };
}
function onClick(e) {
  const { tx, ty } = tileFromEvent(e);
  const u = UNITS.find((u) => u.tx === tx && u.ty === ty);
  if (u) { selected = u; render(); showOrder(u); }
}

const sign = (n) => (n > 0 ? 1 : n < 0 ? -1 : 0);
function advance() {
  for (const u of UNITS) {
    if (u.guard) continue;
    if (u.tx !== ROAD_COL) u.tx += sign(ROAD_COL - u.tx);
    else if (u.ty < ROWS - 1) u.ty += 1;
    else { shipped++; u.tx = HALL_DOOR.tx; u.ty = HALL_DOOR.ty; } // delivered → back to hall
  }
  turn++;
  render();
}
function rally() {
  UNITS.forEach((u, i) => { u.tx = [6, 5, 11, 6, 11][i]; u.ty = [6, 5, 6, 8, 8][i]; });
  selected = null;
  document.getElementById("order").replaceChildren(el("div", "panel-title", "⚑ Orders"), el("p", "muted", "Select a unit on the map."));
  render();
}

document.addEventListener("DOMContentLoaded", () => {
  canvas = document.getElementById("map");
  canvas.width = COLS * TILE;
  canvas.height = ROWS * TILE;
  ctx = canvas.getContext("2d");
  ctx.imageSmoothingEnabled = false;
  canvas.addEventListener("click", onClick);
  document.getElementById("advance").addEventListener("click", advance);
  document.getElementById("rally").addEventListener("click", rally);
  render();
});
