# Workspace Registry Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `bqa workspace init/add/list` command group backed by `.bqa/workspace.yaml`, standing up the persistent registry that later #67 slices (`task start`, Jira, worktrees, context) will read and write.

**Architecture:** Strict hexagonal. The workspace model + deterministic YAML render/parse + init/add/list use cases live in `internal/core/workspace` (ports + stdlib + shared `textutil`/`version` only). Two file/os-level ports — `WorkspaceStore` (content in/out) and `PathInspector` (`IsDir`/`IsGitRepo`) — are implemented by one fs adapter. The cobra command group wires the adapters into the use case, mirroring `internal/app/brain.go`.

**Tech Stack:** Go 1.22, stdlib + cobra only. Deterministic output (`generated_by` from `version.Version`, no wall-clock). v1 YAML envelope reused from `core/knowledge` / `core/memgov`. `textutil.QuoteYAML` / `textutil.UnquoteYAML` for scalars.

---

## File Structure

**New files:**
- `internal/ports/workspace.go` — `WorkspaceStore` + `PathInspector` port interfaces.
- `internal/core/workspace/models.go` — constants, `Project`, `Task`, `Workspace`, result structs, `generatedBy()`.
- `internal/core/workspace/render.go` — deterministic YAML render (`Render`).
- `internal/core/workspace/parse.go` — YAML parse (`Parse`) line scanner.
- `internal/core/workspace/usecase.go` — `UseCase.Init/Add/List`.
- `internal/core/workspace/workspace_test.go` — render/parse round-trip + use-case tests (fakes).
- `internal/adapters/fs/workspace_store.go` — fs `WorkspaceStore` + `PathInspector`.
- `internal/adapters/fs/workspace_store_test.go` — temp-dir round-trip + path probes.
- `internal/app/workspace.go` — `workspaceCmd()` (init/add/list).
- `internal/app/workspace_test.go` — command-tree integration test.

**Modified files:**
- `internal/app/root.go` — register `workspaceCmd()`.

**Rendering note:** all scalar fields (including `branch_role`) are written with `textutil.QuoteYAML` for uniformity and safety; `Parse` uses `textutil.UnquoteYAML`, which handles both quoted and bare scalars. This is a deliberate simplification of the spec's "branch_role bareword" micro-detail — behavior and on-disk readability are unchanged.

---

## Task 1: `WorkspaceStore` + `PathInspector` ports

**Files:**
- Create: `internal/ports/workspace.go`

- [ ] **Step 1: Create the port file**

Create `internal/ports/workspace.go`:

```go
package ports

import "context"

// WorkspaceStore reads and writes the workspace registry file under a base
// directory. It is file-level (content in/out) so all YAML parse/render logic
// lives in core/workspace and this adapter stays pure I/O.
type WorkspaceStore interface {
	// Exists reports whether a workspace file is present under baseDir.
	Exists(ctx context.Context, baseDir string) (bool, error)
	// Load returns the raw workspace file content.
	Load(ctx context.Context, baseDir string) (content string, err error)
	// Save writes the workspace file content under baseDir, creating baseDir.
	Save(ctx context.Context, baseDir string, content string) error
}

// PathInspector answers filesystem questions the workspace use case needs when
// registering a project path, kept behind a port so both branches (missing dir,
// non-git dir) are unit-testable with a fake.
type PathInspector interface {
	// IsDir reports whether path exists and is a directory.
	IsDir(path string) (bool, error)
	// IsGitRepo reports whether path is (or is inside) a git working tree.
	IsGitRepo(path string) (bool, error)
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/ports/`
Expected: no output (success).

- [ ] **Step 3: Commit**

```bash
git add internal/ports/workspace.go
git commit -m "feat(ports): add WorkspaceStore and PathInspector ports"
```

---

## Task 2: `workspace` model + deterministic render/parse

**Files:**
- Create: `internal/core/workspace/models.go`
- Create: `internal/core/workspace/render.go`
- Create: `internal/core/workspace/parse.go`
- Test: `internal/core/workspace/workspace_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/core/workspace/workspace_test.go`:

```go
package workspace

import (
	"strings"
	"testing"
)

func TestRenderParseRoundTrip(t *testing.T) {
	ws := Workspace{
		Name: `bigdata "test"`,
		Projects: []Project{
			{ID: "main", Path: "/repos/bigdata_testing", Repo: "bigdata_testing", ETL: "", BranchRole: "base"},
			{ID: "ns2", Path: "/repos/ns2", Repo: "bigdata_testing", ETL: "NS2", BranchRole: "feature"},
		},
		Tasks: []Task{
			{ID: "DATA-1", Jira: "DATA-1", Repo: "bigdata_testing", ETL: "NS2", Path: "/repos/DATA-1", Branch: "feature/DATA-1"},
		},
	}
	out := Render(ws)
	if !strings.Contains(out, "schema_version: 1\n") || !strings.Contains(out, "kind: workspace\n") {
		t.Fatalf("missing v1 envelope:\n%s", out)
	}
	got := Parse(out)
	if got.Name != ws.Name {
		t.Fatalf("name mismatch: got %q want %q", got.Name, ws.Name)
	}
	if len(got.Projects) != 2 || got.Projects[0] != ws.Projects[0] || got.Projects[1] != ws.Projects[1] {
		t.Fatalf("projects round-trip mismatch: %#v", got.Projects)
	}
	if len(got.Tasks) != 1 || got.Tasks[0] != ws.Tasks[0] {
		t.Fatalf("tasks round-trip mismatch: %#v", got.Tasks)
	}
}

func TestRenderEmptyWorkspace(t *testing.T) {
	out := Render(Workspace{Name: "empty"})
	if !strings.Contains(out, "projects: []\n") || !strings.Contains(out, "tasks: []\n") {
		t.Fatalf("expected empty projects/tasks lists:\n%s", out)
	}
	got := Parse(out)
	if got.Name != "empty" || len(got.Projects) != 0 || len(got.Tasks) != 0 {
		t.Fatalf("expected empty workspace named 'empty', got %#v", got)
	}
}

func TestGeneratedByDeterministic(t *testing.T) {
	// version.Version is "dev" in tests, so the stamp is stable.
	if !strings.Contains(Render(Workspace{Name: "x"}), "generated_by: bqa dev\n") {
		t.Fatalf("expected deterministic generated_by stamp")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/workspace/ -v`
Expected: FAIL — package does not compile (`undefined: Render`, `undefined: Workspace`, etc.).

- [ ] **Step 3: Create `models.go`**

Create `internal/core/workspace/models.go`:

```go
package workspace

import "github.com/mshegolev/bqa-os/internal/version"

// SchemaVersion is the v1 envelope version stamped on the workspace file.
const SchemaVersion = 1

// KindWorkspace is the envelope kind for the workspace registry file.
const KindWorkspace = "workspace"

// DefaultBranchRole is the branch_role assigned when --branch-role is omitted.
const DefaultBranchRole = "base"

// WorkspaceFileName is the registry filename under the base directory.
const WorkspaceFileName = "workspace.yaml"

// generatedBy is the provenance stamp: "bqa dev" in dev/test, "bqa vX.Y.Z" in a
// release build. Deterministic in tests.
func generatedBy() string { return "bqa " + version.Version }

// Project is a registered local repo/worktree root.
type Project struct {
	ID         string
	Path       string
	Repo       string
	ETL        string
	BranchRole string
}

// Task is a registered task worktree. Reserved for slice 2 (bqa task start);
// rendered/parsed now so the schema is stable.
type Task struct {
	ID     string
	Jira   string
	Repo   string
	ETL    string
	Path   string
	Branch string
}

// Workspace is the full registry: a name plus project and task lists.
type Workspace struct {
	Name     string
	Projects []Project
	Tasks    []Task
}

// InitResult reports a completed init.
type InitResult struct {
	Name string
	Path string
}

// AddResult reports an added project and an optional non-fatal warning.
type AddResult struct {
	Project Project
	Warning string
}

// ListResult holds the loaded workspace for display.
type ListResult struct {
	Workspace Workspace
}
```

- [ ] **Step 4: Create `render.go`**

Create `internal/core/workspace/render.go`:

```go
package workspace

import (
	"fmt"
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
)

// Render serializes a Workspace to deterministic v1-envelope YAML.
func Render(ws Workspace) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("schema_version: %d\nkind: %s\ngenerated_by: %s\n", SchemaVersion, KindWorkspace, generatedBy()))
	b.WriteString("name: " + textutil.QuoteYAML(ws.Name) + "\n")
	b.WriteString(renderProjects(ws.Projects))
	b.WriteString(renderTasks(ws.Tasks))
	return b.String()
}

func renderProjects(projects []Project) string {
	var b strings.Builder
	if len(projects) == 0 {
		b.WriteString("projects: []\n")
		return b.String()
	}
	b.WriteString("projects:\n")
	for _, p := range projects {
		b.WriteString("  - id: " + textutil.QuoteYAML(p.ID) + "\n")
		b.WriteString("    path: " + textutil.QuoteYAML(p.Path) + "\n")
		b.WriteString("    repo: " + textutil.QuoteYAML(p.Repo) + "\n")
		b.WriteString("    etl: " + textutil.QuoteYAML(p.ETL) + "\n")
		b.WriteString("    branch_role: " + textutil.QuoteYAML(p.BranchRole) + "\n")
	}
	return b.String()
}

func renderTasks(tasks []Task) string {
	var b strings.Builder
	if len(tasks) == 0 {
		b.WriteString("tasks: []\n")
		return b.String()
	}
	b.WriteString("tasks:\n")
	for _, t := range tasks {
		b.WriteString("  - id: " + textutil.QuoteYAML(t.ID) + "\n")
		b.WriteString("    jira: " + textutil.QuoteYAML(t.Jira) + "\n")
		b.WriteString("    repo: " + textutil.QuoteYAML(t.Repo) + "\n")
		b.WriteString("    etl: " + textutil.QuoteYAML(t.ETL) + "\n")
		b.WriteString("    path: " + textutil.QuoteYAML(t.Path) + "\n")
		b.WriteString("    branch: " + textutil.QuoteYAML(t.Branch) + "\n")
	}
	return b.String()
}
```

- [ ] **Step 5: Create `parse.go`**

Create `internal/core/workspace/parse.go`:

```go
package workspace

import (
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
)

// Parse reads a workspace file produced by Render. Envelope/header lines and
// unknown lines are ignored. "projects:" / "tasks:" switch the active section;
// each "- id:" starts a new entry in that section.
func Parse(content string) Workspace {
	var ws Workspace
	section := ""
	var curP *Project
	var curT *Task
	flushP := func() {
		if curP != nil {
			ws.Projects = append(ws.Projects, *curP)
			curP = nil
		}
	}
	flushT := func() {
		if curT != nil {
			ws.Tasks = append(ws.Tasks, *curT)
			curT = nil
		}
	}

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case strings.HasPrefix(line, "name:"):
			ws.Name = textutil.UnquoteYAML(strings.TrimPrefix(line, "name:"))
		case strings.HasPrefix(line, "projects:"):
			flushP()
			flushT()
			section = "projects"
		case strings.HasPrefix(line, "tasks:"):
			flushP()
			flushT()
			section = "tasks"
		case strings.HasPrefix(line, "- id:"):
			id := textutil.UnquoteYAML(strings.TrimPrefix(line, "- id:"))
			if section == "tasks" {
				flushT()
				curT = &Task{ID: id}
			} else {
				flushP()
				curP = &Project{ID: id}
			}
		default:
			if section == "tasks" && curT != nil {
				assignTaskField(curT, line)
			} else if section == "projects" && curP != nil {
				assignProjectField(curP, line)
			}
		}
	}
	flushP()
	flushT()
	return ws
}

func assignProjectField(p *Project, line string) {
	switch {
	case strings.HasPrefix(line, "path:"):
		p.Path = textutil.UnquoteYAML(strings.TrimPrefix(line, "path:"))
	case strings.HasPrefix(line, "repo:"):
		p.Repo = textutil.UnquoteYAML(strings.TrimPrefix(line, "repo:"))
	case strings.HasPrefix(line, "etl:"):
		p.ETL = textutil.UnquoteYAML(strings.TrimPrefix(line, "etl:"))
	case strings.HasPrefix(line, "branch_role:"):
		p.BranchRole = textutil.UnquoteYAML(strings.TrimPrefix(line, "branch_role:"))
	}
}

func assignTaskField(t *Task, line string) {
	switch {
	case strings.HasPrefix(line, "jira:"):
		t.Jira = textutil.UnquoteYAML(strings.TrimPrefix(line, "jira:"))
	case strings.HasPrefix(line, "repo:"):
		t.Repo = textutil.UnquoteYAML(strings.TrimPrefix(line, "repo:"))
	case strings.HasPrefix(line, "etl:"):
		t.ETL = textutil.UnquoteYAML(strings.TrimPrefix(line, "etl:"))
	case strings.HasPrefix(line, "path:"):
		t.Path = textutil.UnquoteYAML(strings.TrimPrefix(line, "path:"))
	case strings.HasPrefix(line, "branch:"):
		t.Branch = textutil.UnquoteYAML(strings.TrimPrefix(line, "branch:"))
	}
}
```

Note: `assignTaskField` checks `branch:` (the task field). Project uses `branch_role:`. Both are matched by their own field assigner, so there is no cross-section collision.

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/core/workspace/ -v`
Expected: PASS (`TestRenderParseRoundTrip`, `TestRenderEmptyWorkspace`, `TestGeneratedByDeterministic`).

- [ ] **Step 7: Commit**

```bash
git add internal/core/workspace/models.go internal/core/workspace/render.go internal/core/workspace/parse.go internal/core/workspace/workspace_test.go
git commit -m "feat(workspace): model plus deterministic YAML render/parse"
```

---

## Task 3: `UseCase.Init/Add/List`

**Files:**
- Create: `internal/core/workspace/usecase.go`
- Test: `internal/core/workspace/workspace_test.go` (extend — append, keep existing tests)

- [ ] **Step 1: Write the failing test**

Append to `internal/core/workspace/workspace_test.go`. Also add `"context"` to the test file's import block.

```go
// memStore is an in-memory WorkspaceStore keyed by baseDir.
type memStore struct{ files map[string]string }

func newMemStore() *memStore { return &memStore{files: map[string]string{}} }

func (m *memStore) Exists(_ context.Context, baseDir string) (bool, error) {
	_, ok := m.files[baseDir]
	return ok, nil
}

func (m *memStore) Load(_ context.Context, baseDir string) (string, error) {
	return m.files[baseDir], nil
}

func (m *memStore) Save(_ context.Context, baseDir, content string) error {
	m.files[baseDir] = content
	return nil
}

// fakeInspector reports fixed answers for IsDir/IsGitRepo.
type fakeInspector struct {
	dir bool
	git bool
}

func (f fakeInspector) IsDir(string) (bool, error)     { return f.dir, nil }
func (f fakeInspector) IsGitRepo(string) (bool, error) { return f.git, nil }

func newUC(store *memStore, insp fakeInspector) UseCase {
	return UseCase{Store: store, Inspector: insp, BaseDir: ".bqa"}
}

func TestInitWritesEmptyWorkspace(t *testing.T) {
	store := newMemStore()
	uc := newUC(store, fakeInspector{dir: true, git: true})
	res, err := uc.Init(context.Background(), "bigdata")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if res.Name != "bigdata" {
		t.Fatalf("unexpected name: %q", res.Name)
	}
	content := store.files[".bqa"]
	if !strings.Contains(content, "kind: workspace\n") || !strings.Contains(content, "projects: []\n") {
		t.Fatalf("workspace not initialized as empty v1:\n%s", content)
	}
}

func TestInitTwiceErrors(t *testing.T) {
	store := newMemStore()
	uc := newUC(store, fakeInspector{dir: true, git: true})
	if _, err := uc.Init(context.Background(), "x"); err != nil {
		t.Fatalf("first Init: %v", err)
	}
	before := store.files[".bqa"]
	if _, err := uc.Init(context.Background(), "y"); err == nil {
		t.Fatalf("expected error re-initializing")
	}
	if store.files[".bqa"] != before {
		t.Fatalf("workspace overwritten by second init")
	}
}

func TestAddAppendsProject(t *testing.T) {
	store := newMemStore()
	uc := newUC(store, fakeInspector{dir: true, git: true})
	if _, err := uc.Init(context.Background(), "w"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	res, err := uc.Add(context.Background(), Project{ID: "main", Path: "/repos/bt", Repo: "bt", BranchRole: "base"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if res.Warning != "" {
		t.Fatalf("unexpected warning for git repo: %q", res.Warning)
	}
	got := Parse(store.files[".bqa"])
	if len(got.Projects) != 1 || got.Projects[0].ID != "main" {
		t.Fatalf("project not appended: %#v", got.Projects)
	}
}

func TestAddNonGitWarnsButRecords(t *testing.T) {
	store := newMemStore()
	uc := newUC(store, fakeInspector{dir: true, git: false})
	if _, err := uc.Init(context.Background(), "w"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	res, err := uc.Add(context.Background(), Project{ID: "main", Path: "/repos/bt", Repo: "bt", BranchRole: "base"})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if res.Warning == "" {
		t.Fatalf("expected non-git warning")
	}
	if len(Parse(store.files[".bqa"]).Projects) != 1 {
		t.Fatalf("non-git project should still be recorded")
	}
}

func TestAddMissingDirErrors(t *testing.T) {
	store := newMemStore()
	uc := newUC(store, fakeInspector{dir: false, git: false})
	if _, err := uc.Init(context.Background(), "w"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	before := store.files[".bqa"]
	if _, err := uc.Add(context.Background(), Project{ID: "main", Path: "/nope", Repo: "bt"}); err == nil {
		t.Fatalf("expected error for missing dir")
	}
	if store.files[".bqa"] != before {
		t.Fatalf("workspace changed after failed add")
	}
}

func TestAddDuplicateIDErrors(t *testing.T) {
	store := newMemStore()
	uc := newUC(store, fakeInspector{dir: true, git: true})
	if _, err := uc.Init(context.Background(), "w"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	if _, err := uc.Add(context.Background(), Project{ID: "main", Path: "/a", Repo: "bt"}); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	before := store.files[".bqa"]
	if _, err := uc.Add(context.Background(), Project{ID: "main", Path: "/b", Repo: "bt"}); err == nil {
		t.Fatalf("expected duplicate id error")
	}
	if store.files[".bqa"] != before {
		t.Fatalf("workspace changed after failed duplicate add")
	}
}

func TestAddBeforeInitErrors(t *testing.T) {
	uc := newUC(newMemStore(), fakeInspector{dir: true, git: true})
	if _, err := uc.Add(context.Background(), Project{ID: "main", Path: "/a", Repo: "bt"}); err == nil {
		t.Fatalf("expected not-initialized error")
	}
}

func TestListReturnsProjectsInOrder(t *testing.T) {
	store := newMemStore()
	uc := newUC(store, fakeInspector{dir: true, git: true})
	if _, err := uc.Init(context.Background(), "w"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	for _, id := range []string{"a", "b", "c"} {
		if _, err := uc.Add(context.Background(), Project{ID: id, Path: "/" + id, Repo: "bt"}); err != nil {
			t.Fatalf("Add %s: %v", id, err)
		}
	}
	res, err := uc.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	ids := []string{}
	for _, p := range res.Workspace.Projects {
		ids = append(ids, p.ID)
	}
	if strings.Join(ids, ",") != "a,b,c" {
		t.Fatalf("expected insertion order a,b,c, got %v", ids)
	}
}

func TestListBeforeInitErrors(t *testing.T) {
	uc := newUC(newMemStore(), fakeInspector{dir: true, git: true})
	if _, err := uc.List(context.Background()); err == nil {
		t.Fatalf("expected not-initialized error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/workspace/ -run 'Init|Add|List' -v`
Expected: FAIL — `undefined: (UseCase)` / `undefined: UseCase`.

- [ ] **Step 3: Create `usecase.go`**

Create `internal/core/workspace/usecase.go`:

```go
package workspace

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// UseCase manages the workspace registry: initialize it, register projects, and
// list them. All side effects go through the WorkspaceStore and PathInspector
// ports.
type UseCase struct {
	Store     ports.WorkspaceStore
	Inspector ports.PathInspector
	BaseDir   string
}

func (u UseCase) baseDir() string {
	if u.BaseDir == "" {
		return ".bqa"
	}
	return u.BaseDir
}

// Init creates an empty workspace registry. It errors if one already exists,
// writing nothing.
func (u UseCase) Init(ctx context.Context, name string) (InitResult, error) {
	exists, err := u.Store.Exists(ctx, u.baseDir())
	if err != nil {
		return InitResult{}, err
	}
	if exists {
		return InitResult{}, fmt.Errorf("workspace already initialized in %s", u.baseDir())
	}
	if err := u.Store.Save(ctx, u.baseDir(), Render(Workspace{Name: name})); err != nil {
		return InitResult{}, err
	}
	return InitResult{Name: name, Path: filepath.Join(u.baseDir(), WorkspaceFileName)}, nil
}

// Add registers a project. It requires an initialized workspace and an existing
// directory path; a non-git path is recorded with a warning; a duplicate id
// errors and changes nothing.
func (u UseCase) Add(ctx context.Context, p Project) (AddResult, error) {
	exists, err := u.Store.Exists(ctx, u.baseDir())
	if err != nil {
		return AddResult{}, err
	}
	if !exists {
		return AddResult{}, fmt.Errorf("no workspace in %s; run 'bqa workspace init'", u.baseDir())
	}

	isDir, err := u.Inspector.IsDir(p.Path)
	if err != nil {
		return AddResult{}, err
	}
	if !isDir {
		return AddResult{}, fmt.Errorf("path %q is not an existing directory", p.Path)
	}

	content, err := u.Store.Load(ctx, u.baseDir())
	if err != nil {
		return AddResult{}, err
	}
	ws := Parse(content)
	for _, existing := range ws.Projects {
		if existing.ID == p.ID {
			return AddResult{}, fmt.Errorf("project id %q already exists", p.ID)
		}
	}

	var warning string
	isGit, err := u.Inspector.IsGitRepo(p.Path)
	if err != nil {
		return AddResult{}, err
	}
	if !isGit {
		warning = fmt.Sprintf("path %q is not a git repository; recorded anyway", p.Path)
	}

	ws.Projects = append(ws.Projects, p)
	if err := u.Store.Save(ctx, u.baseDir(), Render(ws)); err != nil {
		return AddResult{}, err
	}
	return AddResult{Project: p, Warning: warning}, nil
}

// List returns the registered projects. It requires an initialized workspace.
func (u UseCase) List(ctx context.Context) (ListResult, error) {
	exists, err := u.Store.Exists(ctx, u.baseDir())
	if err != nil {
		return ListResult{}, err
	}
	if !exists {
		return ListResult{}, fmt.Errorf("no workspace in %s; run 'bqa workspace init'", u.baseDir())
	}
	content, err := u.Store.Load(ctx, u.baseDir())
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Workspace: Parse(content)}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/core/workspace/ -v`
Expected: PASS (all workspace tests, including init/add/list/errors).

- [ ] **Step 5: Commit**

```bash
git add internal/core/workspace/usecase.go internal/core/workspace/workspace_test.go
git commit -m "feat(workspace): init/add/list use cases"
```

---

## Task 4: fs adapter — `WorkspaceStore` + `PathInspector`

**Files:**
- Create: `internal/adapters/fs/workspace_store.go`
- Test: `internal/adapters/fs/workspace_store_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapters/fs/workspace_store_test.go`:

```go
package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkspaceStoreExistsLoadSave(t *testing.T) {
	base := filepath.Join(t.TempDir(), ".bqa")
	store := WorkspaceStore{}

	exists, err := store.Exists(context.Background(), base)
	if err != nil || exists {
		t.Fatalf("expected not-exists, got exists=%v err=%v", exists, err)
	}
	if err := store.Save(context.Background(), base, "hello\n"); err != nil {
		t.Fatalf("Save: %v", err)
	}
	exists, err = store.Exists(context.Background(), base)
	if err != nil || !exists {
		t.Fatalf("expected exists after save, got exists=%v err=%v", exists, err)
	}
	content, err := store.Load(context.Background(), base)
	if err != nil || content != "hello\n" {
		t.Fatalf("Load mismatch: content=%q err=%v", content, err)
	}
}

func TestPathInspectorIsDir(t *testing.T) {
	dir := t.TempDir()
	insp := PathInspector{}

	ok, err := insp.IsDir(dir)
	if err != nil || !ok {
		t.Fatalf("expected dir true, got %v err=%v", ok, err)
	}
	file := filepath.Join(dir, "f")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if ok, _ := insp.IsDir(file); ok {
		t.Fatalf("a regular file should not be a dir")
	}
	if ok, _ := insp.IsDir(filepath.Join(dir, "missing")); ok {
		t.Fatalf("missing path should not be a dir")
	}
}

func TestPathInspectorIsGitRepo(t *testing.T) {
	insp := PathInspector{}
	plain := t.TempDir()
	if ok, _ := insp.IsGitRepo(plain); ok {
		t.Fatalf("plain dir should not be a git repo")
	}

	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if ok, err := insp.IsGitRepo(repo); err != nil || !ok {
		t.Fatalf("expected git repo true, got %v err=%v", ok, err)
	}
	// A subdirectory of the repo is inside the working tree.
	sub := filepath.Join(repo, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}
	if ok, err := insp.IsGitRepo(sub); err != nil || !ok {
		t.Fatalf("expected subdir of repo to be inside git tree, got %v err=%v", ok, err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/fs/ -run 'Workspace|PathInspector' -v`
Expected: FAIL — `undefined: WorkspaceStore` / `undefined: PathInspector`.

- [ ] **Step 3: Create `workspace_store.go`**

Create `internal/adapters/fs/workspace_store.go`:

```go
package fs

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// Compile-time checks that the fs types satisfy the workspace ports.
var (
	_ ports.WorkspaceStore = WorkspaceStore{}
	_ ports.PathInspector  = PathInspector{}
)

// WorkspaceStore is the filesystem implementation of ports.WorkspaceStore. It
// reads/writes <baseDir>/workspace.yaml and performs no parsing.
type WorkspaceStore struct{}

func workspaceFilePath(baseDir string) string {
	if baseDir == "" {
		baseDir = ".bqa"
	}
	return filepath.Join(baseDir, "workspace.yaml")
}

func (WorkspaceStore) Exists(ctx context.Context, baseDir string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	_, err := os.Stat(workspaceFilePath(baseDir))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (WorkspaceStore) Load(ctx context.Context, baseDir string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	data, err := os.ReadFile(workspaceFilePath(baseDir))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (WorkspaceStore) Save(ctx context.Context, baseDir string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path := workspaceFilePath(baseDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

// PathInspector is the filesystem implementation of ports.PathInspector.
type PathInspector struct{}

func (PathInspector) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// IsGitRepo walks up from path looking for a `.git` entry (a directory for a
// normal repo, or a file for a linked worktree), so any location inside a
// working tree reports true.
func (PathInspector) IsGitRepo(path string) (bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	for {
		_, err := os.Stat(filepath.Join(abs, ".git"))
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return false, nil
		}
		abs = parent
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/adapters/fs/ -run 'Workspace|PathInspector' -v`
Expected: PASS.

Also run the whole fs package: `go test ./internal/adapters/fs/ -v`

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/fs/workspace_store.go internal/adapters/fs/workspace_store_test.go
git commit -m "feat(fs): filesystem WorkspaceStore and PathInspector adapters"
```

---

## Task 5: `bqa workspace` command group + registration

**Files:**
- Create: `internal/app/workspace.go`
- Modify: `internal/app/root.go`
- Test: `internal/app/workspace_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/app/workspace_test.go`:

```go
package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runWorkspace(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := workspaceCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func TestWorkspaceInitAddListFlow(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, ".bqa")
	repo := filepath.Join(dir, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo/.git: %v", err)
	}

	if _, err := runWorkspace(t, "init", "--name", "bigdata", "--base-dir", base); err != nil {
		t.Fatalf("init: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(base, "workspace.yaml"))
	if err != nil {
		t.Fatalf("read workspace.yaml: %v", err)
	}
	if !strings.Contains(string(content), "name: \"bigdata\"") {
		t.Fatalf("workspace.yaml missing name:\n%s", content)
	}

	addOut, err := runWorkspace(t, "add", "main", repo, "--repo", "bigdata_testing", "--etl", "NS2", "--base-dir", base)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(addOut, "main") {
		t.Fatalf("add output missing project id:\n%s", addOut)
	}

	listOut, err := runWorkspace(t, "list", "--base-dir", base)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(listOut, "Workspace: bigdata") || !strings.Contains(listOut, "main") || !strings.Contains(listOut, "NS2") {
		t.Fatalf("list output missing expected content:\n%s", listOut)
	}
}

func TestWorkspaceAddNonGitWarns(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, ".bqa")
	plain := filepath.Join(dir, "plain")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatalf("mkdir plain: %v", err)
	}
	if _, err := runWorkspace(t, "init", "--name", "w", "--base-dir", base); err != nil {
		t.Fatalf("init: %v", err)
	}
	out, err := runWorkspace(t, "add", "main", plain, "--repo", "bt", "--base-dir", base)
	if err != nil {
		t.Fatalf("add: %v", err)
	}
	if !strings.Contains(out, "warning") {
		t.Fatalf("expected non-git warning in output:\n%s", out)
	}
}

func TestWorkspaceListBeforeInitErrors(t *testing.T) {
	base := filepath.Join(t.TempDir(), ".bqa")
	if _, err := runWorkspace(t, "list", "--base-dir", base); err == nil {
		t.Fatalf("expected error listing before init")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/app/ -run Workspace -v`
Expected: FAIL — `undefined: workspaceCmd`.

- [ ] **Step 3: Create `workspace.go`**

Create `internal/app/workspace.go`:

```go
package app

import (
	"fmt"
	"os"
	"path/filepath"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/core/workspace"
	"github.com/spf13/cobra"
)

func workspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage the BQA workspace registry (projects for task work)",
		Long: "The workspace registry records local repo/worktree roots in .bqa/workspace.yaml.\n" +
			"Later commands (bqa task start) create task worktrees against these projects.\n" +
			"Registering a project never copies or modifies the target repository.",
	}

	newUseCase := func(baseDir string) workspace.UseCase {
		return workspace.UseCase{
			Store:     fsadapter.WorkspaceStore{},
			Inspector: fsadapter.PathInspector{},
			BaseDir:   baseDir,
		}
	}

	var initName, initBaseDir string
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a BQA workspace registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := initName
			if name == "" {
				wd, err := os.Getwd()
				if err != nil {
					return err
				}
				name = filepath.Base(wd)
			}
			res, err := newUseCase(initBaseDir).Init(cmd.Context(), name)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized workspace %q at %s\n", res.Name, res.Path)
			return nil
		},
	}
	initCmd.Flags().StringVar(&initName, "name", "", "workspace name (default: current directory name)")
	initCmd.Flags().StringVar(&initBaseDir, "base-dir", ".bqa", "BQA base directory")
	cmd.AddCommand(initCmd)

	var addRepo, addETL, addBranchRole, addBaseDir string
	addCmd := &cobra.Command{
		Use:   "add <project-id> <path>",
		Short: "Register a local repo/worktree root (does not copy it)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := newUseCase(addBaseDir).Add(cmd.Context(), workspace.Project{
				ID: args[0], Path: args[1], Repo: addRepo, ETL: addETL, BranchRole: addBranchRole,
			})
			if err != nil {
				return err
			}
			if res.Warning != "" {
				fmt.Fprintln(cmd.ErrOrStderr(), "warning: "+res.Warning)
			}
			p := res.Project
			fmt.Fprintf(cmd.OutOrStdout(), "Added project %s (repo=%s etl=%s branch_role=%s): %s\n", p.ID, p.Repo, p.ETL, p.BranchRole, p.Path)
			return nil
		},
	}
	addCmd.Flags().StringVar(&addRepo, "repo", "", "repository name (required)")
	addCmd.Flags().StringVar(&addETL, "etl", "", "ETL pipeline this project targets")
	addCmd.Flags().StringVar(&addBranchRole, "branch-role", workspace.DefaultBranchRole, "branch role for this project")
	addCmd.Flags().StringVar(&addBaseDir, "base-dir", ".bqa", "BQA base directory")
	_ = addCmd.MarkFlagRequired("repo")
	cmd.AddCommand(addCmd)

	var listBaseDir string
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List registered projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := newUseCase(listBaseDir).List(cmd.Context())
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			ws := res.Workspace
			fmt.Fprintf(out, "Workspace: %s\n", ws.Name)
			if len(ws.Projects) == 0 {
				fmt.Fprintln(out, "No projects registered.")
			} else {
				fmt.Fprintf(out, "Projects: %d\n", len(ws.Projects))
				for _, p := range ws.Projects {
					fmt.Fprintf(out, "  %s  repo=%s  etl=%s  branch_role=%s  %s\n", p.ID, p.Repo, p.ETL, p.BranchRole, p.Path)
				}
			}
			fmt.Fprintf(out, "Tasks: %d\n", len(ws.Tasks))
			return nil
		},
	}
	listCmd.Flags().StringVar(&listBaseDir, "base-dir", ".bqa", "BQA base directory")
	cmd.AddCommand(listCmd)

	return cmd
}
```

- [ ] **Step 4: Register the command in `root.go`**

In `internal/app/root.go`, add the registration alongside the others (after the `rootCmd.AddCommand(memoryCmd())` line):

```go
	rootCmd.AddCommand(workspaceCmd())
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./internal/app/ -run Workspace -v`
Expected: PASS (`TestWorkspaceInitAddListFlow`, `TestWorkspaceAddNonGitWarns`, `TestWorkspaceListBeforeInitErrors`).

Also run the whole app package: `go test ./internal/app/ -v` and `go build ./...`.

- [ ] **Step 6: Commit**

```bash
git add internal/app/workspace.go internal/app/root.go internal/app/workspace_test.go
git commit -m "feat(app): wire bqa workspace init/add/list"
```

---

## Task 6: Full suite + determinism

**Files:** none (verification only)

- [ ] **Step 1: Run the whole Go suite**

Run: `go test ./...`
Expected: PASS (all packages).

- [ ] **Step 2: Verify deterministic output**

Run from the repo root:
```bash
cd /opt/develop/bqa-os && \
  TMP=$(mktemp -d) && \
  mkdir -p "$TMP/repo/.git" && \
  go run ./cmd/bqa workspace init --name demo --base-dir "$TMP/.bqa" >/dev/null && \
  go run ./cmd/bqa workspace add main "$TMP/repo" --repo bigdata_testing --etl NS2 --base-dir "$TMP/.bqa" >/dev/null && \
  cp "$TMP/.bqa/workspace.yaml" "$TMP/first.yaml" && \
  go run ./cmd/bqa workspace list --base-dir "$TMP/.bqa" && \
  diff <(go run ./cmd/bqa workspace init --name demo --base-dir "$TMP/.bqa2" 2>/dev/null; cat "$TMP/.bqa2/workspace.yaml") \
       <(printf 'schema_version: 1\nkind: workspace\ngenerated_by: bqa dev\nname: "demo"\nprojects: []\ntasks: []\n') \
    && echo "INIT DETERMINISTIC OK"
```
Expected: the `list` output shows `Workspace: demo`, one project `main` with `etl=NS2`; the second `diff` prints nothing and `INIT DETERMINISTIC OK` is printed (a fresh init produces exactly the canonical empty v1 file). Note `go run` builds with `version.Version == "dev"`, so `generated_by: bqa dev` is stable.

- [ ] **Step 3: Confirm clean tree**

```bash
git status --short
```
Expected: no uncommitted source changes (all work committed in prior tasks).

---

## Self-Review

**Spec coverage:**
- Command surface `bqa workspace init/add/list` with `--name`/`--base-dir`/`--repo`/`--etl`/`--branch-role` → Task 5.
- `.bqa/workspace.yaml` v1 envelope + `name`/`projects`/`tasks` model → Tasks 2, 4.
- `path` stored verbatim (absolute allowed) → `Project.Path` passed through unmodified (Tasks 2, 3, 5).
- `add` validation: missing/non-dir → error, nothing written; non-git dir → recorded with warning; duplicate id → error, nothing changed → Task 3 (tests) + Task 5 (CLI warning to stderr).
- One workspace per `.bqa`, optional `name` (default = cwd base name) → Task 5 init.
- Hexagonal layering: pure `core/workspace`, ports `WorkspaceStore` + `PathInspector`, one fs adapter → Tasks 1–4.
- Deterministic (`generated_by` from `version.Version`, stable field order) → Task 2 + Task 6 determinism check.
- `tasks: []` reserved and schema-stable → Task 2 renders/parses tasks (round-trip test includes a task entry).
- Constraints: stdlib + cobra only; never copies/mutates target repos (add only records + probes) → throughout.

**Placeholder scan:** every code step contains complete code; commands have expected output. No TBD/TODO/"add validation"-style placeholders.

**Type consistency:** `Workspace{Name,Projects,Tasks}`, `Project{ID,Path,Repo,ETL,BranchRole}`, `Task{ID,Jira,Repo,ETL,Path,Branch}`, `UseCase{Store,Inspector,BaseDir}`, result structs (`InitResult{Name,Path}`, `AddResult{Project,Warning}`, `ListResult{Workspace}`), `Render`/`Parse`, and the port method sets (`Exists/Load/Save`, `IsDir/IsGitRepo`) are used identically across tasks and the CLI. Rendering quotes all scalars uniformly and `Parse`/`UnquoteYAML` reads both quoted and bare forms.
</content>
