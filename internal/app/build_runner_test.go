package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestRunBuildCreatesKnowledgeAndBQAArtifacts(t *testing.T) {
	tmp := t.TempDir()
	sessionsDir := filepath.Join(tmp, ".bqa", "input", "sessions")
	normalizedDir := filepath.Join(sessionsDir, "normalized", "synthetic")
	knowledgeDir := filepath.Join(tmp, ".bqa", "knowledge")

	if err := os.MkdirAll(normalizedDir, 0o755); err != nil {
		t.Fatalf("create normalized dir: %v", err)
	}

	normalizedPath := filepath.Join(normalizedDir, "000001-synthetic.md")
	normalizedBody := "Task: test GraphQL query, REST API endpoint, ETL reconciliation, and null check data validation"
	if err := os.WriteFile(normalizedPath, []byte(normalizedBody), 0o600); err != nil {
		t.Fatalf("write normalized session: %v", err)
	}

	index := ports.SessionIndex{Entries: []ports.SessionIndexEntry{{
		Source:         "synthetic",
		OriginalPath:   "/synthetic/session.jsonl",
		RawPath:        filepath.Join(sessionsDir, "raw", "synthetic", "000001-synthetic.jsonl"),
		NormalizedPath: normalizedPath,
		Size:           int64(len(normalizedBody)),
		SHA256:         "synthetic",
		Modified:       "2026-01-01T00:00:00Z",
	}}}
	indexBytes, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		t.Fatalf("marshal index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionsDir, "index.json"), indexBytes, 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}

	summary, err := RunBuild(context.Background(), BuildOptions{SessionBaseDir: sessionsDir, KnowledgeDir: knowledgeDir})
	if err != nil {
		t.Fatalf("RunBuild returned error: %v", err)
	}

	if summary.SessionsProcessed != 1 {
		t.Fatalf("expected 1 processed session, got %d", summary.SessionsProcessed)
	}
	if summary.KnowledgeArtifactsCreated != 9 {
		t.Fatalf("expected 9 knowledge artifacts, got %d", summary.KnowledgeArtifactsCreated)
	}
	if summary.BQAArtifactsCreated != 10 {
		t.Fatalf("expected 10 BQA artifacts, got %d", summary.BQAArtifactsCreated)
	}
	if summary.KnowledgeDir != knowledgeDir {
		t.Fatalf("expected knowledge dir %q, got %q", knowledgeDir, summary.KnowledgeDir)
	}

	assertFileContains(t, filepath.Join(knowledgeDir, "graphql_patterns.yaml"), "graphql_functional_testing")
	assertFileContains(t, filepath.Join(knowledgeDir, "etl_patterns.yaml"), "etl_validation")
	assertFileContains(t, filepath.Join(knowledgeDir, "project_profile.yaml"), "sessions_analyzed: 1")
	assertFileContains(t, filepath.Join(tmp, ".bqa", "skills", "etl-log-investigation.md"), "# ETL Log Investigation")
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("expected %s to contain %q, got %s", path, want, string(data))
	}
}
