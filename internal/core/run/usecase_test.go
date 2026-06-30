package run

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPlan(t *testing.T) {
	base := t.TempDir()
	mustWrite(t, filepath.Join(base, "agents", "etl-qa-agent.md"))
	mustWrite(t, filepath.Join(base, "skills", "log-investigation.md"))
	mustWrite(t, filepath.Join(base, "workflows", "verify.md"))

	plan := Build(base, "Test VSR")
	if plan.Task != "Test VSR" {
		t.Fatalf("task not preserved: %q", plan.Task)
	}
	if len(plan.Agents) != 1 || plan.Agents[0] != "etl-qa-agent" {
		t.Fatalf("agents wrong: %v", plan.Agents)
	}
	if len(plan.Skills) != 1 || len(plan.Workflows) != 1 {
		t.Fatalf("skills/workflows wrong: %v %v", plan.Skills, plan.Workflows)
	}
	if plan.Empty() {
		t.Fatal("plan should not be empty")
	}
}

func TestBuildPlanEmptyWorkspace(t *testing.T) {
	plan := Build(t.TempDir(), "anything")
	if !plan.Empty() {
		t.Fatalf("expected empty plan, got %+v", plan)
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
