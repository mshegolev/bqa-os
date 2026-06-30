package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKnowledgeStoreReadNormalizedSessionResolvesRelativePathUnderSessionBase(t *testing.T) {
	tmp := t.TempDir()
	sessionBase := filepath.Join(tmp, ".bqa", "input", "sessions")
	relativePath := filepath.Join("normalized", "codex", "session.md")
	if err := os.MkdirAll(filepath.Dir(filepath.Join(sessionBase, relativePath)), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionBase, relativePath), []byte("Task: synthetic ETL row count validation"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	store := KnowledgeStore{SessionBaseDir: sessionBase}
	body, err := store.ReadNormalizedSession(context.Background(), relativePath)
	if err != nil {
		t.Fatalf("ReadNormalizedSession returned error: %v", err)
	}
	if !strings.Contains(body, "synthetic ETL row count") {
		t.Fatalf("expected normalized session body, got %q", body)
	}
}

func TestKnowledgeStoreReadNormalizedSessionAcceptsBasePrefixedRelativePath(t *testing.T) {
	tmp := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore Chdir returned error: %v", err)
		}
	})

	sessionBase := filepath.Join(".bqa", "input", "sessions")
	indexedPath := filepath.Join(sessionBase, "normalized", "codex", "session.md")
	if err := os.MkdirAll(filepath.Dir(indexedPath), 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(indexedPath, []byte("Task: synthetic DQ checksum validation"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	store := KnowledgeStore{SessionBaseDir: sessionBase}
	body, err := store.ReadNormalizedSession(context.Background(), indexedPath)
	if err != nil {
		t.Fatalf("ReadNormalizedSession returned error: %v", err)
	}
	if !strings.Contains(body, "synthetic DQ checksum") {
		t.Fatalf("expected normalized session body, got %q", body)
	}
}
