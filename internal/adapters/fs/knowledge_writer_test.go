package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestKnowledgeWriterCreatesKnowledgeDirAndWritesArtifact(t *testing.T) {
	tmp := t.TempDir()
	knowledgeDir := filepath.Join(tmp, ".bqa", "knowledge")
	writer := KnowledgeWriter{KnowledgeDir: knowledgeDir}

	if err := writer.WriteKnowledgeArtifact(context.Background(), "etl_patterns.yaml", "etl_patterns:\n  []\n"); err != nil {
		t.Fatalf("WriteKnowledgeArtifact returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(knowledgeDir, "etl_patterns.yaml"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(data) != "etl_patterns:\n  []\n" {
		t.Fatalf("unexpected artifact content: %q", string(data))
	}
}
