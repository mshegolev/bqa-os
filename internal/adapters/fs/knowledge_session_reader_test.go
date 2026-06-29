package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKnowledgeSessionReaderLoadsIndexAndRelativeNormalizedMarkdown(t *testing.T) {
	tmp := t.TempDir()
	sessionsDir := filepath.Join(tmp, "sessions")
	normalizedDir := filepath.Join(sessionsDir, "normalized")
	if err := os.MkdirAll(normalizedDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	index := `{"generated_at":"2026-06-29T00:00:00Z","entries":[{"source":"synthetic","original_path":"synthetic/source.jsonl","normalized_path":"normalized/session.md","size":123,"sha256":"synthetic","modified":"2026-06-29T00:00:00Z"}]}`
	if err := os.WriteFile(filepath.Join(sessionsDir, "index.json"), []byte(index), 0o600); err != nil {
		t.Fatalf("WriteFile index returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(normalizedDir, "session.md"), []byte("Task: verify synthetic ETL row count"), 0o600); err != nil {
		t.Fatalf("WriteFile session returned error: %v", err)
	}

	reader := KnowledgeSessionReader{SessionBaseDir: sessionsDir}
	loaded, err := reader.LoadSessionIndex(context.Background())
	if err != nil {
		t.Fatalf("LoadSessionIndex returned error: %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 index entry, got %d", len(loaded.Entries))
	}

	body, err := reader.ReadNormalizedSession(context.Background(), loaded.Entries[0].NormalizedPath)
	if err != nil {
		t.Fatalf("ReadNormalizedSession returned error: %v", err)
	}
	if !strings.Contains(body, "synthetic ETL row count") {
		t.Fatalf("expected normalized markdown content, got %q", body)
	}
}

func TestKnowledgeSessionReaderMissingIndexErrorIsActionable(t *testing.T) {
	reader := KnowledgeSessionReader{SessionBaseDir: filepath.Join(t.TempDir(), "sessions")}

	_, err := reader.LoadSessionIndex(context.Background())
	if err == nil {
		t.Fatalf("expected missing index error")
	}

	message := err.Error()
	if !strings.Contains(message, ".bqa/input/sessions/index.json") && !strings.Contains(message, "bqa discover") {
		t.Fatalf("expected actionable missing index error, got %q", message)
	}
}
