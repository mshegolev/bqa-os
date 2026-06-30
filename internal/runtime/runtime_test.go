package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCommandsWritesBQAMasterCommandFiles(t *testing.T) {
	tmp := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore Chdir returned error: %v", err)
		}
	})

	if err := InstallCommands(); err != nil {
		t.Fatalf("InstallCommands returned error: %v", err)
	}

	contextBody, err := os.ReadFile(filepath.Clean(".bqa/prompts/bqa-master-context.md"))
	if err != nil {
		t.Fatalf("expected BQA master context to be written: %v", err)
	}
	context := string(contextBody)
	for _, expected := range []string{
		"You are BQA Master Agent",
		".bqa/agents/",
		".bqa/skills/",
		".bqa/workflows/",
		".bqa/guardrails/",
	} {
		if !strings.Contains(context, expected) {
			t.Fatalf("expected BQA master context to contain %q, got:\n%s", expected, context)
		}
	}

	for _, path := range []string{
		".bqa/runtime-commands/codex/bqa-master.md",
		".bqa/runtime-commands/claude/bqa-master.md",
		".bqa/runtime-commands/opencode/bqa-master.md",
		".claude/commands/bqa-master.md",
	} {
		body, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			t.Fatalf("expected %s to be written: %v", path, err)
		}
		content := string(body)
		for _, expected := range []string{
			"Read .bqa/prompts/bqa-master-context.md",
			"act as BQA Master Agent",
			".bqa/agents/",
			".bqa/skills/",
			".bqa/workflows/",
			".bqa/guardrails/",
		} {
			if !strings.Contains(content, expected) {
				t.Fatalf("expected %s to contain %q, got:\n%s", path, expected, content)
			}
		}
	}
}
