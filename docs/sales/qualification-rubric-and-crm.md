# Pilot Qualification Rubric & CRM Pipeline (Issue #53)

A lightweight way to evaluate replies, inbound interest, and discovery calls for
the first **20–30 BQA-OS pilot conversations**, decide whether to push for
discovery / nurture / disqualify, and hand qualified opportunities to the
Solutions Engineer.

> **Privacy:** Keep populated CRM data in a private sheet/CRM. This repo holds the
> **template and rules only** — no real prospect data or private notes.

---

## 1. Qualification rubric

Score each opportunity on **7 dimensions, 0–3 each** (max 21). Score from what
you actually know — leave unknowns at the value you can defend, and use the call
to fill gaps.

| # | Dimension | 0 | 1 | 2 | 3 |
|---|---|---|---|---|---|
| 1 | **Pain** (QA knowledge loss / regression / onboarding / API-GraphQL-ETL bottleneck) | none evident | vague | one clear pain | clear, urgent, matches BQA-OS angle |
| 2 | **Data availability** (10–30 safe artifacts: sanitized logs, bug reports, test notes, regression checklists) | none | maybe | likely available | confirmed available & sanitizable |
| 3 | **Buyer authority** | no path | influencer only | can introduce budget owner | can approve a paid pilot |
| 4 | **Urgency** | none | someday | this quarter | active bottleneck / near-term release or regression pressure |
| 5 | **Pilot fit** (2 weeks, no custom enterprise integration) | needs enterprise build | heavy scoping | mostly scopeable | clean 2-week scope |
| 6 | **Privacy fit** (local-first / sanitized acceptable) | blocked | unclear | likely OK | explicitly OK |
| 7 | **Success-criteria clarity** | can't define | fuzzy | rough idea | can define what "useful output" means |

### Reading the score

| Total (of 21) | Decision |
|---|---|
| **15–21** | **Push for discovery / advance to pilot proposal.** Strong fit. |
| **9–14** | **Nurture.** Real signal but gaps; set a next action to close the biggest gap. |
| **0–8** | **Disqualify or park.** Don't spend pilot capacity here now. |

**Hard gates (regardless of total):** if **Pilot fit = 0** (needs enterprise
integration) or **Privacy fit = 0** (local-first/sanitized blocked), do **not**
advance to a paid pilot — nurture or disqualify with the reason recorded.

---

## 2. Pilot fit scoring vs. the lead-list score

- The **lead-list `pilot_fit_score` (1–5)** in `icp-and-sourcing.md` is a
  pre-contact gut-check used to order outreach.
- This **rubric (0–21)** is the post-contact qualification used to decide
  discovery / nurture / disqualify and to advance pipeline stages.
- When they disagree after a real conversation, **the rubric wins** — update the
  lead's status from the rubric outcome.

---

## 3. CRM pipeline statuses

Minimal pipeline. Every status carries a `next_action` and `next_action_date`.

| Status | Definition | Typical next action |
|---|---|---|
| **Target account** | In lead list, not yet contacted. | Send initial message. |
| **Contacted** | Initial outreach sent, no reply yet. | Follow-up per cadence. |
| **Replied** | Any reply received (not yet qualified). | Score rubric; propose call. |
| **Discovery scheduled** | Discovery call booked. | Prep with `icp-and-sourcing.md` pain hypothesis. |
| **Discovery completed** | Call done; rubric scored. | Decide: propose pilot / nurture / disqualify. |
| **Pilot proposed** | 2-week QA Memory Pilot proposed. | Confirm scope, artifacts, success criteria. |
| **Paid pilot agreed** | Pilot agreed (paid). | Hand off to Solutions Engineer (see §6). |
| **Not now** | Real fit, wrong timing. | Set nurture follow-up date. |
| **Disqualified** | Not a fit. | Record reason (§5); stop outreach. |

### Status flow

```text
Target account → Contacted → Replied → Discovery scheduled → Discovery completed
   → Pilot proposed → Paid pilot agreed → [handoff]
         │
         └─ at any point → Not now  (nurture) | Disqualified (stop)
```

---

## 4. Next-action rules

- **Every record must have a `next_action` and `next_action_date`.** No record
  sits without one (this is the daily hygiene check in the cadence doc).
- **Contacted → no reply:** follow per the sequence in `discovery-message-pack.md`
  (max 4 touches), then move to `Not now` with a future nurture date.
- **Replied:** score the rubric within 1 business day; book a call if ≥9.
- **Discovery completed:** decide within 1 business day — `Pilot proposed`,
  `Not now`, or `Disqualified`.
- **Pilot proposed → no response (7 days):** one nudge, then `Not now`.
- **Any "not interested" / opt-out:** move to `Disqualified` immediately.

---

## 5. Disqualification reasons (pick one)

- `no-pain` — no real QA knowledge-loss / regression / onboarding / API-GraphQL-ETL pain.
- `no-data` — can't provide 10–30 safe/sanitizable artifacts.
- `no-authority-no-path` — no route to a budget owner.
- `privacy-blocked` — local-first / sanitized workflow not acceptable.
- `scope-too-big` — needs enterprise/custom integration; not a 2-week pilot.
- `no-success-criteria` — buyer can't define useful output even after discovery.
- `timing` — real fit but not now → use `Not now` instead of disqualifying.
- `unresponsive` — no engagement after full cadence.
- `opt-out` — explicitly asked to stop.

---

## 6. Handoff notes — Founder / Solutions Engineer / Pilot Manager

When a deal reaches **Paid pilot agreed**, the SDR hands off a short brief so the
Solutions Engineer can start the pilot without re-discovery.

**Who does what**
- **Founder:** loops in on `Pilot proposed`+ for high-fit deals; owns
  pricing/commercial conversation and final go/no-go.
- **Solutions Engineer (SE):** owns pilot kickoff and execution against the
  agreed artifacts and success criteria.
- **Pilot Manager:** owns the 2-week timeline, checkpoints, and the wrap-up
  review (see `7-day-pilot-cadence.md` for the operating rhythm feeding this).

**Handoff brief — fields the SE needs before kickoff**

```text
Company / contact (role, authority)
Primary pain (rubric dim 1) + supporting quotes
Agreed pilot scope (what's in / out for 2 weeks)
Artifacts: which 10–30, source, sanitization status
Privacy constraints (local-first / sanitized requirements)
Success criteria: buyer's definition of "useful output"
Domain surface: API / GraphQL / ETL / data quality / automation
Rubric total + any hard-gate notes
Timeline & key dates; named Pilot Manager
Open questions / risks
```

> SE should not start kickoff until **artifacts are identified and the
> sanitization path is clear**, and **success criteria are written down**.

---

## 7. CRM template (template only — no real data)

Mirrors the lead-table fields in `icp-and-sourcing.md` plus qualification fields.

```csv
company,persona,segment,outreach_status,rubric_pain,rubric_data,rubric_authority,rubric_urgency,rubric_pilotfit,rubric_privacy,rubric_success,rubric_total,decision,disqualify_reason,next_action,next_action_date,owner,notes
"Example SaaS Co","QA owner",A,"Discovery completed",3,2,2,3,3,3,2,18,"Pilot proposed","","Send pilot scope + artifact list","2026-07-04","SDR","Synthetic placeholder — replace before use"
```

## Acceptance checklist (maps to Issue #53)

- [x] Rubric drives push-for-discovery / nurture / disqualify (§1).
- [x] CRM fields support 100–200 leads (flat, template-only).
- [x] Statuses include next action + next action date (§3, §4, §7).
- [x] Handoff notes state what the SE needs before kickoff (§6).
- [x] No real prospect data committed (synthetic template only).
- [x] Indexed from this directory's `README.md`.
