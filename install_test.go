package bqaos

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallCheckGoPassesWhenMinimumVersionIsPresent(t *testing.T) {
	output, err := runInstallCheckGo(t, "1.22.0")
	if err != nil {
		t.Fatalf("install.sh --check-go failed:\n%s", output)
	}
	for _, expected := range []string{
		"Go version OK: 1.22.0 (minimum 1.22)",
		"Skipping install because --check-go was requested.",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output missing %q in:\n%s", expected, output)
		}
	}
}

func TestInstallCheckGoRejectsOldVersion(t *testing.T) {
	output, err := runInstallCheckGo(t, "1.21.9")
	if err == nil {
		t.Fatalf("install.sh --check-go unexpectedly passed:\n%s", output)
	}
	for _, expected := range []string{
		"Go 1.22 or newer is required.",
		"Found: 1.21.9",
		"brew upgrade go",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output missing %q in:\n%s", expected, output)
		}
	}
}

func runInstallCheckGo(t *testing.T, version string) (string, error) {
	t.Helper()

	fakeBin := t.TempDir()
	fakeGo := filepath.Join(fakeBin, "go")
	script := fmt.Sprintf(`#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == "version" ]]; then
  echo "go version go%s darwin/arm64"
  exit 0
fi
echo "unexpected go command: $*" >&2
exit 42
`, version)
	if err := os.WriteFile(fakeGo, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("bash", "install.sh", "--check-go")
	cmd.Env = append(os.Environ(),
		"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"BQA_INSTALL_DIR="+t.TempDir(),
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return string(output), nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(output), err
	}
	t.Fatalf("install.sh could not run: %v\n%s", err, output)
	return "", err
}
