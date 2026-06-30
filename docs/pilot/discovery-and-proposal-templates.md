# Pilot readiness: discovery and proposal templates

Public-safe, copy-pasteable templates for running the first BQA-OS pilot
conversations and writing a pilot proposal. Three parts:

1. [Discovery call template](#1-discovery-call-template)
2. [Pilot proposal template](#2-pilot-proposal-template)
3. [Pilot readiness checklist](#3-pilot-readiness-checklist)

All examples here are **synthetic**. Keep real prospect data out of this public
repo; fill these in inside a private doc per prospect.

**Scope boundaries (state these up front):**

- This is a **2-week QA Memory Pilot**, not a full SaaS rollout.
- BQA-OS turns sanitized QA artifacts into reusable QA knowledge and grounded
  runtime context — it does **not** promise fully autonomous QA.
- The pilot scope is bounded (one focus area, a fixed artifact count, a fixed
  two-week window). It is **not** an unlimited free pilot.
- BQA-OS runs **local-first**: the prospect keeps their data; nothing is
  required to leave their machine.

---

## 1. Discovery call template

Use this to run a 30–45 minute discovery call. Capture answers in a private doc.

### Buyer persona and project context

- Who owns QA quality on the target project? (role, team size)
- What does the system under test do? (1–2 sentences)
- Tech stack and primary QA surface: **ETL / data quality / API / GraphQL** (circle the focus).
- Where does QA pain show up today? (releases slip, regressions escape, onboarding is slow, knowledge lives in people's heads)

### Current QA workflow

- How are checks run today? (manual, scripted, CI, ad-hoc)
- Where is QA knowledge stored? (wiki, tickets, chat, nowhere)
- How long does it take a new engineer to become productive in QA?
- What breaks most often after a change?

### Repeated regression checks

- Which checks are run on almost every release?
- Which past bugs keep coming back?
- Are these checks written down, or remembered?

### Focus area (pick one for the pilot)

- **ETL:** partitioning, idempotency/retries, duplicates, schema drift, row-count reconciliation, data-quality rules.
- **Data quality:** invariants, null/dup checks, referential integrity, freshness.
- **API:** contract/status codes, auth, pagination, error handling.
- **GraphQL:** query/mutation correctness, validation/auth, pagination, resolver/schema checks.

### Available sanitized artifacts

- Can you provide **10–30 sanitized** QA artifacts? (test notes, regression
  checklists, post-mortems, sanitized log snippets, prompts that worked)
- What must be redacted before sharing? (secrets, client names, internal URLs, tokens, PII)
- Who on your side will sanitize them?

### Decision and next step

- What would make this pilot a clear success for you?
- Who else needs to approve a paid pilot?
- Agreed next step + date: ______________________

**Required inputs from this call:** focus area, an artifact owner, a success
definition, and a decision-maker.
**Expected output:** enough to draft the pilot proposal below.

---

## 2. Pilot proposal template

> **BQA-OS QA Memory Pilot — Proposal**
> Prepared for: _[team / project, synthetic placeholder]_
> Date: _[date]_ · Owner (us): _[name]_ · Owner (you): _[name]_

### Context

_[1–2 sentences restating the prospect's QA pain in their words. Synthetic
example: "Regression checks for the orders ETL live in people's heads, so the
same null-account-id and duplicate-line bugs keep escaping to the warehouse."]_

### Focus area

_[One of: ETL / data quality / API / GraphQL.] Synthetic example: ETL data
quality for the orders + transactions pipelines._

### What we will do (2 weeks)

1. **Intake (Days 1–2):** you provide 10–30 sanitized QA artifacts; we confirm
   the focus area and what must not be shared.
2. **Build (Days 3–7):** import artifacts with `bqa ingest`, generate reusable
   QA knowledge with `bqa build`, and produce a grounded runtime context with
   `bqa codex` — all local-first.
3. **Review (Days 8–10):** a 30-minute review session walking the generated
   knowledge and demonstrating a grounded QA task in your runtime.
4. **Iterate (Days 11–14):** apply review feedback, re-run the build, and hand
   over the workspace + a short "how to keep using it" note.

### Deliverables

- A local `.bqa/` workspace built from **your** sanitized artifacts.
- Reviewed `.bqa/knowledge/*.yaml` QA knowledge for the focus area.
- A `.bqa/prompts/bqa-master-context.md` master context grounded in your QA.
- A live demo of a grounded QA task (e.g. in Codex / Claude Code).
- A short adoption note for your team.

### Timeline

Two weeks, with three touchpoints: kickoff (Day 1), review session (Day 8–10),
handover (Day 14).

### Success criteria (agreed up front)

_[2–4 measurable items. Synthetic examples:]_

- The generated knowledge correctly captures _[N]_ of your top recurring ETL
  checks (e.g. dedup on `(transaction_id, line_no)`, null-account row-count
  reconciliation).
- In the review session, a grounded QA task produces a plan your QA owner rates
  as "usable without rework."
- Your QA owner can re-run `bqa build` / `bqa codex` unaided after handover.

### Scope boundaries

- One focus area; 10–30 artifacts; two-week window.
- **No** promise of fully autonomous QA — BQA-OS grounds and accelerates humans
  and AI runtimes, it does not replace QA judgement.
- **No** unlimited free pilot; out-of-scope work is a separate engagement.
- Generated knowledge is heuristic and **human-reviewed** before it is trusted.

### Privacy / local-first

- Runs on your hardware; your data stays with you.
- Only **sanitized** artifacts are used; secrets, client names, tokens, and
  internal URLs are redacted before intake.
- We never copy raw sessions or secrets into shared output.

### Price and next step

- Pilot fee: _[amount]_ · Start date: _[date]_
- To proceed: _[single concrete next action]_

---

## 3. Pilot readiness checklist

Run this before committing to a pilot. The prospect is **ready** when every box
is checked.

### Buyer & context

- [ ] A named QA owner with authority over the target project.
- [ ] A decision-maker who can approve a paid pilot.
- [ ] A single agreed focus area (ETL / data quality / API / GraphQL).
- [ ] A clear, restated QA pain in the prospect's own words.

### Inputs

- [ ] 10–30 **sanitized** QA artifacts can be provided within Days 1–2.
- [ ] A named person owns sanitization on the prospect side.
- [ ] Agreement on what must be redacted (secrets, client names, URLs, tokens, PII).
- [ ] Artifacts are representative of real recurring checks (not toy data).

### Environment

- [ ] Local environment available to run the `bqa` CLI (Go + terminal) — see the
      [MacBook dogfood guide](macbook-dogfood-guide.md).
- [ ] A runtime to demo against is available (Codex / Claude Code / OpenCode).

### Scope & success

- [ ] Success criteria are written, measurable, and mutually agreed.
- [ ] Scope boundaries acknowledged (no autonomous-QA promise; bounded pilot).
- [ ] Two-week timeline and three touchpoints accepted.
- [ ] A concrete next step with an owner and a date.

### Privacy

- [ ] Local-first handling understood and accepted.
- [ ] Confirmation that **no** real client data lands in any shared/public repo.

When all boxes are checked, move to the
[onboarding checklist](onboarding-checklist.md) for kickoff and the
[first-week founder plan](first-week-founder-plan.md) for sequencing.
