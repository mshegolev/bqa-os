package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCmdDoesNotPrintSalesGeneratedDirByDefault(t *testing.T) {
	tmp := t.TempDir()
	sessionsDir := writeSyntheticSessionFixture(t, tmp)

	cmd := buildCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"--sessions", sessionsDir,
		"--knowledge-dir", filepath.Join(tmp, ".bqa", "knowledge"),
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if strings.Contains(out.String(), ".bqa/sales") {
		t.Fatalf("build output should not list .bqa/sales by default, got %q", out.String())
	}
}

func TestBuildCmdPrintsSalesGeneratedDirWhenEnabled(t *testing.T) {
	tmp := t.TempDir()
	sessionsDir := writeSyntheticSessionFixture(t, tmp)

	cmd := buildCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"--sessions", sessionsDir,
		"--knowledge-dir", filepath.Join(tmp, ".bqa", "knowledge"),
		"--sales-package",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(out.String(), ".bqa/sales") {
		t.Fatalf("build output should list .bqa/sales when enabled, got %q", out.String())
	}
}

func writeSyntheticSessionFixture(t *testing.T, tmp string) string {
	t.Helper()

	sessionsDir := filepath.Join(tmp, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	normalizedPath := filepath.Join(sessionsDir, "synthetic-session.md")
	index := fmt.Sprintf(`{"generated_at":"2026-06-29T00:00:00Z","entries":[{"source":"synthetic","original_path":"synthetic-session.md","normalized_path":%q,"size":123,"sha256":"synthetic","modified":"2026-06-29T00:00:00Z"}]}`, normalizedPath)
	if err := os.WriteFile(filepath.Join(sessionsDir, "index.json"), []byte(index), 0o600); err != nil {
		t.Fatalf("WriteFile index returned error: %v", err)
	}
	if err := os.WriteFile(normalizedPath, []byte("Task: synthetic API regression and GraphQL testing"), 0o600); err != nil {
		t.Fatalf("WriteFile session returned error: %v", err)
	}

	return sessionsDir
}
