package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// buildPrereqCheckCount is the number of non-directory build-prerequisite
// checks appended by Run on top of requiredDirs.
const buildPrereqCheckCount = 6

// writeValidWorkspace lays out a workspace that satisfies every doctor check:
// all required dirs, a session index with one entry, and a knowledge dir.
func writeValidWorkspace(t *testing.T, base string) {
	t.Helper()
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(filepath.Join(base, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(base, "knowledge"), 0o755); err != nil {
		t.Fatal(err)
	}
	indexPath := filepath.Join(base, "input", "sessions", "index.json")
	index := `{"generated_at":"2026-06-29T00:00:00Z","entries":[{"source":"synthetic","original_path":"s.md","normalized_path":"s.md","size":1,"sha256":"x","modified":"2026-06-29T00:00:00Z"}]}`
	if err := os.WriteFile(indexPath, []byte(index), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestRunAllPresent(t *testing.T) {
	base := t.TempDir()
	writeValidWorkspace(t, base)
	report := Run(base)
	if !report.OK {
		t.Fatalf("expected OK report, got failing checks: %+v", report.Checks)
	}
	if want := len(requiredDirs) + buildPrereqCheckCount; len(report.Checks) != want {
		t.Fatalf("expected %d checks, got %d", want, len(report.Checks))
	}
}

func TestRunMissingIndexSuggestsNextCommand(t *testing.T) {
	base := t.TempDir()
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(filepath.Join(base, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// No index.json written.
	report := Run(base)
	if report.OK {
		t.Fatal("expected report to fail when session index is missing")
	}
	var found bool
	for _, c := range report.Checks {
		if c.Name == "session-index" {
			found = true
			if c.OK {
				t.Fatal("expected session-index check to fail")
			}
			if c.Suggestion == "" {
				t.Fatal("expected a suggestion for the missing session index")
			}
		}
	}
	if !found {
		t.Fatal("expected a session-index check")
	}
}

func TestRunEmptyIndexFails(t *testing.T) {
	base := t.TempDir()
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(filepath.Join(base, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	indexPath := filepath.Join(base, "input", "sessions", "index.json")
	if err := os.WriteFile(indexPath, []byte(`{"generated_at":"x","entries":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	report := Run(base)
	if report.OK {
		t.Fatal("expected report to fail when session index has no entries")
	}
	for _, c := range report.Checks {
		if c.Name == "session-entries" && c.OK {
			t.Fatal("expected session-entries check to fail for empty index")
		}
	}
}

func TestRunNonWritableKnowledgeDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission semantics differ on windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}
	base := t.TempDir()
	writeValidWorkspace(t, base)
	// Make the workspace root read-only so the knowledge dir cannot be (re)created/written.
	roBase := filepath.Join(base, "ro")
	if err := os.MkdirAll(roBase, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(roBase, 0o755) })
	report := Run(roBase)
	if report.OK {
		t.Fatal("expected report to fail for a non-writable workspace")
	}
	for _, c := range report.Checks {
		if c.Name == "knowledge-writable" && c.OK {
			t.Fatal("expected knowledge-writable check to fail")
		}
	}
}

func TestRunMissingDirs(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "registry"), 0o755); err != nil {
		t.Fatal(err)
	}
	report := Run(base)
	if report.OK {
		t.Fatal("expected report to fail when workspace dirs are missing")
	}
	failing := 0
	for _, c := range report.Checks {
		if !c.OK {
			failing++
		}
	}
	if failing == 0 {
		t.Fatal("expected at least one failing check")
	}
}
