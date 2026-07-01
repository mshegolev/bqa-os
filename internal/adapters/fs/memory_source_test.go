package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMemorySourceReadsAllowListedDirsOnly(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "knowledge", "etl.yaml"), "k")
	mustWrite(t, filepath.Join(root, "agents", "qa.md"), "a")
	mustWrite(t, filepath.Join(root, "input", "sessions", "raw.md"), "SECRET") // NOT allow-listed

	files, err := MemorySource{}.ReadMemory(context.Background(), root, []string{"knowledge", "agents", "registry"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %+v", len(files), files)
	}
	if files[0].Path != "agents/qa.md" || files[1].Path != "knowledge/etl.yaml" {
		t.Fatalf("unexpected paths: %+v", files)
	}
	for _, f := range files {
		if string(f.Data) == "SECRET" {
			t.Fatal("input/sessions must never be exported")
		}
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
