package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKnowledgeStoreLoadSessionIndexMissingReturnsActionableError(t *testing.T) {
	store := KnowledgeStore{SessionBaseDir: filepath.Join(t.TempDir(), ".bqa", "input", "sessions")}

	_, err := store.LoadSessionIndex(context.Background())
	if err == nil {
		t.Fatalf("LoadSessionIndex returned nil error for missing index")
	}
	if !strings.Contains(err.Error(), "run bqa discover and bqa ingest2 first") {
		t.Fatalf("LoadSessionIndex error = %q, expected discover/ingest2 guidance", err.Error())
	}
}

func TestKnowledgeStoreLoadSessionIndexInvalidJSONReturnsClearError(t *testing.T) {
	sessionBase := filepath.Join(t.TempDir(), ".bqa", "input", "sessions")
	if err := os.MkdirAll(sessionBase, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionBase, "index.json"), []byte("{invalid"), 0o600); err != nil {
		t.Fatalf("WriteFile index returned error: %v", err)
	}
	store := KnowledgeStore{SessionBaseDir: sessionBase}

	_, err := store.LoadSessionIndex(context.Background())
	if err == nil {
		t.Fatalf("LoadSessionIndex returned nil error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid session index JSON") {
		t.Fatalf("LoadSessionIndex error = %q, expected invalid JSON guidance", err.Error())
	}
}

func TestKnowledgeStoreReadNormalizedSessionRejectsUnsafePath(t *testing.T) {
	tmp := t.TempDir()
	sessionBase := filepath.Join(tmp, ".bqa", "input", "sessions")
	if err := os.MkdirAll(sessionBase, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	outside := filepath.Join(tmp, "outside.md")
	if err := os.WriteFile(outside, []byte("outside session"), 0o600); err != nil {
		t.Fatalf("WriteFile outside returned error: %v", err)
	}
	store := KnowledgeStore{SessionBaseDir: sessionBase}

	body, err := store.ReadNormalizedSession(context.Background(), outside)
	if err == nil {
		t.Fatalf("ReadNormalizedSession returned body %q for unsafe path, expected error", body)
	}
	if !strings.Contains(err.Error(), "unsafe normalized session path") {
		t.Fatalf("ReadNormalizedSession error = %q, expected unsafe path guidance", err.Error())
	}
}

func TestKnowledgeStoreReadNormalizedSessionLoadsValidMarkdown(t *testing.T) {
	sessionBase := filepath.Join(t.TempDir(), ".bqa", "input", "sessions")
	normalizedPath := filepath.Join(sessionBase, "normalized", "synthetic", "000001-session.md")
	if err := os.MkdirAll(filepath.Dir(normalizedPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(normalizedPath, []byte("Task: synthetic API regression"), 0o600); err != nil {
		t.Fatalf("WriteFile normalized returned error: %v", err)
	}
	store := KnowledgeStore{SessionBaseDir: sessionBase}

	for _, path := range []string{normalizedPath, filepath.Join("normalized", "synthetic", "000001-session.md")} {
		body, err := store.ReadNormalizedSession(context.Background(), path)
		if err != nil {
			t.Fatalf("ReadNormalizedSession(%q) returned error: %v", path, err)
		}
		if body != "Task: synthetic API regression" {
			t.Fatalf("ReadNormalizedSession(%q) = %q", path, body)
		}
	}
}

func TestKnowledgeStoreReadNormalizedSessionLoadsDefaultIngestPath(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	sessionBase := filepath.Join(".bqa", "input", "sessions")
	normalizedPath := filepath.Join(sessionBase, "normalized", "synthetic", "000001-session.md")
	if err := os.MkdirAll(filepath.Dir(normalizedPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(normalizedPath, []byte("Task: default ingest path"), 0o600); err != nil {
		t.Fatalf("WriteFile normalized returned error: %v", err)
	}
	store := KnowledgeStore{SessionBaseDir: sessionBase}

	body, err := store.ReadNormalizedSession(context.Background(), normalizedPath)
	if err != nil {
		t.Fatalf("ReadNormalizedSession returned error: %v", err)
	}
	if body != "Task: default ingest path" {
		t.Fatalf("ReadNormalizedSession = %q", body)
	}
}

func TestKnowledgeStoreWritesExpectedKnowledgeArtifacts(t *testing.T) {
	knowledgeDir := filepath.Join(t.TempDir(), ".bqa", "knowledge")
	store := KnowledgeStore{KnowledgeDir: knowledgeDir}
	expected := []string{
		"etl_patterns.yaml",
		"graphql_patterns.yaml",
		"api_patterns.yaml",
		"data_quality_patterns.yaml",
		"common_bugs.yaml",
		"successful_prompts.yaml",
		"project_profile.yaml",
	}

	for _, filename := range expected {
		if err := store.WriteKnowledgeArtifact(context.Background(), filename, filename+": {}\n"); err != nil {
			t.Fatalf("WriteKnowledgeArtifact(%q) returned error: %v", filename, err)
		}
		data, err := os.ReadFile(filepath.Join(knowledgeDir, filename))
		if err != nil {
			t.Fatalf("ReadFile(%q) returned error: %v", filename, err)
		}
		if string(data) != filename+": {}\n" {
			t.Fatalf("artifact %q content = %q", filename, string(data))
		}
	}
}

func TestKnowledgeStoreWriteKnowledgeArtifactRejectsUnexpectedFilename(t *testing.T) {
	knowledgeDir := filepath.Join(t.TempDir(), ".bqa", "knowledge")
	store := KnowledgeStore{KnowledgeDir: knowledgeDir}

	for _, filename := range []string{"unexpected.yaml", "../etl_patterns.yaml"} {
		if err := store.WriteKnowledgeArtifact(context.Background(), filename, "unsafe: true\n"); err == nil {
			t.Fatalf("WriteKnowledgeArtifact(%q) returned nil error, expected rejection", filename)
		}
	}
}
