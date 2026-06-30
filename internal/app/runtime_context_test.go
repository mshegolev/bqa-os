package app

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// chdirTemp switches into a fresh temp working directory for the duration of
// the test and restores the previous directory afterwards. The runtime
// commands write files relative to cwd via fs.RuntimeStore{TargetDir: "."},
// so the test must run from a writable scratch directory. This mirrors the
// os.Getwd()+os.Chdir()+defer pattern used by the fs adapter tests so the
// suite stays Go 1.22 compatible (t.Chdir requires Go 1.24+).
func chdirTemp(t *testing.T) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore Chdir returned error: %v", err)
		}
	})
}

func TestRuntimeContextCmdStdoutContract(t *testing.T) {
	chdirTemp(t)

	cmd := runtimeContextCmd("claude")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	out := buf.String()
	const stableLine = "BQA context generated: .bqa/prompts/bqa-master-context.md\n"
	if !strings.Contains(out, stableLine) {
		t.Fatalf("expected stable first line %q in output, got:\n%s", stableLine, out)
	}
}

func TestRuntimeInstallCommandsStdoutContract(t *testing.T) {
	chdirTemp(t)

	root := runtimeCmd()
	if _, ok := findSubcommand(root, "install-commands"); !ok {
		t.Fatalf("install-commands subcommand not found under runtime")
	}

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"install-commands"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	out := buf.String()
	wantLines := []string{
		"BQA runtime command written: .bqa/runtime-commands/codex/bqa-master.md\n",
		"BQA runtime command written: .bqa/runtime-commands/claude/bqa-master.md\n",
		"BQA runtime command written: .bqa/runtime-commands/opencode/bqa-master.md\n",
		"BQA runtime command written: .claude/commands/bqa-master.md\n",
		"BQA master context generated: .bqa/prompts/bqa-master-context.md\n",
		"Claude Code can use /bqa-master in this project.\n",
		"Codex and OpenCode command helpers are available under .bqa/runtime-commands/.\n",
	}
	for _, line := range wantLines {
		if !strings.Contains(out, line) {
			t.Fatalf("expected line %q in output, got:\n%s", line, out)
		}
	}
}

func findSubcommand(parent *cobra.Command, use string) (*cobra.Command, bool) {
	for _, c := range parent.Commands() {
		if c.Name() == use {
			return c, true
		}
	}
	return nil, false
}
