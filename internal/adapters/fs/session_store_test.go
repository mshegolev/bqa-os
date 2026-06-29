package fs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestSessionStoreSaveNormalizedReturnsPathRelativeToSessionBase(t *testing.T) {
	tmp := t.TempDir()
	baseDir := filepath.Join(tmp, ".bqa", "input", "sessions")
	hash := strings.Repeat("a", 64)
	store := &SessionStore{BaseDir: baseDir, seq: 1}

	normalizedPath, err := store.SaveNormalized(context.Background(), ports.NormalizedSession{
		Ref:     ports.SessionRef{Source: "codex"},
		Content: "Task: synthetic ETL row count validation",
		SHA256:  hash,
	})
	if err != nil {
		t.Fatalf("SaveNormalized returned error: %v", err)
	}

	want := filepath.Join("normalized", "codex", "000001-"+hash[:12]+".md")
	if normalizedPath != want {
		t.Fatalf("expected normalized path %q, got %q", want, normalizedPath)
	}
	if _, err := os.Stat(filepath.Join(baseDir, normalizedPath)); err != nil {
		t.Fatalf("expected normalized file under session base: %v", err)
	}
}
