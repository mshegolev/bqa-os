package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedKnowledge writes a couple of synthetic knowledge artifacts under
// .bqa/knowledge in the current working directory, mirroring the flat YAML
// format produced by knowledge.UseCase.
func seedKnowledge(t *testing.T) {
	t.Helper()
	dir := filepath.Join(".bqa", "knowledge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	files := map[string]string{
		"etl_patterns.yaml": `etl_patterns:
  - name: "etl_validation"
    domain: "etl"
    evidence: "spark dag run reprocessed a partition after retries and removed duplicates"
    source: "normalized/etl/session-1.txt"
`,
		"data_quality_patterns.yaml": `data_quality_patterns:
  - name: "data_quality_validation"
    domain: "data_quality"
    evidence: "row count reconciliation flagged schema drift and a null check on required field"
    source: "normalized/etl/session-2.txt"
`,
		"common_bugs.yaml": `common_bugs:
  - name: "common_failure_signal"
    domain: "bugs"
    evidence: "job failed with duplicate rows and a schema drift exception"
    source: "normalized/etl/session-3.txt"
`,
		"successful_prompts.yaml": `successful_prompts:
  - name: "successful_prompt_candidate"
    domain: "prompts"
    evidence: "task: validate etl reconciliation for the daily partition"
    source: "normalized/etl/session-4.txt"
`,
		"graphql_patterns.yaml": "graphql_patterns:\n  []\n",
		"api_patterns.yaml":     "api_patterns:\n  []\n",
		"droid_patterns.yaml":   "droid_patterns:\n  []\n",
		"runtime_patterns.yaml": "runtime_patterns:\n  []\n",
		"project_profile.yaml": `project_profile:
  sessions_analyzed: 4
  signals:
    etl: 2
    graphql: 0
    api: 0
    data_quality: 1
    droid: 0
    runtime: 0
  maturity: initial
`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile %s returned error: %v", name, err)
		}
	}
}

func runCodexContext(t *testing.T) string {
	t.Helper()
	cmd := runtimeContextCmd("codex")
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	// Stdout contract: the codex path must still emit the generation line.
	if !strings.Contains(buf.String(), "BQA context generated: .bqa/prompts/bqa-master-context.md") {
		t.Fatalf("expected stdout to confirm context generation, got %q", buf.String())
	}

	data, err := os.ReadFile(filepath.Join(".bqa", "prompts", "bqa-master-context.md"))
	if err != nil {
		t.Fatalf("ReadFile master context returned error: %v", err)
	}
	return string(data)
}

// TestCodexContextIsIdempotent guards against double-appending the knowledge
// section when `bqa codex` runs more than once.
func TestCodexContextIsIdempotent(t *testing.T) {
	chdirTemp(t)
	seedKnowledge(t)

	_ = runCodexContext(t)
	content := runCodexContext(t)

	if n := strings.Count(content, "## Project QA Knowledge"); n != 1 {
		t.Fatalf("expected exactly one knowledge section after two runs, got %d", n)
	}
}

func TestCodexContextIncludesKnowledge(t *testing.T) {
	chdirTemp(t)
	seedKnowledge(t)

	content := runCodexContext(t)

	wantSections := []string{
		"## Project QA Knowledge",
		"### Project profile summary",
		"### ETL patterns",
		"### Data quality patterns",
		"### Common bugs",
		"### Successful prompts",
		"### How Codex should use this knowledge",
		"### Privacy note",
		"do not expose raw or private logs",
	}
	lower := strings.ToLower(content)
	for _, want := range wantSections {
		if !strings.Contains(lower, strings.ToLower(want)) {
			t.Errorf("expected master context to contain %q\n--- content ---\n%s", want, content)
		}
	}

	// Acceptance criterion: synthetic ETL fixture surfaces these QA signals.
	for _, signal := range []string{"partition", "retries", "duplicates", "schema drift", "data quality"} {
		if !strings.Contains(lower, signal) {
			t.Errorf("expected master context to mention %q from the ETL fixture", signal)
		}
	}

	// The base BQA master context must still be present.
	if !strings.Contains(content, "BQA Master Agent Context") {
		t.Errorf("expected base master context header to remain")
	}

	// No raw source paths should leak into the context.
	if strings.Contains(content, "normalized/etl/session-") {
		t.Errorf("raw session source paths leaked into the master context:\n%s", content)
	}
}

func TestCodexContextWithoutKnowledge(t *testing.T) {
	chdirTemp(t)

	content := runCodexContext(t)

	if !strings.Contains(content, "BQA Master Agent Context") {
		t.Errorf("expected base master context to be generated without knowledge")
	}
	if !strings.Contains(content, "No `.bqa/knowledge/*.yaml` artifacts were found") {
		t.Errorf("expected graceful no-knowledge hint, got:\n%s", content)
	}
	if !strings.Contains(content, "bqa build") {
		t.Errorf("expected suggestion to run bqa build")
	}
	if !strings.Contains(strings.ToLower(content), "do not expose raw or private logs") {
		t.Errorf("expected privacy note even without knowledge")
	}
}
