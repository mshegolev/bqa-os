# Single internal catalog for built-in agents, skills, and workflows (bqa-os)

Date: 2026-06-30
Status: approved design, pending implementation
Scope: bqa-os only (Go engine). No coupling to the bqa-team registry.

## Context

A project audit found that bqa-os defines its built-in agent/skill/workflow
content as hardcoded Go string functions scattered across multiple packages,
with no single source:

- `internal/core/artifacts/usecase.go` hardcodes `etl-qa-agent`, `runtime-agent`,
  two skills, two workflows, and the `registry/*.yaml` index.
- `internal/core/etlpack/usecase.go` hardcodes `codex-etl-qa-agent`,
  `claude-code-etl-qa-agent`, three workflows, plus specs/prompts/examples.
- The "ETL QA agent" persona exists under three ids
  (`etl-qa-agent`, `codex-etl-qa-agent`, `claude-code-etl-qa-agent`) with
  divergent bodies — the same intent re-authored three times.
- `core/artifacts` emits a hardcoded `registry/agents.yaml` listing agents,
  maintained separately from the files it writes.

This refactor consolidates the scattered built-in content into one internal
catalog package and de-duplicates the ETL-QA persona, **without coupling to the
bqa-team `registry.json`** (that "single source across repos" option was
explicitly declined). External behavior of `bqa build` and `bqa etl-agent-pack`
is unchanged: same files, same counts.

## Goals

- One package owns all built-in agents, skills, and workflows as structured data.
- The ETL-QA persona is defined once (base) with per-runtime deltas (Model A).
- `core/artifacts` derives `registry/*.yaml` from the catalog entries it emits,
  so the index can never drift from the emitted files.
- File counts and external behavior of both generators are preserved.

## Non-goals

- No coupling to bqa-team `registry.json` (declined "full" option).
- `internal/runtime` (master context, `/bqa-master` command) is **out of scope**
  — a separate concern handled by the runtime-consolidation thread.
- Generator-specific single-use documents (sales package, ETL specs/prompts/
  examples, statistics summary, README) stay in their generators; they are not
  reusable catalog material and contain no duplication.

## Design

### 1. New package `internal/catalog`

Holds the reusable built-in content as data, not scattered string functions.

```go
package catalog

// Entry is a self-contained built-in skill or workflow.
type Entry struct {
    ID      string
    Type    string // "skill" | "workflow"
    Domain  string // "etl" | "runtime" | "memory" | ...
    Title   string
    Content string // full markdown body
}

// AgentCore is the shared definition of an agent persona.
type AgentCore struct {
    ID      string
    Title   string
    Domain  string
    Mission string
    Rules   []string
}

// RuntimeFlavor adds runtime-specific framing on top of an AgentCore.
type RuntimeFlavor struct {
    TitlePrefix string   // e.g. "Codex", "Claude Code"
    Intro       string   // runtime-specific opening line
    ExtraRules  []string // appended after the base rules
}
```

Catalog contents:
- Skills: `etl-log-investigation`, `runtime-trace-review`.
- Workflows: `etl-verification`, `session-knowledge`, `etl-regression`,
  `data-reconciliation`, `data-quality-validation`.
- Agents: `etlQACore` (the unified ETL-QA persona), plus `runtimeAgent`
  (distinct role — session/runtime trace review — kept as its own entry).

### 2. ETL-QA persona dedup (Model A: base + runtime delta)

The three divergent ETL-QA personas collapse to one `etlQACore` (mission + base
rules) rendered through optional runtime flavors:

```go
func RenderAgent(core AgentCore, flavor *RuntimeFlavor) string
```

- `agents/etl-qa-agent.md` (build) = `RenderAgent(etlQACore, nil)` — generic.
- `agents/codex-etl-qa-agent.md` (etlpack) = `RenderAgent(etlQACore, codexFlavor)`.
- `agents/claude-code-etl-qa-agent.md` (etlpack) =
  `RenderAgent(etlQACore, claudeFlavor)`.

Base rules and mission live in one place; each runtime variant adds only its
intro line and 1–2 extra rules. The variant titles (`# Codex ETL QA Agent`,
`# Claude Code ETL QA Agent`) are preserved.

### 3. Generators render from the catalog

- `core/artifacts` selects its agents/skills/workflows from `catalog` instead of
  local string functions, and **derives** `registry/{index,agents,skills,
  workflows}.yaml` from the catalog entries it emits (id, path, domain). The
  derived YAML must equal the current YAML byte-for-byte (golden check).
- `core/etlpack` renders its agents (via `RenderAgent`) and workflows from
  `catalog`; specs, prompts, examples, statistics, and README stay local.
- The local string functions that move into the catalog are deleted from the
  generators.

### 4. Behavior and output

- File counts unchanged: `bqa build` = 10 (or 18 with `--sales-package`);
  `bqa etl-agent-pack` = 12.
- Output paths unchanged.
- The ETL-QA persona **bodies change** (now base + runtime delta). Titles and
  registry YAML are preserved.

## Affected files

Added:
- `internal/catalog/catalog.go` (Entry/AgentCore/RuntimeFlavor + content + RenderAgent)
- `internal/catalog/catalog_test.go`

Modified:
- `internal/core/artifacts/usecase.go` (render from catalog; derive registry YAML;
  drop the relocated string funcs)
- `internal/core/artifacts/usecase_test.go` (update ETL-QA/agent body assertions
  if their phrases change; counts unchanged)
- `internal/core/etlpack/usecase.go` (render agents/workflows from catalog; drop
  relocated string funcs)
- `internal/core/etlpack/usecase_test.go` (update persona body assertions if
  needed; counts unchanged)

Unchanged: `internal/runtime/*`, sales/specs/prompts/examples generators.

## Testing & verification

- `internal/catalog` unit tests: `RenderAgent` with and without a flavor; base
  rules present in all variants; runtime extras appended; titles correct.
- A golden test asserting the derived `registry/*.yaml` equals the current YAML.
- `core/artifacts` and `core/etlpack` tests keep their count assertions (10/18/12)
  and title-phrase assertions; body-phrase assertions updated to the new unified
  rule text where they changed.
- `go build ./...`, `go vet ./...`, `go test ./...` green.
- Manual: `bqa build` and `bqa etl-agent-pack` produce the same file set; spot
  check the three ETL-QA agent files render base + correct runtime delta.

## Risks

- **Output drift in derived registry YAML** — mitigated by the byte-for-byte
  golden check against the current YAML.
- **Persona body change** is intentional and the main behavioral delta; bounded
  to three agent files, titles preserved, deltas explicit in the catalog.
- **Scope creep into runtime** — explicitly excluded; runtime stays untouched.
