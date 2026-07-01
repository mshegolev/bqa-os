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
