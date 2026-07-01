package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/ports"
)

func writeBqa(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// exportUseCase builds a use case wired with the real fs adapters for export tests.
func exportUseCase() UseCase {
	return UseCase{
		Source:  fs.MemorySource{},
		Auditor: fs.MemoryAuditor{},
		Writers: map[string]ports.ArchiveWriter{"zip": fs.ZipArchive{}, "tar": fs.TarArchive{}},
	}
}

func TestExportZipContainsAllowListAndMetadata(t *testing.T) {
	tmp := t.TempDir()
	bqa := filepath.Join(tmp, ".bqa")
	writeBqa(t, bqa, map[string]string{
		"knowledge/etl_patterns.yaml": "schema_version: 1\n",
		"registry/index.yaml":         "registry:\n",
		"input/sessions/raw.md":       "password: hunter2\n", // must be excluded
	})
	out := filepath.Join(tmp, "bundle.zip")

	res, err := exportUseCase().Export(context.Background(), ExportOptions{SourceRoot: bqa, Target: "zip", OutPath: out})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if res.DryRun {
		t.Fatal("not a dry run")
	}
	files, err := fs.ZipArchive{}.ReadArchive(context.Background(), out)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	got := map[string]bool{}
	for _, f := range files {
		got[f.Path] = true
		if f.Path == "input/sessions/raw.md" {
			t.Fatal("input/sessions must never be exported")
		}
	}
	for _, want := range []string{"knowledge/etl_patterns.yaml", "registry/index.yaml", "manifest.json", "metadata/checksums.json"} {
		if !got[want] {
			t.Fatalf("bundle missing %q", want)
		}
	}
}

func TestExportDryRunWritesNothing(t *testing.T) {
	tmp := t.TempDir()
	bqa := filepath.Join(tmp, ".bqa")
	writeBqa(t, bqa, map[string]string{"knowledge/x.yaml": "k\n"})
	out := filepath.Join(tmp, "bundle.zip")
	if _, err := exportUseCase().Export(context.Background(), ExportOptions{SourceRoot: bqa, Target: "zip", OutPath: out, DryRun: true}); err != nil {
		t.Fatalf("dry-run export: %v", err)
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write the archive")
	}
}

func TestExportStrictBlocksOnSecret(t *testing.T) {
	tmp := t.TempDir()
	bqa := filepath.Join(tmp, ".bqa")
	writeBqa(t, bqa, map[string]string{"prompts/leak.md": "password: hunter2\n"})
	_, err := exportUseCase().Export(context.Background(), ExportOptions{SourceRoot: bqa, Target: "zip", OutPath: filepath.Join(tmp, "b.zip"), Strict: true})
	if err == nil {
		t.Fatal("expected --strict export to be blocked by the audit")
	}
}

func TestExportEmptyErrors(t *testing.T) {
	tmp := t.TempDir()
	if _, err := exportUseCase().Export(context.Background(), ExportOptions{SourceRoot: filepath.Join(tmp, ".bqa"), Target: "zip", OutPath: filepath.Join(tmp, "b.zip")}); err == nil {
		t.Fatal("expected error when there is nothing to export")
	}
}
