package runplan

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeRegistryReader struct {
	registry ports.BQARegistry
	err      error
}

func (f fakeRegistryReader) LoadBQARegistry(ctx context.Context) (ports.BQARegistry, error) {
	return f.registry, f.err
}

func TestUseCaseSelectsDomainArtifactsForTask(t *testing.T) {
	uc := UseCase{RegistryReader: fakeRegistryReader{registry: registryFixture()}}

	plan, err := uc.Run(context.Background(), Options{Task: "Validate ETL pipeline on stage"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if plan.Task != "Validate ETL pipeline on stage" {
		t.Fatalf("Task = %q, expected trimmed input task", plan.Task)
	}
	if len(plan.Agents) != 1 || plan.Agents[0].ID != "etl-qa-agent" {
		t.Fatalf("expected ETL agent only, got %#v", plan.Agents)
	}
	if len(plan.Workflows) != 1 || plan.Workflows[0].ID != "etl-verification-workflow" {
		t.Fatalf("expected ETL workflow only, got %#v", plan.Workflows)
	}
	if len(plan.Skills) != 1 || plan.Skills[0].ID != "etl-log-investigation" {
		t.Fatalf("expected ETL skill only, got %#v", plan.Skills)
	}
	if len(plan.Steps) == 0 {
		t.Fatalf("expected execution plan steps")
	}
	for _, marker := range placeholderMarkers() {
		if strings.Contains(plan.Report, marker) {
			t.Fatalf("report must not include placeholder marker %q, got %q", marker, plan.Report)
		}
	}
}

func TestUseCaseUsesDefaultTaskAndFallbackSelection(t *testing.T) {
	uc := UseCase{RegistryReader: fakeRegistryReader{registry: registryFixture()}}

	plan, err := uc.Run(context.Background(), Options{Task: "   "})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if plan.Task == "" {
		t.Fatalf("expected default task")
	}
	if len(plan.Agents) != 1 || plan.Agents[0].ID != "etl-qa-agent" {
		t.Fatalf("expected default task to select ETL agent, got %#v", plan.Agents)
	}
}

func TestUseCaseFallsBackToAllArtifactsWhenNoDomainMatches(t *testing.T) {
	uc := UseCase{RegistryReader: fakeRegistryReader{registry: registryFixture()}}

	plan, err := uc.Run(context.Background(), Options{Task: "Review the current project"})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if len(plan.Agents) != 2 {
		t.Fatalf("expected all agents for generic fallback, got %#v", plan.Agents)
	}
	if len(plan.Workflows) != 2 {
		t.Fatalf("expected all workflows for generic fallback, got %#v", plan.Workflows)
	}
}

func TestUseCaseRequiresRegistryReader(t *testing.T) {
	_, err := UseCase{}.Run(context.Background(), Options{Task: "Test ETL"})
	if err == nil {
		t.Fatalf("expected missing registry reader error")
	}
}

func TestUseCasePropagatesRegistryLoadError(t *testing.T) {
	expected := errors.New("registry missing")
	uc := UseCase{RegistryReader: fakeRegistryReader{err: expected}}

	_, err := uc.Run(context.Background(), Options{Task: "Test ETL"})
	if !errors.Is(err, expected) {
		t.Fatalf("expected registry error, got %v", err)
	}
}

func registryFixture() ports.BQARegistry {
	return ports.BQARegistry{
		Agents: []ports.BQARegistryItem{
			{ID: "etl-qa-agent", Path: "agents/etl-qa-agent.md", Domain: "etl"},
			{ID: "runtime-agent", Path: "agents/runtime-agent.md", Domain: "runtime"},
		},
		Skills: []ports.BQARegistryItem{
			{ID: "etl-log-investigation", Path: "skills/etl-log-investigation.md", Domain: "etl"},
			{ID: "runtime-trace-review", Path: "skills/runtime-trace-review.md", Domain: "runtime"},
		},
		Workflows: []ports.BQARegistryItem{
			{ID: "etl-verification-workflow", Path: "workflows/etl-verification-workflow.md", Domain: "etl"},
			{ID: "session-knowledge-workflow", Path: "workflows/session-knowledge-workflow.md", Domain: "memory"},
		},
	}
}

func placeholderMarkers() []string {
	return []string{"TO" + "DO", "FIX" + "ME"}
}
