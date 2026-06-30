# Hexagonalize `internal/runtime` and de-duplicate runtime commands (bqa-os)

Date: 2026-06-30
Status: approved design, pending implementation
Scope: bqa-os only. Behavior-preserving (audit thread #3, option "hexagon + dedup").

## Context

`internal/runtime/runtime.go` is a flat package that mixes content building,
filesystem IO, process detection, and stdout printing — unlike the rest of
bqa-os, which follows ports → core → adapters → app. It also contains
duplication:

- `claude`/`codex`/`opencode` commands are three near-identical files that each
  call `runtime.Prepare(name)`; the only behavioral difference is the runtime
  name interpolated into one sentence of the master context.
- `InstallCommands()` re-writes the master context (duplicating `Prepare`'s
  write) and then writes the `bqa-master` command into four locations.
- The package performs `os.WriteFile`/`os.MkdirAll` and `exec.LookPath`
  directly, with no port boundary, and prints with `fmt.Printf` inside the
  business logic.

This refactor moves runtime logic into the hexagonal structure and removes the
duplication, **without changing any external behavior** (same files, same bytes,
same stdout messages). The "build runtime on top of emit" option was declined;
`emit` and `internal/catalog` are untouched.

## Goals

- Runtime logic lives in `internal/core/runtime` behind ports; IO and process
  detection live in adapters; printing lives in the app layer.
- Remove the duplicated master-context write and collapse the three near-
  identical runtime commands into one factory.
- Preserve every output path, file content (byte-for-byte), and stdout message.

## Non-goals

- No coupling of runtime commands to `emit`/registry (declined option).
- No change to `internal/catalog`, `internal/core/runtimeemit`, or the
  `.claude/commands/` ↔ emit relationship.
- No change to the master-context or `/bqa-master` command text.

## Design

### Package structure

- **`internal/core/runtime`** — the use case and pure content builders:
  ```go
  type UseCase struct {
      Writer   ports.RuntimeArtifactWriter // reused from emit
      Detector ports.RuntimeDetector       // new port
  }

  type PrepareResult struct {
      Runtime     string
      ContextPath string
      Detected    bool
      BinaryPath  string
      Command     string
  }
  type InstallResult struct {
      ContextPath string
      Commands    []string // paths written
  }
  type RuntimeStatus struct {
      Name       string
      Found      bool
      BinaryPath string
  }

  func (u UseCase) Prepare(ctx context.Context, runtime string) (PrepareResult, error)
  func (u UseCase) InstallCommands(ctx context.Context) (InstallResult, error)
  func (u UseCase) Detect(ctx context.Context) ([]RuntimeStatus, error)
  ```
  `buildContext(runtimeName)` and `buildMasterCommand()` move here verbatim. A
  single internal `writeMasterContext` step is reused by both `Prepare` and
  `InstallCommands`. The use case returns results and does **no** printing.

- **`internal/ports/runtime_setup.go`** — new port:
  ```go
  type RuntimeDetector interface {
      Detect(binary string) (path string, found bool)
  }
  ```
  Writing reuses the existing `ports.RuntimeArtifactWriter`
  (`WriteRuntimeArtifact(ctx, relativePath, content)`).

- **`internal/adapters/runtimebin`** — `Detector` implementing `RuntimeDetector`
  via `exec.LookPath`.

- **Writing** reuses `internal/adapters/fs.RuntimeStore{TargetDir: "."}`
  (already validates relative paths and writes project-relative files).

- **`internal/runtime`** (old flat package) is **deleted**; its tests are moved
  to `internal/core/runtime`.

### App layer

- A single factory `runtimeContextCmd(name string) *cobra.Command` replaces
  `claude.go`, `codex.go`, `opencode.go`. `root.go` registers three commands
  from it (`runtimeContextCmd("claude")`, etc.). Each builds the use case,
  calls `Prepare`, and prints the same messages as today.
- `internal/app/runtime.go` (`runtime detect` / `runtime install-commands`)
  builds the use case, calls `Detect` / `InstallCommands`, and prints today's
  messages from the returned results.

### Behavior preservation (exact)

- Output paths unchanged: `.bqa/prompts/bqa-master-context.md`,
  `.bqa/runtime-commands/{codex,claude,opencode}/bqa-master.md`,
  `.claude/commands/bqa-master.md`.
- File contents byte-identical, including the runtime-name interpolation
  (`Prepare` uses the runtime name; `InstallCommands` uses
  `"project-local command"`).
- stdout messages identical (same strings, same order), now emitted from app.

## Affected files

Added:
- `internal/core/runtime/usecase.go`
- `internal/core/runtime/usecase_test.go` (moved/expanded from the old test)
- `internal/ports/runtime_setup.go`
- `internal/adapters/runtimebin/detector.go`
- `internal/app/runtime_context.go` (the `runtimeContextCmd` factory)

Deleted:
- `internal/runtime/runtime.go`
- `internal/runtime/runtime_test.go`
- `internal/app/claude.go`, `internal/app/codex.go`, `internal/app/opencode.go`

Modified:
- `internal/app/root.go` (register three commands from the factory; unchanged for
  the others)
- `internal/app/runtime.go` (wire use case + adapters; print from results)

## Testing & verification

- `internal/core/runtime` unit tests with a fake `RuntimeArtifactWriter` and a
  fake `RuntimeDetector`:
  - `Prepare("claude")` writes exactly `.bqa/prompts/bqa-master-context.md` with
    content containing "through the claude runtime"; result reports detection.
  - `InstallCommands` writes the master context plus the four command files; the
    command content equals the current `buildMasterCommand()` output (golden).
  - `Detect` returns one status per adapter reflecting the fake detector.
- A golden test asserting `buildContext` and `buildMasterCommand` output equals
  the current bytes (copied from `origin/main`).
- `go build ./...`, `go vet ./...`, `go test ./...` green.
- Manual: `bqa claude`, `bqa runtime detect`, `bqa runtime install-commands`
  produce the same files and stdout as before.

## Risks

- **Message/content drift** — mitigated by moving builders verbatim and a golden
  test on their output; app-layer print strings copied verbatim from today.
- **Path-writing differences** — `fs.RuntimeStore` already writes
  project-relative paths with a `..` guard; the four install paths are plain
  relative paths, so behavior matches `os.WriteFile` after `MkdirAll`.
