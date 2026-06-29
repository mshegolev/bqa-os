package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCmdPrintsKnowledgeBuildSummary(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	writeSyntheticSessionFixture(t, filepath.Join(tmp, ".bqa", "input", "sessions"))

	cmd := buildCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	expected := "BQA knowledge build completed\n" +
		"Sessions processed: 1\n" +
		"Artifacts created: 7\n" +
		"Output directory: .bqa/knowledge\n"
	if out.String() != expected {
		t.Fatalf("unexpected build output:\nwant:\n%s\ngot:\n%s", expected, out.String())
	}

	expectedFiles := []string{
		"etl_patterns.yaml",
		"graphql_patterns.yaml",
		"api_patterns.yaml",
		"data_quality_patterns.yaml",
		"common_bugs.yaml",
		"successful_prompts.yaml",
		"project_profile.yaml",
	}
	for _, filename := range expectedFiles {
		if _, err := os.Stat(filepath.Join(tmp, ".bqa", "knowledge", filename)); err != nil {
			t.Fatalf("expected knowledge artifact %s: %v", filename, err)
		}
	}
	if _, err := os.Stat(filepath.Join(tmp, ".bqa", "skills")); !os.IsNotExist(err) {
		t.Fatalf("build command should not create starter BQA artifacts for this issue, stat err: %v", err)
	}
}

func TestBuildCmdMissingInputReturnsActionableError(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	cmd := buildCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected missing input error")
	}
	message := err.Error()
	if !strings.Contains(message, ".bqa/input/sessions/index.json") {
		t.Fatalf("expected missing index path in error, got %q", message)
	}
	if !strings.Contains(message, "bqa discover") || !strings.Contains(message, "bqa ingest") {
		t.Fatalf("expected actionable next commands in error, got %q", message)
	}
}

func writeSyntheticSessionFixture(t *testing.T, sessionsDir string) {
	t.Helper()

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
}
