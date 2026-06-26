package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestKnowledgeStoreLoadsIndexAndNormalizedSession(t *testing.T) {
	dir := t.TempDir()
	sessionsDir := filepath.Join(dir, "sessions")
	normalizedDir := filepath.Join(sessionsDir, "normalized", "demo")
	if err := os.MkdirAll(normalizedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "GraphQL query resolver schema"
	if err := os.WriteFile(filepath.Join(normalizedDir, "session.md"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	index := ports.SessionIndex{Entries: []ports.SessionIndexEntry{{Source: "demo", NormalizedPath: "normalized/demo/session.md"}}}
	data, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessionsDir, "index.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	store := KnowledgeStore{SessionBaseDir: sessionsDir, KnowledgeDir: filepath.Join(dir, "knowledge")}
	loaded, err := store.LoadSessionIndex(context.Background())
	if err != nil {
		t.Fatalf("LoadSessionIndex returned error: %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 index entry, got %d", len(loaded.Entries))
	}

	body, err := store.ReadNormalizedSession(context.Background(), loaded.Entries[0].NormalizedPath)
	if err != nil {
		t.Fatalf("ReadNormalizedSession returned error: %v", err)
	}
	if body != content {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestKnowledgeStoreWritesYamlArtifact(t *testing.T) {
	dir := t.TempDir()
	knowledgeDir := filepath.Join(dir, "knowledge")
	store := KnowledgeStore{KnowledgeDir: knowledgeDir}

	content := "etl_patterns:\n  []\n"
	if err := store.WriteKnowledgeArtifact(context.Background(), "etl_patterns.yaml", content); err != nil {
		t.Fatalf("WriteKnowledgeArtifact returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(knowledgeDir, "etl_patterns.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Fatalf("unexpected content: %q", string(data))
	}
}
