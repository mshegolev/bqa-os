package etlpack

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeInputReader struct {
	index     ports.SessionIndex
	sessions  map[string]string
	knowledge map[string]string
}

func (f fakeInputReader) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	if len(f.index.Entries) == 0 {
		return ports.SessionIndex{}, errors.New("no session index")
	}
	return f.index, nil
}

func (f fakeInputReader) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	content, ok := f.sessions[path]
	if !ok {
		return "", errors.New("missing normalized session")
	}
	return content, nil
}

func (f fakeInputReader) ReadKnowledgeArtifact(ctx context.Context, filename string) (string, error) {
	content, ok := f.knowledge[filename]
	if !ok {
		return "", errors.New("missing knowledge artifact")
	}
	return content, nil
}

type fakePackWriter struct {
	files map[string]string
}

func (f fakePackWriter) WriteETLAgentPackArtifact(ctx context.Context, relativePath string, content string) error {
	f.files[relativePath] = content
	return nil
}

func TestUseCaseGeneratesSyntheticPackWhenInputsAreMissing(t *testing.T) {
	writer := fakePackWriter{files: map[string]string{}}
	uc := UseCase{Reader: fakeInputReader{}, Writer: writer, OutputDir: ".bqa/output/etl-agent-pack"}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.ArtifactsCreated != 12 {
		t.Fatalf("expected 12 generated artifacts, got %d", result.ArtifactsCreated)
	}
	if result.SessionsProcessed != 0 {
		t.Fatalf("expected 0 processed sessions, got %d", result.SessionsProcessed)
	}
	if !result.SyntheticExamplesUsed {
		t.Fatalf("expected synthetic examples to be used when inputs are missing")
	}

	expected := map[string]string{
		"statistics/summary.md":                              "Synthetic demo data",
		"agents/codex-etl-qa-agent.md":                       "Codex ETL QA Agent",
		"agents/claude-code-etl-qa-agent.md":                 "Claude Code ETL QA Agent",
		"workflows/etl-regression-workflow.md":               "ETL Regression Workflow",
		"workflows/data-reconciliation-workflow.md":          "Data Reconciliation Workflow",
		"workflows/data-quality-validation-workflow.md":      "Data Quality Validation Workflow",
		"specs/etl-test-spec-template.md":                    "ETL Test Spec Template",
		"specs/source-to-target-mapping-review-checklist.md": "Source-to-Target Mapping Review Checklist",
		"prompts/codex-etl-qa-agent-prompt.md":               "copy-paste into Codex",
		"prompts/claude-code-etl-qa-agent-prompt.md":         "copy-paste into Claude Code",
		"examples/synthetic-etl-reconciliation-example.md":   "synthetic",
		"README_NEXT_STEPS.md":                               "Next steps",
	}

	for path, phrase := range expected {
		content, ok := writer.files[path]
		if !ok {
			t.Fatalf("missing generated artifact %s", path)
		}
		if !strings.Contains(content, phrase) {
			t.Fatalf("artifact %s does not contain %q", path, phrase)
		}
	}

	spec := writer.files["specs/etl-test-spec-template.md"]
	if strings.Contains(spec, "synthetic query placeholder") {
		t.Fatalf("spec template must no longer contain the synthetic query placeholder")
	}
	for _, marker := range []string{"COUNT(*)", ":date_start", ":date_end", "demo_src.source_table", "demo_dw.target_table", "natural_key"} {
		if !strings.Contains(spec, marker) {
			t.Fatalf("spec template should contain real SQL marker %q", marker)
		}
	}
}

func TestUseCaseUsesInputsForAggregateStatisticsWithoutCopyingSessionContent(t *testing.T) {
	reader := fakeInputReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "synthetic-etl-session.md", NormalizedPath: "normalized/synthetic-etl-session.md"},
		}},
		sessions: map[string]string{
			"normalized/synthetic-etl-session.md": "Synthetic ETL reconciliation found row count mismatch and null check failure. synthetic-sensitive-marker-123",
		},
		knowledge: map[string]string{
			"etl_patterns.yaml":          "etl_patterns:\n  - name: synthetic_reconciliation\n",
			"data_quality_patterns.yaml": "data_quality_patterns:\n  - name: synthetic_null_check\n",
			"project_profile.yaml":       "project_profile:\n  sessions_analyzed: 1\n",
		},
	}
	writer := fakePackWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/output/etl-agent-pack"}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.SessionsProcessed != 1 {
		t.Fatalf("expected 1 processed session, got %d", result.SessionsProcessed)
	}
	if result.KnowledgeArtifactsFound != 3 {
		t.Fatalf("expected 3 knowledge artifacts found, got %d", result.KnowledgeArtifactsFound)
	}
	if result.SyntheticExamplesUsed {
		t.Fatalf("did not expect synthetic statistics when ETL inputs exist")
	}

	summary := writer.files["statistics/summary.md"]
	if !strings.Contains(summary, "Sessions processed: 1") {
		t.Fatalf("summary should include processed session count, got %s", summary)
	}
	if !strings.Contains(summary, "Knowledge artifacts found: 3") {
		t.Fatalf("summary should include knowledge artifact count, got %s", summary)
	}
	if strings.Contains(summary, "synthetic-sensitive-marker-123") {
		t.Fatalf("summary must not copy raw normalized session content")
	}
}
