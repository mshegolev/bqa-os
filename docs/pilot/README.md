# BQA-OS Pilot Ops Pack

A cohesive set of public-safe docs for dogfooding BQA-OS locally and running the
first **2-week QA Memory Pilot** end to end — from your own MacBook through to a
pilot proposal.

Everything here uses **synthetic** examples only. Keep real logs, secrets,
client names, tokens, internal URLs, and prospect data out of this repo.

## Contents

| Doc | Use it to |
| --- | --- |
| [MacBook dogfood guide](macbook-dogfood-guide.md) | Run BQA-OS locally on 5 synthetic ETL checks: `bqa init` → `discover` → `ingest --from` → `build` → `codex`, with expected outputs and troubleshooting. |
| [Discovery and proposal templates](discovery-and-proposal-templates.md) | Run pilot discovery calls, write a pilot proposal, and qualify readiness — with scope boundaries and privacy notes. |
| [Onboarding checklist](onboarding-checklist.md) | Run a customer kickoff for the 2-week QA Memory Pilot: data request, install, build, a 30-minute review session, and acceptance decisions. |
| [First-week founder plan](first-week-founder-plan.md) | A one-page 7-day operating plan to reach the first paid pilot conversations, with measurable outcomes and anti-scope-creep warnings. |

## Suggested path

1. **Prove it locally** with the [dogfood guide](macbook-dogfood-guide.md).
2. **Plan the week** with the [first-week founder plan](first-week-founder-plan.md).
3. **Run conversations** with the [discovery and proposal templates](discovery-and-proposal-templates.md).
4. **Deliver the pilot** with the [onboarding checklist](onboarding-checklist.md).

## Related repo docs

- [Knowledge Extractor](../knowledge-extractor.md) — what `bqa build` produces.
- [Knowledge Review Checklist](../knowledge-review-checklist.md) — how to review
  generated `.bqa/knowledge/*.yaml`.
- [Runtime Adapters](../runtimes.md) — Codex / Claude Code / OpenCode targets.

## Honesty and scope

- The 2-week QA Memory Pilot is **bounded**, not an unlimited free pilot.
- BQA-OS grounds and accelerates QA; it does **not** promise fully autonomous QA.
- Generated knowledge is heuristic and **human-reviewed** before it is trusted.
- BQA-OS is **local-first**: customer data stays on customer hardware.
