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

func TestBuildCheckPassesAfterSuccessfulBuild(t *testing.T) {
	tmp := t.TempDir()
	sessionsDir := writeSyntheticSessionFixture(t, tmp)
	knowledgeDir := filepath.Join(tmp, ".bqa", "knowledge")

	// First build to generate artifacts.
	buildIt := buildCmd()
	var buildOut bytes.Buffer
	buildIt.SetOut(&buildOut)
	buildIt.SetErr(&buildOut)
	buildIt.SetArgs([]string{"--sessions", sessionsDir, "--knowledge-dir", knowledgeDir})
	if err := buildIt.Execute(); err != nil {
		t.Fatalf("build Execute returned error: %v", err)
	}

	// Then validate with --check; should exit 0.
	checkIt := buildCmd()
	var checkOut bytes.Buffer
	checkIt.SetOut(&checkOut)
	checkIt.SetErr(&checkOut)
	checkIt.SetArgs([]string{"--sessions", sessionsDir, "--knowledge-dir", knowledgeDir, "--check"})
	if err := checkIt.Execute(); err != nil {
		t.Fatalf("build --check returned error on valid output: %v\noutput: %s", err, checkOut.String())
	}
	if !strings.Contains(checkOut.String(), "valid") {
		t.Fatalf("expected success message, got %q", checkOut.String())
	}
}

func TestBuildCheckFailsOnMissingArtifacts(t *testing.T) {
	tmp := t.TempDir()
	sessionsDir := writeSyntheticSessionFixture(t, tmp)
	knowledgeDir := filepath.Join(tmp, ".bqa", "knowledge")

	// No build performed; knowledge dir is empty/missing.
	checkIt := buildCmd()
	var checkOut bytes.Buffer
	checkIt.SetOut(&checkOut)
	checkIt.SetErr(&checkOut)
	checkIt.SetArgs([]string{"--sessions", sessionsDir, "--knowledge-dir", knowledgeDir, "--check"})
	err := checkIt.Execute()
	if err == nil {
		t.Fatalf("expected non-zero exit (error) when artifacts are missing, output: %s", checkOut.String())
	}
	if !strings.Contains(checkOut.String(), "Invalid build output") {
		t.Fatalf("expected invalid output listing, got %q", checkOut.String())
	}
}

func TestBuildCheckFailsOnEmptyArtifact(t *testing.T) {
	tmp := t.TempDir()
	sessionsDir := writeSyntheticSessionFixture(t, tmp)
	knowledgeDir := filepath.Join(tmp, ".bqa", "knowledge")

	buildIt := buildCmd()
	var buildOut bytes.Buffer
	buildIt.SetOut(&buildOut)
	buildIt.SetErr(&buildOut)
	buildIt.SetArgs([]string{"--sessions", sessionsDir, "--knowledge-dir", knowledgeDir})
	if err := buildIt.Execute(); err != nil {
		t.Fatalf("build Execute returned error: %v", err)
	}

	// Corrupt one artifact by truncating it to empty.
	corrupted := filepath.Join(knowledgeDir, "project_profile.yaml")
	if err := os.WriteFile(corrupted, []byte(""), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	checkIt := buildCmd()
	var checkOut bytes.Buffer
	checkIt.SetOut(&checkOut)
	checkIt.SetErr(&checkOut)
	checkIt.SetArgs([]string{"--sessions", sessionsDir, "--knowledge-dir", knowledgeDir, "--check"})
	if err := checkIt.Execute(); err == nil {
		t.Fatalf("expected non-zero exit for empty artifact, output: %s", checkOut.String())
	}
	if !strings.Contains(checkOut.String(), "project_profile.yaml") {
		t.Fatalf("expected the empty file to be named, got %q", checkOut.String())
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
