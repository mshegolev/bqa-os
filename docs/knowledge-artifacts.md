# Knowledge artifact schema (v1)

`bqa build` generates YAML knowledge artifacts under `.bqa/knowledge/` — one per
domain plus a project profile. This document is the reference for their **v1**
schema: the common envelope, the per-artifact shapes, the field derivation
rules, and the forward-compatibility policy.

## Purpose

The knowledge artifacts capture project-specific QA signals extracted from
normalized sessions so downstream consumers (`bqa codex`, `bqa build --check`,
`bqa doctor`, and future export/import and learning features) can read a
stable, versioned contract rather than an ad-hoc shape.

**Hard cutover — artifacts are derived, not hand-authored.** `bqa build` always
writes v1. There is no in-place migrator: on any schema change, regenerate the
files by re-running `bqa build`. Do not edit `.bqa/knowledge/*.yaml` by hand.

Output is **deterministic**: identical synthetic input produces byte-identical
output. There is no wall-clock timestamp; provenance is limited to
`generated_by` so the files diff cleanly and tests can compare bytes.

## Common envelope (every artifact)

Every artifact begins with the same three flat top-level keys:

```yaml
schema_version: 1
kind: <kind>            # etl_patterns | graphql_patterns | api_patterns |
                        # data_quality_patterns | common_bugs | successful_prompts |
                        # droid_patterns | runtime_patterns | project_profile
generated_by: bqa dev   # internal/version.Version ("dev" in tests → deterministic;
                        # vX.Y.Z in a release build). Metadata / provenance only.
```

There is no `generated_at` wall-clock field — determinism is worth more than a
timestamp here, and provenance can be enriched later without a version bump.

## Pattern artifacts (8 files)

`etl_patterns`, `graphql_patterns`, `api_patterns`, `data_quality_patterns`,
`common_bugs`, `successful_prompts`, `droid_patterns`, `runtime_patterns` all
share the same shape (synthetic example):

```yaml
schema_version: 1
kind: etl_patterns
generated_by: bqa dev
patterns:
  - id: etl-0a1b2c3d
    name: "row-count reconciliation"
    domain: etl
    evidence: "...evidence snippet (rune-safe, length-bounded)..."
    source: "normalized/etl/s1.md"
    reusable_check: "compare source vs target row counts for the window"
    confidence: medium
```

The empty case renders `patterns: []`.

`kind` is file-level; `domain` is finding-level (they differ — e.g.
`common_bugs` entries carry the failing area as `domain`). The list key is
`patterns:` in **all** pattern files (uniform for generic parsing), including
`successful_prompts`.

### Pattern fields

| Field | Meaning |
|---|---|
| `id` | Content hash `<domain>-<8 hex>` — `<domain>-<first 8 hex of sha256(domain + "\|" + name + "\|" + source)>`. Content-derived, so it is stable when other findings are inserted/removed, and fully deterministic. |
| `name` | Short human-readable finding name. |
| `domain` | Finding-level domain (may differ from the file `kind`). |
| `evidence` | A rune-safe, length-bounded snippet from the normalized session. |
| `source` | The normalized session path the finding came from. |
| `reusable_check` | A per-domain check **candidate — review before use**, not an extracted command. |
| `confidence` | Coarse band `low` / `medium` / `high` (see rule below). A heuristic, not a probability. |

### `confidence` rule

`confidence` is computed locally per finding from the number of **distinct
domain keywords** present in its `evidence`:

- `high` — ≥ 3 distinct domain signals in the evidence
- `medium` — exactly 2
- `low` — 1 (or a minimal/single-keyword match)

Documented honestly: confidence reflects how many corroborating domain signals
appear in the evidence; it is a heuristic, not a statistical probability.
Cross-session corroboration is deliberately deferred as a future **additive**
improvement within v1.

### `reusable_check` templates

`reusable_check` is a deterministic per-domain (and matched-sub-signal)
template — a candidate to review before use:

| domain | reusable_check |
|---|---|
| etl (reconciliation) | `compare source vs target row counts for the window` |
| etl (null/dup) | `assert no unexpected nulls or duplicate keys` |
| graphql | `assert query/mutation response shape and error handling` |
| api | `assert endpoint status code and response contract` |
| data_quality | `assert null / duplicate / schema-drift rules pass` |
| common_bugs | `add a regression check reproducing the failure signal` |
| successful_prompts | *(the reusable prompt text itself)* |
| droid / runtime | templated per domain |

## project_profile.yaml

The project profile uses a `profile:` block instead of a `patterns:` list
(synthetic example):

```yaml
schema_version: 1
kind: project_profile
generated_by: bqa dev
profile:
  sessions_analyzed: 12
  domains_detected: [etl, graphql, api]      # domains with signals > 0
  signals: { etl: 8, graphql: 3, api: 2, data_quality: 0, droid: 0, runtime: 0 }
  suggested_next_reviews:
    - "Review etl coverage (8 signals)."
    - "Review graphql coverage (3 signals)."
  maturity: initial                           # existing field, retained
```

### Profile fields

| Field | Meaning |
|---|---|
| `sessions_analyzed` | Number of normalized sessions analyzed for this build. |
| `domains_detected` | The domains with signals > 0, ordered by signal count descending then name (empty list when none). |
| `signals` | Per-domain signal counts for `etl` / `graphql` / `api` / `data_quality` / `droid` / `runtime`. |
| `suggested_next_reviews` | One line per detected domain (`signals > 0`), ordered by signal count descending then name: `"Review <domain> coverage (<N> signals)."` Empty when no domains detected. |
| `maturity` | Coarse maturity marker (currently `initial`). |

## Forward-compatibility policy

- `schema_version` is an integer; **v1** is current.
- An **additive** change (a new optional field) does **not** bump
  `schema_version`; consumers **must ignore unknown fields**.
- A **breaking** change (remove/rename a field, change its type or semantics)
  bumps `schema_version`.
- Consumers **must read `schema_version`** and, on an unsupported major, fail
  with a clear message (`bqa build --check` / `bqa doctor` report it; `bqa codex`
  degrades gracefully).
- Artifacts are **derived** — there is no in-place migrator; on any schema
  change, regenerate with `bqa build`.
