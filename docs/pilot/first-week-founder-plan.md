# First-week founder operating plan

A concise, founder-operable 7-day plan for reaching the first **paid QA Memory
Pilot** conversations. One focus: the 2-week QA Memory Pilot — not a full SaaS
product, and not generic "AI for QA" positioning.

Keep all prospect data **outside this public repo** (a private CRM/notes doc).
Only synthetic examples belong here.

Pairs with:

- [Discovery and proposal templates](discovery-and-proposal-templates.md)
- [Pilot onboarding checklist](onboarding-checklist.md)
- [MacBook dogfood guide](macbook-dogfood-guide.md)

---

## The week at a glance

| Day | Focus | Measurable outcome |
| --- | --- | --- |
| 1 | Pilot offer + demo story | One-paragraph offer + a 3-minute demo narrative, written down. |
| 2 | Landing page | A live one-page pilot landing page linking the sales/onboarding docs. |
| 3 | Synthetic demo data + outputs | A reproducible synthetic demo: `.bqa/knowledge/*` + master context, captured. |
| 4 | Target account research (off-repo) | A private list of 15–20 qualified target accounts. |
| 5 | First messages (off-repo) | 10–15 personalized outbound messages sent. |
| 6 | Discovery calls | 2–3 discovery calls run with the [discovery template](discovery-and-proposal-templates.md#1-discovery-call-template). |
| 7 | Pilot proposals | 1–2 proposals sent using the [proposal template](discovery-and-proposal-templates.md#2-pilot-proposal-template). |

---

## Day 1 — Pilot offer and demo story

- Write the offer in one paragraph: *who* it is for (a QA owner with recurring
  ETL/data-quality/API/GraphQL regressions), *what* the 2-week pilot delivers
  (their sanitized QA turned into reusable, grounded QA knowledge), and the
  bounded scope.
- Draft a 3-minute demo narrative: pain → import sanitized artifacts → `bqa build`
  → grounded QA task in Codex.
- **Outcome:** offer paragraph + demo narrative committed to your private notes.

## Day 2 — Landing page

- Stand up a single landing page describing the QA Memory Pilot. Link the
  **sales package** ([discovery + proposal templates](discovery-and-proposal-templates.md)),
  the **demo packs** (synthetic examples such as `examples/etl-weekend/` and
  `examples/graphql-functional-qa/`), and the
  [**pilot onboarding**](onboarding-checklist.md) flow.
- State the scope boundary plainly: bounded 2-week pilot, local-first, no
  autonomous-QA promise.
- **Outcome:** a live one-page landing page with working links.

## Day 3 — Synthetic demo data and output examples

- Build a reproducible demo from synthetic data by following the
  [MacBook dogfood guide](macbook-dogfood-guide.md): `bqa init` → `bqa ingest --from`
  → `bqa build` → `bqa codex`.
- Capture the real outputs (knowledge artifacts + the master context section) as
  your demo exhibits. `bqa demo archive` can package a synthetic archive for the
  landing page.
- **Outcome:** a repeatable demo with saved output examples — no real data.

## Day 4 — Target account research (outside the repo)

- Build a private list of 15–20 accounts with the QA pain you solve (teams with
  data pipelines / ETL / API / GraphQL and visible regression pain).
- For each: the likely QA owner, the focus area, and one specific pain hypothesis.
- **Outcome:** 15–20 qualified accounts in your private CRM. **Nothing about
  real prospects goes in this repo.**

## Day 5 — First messages (outside the repo)

- Send 10–15 personalized outbound messages. Lead with the specific recurring
  QA pain and the bounded pilot — not generic AI claims.
- Point interested replies to the landing page.
- **Outcome:** 10–15 messages sent; replies tracked privately.

## Day 6 — Discovery calls

- Run 2–3 discovery calls using the
  [discovery call template](discovery-and-proposal-templates.md#1-discovery-call-template).
- Qualify against the
  [pilot readiness checklist](discovery-and-proposal-templates.md#3-pilot-readiness-checklist):
  focus area, an artifact owner, a success definition, a decision-maker.
- **Outcome:** 2–3 calls run; readiness captured per prospect.

## Day 7 — Pilot proposals

- Send 1–2 proposals using the
  [pilot proposal template](discovery-and-proposal-templates.md#2-pilot-proposal-template),
  each with a focus area, agreed success criteria, scope boundaries, and a price.
- **Outcome:** 1–2 proposals out, each with a concrete next step and date.

---

## Weekly outcome targets (measurable)

- 1 offer paragraph + 1 demo narrative.
- 1 live landing page.
- 1 reproducible synthetic demo with saved outputs.
- 15–20 qualified target accounts (private).
- 10–15 personalized messages sent.
- 2–3 discovery calls run.
- 1–2 pilot proposals sent.

---

## Anti-scope-creep warnings

- **Sell the pilot, not a platform.** Do not promise dashboards, integrations,
  or autonomous QA. The deliverable is the 2-week QA Memory Pilot.
- **Don't build product this week.** No new features, no UI. Use the CLI and
  synthetic demos exactly as they exist today.
- **Don't widen the focus.** One focus area per prospect (ETL / data quality /
  API / GraphQL). "Could it also…" goes to a later engagement.
- **Don't run an unbounded free pilot.** Two weeks, fixed artifact count, fixed
  success criteria, and a price.
- **Keep prospect data out of the public repo.** All real names, notes, and
  messages live in private tooling only.
- **Avoid generic AI positioning.** Lead with the specific, recurring QA pain
  you remove — never "AI-powered QA."
