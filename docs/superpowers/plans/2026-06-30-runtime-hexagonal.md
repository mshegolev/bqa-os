# Runtime Hexagonalization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move `internal/runtime` into the hexagonal structure (`internal/core/runtime` + ports + adapters + app), remove the duplicated master-context write, and collapse the three near-identical runtime commands — preserving all outputs and stdout byte-for-byte.

**Architecture:** A `core/runtime.UseCase` holds pure content builders and orchestration, depending on `ports.RuntimeArtifactWriter` (reused from emit) and a new `ports.RuntimeDetector`. A `runtimebin` adapter implements detection via `exec.LookPath`; `fs.RuntimeStore{TargetDir:"."}` does the writing. The app layer constructs these and prints the same messages as today.

**Tech Stack:** Go 1.22, stdlib only, cobra.

Spec: `docs/superpowers/specs/2026-06-30-runtime-hexagonal-design.md`

---

## File Structure

- `internal/ports/runtime_setup.go` — `RuntimeDetector` interface.
- `internal/core/runtime/usecase.go` — `UseCase`, result types, `Prepare`/`InstallCommands`/`Detect`, pure builders `masterContext`/`masterCommand`.
- `internal/core/runtime/usecase_test.go` — unit tests with fakes + golden builders.
- `internal/adapters/runtimebin/detector.go` — `Detector` via `exec.LookPath`.
- `internal/app/runtime_context.go` — `runtimeContextCmd(name)` factory.
- `internal/app/runtime.go` — wire use case for `detect`/`install-commands` (modify).
- `internal/app/root.go` — register three context commands from the factory (modify).
- Deleted: `internal/runtime/`, `internal/app/claude.go`, `codex.go`, `opencode.go`.

---

## Task 1: Detector port + content builders + `Prepare`

**Files:**
- Create: `internal/ports/runtime_setup.go`
- Create: `internal/core/runtime/usecase.go`
- Test: `internal/core/runtime/usecase_test.go`

The current builders to copy VERBATIM live in `origin/main:internal/runtime/runtime.go` (`buildContext` and `buildMasterCommand`). `buildContext` interpolates a runtime name into `running through the %s runtime`.

- [ ] **Step 1: Write the failing test**

```go
package runtime

import (
	"context"
	"strings"
	"testing"
)

type fakeWriter struct{ files map[string]string }

func (f *fakeWriter) WriteRuntimeArtifact(ctx context.Context, path, content string) error {
	f.files[path] = content
	return nil
}

type fakeDetector struct{ found map[string]string }

func (f fakeDetector) Detect(binary string) (string, bool) {
	p, ok := f.found[binary]
	return p, ok
}

func TestPrepareWritesMasterContext(t *testing.T) {
	w := &fakeWriter{files: map[string]string{}}
	uc := UseCase{Writer: w, Detector: fakeDetector{found: map[string]string{"claude": "/usr/bin/claude"}}}

	res, err := uc.Prepare(context.Background(), "claude")
	if err != nil {
		t.Fatalf("Prepare error: %v", err)
	}
	if res.ContextPath != ".bqa/prompts/bqa-master-context.md" {
		t.Fatalf("context path: %s", res.ContextPath)
	}
	got := w.files[".bqa/prompts/bqa-master-context.md"]
	if !strings.Contains(got, "through the claude runtime") {
		t.Fatalf("master context missing runtime name:\n%s", got)
	}
	if !res.Detected || res.BinaryPath != "/usr/bin/claude" {
		t.Fatalf("detection not reported: %+v", res)
	}
}

func TestPrepareRejectsUnknownRuntime(t *testing.T) {
	uc := UseCase{Writer: &fakeWriter{files: map[string]string{}}, Detector: fakeDetector{}}
	if _, err := uc.Prepare(context.Background(), "vim"); err == nil {
		t.Fatal("expected error for unsupported runtime")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /tmp/bqa-runtime-wt && go test ./internal/core/runtime/ -run TestPrepare`
Expected: FAIL — package/identifiers undefined.

- [ ] **Step 3: Write minimal implementation**

Create `internal/ports/runtime_setup.go`:

```go
package ports

// RuntimeDetector reports whether a runtime CLI binary is on PATH.
type RuntimeDetector interface {
	Detect(binary string) (path string, found bool)
}
```

Create `internal/core/runtime/usecase.go`:

```go
// Package runtime prepares BQA Master Agent context and command helpers for the
// supported AI coding runtimes.
package runtime

import (
	"context"
	"fmt"

	"github.com/mshegolev/bqa-os/internal/ports"
)

const masterContextPath = ".bqa/prompts/bqa-master-context.md"

type adapter struct{ name, binary, command string }

var adapters = []adapter{
	{name: "codex", binary: "codex", command: "codex"},
	{name: "claude", binary: "claude", command: "claude"},
	{name: "opencode", binary: "opencode", command: "opencode"},
}

// UseCase prepares runtime context/commands via injected ports.
type UseCase struct {
	Writer   ports.RuntimeArtifactWriter
	Detector ports.RuntimeDetector
}

// PrepareResult reports what Prepare produced.
type PrepareResult struct {
	Runtime     string
	ContextPath string
	Command     string
	Detected    bool
	BinaryPath  string
}

func (u UseCase) Prepare(ctx context.Context, runtime string) (PrepareResult, error) {
	ad, ok := findAdapter(runtime)
	if !ok {
		return PrepareResult{}, fmt.Errorf("unsupported runtime: %s", runtime)
	}
	if err := u.writeMasterContext(ctx, ad.name); err != nil {
		return PrepareResult{}, err
	}
	res := PrepareResult{Runtime: ad.name, ContextPath: masterContextPath, Command: ad.command}
	if path, found := u.Detector.Detect(ad.binary); found {
		res.Detected = true
		res.BinaryPath = path
	}
	return res, nil
}

func (u UseCase) writeMasterContext(ctx context.Context, runtimeName string) error {
	return u.Writer.WriteRuntimeArtifact(ctx, masterContextPath, masterContext(runtimeName))
}

func findAdapter(name string) (adapter, bool) {
	for _, a := range adapters {
		if a.name == name {
			return a, true
		}
	}
	return adapter{}, false
}

func masterContext(runtimeName string) string {
	return fmt.Sprintf(`# BQA Master Agent Context

You are BQA Master Agent running through the %s runtime.

BQA-OS stands for Better QA Operating System.

Responsibilities:

1. Understand the QA task.
2. Detect the domain: Big Data ETL, GraphQL Functional Testing, API Testing, Contract Testing, or general automation.
3. Load local BQA artifacts when available:
   - .bqa/registry/
   - .bqa/memory/
   - .bqa/agents/
   - .bqa/skills/
   - .bqa/workflows/
   - .bqa/rules/
   - .bqa/guardrails/
4. Create a plan before changing code.
5. Delegate logically to specialist agents by reading their definitions from BQA artifacts.
6. Prefer tests, evidence, reproducible commands, and clear reports.

Default behavior:

- For ETL tasks, use Big Data / ETL QA workflows.
- For GraphQL tasks, use GraphQL Functional QA workflows.
- For ambiguous tasks, inspect the repository before selecting a domain.
- After task completion, propose memory updates for BQA Brain.
`, runtimeName)
}
```

IMPORTANT: the string body inside `masterContext` must be byte-identical to
`buildContext` in `origin/main:internal/runtime/runtime.go` (only the function
name and parameter change). Copy it exactly.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /tmp/bqa-runtime-wt && go test ./internal/core/runtime/ -run TestPrepare -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ports/runtime_setup.go internal/core/runtime/
git commit -m "feat(core/runtime): Prepare use case + RuntimeDetector port"
```

---

## Task 2: `InstallCommands` + `Detect` + golden builder test

**Files:**
- Modify: `internal/core/runtime/usecase.go`
- Test: `internal/core/runtime/usecase_test.go`

`buildMasterCommand()` to copy verbatim is in `origin/main:internal/runtime/runtime.go`.

- [ ] **Step 1: Write the failing test**

```go
func TestInstallCommandsWritesAllFiles(t *testing.T) {
	w := &fakeWriter{files: map[string]string{}}
	uc := UseCase{Writer: w, Detector: fakeDetector{}}

	res, err := uc.InstallCommands(context.Background())
	if err != nil {
		t.Fatalf("InstallCommands error: %v", err)
	}
	want := []string{
		".bqa/runtime-commands/codex/bqa-master.md",
		".bqa/runtime-commands/claude/bqa-master.md",
		".bqa/runtime-commands/opencode/bqa-master.md",
		".claude/commands/bqa-master.md",
	}
	for _, p := range want {
		if !strings.Contains(w.files[p], "# /bqa-master") {
			t.Fatalf("missing or wrong command file: %s", p)
		}
	}
	if res.ContextPath != masterContextPath {
		t.Fatalf("install context path: %s", res.ContextPath)
	}
	// master context uses the project-local command name
	if !strings.Contains(w.files[masterContextPath], "through the project-local command runtime") {
		t.Fatalf("install master context name wrong:\n%s", w.files[masterContextPath])
	}
}

func TestDetectReportsEachRuntime(t *testing.T) {
	uc := UseCase{Writer: &fakeWriter{files: map[string]string{}}, Detector: fakeDetector{found: map[string]string{"codex": "/bin/codex"}}}
	statuses, err := uc.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("expected 3 statuses, got %d", len(statuses))
	}
	byName := map[string]RuntimeStatus{}
	for _, s := range statuses {
		byName[s.Name] = s
	}
	if !byName["codex"].Found || byName["codex"].BinaryPath != "/bin/codex" {
		t.Fatal("codex should be found")
	}
	if byName["claude"].Found {
		t.Fatal("claude should be missing")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /tmp/bqa-runtime-wt && go test ./internal/core/runtime/ -run 'TestInstallCommands|TestDetect'`
Expected: FAIL — `InstallCommands`/`Detect`/`RuntimeStatus`/`masterCommand` undefined.

- [ ] **Step 3: Write minimal implementation**

Append to `internal/core/runtime/usecase.go`:

```go
// InstallResult reports what InstallCommands produced.
type InstallResult struct {
	ContextPath string
	Commands    []string
}

// RuntimeStatus is one runtime's detection result.
type RuntimeStatus struct {
	Name       string
	Found      bool
	BinaryPath string
}

func (u UseCase) InstallCommands(ctx context.Context) (InstallResult, error) {
	if err := u.writeMasterContext(ctx, "project-local command"); err != nil {
		return InstallResult{}, err
	}
	paths := []string{
		".bqa/runtime-commands/codex/bqa-master.md",
		".bqa/runtime-commands/claude/bqa-master.md",
		".bqa/runtime-commands/opencode/bqa-master.md",
		".claude/commands/bqa-master.md",
	}
	cmd := masterCommand()
	for _, p := range paths {
		if err := u.Writer.WriteRuntimeArtifact(ctx, p, cmd); err != nil {
			return InstallResult{}, err
		}
	}
	return InstallResult{ContextPath: masterContextPath, Commands: paths}, nil
}

func (u UseCase) Detect(ctx context.Context) ([]RuntimeStatus, error) {
	statuses := make([]RuntimeStatus, 0, len(adapters))
	for _, a := range adapters {
		path, found := u.Detector.Detect(a.binary)
		statuses = append(statuses, RuntimeStatus{Name: a.name, Found: found, BinaryPath: path})
	}
	return statuses, nil
}

func masterCommand() string {
	return `# /bqa-master

Read .bqa/prompts/bqa-master-context.md and act as BQA Master Agent.

Load project-local BQA artifacts before planning:

- .bqa/registry/
- .bqa/memory/
- .bqa/agents/
- .bqa/skills/
- .bqa/workflows/
- .bqa/rules/
- .bqa/guardrails/

Default workflow:

1. Inspect the repository and current task context.
2. Select the applicable BQA domain workflow.
3. Create a short plan before changing files.
4. Execute with tests and reproducible evidence.
5. Propose BQA Brain memory updates for reusable findings.
`
}
```

The `masterCommand` body must be byte-identical to `buildMasterCommand` in `origin/main:internal/runtime/runtime.go`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /tmp/bqa-runtime-wt && go test ./internal/core/runtime/ -v`
Expected: PASS (all tests).

- [ ] **Step 5: Commit**

```bash
git add internal/core/runtime/
git commit -m "feat(core/runtime): InstallCommands + Detect"
```

---

## Task 3: `runtimebin` detector adapter

**Files:**
- Create: `internal/adapters/runtimebin/detector.go`
- Test: `internal/adapters/runtimebin/detector_test.go`

- [ ] **Step 1: Write the failing test**

```go
package runtimebin

import "testing"

func TestDetectKnownBinary(t *testing.T) {
	// "go" is on PATH in the test environment
	if _, ok := (Detector{}).Detect("go"); !ok {
		t.Fatal("expected to find go on PATH")
	}
	if _, ok := (Detector{}).Detect("definitely-not-a-real-binary-xyz"); ok {
		t.Fatal("did not expect to find a bogus binary")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /tmp/bqa-runtime-wt && go test ./internal/adapters/runtimebin/`
Expected: FAIL — `Detector` undefined.

- [ ] **Step 3: Write minimal implementation**

Create `internal/adapters/runtimebin/detector.go`:

```go
// Package runtimebin detects AI coding runtime CLIs on PATH.
package runtimebin

import "os/exec"

// Detector implements ports.RuntimeDetector via exec.LookPath.
type Detector struct{}

func (Detector) Detect(binary string) (string, bool) {
	path, err := exec.LookPath(binary)
	if err != nil {
		return "", false
	}
	return path, true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /tmp/bqa-runtime-wt && go test ./internal/adapters/runtimebin/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/adapters/runtimebin/
git commit -m "feat(adapters/runtimebin): RuntimeDetector via exec.LookPath"
```

---

## Task 4: App wiring + delete old package and stub commands

**Files:**
- Create: `internal/app/runtime_context.go`
- Modify: `internal/app/runtime.go`
- Modify: `internal/app/root.go`
- Delete: `internal/app/claude.go`, `internal/app/codex.go`, `internal/app/opencode.go`, `internal/runtime/` (whole package)

- [ ] **Step 1: Create the runtime command factory**

Create `internal/app/runtime_context.go`:

```go
package app

import (
	"context"
	"fmt"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	"github.com/mshegolev/bqa-os/internal/adapters/runtimebin"
	coreruntime "github.com/mshegolev/bqa-os/internal/core/runtime"
	"github.com/spf13/cobra"
)

func newRuntimeUseCase() coreruntime.UseCase {
	return coreruntime.UseCase{
		Writer:   fsadapter.RuntimeStore{TargetDir: "."},
		Detector: runtimebin.Detector{},
	}
}

func runtimeContextCmd(name string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: fmt.Sprintf("Prepare BQA Master Agent context for %s", runtimeLabel(name)),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := newRuntimeUseCase().Prepare(context.Background(), name)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "BQA context generated: %s\n", res.ContextPath)
			if res.Detected {
				fmt.Fprintf(out, "Detected %s CLI: %s\n", res.Runtime, res.BinaryPath)
				fmt.Fprintf(out, "Next: start %s and paste or reference %s as the initial project instruction.\n", res.Command, res.ContextPath)
				return nil
			}
			fmt.Fprintf(out, "%s CLI was not found in PATH. Install it first, then run this command again.\n", res.Runtime)
			return nil
		},
	}
}

func runtimeLabel(name string) string {
	switch name {
	case "claude":
		return "Claude Code"
	case "codex":
		return "Codex CLI"
	case "opencode":
		return "OpenCode"
	default:
		return name
	}
}
```

Note: the `Short` strings reproduce the originals ("Prepare BQA Master Agent context for Claude Code" / "...for Codex CLI" / "...for OpenCode").

- [ ] **Step 2: Update `runtime.go` to use the use case**

Replace the bodies of the `detect` and `install-commands` subcommands in
`internal/app/runtime.go` so they call the use case and print today's messages:

```go
// detect RunE:
RunE: func(cmd *cobra.Command, args []string) error {
	statuses, err := newRuntimeUseCase().Detect(context.Background())
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	for _, s := range statuses {
		if s.Found {
			fmt.Fprintf(out, "%-8s %s\n", s.Name, s.BinaryPath)
		} else {
			fmt.Fprintf(out, "%-8s missing\n", s.Name)
		}
	}
	return nil
},

// install-commands RunE:
RunE: func(cmd *cobra.Command, args []string) error {
	res, err := newRuntimeUseCase().InstallCommands(context.Background())
	if err != nil {
		return err
	}
	out := cmd.OutOrStdout()
	for _, p := range res.Commands {
		fmt.Fprintf(out, "BQA runtime command written: %s\n", p)
	}
	fmt.Fprintf(out, "BQA master context generated: %s\n", res.ContextPath)
	fmt.Fprintln(out, "Claude Code can use /bqa-master in this project.")
	fmt.Fprintln(out, "Codex and OpenCode command helpers are available under .bqa/runtime-commands/.")
	return nil
},
```

Update imports in `runtime.go`: remove `"github.com/mshegolev/bqa-os/internal/runtime"`; add `"context"` and `"fmt"` if not present. Keep the `runtimeCmd` parent and its `Use`/`Short`.

- [ ] **Step 3: Update `root.go` registration**

In `internal/app/root.go`, replace the three lines:

```go
	rootCmd.AddCommand(codexCmd())
	rootCmd.AddCommand(claudeCmd())
	rootCmd.AddCommand(opencodeCmd())
```

with:

```go
	rootCmd.AddCommand(runtimeContextCmd("codex"))
	rootCmd.AddCommand(runtimeContextCmd("claude"))
	rootCmd.AddCommand(runtimeContextCmd("opencode"))
```

- [ ] **Step 4: Delete the old package and stub command files**

```bash
cd /tmp/bqa-runtime-wt
git rm internal/app/claude.go internal/app/codex.go internal/app/opencode.go
git rm -r internal/runtime/
```

- [ ] **Step 5: Build, vet, format, test**

Run:
```bash
cd /tmp/bqa-runtime-wt
gofmt -w internal/app internal/core/runtime internal/adapters/runtimebin internal/ports
go vet ./...
go build ./...
go test ./...
```
Expected: gofmt clean, vet clean, build OK, all tests green. (No importer of `internal/runtime` remains — confirm with `grep -rn 'internal/runtime\"' --include='*.go' .` → empty, ignoring `internal/core/runtime`.)

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor(app): wire runtime use case; collapse runtime commands; drop internal/runtime"
```

---

## Task 5: Behavior verification

**Files:** none (verification only)

- [ ] **Step 1: Build the binary and compare outputs to current behavior**

Run:
```bash
cd /tmp/bqa-runtime-wt && go build -o /tmp/bqa-rt ./cmd/bqa
TMP=$(mktemp -d); cd "$TMP"
/tmp/bqa-rt claude
echo "---"; cat .bqa/prompts/bqa-master-context.md | head -3
/tmp/bqa-rt runtime install-commands
echo "---"; find .bqa/runtime-commands .claude/commands -type f | sort
/tmp/bqa-rt runtime detect
```
Expected:
- `bqa claude` prints `BQA context generated: .bqa/prompts/bqa-master-context.md` then a detection or "not found" line; the context contains "through the claude runtime".
- `runtime install-commands` prints four `BQA runtime command written:` lines, the master-context line, and the two helper lines; writes the four command files.
- `runtime detect` prints one line per runtime (path or `missing`).

- [ ] **Step 2: Confirm help still lists the three commands**

Run: `/tmp/bqa-rt --help`
Expected: `claude`, `codex`, `opencode` still listed with their original Short text.
