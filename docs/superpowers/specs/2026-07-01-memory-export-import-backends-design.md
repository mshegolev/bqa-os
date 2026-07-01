# Design: Memory export/import backends — GitHub brain + archive bundles (issue #59)

## Context

BQA-OS generates project memory under `.bqa/` (knowledge, agents, skills,
workflows, prompts, registry). Users need to **back it up, move it between
machines, review it, and share a sanitized package with a customer** — via a
portable archive (zip/tar) or a private GitHub "brain" repo. Issue #59 asks for
a hexagonal architecture and an MVP vertical slice.

This builds directly on two things already in the repo:

- `bqa brain connect/pull/sync/status` (git-backed brain cache; `sync` commits +
  pushes, with optional `sanitize`) — the GitHub backend already largely exists.
- `bqa brain install --from <dir> --target <dir>` (issue #41) — copies an
  allow-listed set of dirs (`knowledge, agents, skills, workflows, prompts,
  registry`) into `<target>/.bqa/`, never touching unrelated files.
- The **v1 knowledge schema** (issue #16) — exported knowledge is versioned and
  checksummable.
- `internal/adapters/fs/demo_archive_writer.go` — a working `archive/zip`
  writer (the model for the new archive adapters).
- `internal/sanitize` — `sanitize.Path(root, write)` scans/redacts a tree.

**Constraint:** stdlib + cobra only (no third-party archive/YAML libs). Output
must be **deterministic** (no wall-clock timestamps) so bundles diff cleanly and
tests can compare bytes.

Decisions taken during brainstorming:

1. **MVP slice:** zip **and** tar are both fully implemented behind one archive
   port (two adapters). GitHub is a boundary that reuses the existing brain sync.
2. **Command surface:** everything lives under the existing `bqa brain` group —
   new `export` and `import` subcommands (alongside connect/pull/sync/status/install).
3. **Safety:** export only an **allow-list** of safe dirs (never
   `.bqa/input/sessions`), run an **audit** (sanitize scan) before packaging, and
   a `--strict` flag aborts the export if the audit finds redaction candidates.
4. **Import:** unpack → verify manifest + checksums → **reuse `brain.Install`**
   (via a port) to place files into `<target>/.bqa/`. Nothing is written to the
   target unless verification passes.
5. **Strict hexagonal:** the use case lives in `internal/core/memory/` and
   depends only on ports + stdlib; every external boundary (archive, filesystem
   read, placement, audit, github) is a port with an adapter. The legacy
   `internal/brain` package (which shells out to git inline) is not extended;
   adapters delegate to its existing functions where useful.

## Architecture (strict hexagonal)

```
internal/app/brain.go        (CLI wiring: build adapters, call core)
        ↓ depends on
internal/core/memory/        (Export/Import use cases — ports + stdlib only)
        ↓ depends on
internal/ports/{archive,memory}.go   (interfaces)
        ↑ implemented by
internal/adapters/fs/*        (zip/tar archive, memory source, installer, auditor)
internal/adapters/brainstore/* (github/git brain store)
```

Pure logic (SHA-256 checksums, manifest JSON build/parse, allow-list staging
decisions) lives in `core/memory` as plain functions — it has no external
dependency, so it does not need a port.

### Ports

`internal/ports/archive.go`:

```go
type ArchiveFile struct {
	Path string // bundle-relative path, forward slashes
	Data []byte
}

type ArchiveWriter interface {
	WriteArchive(ctx context.Context, outPath string, files []ArchiveFile) error
}

type ArchiveReader interface {
	ReadArchive(ctx context.Context, inPath string) ([]ArchiveFile, error)
}
```

`internal/ports/memory.go`:

```go
type MemorySource interface {
	// ReadMemory returns allow-listed files under root (e.g. ".bqa"), each with a
	// bundle-relative path. Missing allow-listed dirs are simply absent.
	ReadMemory(ctx context.Context, root string, allow []string) ([]ArchiveFile, error)
}

type InstalledSummary struct {
	Target string
	Files  []string
}

type MemoryInstaller interface {
	// InstallMemory places a verified bundle directory into <target>/.bqa.
	InstallMemory(ctx context.Context, sourceDir, target string) (InstalledSummary, error)
}

type AuditReport struct {
	FilesScanned int
	Candidates   int      // files that would be redacted (potential secrets)
	Details      []string // human-readable per-file notes
}

type MemoryAuditor interface {
	Audit(ctx context.Context, dir string) (AuditReport, error)
}

type BrainStore interface {
	// Push writes the assembled files into the connected brain cache and syncs
	// (commit + push). sanitize toggles a redaction pass before commit.
	Push(ctx context.Context, files []ArchiveFile, sanitize bool) error
}
```

### Adapters

- `internal/adapters/fs/zip_archive.go` — `ArchiveWriter`/`ArchiveReader` via
  `archive/zip` (modelled on `demo_archive_writer.go`: deterministic, sorted,
  dir-safe).
- `internal/adapters/fs/tar_archive.go` — same via `archive/tar`.
- `internal/adapters/fs/memory_source.go` — `MemorySource`: walk `<root>/<dir>`
  for each allow-listed dir, return files with bundle-relative paths, using
  `pathsafe.RelClean` for safety.
- `internal/adapters/fs/memory_installer.go` — `MemoryInstaller`: delegates to
  the existing `brain.Install(sourceDir, target)` (DRY; same allow-list +
  non-destructive guarantees).
- `internal/adapters/fs/memory_auditor.go` — `MemoryAuditor`: wraps
  `sanitize.Path(dir, false)`; `Candidates = Result.FilesChanged` (scan mode).
- `internal/adapters/brainstore/git_brain_store.go` — `BrainStore`: reads the
  brain cache dir from the existing brain config, materializes the files there,
  and delegates to `brain.Sync(sanitize)`.

### Core use cases (`internal/core/memory/`)

`export.go`:

```go
type ExportOptions struct {
	SourceRoot string   // default ".bqa"
	Target     string   // "zip" | "tar" | "github"
	OutPath    string   // archive output path (zip/tar)
	Exclude    []string // extra glob patterns to drop
	DryRun     bool
	Strict     bool     // abort if the audit finds redaction candidates
}

type ExportResult struct {
	Target   string
	OutPath  string
	Files    []string    // bundle-relative payload paths (sorted)
	Audit    AuditReport
	DryRun   bool
}

func (u UseCase) Export(ctx context.Context, opts ExportOptions) (ExportResult, error)
```

Flow: `MemorySource.ReadMemory(SourceRoot, allowList)` → apply `Exclude` globs →
error if empty ("nothing to export; run `bqa build`") → stage to a temp dir →
`MemoryAuditor.Audit(temp)` → if `Strict && Candidates>0` abort → build
`manifest.json` + `metadata/checksums.json` + `README.md`/`install.md`/`audit.md`
+ `metadata/memory_audit.yaml` (all pure) → if `DryRun` print plan and stop →
else for zip/tar call `ArchiveWriter.WriteArchive`, for github call
`BrainStore.Push(files, sanitize=true)`.

`import.go`:

```go
type ImportOptions struct {
	FromPath string
	Target   string
	DryRun   bool
}

type ImportResult struct {
	Verified  bool
	Installed InstalledSummary
	DryRun    bool
}

func (u UseCase) Import(ctx context.Context, opts ImportOptions) (ImportResult, error)
```

Flow: pick `ArchiveReader` by extension (`.zip`/`.tar`) → `ReadArchive` → verify
`manifest.json` present and `bundle_version == 1` → recompute SHA-256 of every
payload file and compare to `checksums.json`; any missing/mismatch → error,
**nothing written** → if `DryRun` print planned files + "verified" and stop →
else materialize verified allow-listed dirs to a temp dir →
`MemoryInstaller.InstallMemory(temp, Target)`.

`bundle.go`: the `Manifest`/`Checksums` structs and pure build/verify helpers,
plus the allow-list constant and the generated `README.md`/`install.md` text.

### CLI (`internal/app/brain.go`)

```
bqa brain export --target zip|tar|github [--out <path>] [--source .bqa] [--dry-run] [--strict] [--exclude <glob> ...]
bqa brain import --from <bundle.zip|.tar> --target <dir> [--dry-run]
```

The command builds the concrete adapters and calls `memory.UseCase`. `--out` is
required for zip/tar; ignored for github.

## Bundle format

```text
manifest.json
README.md
install.md
audit.md
knowledge/  agents/  skills/  workflows/  prompts/  registry/
metadata/
  memory_audit.yaml
  checksums.json
```

**Allow-list (hard):** `knowledge, agents, skills, workflows, prompts, registry`
— the same set `brain.Install` understands. **Never** `input/sessions`, caches,
or `.git`. `.bqa/output/runtime/` is **out of MVP scope** (not in the install
allow-list; keeping it out keeps import↔install DRY) — documented as a next step.

`manifest.json` (deterministic — no wall-clock):

```json
{
  "bundle_version": 1,
  "tool": "bqa",
  "generated_by": "bqa dev",
  "included": ["knowledge", "agents", "skills", "workflows", "prompts", "registry"],
  "file_count": 12
}
```

`metadata/checksums.json`:

```json
{ "sha256": { "knowledge/etl_patterns.yaml": "<hex>", "agents/etl-qa-agent.md": "<hex>" } }
```

Covers every payload file; `manifest.json` and the files under `metadata/` are
not themselves checksummed.

## Safety

- **Allow-list** is the primary guard; `input/sessions` is never staged.
- **Audit** before packaging: `MemoryAuditor.Audit(stagedTemp)` runs a
  sanitize scan; results go to human-readable `audit.md` and
  `metadata/memory_audit.yaml` (`files_scanned`, `candidates`).
- `--strict`: if `Candidates > 0`, abort with a clear error and write nothing.
- `--exclude <glob>` (repeatable): drop extra paths from the staged set.
- `--dry-run`: print the planned file list + audit summary; write no archive.
- Missing allow-listed dirs / empty staged set → a clear actionable error.

## Testing

- **Archive round-trip (zip and tar):** `WriteArchive` then `ReadArchive`
  returns identical paths + bytes; output is deterministic across runs.
- **Export (zip and tar):** synthetic `.bqa` → archive contains `manifest.json`,
  `metadata/checksums.json`, and the allow-listed dirs; `input/sessions` is
  absent; `--dry-run` writes no file; `--strict` aborts when a planted secret
  marker is present in a staged file.
- **Import:** a good bundle installs into `<target>/.bqa`; a tampered checksum →
  error and nothing written to the target; a missing/unsupported `manifest.json`
  → error.
- **Errors:** missing `.bqa` allow-listed dirs → clear actionable error.
- `go test ./...` passes.

## Out of scope (MVP, documented next steps)

- `.bqa/output/runtime/` in the bundle (would need a wider install allow-list).
- Full `--target github` beyond assemble-into-cache + `brain.Sync` (the actual
  push path is already covered by `brain.Sync`).
- gzip-compressed tar (`.tar.gz`).
- Refactoring the legacy `internal/brain` git logic into ports (adapters wrap it
  as-is).
