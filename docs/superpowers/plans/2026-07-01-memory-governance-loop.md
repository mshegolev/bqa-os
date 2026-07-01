# Memory Governance Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `bqa memory` command group that extracts *candidate* QA memory from normalized sessions, lets a human review it, and promotes/rejects candidates by id — so nothing enters stable memory without explicit human approval (issue #40, slice 2).

**Architecture:** Strict hexagonal. All pure logic (candidate extraction, content-hash ids, YAML render/parse, the learn/review/promote/reject use cases) lives in `internal/core/memgov` and depends only on `ports` + stdlib (plus the shared `textutil`/`version` helpers, matching `core/knowledge`). The only side effects are behind two ports: the **existing** `ports.NormalizedSessionReader` (reused, not new) for reading sessions, and a new file-oriented `ports.GovernanceStore` for reading/writing the memory files. The fs adapter that implements `GovernanceStore` is pure file I/O and imports only `ports` + stdlib — no adapter in this repo imports `core`, and this plan keeps that invariant.

**Deviation from spec (intentional, noted here):** The spec's Ports sketch shows `GovernanceStore.Load/Save` returning a parsed `GovernanceState`. That would force YAML parsing *into the adapter*, contradicting the spec's own statement that "YAML render/parse of the memory files lives in `core/memgov`" and the repo rule that adapters depend on ports+stdlib only. We resolve this by making `GovernanceStore` a **file-level** port (`ReadFile`/`WriteFile`) and keeping `GovernanceState` + all parse/render in `core/memgov`. Same behavior, same files on disk, cleaner layering.

**Tech Stack:** Go 1.22, stdlib + cobra only. Deterministic output (no wall-clock; `generated_by` comes from `version.Version`, which is `dev` in tests → `bqa dev`). v1 YAML envelope reused from `core/knowledge`.

---

## File Structure

**New files:**
- `internal/ports/memgov.go` — the `GovernanceStore` port (file-level read/write). Reuses existing `ports.NormalizedSessionReader`.
- `internal/core/memgov/models.go` — kinds/filenames constants, `SchemaVersion`, `MemoryItem`, `DecisionEntry`, `GovernanceState`, result structs, `generatedBy()`.
- `internal/core/memgov/id.go` — `ItemID(kind, name, source)` content-hash id.
- `internal/core/memgov/render.go` — `renderItems`, `renderLog`, `header` (v1 envelope).
- `internal/core/memgov/parse.go` — `parseItems`, `parseLog` (line scanners mirroring the render format).
- `internal/core/memgov/state.go` — `loadState`/`saveState` over the `GovernanceStore` port, plus `GovernanceState` helpers.
- `internal/core/memgov/learn.go` — keyword extraction heuristics + `Learn`.
- `internal/core/memgov/review.go` — `Review`.
- `internal/core/memgov/decide.go` — `Promote`/`Reject` (shared `decide`).
- `internal/core/memgov/usecase.go` — `UseCase` struct wiring the ports + `MemoryDir`.
- `internal/core/memgov/memgov_test.go` — learn/review/promote/reject/errors/idempotency, using in-memory fakes.
- `internal/adapters/fs/governance_store.go` — fs implementation of `GovernanceStore`.
- `internal/adapters/fs/governance_store_test.go` — round-trips a temp dir.
- `internal/app/memory.go` — `memoryCmd()` (learn/review/promote/reject wiring).
- `internal/app/memory_test.go` — app-level smoke test of the command tree over a temp dir.

**Modified files:**
- `internal/textutil/textutil.go` — add `UnquoteYAML` (exact inverse of `QuoteYAML`).
- `internal/textutil/textutil_test.go` — round-trip test for `QuoteYAML`/`UnquoteYAML`.
- `internal/app/root.go` — register `memoryCmd()`.

---

## Task 1: `textutil.UnquoteYAML` round-trip helper

Core parsing needs to reverse `QuoteYAML`. There is an `unquoteYAMLScalar` in package `app`, but `core/*` cannot import `app`. Add a faithful inverse to the shared `textutil` package.

**Files:**
- Modify: `internal/textutil/textutil.go`
- Test: `internal/textutil/textutil_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/textutil/textutil_test.go`:

```go
package textutil

import "testing"

func TestQuoteUnquoteYAMLRoundTrip(t *testing.T) {
	cases := []string{
		"plain",
		`has "quotes"`,
		"has\nnewline\tand tab",
		`back\slash`,
		"",
	}
	for _, in := range cases {
		got := UnquoteYAML(QuoteYAML(in))
		if got != in {
			t.Fatalf("round-trip mismatch for %q: got %q", in, got)
		}
	}
}

func TestUnquoteYAMLBareScalar(t *testing.T) {
	if got := UnquoteYAML("  bareword  "); got != "bareword" {
		t.Fatalf("expected trimmed bareword, got %q", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/textutil/ -run YAML -v`
Expected: FAIL — `undefined: UnquoteYAML`.

- [ ] **Step 3: Add the implementation**

Append to `internal/textutil/textutil.go`:

```go
// UnquoteYAML reverses QuoteYAML: it trims surrounding whitespace and, when the
// value is a double-quoted scalar, unescapes the exact set of escapes QuoteYAML
// produces (\\, \", \r, \n, \t). Bare (unquoted) scalars are returned trimmed.
func UnquoteYAML(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		inner := value[1 : len(value)-1]
		r := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\"`, "\"", `\\`, "\\")
		return r.Replace(inner)
	}
	return value
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/textutil/ -run YAML -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/textutil/textutil.go internal/textutil/textutil_test.go
git commit -m "feat(textutil): add UnquoteYAML inverse of QuoteYAML"
```

---

## Task 2: `GovernanceStore` port

A file-level port: read a named memory file (report absence), write one atomically. `NormalizedSessionReader` already exists and is reused as-is.

**Files:**
- Create: `internal/ports/memgov.go`

- [ ] **Step 1: Create the port file**

Create `internal/ports/memgov.go`:

```go
package ports

import "context"

// GovernanceStore reads and writes the governance memory files under a memory
// directory. It is deliberately file-level (bytes in, bytes out) so that all
// YAML parse/render logic lives in core/memgov and this adapter stays pure I/O.
type GovernanceStore interface {
	// ReadFile returns the file content. exists is false (with nil error) when
	// the file is absent, so a first run reads an empty governance state.
	ReadFile(ctx context.Context, memoryDir string, name string) (content string, exists bool, err error)
	// WriteFile writes content to the named file under memoryDir, creating the
	// directory if needed.
	WriteFile(ctx context.Context, memoryDir string, name string, content string) error
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/ports/`
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add internal/ports/memgov.go
git commit -m "feat(ports): add GovernanceStore file-level port"
```

---

## Task 3: `memgov` models, ids, render, parse (pure)

Pure data model plus deterministic YAML render/parse. No I/O.

**Files:**
- Create: `internal/core/memgov/models.go`
- Create: `internal/core/memgov/id.go`
- Create: `internal/core/memgov/render.go`
- Create: `internal/core/memgov/parse.go`
- Test: `internal/core/memgov/memgov_test.go` (created here, extended in later tasks)

- [ ] **Step 1: Write the failing test**

Create `internal/core/memgov/memgov_test.go`:

```go
package memgov

import (
	"strings"
	"testing"
)

func TestItemIDStableAndPrefixed(t *testing.T) {
	a := ItemID(KindSkillCandidates, "etl reconciliation check", "normalized/s1.md")
	b := ItemID(KindSkillCandidates, "etl reconciliation check", "normalized/s1.md")
	if a != b {
		t.Fatalf("id not stable: %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, "skill-") {
		t.Fatalf("expected skill- prefix, got %q", a)
	}
	if ItemID(KindLessons, "x", "y") == ItemID(KindSkillCandidates, "x", "y") {
		t.Fatalf("different kinds must yield different ids")
	}
	if !strings.HasPrefix(ItemID(KindLessons, "x", "y"), "lesson-") {
		t.Fatalf("expected lesson- prefix")
	}
}

func TestRenderParseItemsRoundTrip(t *testing.T) {
	items := []MemoryItem{
		{ID: "skill-0a1b2c3d", Name: `a "quoted" name`, Domain: "etl", Evidence: "line one\nline two", Source: "normalized/s1.md", Status: "pending"},
	}
	out := renderItems(KindSkillCandidates, items)
	if !strings.Contains(out, "schema_version: 1\n") || !strings.Contains(out, "kind: skill_candidates\n") {
		t.Fatalf("missing v1 envelope:\n%s", out)
	}
	got := parseItems(out)
	if len(got) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got))
	}
	if got[0] != items[0] {
		t.Fatalf("round-trip mismatch:\n want %#v\n got  %#v", items[0], got[0])
	}
}

func TestRenderEmptyItems(t *testing.T) {
	out := renderItems(KindApproved, nil)
	if !strings.Contains(out, "items: []\n") {
		t.Fatalf("expected empty items list, got:\n%s", out)
	}
	if len(parseItems(out)) != 0 {
		t.Fatalf("expected zero parsed items")
	}
}

func TestRenderParseLogRoundTrip(t *testing.T) {
	entries := []DecisionEntry{{ID: "skill-0a1b2c3d", Action: "promoted", Name: "etl check"}}
	out := renderLog(entries)
	if !strings.Contains(out, "kind: decision_log\n") {
		t.Fatalf("missing decision_log kind:\n%s", out)
	}
	got := parseLog(out)
	if len(got) != 1 || got[0] != entries[0] {
		t.Fatalf("log round-trip mismatch: %#v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/memgov/ -v`
Expected: FAIL — package does not compile (`undefined: ItemID`, etc.).

- [ ] **Step 3: Create `models.go`**

Create `internal/core/memgov/models.go`:

```go
package memgov

import "github.com/mshegolev/bqa-os/internal/version"

// SchemaVersion is the v1 envelope version stamped on every governance file.
const SchemaVersion = 1

// Governance file kinds. Each kind maps 1:1 to a "<kind>.yaml" file.
const (
	KindLessons         = "lessons_learned"
	KindSkillCandidates = "skill_candidates"
	KindApproved        = "approved_patterns"
	KindRejected        = "rejected_patterns"
	KindDecisionLog     = "decision_log"
)

// Item statuses.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// DefaultMemoryDir is the private memory area (portable via `bqa brain`).
const DefaultMemoryDir = ".bqa/memory"

// kindPrefix is the short id prefix per candidate kind.
var kindPrefix = map[string]string{
	KindLessons:         "lesson",
	KindSkillCandidates: "skill",
}

// candidateFiles lists the two files that hold reviewable candidates.
// stateFiles lists every governance file (candidates + decided + log).
var (
	candidateFiles = []string{KindLessons, KindSkillCandidates}
	stateFiles     = []string{KindLessons, KindSkillCandidates, KindApproved, KindRejected, KindDecisionLog}
)

// fileName returns the on-disk filename for a kind.
func fileName(kind string) string { return kind + ".yaml" }

// generatedBy is the provenance stamp: "bqa dev" in dev/test, "bqa vX.Y.Z" in a
// release build. Deterministic in tests.
func generatedBy() string { return "bqa " + version.Version }

// MemoryItem is one governed memory record carrying the v1 envelope fields plus a
// governance status.
type MemoryItem struct {
	ID       string
	Name     string
	Domain   string
	Evidence string
	Source   string
	Status   string
}

// DecisionEntry is one appended decision_log record. No timestamp — decisions are
// deterministic for now (a future slice may add wall-clock).
type DecisionEntry struct {
	ID     string
	Action string // "promoted" | "rejected"
	Name   string
}

// GovernanceState is the full in-memory governance state.
type GovernanceState struct {
	Lessons         []MemoryItem
	SkillCandidates []MemoryItem
	Approved        []MemoryItem
	Rejected        []MemoryItem
	Log             []DecisionEntry
}

// LearnResult reports what a learn run added.
type LearnResult struct {
	SessionsProcessed int
	LessonsAdded      int
	SkillsAdded       int
}

// ReviewResult holds the pending candidates for display.
type ReviewResult struct {
	Pending []MemoryItem
}

// DecideResult reports the moved item and the action taken.
type DecideResult struct {
	Item   MemoryItem
	Action string // "promoted" | "rejected"
}
```

- [ ] **Step 4: Create `id.go`**

Create `internal/core/memgov/id.go`:

```go
package memgov

import (
	"crypto/sha256"
	"encoding/hex"
)

// ItemID returns a stable, content-derived id "<prefix>-<8 hex>" from the
// candidate kind, name, and source. Stable across insertions/removals so the
// governance files diff and merge cleanly.
func ItemID(kind, name, source string) string {
	sum := sha256.Sum256([]byte(kind + "|" + name + "|" + source))
	return kindPrefix[kind] + "-" + hex.EncodeToString(sum[:])[:8]
}
```

- [ ] **Step 5: Create `render.go`**

Create `internal/core/memgov/render.go`:

```go
package memgov

import (
	"fmt"
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
)

// header renders the shared v1 envelope.
func header(kind string) string {
	return fmt.Sprintf("schema_version: %d\nkind: %s\ngenerated_by: %s\n", SchemaVersion, kind, generatedBy())
}

// renderItems renders a kind's items file. id and status are barewords (safe by
// construction); free-text fields are YAML-quoted.
func renderItems(kind string, items []MemoryItem) string {
	var b strings.Builder
	b.WriteString(header(kind))
	if len(items) == 0 {
		b.WriteString("items: []\n")
		return b.String()
	}
	b.WriteString("items:\n")
	for _, it := range items {
		b.WriteString("  - id: " + it.ID + "\n")
		b.WriteString("    name: " + textutil.QuoteYAML(it.Name) + "\n")
		b.WriteString("    domain: " + textutil.QuoteYAML(it.Domain) + "\n")
		b.WriteString("    evidence: " + textutil.QuoteYAML(it.Evidence) + "\n")
		b.WriteString("    source: " + textutil.QuoteYAML(it.Source) + "\n")
		b.WriteString("    status: " + it.Status + "\n")
	}
	return b.String()
}

// renderLog renders the decision_log file.
func renderLog(entries []DecisionEntry) string {
	var b strings.Builder
	b.WriteString(header(KindDecisionLog))
	if len(entries) == 0 {
		b.WriteString("entries: []\n")
		return b.String()
	}
	b.WriteString("entries:\n")
	for _, e := range entries {
		b.WriteString("  - id: " + e.ID + "\n")
		b.WriteString("    action: " + e.Action + "\n")
		b.WriteString("    name: " + textutil.QuoteYAML(e.Name) + "\n")
	}
	return b.String()
}
```

- [ ] **Step 6: Create `parse.go`**

Create `internal/core/memgov/parse.go`:

```go
package memgov

import (
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
)

// parseItems parses an items file rendered by renderItems. Unknown lines and the
// envelope header are ignored; each "- id:" starts a new item.
func parseItems(content string) []MemoryItem {
	var items []MemoryItem
	var cur *MemoryItem
	flush := func() {
		if cur != nil {
			items = append(items, *cur)
			cur = nil
		}
	}
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "- id:"):
			flush()
			cur = &MemoryItem{ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:"))}
		case cur == nil:
			continue
		case strings.HasPrefix(line, "name:"):
			cur.Name = textutil.UnquoteYAML(strings.TrimPrefix(line, "name:"))
		case strings.HasPrefix(line, "domain:"):
			cur.Domain = textutil.UnquoteYAML(strings.TrimPrefix(line, "domain:"))
		case strings.HasPrefix(line, "evidence:"):
			cur.Evidence = textutil.UnquoteYAML(strings.TrimPrefix(line, "evidence:"))
		case strings.HasPrefix(line, "source:"):
			cur.Source = textutil.UnquoteYAML(strings.TrimPrefix(line, "source:"))
		case strings.HasPrefix(line, "status:"):
			cur.Status = strings.TrimSpace(strings.TrimPrefix(line, "status:"))
		}
	}
	flush()
	return items
}

// parseLog parses a decision_log file rendered by renderLog.
func parseLog(content string) []DecisionEntry {
	var entries []DecisionEntry
	var cur *DecisionEntry
	flush := func() {
		if cur != nil {
			entries = append(entries, *cur)
			cur = nil
		}
	}
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "- id:"):
			flush()
			cur = &DecisionEntry{ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:"))}
		case cur == nil:
			continue
		case strings.HasPrefix(line, "action:"):
			cur.Action = strings.TrimSpace(strings.TrimPrefix(line, "action:"))
		case strings.HasPrefix(line, "name:"):
			cur.Name = textutil.UnquoteYAML(strings.TrimPrefix(line, "name:"))
		}
	}
	flush()
	return entries
}
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test ./internal/core/memgov/ -v`
Expected: PASS (`TestItemIDStableAndPrefixed`, `TestRenderParseItemsRoundTrip`, `TestRenderEmptyItems`, `TestRenderParseLogRoundTrip`).

- [ ] **Step 8: Commit**

```bash
git add internal/core/memgov/models.go internal/core/memgov/id.go internal/core/memgov/render.go internal/core/memgov/parse.go internal/core/memgov/memgov_test.go
git commit -m "feat(memgov): pure model, ids, and YAML render/parse"
```

---

## Task 4: State load/save over the port + `UseCase`

`loadState` reads every governance file via the port (absent → empty list); `saveState` writes all five deterministically. `UseCase` wires the two ports plus the memory dir.

**Files:**
- Create: `internal/core/memgov/state.go`
- Create: `internal/core/memgov/usecase.go`
- Test: `internal/core/memgov/memgov_test.go` (extend)

- [ ] **Step 1: Write the failing test**

Append to `internal/core/memgov/memgov_test.go`:

```go
// memStore is an in-memory GovernanceStore keyed by "<dir>/<name>".
type memStore struct{ files map[string]string }

func newMemStore() *memStore { return &memStore{files: map[string]string{}} }

func (m *memStore) ReadFile(_ context.Context, dir, name string) (string, bool, error) {
	c, ok := m.files[dir+"/"+name]
	return c, ok, nil
}

func (m *memStore) WriteFile(_ context.Context, dir, name, content string) error {
	m.files[dir+"/"+name] = content
	return nil
}

func TestSaveLoadStateRoundTrip(t *testing.T) {
	store := newMemStore()
	want := GovernanceState{
		SkillCandidates: []MemoryItem{{ID: "skill-1", Name: "n", Domain: "etl", Evidence: "e", Source: "s", Status: StatusPending}},
		Log:             []DecisionEntry{{ID: "skill-1", Action: "promoted", Name: "n"}},
	}
	if err := saveState(context.Background(), store, ".bqa/memory", want); err != nil {
		t.Fatalf("saveState: %v", err)
	}
	got, err := loadState(context.Background(), store, ".bqa/memory")
	if err != nil {
		t.Fatalf("loadState: %v", err)
	}
	if len(got.SkillCandidates) != 1 || got.SkillCandidates[0].ID != "skill-1" {
		t.Fatalf("skill candidates not round-tripped: %#v", got.SkillCandidates)
	}
	if len(got.Log) != 1 || got.Log[0].Action != "promoted" {
		t.Fatalf("log not round-tripped: %#v", got.Log)
	}
}

func TestLoadStateEmptyWhenAbsent(t *testing.T) {
	got, err := loadState(context.Background(), newMemStore(), ".bqa/memory")
	if err != nil {
		t.Fatalf("loadState: %v", err)
	}
	if len(got.SkillCandidates) != 0 || len(got.Lessons) != 0 || len(got.Log) != 0 {
		t.Fatalf("expected empty state, got %#v", got)
	}
}
```

Add `"context"` to the test file's import block (alongside the existing `strings`/`testing`).

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/memgov/ -run State -v`
Expected: FAIL — `undefined: saveState` / `loadState`.

- [ ] **Step 3: Create `state.go`**

Create `internal/core/memgov/state.go`:

```go
package memgov

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// loadState reads every governance file via the store and parses it. A missing
// file yields an empty list, so a first run starts from an empty state.
func loadState(ctx context.Context, store ports.GovernanceStore, memoryDir string) (GovernanceState, error) {
	var st GovernanceState
	read := func(kind string) ([]MemoryItem, error) {
		content, exists, err := store.ReadFile(ctx, memoryDir, fileName(kind))
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, nil
		}
		return parseItems(content), nil
	}

	var err error
	if st.Lessons, err = read(KindLessons); err != nil {
		return GovernanceState{}, err
	}
	if st.SkillCandidates, err = read(KindSkillCandidates); err != nil {
		return GovernanceState{}, err
	}
	if st.Approved, err = read(KindApproved); err != nil {
		return GovernanceState{}, err
	}
	if st.Rejected, err = read(KindRejected); err != nil {
		return GovernanceState{}, err
	}

	logContent, exists, err := store.ReadFile(ctx, memoryDir, fileName(KindDecisionLog))
	if err != nil {
		return GovernanceState{}, err
	}
	if exists {
		st.Log = parseLog(logContent)
	}
	return st, nil
}

// saveState writes all governance files deterministically. Re-writing unchanged
// files is safe because render output is stable.
func saveState(ctx context.Context, store ports.GovernanceStore, memoryDir string, st GovernanceState) error {
	byKind := map[string][]MemoryItem{
		KindLessons:         st.Lessons,
		KindSkillCandidates: st.SkillCandidates,
		KindApproved:        st.Approved,
		KindRejected:        st.Rejected,
	}
	for _, kind := range candidateFiles {
		if err := store.WriteFile(ctx, memoryDir, fileName(kind), renderItems(kind, byKind[kind])); err != nil {
			return err
		}
	}
	for _, kind := range []string{KindApproved, KindRejected} {
		if err := store.WriteFile(ctx, memoryDir, fileName(kind), renderItems(kind, byKind[kind])); err != nil {
			return err
		}
	}
	return store.WriteFile(ctx, memoryDir, fileName(KindDecisionLog), renderLog(st.Log))
}

// idSet returns the set of every id present anywhere in the state, used to keep
// learn idempotent.
func (s GovernanceState) idSet() map[string]bool {
	set := map[string]bool{}
	for _, list := range [][]MemoryItem{s.Lessons, s.SkillCandidates, s.Approved, s.Rejected} {
		for _, it := range list {
			set[it.ID] = true
		}
	}
	return set
}
```

- [ ] **Step 4: Create `usecase.go`**

Create `internal/core/memgov/usecase.go`:

```go
package memgov

import "github.com/mshegolev/bqa-os/internal/ports"

// UseCase governs candidate QA memory: learn candidates, review pending ones, and
// promote/reject them by id. All side effects go through the two ports.
type UseCase struct {
	Reader    ports.NormalizedSessionReader
	Store     ports.GovernanceStore
	MemoryDir string
}

// dir returns the configured memory dir or the default.
func (u UseCase) dir() string {
	if u.MemoryDir == "" {
		return DefaultMemoryDir
	}
	return u.MemoryDir
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/core/memgov/ -run State -v`
Expected: PASS (`TestSaveLoadStateRoundTrip`, `TestLoadStateEmptyWhenAbsent`).

- [ ] **Step 6: Commit**

```bash
git add internal/core/memgov/state.go internal/core/memgov/usecase.go internal/core/memgov/memgov_test.go
git commit -m "feat(memgov): governance state load/save and UseCase wiring"
```

---

## Task 5: `Learn` — extract candidates (idempotent, bounded evidence)

Keyword heuristic over normalized sessions. Skill candidates come from domain signal words; lessons from failure signals. Evidence is a bounded, single-line snippet (never a raw body). An id already present anywhere is not re-added.

**Files:**
- Create: `internal/core/memgov/learn.go`
- Test: `internal/core/memgov/memgov_test.go` (extend)

- [ ] **Step 1: Write the failing test**

Append to `internal/core/memgov/memgov_test.go`:

```go
type fakeSessions struct {
	index ports.SessionIndex
	files map[string]string
}

func (f fakeSessions) LoadSessionIndex(context.Context) (ports.SessionIndex, error) {
	return f.index, nil
}

func (f fakeSessions) ReadNormalizedSession(_ context.Context, path string) (string, error) {
	return f.files[path], nil
}

func learnFixture() (UseCase, *memStore) {
	store := newMemStore()
	reader := fakeSessions{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "/s1.jsonl", NormalizedPath: "s1.md"},
		}},
		files: map[string]string{
			"s1.md": "We ran the ETL reconciliation and compared source table vs target table row count. " +
				"The job failed with a traceback during the null check.",
		},
	}
	return UseCase{Reader: reader, Store: store, MemoryDir: ".bqa/memory"}, store
}

func TestLearnExtractsPendingCandidates(t *testing.T) {
	uc, store := learnFixture()
	res, err := uc.Learn(context.Background())
	if err != nil {
		t.Fatalf("Learn: %v", err)
	}
	if res.SessionsProcessed != 1 {
		t.Fatalf("expected 1 session processed, got %d", res.SessionsProcessed)
	}
	if res.SkillsAdded == 0 || res.LessonsAdded == 0 {
		t.Fatalf("expected skills and lessons added, got %+v", res)
	}
	skills := store.files[".bqa/memory/skill_candidates.yaml"]
	if !strings.Contains(skills, "status: pending") || !strings.Contains(skills, "domain: etl") {
		t.Fatalf("skill_candidates missing pending etl item:\n%s", skills)
	}
	// Evidence must be bounded (no raw dumps) and single line.
	for _, it := range parseItems(skills) {
		if len(it.Evidence) > evidenceWindow {
			t.Fatalf("evidence not bounded: %d chars", len(it.Evidence))
		}
		if strings.Contains(it.Evidence, "\n") {
			t.Fatalf("evidence should be single-line")
		}
	}
	lessons := store.files[".bqa/memory/lessons_learned.yaml"]
	if !strings.Contains(lessons, "domain: bugs") {
		t.Fatalf("lessons_learned missing bugs item:\n%s", lessons)
	}
}

func TestLearnIsIdempotent(t *testing.T) {
	uc, _ := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("first Learn: %v", err)
	}
	res, err := uc.Learn(context.Background())
	if err != nil {
		t.Fatalf("second Learn: %v", err)
	}
	if res.SkillsAdded != 0 || res.LessonsAdded != 0 {
		t.Fatalf("second learn should add nothing, got %+v", res)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/memgov/ -run Learn -v`
Expected: FAIL — `undefined: (UseCase).Learn` and `undefined: evidenceWindow`.

- [ ] **Step 3: Create `learn.go`**

Create `internal/core/memgov/learn.go`:

```go
package memgov

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

// evidenceWindow bounds a captured snippet: wide enough for a QA context sentence,
// far below a transcript body so no raw session leaks verbatim.
const evidenceWindow = 480

// skillRule maps signal words to a stable skill-candidate name + domain.
type skillRule struct {
	needles []string
	name    string
	domain  string
}

var skillRules = []skillRule{
	{[]string{"reconciliation", "row count", "source table", "target table"}, "etl reconciliation check", "etl"},
	{[]string{"data quality", "null check", "duplicate", "schema drift", "checksum"}, "data quality validation check", "data_quality"},
	{[]string{"rest api", "endpoint", "http status", "status code", "contract test"}, "api contract check", "api"},
	{[]string{"graphql query", "graphql mutation", "graphql schema", "graphql resolver"}, "graphql operation check", "graphql"},
}

// failureNeedles signal a lesson worth remembering.
var failureNeedles = []string{"traceback", "exception", "failed", "failure", "error:", "panic", "regression", "flaky"}

// Learn reads normalized sessions, extracts skill + lesson candidates with bounded
// evidence, and merges any new ones (by id) into the candidate files as pending.
// Idempotent: an id already present anywhere in the state is skipped.
func (u UseCase) Learn(ctx context.Context) (LearnResult, error) {
	index, err := u.Reader.LoadSessionIndex(ctx)
	if err != nil {
		return LearnResult{}, err
	}
	state, err := loadState(ctx, u.Store, u.dir())
	if err != nil {
		return LearnResult{}, err
	}
	seen := state.idSet()
	res := LearnResult{}

	for _, entry := range index.Entries {
		body, err := u.Reader.ReadNormalizedSession(ctx, entry.NormalizedPath)
		if err != nil {
			return LearnResult{}, fmt.Errorf("read normalized session %q: %w", entry.NormalizedPath, err)
		}
		res.SessionsProcessed++
		text := collapse(body)
		lower := strings.ToLower(text)

		for _, item := range extractSkillCandidates(text, lower, entry.NormalizedPath) {
			if seen[item.ID] {
				continue
			}
			seen[item.ID] = true
			state.SkillCandidates = append(state.SkillCandidates, item)
			res.SkillsAdded++
		}
		for _, item := range extractLessons(text, lower, entry.NormalizedPath) {
			if seen[item.ID] {
				continue
			}
			seen[item.ID] = true
			state.Lessons = append(state.Lessons, item)
			res.LessonsAdded++
		}
	}

	if err := saveState(ctx, u.Store, u.dir(), state); err != nil {
		return LearnResult{}, err
	}
	return res, nil
}

// extractSkillCandidates returns one pending skill candidate per matched rule.
func extractSkillCandidates(text, lower, source string) []MemoryItem {
	var out []MemoryItem
	for _, rule := range skillRules {
		if needle, ok := firstMatch(lower, rule.needles); ok {
			out = append(out, MemoryItem{
				ID:       ItemID(KindSkillCandidates, rule.name, source),
				Name:     rule.name,
				Domain:   rule.domain,
				Evidence: snippet(text, lower, needle),
				Source:   source,
				Status:   StatusPending,
			})
		}
	}
	return out
}

// extractLessons returns at most one pending lesson per session (the first
// failure signal), so lessons stay high-signal.
func extractLessons(text, lower, source string) []MemoryItem {
	needle, ok := firstMatch(lower, failureNeedles)
	if !ok {
		return nil
	}
	const name = "lesson from failure signal"
	return []MemoryItem{{
		ID:       ItemID(KindLessons, name, source),
		Name:     name,
		Domain:   "bugs",
		Evidence: snippet(text, lower, needle),
		Source:   source,
		Status:   StatusPending,
	}}
}

// firstMatch returns the first needle contained in lower and whether any matched.
func firstMatch(lower string, needles []string) (string, bool) {
	for _, n := range needles {
		if strings.Contains(lower, n) {
			return n, true
		}
	}
	return "", false
}

// collapse turns all whitespace runs into single spaces, giving single-line
// evidence and stable YAML scalars.
func collapse(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// snippet returns a rune-bounded window of the (already collapsed) text centered
// on the needle. It never returns more than evidenceWindow runes, so a raw body
// is never copied whole.
func snippet(text, lower, needle string) string {
	idx := strings.Index(lower, needle)
	if idx < 0 {
		return boundRunes(text, evidenceWindow)
	}
	start := idx - 120
	if start < 0 {
		start = 0
	}
	end := start + evidenceWindow
	if end > len(text) {
		end = len(text)
	}
	for start > 0 && !utf8.RuneStart(text[start]) {
		start--
	}
	for end < len(text) && !utf8.RuneStart(text[end]) {
		end++
	}
	return strings.TrimSpace(text[start:end])
}

// boundRunes caps text at n runes on a rune boundary.
func boundRunes(text string, n int) string {
	if r := []rune(text); len(r) > n {
		return strings.TrimSpace(string(r[:n]))
	}
	return text
}
```

Note: `evidenceWindow` bounds the byte-window; the test asserts `len(evidence) <= evidenceWindow` where `len` is byte length, and the window end is a byte offset `start + evidenceWindow` (then snapped forward to a rune boundary, adding at most 3 bytes). For the ASCII fixture this holds exactly; the `boundRunes` fallback path caps by rune count. This is consistent with `core/knowledge`'s evidence bounding.

- [ ] **Step 4: Verify it compiles**

Run: `go build ./internal/core/memgov/`
Expected: no output (success).

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/core/memgov/ -run Learn -v`
Expected: PASS (`TestLearnExtractsPendingCandidates`, `TestLearnIsIdempotent`).

- [ ] **Step 6: Commit**

```bash
git add internal/core/memgov/learn.go internal/core/memgov/memgov_test.go
git commit -m "feat(memgov): learn extracts pending candidates (idempotent, bounded evidence)"
```

---

## Task 6: `Review` — list pending candidates

**Files:**
- Create: `internal/core/memgov/review.go`
- Test: `internal/core/memgov/memgov_test.go` (extend)

- [ ] **Step 1: Write the failing test**

Append to `internal/core/memgov/memgov_test.go`:

```go
func TestReviewListsOnlyPending(t *testing.T) {
	uc, _ := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	res, err := uc.Review(context.Background())
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if len(res.Pending) == 0 {
		t.Fatalf("expected pending candidates")
	}
	for _, it := range res.Pending {
		if it.Status != StatusPending {
			t.Fatalf("review returned non-pending item: %#v", it)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/memgov/ -run Review -v`
Expected: FAIL — `undefined: (UseCase).Review`.

- [ ] **Step 3: Create `review.go`**

Create `internal/core/memgov/review.go`:

```go
package memgov

import "context"

// Review returns the pending candidates (lessons first, then skills) for display.
func (u UseCase) Review(ctx context.Context) (ReviewResult, error) {
	state, err := loadState(ctx, u.Store, u.dir())
	if err != nil {
		return ReviewResult{}, err
	}
	var pending []MemoryItem
	pending = append(pending, filterPending(state.Lessons)...)
	pending = append(pending, filterPending(state.SkillCandidates)...)
	return ReviewResult{Pending: pending}, nil
}

// filterPending returns only items still in the pending state.
func filterPending(items []MemoryItem) []MemoryItem {
	var out []MemoryItem
	for _, it := range items {
		if it.Status == StatusPending {
			out = append(out, it)
		}
	}
	return out
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/memgov/ -run Review -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/memgov/review.go internal/core/memgov/memgov_test.go
git commit -m "feat(memgov): review lists pending candidates"
```

---

## Task 7: `Promote` / `Reject` — move by id + log

Find a pending candidate by id, move it to `approved_patterns`/`rejected_patterns` with the new status, remove it from candidates, append a `decision_log` entry. Unknown or already-decided id → clear error, nothing changed.

**Files:**
- Create: `internal/core/memgov/decide.go`
- Test: `internal/core/memgov/memgov_test.go` (extend)

- [ ] **Step 1: Write the failing test**

Append to `internal/core/memgov/memgov_test.go`:

```go
func firstPendingID(t *testing.T, uc UseCase) string {
	t.Helper()
	res, err := uc.Review(context.Background())
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if len(res.Pending) == 0 {
		t.Fatalf("no pending candidates to decide")
	}
	return res.Pending[0].ID
}

func TestPromoteMovesItemAndLogs(t *testing.T) {
	uc, store := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	id := firstPendingID(t, uc)

	res, err := uc.Promote(context.Background(), id)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if res.Action != "promoted" || res.Item.ID != id || res.Item.Status != StatusApproved {
		t.Fatalf("unexpected promote result: %#v", res)
	}
	approved := store.files[".bqa/memory/approved_patterns.yaml"]
	if !strings.Contains(approved, "id: "+id) || !strings.Contains(approved, "status: approved") {
		t.Fatalf("approved_patterns missing promoted item:\n%s", approved)
	}
	// Removed from its candidate list.
	for _, list := range []string{"lessons_learned.yaml", "skill_candidates.yaml"} {
		if strings.Contains(store.files[".bqa/memory/"+list], "id: "+id) {
			t.Fatalf("promoted id still present in %s", list)
		}
	}
	if !strings.Contains(store.files[".bqa/memory/decision_log.yaml"], "action: promoted") {
		t.Fatalf("decision_log missing promotion entry")
	}
}

func TestRejectMovesItemToRejected(t *testing.T) {
	uc, store := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	id := firstPendingID(t, uc)

	res, err := uc.Reject(context.Background(), id)
	if err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if res.Action != "rejected" || res.Item.Status != StatusRejected {
		t.Fatalf("unexpected reject result: %#v", res)
	}
	rejected := store.files[".bqa/memory/rejected_patterns.yaml"]
	if !strings.Contains(rejected, "id: "+id) || !strings.Contains(rejected, "status: rejected") {
		t.Fatalf("rejected_patterns missing item:\n%s", rejected)
	}
}

func TestDecideTwiceErrorsAndLeavesFilesUnchanged(t *testing.T) {
	uc, store := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	id := firstPendingID(t, uc)
	if _, err := uc.Promote(context.Background(), id); err != nil {
		t.Fatalf("first Promote: %v", err)
	}
	before := store.files[".bqa/memory/approved_patterns.yaml"]

	_, err := uc.Promote(context.Background(), id)
	if err == nil {
		t.Fatalf("expected error promoting an already-decided id")
	}
	if store.files[".bqa/memory/approved_patterns.yaml"] != before {
		t.Fatalf("approved_patterns changed after failed second promote")
	}
}

func TestPromoteUnknownIDErrors(t *testing.T) {
	uc, _ := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	if _, err := uc.Promote(context.Background(), "skill-doesnotexist"); err == nil {
		t.Fatalf("expected error for unknown id")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/memgov/ -run 'Promote|Reject|Decide' -v`
Expected: FAIL — `undefined: (UseCase).Promote` / `Reject`.

- [ ] **Step 3: Create `decide.go`**

Create `internal/core/memgov/decide.go`:

```go
package memgov

import (
	"context"
	"fmt"
)

// Promote moves a pending candidate to approved_patterns.
func (u UseCase) Promote(ctx context.Context, id string) (DecideResult, error) {
	return u.decide(ctx, id, StatusApproved, "promoted")
}

// Reject moves a pending candidate to rejected_patterns.
func (u UseCase) Reject(ctx context.Context, id string) (DecideResult, error) {
	return u.decide(ctx, id, StatusRejected, "rejected")
}

// decide is the shared promote/reject transition. It finds a pending candidate by
// id, sets its status, moves it to the approved/rejected list, removes it from its
// candidate list, and appends a decision_log entry. An unknown or already-decided
// id returns an error and writes nothing.
func (u UseCase) decide(ctx context.Context, id, newStatus, action string) (DecideResult, error) {
	state, err := loadState(ctx, u.Store, u.dir())
	if err != nil {
		return DecideResult{}, err
	}

	item, found := takePendingCandidate(&state, id)
	if !found {
		return DecideResult{}, fmt.Errorf("no pending candidate with id %q", id)
	}
	item.Status = newStatus
	if newStatus == StatusApproved {
		state.Approved = append(state.Approved, item)
	} else {
		state.Rejected = append(state.Rejected, item)
	}
	state.Log = append(state.Log, DecisionEntry{ID: item.ID, Action: action, Name: item.Name})

	if err := saveState(ctx, u.Store, u.dir(), state); err != nil {
		return DecideResult{}, err
	}
	return DecideResult{Item: item, Action: action}, nil
}

// takePendingCandidate removes and returns the pending candidate with the given
// id from whichever candidate list holds it. Reports false if no pending
// candidate matches.
func takePendingCandidate(state *GovernanceState, id string) (MemoryItem, bool) {
	if item, rest, ok := removePending(state.Lessons, id); ok {
		state.Lessons = rest
		return item, true
	}
	if item, rest, ok := removePending(state.SkillCandidates, id); ok {
		state.SkillCandidates = rest
		return item, true
	}
	return MemoryItem{}, false
}

// removePending returns the matching pending item, the list without it, and
// whether it was found.
func removePending(items []MemoryItem, id string) (MemoryItem, []MemoryItem, bool) {
	for i, it := range items {
		if it.ID == id && it.Status == StatusPending {
			rest := make([]MemoryItem, 0, len(items)-1)
			rest = append(rest, items[:i]...)
			rest = append(rest, items[i+1:]...)
			return it, rest, true
		}
	}
	return MemoryItem{}, nil, false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/core/memgov/ -v`
Expected: PASS (all memgov tests, including promote/reject/errors).

- [ ] **Step 5: Commit**

```bash
git add internal/core/memgov/decide.go internal/core/memgov/memgov_test.go
git commit -m "feat(memgov): promote/reject candidates by id with decision log"
```

---

## Task 8: fs adapter for `GovernanceStore`

Pure file I/O under the memory dir. Imports only `ports` + stdlib.

**Files:**
- Create: `internal/adapters/fs/governance_store.go`
- Test: `internal/adapters/fs/governance_store_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapters/fs/governance_store_test.go`:

```go
package fs

import (
	"context"
	"path/filepath"
	"testing"
)

func TestGovernanceStoreReadMissing(t *testing.T) {
	store := GovernanceStore{}
	_, exists, err := store.ReadFile(context.Background(), t.TempDir(), "skill_candidates.yaml")
	if err != nil {
		t.Fatalf("ReadFile on missing: %v", err)
	}
	if exists {
		t.Fatalf("expected exists=false for missing file")
	}
}

func TestGovernanceStoreWriteThenRead(t *testing.T) {
	dir := t.TempDir()
	store := GovernanceStore{}
	memoryDir := filepath.Join(dir, ".bqa", "memory")

	if err := store.WriteFile(context.Background(), memoryDir, "skill_candidates.yaml", "hello\n"); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	content, exists, err := store.ReadFile(context.Background(), memoryDir, "skill_candidates.yaml")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !exists || content != "hello\n" {
		t.Fatalf("round-trip mismatch: exists=%v content=%q", exists, content)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/fs/ -run GovernanceStore -v`
Expected: FAIL — `undefined: GovernanceStore`.

- [ ] **Step 3: Create `governance_store.go`**

Create `internal/adapters/fs/governance_store.go`:

```go
package fs

import (
	"context"
	"errors"
	"os"
	"path/filepath"
)

// GovernanceStore is the filesystem implementation of ports.GovernanceStore. It
// reads and writes governance memory files under a memory directory and performs
// no parsing (that lives in core/memgov).
type GovernanceStore struct{}

func (GovernanceStore) ReadFile(ctx context.Context, memoryDir string, name string) (string, bool, error) {
	select {
	case <-ctx.Done():
		return "", false, ctx.Err()
	default:
	}
	path := filepath.Join(governanceDir(memoryDir), filepath.Clean(name))
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}

func (GovernanceStore) WriteFile(ctx context.Context, memoryDir string, name string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path := filepath.Join(governanceDir(memoryDir), filepath.Clean(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

func governanceDir(memoryDir string) string {
	if memoryDir == "" {
		return ".bqa/memory"
	}
	return memoryDir
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/fs/ -run GovernanceStore -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/fs/governance_store.go internal/adapters/fs/governance_store_test.go
git commit -m "feat(fs): filesystem GovernanceStore adapter"
```

---

## Task 9: `bqa memory` command group + registration

Wire the fs adapters into `memgov.UseCase` and expose learn/review/promote/reject.

**Files:**
- Create: `internal/app/memory.go`
- Modify: `internal/app/root.go`
- Test: `internal/app/memory_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/app/memory_test.go`:

```go
package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// writeSessionFixture writes a normalized session + index.json under base so the
// memory command has something to learn from.
func writeSessionFixture(t *testing.T, base string) {
	t.Helper()
	normDir := filepath.Join(base, "normalized")
	if err := os.MkdirAll(normDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "ETL reconciliation compared source table vs target table row count; job failed with a traceback."
	if err := os.WriteFile(filepath.Join(normDir, "s1.md"), []byte(body), 0o600); err != nil {
		t.Fatalf("write session: %v", err)
	}
	index := ports.SessionIndex{Entries: []ports.SessionIndexEntry{
		{OriginalPath: "/s1.jsonl", NormalizedPath: "normalized/s1.md"},
	}}
	data, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("marshal index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "index.json"), data, 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}
}

func runMemory(t *testing.T, args ...string) string {
	t.Helper()
	cmd := memoryCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("memory %v: %v\noutput:\n%s", args, err, out.String())
	}
	return out.String()
}

func TestMemoryLearnReviewPromoteFlow(t *testing.T) {
	dir := t.TempDir()
	sessions := filepath.Join(dir, "sessions")
	memoryDir := filepath.Join(dir, "memory")
	writeSessionFixture(t, sessions)

	learnOut := runMemory(t, "learn", "--sessions", sessions, "--memory-dir", memoryDir)
	if !strings.Contains(learnOut, "Skill candidates added:") {
		t.Fatalf("learn output missing counts:\n%s", learnOut)
	}

	reviewOut := runMemory(t, "review", "--memory-dir", memoryDir)
	if !strings.Contains(reviewOut, "etl reconciliation check") {
		t.Fatalf("review output missing pending item:\n%s", reviewOut)
	}

	// Grab an id from the rendered file and promote it.
	content, err := os.ReadFile(filepath.Join(memoryDir, "skill_candidates.yaml"))
	if err != nil {
		t.Fatalf("read skill_candidates: %v", err)
	}
	id := firstIDInYAML(t, string(content))
	promoteOut := runMemory(t, "promote", id, "--memory-dir", memoryDir)
	if !strings.Contains(promoteOut, "promoted") {
		t.Fatalf("promote output missing confirmation:\n%s", promoteOut)
	}

	approved, err := os.ReadFile(filepath.Join(memoryDir, "approved_patterns.yaml"))
	if err != nil {
		t.Fatalf("read approved_patterns: %v", err)
	}
	if !strings.Contains(string(approved), "id: "+id) {
		t.Fatalf("promoted id not in approved_patterns:\n%s", approved)
	}
}

// firstIDInYAML returns the first "- id: <value>" from a rendered items file.
func firstIDInYAML(t *testing.T, content string) string {
	t.Helper()
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- id:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "- id:"))
		}
	}
	t.Fatalf("no id found in:\n%s", content)
	return ""
}
```

Note: the fixture writes the session under `<sessions>/normalized/s1.md` with the index's `NormalizedPath` set to `normalized/s1.md`. The reused `fsadapter.KnowledgeStore.ReadNormalizedSession` resolves that relative path under `SessionBaseDir`, so no absolute paths are needed.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/app/ -run Memory -v`
Expected: FAIL — `undefined: memoryCmd`.

- [ ] **Step 3: Create `memory.go`**

Create `internal/app/memory.go`:

```go
package app

import (
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/core/memgov"
	"github.com/spf13/cobra"
)

func memoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Govern candidate QA memory (learn/review/promote/reject)",
		Long: "The memory governance loop extracts candidate memory from normalized sessions,\n" +
			"lets a human review it, and promotes or rejects candidates by id. Nothing enters\n" +
			"stable memory (approved_patterns.yaml) without an explicit promote.",
	}

	var sessions string
	var memoryDir string

	useCase := func() memgov.UseCase {
		return memgov.UseCase{
			Reader:    fsadapter.KnowledgeStore{SessionBaseDir: sessions},
			Store:     fsadapter.GovernanceStore{},
			MemoryDir: memoryDir,
		}
	}

	learnCmd := &cobra.Command{
		Use:   "learn",
		Short: "Extract candidate memory from normalized sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := useCase().Learn(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Sessions processed: %d\n", res.SessionsProcessed)
			fmt.Fprintf(out, "Lessons added: %d\n", res.LessonsAdded)
			fmt.Fprintf(out, "Skill candidates added: %d\n", res.SkillsAdded)
			fmt.Fprintf(out, "Memory dir: %s\n", memoryDir)
			return nil
		},
	}
	learnCmd.Flags().StringVar(&sessions, "sessions", ".bqa/input/sessions", "session input directory")
	learnCmd.Flags().StringVar(&memoryDir, "memory-dir", memgov.DefaultMemoryDir, "governance memory directory")
	cmd.AddCommand(learnCmd)

	reviewCmd := &cobra.Command{
		Use:   "review",
		Short: "List pending candidates awaiting a decision",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := useCase().Review(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(res.Pending) == 0 {
				fmt.Fprintln(out, "No pending candidates.")
				return nil
			}
			fmt.Fprintf(out, "Pending candidates: %d\n", len(res.Pending))
			for _, it := range res.Pending {
				fmt.Fprintf(out, "  %s [%s] %s\n", it.ID, it.Domain, it.Name)
			}
			return nil
		},
	}
	reviewCmd.Flags().StringVar(&memoryDir, "memory-dir", memgov.DefaultMemoryDir, "governance memory directory")
	cmd.AddCommand(reviewCmd)

	promoteCmd := &cobra.Command{
		Use:   "promote <id>",
		Short: "Approve a pending candidate into approved_patterns.yaml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := useCase().Promote(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s): %s\n", res.Action, res.Item.ID, res.Item.Domain, res.Item.Name)
			return nil
		},
	}
	promoteCmd.Flags().StringVar(&memoryDir, "memory-dir", memgov.DefaultMemoryDir, "governance memory directory")
	cmd.AddCommand(promoteCmd)

	rejectCmd := &cobra.Command{
		Use:   "reject <id>",
		Short: "Reject a pending candidate into rejected_patterns.yaml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := useCase().Reject(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%s): %s\n", res.Action, res.Item.ID, res.Item.Domain, res.Item.Name)
			return nil
		},
	}
	rejectCmd.Flags().StringVar(&memoryDir, "memory-dir", memgov.DefaultMemoryDir, "governance memory directory")
	cmd.AddCommand(rejectCmd)

	return cmd
}
```

- [ ] **Step 4: Register the command in `root.go`**

In `internal/app/root.go`, add the registration alongside the others (after the `brainCmd()` line at `internal/app/root.go:25`):

```go
	rootCmd.AddCommand(memoryCmd())
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/app/ -run Memory -v`
Expected: PASS (`TestMemoryLearnReviewPromoteFlow`).

- [ ] **Step 6: Commit**

```bash
git add internal/app/memory.go internal/app/root.go internal/app/memory_test.go
git commit -m "feat(app): wire bqa memory learn/review/promote/reject"
```

---

## Task 10: Full suite + determinism check

**Files:** none (verification only)

- [ ] **Step 1: Run the whole Go suite**

Run: `go test ./...`
Expected: PASS (all packages).

- [ ] **Step 2: Verify deterministic output**

Run:
```bash
cd "$(mktemp -d)" && \
  mkdir -p s/normalized && \
  printf 'ETL reconciliation source table vs target table row count; failed with traceback.' > s/normalized/s1.md && \
  printf '{"entries":[{"normalized_path":"normalized/s1.md","original_path":"/s1.jsonl"}]}' > s/index.json && \
  go run github.com/mshegolev/bqa-os -h >/dev/null 2>&1 || true
```
Then, from the repo root, confirm two learn runs into the same dir are byte-identical:
```bash
cd /opt/develop/bqa-os && \
  TMP=$(mktemp -d) && \
  mkdir -p "$TMP/s/normalized" && \
  printf 'ETL reconciliation source table vs target table row count; failed with traceback.' > "$TMP/s/normalized/s1.md" && \
  printf '{"entries":[{"normalized_path":"normalized/s1.md","original_path":"/s1.jsonl"}]}' > "$TMP/s/index.json" && \
  go run . memory learn --sessions "$TMP/s" --memory-dir "$TMP/m" >/dev/null && \
  cp "$TMP/m/skill_candidates.yaml" "$TMP/first.yaml" && \
  go run . memory learn --sessions "$TMP/s" --memory-dir "$TMP/m" >/dev/null && \
  diff "$TMP/first.yaml" "$TMP/m/skill_candidates.yaml" && echo "DETERMINISTIC OK"
```
Expected: `DETERMINISTIC OK` (no diff), and the second run reports `Skill candidates added: 0` (idempotent).

- [ ] **Step 3: Final commit (if anything is uncommitted)**

```bash
git status --short
```
Expected: clean tree (all work already committed in prior tasks).

---

## Self-Review

**Spec coverage:**
- Command surface `bqa memory learn/review/promote/reject` with `--sessions`/`--memory-dir` → Task 9.
- Storage under `.bqa/memory/` as v1-schema YAML; promote/reject transition by id into `approved_patterns.yaml`/`rejected_patterns.yaml` + append `decision_log.yaml` → Tasks 3, 7, 8.
- Hexagonal layering: pure `core/memgov`, ports `NormalizedSessionReader` (reused) + `GovernanceStore` (new), fs adapter → Tasks 2–8. Deviation from the spec's `Load/Save` signature documented in the Architecture header (file-level port keeps parse/render in core and adapters ports-only).
- Data model: v1 envelope + `status`; `decision_log` entries `{id, action, name}`, no timestamp → Task 3.
- Learn idempotent + bounded evidence, no raw body → Task 5.
- Review lists pending → Task 6.
- Promote/Reject move-by-id, remove from candidates, append log; unknown/already-decided → clear error, nothing changed → Task 7.
- Testing (learn/review/promote/reject/errors/idempotent, deterministic, `go test ./...`) → Tasks 3–10.
- Constraints: stdlib + cobra only (no new imports); deterministic (`generated_by` from `version.Version`, no wall-clock) → throughout.

**Placeholder scan:** every code step contains complete code; commands have expected output. No TBD/TODO/"add error handling"-style placeholders remain.

**Type consistency:** `MemoryItem`/`DecisionEntry`/`GovernanceState` fields, `UseCase{Reader, Store, MemoryDir}`, kind constants (`KindSkillCandidates` etc.), status constants, `ItemID(kind,name,source)`, `loadState`/`saveState`, and result structs (`LearnResult.SkillsAdded/LessonsAdded/SessionsProcessed`, `DecideResult.Item/Action`) are used identically across tasks and the CLI output lines.
</content>
</invoke>
