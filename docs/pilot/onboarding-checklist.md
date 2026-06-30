# Pilot onboarding checklist — QA Memory Pilot

The first-customer onboarding flow for a **2-week BQA-OS QA Memory Pilot**. A
Solutions Engineer (SE) can run a kickoff straight from this checklist.

BQA-OS is **local-first**: the build runs on hardware the customer controls, and
their data stays with them. Only **sanitized** artifacts are used.

Related docs:

- [Discovery and proposal templates](discovery-and-proposal-templates.md) — to
  get here, the prospect should pass the pilot readiness checklist.
- [MacBook dogfood guide](macbook-dogfood-guide.md) — the exact `bqa` workflow,
  with expected outputs, that the install steps below follow.
- [Knowledge Review Checklist](../knowledge-review-checklist.md) — the
  artifact-by-artifact review used in the review session.

---

## 0. Prerequisites (before kickoff)

- [ ] Signed/agreed pilot proposal with one focus area (ETL / data quality / API / GraphQL).
- [ ] Named customer QA owner and a named artifact/sanitization owner.
- [ ] Agreed success criteria and a two-week window with three touchpoints.
- [ ] A machine to run the pilot on with **Go** and a terminal
      (see [dogfood guide §1](macbook-dogfood-guide.md#1-prerequisites)).
- [ ] A target AI runtime available for the demo (Codex / Claude Code / OpenCode).

---

## 1. Data request: 10–30 sanitized QA artifacts

Ask the customer to gather **10–30 sanitized** QA artifacts in the focus area:

- [ ] Regression checklists / "things we always check" notes.
- [ ] Post-mortems or bug write-ups for recurring failures.
- [ ] Sanitized log snippets that illustrate a real failure pattern.
- [ ] Prompts or queries that have worked well for QA tasks.
- [ ] Test notes describing pipeline → source → target and expected invariants.

Format: plain `*.md`, `*.log`, or `*.txt` files in a single folder.

### What must **not** be provided

- [ ] No secrets, tokens, API keys, or credentials.
- [ ] No real client/customer names.
- [ ] No internal URLs or hostnames.
- [ ] No PII or regulated data.
- [ ] No raw production dumps.

> **Public examples vs. private client data.** The synthetic samples in this
> repo (e.g. `examples/etl-weekend/`, the dogfood guide) are for illustration
> only. Real customer artifacts are **private**: they live on the customer's
> machine, never in this repo or any shared location. Redaction during ingest is
> a safety net, not a substitute for sanitizing first.

---

## 2. Install

Follow [dogfood guide §2–§3](macbook-dogfood-guide.md#2-build-or-install-the-bqa-cli):

- [ ] Build or install the CLI: `go build -o ./bqa ./cmd/bqa` (or run `install.sh`).
- [ ] Verify: `bqa --help` and `bqa version` work.
- [ ] Create the pilot workspace directory and `cd` into it.
- [ ] Initialize: `bqa init` → expect `BQA workspace initialized in .bqa/`.
- [ ] Place the customer's sanitized artifact folder alongside (e.g. `./qa-artifacts/`).

---

## 3. `bqa discover`

- [ ] Run `bqa discover --global=false --local=true`.
- [ ] Confirm it completes and writes `.bqa/input/sessions/manifest.json`.

> For hand-curated artifact folders, `discover` typically reports `0` AI-session
> files — that is expected. The artifacts are brought in via `ingest --from` in
> the next step.

---

## 4. `bqa ingest`

Import the customer's sanitized artifacts:

- [ ] Run `bqa ingest --from ./qa-artifacts`.
- [ ] Confirm `Discovered:` and `Imported:` counts match the artifact count.
- [ ] **Check the `Redactions:` count.** If `> 0`, stop and review with the
      customer — a secret reached intake and the source must be re-sanitized.
- [ ] Confirm `.bqa/input/sessions/index.json` exists with the right entry count
      (`bqa doctor` will report this too).

---

## 5. `bqa build`

Generate reusable QA knowledge plus starter runtime artifacts:

- [ ] Run `bqa build`.
- [ ] Confirm `Sessions processed` matches the imported count.
- [ ] Confirm `.bqa/knowledge/` contains the nine `*.yaml` artifacts.
- [ ] Validate: `bqa build --check` → `All knowledge artifacts are present and valid.`
- [ ] Health check: `bqa doctor` → `All checks passed.`
- [ ] Generate the grounded runtime context: `bqa codex` → writes
      `.bqa/prompts/bqa-master-context.md` (run `bqa claude` / `bqa opencode`
      instead if that is the customer's runtime).

> Generated knowledge is **heuristic and keyword-driven**. It is reviewed with
> the customer in the next step before being trusted.

---

## 6. 30-minute review session agenda

Run this live with the customer QA owner. Drive the artifact review from the
[Knowledge Review Checklist](../knowledge-review-checklist.md).

| Time | Item |
| --- | --- |
| 0:00–0:03 | Recap focus area, success criteria, and what was imported. |
| 0:03–0:10 | Walk `.bqa/knowledge/` for the focus area against the [Knowledge Review Checklist](../knowledge-review-checklist.md). Customer flags wrong / missing / noisy findings. |
| 0:10–0:15 | Review `project_profile.yaml` signal counts — do they match reality? |
| 0:15–0:23 | Live demo: a grounded QA task in the customer's runtime, referencing `.bqa/prompts/bqa-master-context.md`. Customer judges usefulness. |
| 0:23–0:28 | Capture accept / revise / reject decisions per artifact (below). |
| 0:28–0:30 | Confirm the single next iteration and owner. |

---

## 7. Acceptance decisions for generated artifacts

For each knowledge artifact in the focus area, record one decision:

- [ ] **Accept** — captures real, useful QA knowledge as-is.
- [ ] **Revise** — close but needs added/removed/corrected findings (note what).
- [ ] **Reject** — not useful for this focus area (note why; near-empty
      off-domain artifacts are expected, not failures).

Also record against the agreed success criteria:

- [ ] Did the grounded QA task meet the "usable without rework" bar?
- [ ] Can the customer re-run `bqa build` / `bqa codex` unaided?

---

## 8. Next iteration

- [ ] Apply review feedback: add/sanitize a few more artifacts where coverage was thin.
- [ ] Re-run `bqa ingest --from … && bqa build && bqa codex`.
- [ ] Re-validate: `bqa build --check` and `bqa doctor`.
- [ ] Hand over the `.bqa/` workspace and a short adoption note.
- [ ] (Optional) Install the brain into another of the customer's projects:
      `bqa brain install --from .bqa --target <project>`.
- [ ] Confirm the post-pilot decision and the next concrete step.

---

## Privacy reminder

- The pilot is **local-first**: data stays on customer-controlled hardware.
- Only **sanitized** artifacts enter the workspace; never real client names,
  secrets, tokens, or internal URLs.
- Before sharing anything generated, run `bqa sanitize .bqa` (review the dry-run
  before adding `--write`).
- Keep public examples and private client data strictly separate.
