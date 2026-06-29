package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestETLAgentPackCmdWritesPackToOutputDirectory(t *testing.T) {
	tmp := t.TempDir()
	outputDir := filepath.Join(tmp, ".bqa", "output", "etl-agent-pack")

	cmd := etlAgentPackCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"--sessions", filepath.Join(tmp, ".bqa", "input", "sessions"),
		"--knowledge-dir", filepath.Join(tmp, ".bqa", "knowledge"),
		"--output-dir", outputDir,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	expectedFiles := []string{
		"statistics/summary.md",
		"agents/codex-etl-qa-agent.md",
		"agents/claude-code-etl-qa-agent.md",
		"workflows/etl-regression-workflow.md",
		"workflows/data-reconciliation-workflow.md",
		"workflows/data-quality-validation-workflow.md",
		"specs/etl-test-spec-template.md",
		"specs/source-to-target-mapping-review-checklist.md",
		"prompts/codex-etl-qa-agent-prompt.md",
		"prompts/claude-code-etl-qa-agent-prompt.md",
		"examples/synthetic-etl-reconciliation-example.md",
		"README_NEXT_STEPS.md",
	}

	for _, relativePath := range expectedFiles {
		if _, err := os.Stat(filepath.Join(outputDir, relativePath)); err != nil {
			t.Fatalf("expected generated file %s: %v", relativePath, err)
		}
	}

	if !strings.Contains(out.String(), "ETL agent pack artifacts created: 12") {
		t.Fatalf("command output should include artifact count, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Output dir: "+outputDir) {
		t.Fatalf("command output should include output dir, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Synthetic examples used: true") {
		t.Fatalf("command output should mention synthetic fallback, got %q", out.String())
	}
}
