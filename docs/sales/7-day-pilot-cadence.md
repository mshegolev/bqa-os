# 7-Day Pilot Pipeline Operating Cadence (Issue #54)

A practical **7-day operating plan** that turns the artifacts in this directory
into **booked discovery calls** for the BQA-OS **2-week QA Memory Pilot**.

Designed for a **part-time SDR / Growth** person spending **30–90 minutes/day**,
working with the Founder, Solutions Engineer (SE), and Pilot Manager.

> **No automated bulk outreach, no paid ads, no enterprise process, no unlimited
> free pilots.** Every message is human-reviewed. No real lead data in this repo.

## Inputs (read once before Day 1)

- `icp-and-sourcing.md` — ICP, filters, triggers, priority scoring, lead fields.
- `discovery-message-pack.md` — persona sequences and reply handling.
- `qualification-rubric-and-crm.md` — rubric, statuses, next-action rules, handoff.

## Targets for the week (adjust to your capacity)

- **30–50 targeted leads** sourced.
- All leads contacted with **human-reviewed** first messages.
- **Goal: 2–4 discovery calls scheduled** by end of week (volume-dependent;
  treat as a directional target, not a guarantee).

---

## Daily plan

Each day lists the **focus**, the **concrete output**, and the **time box**.

### Day 1 — Finalize ICP filters & lead-list template (45–60 min)
- Confirm which 1–2 segments (A–D) to lead with this week.
- Lock company filters and trigger definitions from `icp-and-sourcing.md`.
- Set up the private lead sheet using the CRM template fields.
- **Output:** ready lead template + chosen segments + filter notes.

### Day 2 — Source first 30–50 leads (60–90 min)
- Source within filters; add `trigger` and `personalization_angle` per lead.
- Score `pilot_fit_score` (1–5) and `priority_score` (1–100); sort descending.
- **Output:** 30–50 leads scored and prioritized; status `Target account`.

### Day 3 — Send first human-reviewed messages (45–75 min)
- Work top of the priority list. Pick the matching persona sequence; fill
  personalization slots; **read each message before sending.**
- Move sent leads to `Contacted`; set `next_action` = follow-up + date.
- **Output:** initial messages out to the top of the list (aim ~15–25).

### Day 4 — Follow up & refine personalization (30–60 min)
- Send follow-ups (sequence #2) to no-replies whose date is due.
- Note which angles draw replies; refine wording for the next batch.
- Continue first-touch on remaining sourced leads.
- **Output:** follow-ups sent; refined angle notes; more first-touch coverage.

### Day 5 — Run / schedule first discovery calls (45–75 min)
- Book calls with positive replies; for live calls, prep with the segment pain
  hypothesis. Score the rubric on any completed calls.
- Move records to `Discovery scheduled` / `Discovery completed`.
- **Output:** calls booked/run; rubric scored; statuses updated.

### Day 6 — Pilot-fit summaries for positive replies (30–60 min)
- For rubric ≥15, draft the handoff brief (`qualification-rubric-and-crm.md` §6);
  loop in Founder/SE. Move clear fits toward `Pilot proposed`.
- For 9–14, set a nurture next action; for ≤8, disqualify with a reason.
- **Output:** pilot-fit summaries for strong replies; clean decisions logged.

### Day 7 — Review, update messaging, pick next segment (30–45 min)
- Compute the weekly metrics (below); answer the review questions.
- Update the message pack with what worked; choose next week's segment.
- **Output:** weekly metrics + decisions + next-segment plan.

---

## Reply handling rules (every day)

- **Positive →** book the call; score rubric same/next business day.
- **Question/objection →** answer honestly within the messaging rules; offer call.
- **"Not now" →** status `Not now`, set nurture date, stop the sequence.
- **"Not interested"/opt-out →** `Disqualified` (`opt-out`); stop immediately.

## Follow-up cadence

- Max **4 touches** per lead (initial + 3), spaced ~2–3 days, per
  `discovery-message-pack.md`.
- Stop on any reply or opt-out. After 4 touches with no reply → `Not now` with a
  future nurture date (don't burn the lead with more cold touches).

---

## Daily CRM hygiene checklist (5 min, every day)

- [ ] Every active record has a `next_action` and a `next_action_date`.
- [ ] Today's due `next_action_date` items are actioned or rescheduled.
- [ ] New replies have a status and (if applicable) a rubric score.
- [ ] No real secrets / private data entered into shared/public locations.
- [ ] Disqualified records have a reason; `Not now` records have a nurture date.

---

## Weekly review metrics (Day 7)

Track week over week:

| Metric | Definition |
|---|---|
| Targeted leads added | New leads sourced into the list. |
| Contacts reached | Leads that received a first message. |
| Replies | Any reply received. |
| Positive replies | Replies showing genuine interest. |
| Discovery calls scheduled | Calls booked. |
| Pilot opportunities | Reached `Pilot proposed`+. |
| Disqualified leads | Moved to `Disqualified`. |
| Top pain patterns | Most-common pains heard (qualitative). |

### Review questions

- Which **segment** and which **angle** produced the best reply rate?
- Where did leads **stall** (no reply, no-show, post-discovery), and why?
- Which **pain patterns** showed up most — should messaging or ICP shift?
- Is the **2-week pilot scope** landing, or do prospects want something different?
- What single change would most improve next week's discovery-call count?

---

## When to involve others

- **Founder:** any `Pilot proposed`+ high-fit deal, pricing/commercial questions,
  or a segment-level pivot decision.
- **Solutions Engineer:** once a deal reaches `Paid pilot agreed` — receives the
  handoff brief and owns kickoff (artifacts + sanitization + success criteria).
- **Pilot Manager:** owns the 2-week pilot timeline and checkpoints once a pilot
  starts.

## Acceptance checklist (maps to Issue #54)

- [x] Specific to the BQA-OS pilot motion (ties to the other three docs).
- [x] Runnable by a part-time SDR in 30–90 min/day.
- [x] Clear daily outputs.
- [x] Metrics + review questions included.
- [x] Rules for involving Founder / SE / Pilot Manager.
- [x] No private lead data committed.
- [x] Indexed from this directory's `README.md`.
