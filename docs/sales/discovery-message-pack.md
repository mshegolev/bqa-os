# Discovery Conversation Message Pack (Issue #51)

Persona-specific message sequences for starting **respectful, human-reviewed**
pilot-discovery conversations for the BQA-OS **2-week QA Memory Pilot**.

> **These are not for automation or bulk sending.** Every message is read,
> tailored, and sent by a person. The CTA always asks for a *conversation*, never
> a hard sell. Wording marked "example" is illustrative, not a guaranteed result.

## Messaging rules (apply to every message)

- No hype, no buzzword stacking.
- No promise of fully autonomous QA — BQA-OS captures and reuses the team's QA
  knowledge; humans stay in the loop.
- No unsupported claims, no invented metrics, no fake customers/logos.
- Mention **local-first / sanitized** handling where data comes up.
- Frame the offer as a **2-week QA Memory Pilot**.
- Anchor on real pains: reusable QA knowledge, regression workflows, QA
  onboarding, API/GraphQL/ETL checks.

## Personalization slots (used across sequences)

- `{first_name}` — contact first name.
- `{company}` — company name.
- `{trigger}` — the observed trigger (e.g. "your GraphQL API launch").
- `{pain_angle}` — chosen angle from `icp-and-sourcing.md` §5.
- `{surface}` — API / GraphQL / ETL / data pipelines, as relevant.

Each sequence has **4 short messages**: (1) initial, (2) follow-up with a
concrete pain, (3) value-proof-style note (example wording only), (4) polite
close-the-loop. Send #2–#4 only if there's no reply; **stop immediately on any
"not interested."**

---

## Sequence 1 — QA Lead / QA Automation Lead (Segment A)

**1. Initial**
> Hi {first_name} — saw {trigger} at {company}. Quick one: how much of your
> regression is re-running the same checks each release, and where does that QA
> knowledge actually live? Open to a short chat?

**2. Follow-up (concrete pain)**
> Following up, {first_name}. The pattern I keep hearing from QA leads: the
> hard-won regression and project-specific knowledge sits in one or two people,
> so onboarding and repeat checks stay slow. Is that close to your reality?

**3. Value-proof note (example wording only)**
> For context on what I'm working on: BQA-OS turns existing QA sessions, test
> notes, and regression checklists into reusable agents and workflows — local-first
> and sanitized, so nothing sensitive leaves your side. *Example:* a team's
> regression checklist becomes a reusable workflow a new hire can run on day one.

**4. Close the loop**
> Last note, {first_name} — don't want to crowd your inbox. If reusable
> regression/onboarding QA knowledge is worth 20 minutes, I'd love to compare
> notes. If now's not the time, no problem — happy to circle back later.

**CTA:** "Worth a 20-min call to see if a 2-week QA Memory Pilot fits?"

---

## Sequence 2 — CTO / VP Engineering (Segment B)

**1. Initial**
> Hi {first_name} — {trigger} suggests {company} is shipping fast. Curious:
> with AI coding speeding up dev, is QA keeping pace or becoming the constraint?
> Happy to trade notes.

**2. Follow-up (concrete pain)**
> Quick follow-up. The recurring story: AI coding lifts dev throughput, then QA
> becomes the bottleneck and release-quality risk creeps up — with no system to
> capture QA know-how at the new pace. Sound familiar at {company}?

**3. Value-proof note (example wording only)**
> What I'm building, briefly: BQA-OS is an AI-native QA layer that turns real QA
> work into reusable agents/skills/workflows across {surface}. Local-first and
> sanitized by default. *Example:* QA knowledge scales alongside AI-accelerated
> dev instead of lagging it. No autonomous-QA magic claimed — humans stay in the
> loop.

**4. Close the loop**
> Final note — I'll stop here unless useful. If the dev-vs-QA throughput gap is on
> your radar, a 2-week pilot on 10–30 safe artifacts is a low-risk way to test it.
> Worth a short call?

**CTA:** "Open to a 20-min call on closing the dev/QA throughput gap?"

---

## Sequence 3 — QA Consultant / Boutique QA Agency Owner (Segment C)

**1. Initial**
> Hi {first_name} — running QA across clients, reusable QA assets seem like real
> leverage. How do you carry regression/onboarding knowledge between engagements
> today? Curious how you've solved it.

**2. Follow-up (concrete pain)**
> Following up. What I hear from agency owners: you re-derive similar QA knowledge
> per client, and a lot walks out when a contractor rolls off. Is reusability a
> margin lever you're actively working on?

**3. Value-proof note (example wording only)**
> For context: BQA-OS turns QA sessions into reusable, sanitized agents and
> workflows you can apply across engagements — private value stays in your own
> workspace. *Example:* a regression playbook built once becomes reusable client
> to client, with secrets redacted before anything is shared.

**4. Close the loop**
> Last one, {first_name}. If turning client QA work into reusable assets is
> interesting, I'd value 20 minutes to compare approaches — and there's a 2-week
> pilot if you want to try it on a real (sanitized) example. No pressure either way.

**CTA:** "Worth 20 min on reusable QA assets across clients?"

---

## Sequence 4 — Big Data / ETL QA Lead or Data-Platform QA Owner (Segment D)

**1. Initial**
> Hi {first_name} — saw {trigger} at {company}. How manual is your ETL /
> data-quality validation right now, and how much of it is the same checks each
> run? Curious how your team handles it.

**2. Follow-up (concrete pain)**
> Quick follow-up. The pattern on data teams: ETL and data-quality validation is
> repetitive and manual, and pipeline quirks live as tribal knowledge — tough to
> onboard QA into. Does that match {company}?

**3. Value-proof note (example wording only)**
> Briefly: BQA-OS supports ETL/Big-Data testing and data-quality validation as
> first-class domains, turning your validation knowledge into reusable workflows —
> local-first and sanitized. *Example:* a set of data-quality rules becomes a
> reusable workflow new data-QA hires can run, instead of re-learning each pipeline.

**4. Close the loop**
> Final note — won't keep nudging. If repetitive ETL/data-quality validation is a
> real cost, a 2-week pilot on 10–30 safe artifacts is a low-commitment way to see
> the reuse. Happy to walk through it on a quick call if useful.

**CTA:** "Open to 20 min on reusable ETL/data-quality validation?"

---

## Reply handling (quick reference)

- **Positive / curious →** book the discovery call; score in
  `qualification-rubric-and-crm.md`; loop in the Founder if it looks like a fit.
- **Question / objection →** answer honestly, stay within the messaging rules,
  offer the call. Never overclaim to win the reply.
- **"Not now" →** mark `Not now`, set a polite follow-up date, stop the sequence.
- **"Not interested" / unsubscribe →** stop all outreach immediately; mark
  disqualified with reason.

## Acceptance checklist (maps to Issue #51)

- [x] 4 persona-specific sequences (A–D).
- [x] Each sequence has 4 short messages.
- [x] Each includes a concrete BQA-OS pain angle.
- [x] CTAs ask for a conversation, not a hard sell.
- [x] No private references or fake claims; example wording labeled as such.
- [x] Local-first / sanitized framing included; positioned as a 2-week pilot.
- [x] Indexed from this directory's `README.md`.
