# Knowledge Artifact Schema v1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give `bqa build` a stable, versioned v1 YAML schema for `.bqa/knowledge/*` artifacts (envelope + `id`/`confidence`/`reusable_check` + richer profile), update all consumers, and document it.

**Architecture:** Artifacts are still rendered/parsed by hand (stdlib + cobra only, no YAML library). Each file gets a flat envelope (`schema_version`, `kind`, `generated_by`) plus a uniform `patterns:` list (or a `profile:` block). Output stays deterministic (no wall-clock timestamps) so it diffs cleanly and tests compare bytes. Hard cutover: `bqa build` always writes v1; derived artifacts are regenerated, no migrator.

**Tech Stack:** Go 1.22, stdlib only (`crypto/sha256`, `encoding/hex`, `strings`, `fmt`, `sort`), cobra. Vanilla JS + `node --test` for the game mirror.

**Reference spec:** `docs/superpowers/specs/2026-07-01-knowledge-artifact-schema-v1-design.md`

---

### Task 1: Schema version constant

**Files:**
- Modify: `internal/core/knowledge/contract.go`
- Test: `internal/core/knowledge/contract_test.go` (create)

Note: `kind` reuses the existing `ArtifactSpec.RootKey` (filename minus `.yaml`) — no new struct field is needed. This task only adds the version constant.

- [ ] **Step 1: Write the failing test**

Create `internal/core/knowledge/contract_test.go`:

```go
package knowledge

import "testing"

func TestSchemaVersionIsOne(t *testing.T) {
	if SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", SchemaVersion)
	}
}

func TestExpectedArtifactsRootKeysServeAsKind(t *testing.T) {
	for _, spec := range ExpectedArtifacts() {
		if spec.RootKey == "" {
			t.Fatalf("artifact %q has empty RootKey (used as kind)", spec.Filename)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/knowledge/ -run TestSchemaVersion -v`
Expected: FAIL — `undefined: SchemaVersion`.

- [ ] **Step 3: Add the constant**

In `internal/core/knowledge/contract.go`, add below the imports:

```go
// SchemaVersion is the current knowledge artifact schema version. Every artifact
// written by bqa build carries it as `schema_version`. Additive field changes
// keep this version; removing/renaming a field or changing its meaning bumps it.
const SchemaVersion = 1
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/knowledge/ -run 'TestSchemaVersion|TestExpectedArtifactsRootKeys' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/knowledge/contract.go internal/core/knowledge/contract_test.go
git commit -m "feat(knowledge): add SchemaVersion constant for v1 artifact schema (#16)"
```

---

### Task 2: Field derivation helpers (id, confidence, reusable_check, signals)

**Files:**
- Create: `internal/core/knowledge/schema.go`
- Test: `internal/core/knowledge/schema_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/core/knowledge/schema_test.go`:

```go
package knowledge

import (
	"strings"
	"testing"
)

func TestFindingIDIsStableAndDomainPrefixed(t *testing.T) {
	f := Finding{Name: "etl_validation", Domain: "etl", SourcePath: "normalized/etl/s1.md"}
	id1 := findingID(f)
	id2 := findingID(f)
	if id1 != id2 {
		t.Fatalf("findingID not deterministic: %q vs %q", id1, id2)
	}
	if !strings.HasPrefix(id1, "etl-") || len(id1) != len("etl-")+8 {
		t.Fatalf("unexpected id shape: %q", id1)
	}
	// Different source => different id.
	if findingID(Finding{Name: "etl_validation", Domain: "etl", SourcePath: "other.md"}) == id1 {
		t.Fatalf("id should change with source")
	}
}

func TestFindingConfidenceBandsBySignalCount(t *testing.T) {
	high := Finding{Domain: "etl", Evidence: "reconciliation row count between source table and target table"}
	if got := findingConfidence(high); got != "high" {
		t.Fatalf("expected high, got %q", got)
	}
	med := Finding{Domain: "api", Evidence: "the endpoint returned a 500 status code"}
	if got := findingConfidence(med); got != "medium" {
		t.Fatalf("expected medium, got %q", got)
	}
	low := Finding{Domain: "graphql", Evidence: "a graphql thing happened"}
	if got := findingConfidence(low); got != "low" {
		t.Fatalf("expected low, got %q", got)
	}
}

func TestReusableCheckIsDomainSpecific(t *testing.T) {
	if got := reusableCheck(Finding{Domain: "etl", Evidence: "row count reconciliation"}); !strings.Contains(got, "row counts") {
		t.Fatalf("etl reusable_check unexpected: %q", got)
	}
	if got := reusableCheck(Finding{Domain: "etl", Evidence: "null check and duplicate keys"}); !strings.Contains(got, "nulls") {
		t.Fatalf("etl null/dup reusable_check unexpected: %q", got)
	}
	prompt := Finding{Domain: "prompts", Evidence: "Task: verify X"}
	if got := reusableCheck(prompt); got != "Task: verify X" {
		t.Fatalf("prompts reusable_check should echo the prompt, got %q", got)
	}
}

func TestDetectedSignalsSortedByCountThenName(t *testing.T) {
	p := ProjectProfile{ETLSignals: 8, GraphQLSignals: 3, APISignals: 3}
	sigs := detectedSignals(p)
	if len(sigs) != 3 {
		t.Fatalf("expected 3 detected domains, got %d", len(sigs))
	}
	if sigs[0].name != "etl" || sigs[1].name != "api" || sigs[2].name != "graphql" {
		t.Fatalf("unexpected order: %+v", sigs)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/knowledge/ -run 'TestFindingID|TestFindingConfidence|TestReusableCheck|TestDetectedSignals' -v`
Expected: FAIL — `undefined: findingID` (etc.).

- [ ] **Step 3: Write the helpers**

Create `internal/core/knowledge/schema.go`:

```go
package knowledge

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
	"github.com/mshegolev/bqa-os/internal/version"
)

// generatedBy is the provenance stamp written as `generated_by`. It is "bqa dev"
// in dev/test builds (deterministic) and "bqa vX.Y.Z" in a release build.
func generatedBy() string { return "bqa " + version.Version }

// findingID returns a stable, content-derived id "<domain>-<8 hex>". It is stable
// when other findings are inserted or removed, so artifacts diff and merge cleanly.
func findingID(f Finding) string {
	sum := sha256.Sum256([]byte(f.Domain + "|" + f.Name + "|" + f.SourcePath))
	return f.Domain + "-" + hex.EncodeToString(sum[:])[:8]
}

// domainKeywords are the per-domain signal words used to gauge how many
// corroborating signals appear in a finding's evidence.
var domainKeywords = map[string][]string{
	"etl":          {"airflow", "spark", "hive", "oozie", "dag", "reconciliation", "source table", "target table", "row count", "parquet", "pipeline", "partition", "schedule"},
	"graphql":      {"graphql", "query", "mutation", "schema", "resolver", "variables", "pagination", "auth", "error"},
	"api":          {"rest api", "http status", "status code", "endpoint", "contract", "openapi", "swagger", "request", "response"},
	"data_quality": {"data quality", "schema drift", "null", "duplicate", "row count", "checksum", "unique", "validation"},
	"bugs":         {"failed", "failure", "error", "panic", "regression", "flaky", "stack trace", "exception", "traceback"},
	"runtime":      {"runtime", "execution", "exit code", "stdout", "stderr", "command"},
	"droid":        {"droid", "factory", "automation", "agent"},
	"prompts":      {"task", "goal", "acceptance", "implement", "verify", "context"},
}

// findingConfidence returns low/medium/high by counting distinct domain keywords
// present in the finding's evidence. It is a heuristic, not a probability.
func findingConfidence(f Finding) string {
	lower := strings.ToLower(f.Evidence)
	n := 0
	for _, kw := range domainKeywords[f.Domain] {
		if strings.Contains(lower, kw) {
			n++
		}
	}
	switch {
	case n >= 3:
		return "high"
	case n == 2:
		return "medium"
	default:
		return "low"
	}
}

// reusableCheck returns a per-domain check candidate — a suggestion for a human
// to review, not an extracted command.
func reusableCheck(f Finding) string {
	lower := strings.ToLower(f.Evidence)
	switch f.Domain {
	case "etl":
		if textutil.HasAny(lower, "null", "duplicate") {
			return "assert no unexpected nulls or duplicate keys"
		}
		return "compare source vs target row counts for the window"
	case "graphql":
		return "assert query/mutation response shape and error handling"
	case "api":
		return "assert endpoint status code and response contract"
	case "data_quality":
		return "assert null / duplicate / schema-drift rules pass"
	case "bugs":
		return "add a regression check reproducing the failure signal"
	case "prompts":
		return f.Evidence // the reusable prompt text itself
	case "runtime":
		return "assert the runtime command exits cleanly and emits expected output"
	case "droid":
		return "capture the automation step as a repeatable check"
	default:
		return "add a check that reproduces this signal"
	}
}

// domainSignal pairs a profile domain with its signal count.
type domainSignal struct {
	name  string
	count int
}

// detectedSignals returns the domains with signals > 0, ordered by count
// descending then name ascending (deterministic).
func detectedSignals(p ProjectProfile) []domainSignal {
	all := []domainSignal{
		{"etl", p.ETLSignals}, {"graphql", p.GraphQLSignals}, {"api", p.APISignals},
		{"data_quality", p.DQSignals}, {"droid", p.DroidSignals}, {"runtime", p.RuntimeSignals},
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].count != all[j].count {
			return all[i].count > all[j].count
		}
		return all[i].name < all[j].name
	})
	out := make([]domainSignal, 0, len(all))
	for _, d := range all {
		if d.count > 0 {
			out = append(out, d)
		}
	}
	return out
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/knowledge/ -run 'TestFindingID|TestFindingConfidence|TestReusableCheck|TestDetectedSignals' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/knowledge/schema.go internal/core/knowledge/schema_test.go
git commit -m "feat(knowledge): id/confidence/reusable_check/signal derivation helpers (#16)"
```

---

### Task 3: Render pattern artifacts in the v1 envelope

**Files:**
- Modify: `internal/core/knowledge/usecase.go` (function `renderFindings`, ~line 392)
- Test: `internal/core/knowledge/schema_test.go` (append)

Note: the `Run` call sites already pass the kind string (e.g. `renderFindings("etl_patterns", ...)`), so no call-site change is needed — only the function body.

- [ ] **Step 1: Write the failing test**

Append to `internal/core/knowledge/schema_test.go`:

```go
func TestRenderFindingsV1Envelope(t *testing.T) {
	out := renderFindings("etl_patterns", []Finding{
		{Name: "etl_validation", Domain: "etl", Evidence: "row count reconciliation source table target table", SourcePath: "normalized/etl/s1.md"},
	})
	for _, want := range []string{
		"schema_version: 1\n",
		"kind: etl_patterns\n",
		"generated_by: bqa ",
		"patterns:\n",
		"    domain:",
		"    reusable_check:",
		"    confidence: high\n",
		"    id:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("render missing %q, got:\n%s", want, out)
		}
	}
}

func TestRenderFindingsEmptyUsesFlowList(t *testing.T) {
	out := renderFindings("graphql_patterns", nil)
	if !strings.Contains(out, "kind: graphql_patterns\n") || !strings.Contains(out, "patterns: []\n") {
		t.Fatalf("unexpected empty render:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/knowledge/ -run TestRenderFindings -v`
Expected: FAIL — output still uses the old `etl_patterns:` root key, no `schema_version`.

- [ ] **Step 3: Replace `renderFindings`**

In `internal/core/knowledge/usecase.go`, replace the whole `renderFindings` function with:

```go
func renderFindings(kind string, items []Finding) string {
	var b strings.Builder
	b.WriteString(artifactHeader(kind))
	if len(items) == 0 {
		b.WriteString("patterns: []\n")
		return b.String()
	}
	items = uniqueFindings(items)
	b.WriteString("patterns:\n")
	for _, item := range items {
		b.WriteString("  - id: " + textutil.QuoteYAML(findingID(item)) + "\n")
		b.WriteString("    name: " + textutil.QuoteYAML(item.Name) + "\n")
		b.WriteString("    domain: " + textutil.QuoteYAML(item.Domain) + "\n")
		b.WriteString("    evidence: " + textutil.QuoteYAML(item.Evidence) + "\n")
		b.WriteString("    source: " + textutil.QuoteYAML(item.SourcePath) + "\n")
		b.WriteString("    reusable_check: " + textutil.QuoteYAML(reusableCheck(item)) + "\n")
		b.WriteString("    confidence: " + findingConfidence(item) + "\n")
	}
	return b.String()
}

// artifactHeader renders the common v1 envelope shared by every artifact.
func artifactHeader(kind string) string {
	return fmt.Sprintf("schema_version: %d\nkind: %s\ngenerated_by: %s\n", SchemaVersion, kind, generatedBy())
}
```

(`fmt`, `strings`, and `textutil` are already imported in this file.)

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/knowledge/ -run 'TestRenderFindings' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/knowledge/usecase.go internal/core/knowledge/schema_test.go
git commit -m "feat(knowledge): render pattern artifacts in v1 envelope (#16)"
```

---

### Task 4: Render project_profile in the v1 schema

**Files:**
- Modify: `internal/core/knowledge/usecase.go` (function `renderProfile`, ~line 409)
- Test: `internal/core/knowledge/schema_test.go` (append)

- [ ] **Step 1: Write the failing test**

Append to `internal/core/knowledge/schema_test.go`:

```go
func TestRenderProfileV1(t *testing.T) {
	out := renderProfile(ProjectProfile{Sessions: 12, ETLSignals: 8, GraphQLSignals: 3, APISignals: 2})
	for _, want := range []string{
		"schema_version: 1\n",
		"kind: project_profile\n",
		"profile:\n",
		"  sessions_analyzed: 12\n",
		"  domains_detected: [etl, graphql, api]\n",
		"  signals:\n",
		"    etl: 8\n",
		"  suggested_next_reviews:\n",
		"Review etl coverage (8 signals).",
		"  maturity: initial\n",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("profile render missing %q, got:\n%s", want, out)
		}
	}
}

func TestRenderProfileNoDomainsUsesEmptyList(t *testing.T) {
	out := renderProfile(ProjectProfile{Sessions: 1})
	if !strings.Contains(out, "domains_detected: []\n") || !strings.Contains(out, "suggested_next_reviews:\n    []\n") {
		t.Fatalf("unexpected empty-profile render:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/knowledge/ -run TestRenderProfile -v`
Expected: FAIL — old profile has no `schema_version`/`domains_detected`.

- [ ] **Step 3: Replace `renderProfile`**

In `internal/core/knowledge/usecase.go`, replace the whole `renderProfile` function with:

```go
func renderProfile(p ProjectProfile) string {
	var b strings.Builder
	b.WriteString(artifactHeader("project_profile"))
	b.WriteString("profile:\n")
	b.WriteString(fmt.Sprintf("  sessions_analyzed: %d\n", p.Sessions))

	sigs := detectedSignals(p)
	names := make([]string, 0, len(sigs))
	for _, s := range sigs {
		names = append(names, s.name)
	}
	if len(names) == 0 {
		b.WriteString("  domains_detected: []\n")
	} else {
		b.WriteString("  domains_detected: [" + strings.Join(names, ", ") + "]\n")
	}

	b.WriteString("  signals:\n")
	b.WriteString(fmt.Sprintf("    etl: %d\n", p.ETLSignals))
	b.WriteString(fmt.Sprintf("    graphql: %d\n", p.GraphQLSignals))
	b.WriteString(fmt.Sprintf("    api: %d\n", p.APISignals))
	b.WriteString(fmt.Sprintf("    data_quality: %d\n", p.DQSignals))
	b.WriteString(fmt.Sprintf("    droid: %d\n", p.DroidSignals))
	b.WriteString(fmt.Sprintf("    runtime: %d\n", p.RuntimeSignals))

	b.WriteString("  suggested_next_reviews:\n")
	if len(sigs) == 0 {
		b.WriteString("    []\n")
	} else {
		for _, s := range sigs {
			b.WriteString("    - " + textutil.QuoteYAML(fmt.Sprintf("Review %s coverage (%d signals).", s.name, s.count)) + "\n")
		}
	}
	b.WriteString("  maturity: initial\n")
	return b.String()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/knowledge/ -run TestRenderProfile -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/knowledge/usecase.go internal/core/knowledge/schema_test.go
git commit -m "feat(knowledge): render project_profile in v1 schema (#16)"
```

---

### Task 5: Validate against the v1 schema

**Files:**
- Modify: `internal/core/knowledge/validate.go` (the per-spec checks, lines 46-57)
- Test: `internal/core/knowledge/validate_test.go` (create)

- [ ] **Step 1: Write the failing test**

Create `internal/core/knowledge/validate_test.go`:

```go
package knowledge

import (
	"context"
	"testing"
)

type mapReader struct{ files map[string]string }

func (m mapReader) ReadKnowledgeArtifact(_ context.Context, name string) (string, error) {
	c, ok := m.files[name]
	if !ok {
		return "", errMissing
	}
	return c, nil
}

var errMissing = &missingErr{}

type missingErr struct{}

func (*missingErr) Error() string { return "missing" }

// buildValidFiles renders a full, valid v1 artifact set for the given profile.
func buildValidFiles() map[string]string {
	files := map[string]string{}
	for _, spec := range ExpectedArtifacts() {
		if spec.RootKey == "project_profile" {
			files[spec.Filename] = renderProfile(ProjectProfile{Sessions: 3, ETLSignals: 1})
		} else {
			files[spec.Filename] = renderFindings(spec.RootKey, nil)
		}
	}
	return files
}

func TestValidateAcceptsV1(t *testing.T) {
	rep := Validate(context.Background(), mapReader{files: buildValidFiles()})
	if !rep.OK() {
		t.Fatalf("expected valid v1 set, got issues: %+v", rep.Issues)
	}
}

func TestValidateRejectsMissingSchemaVersion(t *testing.T) {
	files := buildValidFiles()
	files["etl_patterns.yaml"] = "kind: etl_patterns\npatterns: []\n" // no schema_version
	rep := Validate(context.Background(), mapReader{files: files})
	if rep.OK() {
		t.Fatalf("expected failure for missing schema_version")
	}
}

func TestValidateRejectsWrongKind(t *testing.T) {
	files := buildValidFiles()
	files["etl_patterns.yaml"] = "schema_version: 1\nkind: wrong\npatterns: []\n"
	rep := Validate(context.Background(), mapReader{files: files})
	if rep.OK() {
		t.Fatalf("expected failure for wrong kind")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/knowledge/ -run TestValidate -v`
Expected: FAIL — `TestValidateRejectsMissingSchemaVersion`/`WrongKind` fail because current `Validate` only checks the old root key.

- [ ] **Step 3: Replace the per-spec check block in `Validate`**

In `internal/core/knowledge/validate.go`, replace lines 50-57 (the `missing root key` + `sessions_analyzed` checks) with:

```go
		if !strings.Contains(content, fmt.Sprintf("schema_version: %d", SchemaVersion)) {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: fmt.Sprintf("missing or unsupported schema_version (want %d)", SchemaVersion)})
			continue
		}
		if !strings.Contains(content, "kind: "+spec.RootKey) {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: fmt.Sprintf("missing or wrong kind (want %q)", spec.RootKey)})
			continue
		}
		if spec.RootKey == "project_profile" {
			if !strings.Contains(content, "profile:") || !strings.Contains(content, "sessions_analyzed:") {
				report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "missing profile block or sessions_analyzed"})
				continue
			}
		} else if !strings.Contains(content, "patterns:") {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "missing patterns list"})
			continue
		}
```

Also update the `Validate` doc comment (lines 31-35) to describe the v1 checks (schema_version, kind, patterns/profile).

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/knowledge/ -run TestValidate -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/knowledge/validate.go internal/core/knowledge/validate_test.go
git commit -m "feat(knowledge): validate the v1 schema (schema_version, kind, patterns/profile) (#16)"
```

---

### Task 6: Update the Codex knowledge parser for the v1 shape

**Files:**
- Modify: `internal/app/codex_knowledge.go` (`summarizeFindings`, and the `ingest2` text)
- Test: `internal/app/codex_knowledge_test.go` (update fixtures — see Task 7)

- [ ] **Step 1: Update `summarizeFindings`**

In `internal/app/codex_knowledge.go`, replace the `switch` body inside `summarizeFindings` (lines 71-85) with:

```go
		switch {
		case strings.HasPrefix(line, "- id:"):
			summary.count++
		case strings.HasPrefix(line, "name:"):
			name := unquoteYAMLScalar(strings.TrimSpace(strings.TrimPrefix(line, "name:")))
			if name != "" && !nameSeen[name] {
				nameSeen[name] = true
				summary.names = append(summary.names, name)
			}
		case strings.HasPrefix(line, "evidence:"):
			ev := unquoteYAMLScalar(strings.TrimSpace(strings.TrimPrefix(line, "evidence:")))
			ev = condenseEvidence(ev)
			if ev != "" && len(summary.samples) < 2 {
				summary.samples = append(summary.samples, ev)
			}
		}
```

Update the function's doc comment (lines 61-64) to say it parses the v1 `patterns:` list (`- id:` starts each entry; `name:`/`evidence:` are its fields).

- [ ] **Step 2: Fix the stale command hint**

In `codexNoKnowledgeSection` (line ~200), change `bqa ingest2` to `bqa ingest` so the hint matches the real command.

- [ ] **Step 3: Build to check compilation**

Run: `go build ./...`
Expected: success (tests updated in Task 7).

- [ ] **Step 4: Commit**

```bash
git add internal/app/codex_knowledge.go
git commit -m "feat(codex): parse v1 patterns list; fix stale ingest hint (#16)"
```

---

### Task 7: Update existing tests to the v1 shape

**Files:**
- Modify: `internal/app/codex_knowledge_test.go` (fixtures)
- Modify: `internal/core/knowledge/usecase_test.go`
- Modify: `internal/core/knowledge/quality_test.go`
- Modify: `internal/app/e2e_test.go`

- [ ] **Step 1: Update the Codex test fixtures**

In `internal/app/codex_knowledge_test.go`, the inline YAML fixtures use the old shape. Replace the pattern fixtures with v1 shape. For example, replace the `etl_patterns.yaml` fixture block (lines ~22-28) with:

```go
		"etl_patterns.yaml": `schema_version: 1
kind: etl_patterns
generated_by: bqa dev
patterns:
  - id: etl-00000000
    name: "etl_validation"
    domain: "etl"
    evidence: "row count reconciliation across partitions with retries and duplicates and schema drift"
    source: "normalized/etl/s1.md"
    reusable_check: "compare source vs target row counts for the window"
    confidence: high
`,
```

Apply the same transformation to the `data_quality_patterns.yaml`, `common_bugs.yaml`, and `successful_prompts.yaml` fixtures (envelope header + `patterns:` list with `- id:`/`name:`/`evidence:`/`source:`/`reusable_check:`/`confidence:`). Replace the empty `graphql_patterns.yaml` fixture (line 46) with:

```go
		"graphql_patterns.yaml": "schema_version: 1\nkind: graphql_patterns\ngenerated_by: bqa dev\npatterns: []\n",
```

For the `project_profile.yaml` fixture (lines ~49-53), wrap it in the v1 profile shape:

```go
		"project_profile.yaml": `schema_version: 1
kind: project_profile
generated_by: bqa dev
profile:
  sessions_analyzed: 4
  domains_detected: [etl]
  signals:
    etl: 2
    graphql: 0
    api: 0
    data_quality: 0
    droid: 0
    runtime: 0
  suggested_next_reviews:
    - "Review etl coverage (2 signals)."
  maturity: initial
`,
```

Keep the existing assertions that check the rendered Codex section contains the pattern names, ETL signal words, and privacy note — those still hold because the parser now reads the v1 list.

- [ ] **Step 2: Run the Codex tests**

Run: `go test ./internal/app/ -run Codex -v`
Expected: PASS. If an assertion referenced a raw source path or an old-shape string, update it to the v1 equivalent.

- [ ] **Step 3: Update knowledge unit/quality tests**

In `internal/core/knowledge/usecase_test.go` and `quality_test.go`, most assertions check for substrings that survive (finding `name` values like `etl_validation`, evidence keywords). Run them and fix only failures:

Run: `go test ./internal/core/knowledge/ -v`
Expected: PASS. For any failure asserting the old root key (e.g. a check for `"etl_patterns:\n  - name:"`), change it to assert `"kind: etl_patterns"` and `"patterns:"`. For any check of `project_profile:` as a root key, change to `"kind: project_profile"`.

- [ ] **Step 4: Update the e2e assertions**

In `internal/app/e2e_test.go`, the domain checks (lines ~96-108) assert the old root keys and `- name:`. Change the per-file expected substring from the root key (e.g. `"etl_patterns:"`) to `"kind: etl_patterns"`, and change the findings check from `strings.Contains(body, "- name:")` to `strings.Contains(body, "- id:")`. The `sessions_analyzed:` profile assertion (lines 118-119) still holds.

- [ ] **Step 5: Run the full suite**

Run: `go test ./...`
Expected: PASS (all packages).

- [ ] **Step 6: Commit**

```bash
git add internal/app/codex_knowledge_test.go internal/core/knowledge/usecase_test.go internal/core/knowledge/quality_test.go internal/app/e2e_test.go
git commit -m "test: update knowledge/codex/e2e assertions to v1 schema (#16)"
```

---

### Task 8: Determinism + required-fields acceptance tests

**Files:**
- Test: `internal/core/knowledge/schema_test.go` (append)

- [ ] **Step 1: Write the tests**

Append to `internal/core/knowledge/schema_test.go`:

```go
func TestRenderIsDeterministic(t *testing.T) {
	items := []Finding{
		{Name: "etl_validation", Domain: "etl", Evidence: "row count reconciliation", SourcePath: "normalized/etl/s1.md"},
		{Name: "api_contract_testing", Domain: "api", Evidence: "status code 500", SourcePath: "normalized/api/s2.md"},
	}
	if renderFindings("etl_patterns", items) != renderFindings("etl_patterns", items) {
		t.Fatal("renderFindings is not deterministic")
	}
	p := ProjectProfile{Sessions: 5, ETLSignals: 2, APISignals: 1}
	if renderProfile(p) != renderProfile(p) {
		t.Fatal("renderProfile is not deterministic")
	}
}

func TestPatternArtifactHasAllRequiredFields(t *testing.T) {
	out := renderFindings("api_patterns", []Finding{
		{Name: "api_contract_testing", Domain: "api", Evidence: "endpoint status code contract", SourcePath: "normalized/api/s1.md"},
	})
	for _, field := range []string{"schema_version:", "kind:", "generated_by:", "- id:", "name:", "domain:", "evidence:", "source:", "reusable_check:", "confidence:"} {
		if !strings.Contains(out, field) {
			t.Fatalf("pattern artifact missing required field %q", field)
		}
	}
}

func TestProfileArtifactHasAllRequiredFields(t *testing.T) {
	out := renderProfile(ProjectProfile{Sessions: 4, ETLSignals: 2})
	for _, field := range []string{"schema_version:", "kind: project_profile", "sessions_analyzed:", "domains_detected:", "signals:", "suggested_next_reviews:"} {
		if !strings.Contains(out, field) {
			t.Fatalf("profile artifact missing required field %q", field)
		}
	}
}
```

- [ ] **Step 2: Run to verify pass**

Run: `go test ./internal/core/knowledge/ -run 'TestRenderIsDeterministic|TestPatternArtifactHasAll|TestProfileArtifactHasAll' -v`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/core/knowledge/schema_test.go
git commit -m "test(knowledge): determinism + required-field presence for v1 (#16)"
```

---

### Task 9: Schema documentation

**Files:**
- Create: `docs/knowledge-artifacts.md`

- [ ] **Step 1: Write the doc**

Create `docs/knowledge-artifacts.md` covering:
- Purpose and the hard-cutover note (artifacts are derived; regenerate with `bqa build`).
- The common envelope (`schema_version`, `kind`, `generated_by`), with the exact synthetic pattern example and the exact profile example from the design spec (copy them verbatim so the doc matches real output).
- A field table: `id` (content hash `<domain>-<8hex>`), `name`, `domain`, `evidence`, `source`, `reusable_check` (a candidate — review before use), `confidence` (low/medium/high by the documented rule: distinct domain keywords in evidence — high ≥3, medium 2, low 1; a heuristic, not a probability).
- The profile fields: `sessions_analyzed`, `domains_detected`, `signals`, `suggested_next_reviews`, `maturity`.
- The **forward-compatibility policy** section verbatim from the spec (additive = no bump + ignore unknown fields; breaking = bump; consumers read `schema_version`; no migrator).

Use only synthetic examples.

- [ ] **Step 2: Sanity check**

Run: `go build ./...`
Expected: success (no code touched).

- [ ] **Step 3: Commit**

```bash
git add docs/knowledge-artifacts.md
git commit -m "docs: document the v1 knowledge artifact schema (#16)"
```

---

### Task 10: Align the JS mirror and example artifacts

**Files:**
- Modify: `docs/assets/upload.js` (client-side artifact generation)
- Modify: `examples/knowledge/successful_prompts.yaml`
- Modify: `docs/knowledge-review-checklist.md`
- Test: `node --test docs/assets/*.test.js`

The game reimplements the artifact shape client-side; align it so the demo stays honest.

- [ ] **Step 1: Inspect the JS generator**

Read the function in `docs/assets/upload.js` that builds the `.bqa/knowledge/*.yaml` strings (search for `patterns` / `_patterns` / `sessions_analyzed`). Note how each artifact string is assembled.

- [ ] **Step 2: Update the JS to emit the v1 envelope**

For each pattern artifact string, emit:

```
schema_version: 1
kind: <name>
generated_by: bqa web
patterns:
  - id: <domain>-<8hex>
    name: "..."
    domain: <domain>
    evidence: "..."
    source: "..."
    reusable_check: "..."
    confidence: <low|medium|high>
```

and for the profile emit the `profile:` block with `domains_detected` and `suggested_next_reviews`. Keep a JS `sha256`→8hex (or a small deterministic hash already present in the file) for `id`; if none exists, use a short deterministic hash of `domain|name|source`. Match the Go rules closely enough that the demo output looks like real output.

- [ ] **Step 3: Update the tests and example files**

- Update any assertions in `docs/assets/*.test.js` (e.g. `upload.test.js`/`scorecard.test.js`) that check the old artifact shape to the v1 shape.
- Rewrite `examples/knowledge/successful_prompts.yaml` in the v1 envelope (`schema_version: 1`, `kind: successful_prompts`, `patterns:` with the new fields; keep the ≥5 synthetic records).
- In `docs/knowledge-review-checklist.md`, update any inline YAML snippet that shows the old shape to the v1 shape.

- [ ] **Step 4: Run the JS tests**

Run: `node --test docs/assets/game.test.js docs/assets/game-stage.test.js docs/assets/scorecard.test.js`
Expected: PASS (0 fail; the Playwright test may skip without `BQA_BROWSER_QA=1`).

- [ ] **Step 5: Commit**

```bash
git add docs/assets/upload.js docs/assets/*.test.js examples/knowledge/successful_prompts.yaml docs/knowledge-review-checklist.md
git commit -m "chore: align JS mirror + example artifacts to v1 schema (#16)"
```

---

### Task 11: Final verification

- [ ] **Step 1: Full Go suite + vet + build**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: build clean, vet clean, all packages `ok`.

- [ ] **Step 2: JS suite**

Run: `node --test docs/assets/game.test.js docs/assets/game-stage.test.js docs/assets/scorecard.test.js`
Expected: 0 failures.

- [ ] **Step 3: End-to-end sanity (optional, manual)**

Build the binary and run the workflow on synthetic input, then confirm `.bqa/knowledge/etl_patterns.yaml` begins with `schema_version: 1` / `kind: etl_patterns` and that `bqa build --check` passes:

```bash
go run ./cmd/bqa build --check --sessions <tmp>/.bqa/input/sessions --knowledge-dir <tmp>/.bqa/knowledge
```

- [ ] **Step 4: Push and open PR**

```bash
git push -u origin feature/issue-16-knowledge-schema
gh pr create --title "feat: stable v1 knowledge artifact schema (#16)" --body "Implements the v1 schema per docs/superpowers/specs/2026-07-01-knowledge-artifact-schema-v1-design.md. Closes #16."
```

---

## Notes for the implementer

- **DRY:** `artifactHeader` is the single source of the envelope; both renderers and the JS mirror follow it.
- **YAGNI:** no YAML library, no migrator, no cross-session confidence — all deferred per the spec.
- **Determinism:** never add a wall-clock timestamp; tests compare bytes and `generated_by` is `bqa dev` under test.
- **Hard cutover:** there is intentionally no reader for the old unversioned shape — old files are regenerated by `bqa build`.
