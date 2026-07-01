package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/ports"
)

func exportFixture(t *testing.T) (string, string) {
	t.Helper()
	tmp := t.TempDir()
	bqa := filepath.Join(tmp, ".bqa")
	writeBqa(t, bqa, map[string]string{
		"knowledge/etl.yaml":  "schema_version: 1\n",
		"registry/index.yaml": "registry:\n  version: 1\n",
		"agents/qa.md":        "# agent\n",
	})
	out := filepath.Join(tmp, "bundle.zip")
	uc := UseCase{Source: fs.MemorySource{}, Auditor: fs.MemoryAuditor{}, Writers: map[string]ports.ArchiveWriter{"zip": fs.ZipArchive{}}}
	if _, err := uc.Export(context.Background(), ExportOptions{SourceRoot: bqa, Target: "zip", OutPath: out}); err != nil {
		t.Fatalf("export fixture: %v", err)
	}
	return tmp, out
}

func importUseCase() UseCase {
	return UseCase{
		Installer: fs.MemoryInstaller{},
		Readers:   map[string]ports.ArchiveReader{"zip": fs.ZipArchive{}, "tar": fs.TarArchive{}},
	}
}

func TestImportInstallsVerifiedBundle(t *testing.T) {
	tmp, bundle := exportFixture(t)
	target := filepath.Join(tmp, "client")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := importUseCase().Import(context.Background(), ImportOptions{FromPath: bundle, Target: target})
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if !res.Verified {
		t.Fatal("expected verified")
	}
	if _, err := os.Stat(filepath.Join(target, ".bqa", "knowledge", "etl.yaml")); err != nil {
		t.Fatalf("expected installed knowledge file: %v", err)
	}
}

func TestImportDryRunWritesNothing(t *testing.T) {
	tmp, bundle := exportFixture(t)
	target := filepath.Join(tmp, "client")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := importUseCase().Import(context.Background(), ImportOptions{FromPath: bundle, Target: target, DryRun: true}); err != nil {
		t.Fatalf("dry-run import: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".bqa")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not install anything")
	}
}

func TestImportRejectsTamperedChecksum(t *testing.T) {
	tmp, bundle := exportFixture(t)
	// Rewrite the archive with a tampered payload file but the original checksums.
	files, err := fs.ZipArchive{}.ReadArchive(context.Background(), bundle)
	if err != nil {
		t.Fatal(err)
	}
	for i := range files {
		if files[i].Path == "knowledge/etl.yaml" {
			files[i].Data = []byte("TAMPERED\n")
		}
	}
	tampered := filepath.Join(tmp, "tampered.zip")
	if err := (fs.ZipArchive{}).WriteArchive(context.Background(), tampered, files); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(tmp, "client")
	_ = os.MkdirAll(target, 0o755)
	if _, err := importUseCase().Import(context.Background(), ImportOptions{FromPath: tampered, Target: target}); err == nil {
		t.Fatal("expected checksum-mismatch error")
	}
	if _, err := os.Stat(filepath.Join(target, ".bqa")); !os.IsNotExist(err) {
		t.Fatal("nothing must be written to target on verification failure")
	}
}

func TestImportRejectsMissingManifest(t *testing.T) {
	tmp := t.TempDir()
	bundle := filepath.Join(tmp, "nomanifest.zip")
	if err := (fs.ZipArchive{}).WriteArchive(context.Background(), bundle, []ports.ArchiveFile{{Path: "knowledge/x.yaml", Data: []byte("k")}}); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(tmp, "client")
	_ = os.MkdirAll(target, 0o755)
	if _, err := importUseCase().Import(context.Background(), ImportOptions{FromPath: bundle, Target: target}); err == nil {
		t.Fatal("expected missing-manifest error")
	}
}
