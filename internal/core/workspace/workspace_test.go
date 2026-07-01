package workspace

import (
	"context"
	"errors"
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

// errInspector returns a configured error from the named method.
type errInspector struct {
	failDir bool
	failGit bool
}

func (e errInspector) IsDir(string) (bool, error) {
	if e.failDir {
		return false, errors.New("isdir boom")
	}
	return true, nil
}

func (e errInspector) IsGitRepo(string) (bool, error) {
	if e.failGit {
		return false, errors.New("isgit boom")
	}
	return true, nil
}

func TestAddPropagatesIsDirError(t *testing.T) {
	store := newMemStore()
	uc := UseCase{Store: store, Inspector: errInspector{failDir: true}, BaseDir: ".bqa"}
	if _, err := uc.Init(context.Background(), "w"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	before := store.files[".bqa"]
	if _, err := uc.Add(context.Background(), Project{ID: "main", Path: "/a", Repo: "bt"}); err == nil {
		t.Fatalf("expected IsDir error to propagate")
	}
	if store.files[".bqa"] != before {
		t.Fatalf("workspace changed after IsDir error")
	}
}

func TestAddPropagatesIsGitRepoError(t *testing.T) {
	store := newMemStore()
	uc := UseCase{Store: store, Inspector: errInspector{failGit: true}, BaseDir: ".bqa"}
	if _, err := uc.Init(context.Background(), "w"); err != nil {
		t.Fatalf("Init: %v", err)
	}
	before := store.files[".bqa"]
	if _, err := uc.Add(context.Background(), Project{ID: "main", Path: "/a", Repo: "bt"}); err == nil {
		t.Fatalf("expected IsGitRepo error to propagate")
	}
	if store.files[".bqa"] != before {
		t.Fatalf("workspace changed after IsGitRepo error")
	}
}
