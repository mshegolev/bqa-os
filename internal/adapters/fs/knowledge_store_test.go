package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

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
