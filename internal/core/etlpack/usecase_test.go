package etlpack

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeStore struct {
	index     ports.SessionIndex
	indexErr  error
	knowledge map[string]string
	files     map[string]string
}

func (f *fakeStore) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	if f.indexErr != nil {
		return ports.SessionIndex{}, f.indexErr
	}
	return f.index, nil
}

func (f *fakeStore) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	return "", nil
}

func (f *fakeStore) ReadKnowledgeArtifact(ctx context.Context, filename string) (string, error) {
	content, ok := f.knowledge[filename]
	if !ok {
		return "", os.ErrNotExist
	}
	return content, nil
}

func (f *fakeStore) WriteBQAArtifact(ctx context.Context, relativePath string, content string) error {
	if f.files == nil {
		f.files = map[string]string{}
	}
	f.files[relativePath] = content
	return nil
}

func TestUseCaseGeneratesETLAgentPackWithoutCopyingPrivateInputs(t *testing.T) {
	store := &fakeStore{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{Source: "synthetic", NormalizedPath: ".bqa/input/sessions/normalized/etl-1.md"},
			{Source: "synthetic", NormalizedPath: ".bqa/input/sessions/normalized/etl-2.md"},
		}},
		knowledge: map[string]string{
			"etl_patterns.yaml":          "etl_patterns:\n  - name: \"reconcile\"\n    evidence: \"PRIVATE CUSTOMER TABLE alpha.orders\"",
			"data_quality_patterns.yaml": "data_quality_patterns:\n  - name: \"null_check\"\n    evidence: \"PRIVATE CUSTOMER TABLE alpha.users\"",
			"project_profile.yaml":       "project_profile:\n  sessions_analyzed: 2\n  signals:\n    etl: 2\n    data_quality: 1\n",
		},
		files: map[string]string{},
	}

	result, err := UseCase{Sessions: store, Knowledge: store, Writer: store}.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.ArtifactsCreated != 12 {
		t.Fatalf("expected 12 generated files, got %d", result.ArtifactsCreated)
	}
	if result.SessionsProcessed != 2 {
		t.Fatalf("expected 2 sessions processed, got %d", result.SessionsProcessed)
	}
	if result.KnowledgeArtifactsFound != 3 {
		t.Fatalf("expected 3 knowledge artifacts found, got %d", result.KnowledgeArtifactsFound)
	}

	expectedPaths := []string{
		"output/etl-agent-pack/statistics/summary.md",
		"output/etl-agent-pack/agents/codex-etl-qa-agent.md",
		"output/etl-agent-pack/agents/claude-code-etl-qa-agent.md",
		"output/etl-agent-pack/prompts/codex-etl-qa-agent-prompt.md",
		"output/etl-agent-pack/prompts/claude-code-etl-qa-agent-prompt.md",
		"output/etl-agent-pack/workflows/etl-regression-workflow.md",
		"output/etl-agent-pack/workflows/data-reconciliation-workflow.md",
		"output/etl-agent-pack/workflows/data-quality-validation-workflow.md",
		"output/etl-agent-pack/specs/etl-test-spec-template.md",
		"output/etl-agent-pack/specs/source-to-target-mapping-review-checklist.md",
		"output/etl-agent-pack/examples/synthetic-etl-case.md",
		"output/etl-agent-pack/README_NEXT_STEPS.md",
	}
	for _, path := range expectedPaths {
		if _, ok := store.files[path]; !ok {
			t.Fatalf("missing generated file %s", path)
		}
	}

	summary := store.files["output/etl-agent-pack/statistics/summary.md"]
	if !strings.Contains(summary, "Sessions analyzed: 2") {
		t.Fatalf("summary does not include session count: %s", summary)
	}
	if !strings.Contains(summary, "Knowledge artifacts available: 3") {
		t.Fatalf("summary does not include knowledge artifact count: %s", summary)
	}
	if !strings.Contains(store.files["output/etl-agent-pack/prompts/codex-etl-qa-agent-prompt.md"], "Codex ETL QA Agent Prompt") {
		t.Fatalf("missing Codex prompt content")
	}
	if !strings.Contains(store.files["output/etl-agent-pack/prompts/claude-code-etl-qa-agent-prompt.md"], "Claude Code ETL QA Agent Prompt") {
		t.Fatalf("missing Claude Code prompt content")
	}
	if !strings.Contains(store.files["output/etl-agent-pack/examples/synthetic-etl-case.md"], "Synthetic ETL QA Example") {
		t.Fatalf("missing synthetic example content")
	}
	for path, content := range store.files {
		if strings.Contains(content, "PRIVATE CUSTOMER TABLE") {
			t.Fatalf("generated file %s copied private input content", path)
		}
	}
}

func TestUseCaseFallsBackToSyntheticStatisticsWhenInputsAreMissing(t *testing.T) {
	store := &fakeStore{
		indexErr:  errors.New("missing session index"),
		knowledge: map[string]string{},
		files:     map[string]string{},
	}

	result, err := UseCase{Sessions: store, Knowledge: store, Writer: store}.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if !result.UsedSyntheticDemo {
		t.Fatalf("expected synthetic demo fallback")
	}
	summary := store.files["output/etl-agent-pack/statistics/summary.md"]
	if !strings.Contains(summary, "Input basis: synthetic demo data") {
		t.Fatalf("summary should document synthetic fallback, got: %s", summary)
	}
	if !strings.Contains(summary, "Sessions analyzed: 0") {
		t.Fatalf("summary should report zero sessions, got: %s", summary)
	}
}
