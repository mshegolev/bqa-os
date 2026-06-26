package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Adapter struct {
	Name       string
	BinaryName string
	Command    string
}

var adapters = []Adapter{
	{Name: "codex", BinaryName: "codex", Command: "codex"},
	{Name: "claude", BinaryName: "claude", Command: "claude"},
	{Name: "opencode", BinaryName: "opencode", Command: "opencode"},
}

func Detect() error {
	for _, adapter := range adapters {
		path, err := exec.LookPath(adapter.BinaryName)
		if err != nil {
			fmt.Printf("%-8s missing\n", adapter.Name)
			continue
		}
		fmt.Printf("%-8s %s\n", adapter.Name, path)
	}
	return nil
}

func Prepare(name string) error {
	adapter, ok := findAdapter(name)
	if !ok {
		return fmt.Errorf("unsupported runtime: %s", name)
	}

	if err := os.MkdirAll(filepath.Clean(".bqa/prompts"), 0o755); err != nil {
		return err
	}

	contextPath := filepath.Clean(".bqa/prompts/bqa-master-context.md")
	content := buildContext(adapter)
	if err := os.WriteFile(contextPath, []byte(content), 0o644); err != nil {
		return err
	}

	fmt.Printf("BQA context generated: %s\n", contextPath)
	if path, err := exec.LookPath(adapter.BinaryName); err == nil {
		fmt.Printf("Detected %s CLI: %s\n", adapter.Name, path)
		fmt.Printf("Next: start %s and paste or reference %s as the initial project instruction.\n", adapter.Command, contextPath)
		return nil
	}

	fmt.Printf("%s CLI was not found in PATH. Install it first, then run this command again.\n", adapter.Name)
	return nil
}

func findAdapter(name string) (Adapter, bool) {
	for _, adapter := range adapters {
		if adapter.Name == name {
			return adapter, true
		}
	}
	return Adapter{}, false
}

func buildContext(adapter Adapter) string {
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
`, adapter.Name)
}
