package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const masterContextPath = ".bqa/prompts/bqa-master-context.md"

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

	contextPath, err := writeMasterContext(adapter)
	if err != nil {
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

func InstallCommands() error {
	contextPath, err := writeMasterContext(Adapter{Name: "project-local command"})
	if err != nil {
		return err
	}

	commandContent := buildMasterCommand()
	for _, path := range []string{
		".bqa/runtime-commands/codex/bqa-master.md",
		".bqa/runtime-commands/claude/bqa-master.md",
		".bqa/runtime-commands/opencode/bqa-master.md",
		".claude/commands/bqa-master.md",
	} {
		cleanPath := filepath.Clean(path)
		if err := os.MkdirAll(filepath.Dir(cleanPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(cleanPath, []byte(commandContent), 0o644); err != nil {
			return err
		}
		fmt.Printf("BQA runtime command written: %s\n", cleanPath)
	}

	fmt.Printf("BQA master context generated: %s\n", contextPath)
	fmt.Println("Claude Code can use /bqa-master in this project.")
	fmt.Println("Codex and OpenCode command helpers are available under .bqa/runtime-commands/.")
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

func writeMasterContext(adapter Adapter) (string, error) {
	if err := os.MkdirAll(filepath.Clean(".bqa/prompts"), 0o755); err != nil {
		return "", err
	}

	contextPath := filepath.Clean(masterContextPath)
	content := buildContext(adapter)
	if err := os.WriteFile(contextPath, []byte(content), 0o644); err != nil {
		return "", err
	}
	return contextPath, nil
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

func buildMasterCommand() string {
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
