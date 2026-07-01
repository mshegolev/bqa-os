package memory

import (
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func samplePayload() []ports.ArchiveFile {
	return []ports.ArchiveFile{
		{Path: "knowledge/etl.yaml", Data: []byte("schema_version: 1\n")},
		{Path: "registry/index.yaml", Data: []byte("registry:\n")},
	}
}

func TestAssembleBundleIncludesMetadata(t *testing.T) {
	files, err := assembleBundle(samplePayload(), ports.AuditReport{FilesScanned: 2, Candidates: 0})
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	index := map[string]bool{}
	for _, f := range files {
		index[f.Path] = true
	}
	for _, want := range []string{"knowledge/etl.yaml", "registry/index.yaml", "manifest.json", "metadata/checksums.json", "metadata/memory_audit.yaml", "README.md", "install.md", "audit.md"} {
		if !index[want] {
			t.Fatalf("bundle missing %q", want)
		}
	}
}

func TestVerifyChecksumsDetectsTamper(t *testing.T) {
	payload := samplePayload()
	c := buildChecksums(payload)
	if err := verifyChecksums(payload, c); err != nil {
		t.Fatalf("expected clean verify, got %v", err)
	}
	payload[0].Data = []byte("TAMPERED")
	if err := verifyChecksums(payload, c); err == nil {
		t.Fatal("expected checksum mismatch after tamper")
	}
}

func TestParseManifestRejectsGarbage(t *testing.T) {
	if _, err := parseManifest([]byte("not json")); err == nil {
		t.Fatal("expected parse error")
	}
	m, err := parseManifest([]byte(`{"bundle_version":1}`))
	if err != nil || m.BundleVersion != 1 {
		t.Fatalf("parse: %+v %v", m, err)
	}
}

func TestManifestOmitsWallClock(t *testing.T) {
	files, _ := assembleBundle(samplePayload(), ports.AuditReport{})
	for _, f := range files {
		if f.Path == "manifest.json" && strings.Contains(string(f.Data), "created_at") {
			t.Fatal("manifest must be deterministic (no created_at)")
		}
	}
}
