# Design: Workspace registry foundation (issue #67, slice 1)

## Context

Issue #67 asks for a workspace-aware, one-command flow — `bqa task start <JIRA_KEY>` —
that prepares a git worktree, resolves layered BQA context, and opens an
interactive runtime console for a QA engineer. That is five independent
subsystems (workspace registry, Jira source, git-worktree manager, task context
resolution, interactive runtime launch). It is too large for one spec, so #67 is
**sliced**.

**Target vision (decided during brainstorming, informs the roadmap):** the eventual
`task start` uses a **live Jira REST** source and **execs** the detected runtime
(codex/claude/opencode). Those are later slices.

**This slice (1)** builds only the **workspace registry** — the persistent model
and CLI that every later `task`/context/worktree command reads and writes. It
touches no Jira, no git worktree, and no runtime exec, so it stays small and fully
deterministic/testable while unblocking the rest.

Builds on the repo's established patterns: hexagonal core/ports/adapters, v1 YAML
envelope (`schema_version` / `kind` / `generated_by`) as used by `core/knowledge`
and `core/memgov`, deterministic output (no wall-clock; `generated_by` from
`version.Version`), the `internal/app/brain.go` cobra command-group style, the
`fsadapter` write pattern (`pathsafe`, `0o600`), and `textutil.QuoteYAML` /
`textutil.UnquoteYAML`.

**Decisions from brainstorming:**

1. **Scope:** slice 1 = `bqa workspace init/add/list` over `.bqa/workspace.yaml`.
   The `task`/Jira/worktree/context/exec pieces are later slices.
2. **`workspace add` validation:** require the `<path>` to exist and be a
   directory (else a clear error, nothing written); if it exists but is **not** a
   git repo, record it anyway with a printed warning.
3. **Workspace model:** one workspace per `.bqa` (single top-level object with an
   optional `name`), not a list of named workspaces.

**Constraint:** stdlib + cobra only; deterministic output; strict hexagonal
(`core/*` depends only on ports + stdlib, plus the shared `textutil`/`version`
helpers). `workspace add` never copies or mutates a target repository — it only
records a path.

## Command surface

A new `bqa workspace` command group:

```
bqa workspace init  [--name <name>] [--base-dir .bqa]
bqa workspace add <project-id> <path> --repo <repo-name> [--etl <ETL>] [--branch-role base] [--base-dir .bqa]
bqa workspace list  [--base-dir .bqa]
```

- `--base-dir` defaults to `.bqa`, matching `bqa run`/`bqa doctor`.
- `init --name` sets the workspace name; when omitted it defaults to the base name
  of the current working directory.
- `add` positional args: `<project-id>` and `<path>`. `--repo` is required;
  `--etl` defaults to `""`; `--branch-role` defaults to `base`.

## Data model

`.bqa/workspace.yaml`, v1 envelope + a single workspace object:

```yaml
schema_version: 1
kind: workspace
generated_by: bqa dev
name: bigdata-testing
projects:
  - id: main
    path: /Users/me/develop/bigdata_testing
    repo: bigdata_testing
    etl: ""
    branch_role: base
tasks: []
```

- `name` — workspace name (from `--name` or the cwd base name).
- `projects` — registered local repo/worktree roots. Each: `id`, `path`, `repo`,
  `etl`, `branch_role`.
  - `path` is stored **verbatim** (it points at a real repo/worktree that may live
    anywhere, so it is not forced relative or under `.bqa`). Free-text fields are
    YAML-quoted on write.
- `tasks: []` — reserved and always rendered (empty in slice 1) so the schema is
  stable for slice 2's `task start`, which appends task entries.

## Architecture (hexagonal)

```
internal/app/workspace.go            (CLI wiring: bqa workspace init/add/list)
        ↓
internal/core/workspace/             (use case — ports + stdlib + textutil/version)
        ↓
internal/ports/workspace.go          (WorkspaceStore, PathInspector)
        ↑
internal/adapters/fs/workspace_store.go  (fs store + fs PathInspector)
```

Pure logic (the workspace model, deterministic YAML render/parse, and the
init/add/list use cases) lives in `core/workspace`. The only side effects are
behind two ports: reading/writing `workspace.yaml`, and inspecting a path on disk.

### Ports (`internal/ports/workspace.go`)

```go
type WorkspaceStore interface {
    // Exists reports whether a workspace file is present under baseDir.
    Exists(ctx context.Context, baseDir string) (bool, error)
    // Load reads and returns the raw workspace file content.
    Load(ctx context.Context, baseDir string) (content string, err error)
    // Save writes the workspace file content under baseDir (creating baseDir).
    Save(ctx context.Context, baseDir string, content string) error
}

type PathInspector interface {
    // IsDir reports whether path exists and is a directory.
    IsDir(path string) (bool, error)
    // IsGitRepo reports whether path is (or is inside) a git working tree.
    IsGitRepo(path string) (bool, error)
}
```

`WorkspaceStore` is file-level (content in/out) so all YAML parse/render stays in
`core/workspace` and the fs adapter stays pure I/O — consistent with the
`GovernanceStore` decision in #40 and with the repo rule that adapters depend on
ports + stdlib only. `PathInspector` keeps the `add` filesystem checks out of the
core so both branches (missing dir → error; non-git dir → warn) are unit-testable
with a fake.

### Data model types (`internal/core/workspace/models.go`)

```go
const SchemaVersion = 1
const KindWorkspace = "workspace"
const DefaultBranchRole = "base"
const workspaceFileName = "workspace.yaml"

type Project struct {
    ID         string
    Path       string
    Repo       string
    ETL        string
    BranchRole string
}

// Task is reserved for slice 2 (task start). Rendered as an empty list in slice 1.
type Task struct {
    ID     string
    Jira   string
    Repo   string
    ETL    string
    Path   string
    Branch string
}

type Workspace struct {
    Name     string
    Projects []Project
    Tasks    []Task
}
```

### Use cases (`internal/core/workspace/`)

- **Init** (`usecase.go`): if `WorkspaceStore.Exists` → error
  (`workspace already initialized …`), nothing written. Else render an empty
  workspace (`Name` from the option, empty `projects`/`tasks`) and `Save`. Returns
  the name and the file path for display.
- **Add** (`usecase.go`): if `WorkspaceStore.Exists` is false → error
  (`no workspace; run 'bqa workspace init'`), nothing changed. Else `Load` + parse.
  Validate via `PathInspector.IsDir(path)`: false → error, nothing changed.
  `IsGitRepo(path)` false → set a `Warning` on the result but continue. Reject a
  duplicate `project-id` → error, nothing changed. Else append the `Project`,
  render, `Save`. Returns the added project + optional warning.
- **List** (`usecase.go`): if `WorkspaceStore.Exists` is false → same
  not-initialized error. Else `Load` + parse and return the `Workspace` for display
  (name, projects in insertion order, task count).

Render/parse (`render.go` / `parse.go`) mirror the `memgov` approach: a v1 header
(`schema_version` / `kind: workspace` / `generated_by`), a `name:` scalar, a
`projects:` list (`- id:` starting each entry; `path`/`repo`/`etl` quoted;
`branch_role` bareword), and a `tasks:` list (empty `[]` in this slice; the parser
tolerates entries so slice 2 needs no parser change). `generatedBy()` returns
`"bqa " + version.Version` (deterministic: `bqa dev` in tests).

### CLI (`internal/app/workspace.go`)

`workspaceCmd()` builds the group with three subcommands, each owning its own flag
variables (per the `brain.go` convention — no shared-flag bleed), wiring
`fsadapter.WorkspaceStore{}` and `fsadapter.PathInspector{}` into
`workspace.UseCase`. Registered in `internal/app/root.go` via
`rootCmd.AddCommand(workspaceCmd())`.

- `init` prints the created path and name.
- `add` prints the added project; if a non-git warning was returned, prints it to
  stderr.
- `list` prints the workspace name, one line per project
  (`id  repo  etl  branch_role  path`), and the task count.

## Testing

- **core/workspace** (fake `WorkspaceStore` + fake `PathInspector`):
  - Init writes a v1 workspace with the given name and empty `projects`/`tasks`; a
    second Init errors and does not overwrite.
  - Add appends a project and persists it; a duplicate id errors and changes
    nothing; a missing/ non-dir path errors and changes nothing; a non-git dir is
    recorded and returns a warning.
  - List returns the parsed workspace with projects in insertion order.
  - Render↔parse round-trip (including a project whose fields contain quotes).
- **adapters/fs**: `WorkspaceStore` load/save round-trip and `Exists` under a temp
  dir; `PathInspector.IsDir`/`IsGitRepo` against temp dirs (a plain dir, a
  `git init`-ed dir, and a missing path).
- **app**: drive the real command tree — `workspace init` then `add` then `list`
  over a temp `--base-dir` — asserting the workspace file contents and the list
  output; and that `add` before `init` errors.
- `go test ./...` passes; output is deterministic.

## Out of scope (later #67 slices)

- `bqa task start/status/finish` and the `tasks:` entries they create.
- Jira issue source (live Jira REST) and ETL/repo detection from Jira fields.
- Git worktree creation and `feature/<KEY>` branch management.
- Task context resolution into `.bqa/runtime/<task-id>/context.md` (layering global
  QA + repo knowledge + ETL + Jira + worktree diff).
- Interactive runtime console launch (`exec` codex/claude/opencode).
- Upward discovery of `.bqa` (commands keep the explicit `--base-dir` flag for now).
</content>
