# Single Internal Catalog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate bqa-os's scattered hardcoded agents/skills/workflows into one `internal/catalog` package, de-duplicate the ETL-QA persona (base + runtime delta), and derive `registry/*.yaml` from the catalog — with unchanged file counts.

**Architecture:** A new leaf package `internal/catalog` holds reusable content as data (`Entry`, `AgentCore`, `RuntimeFlavor`) plus `RenderAgent`. `core/artifacts` and `core/etlpack` render from the catalog instead of local string functions; `core/artifacts` generates `registry/*.yaml` from the catalog entries it emits. `internal/runtime` and generator-specific docs are untouched.

**Tech Stack:** Go 1.22, standard library only.

Spec: `docs/superpowers/specs/2026-06-30-catalog-single-source-of-truth-design.md`

---

## File Structure

- `internal/catalog/catalog.go` — types (`Entry`, `AgentCore`, `RuntimeFlavor`), content (skills, workflows, agents), `RenderAgent`, registry-derivation helpers.
- `internal/catalog/catalog_test.go` — unit tests for `RenderAgent` and registry derivation.
- `internal/core/artifacts/usecase.go` — render from catalog; derive registry YAML; drop relocated funcs.
- `internal/core/etlpack/usecase.go` — render agents/workflows from catalog; drop relocated funcs.
- `internal/core/artifacts/usecase_test.go`, `internal/core/etlpack/usecase_test.go` — update body-phrase assertions; keep counts.

---

## Task 1: Catalog types and `RenderAgent`

**Files:**
- Create: `internal/catalog/catalog.go`
- Test: `internal/catalog/catalog_test.go`

- [ ] **Step 1: Write the failing test**

```go
package catalog

import (
	"strings"
	"testing"
)

func TestRenderAgentGeneric(t *testing.T) {
	out := RenderAgent(etlQACore, nil)
	if !strings.HasPrefix(out, "# ETL QA Agent\n") {
		t.Fatalf("generic title missing:\n%s", out)
	}
	if !strings.Contains(out, "## Operating Rules") {
		t.Fatal("operating rules section missing")
	}
	for _, rule := range etlQACore.Rules {
		if !strings.Contains(out, "- "+rule+"\n") {
			t.Fatalf("base rule missing: %s", rule)
		}
	}
}

func TestRenderAgentRuntimeFlavor(t *testing.T) {
	out := RenderAgent(etlQACore, &codexFlavor)
	if !strings.HasPrefix(out, "# Codex ETL QA Agent\n") {
		t.Fatalf("codex title missing:\n%s", out)
	}
	if !strings.Contains(out, codexFlavor.Intro) {
		t.Fatal("codex intro missing")
	}
	for _, rule := range codexFlavor.ExtraRules {
		if !strings.Contains(out, "- "+rule+"\n") {
			t.Fatalf("codex extra rule missing: %s", rule)
		}
	}
	// base rules still present in every variant
	if !strings.Contains(out, "- "+etlQACore.Rules[0]+"\n") {
		t.Fatal("base rule missing from flavored output")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/catalog/ -run TestRenderAgent`
Expected: FAIL — package/identifiers undefined.

- [ ] **Step 3: Write minimal implementation**

Create `internal/catalog/catalog.go`:

```go
// Package catalog is the single source of bqa-os's built-in agents, skills,
// and workflows.
package catalog

import "strings"

// Entry is a self-contained built-in skill or workflow.
type Entry struct {
	ID      string
	Type    string // "skill" | "workflow"
	Domain  string
	Title   string
	Content string
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
	TitlePrefix string
	Intro       string
	ExtraRules  []string
}

// RenderAgent renders an agent persona, optionally with a runtime flavor.
func RenderAgent(core AgentCore, flavor *RuntimeFlavor) string {
	title := core.Title
	role := core.Mission
	rules := append([]string(nil), core.Rules...)
	if flavor != nil {
		if flavor.TitlePrefix != "" {
			title = flavor.TitlePrefix + " " + core.Title
		}
		if flavor.Intro != "" {
			role = flavor.Intro + " " + core.Mission
		}
		rules = append(rules, flavor.ExtraRules...)
	}
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString("## Role\n\n" + role + "\n\n")
	b.WriteString("## Operating Rules\n\n")
	for _, rule := range rules {
		b.WriteString("- " + rule + "\n")
	}
	return b.String()
}

// --- ETL-QA persona (unified) ---

var etlQACore = AgentCore{
	ID:      "etl-qa-agent",
	Title:   "ETL QA Agent",
	Domain:  "etl",
	Mission: "Validate Big Data / ETL changes (Airflow, Spark, Hive, reconciliation, data quality) with reproducible evidence.",
	Rules: []string{
		"Build concise investigation plans before changing anything.",
		"Prefer commands, logs, row counts, and schema evidence over guesses.",
		"Keep private data, raw logs, secrets, and customer records out of output.",
		"Report scope, checks, evidence, result, risks, and next action.",
	},
}

var codexFlavor = RuntimeFlavor{
	TitlePrefix: "Codex",
	Intro:       "You are the ETL QA Agent working in Codex.",
	ExtraRules: []string{
		"Prefer existing repository test commands before inventing new helpers.",
		"When inputs are missing, use the synthetic example in this pack to demonstrate the workflow.",
	},
}

var claudeFlavor = RuntimeFlavor{
	TitlePrefix: "Claude Code",
	Intro:       "You are the ETL QA Agent working in Claude Code.",
	ExtraRules: []string{
		"Follow the project's existing tooling and test framework.",
		"Ask a blocker question when environment, dataset, or acceptance criteria are unclear.",
	},
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/catalog/ -run TestRenderAgent -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/catalog/catalog.go internal/catalog/catalog_test.go
git commit -m "feat(catalog): add agent persona types and RenderAgent"
```

---

## Task 2: Relocate skills, workflows, and runtime-agent into the catalog

**Files:**
- Modify: `internal/catalog/catalog.go`
- Test: `internal/catalog/catalog_test.go`

Move the exact current markdown bodies from the generators into catalog `Entry`
values. Source content (copy verbatim, only the Go string literal):
- `internal/core/artifacts/usecase.go` `etlSkill()` → Entry `etl-log-investigation` (skill, domain `etl`).
- `internal/core/artifacts/usecase.go` `runtimeSkill()` → Entry `runtime-trace-review` (skill, domain `runtime`).
- `internal/core/artifacts/usecase.go` `etlWorkflow()` → Entry `etl-verification-workflow` (workflow, domain `etl`).
- `internal/core/artifacts/usecase.go` `sessionWorkflow()` → Entry `session-knowledge-workflow` (workflow, domain `memory`).
- `internal/core/artifacts/usecase.go` `runtimeAgent()` → keep as a separate `AgentCore`-free raw entry `runtime-agent` (agent, domain `runtime`) rendered verbatim (distinct role; not de-duplicated).
- `internal/core/etlpack/usecase.go` `etlRegressionWorkflow()` → Entry `etl-regression-workflow` (workflow, domain `etl`).
- `internal/core/etlpack/usecase.go` `dataReconciliationWorkflow()` → Entry `data-reconciliation-workflow` (workflow, domain `etl`).
- `internal/core/etlpack/usecase.go` `dataQualityValidationWorkflow()` → Entry `data-quality-validation-workflow` (workflow, domain `etl`).

- [ ] **Step 1: Write the failing test**

```go
func TestCatalogLookups(t *testing.T) {
	if Skill("etl-log-investigation").Content == "" {
		t.Fatal("etl-log-investigation skill missing")
	}
	if Workflow("session-knowledge-workflow").Domain != "memory" {
		t.Fatal("session-knowledge-workflow domain wrong")
	}
	if !strings.Contains(RuntimeAgentContent(), "Runtime Agent") {
		t.Fatal("runtime-agent content missing")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/catalog/ -run TestCatalogLookups`
Expected: FAIL — `Skill`/`Workflow`/`RuntimeAgentContent` undefined.

- [ ] **Step 3: Write minimal implementation**

Add to `internal/catalog/catalog.go` the `Entry` values (verbatim content from the
source funcs above) and accessors:

```go
var skills = []Entry{
	{ID: "etl-log-investigation", Type: "skill", Domain: "etl", Title: "ETL Log Investigation", Content: /* exact body of etlSkill() */ ""},
	{ID: "runtime-trace-review", Type: "skill", Domain: "runtime", Title: "Runtime Trace Review", Content: /* exact body of runtimeSkill() */ ""},
}

var workflows = []Entry{
	{ID: "etl-verification-workflow", Type: "workflow", Domain: "etl", Title: "ETL Verification Workflow", Content: /* etlWorkflow() */ ""},
	{ID: "session-knowledge-workflow", Type: "workflow", Domain: "memory", Title: "Session Knowledge Workflow", Content: /* sessionWorkflow() */ ""},
	{ID: "etl-regression-workflow", Type: "workflow", Domain: "etl", Title: "ETL Regression Workflow", Content: /* etlRegressionWorkflow() */ ""},
	{ID: "data-reconciliation-workflow", Type: "workflow", Domain: "etl", Title: "Data Reconciliation Workflow", Content: /* dataReconciliationWorkflow() */ ""},
	{ID: "data-quality-validation-workflow", Type: "workflow", Domain: "etl", Title: "Data Quality Validation Workflow", Content: /* dataQualityValidationWorkflow() */ ""},
}

var runtimeAgentContent = /* exact body of runtimeAgent() */ ""

func Skill(id string) Entry    { return find(skills, id) }
func Workflow(id string) Entry { return find(workflows, id) }
func RuntimeAgentContent() string { return runtimeAgentContent }

func find(entries []Entry, id string) Entry {
	for _, e := range entries {
		if e.ID == id {
			return e
		}
	}
	return Entry{}
}
```

Replace each `/* exact body of X() */ ""` with the verbatim string literal copied
from the corresponding source function. Do not alter the text.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/catalog/ -run TestCatalogLookups -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/catalog/
git commit -m "feat(catalog): relocate skills, workflows, runtime-agent content"
```

---

## Task 3: Derive `registry/*.yaml` from the catalog

**Files:**
- Modify: `internal/catalog/catalog.go`
- Test: `internal/catalog/catalog_test.go`

The current YAML (must be reproduced byte-for-byte) is:

```
# registry/index.yaml
registry:
  version: 1
  skills: registry/skills.yaml
  agents: registry/agents.yaml
  workflows: registry/workflows.yaml
  knowledge: knowledge/project_profile.yaml

# registry/skills.yaml
skills:
  - id: etl-log-investigation
    path: skills/etl-log-investigation.md
    domain: etl
  - id: runtime-trace-review
    path: skills/runtime-trace-review.md
    domain: runtime

# registry/agents.yaml
agents:
  - id: etl-qa-agent
    path: agents/etl-qa-agent.md
    domain: etl
  - id: runtime-agent
    path: agents/runtime-agent.md
    domain: runtime

# registry/workflows.yaml
workflows:
  - id: etl-verification-workflow
    path: workflows/etl-verification-workflow.md
    domain: etl
  - id: session-knowledge-workflow
    path: workflows/session-knowledge-workflow.md
    domain: memory
```

- [ ] **Step 1: Write the failing test (golden)**

```go
func TestRegistryYAMLGolden(t *testing.T) {
	wantAgents := "agents:\n  - id: etl-qa-agent\n    path: agents/etl-qa-agent.md\n    domain: etl\n  - id: runtime-agent\n    path: agents/runtime-agent.md\n    domain: runtime\n"
	if got := RegistryAgentsYAML(); got != wantAgents {
		t.Fatalf("agents yaml drift:\n got=%q\nwant=%q", got, wantAgents)
	}
	wantSkills := "skills:\n  - id: etl-log-investigation\n    path: skills/etl-log-investigation.md\n    domain: etl\n  - id: runtime-trace-review\n    path: skills/runtime-trace-review.md\n    domain: runtime\n"
	if got := RegistrySkillsYAML(); got != wantSkills {
		t.Fatalf("skills yaml drift:\n got=%q\nwant=%q", got, wantSkills)
	}
	wantWorkflows := "workflows:\n  - id: etl-verification-workflow\n    path: workflows/etl-verification-workflow.md\n    domain: etl\n  - id: session-knowledge-workflow\n    path: workflows/session-knowledge-workflow.md\n    domain: memory\n"
	if got := RegistryWorkflowsYAML(); got != wantWorkflows {
		t.Fatalf("workflows yaml drift:\n got=%q\nwant=%q", got, wantWorkflows)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/catalog/ -run TestRegistryYAMLGolden`
Expected: FAIL — derivation functions undefined.

- [ ] **Step 3: Write minimal implementation**

Add to `internal/catalog/catalog.go`. Note: the `agents` list for derivation is
exactly `etl-qa-agent` (domain etl) then `runtime-agent` (domain runtime), in
that order, to match the current YAML.

```go
var registryAgents = []Entry{
	{ID: etlQACore.ID, Type: "agent", Domain: etlQACore.Domain},
	{ID: "runtime-agent", Type: "agent", Domain: "runtime"},
}

func RegistryAgentsYAML() string    { return renderRegistry("agents", "agents", registryAgents) }
func RegistrySkillsYAML() string    { return renderRegistry("skills", "skills", skills) }
func RegistryWorkflowsYAML() string { return renderRegistry("workflows", "workflows", workflows) }

func renderRegistry(key, dir string, entries []Entry) string {
	var b strings.Builder
	b.WriteString(key + ":\n")
	for _, e := range entries {
		b.WriteString("  - id: " + e.ID + "\n")
		b.WriteString("    path: " + dir + "/" + e.ID + ".md\n")
		b.WriteString("    domain: " + e.Domain + "\n")
	}
	return b.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/catalog/ -run TestRegistryYAMLGolden -v`
Expected: PASS — byte-identical to the current YAML.

- [ ] **Step 5: Commit**

```bash
git add internal/catalog/
git commit -m "feat(catalog): derive registry yaml from catalog entries"
```

---

## Task 4: Render `core/artifacts` from the catalog

**Files:**
- Modify: `internal/core/artifacts/usecase.go`
- Modify (if phrases change): `internal/core/artifacts/usecase_test.go`

- [ ] **Step 1: Update the use case to read from the catalog**

In `internal/core/artifacts/usecase.go`, replace the hardcoded entries in the
`artifacts` map:

```go
import "github.com/mshegolev/bqa-os/internal/catalog"

// agents
"agents/etl-qa-agent.md": catalog.RenderAgent(catalog.ETLQA(), nil),
"agents/runtime-agent.md": catalog.RuntimeAgentContent(),
// skills
"skills/etl-log-investigation.md": catalog.Skill("etl-log-investigation").Content,
"skills/runtime-trace-review.md": catalog.Skill("runtime-trace-review").Content,
// workflows
"workflows/etl-verification-workflow.md": catalog.Workflow("etl-verification-workflow").Content,
"workflows/session-knowledge-workflow.md": catalog.Workflow("session-knowledge-workflow").Content,
// registry (derived)
"registry/index.yaml": registryIndex(),               // index stays local (unchanged literal)
"registry/skills.yaml": catalog.RegistrySkillsYAML(),
"registry/agents.yaml": catalog.RegistryAgentsYAML(),
"registry/workflows.yaml": catalog.RegistryWorkflowsYAML(),
```

Add an exported accessor `ETLQA()` to the catalog returning `etlQACore` (since
`etlQACore` is unexported):

```go
// in catalog.go
func ETLQA() AgentCore { return etlQACore }
```

Delete the now-unused local funcs in `artifacts/usecase.go`: `etlSkill`,
`runtimeSkill`, `etlAgent`, `runtimeAgent`, `etlWorkflow`, `sessionWorkflow`,
`registrySkills`, `registryAgents`, `registryWorkflows`. Keep `registryIndex`,
`registryIndexWithSales`, `registrySales`, and all `sales*` funcs.

- [ ] **Step 2: Update tests for the new ETL-QA body**

The count assertion (`ArtifactsCreated != 10` / `!= 18`) stays. Update only
assertions that check changed body text. Current `usecase_test.go` checks:
`skills/etl-log-investigation.md` contains "ETL Log Investigation" (unchanged),
`agents/runtime-agent.md` contains "Runtime Agent" (unchanged),
`workflows/session-knowledge-workflow.md` contains "bqa build" (unchanged),
`registry/index.yaml` contains "registry:" (unchanged). No changes needed unless
an assertion references etl-qa-agent body text — add/keep:

```go
if !strings.Contains(writer.files["agents/etl-qa-agent.md"], "ETL QA Agent") {
	t.Fatal("etl-qa-agent persona missing")
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/core/artifacts/ -v`
Expected: PASS, count 10 and 18 unchanged.

- [ ] **Step 4: Commit**

```bash
git add internal/core/artifacts/ internal/catalog/
git commit -m "refactor(artifacts): render agents/skills/workflows and registry from catalog"
```

---

## Task 5: Render `core/etlpack` from the catalog

**Files:**
- Modify: `internal/core/etlpack/usecase.go`
- Modify (if phrases change): `internal/core/etlpack/usecase_test.go`

- [ ] **Step 1: Update the use case**

In `internal/core/etlpack/usecase.go`, replace the agent/workflow map entries:

```go
import "github.com/mshegolev/bqa-os/internal/catalog"

"agents/codex-etl-qa-agent.md": catalog.RenderAgent(catalog.ETLQA(), catalog.CodexFlavor()),
"agents/claude-code-etl-qa-agent.md": catalog.RenderAgent(catalog.ETLQA(), catalog.ClaudeFlavor()),
"workflows/etl-regression-workflow.md": catalog.Workflow("etl-regression-workflow").Content,
"workflows/data-reconciliation-workflow.md": catalog.Workflow("data-reconciliation-workflow").Content,
"workflows/data-quality-validation-workflow.md": catalog.Workflow("data-quality-validation-workflow").Content,
```

Add exported flavor accessors to the catalog:

```go
// in catalog.go
func CodexFlavor() *RuntimeFlavor  { return &codexFlavor }
func ClaudeFlavor() *RuntimeFlavor { return &claudeFlavor }
```

Delete the now-unused local funcs in `etlpack/usecase.go`: `codexAgent`,
`claudeCodeAgent`, `etlRegressionWorkflow`, `dataReconciliationWorkflow`,
`dataQualityValidationWorkflow`. Keep specs, prompts, examples, statistics,
README funcs.

- [ ] **Step 2: Update tests for the new persona titles**

Count assertion (`ArtifactsCreated != 12`) stays. The test checks
`agents/codex-etl-qa-agent.md` contains "Codex ETL QA Agent" and
`agents/claude-code-etl-qa-agent.md` contains "Claude Code ETL QA Agent" — both
preserved by `RenderAgent` (title prefix). No change unless a rule-text phrase is
asserted (it is not). Leave as-is.

- [ ] **Step 3: Run tests**

Run: `go test ./internal/core/etlpack/ -v`
Expected: PASS, count 12 unchanged.

- [ ] **Step 4: Commit**

```bash
git add internal/core/etlpack/ internal/catalog/
git commit -m "refactor(etlpack): render ETL-QA agents and workflows from catalog"
```

---

## Task 6: Full verification

**Files:** none (verification only)

- [ ] **Step 1: Build, vet, format, test**

Run:
```bash
gofmt -l internal/catalog internal/core/artifacts internal/core/etlpack
go vet ./...
go build ./...
go test ./...
```
Expected: gofmt clean, vet clean, build OK, all tests green.

- [ ] **Step 2: Manual output check**

Run:
```bash
go build -o /tmp/bqa ./cmd/bqa
TMP=$(mktemp -d); cd "$TMP"
/tmp/bqa build || true
find .bqa -type f | sort
grep -l "ETL QA Agent" .bqa/agents/*.md
```
Expected: `bqa build` writes the same `.bqa/{agents,skills,workflows,registry}` set; `agents/etl-qa-agent.md` renders the unified persona.

- [ ] **Step 3: Final commit (if any formatting fixups)**

```bash
git add -A
git commit -m "chore: gofmt catalog refactor" || true
```
