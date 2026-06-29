/* ============================================================
   CITADEL: The Release War — demo data & taxonomy.
   Original "inspired by 90s fantasy RTS" setting (no Warcraft
   assets/names). This file is where you tune the demo: the unit
   and threat taxonomy, the 5-phase arc, the log->spawn mapping,
   and the scripted DEMO_LOGS timeline.
   ============================================================ */

"use strict";

/* The 5-phase SDLC story arc (ProgressTimeline). */
const PHASES = [
  { key: "Bootstrap",        blurb: "Raise the first walls." },
  { key: "Feature Delivery", blurb: "Mine product capability." },
  { key: "Prompt Hardening", blurb: "Reforge prompts, lay stone." },
  { key: "Incident Response",blurb: "Hold the perimeter." },
  { key: "Deadline Battle",  blurb: "Survive the Warlord, ship." },
];

/* Human faction — each maps to an SDLC role. job: which resource node
   they harvest (or 'defend'/'build'/'release'). */
const UNIT_TYPES = {
  builder:             { label: "Builder Agent",       sdlc: "Setup / CI/CD / architecture",     color: "#b9a37a", glyph: "⌂", job: "build" },
  feature_worker:      { label: "Feature Worker",      sdlc: "Feature delivery",                 color: "#f2c14e", glyph: "★", job: "gold" },
  prompt_smith:        { label: "Prompt Smith",        sdlc: "Prompt engineering / guardrails",  color: "#8e6fc4", glyph: "✦", job: "wood" },
  hardening_engineer:  { label: "Hardening Engineer",  sdlc: "Security hardening / validation",  color: "#6fb3e0", glyph: "⛨", job: "stone" },
  context_logger:      { label: "Context Logger",      sdlc: "Logs / telemetry / observability", color: "#7bc45a", glyph: "≋", job: "mana" },
  incident_defender:   { label: "Incident Defender",   sdlc: "Triage / mitigation / response",   color: "#e0883a", glyph: "⚔", job: "defend" },
  release_captain:     { label: "Release Captain",     sdlc: "Release readiness / launch",       color: "#f4eddc", glyph: "✚", job: "release" },
  sentinel_archer:     { label: "Sentinel Archer",     sdlc: "Static analysis / scanning (ranged)", color: "#9ed36a", glyph: "➹", job: "defend", ranged: true, range: 150, dmg: 24 },
};

/* Orc faction — threats. size scales the sprite; boss flags the Warlord. */
const ENEMY_TYPES = {
  bug_grunt:        { label: "Bug Grunt",         color: "#5a8a3a", size: 1.0, hp: 30,  dmg: 6,  speed: 1.4, trust: 4 },
  regression_raider:{ label: "Regression Raider", color: "#7bc45a", size: 0.9, hp: 24,  dmg: 5,  speed: 2.2, trust: 4 },
  spear_hurler:     { label: "Spear Hurler",      color: "#6f9a3a", size: 1.0, hp: 34,  dmg: 9,  speed: 1.1, trust: 6, ranged: true, range: 130 },
  cve_shaman:       { label: "CVE Shaman",        color: "#6e8e3a", size: 1.1, hp: 55,  dmg: 12, speed: 1.2, trust: 9, cve: true },
  incident_ogre:    { label: "Incident Ogre",     color: "#46662a", size: 1.5, hp: 110, dmg: 18, speed: 1.0, trust: 12 },
  tech_debt_troll:  { label: "Tech Debt Troll",   color: "#3f5a2a", size: 1.6, hp: 160, dmg: 10, speed: 0.6, trust: 10 },
  deadline_warlord: { label: "Deadline Warlord",  color: "#7a1f15", size: 2.0, hp: 210, dmg: 16, speed: 0.8, trust: 25, cve: true, boss: true },
};

/* log event type -> what it spawns ('unit:x' | 'enemy:x' | 'horde'). */
const LOG_SPAWN = {
  feature_detected:   "unit:feature_worker",
  prompt_loaded:      "unit:prompt_smith",
  hardening_rule_added:"unit:hardening_engineer",
  log_ingested:       "unit:context_logger",
  bug_found:          "enemy:bug_grunt",
  regression:         "enemy:regression_raider",
  cve_detected:       "enemy:cve_shaman",
  incident_started:   "enemy:incident_ogre",
  tech_debt:          "enemy:tech_debt_troll",
  mass_incident:      "horde",
  deadline_critical:  "enemy:deadline_warlord",
  release_ready:      "unit:release_captain",
};

/* example DemoLogEvent shape (documented for extension):
   { id, t (seconds offset), type, title, msg, phase, severity? }   */
const DEMO_LOGS = [
  // Phase 0 — Bootstrap
  { id: "l01", t: 0.5, phase: 0, type: "bootstrap",         title: "Bootstrap",       msg: "Builder Agent raised the first wall of architecture." },
  { id: "l02", t: 3.0, phase: 0, type: "bug_found",         title: "Bug found",       msg: "A Bug Grunt skitters out of the legacy thicket." },
  { id: "l03", t: 5.0, phase: 0, type: "log_ingested",      title: "Logs ingested",   msg: "Context Logger taps the telemetry stream." },
  // Phase 1 — Feature Delivery
  { id: "l04", t: 7.5,  phase: 1, type: "feature_detected", title: "Feature",         msg: "Feature Worker mined a new product capability." },
  { id: "l05", t: 9.5,  phase: 1, type: "feature_detected", title: "Feature",         msg: "Feature Worker hauls gold to the keep." },
  { id: "l06", t: 11.0, phase: 1, type: "regression",       title: "Regression",      msg: "Regression Raider charges after the last change." },
  { id: "l07", t: 13.0, phase: 1, type: "bug_found",        title: "Bug found",       msg: "Another Bug Grunt joins the fray." },
  // Phase 2 — Prompt Hardening
  { id: "l08", t: 15.5, phase: 2, type: "prompt_loaded",       title: "Prompt",        msg: "Prompt Smith reforged brittle instructions into hardened prompts." },
  { id: "l09", t: 17.5, phase: 2, type: "hardening_rule_added",title: "Hardening",     msg: "Hardening Engineer added validation stonework." },
  { id: "l10", t: 19.5, phase: 2, type: "cve_detected",        title: "CVE",           msg: "CVE Shaman casts an exploit near the northern wall.", severity: "high" },
  { id: "l11", t: 21.5, phase: 2, type: "hardening_rule_added",title: "Hardening",     msg: "Secure defaults mortared into place." },
  // Phase 3 — Incident Response
  { id: "l12", t: 24.0, phase: 3, type: "incident_started", title: "Incident",       msg: "Incident Ogre breached the observability perimeter.", severity: "high" },
  { id: "l13", t: 26.5, phase: 3, type: "tech_debt",        title: "Tech debt",      msg: "A Tech Debt Troll lumbers in, slow but stubborn." },
  { id: "l14", t: 28.5, phase: 3, type: "mass_incident",    title: "Mass incident",  msg: "A horde pours from the breach — all hands!", severity: "critical" },
  { id: "l15", t: 30.5, phase: 3, type: "cve_detected",     title: "CVE",            msg: "Second CVE Shaman flanks from the east.", severity: "high" },
  // Phase 4 — Deadline Battle
  { id: "l16", t: 33.0, phase: 4, type: "deadline_critical",title: "DEADLINE",       msg: "The Deadline Warlord has entered the battlefield.", severity: "critical" },
  { id: "l17", t: 36.0, phase: 4, type: "release_ready",    title: "Release",        msg: "Release Captain rallies the agents for final deployment." },
];
