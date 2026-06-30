package runtime

import (
	"context"
	"strings"
	"testing"
)

func TestInstallCommandsWritesAllFiles(t *testing.T) {
	writer := newFakeWriter()
	uc := UseCase{Writer: writer, Detector: fakeDetector{}}

	res, err := uc.InstallCommands(context.Background())
	if err != nil {
		t.Fatalf("InstallCommands returned error: %v", err)
	}

	if res.ContextPath != masterContextPath {
		t.Errorf("ContextPath = %q, want %q", res.ContextPath, masterContextPath)
	}

	ctxContent, ok := writer.files[masterContextPath]
	if !ok {
		t.Fatalf("master context not written to %q", masterContextPath)
	}
	if !strings.Contains(ctxContent, "through the project-local command runtime") {
		t.Errorf("master context missing project-local marker, got:\n%s", ctxContent)
	}

	wantCmds := []string{
		".bqa/runtime-commands/codex/bqa-master.md",
		".bqa/runtime-commands/claude/bqa-master.md",
		".bqa/runtime-commands/opencode/bqa-master.md",
		".claude/commands/bqa-master.md",
	}
	if len(res.Commands) != len(wantCmds) {
		t.Fatalf("Commands = %v, want %v", res.Commands, wantCmds)
	}
	for i, p := range wantCmds {
		if res.Commands[i] != p {
			t.Errorf("Commands[%d] = %q, want %q", i, res.Commands[i], p)
		}
		content, ok := writer.files[p]
		if !ok {
			t.Fatalf("command file not written to %q", p)
		}
		if !strings.Contains(content, "# /bqa-master") {
			t.Errorf("command file %q missing header, got:\n%s", p, content)
		}
	}
}

func TestDetectReportsEachRuntime(t *testing.T) {
	detector := fakeDetector{found: map[string]string{"codex": "/usr/local/bin/codex"}}
	uc := UseCase{Writer: newFakeWriter(), Detector: detector}

	statuses, err := uc.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect returned error: %v", err)
	}
	if len(statuses) != len(adapters) {
		t.Fatalf("got %d statuses, want %d", len(statuses), len(adapters))
	}

	byName := map[string]RuntimeStatus{}
	for _, s := range statuses {
		byName[s.Name] = s
	}

	codex, ok := byName["codex"]
	if !ok {
		t.Fatalf("missing codex status")
	}
	if !codex.Found || codex.BinaryPath != "/usr/local/bin/codex" {
		t.Errorf("codex status = %+v, want Found with path", codex)
	}

	claude, ok := byName["claude"]
	if !ok {
		t.Fatalf("missing claude status")
	}
	if claude.Found {
		t.Errorf("claude status = %+v, want not found", claude)
	}
}
