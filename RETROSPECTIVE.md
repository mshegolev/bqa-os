# RETROSPECTIVE — Citadel: The Release War

## AI tools used

- **Claude Code** (Anthropic Opus) — orchestrator: planning, code review, filing
  and triaging issues, unblocking stuck work, the UI/UX redesign, and
  importing/deploying the project to GitLab Pages.
- **Codex CLI** (`codex exec`, GPT‑5.5) — developer/QA roles, driven automatically
  by the BQA Team autopilot (`scripts/bqa_team_orchestrator.py`).
- A dedicated **`bqa-team`** project was built first to make this possible — a
  reusable role-orchestrator pack (issues → dev → PR → QA → acceptance).
- **NotebookLM** and **Perplexity** — separate research.
- **In-browser LLM at runtime** — Transformers.js (`onnx-community/gemma-3-270m-it-ONNX`,
  WebGPU/WASM) generates hero lore client-side, key-less, no backend.

## What the game is

A static, synthetic, vanilla HTML/CSS/JS retro-fantasy RTS that dramatizes the
SDLC/security workflow: human agents defend a citadel against orc threats (bugs,
CVEs, incidents, the Deadline Warlord). One mode — **Survival** (endless escalating
waves). Deployed via GitLab Pages from `docs/`.

## Workflow

Backlog items became issues; the autopilot turned them into branches with code +
tests and opened PRs; a QA role reviewed them. Claude Code supervised the loop and
then drove a long, interactive UI/UX iteration directly on the static game:
single-mode simplification, a redesigned start menu, a single-viewport responsive
layout, click-to-inspect, procedural heroes, and the in-browser lore LLM. Each
change was verified in a real browser (DOM measurements + console checks) and
deployed on every push to `main`.

## What worked

- **Issue-driven autopilot** kept early work parallel and traceable.
- **Static + synthetic** meant trivial, safe hosting and a fast feedback loop.
- **Verify in the real artifact**: every UI change was checked by measuring the
  live DOM (overlaps, viewport fit, button counts) and the console — not by eye
  alone. This caught the start-overlay overlap and the canvas squash early.
- **One-command onboarding**: `make-archive.sh` turns local AI sessions into an
  uploadable `archive.zip` with a single `curl | bash`.
- **Key-less in-browser AI**: lore generation runs fully client-side, so the
  static page gains an LLM feature with no secrets and no server.

## What didn't (and the fixes)

- **False-positive blocks**: a no-op dev run produced an empty commit that the
  orchestrator mislabeled as a failure → recognize "nothing to commit" and retry.
- **Duplicate issues** had to be closed against the canonical one.
- **UI overlap**: the start overlay's content bled onto the HUD and command bar →
  clipped + scrollable overlay; later replaced by a clean menu + collapsible prep.
- **Canvas didn't scale** (tiny in fullscreen, squashed in-window) → a render
  scale-to-fit transform + `ResizeObserver` so the field fills any viewport.
- **GitLab vs GitHub**: the autopilot assumes GitHub/`go test`; the static GitLab
  deliverable needed a different path (export + GitLab Pages + a project runner).
- **CI runner**: Pages sat pending until a project runner was registered; one
  analyzer also needed "Run untagged jobs" enabled.
- **CDN cache**: edge nodes briefly served a stale `make-archive.sh` after deploy.
- **Pages HTTPS**: the instance serves a wildcard cert for another domain that does not
  cover the Pages subdomain, so the demo is reached over **http** (infra-level).

## Lessons learned

- Automated agent loops are productive but need **guardrails against false failures**
  and **duplicate work**.
- Keeping the deliverable **static + synthetic** removed whole classes of hosting,
  secret, and security problems — and still allowed a real AI feature via in-browser
  inference.
- **Measure the live UI**, don't trust green tests alone, when the work is visual.
- **Strong visual hierarchy** (one dominant CTA, collapsible secondary controls,
  single-viewport layout) made the game immediately understandable.
- Match the tool to the target: a GitHub-centric pipeline needs adaptation for a
  GitLab Pages static deliverable, down to runners and certificates.
