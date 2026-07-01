# Memory Export/Import Backends Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `bqa brain export` / `bqa brain import` so project memory (`.bqa/{knowledge,agents,skills,workflows,prompts,registry}`) can be packaged into a portable zip/tar bundle (manifest + checksums + audit) or pushed to the GitHub brain, and safely restored elsewhere.

**Architecture:** Strict hexagonal — the use case in `internal/core/memory/` depends only on ports + stdlib; every external boundary is a port with an adapter. Pure logic (SHA-256, manifest JSON, allow-list) lives in the core. Output is deterministic (fixed archive mod-times, no wall-clock manifest field).

**Tech Stack:** Go 1.22, stdlib only (`archive/zip`, `archive/tar`, `crypto/sha256`, `encoding/json`), cobra. Reuses `brain.Install` (placement), `brain.Sync` (github), `sanitize.Path` (audit), `pathsafe.RelClean` (path safety).

**Reference spec:** `docs/superpowers/specs/2026-07-01-memory-export-import-backends-design.md`

---

### Task 1: Ports and shared archive helpers

**Files:**
- Create: `internal/ports/archive.go`
- Create: `internal/ports/memory.go`
- Create: `internal/adapters/fs/archive_common.go`
- Test: `internal/adapters/fs/archive_common_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapters/fs/archive_common_test.go`:

```go
package fs

import "testing"

func TestCleanArchivePathRejectsTraversal(t *testing.T) {
	if _, err := cleanArchivePath("../evil"); err == nil {
		t.Fatal("expected error for traversal path")
	}
	if _, err := cleanArchivePath("/abs"); err == nil {
		t.Fatal("expected error for absolute path")
	}
	got, err := cleanArchivePath("knowledge\\etl.yaml")
	if err != nil || got != "knowledge/etl.yaml" {
		t.Fatalf("cleanArchivePath normalize = %q, %v", got, err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/fs/ -run TestCleanArchivePath -v`
Expected: FAIL — `undefined: cleanArchivePath`.

- [ ] **Step 3: Write the ports and helpers**

Create `internal/ports/archive.go`:

```go
package ports

import "context"

// ArchiveFile is one file in a memory bundle. Path is bundle-relative with
// forward slashes.
type ArchiveFile struct {
	Path string
	Data []byte
}

type ArchiveWriter interface {
	WriteArchive(ctx context.Context, outPath string, files []ArchiveFile) error
}

type ArchiveReader interface {
	ReadArchive(ctx context.Context, inPath string) ([]ArchiveFile, error)
}
```

Create `internal/ports/memory.go`:

```go
package ports

import "context"

// MemorySource reads the allow-listed memory files under a root (e.g. ".bqa").
type MemorySource interface {
	ReadMemory(ctx context.Context, root string, allow []string) ([]ArchiveFile, error)
}

// InstalledSummary reports where a bundle was installed.
type InstalledSummary struct {
	Target string
	Files  []string
}

// MemoryInstaller places a verified bundle directory into <target>/.bqa.
type MemoryInstaller interface {
	InstallMemory(ctx context.Context, sourceDir, target string) (InstalledSummary, error)
}

// AuditReport summarizes a pre-export sensitivity scan.
type AuditReport struct {
	FilesScanned int
	Candidates   int
}

// MemoryAuditor scans a staged directory for secret/PII candidates.
type MemoryAuditor interface {
	Audit(ctx context.Context, dir string) (AuditReport, error)
}

// BrainStore pushes assembled memory files to the connected GitHub brain.
type BrainStore interface {
	Push(ctx context.Context, files []ArchiveFile, sanitize bool) error
}
```

Create `internal/adapters/fs/archive_common.go`:

```go
package fs

import (
	"fmt"
	archivepath "path"
	"sort"
	"strings"
	"time"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// archiveModTime is a fixed timestamp so archives are byte-deterministic.
var archiveModTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func sortedArchiveFiles(files []ports.ArchiveFile) []ports.ArchiveFile {
	out := append([]ports.ArchiveFile(nil), files...)
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func cleanArchivePath(value string) (string, error) {
	value = strings.ReplaceAll(value, "\\", "/")
	cleaned := archivepath.Clean(value)
	if cleaned == "." || strings.HasPrefix(cleaned, "../") || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("invalid archive path %q", value)
	}
	return cleaned, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go build ./... && go test ./internal/adapters/fs/ -run TestCleanArchivePath -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ports/archive.go internal/ports/memory.go internal/adapters/fs/archive_common.go internal/adapters/fs/archive_common_test.go
git commit -m "feat(memory): archive/memory ports and shared archive helpers (#59)"
```

---

### Task 2: Zip archive adapter

**Files:**
- Create: `internal/adapters/fs/zip_archive.go`
- Test: `internal/adapters/fs/zip_archive_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapters/fs/zip_archive_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/fs/ -run TestZipArchiveRoundTrip -v`
Expected: FAIL — `undefined: ZipArchive`.

- [ ] **Step 3: Write the adapter**

Create `internal/adapters/fs/zip_archive.go`:

```go
package fs

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type ZipArchive struct{}

func (ZipArchive) WriteArchive(ctx context.Context, outPath string, files []ports.ArchiveFile) error {
	if outPath == "" {
		return errors.New("archive output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	if err := writeZipFiles(ctx, zw, files); err != nil {
		_ = zw.Close()
		_ = out.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func writeZipFiles(ctx context.Context, zw *zip.Writer, files []ports.ArchiveFile) error {
	files = sortedArchiveFiles(files)
	seen := map[string]bool{}
	for _, f := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		name, err := cleanArchivePath(f.Path)
		if err != nil {
			return err
		}
		if seen[name] {
			return fmt.Errorf("duplicate archive path %q", name)
		}
		seen[name] = true
		h := &zip.FileHeader{Name: name, Method: zip.Store}
		h.SetModTime(archiveModTime)
		h.SetMode(0o600)
		w, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		if _, err := w.Write(f.Data); err != nil {
			return err
		}
	}
	return nil
}

func (ZipArchive) ReadArchive(ctx context.Context, inPath string) ([]ports.ArchiveFile, error) {
	r, err := zip.OpenReader(inPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var out []ports.ArchiveFile
	for _, zf := range r.File {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if zf.FileInfo().IsDir() {
			continue
		}
		name, err := cleanArchivePath(zf.Name)
		if err != nil {
			return nil, err
		}
		rc, err := zf.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		out = append(out, ports.ArchiveFile{Path: name, Data: data})
	}
	return sortedArchiveFiles(out), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/fs/ -run TestZipArchiveRoundTrip -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/fs/zip_archive.go internal/adapters/fs/zip_archive_test.go
git commit -m "feat(memory): zip archive adapter (#59)"
```

---

### Task 3: Tar archive adapter

**Files:**
- Create: `internal/adapters/fs/tar_archive.go`
- Test: `internal/adapters/fs/tar_archive_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapters/fs/tar_archive_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/fs/ -run TestTarArchiveRoundTrip -v`
Expected: FAIL — `undefined: TarArchive`.

- [ ] **Step 3: Write the adapter**

Create `internal/adapters/fs/tar_archive.go`:

```go
package fs

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type TarArchive struct{}

func (TarArchive) WriteArchive(ctx context.Context, outPath string, files []ports.ArchiveFile) error {
	if outPath == "" {
		return errors.New("archive output path is required")
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(out)
	if err := writeTarFiles(ctx, tw, files); err != nil {
		_ = tw.Close()
		_ = out.Close()
		return err
	}
	if err := tw.Close(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func writeTarFiles(ctx context.Context, tw *tar.Writer, files []ports.ArchiveFile) error {
	files = sortedArchiveFiles(files)
	seen := map[string]bool{}
	for _, f := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		name, err := cleanArchivePath(f.Path)
		if err != nil {
			return err
		}
		if seen[name] {
			return fmt.Errorf("duplicate archive path %q", name)
		}
		seen[name] = true
		h := &tar.Header{Name: name, Mode: 0o600, Size: int64(len(f.Data)), ModTime: archiveModTime, Format: tar.FormatUSTAR}
		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		if _, err := tw.Write(f.Data); err != nil {
			return err
		}
	}
	return nil
}

func (TarArchive) ReadArchive(ctx context.Context, inPath string) ([]ports.ArchiveFile, error) {
	in, err := os.Open(inPath)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	tr := tar.NewReader(in)
	var out []ports.ArchiveFile
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeDir {
			continue
		}
		name, err := cleanArchivePath(h.Name)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		out = append(out, ports.ArchiveFile{Path: name, Data: data})
	}
	return sortedArchiveFiles(out), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/fs/ -run TestTarArchiveRoundTrip -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/fs/tar_archive.go internal/adapters/fs/tar_archive_test.go
git commit -m "feat(memory): tar archive adapter (#59)"
```

---

### Task 4: Memory source adapter

**Files:**
- Create: `internal/adapters/fs/memory_source.go`
- Test: `internal/adapters/fs/memory_source_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapters/fs/memory_source_test.go`:

```go
package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMemorySourceReadsAllowListedDirsOnly(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "knowledge", "etl.yaml"), "k")
	mustWrite(t, filepath.Join(root, "agents", "qa.md"), "a")
	mustWrite(t, filepath.Join(root, "input", "sessions", "raw.md"), "SECRET") // NOT allow-listed

	files, err := MemorySource{}.ReadMemory(context.Background(), root, []string{"knowledge", "agents", "registry"})
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %+v", len(files), files)
	}
	if files[0].Path != "agents/qa.md" || files[1].Path != "knowledge/etl.yaml" {
		t.Fatalf("unexpected paths: %+v", files)
	}
	for _, f := range files {
		if string(f.Data) == "SECRET" {
			t.Fatal("input/sessions must never be exported")
		}
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/fs/ -run TestMemorySource -v`
Expected: FAIL — `undefined: MemorySource`.

- [ ] **Step 3: Write the adapter**

Create `internal/adapters/fs/memory_source.go`:

```go
package fs

import (
	"context"
	"fmt"
	"os"
	archivepath "path"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/pathsafe"
	"github.com/mshegolev/bqa-os/internal/ports"
)

type MemorySource struct{}

func (MemorySource) ReadMemory(ctx context.Context, root string, allow []string) ([]ports.ArchiveFile, error) {
	var out []ports.ArchiveFile
	for _, dir := range allow {
		base := filepath.Join(root, dir)
		info, err := os.Stat(base)
		if err != nil || !info.IsDir() {
			continue // a missing allow-listed dir is simply absent
		}
		err = filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if d.IsDir() || !d.Type().IsRegular() {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			if _, ok := pathsafe.RelClean(rel); !ok {
				return fmt.Errorf("unsafe memory path %q", rel)
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			out = append(out, ports.ArchiveFile{Path: archivepath.Clean(filepath.ToSlash(rel)), Data: data})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return sortedArchiveFiles(out), nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/fs/ -run TestMemorySource -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/fs/memory_source.go internal/adapters/fs/memory_source_test.go
git commit -m "feat(memory): fs memory-source adapter (allow-list only) (#59)"
```

---

### Task 5: Auditor and installer adapters

**Files:**
- Create: `internal/adapters/fs/memory_auditor.go`
- Create: `internal/adapters/fs/memory_installer.go`
- Test: `internal/adapters/fs/memory_auditor_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapters/fs/memory_auditor_test.go`:

```go
package fs

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMemoryAuditorFlagsSecretCandidate(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "agents", "clean.md"), "just docs\n")
	mustWrite(t, filepath.Join(dir, "agents", "leak.md"), "password: hunter2\n")

	rep, err := MemoryAuditor{}.Audit(context.Background(), dir)
	if err != nil {
		t.Fatalf("audit: %v", err)
	}
	if rep.FilesScanned < 2 {
		t.Fatalf("expected >=2 scanned, got %d", rep.FilesScanned)
	}
	if rep.Candidates < 1 {
		t.Fatalf("expected the leaked secret flagged as a candidate, got %d", rep.Candidates)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/fs/ -run TestMemoryAuditor -v`
Expected: FAIL — `undefined: MemoryAuditor`.

- [ ] **Step 3: Write the adapters**

Create `internal/adapters/fs/memory_auditor.go`:

```go
package fs

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/mshegolev/bqa-os/internal/sanitize"
)

type MemoryAuditor struct{}

// Audit runs a non-destructive sanitize scan; Candidates is the number of files
// that contain redaction candidates (potential secrets/PII).
func (MemoryAuditor) Audit(ctx context.Context, dir string) (ports.AuditReport, error) {
	res, err := sanitize.Path(dir, false)
	if err != nil {
		return ports.AuditReport{}, err
	}
	return ports.AuditReport{FilesScanned: res.FilesScanned, Candidates: res.FilesChanged}, nil
}
```

Create `internal/adapters/fs/memory_installer.go`:

```go
package fs

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/brain"
	"github.com/mshegolev/bqa-os/internal/ports"
)

type MemoryInstaller struct{}

// InstallMemory delegates to the existing brain.Install so bundle placement
// keeps the same allow-list and non-destructive guarantees.
func (MemoryInstaller) InstallMemory(ctx context.Context, sourceDir, target string) (ports.InstalledSummary, error) {
	res, err := brain.Install(sourceDir, target)
	if err != nil {
		return ports.InstalledSummary{}, err
	}
	return ports.InstalledSummary{Target: res.BqaDir, Files: res.Files}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go build ./... && go test ./internal/adapters/fs/ -run TestMemoryAuditor -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/fs/memory_auditor.go internal/adapters/fs/memory_installer.go internal/adapters/fs/memory_auditor_test.go
git commit -m "feat(memory): auditor (sanitize) and installer (brain.Install) adapters (#59)"
```

---

### Task 6: Brain store adapter (github boundary)

**Files:**
- Modify: `internal/brain/brain.go` (add exported `CacheDir`)
- Create: `internal/adapters/brainstore/git_brain_store.go`
- Test: `internal/adapters/brainstore/git_brain_store_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/adapters/brainstore/git_brain_store_test.go`:

```go
package brainstore

import (
	"context"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestPushRequiresConnectedBrain(t *testing.T) {
	t.Setenv("BQA_HOME", t.TempDir()) // no config.yaml => not connected
	err := GitBrainStore{}.Push(context.Background(), []ports.ArchiveFile{{Path: "knowledge/x.yaml", Data: []byte("k")}}, false)
	if err == nil {
		t.Fatal("expected error when brain is not connected")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapters/brainstore/ -run TestPush -v`
Expected: FAIL — package/symbol undefined.

- [ ] **Step 3: Add `CacheDir` and write the adapter**

In `internal/brain/brain.go`, add:

```go
// CacheDir returns the connected brain cache directory, or an error if the
// brain is not connected (used by the github export backend).
func CacheDir() (string, error) {
	_, cacheDir, err := readConfig()
	return cacheDir, err
}
```

Create `internal/adapters/brainstore/git_brain_store.go`:

```go
package brainstore

import (
	"context"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/brain"
	"github.com/mshegolev/bqa-os/internal/ports"
)

type GitBrainStore struct{}

// Push materializes the assembled files into the connected brain cache, then
// commits and pushes via the existing brain.Sync.
func (GitBrainStore) Push(ctx context.Context, files []ports.ArchiveFile, sanitize bool) error {
	cacheDir, err := brain.CacheDir()
	if err != nil {
		return err
	}
	for _, f := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		dst := filepath.Join(cacheDir, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, f.Data, 0o600); err != nil {
			return err
		}
	}
	return brain.Sync(sanitize)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/adapters/brainstore/ -run TestPush -v`
Expected: PASS (Push returns the "brain is not connected" error before any git call).

- [ ] **Step 5: Commit**

```bash
git add internal/brain/brain.go internal/adapters/brainstore/git_brain_store.go internal/adapters/brainstore/git_brain_store_test.go
git commit -m "feat(memory): github brain-store adapter over brain.Sync (#59)"
```

---

### Task 7: Bundle metadata (manifest, checksums, docs) in core

**Files:**
- Create: `internal/core/memory/bundle.go`
- Test: `internal/core/memory/bundle_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/core/memory/bundle_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/memory/ -run 'TestAssembleBundle|TestVerifyChecksums|TestParseManifest|TestManifestOmits' -v`
Expected: FAIL — undefined symbols.

- [ ] **Step 3: Write bundle.go**

Create `internal/core/memory/bundle.go`:

```go
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/memory/ -run 'TestAssembleBundle|TestVerifyChecksums|TestParseManifest|TestManifestOmits' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/memory/bundle.go internal/core/memory/bundle_test.go
git commit -m "feat(memory): bundle manifest/checksums/docs assembly + verify (#59)"
```

---

### Task 8: Export use case

**Files:**
- Create: `internal/core/memory/export.go`
- Create: `internal/core/memory/usecase.go`
- Test: `internal/core/memory/export_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/core/memory/export_test.go`:

```go
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
```


- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/memory/ -run TestExport -v`
Expected: FAIL — `UseCase`/`ExportOptions` undefined.

- [ ] **Step 3: Write usecase.go and export.go**

Create `internal/core/memory/usecase.go`:

```go
package memory

import "github.com/mshegolev/bqa-os/internal/ports"

// UseCase orchestrates memory export/import over ports only.
type UseCase struct {
	Source    ports.MemorySource
	Auditor   ports.MemoryAuditor
	Installer ports.MemoryInstaller
	Brain     ports.BrainStore
	Writers   map[string]ports.ArchiveWriter // keyed "zip","tar"
	Readers   map[string]ports.ArchiveReader // keyed "zip","tar"
}
```

Create `internal/core/memory/export.go`:

```go
package memory

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type ExportOptions struct {
	SourceRoot string   // default ".bqa"
	Target     string   // "zip" | "tar" | "github"
	OutPath    string   // required for zip/tar
	Exclude    []string // extra glob patterns (matched against bundle paths)
	DryRun     bool
	Strict     bool
}

type ExportResult struct {
	Target  string
	OutPath string
	Files   []string
	Audit   ports.AuditReport
	DryRun  bool
}

func (u UseCase) Export(ctx context.Context, opts ExportOptions) (ExportResult, error) {
	root := opts.SourceRoot
	if root == "" {
		root = ".bqa"
	}
	payload, err := u.Source.ReadMemory(ctx, root, AllowList)
	if err != nil {
		return ExportResult{}, err
	}
	payload, err = applyExcludes(payload, opts.Exclude)
	if err != nil {
		return ExportResult{}, err
	}
	if len(payload) == 0 {
		return ExportResult{}, fmt.Errorf("nothing to export from %s (run `bqa build` first)", root)
	}

	stage, err := os.MkdirTemp("", "bqa-export-*")
	if err != nil {
		return ExportResult{}, err
	}
	defer os.RemoveAll(stage)
	if err := writeFilesToDir(stage, payload); err != nil {
		return ExportResult{}, err
	}
	audit, err := u.Auditor.Audit(ctx, stage)
	if err != nil {
		return ExportResult{}, err
	}
	if opts.Strict && audit.Candidates > 0 {
		return ExportResult{}, fmt.Errorf("export blocked by --strict: audit flagged %d file(s) with redaction candidates", audit.Candidates)
	}

	res := ExportResult{Target: opts.Target, OutPath: opts.OutPath, Files: pathsOf(payload), Audit: audit, DryRun: opts.DryRun}
	if opts.DryRun {
		return res, nil
	}

	switch opts.Target {
	case "zip", "tar":
		if opts.OutPath == "" {
			return ExportResult{}, errors.New("--out is required for zip/tar export")
		}
		w, ok := u.Writers[opts.Target]
		if !ok {
			return ExportResult{}, fmt.Errorf("no archive writer for target %q", opts.Target)
		}
		bundle, err := assembleBundle(payload, audit)
		if err != nil {
			return ExportResult{}, err
		}
		if err := w.WriteArchive(ctx, opts.OutPath, bundle); err != nil {
			return ExportResult{}, err
		}
	case "github":
		if u.Brain == nil {
			return ExportResult{}, errors.New("github target is not configured")
		}
		if err := u.Brain.Push(ctx, payload, true); err != nil {
			return ExportResult{}, err
		}
	default:
		return ExportResult{}, fmt.Errorf("unknown target %q (want zip|tar|github)", opts.Target)
	}
	return res, nil
}

func applyExcludes(files []ports.ArchiveFile, patterns []string) ([]ports.ArchiveFile, error) {
	if len(patterns) == 0 {
		return files, nil
	}
	var out []ports.ArchiveFile
	for _, f := range files {
		drop := false
		for _, pat := range patterns {
			ok, err := path.Match(pat, f.Path)
			if err != nil {
				return nil, fmt.Errorf("invalid --exclude pattern %q: %w", pat, err)
			}
			if ok {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, f)
		}
	}
	return out, nil
}

func writeFilesToDir(dir string, files []ports.ArchiveFile) error {
	for _, f := range files {
		dst := filepath.Join(dir, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, f.Data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func pathsOf(files []ports.ArchiveFile) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.Path)
	}
	return out
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/memory/ -run TestExport -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/core/memory/usecase.go internal/core/memory/export.go internal/core/memory/export_test.go
git commit -m "feat(memory): export use case (zip/tar/github, dry-run, strict, exclude) (#59)"
```

---

### Task 9: Import use case

**Files:**
- Create: `internal/core/memory/import.go`
- Test: `internal/core/memory/import_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/core/memory/import_test.go`:

```go
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
		"knowledge/etl.yaml": "schema_version: 1\n",
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/core/memory/ -run TestImport -v`
Expected: FAIL — `ImportOptions`/`Import` undefined.

- [ ] **Step 3: Write import.go**

Create `internal/core/memory/import.go`:

```go
package memory

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type ImportOptions struct {
	FromPath string
	Target   string
	DryRun   bool
}

type ImportResult struct {
	Verified  bool
	Installed ports.InstalledSummary
	DryRun    bool
}

func (u UseCase) Import(ctx context.Context, opts ImportOptions) (ImportResult, error) {
	reader, err := u.readerFor(opts.FromPath)
	if err != nil {
		return ImportResult{}, err
	}
	files, err := reader.ReadArchive(ctx, opts.FromPath)
	if err != nil {
		return ImportResult{}, err
	}
	index := map[string][]byte{}
	for _, f := range files {
		index[f.Path] = f.Data
	}

	mdata, ok := index["manifest.json"]
	if !ok {
		return ImportResult{}, errors.New("bundle is missing manifest.json")
	}
	manifest, err := parseManifest(mdata)
	if err != nil {
		return ImportResult{}, err
	}
	if manifest.BundleVersion != BundleVersion {
		return ImportResult{}, fmt.Errorf("unsupported bundle_version %d (want %d)", manifest.BundleVersion, BundleVersion)
	}

	cdata, ok := index["metadata/checksums.json"]
	if !ok {
		return ImportResult{}, errors.New("bundle is missing metadata/checksums.json")
	}
	checksums, err := parseChecksums(cdata)
	if err != nil {
		return ImportResult{}, err
	}

	payload := payloadFiles(files)
	if err := verifyChecksums(payload, checksums); err != nil {
		return ImportResult{}, err
	}

	res := ImportResult{Verified: true, DryRun: opts.DryRun}
	if opts.DryRun {
		return res, nil
	}

	stage, err := os.MkdirTemp("", "bqa-import-*")
	if err != nil {
		return ImportResult{}, err
	}
	defer os.RemoveAll(stage)
	if err := writeFilesToDir(stage, payload); err != nil {
		return ImportResult{}, err
	}
	summary, err := u.Installer.InstallMemory(ctx, stage, opts.Target)
	if err != nil {
		return ImportResult{}, err
	}
	res.Installed = summary
	return res, nil
}

// readerFor selects the archive reader by file extension.
func (u UseCase) readerFor(fromPath string) (ports.ArchiveReader, error) {
	switch {
	case strings.HasSuffix(fromPath, ".zip"):
		if r, ok := u.Readers["zip"]; ok {
			return r, nil
		}
	case strings.HasSuffix(fromPath, ".tar"):
		if r, ok := u.Readers["tar"]; ok {
			return r, nil
		}
	}
	return nil, fmt.Errorf("unsupported bundle format for %q (want .zip or .tar)", fromPath)
}

// payloadFiles returns only the allow-listed payload (drops manifest, metadata, docs).
func payloadFiles(files []ports.ArchiveFile) []ports.ArchiveFile {
	allow := map[string]bool{}
	for _, d := range AllowList {
		allow[d] = true
	}
	var out []ports.ArchiveFile
	for _, f := range files {
		i := strings.IndexByte(f.Path, '/')
		if i <= 0 {
			continue // manifest.json, README.md, etc.
		}
		if allow[f.Path[:i]] {
			out = append(out, f)
		}
	}
	return out
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/core/memory/ -run TestImport -v && go test ./...`
Expected: PASS (all packages).

- [ ] **Step 5: Commit**

```bash
git add internal/core/memory/import.go internal/core/memory/import_test.go
git commit -m "feat(memory): import use case (verify manifest+checksums, then install) (#59)"
```

---

### Task 10: CLI wiring — `bqa brain export` / `import`

**Files:**
- Modify: `internal/app/brain.go`
- Test: `internal/app/brain_memory_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/app/brain_memory_test.go`:

```go
package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBrainExportThenImportRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	bqa := filepath.Join(tmp, ".bqa")
	for rel, content := range map[string]string{
		"knowledge/etl.yaml":  "schema_version: 1\n",
		"registry/index.yaml": "registry:\n  version: 1\n",
	} {
		p := filepath.Join(bqa, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	bundle := filepath.Join(tmp, "bundle.zip")

	// export
	exp := brainCmd()
	var out bytes.Buffer
	exp.SetOut(&out)
	exp.SetErr(&out)
	exp.SetArgs([]string{"export", "--source", bqa, "--target", "zip", "--out", bundle})
	if err := exp.Execute(); err != nil {
		t.Fatalf("export: %v\n%s", err, out.String())
	}
	if _, err := os.Stat(bundle); err != nil {
		t.Fatalf("bundle not written: %v", err)
	}

	// import
	target := filepath.Join(tmp, "client")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	imp := brainCmd()
	var out2 bytes.Buffer
	imp.SetOut(&out2)
	imp.SetErr(&out2)
	imp.SetArgs([]string{"import", "--from", bundle, "--target", target})
	if err := imp.Execute(); err != nil {
		t.Fatalf("import: %v\n%s", err, out2.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".bqa", "knowledge", "etl.yaml")); err != nil {
		t.Fatalf("expected installed file: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/app/ -run TestBrainExportThenImport -v`
Expected: FAIL — unknown command "export".

- [ ] **Step 3: Add the subcommands**

In `internal/app/brain.go`, add imports for `context`, the fs adapters, the brainstore adapter, the memory core, and the ports package. Then add, before `return cmd`:

```go
	memoryUseCase := func() coreMemoryUseCase {
		return coreMemoryUseCase{
			Source:    fsadapter.MemorySource{},
			Auditor:   fsadapter.MemoryAuditor{},
			Installer: fsadapter.MemoryInstaller{},
			Brain:     brainstore.GitBrainStore{},
			Writers:   map[string]ports.ArchiveWriter{"zip": fsadapter.ZipArchive{}, "tar": fsadapter.TarArchive{}},
			Readers:   map[string]ports.ArchiveReader{"zip": fsadapter.ZipArchive{}, "tar": fsadapter.TarArchive{}},
		}
	}

	var expSource, expTarget, expOut string
	var expDryRun, expStrict bool
	var expExclude []string
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export project memory to a zip/tar bundle or the GitHub brain",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := memoryUseCase().Export(cmd.Context(), memory.ExportOptions{
				SourceRoot: expSource, Target: expTarget, OutPath: expOut,
				Exclude: expExclude, DryRun: expDryRun, Strict: expStrict,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Target: %s\n", res.Target)
			fmt.Fprintf(out, "Files: %d\n", len(res.Files))
			fmt.Fprintf(out, "Audit: scanned=%d candidates=%d\n", res.Audit.FilesScanned, res.Audit.Candidates)
			if res.DryRun {
				fmt.Fprintln(out, "Dry run — no archive written. Planned files:")
				for _, f := range res.Files {
					fmt.Fprintf(out, "  %s\n", f)
				}
				return nil
			}
			if res.OutPath != "" {
				fmt.Fprintf(out, "Wrote: %s\n", res.OutPath)
			}
			return nil
		},
	}
	exportCmd.Flags().StringVar(&expSource, "source", ".bqa", "source memory directory")
	exportCmd.Flags().StringVar(&expTarget, "target", "zip", "target: zip|tar|github")
	exportCmd.Flags().StringVar(&expOut, "out", "", "output archive path (required for zip/tar)")
	exportCmd.Flags().BoolVar(&expDryRun, "dry-run", false, "print planned files without writing")
	exportCmd.Flags().BoolVar(&expStrict, "strict", false, "abort if the audit finds redaction candidates")
	exportCmd.Flags().StringArrayVar(&expExclude, "exclude", nil, "glob patterns to exclude (repeatable)")
	cmd.AddCommand(exportCmd)

	var impFrom, impTarget string
	var impDryRun bool
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import a memory bundle (verifies manifest + checksums, then installs)",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := memoryUseCase().Import(cmd.Context(), memory.ImportOptions{FromPath: impFrom, Target: impTarget, DryRun: impDryRun})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Verified: %v\n", res.Verified)
			if res.DryRun {
				fmt.Fprintln(out, "Dry run — nothing installed.")
				return nil
			}
			fmt.Fprintf(out, "Installed into: %s (%d files)\n", res.Installed.Target, len(res.Installed.Files))
			return nil
		},
	}
	importCmd.Flags().StringVar(&impFrom, "from", "", "bundle path (.zip or .tar, required)")
	importCmd.Flags().StringVar(&impTarget, "target", "", "target project directory (required)")
	importCmd.Flags().BoolVar(&impDryRun, "dry-run", false, "verify only; do not install")
	_ = importCmd.MarkFlagRequired("from")
	_ = importCmd.MarkFlagRequired("target")
	cmd.AddCommand(importCmd)
```

Add the imports and a type alias at the top of `brain.go`:

```go
import (
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/adapters/brainstore"
	"github.com/mshegolev/bqa-os/internal/brain"
	"github.com/mshegolev/bqa-os/internal/core/memory"
	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/spf13/cobra"
)

// coreMemoryUseCase aliases the memory use case for brevity in wiring.
type coreMemoryUseCase = memory.UseCase
```

(If `brain` is already imported for `install`, keep a single import.)

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/app/ -run TestBrainExportThenImport -v && go test ./...`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/app/brain.go internal/app/brain_memory_test.go
git commit -m "feat(memory): wire bqa brain export/import commands (#59)"
```

---

### Task 11: Docs and final verification

**Files:**
- Create: `docs/memory-bundles.md`

- [ ] **Step 1: Write the doc**

Create `docs/memory-bundles.md` covering: what a memory bundle is; the allow-list (and that `input/sessions` is never exported); `bqa brain export --target zip|tar|github` with `--out`, `--dry-run`, `--strict`, `--exclude`; the bundle layout (manifest.json, metadata/checksums.json, audit.md, allow-listed dirs); `bqa brain import --from <bundle> --target <dir>` (verifies manifest + checksums before installing, reuses the non-destructive brain install); and the github flow (`bqa brain connect` then `bqa brain export --target github`). Use synthetic examples only.

- [ ] **Step 2: Full verification**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: build clean, vet clean, all packages `ok`.

- [ ] **Step 3: Manual smoke (optional)**

```bash
go run ./cmd/bqa brain export --source .bqa --target zip --out /tmp/bqa-memory.zip --dry-run
```
Expected: prints planned files + audit summary; writes nothing.

- [ ] **Step 4: Commit and open PR**

```bash
git add docs/memory-bundles.md
git commit -m "docs: memory export/import bundles guide (#59)"
git push -u origin feature/issue-59-memory-backends
gh pr create --title "feat: memory export/import backends — zip/tar bundles + github (#59)" --body "Implements docs/superpowers/specs/2026-07-01-memory-export-import-backends-design.md. Closes #59."
```

---

## Notes for the implementer

- **DRY:** import placement reuses `brain.Install`; github reuses `brain.Sync`; audit reuses `sanitize.Path`; archive helpers (`cleanArchivePath`, `sortedArchiveFiles`, `archiveModTime`) are shared by the zip and tar adapters.
- **YAGNI:** only one archive port with two adapters; `.bqa/output/runtime/`, `.tar.gz`, and a richer github flow are deferred (see the spec's Out-of-scope).
- **Determinism:** archives use a fixed mod-time; `manifest.json` has no wall-clock field; `json.MarshalIndent` sorts map keys — bundles are byte-reproducible for the same input.
- **Safety:** the allow-list is the primary guard (never `input/sessions`); the audit + `--strict` are the secondary guard; import verifies checksums before writing anything to the target.
