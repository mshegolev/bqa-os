package fs

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestZipArchiveRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	out := filepath.Join(tmp, "bundle.zip")
	in := []ports.ArchiveFile{
		{Path: "knowledge/etl.yaml", Data: []byte("schema_version: 1\n")},
		{Path: "manifest.json", Data: []byte(`{"bundle_version":1}`)},
	}
	w := ZipArchive{}
	if err := w.WriteArchive(context.Background(), out, in); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := ZipArchive{}.ReadArchive(context.Background(), out)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(got) != 2 || got[0].Path != "knowledge/etl.yaml" || string(got[1].Data) != `{"bundle_version":1}` {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}
