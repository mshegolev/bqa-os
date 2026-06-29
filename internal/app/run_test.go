package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCmdLoadsRegistryAndPrintsStructuredPlan(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	writeSyntheticBQAWorkspace(t, tmp)

	cmd := runCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"Validate", "ETL", "pipeline"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := out.String()
	for _, expected := range []string{
		"BQA Master task: Validate ETL pipeline",
		"Registry loaded: agents=2 skills=2 workflows=2",
		"Selected agents:",
		"- etl-qa-agent (etl): agents/etl-qa-agent.md",
		"Selected workflows:",
		"- etl-verification-workflow (etl): workflows/etl-verification-workflow.md",
		"Execution plan:",
		"Report:",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("run output missing %q, got:\n%s", expected, output)
		}
	}
	for _, marker := range placeholderMarkers() {
		if strings.Contains(output, marker) {
			t.Fatalf("run output must not include placeholder marker %q, got:\n%s", marker, output)
		}
	}
}

func writeSyntheticBQAWorkspace(t *testing.T, root string) {
	t.Helper()
	for _, dir := range []string{".bqa/registry", ".bqa/memory", ".bqa/agents", ".bqa/skills", ".bqa/workflows"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("MkdirAll %s returned error: %v", dir, err)
		}
	}
	files := map[string]string{
		".bqa/registry/index.yaml": "registry:\n  version: 1\n  skills: registry/skills.yaml\n  agents: registry/agents.yaml\n  workflows: registry/workflows.yaml\n",
		".bqa/registry/agents.yaml": "agents:\n" +
			"  - id: etl-qa-agent\n" +
			"    path: agents/etl-qa-agent.md\n" +
			"    domain: etl\n" +
			"  - id: runtime-agent\n" +
			"    path: agents/runtime-agent.md\n" +
			"    domain: runtime\n",
		".bqa/registry/skills.yaml": "skills:\n" +
			"  - id: etl-log-investigation\n" +
			"    path: skills/etl-log-investigation.md\n" +
			"    domain: etl\n" +
			"  - id: runtime-trace-review\n" +
			"    path: skills/runtime-trace-review.md\n" +
			"    domain: runtime\n",
		".bqa/registry/workflows.yaml": "workflows:\n" +
			"  - id: etl-verification-workflow\n" +
			"    path: workflows/etl-verification-workflow.md\n" +
			"    domain: etl\n" +
			"  - id: session-knowledge-workflow\n" +
			"    path: workflows/session-knowledge-workflow.md\n" +
			"    domain: memory\n",
	}
	for path, content := range files {
		if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile %s returned error: %v", path, err)
		}
	}
}

func placeholderMarkers() []string {
	return []string{"TO" + "DO", "FIX" + "ME"}
}
