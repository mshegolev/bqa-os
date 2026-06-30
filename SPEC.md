# SPEC — Citadel: The Release War

## Overview

Citadel: The Release War is a small, retro-fantasy RTS-style **demo game** that
visualizes the software delivery (SDLC) and security workflow. **Human agents**
(SDLC roles) build, gather, and defend a citadel while **green-orc threats** —
bugs, regressions, CVEs, incidents, tech debt, and the Deadline Warlord — march
on it. It ships as a static web page served from `docs/`.

## Scope

In scope:

- A single-file-driven browser game (HTML + CSS + JS), no backend, no login.
- One play mode: **Survival** — endless, escalating waves.
- An armory (meta-upgrades), a command bar (recruit/spend resources), a HUD, a
  minimap, and a leaderboard.
- Static, synthetic demo data only — no real customer data, secrets, or backend.

Out of scope:

- Networked multiplayer, persistence beyond `localStorage`, server APIs.
- Any runtime LLM/AI service calls (the game is fully static).

## Functional requirements

1. **Mode** — *Survival*: endless, escalating waves; clearing one wave spawns the
   next, with tougher enemy types unlocking at higher wave thresholds.
2. **Start menu (guided flow)**: ① upload agents (optional) → ② review the
   procedurally-styled warband → ③ to battle. Uploading is reached from the
   start screen; no top navigation buttons.
3. **Procedural warband**: an uploaded archive musters up to 6 heroes, each with
   a name, colour, glyph, skills, and MCP generated (stably) from the session.
4. **Inspect on click**: clicking a human shows its SDLC role, skills, and MCP;
   clicking an orc shows its threat type, damage, HP, trust impact, and tags.
5. **Resources & economy**
   - Players spend resources to recruit units (Feature Worker, Prompt Smith,
     Hardening Engineer, Defender, ranged Archer), deploy an MCP relay, or boost
     Attack / Gather. A "Clear Context" action is available.
6. **Armory meta-upgrades** (spend points): Prompt hardening (+starting Trust),
   MCP relays (+starting MCP), Standing agents (+starting defender), Workflow
   tuning (+base Gather & Attack).
7. **Trust & citadel**: a Trust score and citadel integrity are tracked in the
   HUD; the citadel can take damage from threats.
8. **Leaderboard**: official top-3 is read from `docs/leaderboard.json`; the
   player's personal best (by waves survived) is stored locally.
9. **Upload incentive**: uploading your own agents/prompts/logs musters a bigger
   warband (+25% points).

## Acceptance criteria

- The page loads at `/game.html` with **no console errors or warnings**.
- **Survival** starts and runs without runtime errors; the start menu (upload →
  review warband → battle) works and clicking units/orcs shows their details.
- A win is achievable by shipping with **Trust ≥ 70%** and the citadel standing.
- Survival waves escalate and the run ends when the citadel falls; the result is
  recorded against the leaderboard (waves survived).
- All assets (`game.html`, `assets/game.js`, `leaderboard.json`) return HTTP 200
  when served statically.
- The site is publishable via GitLab Pages from `docs/` with no build step.
