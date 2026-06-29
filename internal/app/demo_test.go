package app

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDemoArchiveCmdWritesSyntheticArchive(t *testing.T) {
	tmp := t.TempDir()
	output := filepath.Join(tmp, "demo-archive.zip")
	cmd := demoArchiveCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"--output", output})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if !strings.Contains(out.String(), "Demo archive written: "+output) {
		t.Fatalf("output should include archive path, got %q", out.String())
	}
	if !strings.Contains(out.String(), "Files included: 15") {
		t.Fatalf("output should include file count, got %q", out.String())
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected archive at %s: %v", output, err)
	}

	reader, err := zip.OpenReader(output)
	if err != nil {
		t.Fatalf("OpenReader returned error: %v", err)
	}
	defer reader.Close()

	var foundReadme bool
	for _, file := range reader.File {
		if file.Name == "README_NEXT_STEPS.md" {
			foundReadme = true
			break
		}
	}
	if !foundReadme {
		t.Fatalf("archive missing README_NEXT_STEPS.md")
	}
}
