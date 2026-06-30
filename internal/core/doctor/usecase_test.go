package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAllPresent(t *testing.T) {
	base := t.TempDir()
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(filepath.Join(base, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	report := Run(base)
	if !report.OK {
		t.Fatalf("expected OK report, got failing checks: %+v", report.Checks)
	}
	if len(report.Checks) != len(requiredDirs) {
		t.Fatalf("expected %d checks, got %d", len(requiredDirs), len(report.Checks))
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
