package fs

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestTarArchiveRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "bundle.tar")
	in := []ports.ArchiveFile{
		{Path: "agents/qa.md", Data: []byte("# agent\n")},
		{Path: "manifest.json", Data: []byte(`{"bundle_version":1}`)},
	}
	if err := (TarArchive{}).WriteArchive(context.Background(), out, in); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := TarArchive{}.ReadArchive(context.Background(), out)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got) != 2 || got[0].Path != "agents/qa.md" || string(got[0].Data) != "# agent\n" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}
