# Design: Stable v1 schema for knowledge artifacts (issue #16)

## Context

`bqa build` generates YAML knowledge artifacts under `.bqa/knowledge/` (one per
domain plus a project profile). Today they have an ad-hoc, unversioned shape:

```yaml
etl_patterns:
  - name: "..."
    domain: "..."
    evidence: "..."
    source: "..."
```

Without a stable, versioned schema these files are hard to diff, validate, and
evolve — and they are the foundation for two planned features: **#59** (memory
export/import backends) and **#40** (continuous learning / memory governance).
Issue #16 asks for a minimal but stable v1 schema with required fields, a
schema doc, and tests that verify required fields are present.

Decisions taken during brainstorming:

1. **Ambition:** full versioned contract now (envelope + `schema_version` +
   forward-compatibility policy), so #59/#40 can build on it without a reshape.
2. **Envelope form:** flat top-level metadata keys + a single uniform `patterns:`
   list (a `profile:` block for the project profile).
3. **Confidence:** a coarse `low`/`medium`/`high` band derived by a documented,
   deterministic rule — no false-precision numeric score.
4. **Transition:** hard cutover. `bqa build` always writes v1; old files are
   simply regenerated (artifacts are derived, not hand-authored). No in-place
   migrator. A forward-compat policy governs future v1→v2 changes.

**Constraint:** the project is stdlib + cobra only (no YAML library). Artifacts
are rendered and parsed by hand, so the schema must stay simple enough to
render and parse with string operations. Output must be **deterministic** (no
wall-clock timestamps) so it diffs cleanly and tests can compare bytes.

## The v1 schema

### Common envelope (every artifact)

```yaml
schema_version: 1
kind: <kind>            # etl_patterns | graphql_patterns | api_patterns |
                        # data_quality_patterns | common_bugs | successful_prompts |
                        # droid_patterns | runtime_patterns | project_profile
generated_by: bqa dev   # internal/version.Version ("dev" in tests → deterministic;
                        # vX.Y.Z in a release build). Metadata / provenance only.
```

No `generated_at` wall-clock field — determinism is worth more than a timestamp
here, and provenance can be enriched later under #40 without a version bump.

### Pattern artifacts (8 files)

`etl_patterns`, `graphql_patterns`, `api_patterns`, `data_quality_patterns`,
`common_bugs`, `successful_prompts`, `droid_patterns`, `runtime_patterns`:

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

Empty case renders `patterns: []`.

`kind` is file-level; `domain` is finding-level (they differ — e.g.
`common_bugs` entries carry the failing area as `domain`). The list key is
`patterns:` in **all** pattern files (uniform for generic parsing), including
`successful_prompts`.

### project_profile.yaml

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

## Field derivation rules (deterministic, honest)

### `id`

`<domain>-<first 8 hex of sha256(domain + "|" + name + "|" + source)>` using
`crypto/sha256` (stdlib). Content-derived, so it is stable when other findings
are inserted/removed (good for #59 diff/merge) and fully deterministic.

### `confidence`

Computed locally per finding from the number of **distinct domain keywords**
present in its `evidence` (the per-domain keyword lists already exist in the
extractor):

- `high` — ≥ 3 distinct domain signals in the evidence
- `medium` — exactly 2
- `low` — 1 (or a minimal/single-keyword match)

Documented honestly as: *"confidence reflects how many corroborating domain
signals appear in the evidence; it is a heuristic, not a statistical
probability."* Cross-session corroboration (same pattern in ≥2 sessions) is
deliberately deferred as a future **additive** improvement within v1.

### `reusable_check`

A deterministic per-domain (and matched-sub-signal) template, documented as a
*"candidate — review before use."* Mapping:

| domain | reusable_check |
|---|---|
| etl (reconciliation) | `compare source vs target row counts for the window` |
| etl (null/dup) | `assert no unexpected nulls or duplicate keys` |
| graphql | `assert query/mutation response shape and error handling` |
| api | `assert endpoint status code and response contract` |
| data_quality | `assert null / duplicate / schema-drift rules pass` |
| common_bugs | `add a regression check reproducing the failure signal` |
| successful_prompts | *(the reusable prompt text itself, from `reusablePrompt`)* |
| droid / runtime | templated per domain |

### `suggested_next_reviews` (profile)

One line per detected domain (`signals > 0`), ordered by signal count
descending then domain name (for determinism):
`"Review <domain> coverage (<N> signals)."` Empty when no domains detected.

## Consumer changes (hard cutover, one cohesive PR)

| File | Change |
|---|---|
| `internal/core/knowledge/contract.go` | add `Kind` to `ArtifactSpec`; add package const `SchemaVersion = 1` |
| `internal/core/knowledge/usecase.go` | rewrite `renderFindings` / `renderProfile` to emit the envelope + new fields; compute `id` / `confidence` / `reusable_check` at render time (minimal `Finding` struct change) |
| `internal/core/knowledge/validate.go` | validate `schema_version: 1`, `kind: <expected>`, presence of `patterns:` (pattern files) or `profile:` + `sessions_analyzed` (profile), replacing the current root-key / `sessions_analyzed` checks; reject unknown/missing `schema_version` and wrong `kind` with a clear error |
| `internal/app/codex_knowledge.go` | update the hand-rolled parser to read entries under `patterns:` and scalars under `profile:`; optionally surface `confidence` in the summary |
| `docs/knowledge-artifacts.md` | **new** — schema spec, synthetic examples, field definitions, the confidence rule, and the forward-compat policy (file named in issue scope) |
| Tests | update existing assertions to the v1 shape; add "required fields present" tests (AC) and a determinism (byte-identical) test |
| JS mirror & examples | `docs/assets/upload.js` (the game reimplements the same artifact shape client-side), `examples/knowledge/successful_prompts.yaml`, `docs/knowledge-review-checklist.md` → align to v1 so the demo/docs stay honest (keep the game `node --test` suite green) |

## Forward-compatibility policy (documented in `docs/knowledge-artifacts.md`)

- `schema_version` is an integer; **v1** is current.
- **Additive** change (a new optional field) does **not** bump `schema_version`;
  consumers **must ignore unknown fields**.
- **Breaking** change (remove/rename a field, change its type or semantics)
  bumps `schema_version`.
- Consumers **must read `schema_version`** and, on an unsupported major, fail
  with a clear message (`bqa build --check` / `bqa doctor` report it; `bqa codex`
  degrades gracefully).
- Artifacts are **derived** — there is no in-place migrator; on any schema
  change, regenerate with `bqa build`.

## Testing

- **Required-fields tests (AC):** every generated file contains `schema_version:`
  and `kind:`; pattern findings carry `id` / `domain` / `evidence` / `source` /
  `reusable_check` / `confidence`; the profile carries `sessions_analyzed` /
  `domains_detected` / `signals` / `suggested_next_reviews`.
- **Determinism test:** identical synthetic input → byte-identical output.
- **Validation test:** `validate` returns a non-zero result for a file with a
  missing/unknown `schema_version` or a wrong `kind`.
- `go test ./...` and the game `node --test docs/assets/*.test.js` stay green.

## Out of scope (v1)

- Cross-session confidence corroboration (future additive improvement).
- A real YAML library / structured emitter (stays hand-rolled per project
  constraint).
- Export/import backends (#59) and learning/governance (#40) — this schema is
  their foundation, not their implementation.
- An in-place migrator (`bqa knowledge migrate`) — unnecessary for derived
  artifacts.
