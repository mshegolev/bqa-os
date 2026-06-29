package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestETLAgentPackCmdWritesPackAndPrintsSummary(t *testing.T) {
	tmp := t.TempDir()
	sessionsDir := writeSyntheticSessionFixture(t, tmp)
	knowledgeDir := filepath.Join(tmp, ".bqa", "knowledge")

	cmd := etlAgentPackCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"--sessions", sessionsDir,
		"--knowledge-dir", knowledgeDir,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	readmePath := filepath.Join(tmp, ".bqa", "output", "etl-agent-pack", "README_NEXT_STEPS.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Fatalf("expected README_NEXT_STEPS.md to be written: %v", err)
	}
	output := out.String()
	if !strings.Contains(output, "ETL agent pack files created: 12") {
		t.Fatalf("missing file count in command output: %q", output)
	}
	if !strings.Contains(output, "Pack dir: "+filepath.Join(tmp, ".bqa", "output", "etl-agent-pack")) {
		t.Fatalf("missing pack dir in command output: %q", output)
	}
}
