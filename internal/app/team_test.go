package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTeamPipelineCmdPrintsReadyQAPlan(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "issue.json")
	body := "{\n" +
		"  \"title\": \"Ready QA workflow bug\",\n" +
		"  \"body\": \"## Manual verification\\n\\n```bash\\ngo test ./...\\n```\",\n" +
		"  \"labels\": [\n" +
		"    {\"name\": \"bqa:arch-approved\"},\n" +
		"    {\"name\": \"bqa:ready-qa\"},\n" +
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
		"--subagent", "go-cli-implementer",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := out.String()
	for _, expected := range []string{
		"Mode: dry-run",
		"Source of truth: GitHub issue",
		"Issue: #27 Ready QA workflow bug",
		"Current state: ready-qa",
		"Development: complete",
		"QA: ready",
		"Verify acceptance criteria and manual checks",
		"verify: go test ./...",
		"QA rejection creates bug issue",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
	// ready-qa is forward-looking: QA has not rejected anything yet, so no
	// concrete bug draft should be rendered.
	if strings.Contains(output, "bug title:") {
		t.Fatalf("ready-qa must not render a bug draft, got:\n%s", output)
	}
	if strings.Contains(output, "Architecture is approved, but the issue is not labeled ready for development") {
		t.Fatalf("ready-qa issue must not be routed back to ready-dev, got:\n%s", output)
	}
}

func TestTeamPipelineCmdRendersBugDraftOnQAFailed(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "issue.json")
	body := "{\n" +
		"  \"title\": \"QA failed workflow bug\",\n" +
		"  \"body\": \"## Manual verification\\n\\n```bash\\ngo test ./...\\n```\",\n" +
		"  \"labels\": [\n" +
		"    {\"name\": \"bqa:ready-dev\"},\n" +
		"    {\"name\": \"bqa:qa-failed\"},\n" +
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
		"--subagent", "go-cli-implementer",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := out.String()
	for _, expected := range []string{
		"Current state: qa-failed",
		"bug title: Bug: QA failed workflow bug",
		"bug labels: bqa:bug",
		"bug body:",
		"mshegolev/bqa-os#27",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}
