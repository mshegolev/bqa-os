# Knowledge Review Checklist

A practical human-review checklist for generated `.bqa/knowledge/*.yaml`
artifacts. Generated knowledge is heuristic and keyword-driven (see
[knowledge-extractor.md](knowledge-extractor.md)); it must pass a human review
before it is trusted as reusable QA memory.

This checklist is designed for a **single 30-minute review session** of one
build's output. Reviewers do not need to read raw session logs — every check can
be answered from the artifact text plus a quick glance at the cited `source`.

> All examples below are **synthetic**. Do not paste real session logs, repo
> data, secrets, or customer records into reviews or into the artifacts.

---

## What you are reviewing

`bqa build` writes these artifacts into `.bqa/knowledge/`:

| File | What it should capture |
| --- | --- |
| `etl_patterns.yaml` | ETL / Big Data testing patterns (Airflow, Spark, Hive, reconciliation, row counts) |
| `graphql_patterns.yaml` | GraphQL functional / contract testing patterns |
| `api_patterns.yaml` | REST / API contract testing patterns |
| `data_quality_patterns.yaml` | Data-quality validation (null checks, schema drift, duplicates, checksums) |
| `common_bugs.yaml` | Recurring bug / failure patterns |
| `successful_prompts.yaml` | Prompt candidates that worked well |
| `droid_patterns.yaml` | factory.ai / droid session patterns |
| `runtime_patterns.yaml` | Runtime execution patterns (tool calls, sandbox, approvals) |
| `project_profile.yaml` | Aggregate signal counts and maturity |

Every artifact opens with the same v1 envelope (`schema_version` / `kind` /
`generated_by`). Each pattern entry has the same shape:

```yaml
schema_version: 1
kind: etl_patterns
generated_by: bqa dev
patterns:
  - id: "etl-0a1b2c3d"
    name: "reconciliation_row_count_check"
    domain: "etl"
    evidence: "verified target row count matched source after the spark reconciliation job"
    source: "normalized/claude/session-0007.json"
    reusable_check: "compare source vs target row counts for the window"
    confidence: medium
```

`project_profile.yaml` is counts only:

```yaml
schema_version: 1
kind: project_profile
generated_by: bqa dev
profile:
  sessions_analyzed: 12
  domains_detected: [etl, runtime, data_quality, api]
  signals:
    etl: 9
    graphql: 0
    api: 3
    data_quality: 4
    droid: 0
    runtime: 7
  suggested_next_reviews:
    - "Review etl coverage (9 signals)."
  maturity: initial
```

---

## How to run a 30-minute session

Suggested time budget per build:

| Step | Time | Focus |
| --- | --- | --- |
| 1. Completeness scan | 4 min | All expected files present, profile sanity |
| 2. Per-artifact review | 16 min | Domain specificity, evidence quality, rejection criteria |
| 3. Cross-cutting checks | 5 min | Repeated-workflow potential, missed bugs, privacy |
| 4. Decision + follow-ups | 5 min | Keep / rewrite / discard / ask, write follow-up questions |

Record one decision **per artifact**, not one decision for the whole build. A
strong `etl_patterns.yaml` and a junk `graphql_patterns.yaml` should get
different verdicts.

---

## 1. Artifact completeness

- [ ] All nine expected files exist in `.bqa/knowledge/`.
- [ ] Each artifact opens with the v1 envelope (`schema_version: 1`,
      `kind: <filename minus .yaml>`, `generated_by:`), and `kind` matches the
      filename (e.g. `etl_patterns.yaml` → `kind: etl_patterns`).
- [ ] Empty domains render an explicit empty list (`patterns: []`), **not** a
      malformed or missing key.
- [ ] Every pattern entry has all fields populated: `id`, `name`, `domain`,
      `evidence`, `source`, `reusable_check`, `confidence`. An entry missing
      `evidence` or `source` is incomplete and cannot be verified — flag it.
- [ ] `project_profile.yaml` signal counts are roughly consistent with the
      number of entries in the matching artifacts (e.g. `signals.etl: 9` but
      `etl_patterns` has 0 entries is a red flag — investigate before trusting
      anything in the build).
- [ ] `sessions_analyzed` is greater than zero. A build over zero sessions
      produces empty artifacts that should be discarded, not kept.

---

## 2. Domain specificity

A good artifact tells you something specific about *this* project's QA work, not
a generic restatement of the domain.

For each non-empty artifact, confirm:

- [ ] `domain` matches the file (no `domain: "api"` entries living in
      `etl_patterns.yaml`).
- [ ] `name` is a specific, named behavior or check, not a bare keyword.
- [ ] The entry would help a QA engineer who has never seen this project.

Synthetic contrast:

| Verdict | `name` | Why |
| --- | --- | --- |
| Specific | `partition_aware_reconciliation_per_dag_run` | Names a concrete check tied to a real workflow |
| Specific | `graphql_persisted_query_hash_mismatch_check` | Concrete GraphQL contract behavior |
| Generic | `etl` | Bare keyword, no behavior |
| Generic | `api_testing` | Restates the domain, says nothing project-specific |

### Per-domain specificity expectations

- **ETL** — should reference a real pipeline concern: source-vs-target row
  counts, partition / scheduler boundaries, reconciliation after a Spark/Hive
  job, late-arriving data, DAG-run scoping. Reject entries that only say "ran an
  ETL job."
- **GraphQL** — should reference schema/contract behavior: field deprecation,
  nullability changes, query/mutation contract drift, persisted-query handling.
  Reject entries whose only evidence is an environment variable like
  `github_graphql_url` (this is the known false-positive the extractor tries to
  filter — verify it actually did).
- **API** — should reference contract behavior: status codes, pagination,
  auth/error envelopes, request/response schema validation. Reject entries that
  only say "called an endpoint."
- **Data quality** — should reference a concrete rule: null/not-null checks,
  uniqueness/duplicate detection, schema drift, checksum/row-count
  reconciliation. Reject vague "checked the data" entries.

---

## 3. Evidence quality

The `evidence` field is the quoted snippet that justifies the pattern. It is the
single most important field — a pattern with weak evidence is unverifiable.

- [ ] `evidence` is a real sentence/snippet describing what happened, not a
      single matched keyword.
- [ ] The evidence actually demonstrates the named behavior (the snippet about
      "row count mismatch" supports a reconciliation pattern; a snippet that just
      contains the word "spark" does not).
- [ ] `source` points at a plausible normalized session path under
      `.bqa/input/sessions/normalized/`. Spot-check 1–2 sources per artifact.
- [ ] The same `(name, source)` pair does not appear many times (the extractor
      dedupes, but verify a domain isn't dominated by one repeated snippet).
- [ ] Evidence is **not** drawn from config/boilerplate (env vars, sample
      fixtures, README text) — those are noise, not QA experience.

Synthetic contrast:

| Verdict | `evidence` |
| --- | --- |
| Strong | `"row count mismatch: source had 10,422 rows, target had 10,418 after the nightly reconciliation; fixed by reprocessing the late partition"` |
| Strong | `"mutation returned 200 but error envelope had code FORBIDDEN; contract test now asserts on the envelope, not the status"` |
| Weak | `"spark"` (single keyword) |
| Weak | `"see github_graphql_url in the workflow env"` (config noise, not testing) |

---

## 4. Repeated-workflow potential

The point of QA memory is to capture work that *recurs*. Single-occurrence,
one-off observations rarely earn a place.

- [ ] The pattern describes something that will happen again (a recurring check,
      a class of bug, a reusable verification step) — not a one-time incident.
- [ ] If a pattern appears across multiple `source` sessions, treat it as
      higher-value and prefer to **keep** it.
- [ ] Could this pattern become a `.bqa/skill` or `workflow` step? If yes, note
      it as a candidate in the follow-ups (e.g. an ETL reconciliation pattern
      that maps to an `etl-verification-workflow` step).
- [ ] Reject patterns that are purely incidental to a single session and carry
      no reusable lesson.

---

## 5. Missed bug patterns

Review `common_bugs.yaml` last, and cross-reference it against the other
artifacts — recurring failures often hide as "patterns" elsewhere.

- [ ] Recurring failures mentioned in ETL/API/GraphQL/DQ evidence are also
      captured in `common_bugs.yaml`. If a reconciliation mismatch shows up three
      times but `common_bugs.yaml` is empty, the build **missed a bug pattern**.
- [ ] Each captured bug names a symptom *and* (where the evidence allows) a root
      cause or fix, not just "something failed."
- [ ] Look for absent-but-expected categories for this project's domains, e.g.:
  - ETL: schema drift, late/duplicate partitions, row-count drift, timezone/DST
    boundary errors.
  - API: pagination off-by-one, auth-token expiry, inconsistent error envelopes.
  - GraphQL: N+1 resolver explosions, nullability regressions, breaking field
    removals.
  - Data quality: silent null inflation, duplicate keys, checksum mismatch.
- [ ] Note any missed bug pattern as a follow-up; a systematic miss is a reason
      to **ask for more context** rather than keep.

---

## 6. Privacy review

Generated knowledge can be shared, reused, and shipped in packs. It must be
clean. **Any single privacy hit blocks "keep" for that artifact** until fixed.

- [ ] No secrets: passwords, tokens, API keys, JDBC connection strings,
      private URLs/hostnames, certificates.
- [ ] No PII / customer data: names, emails, phone numbers, account IDs,
      customer record contents.
- [ ] No internal-only identifiers that should not leave the repo (internal
      hostnames, ticket contents pasted verbatim, employee handles).
- [ ] `source` paths reference normalized session files only — not absolute paths
      that leak a user's home directory or machine layout.
- [ ] Evidence snippets are sanitized: a real row-count value is fine; a real
      customer email inside that snippet is not.

If an artifact is otherwise good but contains sensitive evidence, the verdict is
**rewrite** (sanitize the snippet), not keep.

---

## 7. Review-session questions

Ask these out loud while reviewing each artifact. They surface most problems
fast:

1. If I handed this artifact to a QA engineer new to the project, would it make
   them more effective — or just restate the obvious?
2. Does each `name` describe a behavior I could turn into a test or a workflow
   step?
3. Does each `evidence` snippet actually prove its `name`, or is it just a
   keyword match?
4. Is anything here a one-off incident dressed up as a reusable pattern?
5. Do the `project_profile` counts match what I'm seeing in the artifacts?
6. Would I be comfortable shipping every line of this artifact in a public pack?
7. What recurring failure from the sessions is *missing* from `common_bugs.yaml`?

---

## 8. Decision

Record one verdict per artifact.

### Keep

All of:

- [ ] Completeness, specificity, and evidence checks pass.
- [ ] At least one pattern has clear repeated-workflow value.
- [ ] Privacy review is clean (zero hits).

→ Promote the artifact to trusted QA memory.

### Rewrite

Any of:

- [ ] Good signal buried in noise — real patterns are present but mixed with
      generic/keyword-only entries to prune.
- [ ] Evidence is sound but a snippet contains sensitive data to sanitize.
- [ ] `name`s are vague but fixable into specific behaviors.

→ Edit the artifact in place: drop the junk entries, sanitize evidence, rename
vague patterns. Re-run this checklist on the rewritten file.

### Discard

Any of:

- [ ] Every entry is generic / keyword-only with no project-specific behavior.
- [ ] Evidence is entirely config noise or single-keyword matches.
- [ ] The artifact is empty for a domain the project doesn't actually do (an
      empty `graphql_patterns.yaml` on a pure-ETL project is correct — discard
      the *file from review*, don't fabricate entries).

→ Do not promote. An empty/irrelevant artifact is not a failure; it just isn't
memory.

### Ask for more context

Any of:

- [ ] `project_profile` counts and artifact contents disagree (possible
      extraction bug).
- [ ] A clear bug pattern recurs in evidence but is missing from
      `common_bugs.yaml` (systematic miss).
- [ ] You can't tell whether a pattern is reusable without seeing more sessions.

→ Capture the follow-up questions below, request more sessions / a re-run, and
re-review.

---

## Follow-up questions

Attach the relevant questions to the verdict so the next reviewer (or the build
owner) has context:

- Which sessions produced this pattern, and did it recur across more than one?
- Is this `name` specific enough to become a test, skill, or workflow step?
- Does the cited `evidence` actually demonstrate the behavior, or is it a
  keyword false-positive (e.g. a GraphQL env var)?
- Are there recurring failures in the sessions that `common_bugs.yaml` missed?
- Do `project_profile` signal counts match the artifact contents? If not, is the
  extractor over- or under-counting a domain?
- Does any evidence snippet contain data we cannot ship publicly?
- For ETL/GraphQL/API/data-quality specifically: is the captured behavior tied
  to a concrete check (row counts, schema/contract drift, status/error
  envelopes, null/duplicate rules) — or is it just naming the domain?
