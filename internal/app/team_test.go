package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTeamPipelineCmdPrintsDryRunPlan(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "issue.json")
	body := "{\n" +
		"  \"title\": \"Codex Team Pipeline MVP\",\n" +
		"  \"body\": \"## Manual verification\\n\\n```bash\\ngo test ./...\\n```\",\n" +
		"  \"labels\": [\n" +
		"    {\"name\": \"bqa:arch-approved\"},\n" +
		"    {\"name\": \"bqa:ready-dev\"},\n" +
		"    {\"name\": \"bqa:codex-team\"}\n" +
		"  ]\n" +
		"}"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	cmd := teamCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"pipeline",
		"--issue-json", path,
		"--issue-number", "27",
		"--repo", "mshegolev/bqa-os",
		"--subagent", "senior-go-ai-engineer",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := out.String()
	for _, expected := range []string{
		"Mode: dry-run",
		"Source of truth: GitHub issue",
		"Issue: #27 Codex Team Pipeline MVP",
		"Selected subagent: senior-go-ai-engineer",
		"Development: ready",
		"go test ./...",
		"QA rejection creates bug issue",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}
