# ICP, Lead List Criteria & Sourcing (Issue #48)

Goal: build a **prioritized lead list** for the BQA-OS **2-week QA Memory Pilot**
— not generic AI outreach. Every lead should map to a concrete BQA-OS pain:
QA knowledge loss, repeated regression checks, slow QA onboarding, API/GraphQL
regression bottlenecks, ETL / data-quality validation, or "AI coding sped up
delivery and QA became the bottleneck."

> **Privacy:** This repo is public. Do **not** commit real lead data, scraped
> personal data, secrets, or private notes. Keep the actual lead list in a
> private sheet/CRM. The template below is synthetic and structural only.

---

## 1. ICP segments

Cover at least these four. A lead can match more than one segment; record the
**primary** segment in the lead table.

### Segment A — QA Lead / QA Automation Lead (B2B SaaS, 20–200 employees)
- Owns regression and test automation for products with **APIs, GraphQL, or data
  pipelines**.
- Pain: regression checks repeat every release; QA knowledge lives in a few
  heads; onboarding a new QA engineer into project-specific workflows is slow.
- Why BQA-OS fits: turns existing session logs / test notes / regression
  checklists into reusable agents, skills, and workflows.

### Segment B — CTO / VP Engineering (startup where AI coding accelerated delivery)
- Engineers ship faster with AI coding; **QA is now the bottleneck**.
- Pain: throughput mismatch between dev and QA, quality risk on releases,
  no system to capture QA know-how at the new pace.
- Why BQA-OS fits: an AI-native QA layer that scales reusable QA knowledge
  alongside AI-accelerated development.

### Segment C — QA consultants / boutique QA agencies
- Run QA for multiple clients; reusable QA assets are leverage and margin.
- Pain: re-deriving the same regression/onboarding knowledge per client;
  knowledge walks out when a contractor rolls off.
- Why BQA-OS fits: a reusable, sanitized QA-knowledge system they can apply
  across engagements (private value stays in their own BQA Brain / local `.bqa`).

### Segment D — Big Data / ETL QA teams & data-platform QA owners
- Validate ETL pipelines and data quality; heavy on data-quality rules.
- Pain: manual, repetitive ETL / data-quality validation; tribal knowledge of
  pipeline quirks; hard to onboard QA into data workflows.
- Why BQA-OS fits: directly supports ETL/Big-Data testing and data-quality
  validation as first-class domains.

---

## 2. Company filters

Use as inclusion/exclusion filters when sourcing. These are **starting
heuristics**, not hard rules.

**Include when:**
- B2B software or data product with APIs, GraphQL, and/or data pipelines.
- ~20–200 employees (Segments A/B); boutique agencies of any size (C);
  data/platform teams of any size (D).
- Signals of an active QA or release-quality function (QA roles posted, QA tooling
  in stack, public engineering blog on testing/quality).
- Reachable, relevant decision-influencer in a target persona.

**Exclude / deprioritize when:**
- Pure hardware, no software QA surface.
- Heavily regulated environments that forbid any external/AI tooling in QA.
- Org so large the first conversation requires enterprise procurement (out of
  scope for a 2-week pilot).
- No identifiable QA pain and no API/GraphQL/ETL surface.

---

## 3. Target personas & titles

| Persona | Example titles | Primary segment |
|---|---|---|
| QA owner | QA Lead, QA Automation Lead, SDET Lead, Head of QA, QA Manager | A |
| Eng leader | CTO, VP Engineering, Head of Engineering, Eng Director | B |
| Agency owner | Founder / Owner / Principal of a QA consultancy, QA Practice Lead | C |
| Data QA owner | Big Data QA Lead, ETL QA Engineer, Data Platform QA, Analytics Eng (QA-leaning) | D |

---

## 4. Trigger events

A trigger makes outreach timely and relevant. Prefer leads with at least one.

- Hiring QA / SDET / data-QA roles (regression or automation in the JD).
- Public post / talk about QA bottlenecks, flaky tests, or slow releases.
- Announced or visible adoption of AI coding tools (Copilot/Codex/Claude/Cursor).
- New API or GraphQL launch, or migration to GraphQL.
- New / expanding data platform, ETL, or data-quality initiative.
- Recent funding or fast headcount growth in engineering (throughput pressure).
- Agency publicly expanding QA services or taking on new clients (Segment C).

---

## 5. Personalization angles

Pick the angle that matches the lead's trigger + segment. Keep it specific and
honest.

- **Knowledge loss:** "Looks like your QA know-how lives in a few people —
  curious how you reuse it across releases."
- **Regression repetition:** "Saw you're shipping fast on \<product\> — how much
  of regression is re-run-the-same-checks each release?"
- **QA onboarding:** "You're hiring QA — onboarding into project-specific
  workflows is usually the slow part. How do you handle it today?"
- **AI-coding bottleneck (B):** "AI coding sped your team up — is QA keeping pace
  or becoming the constraint?"
- **API/GraphQL:** "Noticed your GraphQL API — how do you keep functional /
  regression coverage from drifting?"
- **ETL / data quality (D):** "ETL data-quality validation tends to be manual and
  repetitive — what does your current check look like?"
- **Agency leverage (C):** "You run QA across clients — reusable QA assets seem
  like real leverage; how do you carry knowledge between engagements?"

---

## 6. Likely QA pain hypothesis (by segment)

State the **hypothesis** to test in discovery — do not assert it as fact.

| Segment | Likely primary pain to validate |
|---|---|
| A | Regression repeats; QA knowledge siloed; slow onboarding into workflows. |
| B | QA throughput can't match AI-accelerated dev; release-quality risk. |
| C | Re-deriving QA knowledge per client; knowledge loss on rolloff. |
| D | Manual/repetitive ETL & data-quality validation; tribal pipeline knowledge. |

---

## 7. Priority scoring rubric

Two scores per lead: a quick **pilot fit (1–5)** and a composite **priority
(1–100)** that decides contact order.

### 7a. Pilot fit (1–5) — quick gut-check
| Score | Meaning |
|---|---|
| 5 | Clear BQA-OS pain + API/GraphQL/ETL surface + reachable target persona + a trigger. |
| 4 | Clear pain and right persona; trigger weak or inferred. |
| 3 | Plausible fit; persona right but pain unconfirmed. |
| 2 | Weak fit; persona or surface unclear. |
| 1 | Poor fit; likely disqualify. |

### 7b. Priority score (1–100) — contact order
Weighted sum of five factors (each scored 1–5, then weighted):

| Factor | Weight | What "5" looks like |
|---|---|---|
| Pain clarity | ×6 | Public, specific QA pain matching a BQA-OS angle. |
| Trigger freshness | ×5 | Strong, recent trigger event. |
| Persona reachability | ×4 | Direct, named target persona contact identified. |
| Segment value | ×3 | Falls in a high-conviction segment for this batch. |
| Pilot scopeability | ×2 | Obvious 2-week, 10–30-artifact pilot scope; no enterprise procurement. |

`priority = (pain×6 + trigger×5 + reach×4 + segment×3 + scope×2)` → range 20–100.

**Contact order:** sort descending by priority; within ties, higher pilot fit
first. Work the top of the list each day (see `7-day-pilot-cadence.md`).

---

## 8. Lead table fields (CRM-ready)

The template supports **100–200 targeted leads**. Columns map 1:1 to the CRM
fields used in `qualification-rubric-and-crm.md`.

| Field | Description |
|---|---|
| `company` | Company name. |
| `website_or_linkedin` | Company site or LinkedIn URL (no personal scraped data). |
| `segment` | A / B / C / D. |
| `persona` | QA owner / Eng leader / Agency owner / Data QA owner. |
| `title` | Target contact's role/title. |
| `trigger` | Trigger event observed (see §4). |
| `personalization_angle` | The angle chosen (see §5). |
| `likely_qa_pain` | Pain hypothesis to validate (see §6). |
| `pilot_fit_score` | 1–5 (see §7a). |
| `priority_score` | 1–100 (see §7b). |
| `outreach_status` | Pipeline status (see CRM doc). |
| `next_action` | The single next step. |
| `next_action_date` | When that step is due. |
| `notes` | Free-text context (keep sanitized). |

### Synthetic template row (copy-paste, replace placeholders)

```csv
company,website_or_linkedin,segment,persona,title,trigger,personalization_angle,likely_qa_pain,pilot_fit_score,priority_score,outreach_status,next_action,next_action_date,notes
"Example SaaS Co","https://example.com",A,"QA owner","QA Automation Lead","Hiring SDET (regression in JD)","Regression repetition","Regression repeats each release; knowledge siloed",4,78,"Target account","Send initial message (QA Lead seq #1)","2026-07-02","Synthetic placeholder — replace before use"
```

> Keep the real, populated version of this CSV **out of the public repo**.
> The line above is purely structural.

---

## Acceptance checklist (maps to Issue #48)

- [x] Criteria specific to BQA-OS, not generic lead sourcing (pain angles, domains).
- [x] At least 4 ICP segments covered (A–D).
- [x] Priority scoring makes contact order explicit (§7).
- [x] Template supports 100–200 leads (flat CRM-ready columns).
- [x] No private/scraped data committed (synthetic template only).
- [x] Indexed from this directory's `README.md`.

Next: pick messaging from [discovery-message-pack.md](discovery-message-pack.md).
