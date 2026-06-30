# BQA-OS Sales & Growth Playbook — First Pilots Motion

This directory is a **practical, copy-pasteable playbook** for sourcing,
contacting, qualifying, and running the first **2-week QA Memory Pilot**
conversations for BQA-OS.

It is written for a part-time SDR / Growth person working alongside the
Founder, a Solutions Engineer, and a Pilot Manager (these can be the same one
or two people early on). Everything here is generic and synthetic — there are
**no real customers, logos, or invented metrics**. Fill in your own data as you
run the motion.

## What BQA-OS is (so the value prop stays honest)

BQA-OS (Better QA Operating System) is an **AI-native operating system for
quality engineering**. It turns real QA work sessions into **reusable AI agents,
skills, and workflows** that plug into AI coding runtimes (Codex, Claude Code,
OpenCode). Core ideas the sales motion must stay grounded in:

- **QA knowledge becomes reusable**: sessions → sanitized knowledge → skills →
  agents → workflows, instead of living in one engineer's head.
- **Local-first / sanitized**: `bqa sanitize` redacts secrets before anything is
  shared; private value is meant to live in a private BQA Brain or local `.bqa`
  workspace, not in public repos.
- **Supported domains**: Big Data & ETL testing, GraphQL functional testing, API
  testing, contract testing, data quality validation, test automation.
- **The pilot**: a **2-week QA Memory Pilot** — take 10–30 safe QA artifacts,
  decode them into reusable agents/skills/workflows, and show useful QA output.

### What it is **not** (do not claim these)

- Not a fully autonomous QA agent that needs no humans.
- Not a guarantee of specific bug-catch rates, coverage numbers, or ROI.
- Not a system that ingests private/customer data into a public repo.
- Not a replacement for the team's QA judgment — it captures and reuses it.

## The four documents

| # | Document | Use it for |
|---|----------|-----------|
| [icp-and-sourcing.md](icp-and-sourcing.md) | **ICP & sourcing** | Who to target, filters, trigger events, personalization angles, pain hypotheses, priority scoring, lead-table fields. |
| [discovery-message-pack.md](discovery-message-pack.md) | **Discovery message pack** | Persona-specific 4-message sequences to start respectful, human-reviewed pilot conversations. |
| [qualification-rubric-and-crm.md](qualification-rubric-and-crm.md) | **Qualification & CRM** | Score replies and calls, define pipeline statuses + next-action rules, disqualify cleanly, hand off to the Solutions Engineer. |
| [7-day-pilot-cadence.md](7-day-pilot-cadence.md) | **7-day operating cadence** | A day-by-day plan a part-time SDR can run in 30–90 min/day to turn the artifacts above into booked discovery calls. |

## How they fit together

```text
ICP & sourcing ─▶ build a prioritized lead list (segments, scores, CRM fields)
        │
        ▼
Discovery message pack ─▶ send human-reviewed, persona-specific outreach
        │
        ▼
Qualification & CRM ─▶ score replies/calls, move pipeline stages, hand off fit ones
        │
        ▼
7-day cadence ─▶ the daily operating rhythm that runs the whole loop and reviews metrics
```

Read them in that order the first time. After that, the **7-day cadence** is the
daily driver and the other three are reference material it points back to.

## Operating principles (apply to every doc here)

1. **Human-reviewed, not automated.** Every message is read and tailored by a
   person before it goes out. No bulk sending, no scraping personal data.
2. **Honest framing.** Lead with the pain and the pilot; avoid hype and
   unsupported claims. "Example wording" in the message pack is illustrative.
3. **Privacy first.** Mention local-first / sanitized data where relevant. Never
   commit real prospect data, customer notes, or secrets to this (public) repo.
4. **Small and time-boxed.** The offer is a 2-week pilot on 10–30 safe artifacts,
   not an open-ended enterprise integration.

## Related internal material

The `bqa build --sales-package` command generates a complementary internal
"Monday sales package" (pilot one-pager, demo script, discovery call script,
onboarding checklist, sample outreach, pricing hypothesis, stakeholder FAQ). The
documents here are the **public, repo-tracked, reusable framework**; the
generated package is the **per-run, internal** companion. Keep private artifacts
out of this directory.
