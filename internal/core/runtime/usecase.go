// Package runtime prepares a target repository for an AI coding runtime
// (Claude Code, Codex, OpenCode) by writing the BQA master context and the
// project-local command helpers, and by detecting which runtime binaries are
// installed on the host.
package runtime

import (
	"context"
	"fmt"

	"github.com/mshegolev/bqa-os/internal/ports"
)

const masterContextPath = ".bqa/prompts/bqa-master-context.md"

// adapter describes a supported AI coding runtime.
type adapter struct {
	name    string
	binary  string
	command string
}

var adapters = []adapter{
	{name: "codex", binary: "codex", command: "codex"},
	{name: "claude", binary: "claude", command: "claude"},
	{name: "opencode", binary: "opencode", command: "opencode"},
}

// UseCase prepares a repository for a runtime and detects installed binaries.
type UseCase struct {
	Writer   ports.RuntimeArtifactWriter
	Detector ports.RuntimeDetector
}

// PrepareResult reports what Prepare produced for a single runtime.
type PrepareResult struct {
	Runtime     string
	ContextPath string
	Command     string
	Detected    bool
	BinaryPath  string
}

// Prepare writes the BQA master context for the given runtime and reports
// whether the runtime binary is installed.
func (u UseCase) Prepare(ctx context.Context, runtime string) (PrepareResult, error) {
	ad, ok := findAdapter(runtime)
	if !ok {
		return PrepareResult{}, fmt.Errorf("unsupported runtime: %s", runtime)
	}

	if err := u.Writer.WriteRuntimeArtifact(ctx, masterContextPath, masterContext(ad.name)); err != nil {
		return PrepareResult{}, err
	}

	res := PrepareResult{
		Runtime:     ad.name,
		ContextPath: masterContextPath,
		Command:     ad.command,
	}
	if path, found := u.Detector.Detect(ad.binary); found {
		res.Detected = true
		res.BinaryPath = path
	}
	return res, nil
}

func findAdapter(name string) (adapter, bool) {
	for _, ad := range adapters {
		if ad.name == name {
			return ad, true
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
