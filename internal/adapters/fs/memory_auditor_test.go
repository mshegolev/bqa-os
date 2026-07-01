package fs

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMemoryAuditorFlagsSecretCandidate(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "agents", "clean.md"), "just docs\n")
	mustWrite(t, filepath.Join(dir, "agents", "leak.md"), "password: hunter2\n")

	rep, err := MemoryAuditor{}.Audit(context.Background(), dir)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	if rep.FilesScanned < 2 {
		t.Fatalf("expected >=2 scanned, got %d", rep.FilesScanned)
	}
	if rep.Candidates < 1 {
		t.Fatalf("expected the leaked secret flagged as a candidate, got %d", rep.Candidates)
	}
}
