# Cleanup: dead code and duplicate helpers (bqa-os)

Date: 2026-06-30
Status: approved design, pending implementation
Scope: bqa-os only (Go engine)

## Context

A project audit found low-risk dead code and duplicated helpers in `bqa-os`:

- `internal/app/ingest2.go` is a functional duplicate of `internal/app/ingest.go`
  (same `coreingest.UseCase`, same flags, same output; differs only in command
  name and `Short`). Repository docs (`README.md:307`, `AGENTS.md:71`,
  `docs/knowledge-extractor.md:9,27`) standardize on `bqa ingest`; `ingest2`
  appears only in non-authoritative runtime artifacts.
- `internal/ingest/` is an abandoned pre-hexagonal package with **zero
  importers** (superseded by `internal/core/ingest` + `internal/adapters/fs`).
- Small helpers are copy-pasted across packages: a relative-path safety guard,
  `hasAny`, and a YAML value quoter.

This is the "remove dead code / duplicates" cleanup (Approach C of the audit):
deletions plus consolidation of the *genuinely identical* helpers, while
preserving all external behavior.

## Goals

- Remove dead and duplicate code without changing any CLI behavior.
- Consolidate only the helper duplications that are byte-for-byte equivalent, so
  consolidation cannot alter behavior.
- Keep `go build ./...` and `go test ./...` green.

## Non-goals

- No registry/source-of-truth refactor (separate, larger effort).
- No removal of `RuntimeArtifact.Destination`/`.Tags` fields (Approach B, out of
  scope here).
- No folding of the two *divergent* path guards (see Design → Path safety).

## Design

### 1. Deletions

- Delete `internal/app/ingest2.go`.
- Remove `rootCmd.AddCommand(ingest2Cmd())` from `internal/app/root.go`.
- Delete the entire `internal/ingest/` package (confirmed zero importers via
  `grep -rn 'internal/ingest"'`). This also removes the duplicate `looksBinary`
  (the surviving copy lives in `internal/sanitize/sanitize.go`).

`bqa ingest` is kept unchanged. No documentation changes are required because the
authoritative docs already reference `ingest`, not `ingest2`.

### 2. New leaf utility packages

Pure functions shared across `core` and `adapters`; no ports needed.

- `internal/textutil`
  - `HasAny(text string, needles ...string) bool` — substring-any check.
  - `QuoteYAML(value string) string` — escape `\` and `"`, wrap in double quotes.
- `internal/pathsafe`
  - `RelClean(path string) (string, error)` — `filepath.Clean` then reject
    absolute paths, `".."`, and `".."+separator` prefixes. Mirrors the exact
    current rejection set of the identical pair below.

### 3. Call-site rewrites (behavior preserved exactly)

- `internal/core/etlpack/usecase.go` `hasAny` and
  `internal/core/knowledge/usecase.go` `hasAny` → `textutil.HasAny`
  (the two are identical apart from a parameter name).
- `internal/core/knowledge/usecase.go` `yamlString` → `textutil.QuoteYAML`
  (identical escaping logic).
- `internal/core/runtimeemit/usecase.go` `yamlQuote` → 
  `textutil.QuoteYAML(strings.ReplaceAll(value, "\n", " "))`. The newline→space
  collapse, which is unique to `yamlQuote`, is applied *before* the shared
  quoter, so output is byte-identical to today.
- `internal/adapters/registry/reader.go` and
  `internal/adapters/fs/runtime_store.go` path guards → `pathsafe.RelClean`.
  These two guards are currently identical, so the substitution is exact.

### 4. Intentionally left unchanged (behavior risk)

- `internal/adapters/fs/knowledge_store.go` `normalizedSessionPath`: its guard is
  embedded in larger logic (allows absolute paths, joins with a base dir, and
  additionally rejects `"."`). Different shape → not folded.
- `internal/adapters/fs/demo_archive_writer.go` `cleanDemoArchivePath`: uses
  archive `path` semantics (slash normalization, `"../"` and leading `"/"`
  rejection). Different shape → not folded.

Folding these would change their rejection sets; preserving them is the
lower-risk choice and the explicit boundary of this cleanup.

## Affected files

Deleted:
- `internal/app/ingest2.go`
- `internal/ingest/` (entire package)

Added:
- `internal/textutil/textutil.go` (+ `textutil_test.go`)
- `internal/pathsafe/pathsafe.go` (+ `pathsafe_test.go`)

Modified:
- `internal/app/root.go` (drop `ingest2Cmd()` registration)
- `internal/core/etlpack/usecase.go`
- `internal/core/knowledge/usecase.go`
- `internal/core/runtimeemit/usecase.go`
- `internal/adapters/registry/reader.go`
- `internal/adapters/fs/runtime_store.go`

## Testing & verification

- Unit tests for `textutil.HasAny`, `textutil.QuoteYAML`, `pathsafe.RelClean`
  (including the rejection set: absolute, `..`, `../x`).
- `go build ./...` succeeds.
- `go test ./...` stays green (existing emit/knowledge/etlpack tests exercise the
  rewired helpers).
- `go vet ./...` clean.
- Manual: `bqa ingest --help` works; `bqa ingest2` no longer exists.

## Risks

- Low. All changes are deletions of unreachable code or substitutions with
  byte-equivalent helpers. The divergent path guards are explicitly excluded.
- Only residual risk: an unnoticed importer of `internal/ingest` — mitigated by
  the grep check and `go build ./...`.
