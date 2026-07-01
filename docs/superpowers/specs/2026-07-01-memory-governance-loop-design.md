# Design: Memory governance loop — learn → review → promote/reject (issue #40, slice 2)

## Context

The first #40 slice added static critical-thinking guardrails. This second slice
adds the **governance loop**: BQA extracts *candidate* memory from sessions, a
human reviews it, and only approved items enter stable memory. This delivers the
core #40 guarantee: **nothing is promoted into reusable memory without human
approval.**

Builds on: the v1 knowledge schema (#16 — envelope + content-hash ids), the
normalized-session reader used by `bqa build`, and `.bqa/memory/` as the private
memory area (portable via #59). No auto-promotion; extraction is the existing
keyword heuristic (no new intelligence in this slice).

Decisions from brainstorming:

1. **Scope:** the full loop with a minimal candidate set — `lessons_learned` and
   `skill_candidates`. (Agent/workflow candidates, fact-check queue, source
   registry, hygiene/diff are later slices.)
2. **Storage:** everything under `.bqa/memory/` as v1-schema YAML files; promote/
   reject transition an item **by id** into `approved_patterns.yaml` /
   `rejected_patterns.yaml` and append to `decision_log.yaml`. Reviewable,
   diffable, portable.

**Constraint:** stdlib + cobra only; deterministic output (no wall-clock;
`decision_log` records id/action/name, not a timestamp — timestamps are a future
enhancement). No raw session bodies copied into candidates (bounded evidence,
reusing the knowledge extractor's guardrail).

## Command surface

A new `bqa memory` command group (keeps all governance verbs together; `bqa brain`
stays for git/archive from #59):

```
bqa memory learn   [--sessions .bqa/input/sessions] [--memory-dir .bqa/memory]
bqa memory review  [--memory-dir .bqa/memory]
bqa memory promote <id> [--memory-dir .bqa/memory]
bqa memory reject  <id> [--memory-dir .bqa/memory]
```

## Architecture (hexagonal)

```
internal/app/memory.go            (CLI wiring: bqa memory learn/review/promote/reject)
        ↓
internal/core/memgov/             (use case — ports + stdlib only)
        ↓
internal/ports/memgov.go          (MemorySessionReader, GovernanceStore)
        ↑
internal/adapters/fs/*            (fs governance store; reuse fs session reader)
```

Pure logic (candidate extraction from session text, id via content hash, YAML
render/parse of the memory files) lives in `core/memgov`. Only session reading
and the governance-file store are ports.

### Ports (`internal/ports/memgov.go`)

```go
type GovernanceStore interface {
    // Load returns the current governance state (candidates, approved, rejected, log).
    Load(ctx context.Context, memoryDir string) (GovernanceState, error)
    // Save writes the state back atomically (deterministic YAML).
    Save(ctx context.Context, memoryDir string, state GovernanceState) error
}
```

`MemorySessionReader` is the existing normalized-session reader interface (reused,
not new). `GovernanceState` is a plain struct holding the four lists.

### Data model

Each memory item carries the v1 envelope fields plus governance status:

```yaml
# .bqa/memory/skill_candidates.yaml
schema_version: 1
kind: skill_candidates
generated_by: bqa dev
items:
  - id: skill-0a1b2c3d          # content hash of kind|name|source
    name: "etl reconciliation check"
    domain: etl
    evidence: "...bounded snippet..."
    source: "normalized/etl/s1.md"
    status: pending
```

`lessons_learned.yaml` (kind: lessons_learned), `approved_patterns.yaml`
(kind: approved_patterns), `rejected_patterns.yaml` (kind: rejected_patterns)
follow the same envelope. `decision_log.yaml` (kind: decision_log) holds entries
`{ id, action: promoted|rejected, name }`.

### Use cases (`internal/core/memgov/`)

- **Learn** (`learn.go`): read normalized sessions → extract lessons + skill
  candidates (keyword heuristic; bounded evidence) → merge into the candidate
  files with `status: pending`. **Idempotent** — an item whose id already exists
  (in any of candidates/approved/rejected) is not re-added.
- **Review** (`review.go`): return the pending candidates (id, kind, name,
  evidence) for display.
- **Promote / Reject** (`decide.go`): find the pending candidate by id; move it to
  `approved_patterns` (promote) or `rejected_patterns` (reject) with the new
  status; remove it from candidates; append a `decision_log` entry. Unknown or
  already-decided id → a clear error, nothing changed.

### CLI (`internal/app/memory.go`)

Builds the fs adapters and calls `memgov.UseCase`. `learn` prints counts;
`review` lists pending items; `promote`/`reject` print the moved item + decision.

## Testing

- **Learn:** synthetic sessions → `lessons_learned.yaml` / `skill_candidates.yaml`
  contain pending items with stable ids and bounded evidence (no raw body); a
  second `learn` run adds nothing new (idempotent).
- **Review:** lists exactly the pending candidates.
- **Promote:** a pending id moves to `approved_patterns.yaml` (status approved),
  is removed from candidates, and a `decision_log` entry is appended; a second
  promote of the same id errors; nothing is auto-promoted without this call.
- **Reject:** symmetric to promote into `rejected_patterns.yaml`.
- **Errors:** unknown id → clear error, files unchanged.
- `go test ./...` passes; output is deterministic.

## Out of scope (later #40 slices)

- Agent/workflow candidates, `fact_check_queue`, `source_registry`, memory
  hygiene/audit/diff commands.
- Promotion that materializes a candidate into a real `.bqa/skills/*.md` (this
  slice keeps approved items in `approved_patterns.yaml`).
- Wall-clock timestamps in `decision_log` (kept deterministic for now).
- Smarter (non-keyword) candidate extraction.
