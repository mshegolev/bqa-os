package fs

import (
	"context"
	"path/filepath"
	"testing"
)

func TestGovernanceStoreReadMissing(t *testing.T) {
	store := GovernanceStore{}
	_, exists, err := store.ReadFile(context.Background(), t.TempDir(), "skill_candidates.yaml")
	if err != nil {
		t.Fatalf("ReadFile on missing: %v", err)
	}
	if exists {
		t.Fatalf("expected exists=false for missing file")
	}
}

func TestGovernanceStoreWriteThenRead(t *testing.T) {
	dir := t.TempDir()
	store := GovernanceStore{}
	memoryDir := filepath.Join(dir, ".bqa", "memory")

	if err := store.WriteFile(context.Background(), memoryDir, "skill_candidates.yaml", "hello\n"); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	content, exists, err := store.ReadFile(context.Background(), memoryDir, "skill_candidates.yaml")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !exists || content != "hello\n" {
		t.Fatalf("round-trip mismatch: exists=%v content=%q", exists, content)
	}
}

func TestGovernanceStoreRejectsEscapingName(t *testing.T) {
	store := GovernanceStore{}
	if err := store.WriteFile(context.Background(), t.TempDir(), "../escape.yaml", "x"); err == nil {
		t.Fatalf("expected error for escaping name")
	}
	if _, _, err := store.ReadFile(context.Background(), t.TempDir(), "../escape.yaml"); err == nil {
		t.Fatalf("expected error reading escaping name")
	}
}

func TestGovernanceStoreCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := GovernanceStore{}
	if _, _, err := store.ReadFile(ctx, t.TempDir(), "skill_candidates.yaml"); err == nil {
		t.Fatalf("expected error on cancelled context (read)")
	}
	if err := store.WriteFile(ctx, t.TempDir(), "skill_candidates.yaml", "x"); err == nil {
		t.Fatalf("expected error on cancelled context (write)")
	}
}
