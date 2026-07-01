# Design: Critical-thinking guardrails for BQA agents (issue #40, MVP slice)

## Context

Issue #40 asks for a continuous-learning and memory-governance layer. That is a
large surface (candidate → review → promote lifecycle, hygiene, source
attribution, guardrails). This spec covers the **first MVP vertical slice**
chosen during brainstorming: the **critical-thinking + memory-safety
guardrails** that generated/installed BQA agents must follow. The
learn/review/promote governance loop is a separate later slice.

The guardrails are the 8 critical-thinking rules and the memory-safety rules
from issue #40 (do not invent facts for critical decisions; prefer local
project memory and generated BQA artifacts; require source checking for external
facts; ask a human when evidence is insufficient; mark assumptions; separate
fact / inference / recommendation; keep human-in-the-loop for high-impact QA,
product, release, security, compliance, or customer decisions; never promote
private/raw data into reusable artifacts).

Existing seams this builds on:
- `bqa build` writes `.bqa/` artifacts via `internal/core/artifacts` (agents,
  skills, workflows, registry) through a writer port.
- `bqa codex` appends sections (usage instructions, privacy note) to
  `.bqa/prompts/bqa-master-context.md` — the runtime context every agent reads
  (`internal/app/codex_knowledge.go`, `runtime_context.go`).

Decisions taken during brainstorming:

1. **MVP slice:** critical-thinking guardrails only (the learn/promote loop is deferred).
2. **Surfacing:** a canonical guardrails artifact written by `bqa build`
   (`.bqa/guardrails/critical-thinking.md`) **and** a guardrails section injected
   into the `bqa codex` master-context — a reusable file plus an active runtime rule set.
3. **Static canonical text:** the rules are fixed, deterministic text (no
   configuration in the MVP); configurability is deferred.

**Constraint:** stdlib + cobra only; deterministic output (tests compare bytes).

## Architecture

Single source of truth for the guardrails text, consumed by both producers:

```
internal/core/guardrails/guardrails.go   -> CriticalThinking() string   (pure, canonical text)
        ↑ consumed by
internal/core/artifacts/usecase.go        -> writes .bqa/guardrails/critical-thinking.md (bqa build)
internal/app/codex_knowledge.go           -> appends a guardrails section to the codex master-context (bqa codex)
```

### `internal/core/guardrails/guardrails.go` (new)

A pure function `CriticalThinking() string` returning the canonical markdown: a
titled section listing the numbered critical-thinking rules and the memory-safety
rules. No dependencies, deterministic. This is the ONE place the text lives, so
the build artifact and the codex section never drift.

### `internal/core/artifacts/usecase.go` (modified)

Add one artifact: `guardrails/critical-thinking.md` = `guardrails.CriticalThinking()`,
written through the existing `WriteBQAArtifact` writer port. The artifact count
(asserted in tests) increases by one; update those assertions. No new port.

### `internal/app/codex_knowledge.go` (modified)

Add a `## Critical thinking & memory guardrails` section (the same
`guardrails.CriticalThinking()` text) to the generated codex master-context. It
is appended in **both** the has-knowledge path (`renderKnowledgeSection`) and the
no-knowledge path (`codexNoKnowledgeSection`) so the guardrails are always
present, before the existing privacy note. Reuses the current append mechanism
(no new writes).

## Data flow

- `bqa build` → `.bqa/guardrails/critical-thinking.md` exists with the canonical text.
- `bqa codex` → `.bqa/prompts/bqa-master-context.md` contains the guardrails
  section (whether or not knowledge artifacts are present).
- Both derive from `guardrails.CriticalThinking()`, so they are byte-identical
  in content and stay in sync.

## Testing

- **Canonical text (`guardrails` package):** `CriticalThinking()` contains the
  key rule markers (e.g. "Do not invent", "human", "assumptions", "facts, inference,
  and recommendations", "private") and is stable across calls (deterministic).
- **Build (`artifacts`):** a `bqa build` writes `guardrails/critical-thinking.md`
  with the canonical text; the artifact-count assertions are updated for the +1 file.
- **Codex (`app`):** the generated master-context contains the guardrails section
  in both the with-knowledge and no-knowledge paths; the existing privacy-note and
  "BQA context generated" assertions still hold.
- `go test ./...` passes.

## Out of scope (later #40 slices)

- The `bqa learn` / `bqa review memory` / `bqa promote` / `bqa reject` governance loop.
- Candidate artifacts (lessons_learned / *_candidates / decision_log / fact_check_queue).
- Incremental knowledge updates, source attribution registry, memory hygiene/diff.
- Configurable / per-project guardrail overrides (MVP text is static).
