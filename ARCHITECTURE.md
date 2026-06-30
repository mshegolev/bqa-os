# ARCHITECTURE — Citadel: The Release War

## Tech stack

- **Frontend / game**: vanilla **HTML + CSS + JavaScript** (no framework, no build
  step). The game is a static site under `docs/`.
- **Hosting**: **GitLab Pages**, published from `docs/` by the `pages` job in
  `.gitlab-ci.yml` on every push to `main`.
- **Engine repo**: the surrounding **BQA-OS** project is a Go (Cobra) CLI; it is
  not required to run the game, but it provides the agent tooling used to build it.

## Game structure

```text
docs/
  game.html            # markup, HUD, panels, canvas/field, styling hooks
  assets/
    game.js            # game loop, state, spawning, economy, leaderboard (~750 lines)
    game-logs.js       # unit/threat type definitions, threat categories, procedural-style pools
    lore.js            # in-browser, key-less LLM (Transformers.js) — generates hero lore on demand
    preview.gif        # gameplay preview shown in the README
  leaderboard.json     # git-tracked official top-3
  index.html           # project landing page
```

### Runtime design

- A single in-memory **game state** object (`G`) holds mode, clock, units,
  enemies, resources, skills, wave counters, Trust, and leaderboard data.
- A **game loop** advances by delta time and spawns escalating *survival* waves;
  clearing a wave spawns the next, with tougher orc types at higher thresholds.
- **Spawning** maps log event types and wave thresholds to unit/enemy types.
- **Economy** governs recruit costs and meta-upgrades; the **command bar** and
  **armory** are the two spend surfaces.
- The **leaderboard** merges the git-tracked official top-3 with a local
  personal best (`localStorage`), ranked by waves survived.

## Design decisions

- **In-browser LLM for generation**: the optional hero-lore generator runs a
  small model (Transformers.js, WebGPU/WASM) fully client-side — **no API key,
  no backend**. Weights load from the public Hugging Face CDN and cache locally.
- **Static and synthetic**: no backend, no auth, no server-side AI calls — so it
  deploys as plain files and is safe to host publicly. All data is synthetic.
- **Theme as teaching tool**: SDLC roles and security threats are dramatized as
  humans vs. green orcs to make the delivery/security workflow legible.
- **Single-file game logic** keeps the demo approachable and easy to host.
- **Pages from `docs/`** avoids any build pipeline; the CI job only copies files.

## AI / agent workflow (how it was built)

To make projects like this buildable, a dedicated **`bqa-team`** project was
created first: a reusable role-orchestrator pack (GitHub Issues + Codex CLI + QA
review + business acceptance), shipped as `scripts/bqa_team_orchestrator.py` and
the `.bqa-team/` roles/templates. That pack is what drives the pipeline below.

The game was produced by this automated agent pipeline rather than hand-written:

```text
business backlog
  → architect review
  → GitHub issue (bqa:* labels)
  → developer role via Codex CLI  → branch + code + tests + PR
  → QA role  → bug issue if QA fails
  → business acceptance
```

- **Claude Code** (Opus) orchestrated: planning, filing issues, reviewing,
  unblocking false-positive blocks, and importing/deploying to GitLab Pages.
- The **BQA Team autopilot** (`scripts/bqa_team_orchestrator.py`) ran the loop
  and invoked the **Codex CLI** (`codex exec`) as the developer/QA roles.
- Game features were tracked as issues — e.g. level-up stage progression,
  the leaderboard fanfare / fly-to-#1 animation, progressive unit unlocks in the
  tutorial, and a UI overlap bug fix — each implemented on its own branch + PR.
- **NotebookLM** and **Perplexity** were used separately for research.
