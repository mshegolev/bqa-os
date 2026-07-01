package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestAcceptanceInstalledQAAgentWorkflow is an end-to-end business-acceptance
// test for the *installed* QA agent workflow described in issue #39.
//
// It proves the user-facing promise of `bqa brain install`: after a brain
// package has been generated, installing it into a target client project makes
// the expected agent/context entrypoints present and usable under <target>/.bqa/.
//
// The test runs entirely from t.TempDir() with synthetic, non-private content
// and drives the real cobra `brain install` command (the same code path a user
// gets on the CLI). It asserts that:
//   - the install command succeeds and reports the install,
//   - the target gains a .bqa/ directory holding registry + agent + prompt
//     entrypoints,
//   - each installed entrypoint is non-empty (a usable file, not a stub),
//   - unrelated files already present in the target are left untouched.
func TestAcceptanceInstalledQAAgentWorkflow(t *testing.T) {
	tmp := t.TempDir()

	// Build a synthetic brain package: the dirs brain.Install expects
	// (registry/ + at least one artifact dir). We populate registry, an agent,
	// a prompt and a knowledge file so the installed package is realistic.
	pkgDir := filepath.Join(tmp, "brain-package")
	pkgFiles := map[string]string{
		"registry/index.yaml": "" +
			"registry:\n" +
			"  version: 1\n" +
			"  agents:\n" +
			"    - qa-etl-acceptance\n" +
			"  prompts:\n" +
			"    - qa-planning\n",
		"agents/qa-etl-acceptance.md": "" +
			"# QA ETL Acceptance Agent\n\n" +
			"You are a QA engineer. Given an ETL change, produce an acceptance test\n" +
			"plan covering row-count reconciliation, null/duplicate data-quality\n" +
			"checks, and failure-path verification using only synthetic data.\n",
		"prompts/qa-planning.md": "" +
			"Task: plan acceptance testing for a nightly ETL pipeline run.\n" +
			"Acceptance criteria: reconciliation passes and no data-quality regressions.\n",
		"knowledge/etl_patterns.yaml": "" +
			"schema_version: 1\n" +
			"kind: etl_patterns\n" +
			"generated_by: bqa dev\n" +
			"patterns:\n" +
			"  - id: \"etl-00000000\"\n" +
			"    name: \"row_count_reconciliation\"\n" +
			"    domain: \"etl\"\n" +
			"    evidence: \"compare source and target row counts for the window\"\n" +
			"    source: \"normalized/etl/s1.md\"\n" +
			"    reusable_check: \"compare source vs target row counts for the window\"\n" +
			"    confidence: low\n",
	}
	writeAcceptancePackage(t, pkgDir, pkgFiles)

	// Target client project with a pre-existing, unrelated file that must be
	// preserved by the install (it only writes under .bqa/).
	targetDir := filepath.Join(tmp, "client-project")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("MkdirAll target dir returned error: %v", err)
	}
	preExisting := filepath.Join(targetDir, "README.md")
	const preExistingBody = "# Client Project\nDo not touch.\n"
	if err := os.WriteFile(preExisting, []byte(preExistingBody), 0o600); err != nil {
		t.Fatalf("WriteFile pre-existing target file returned error: %v", err)
	}

	// Drive the real cobra `brain install` command.
	out := executeAcceptanceCmd(t, brainCmd(), []string{
		"install",
		"--from", pkgDir,
		"--target", targetDir,
	})
	if !strings.Contains(out, "Installed into:") {
		t.Fatalf("install output missing confirmation, got:\n%s", out)
	}

	// The installed .bqa entrypoints must exist and be non-empty so the agent
	// context is actually usable after install.
	bqaDir := filepath.Join(targetDir, ".bqa")
	entrypoints := []string{
		filepath.Join("registry", "index.yaml"),
		filepath.Join("agents", "qa-etl-acceptance.md"),
		filepath.Join("prompts", "qa-planning.md"),
		filepath.Join("knowledge", "etl_patterns.yaml"),
	}
	for _, rel := range entrypoints {
		path := filepath.Join(bqaDir, rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected installed entrypoint %q missing: %v", rel, err)
		}
		if info.Size() == 0 {
			t.Fatalf("installed entrypoint %q is empty", rel)
		}
	}

	// The installed agent must still carry usable guidance (content survived the
	// copy, not just an empty placeholder of the right name).
	agentBody := readAcceptanceFile(t, filepath.Join(bqaDir, "agents", "qa-etl-acceptance.md"))
	if !strings.Contains(agentBody, "QA ETL Acceptance Agent") {
		t.Fatalf("installed agent lost its content, got:\n%s", agentBody)
	}

	// Unrelated files in the target must be left untouched.
	if got := readAcceptanceFile(t, preExisting); got != preExistingBody {
		t.Fatalf("install modified an unrelated target file, got:\n%s", got)
	}
}

// executeAcceptanceCmd runs a cobra command with the given args, capturing
// stdout/stderr into a buffer and failing the test on error. It is named
// uniquely to avoid colliding with executeCmd in e2e_test.go.
func executeAcceptanceCmd(t *testing.T, cmd *cobra.Command, args []string) string {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("%q execute returned error: %v\noutput: %s", cmd.Name(), err, out.String())
	}
	return out.String()
}

// writeAcceptancePackage materializes a synthetic brain package from a map of
// relative paths to file bodies, creating parent directories as needed.
func writeAcceptancePackage(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll %q returned error: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatalf("WriteFile %q returned error: %v", path, err)
		}
	}
}

// readAcceptanceFile reads a file or fails the test. Named uniquely to avoid
// colliding with readFile in e2e_test.go.
func readAcceptanceFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %q returned error: %v", path, err)
	}
	return string(data)
}
