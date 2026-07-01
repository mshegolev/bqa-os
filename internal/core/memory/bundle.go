package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/mshegolev/bqa-os/internal/version"
)

// BundleVersion is the memory-bundle format version.
const BundleVersion = 1

// AllowList is the set of .bqa subdirectories safe to export. It mirrors the
// brain-install allow-list so an imported bundle installs cleanly, and never
// includes input/sessions or caches.
var AllowList = []string{"knowledge", "agents", "skills", "workflows", "prompts", "registry"}

type Manifest struct {
	BundleVersion int      `json:"bundle_version"`
	Tool          string   `json:"tool"`
	GeneratedBy   string   `json:"generated_by"`
	Included      []string `json:"included"`
	FileCount     int      `json:"file_count"`
}

type Checksums struct {
	SHA256 map[string]string `json:"sha256"`
}

func buildManifest(payload []ports.ArchiveFile) Manifest {
	return Manifest{
		BundleVersion: BundleVersion,
		Tool:          "bqa",
		GeneratedBy:   "bqa " + version.Version,
		Included:      includedDirs(payload),
		FileCount:     len(payload),
	}
}

func includedDirs(payload []ports.ArchiveFile) []string {
	seen := map[string]bool{}
	for _, f := range payload {
		if i := strings.IndexByte(f.Path, '/'); i > 0 {
			seen[f.Path[:i]] = true
		}
	}
	var out []string
	for _, d := range AllowList {
		if seen[d] {
			out = append(out, d)
		}
	}
	return out
}

func buildChecksums(payload []ports.ArchiveFile) Checksums {
	c := Checksums{SHA256: map[string]string{}}
	for _, f := range payload {
		sum := sha256.Sum256(f.Data)
		c.SHA256[f.Path] = hex.EncodeToString(sum[:])
	}
	return c
}

func verifyChecksums(payload []ports.ArchiveFile, c Checksums) error {
	if len(c.SHA256) != len(payload) {
		return fmt.Errorf("checksum count mismatch: manifest lists %d, bundle has %d", len(c.SHA256), len(payload))
	}
	for _, f := range payload {
		want, ok := c.SHA256[f.Path]
		if !ok {
			return fmt.Errorf("no checksum recorded for %q", f.Path)
		}
		sum := sha256.Sum256(f.Data)
		if hex.EncodeToString(sum[:]) != want {
			return fmt.Errorf("checksum mismatch for %q", f.Path)
		}
	}
	return nil
}

func parseManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("invalid manifest.json: %w", err)
	}
	return m, nil
}

func parseChecksums(data []byte) (Checksums, error) {
	var c Checksums
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("invalid metadata/checksums.json: %w", err)
	}
	return c, nil
}

// assembleBundle returns the full bundle: the payload plus manifest, checksums,
// audit metadata, and the human-readable docs. Output is deterministic.
func assembleBundle(payload []ports.ArchiveFile, audit ports.AuditReport) ([]ports.ArchiveFile, error) {
	manifest := buildManifest(payload)
	mjson, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	cjson, err := json.MarshalIndent(buildChecksums(payload), "", "  ")
	if err != nil {
		return nil, err
	}
	out := append([]ports.ArchiveFile(nil), payload...)
	out = append(out,
		ports.ArchiveFile{Path: "manifest.json", Data: mjson},
		ports.ArchiveFile{Path: "metadata/checksums.json", Data: cjson},
		ports.ArchiveFile{Path: "metadata/memory_audit.yaml", Data: []byte(auditYAML(audit))},
		ports.ArchiveFile{Path: "README.md", Data: []byte(readmeText())},
		ports.ArchiveFile{Path: "install.md", Data: []byte(installText())},
		ports.ArchiveFile{Path: "audit.md", Data: []byte(auditText(audit))},
	)
	return out, nil
}

func auditYAML(a ports.AuditReport) string {
	return fmt.Sprintf("memory_audit:\n  files_scanned: %d\n  candidates: %d\n", a.FilesScanned, a.Candidates)
}

func auditText(a ports.AuditReport) string {
	return fmt.Sprintf("# Memory Audit\n\nFiles scanned: %d\nRedaction candidates: %d\n\nCandidates are files that a sanitize scan flagged as possibly containing secrets or PII. Review before sharing.\n", a.FilesScanned, a.Candidates)
}

func readmeText() string {
	return "# BQA Memory Bundle\n\nA portable, sanitized snapshot of BQA-OS project memory (knowledge, agents, skills, workflows, prompts, registry).\n\n- `manifest.json` — bundle metadata and included directories.\n- `metadata/checksums.json` — SHA-256 of every payload file.\n- `audit.md` / `metadata/memory_audit.yaml` — pre-export sensitivity scan.\n\nSee `install.md` to restore it.\n"
}

func installText() string {
	return "# Installing this bundle\n\n```bash\nbqa brain import --from <this-bundle>.zip --target /path/to/project\n```\n\nImport verifies the manifest and checksums before installing anything, then copies the allow-listed directories into `<target>/.bqa/`. Unrelated files in the target are never modified.\n"
}
