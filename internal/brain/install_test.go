package brain

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile creates a file and its parent directories with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// buildSourcePackage creates a minimal synthetic brain package and returns its path.
func buildSourcePackage(t *testing.T) string {
	t.Helper()
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "registry", "brain.yaml"), "version: 1\n")
	writeFile(t, filepath.Join(src, "registry", "agents.yaml"), "agents: []\n")
	writeFile(t, filepath.Join(src, "agents", "master.md"), "# master agent\n")
	writeFile(t, filepath.Join(src, "skills", "etl-qa.md"), "# etl qa\n")
	writeFile(t, filepath.Join(src, "workflows", "review.md"), "# review\n")
	writeFile(t, filepath.Join(src, "prompts", "bqa-master-context.md"), "context\n")
	writeFile(t, filepath.Join(src, "knowledge", "facts.md"), "facts\n")
	// Unsafe content that must NOT be copied.
	writeFile(t, filepath.Join(src, "sessions", "raw.log"), "secret session\n")
	writeFile(t, filepath.Join(src, ".secrets"), "token=abc\n")
	return src
}

func TestInstallCopiesArtifacts(t *testing.T) {
	src := buildSourcePackage(t)
	target := t.TempDir()

	result, err := Install(src, target)
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	bqa := filepath.Join(target, ".bqa")
	wantFiles := []string{
		"registry/brain.yaml",
		"registry/agents.yaml",
		"agents/master.md",
		"skills/etl-qa.md",
		"workflows/review.md",
		"prompts/bqa-master-context.md",
		"knowledge/facts.md",
	}
	for _, rel := range wantFiles {
		if _, err := os.Stat(filepath.Join(bqa, filepath.FromSlash(rel))); err != nil {
			t.Errorf("expected installed file %s: %v", rel, err)
		}
	}

	// Unsafe artifacts must never land in the target.
	for _, rel := range []string{"sessions/raw.log", ".secrets"} {
		if _, err := os.Stat(filepath.Join(bqa, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Errorf("unsafe artifact %s was copied (err=%v)", rel, err)
		}
	}

	if len(result.Files) != len(wantFiles) {
		t.Errorf("result.Files = %d entries, want %d: %v", len(result.Files), len(wantFiles), result.Files)
	}
	if result.BqaDir != bqa {
		t.Errorf("result.BqaDir = %q, want %q", result.BqaDir, bqa)
	}
}

func TestInstallPreservesExistingTargetFiles(t *testing.T) {
	src := buildSourcePackage(t)
	target := t.TempDir()
	existing := filepath.Join(target, "README.md")
	writeFile(t, existing, "client readme\n")

	if _, err := Install(src, target); err != nil {
		t.Fatalf("Install returned error: %v", err)
	}

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatalf("existing file removed: %v", err)
	}
	if string(data) != "client readme\n" {
		t.Errorf("existing file was modified: %q", string(data))
	}
}

func TestInstallRejectsMissingSource(t *testing.T) {
	target := t.TempDir()
	if _, err := Install(filepath.Join(t.TempDir(), "nope"), target); err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
}

func TestInstallRejectsMissingTarget(t *testing.T) {
	src := buildSourcePackage(t)
	if _, err := Install(src, filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing target, got nil")
	}
}

func TestInstallRejectsInvalidPackage(t *testing.T) {
	// Directory exists but has no registry / artifact directories.
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "random.txt"), "x\n")
	target := t.TempDir()
	if _, err := Install(src, target); err == nil {
		t.Fatal("expected error for invalid brain package, got nil")
	}
}

func TestInstallRejectsEmptyFlags(t *testing.T) {
	if _, err := Install("", t.TempDir()); err == nil {
		t.Fatal("expected error for empty source")
	}
	if _, err := Install(t.TempDir(), ""); err == nil {
		t.Fatal("expected error for empty target")
	}
}

func TestInstallRejectsFileAsTarget(t *testing.T) {
	src := buildSourcePackage(t)
	file := filepath.Join(t.TempDir(), "file.txt")
	writeFile(t, file, "x\n")
	if _, err := Install(src, file); err == nil {
		t.Fatal("expected error when target is a file, got nil")
	}
}
