package doctor

import (
	"context"
	"errors"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeInspector struct {
	entries map[string]ports.WorkspaceEntry
}

func (f fakeInspector) InspectWorkspacePath(ctx context.Context, path string) (ports.WorkspaceEntry, error) {
	entry, ok := f.entries[path]
	if !ok {
		return ports.WorkspaceEntry{Path: path}, nil
	}
	return entry, nil
}

type fakeRegistryReader struct {
	registry ports.BQARegistry
	err      error
}

func (f fakeRegistryReader) LoadBQARegistry(ctx context.Context) (ports.BQARegistry, error) {
	return f.registry, f.err
}

func TestUseCasePassesHealthyWorkspace(t *testing.T) {
	uc := UseCase{
		Inspector:      fakeInspector{entries: healthyWorkspaceEntries()},
		RegistryReader: fakeRegistryReader{registry: healthyRegistry()},
	}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if !result.Healthy {
		t.Fatalf("expected healthy workspace, got %#v", result.Checks)
	}

	for _, name := range []string{"workspace", "registry", "memory", "agents", "skills", "workflows"} {
		check := findCheck(result, name)
		if check == nil {
			t.Fatalf("missing check %q in %#v", name, result.Checks)
		}
		if check.Status != StatusPass {
			t.Fatalf("check %q status = %q, expected PASS", name, check.Status)
		}
	}
}

func TestUseCaseFailsMissingWorkspacePieces(t *testing.T) {
	uc := UseCase{
		Inspector: fakeInspector{entries: map[string]ports.WorkspaceEntry{
			".bqa":          {Path: ".bqa", Exists: true, IsDir: true},
			".bqa/registry": {Path: ".bqa/registry", Exists: true, IsDir: true},
			".bqa/agents":   {Path: ".bqa/agents", Exists: true, IsDir: true},
			".bqa/skills":   {Path: ".bqa/skills", Exists: true, IsDir: true},
		}},
		RegistryReader: fakeRegistryReader{err: errors.New("missing registry index")},
	}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Healthy {
		t.Fatalf("expected unhealthy workspace")
	}
	if findCheck(result, "memory").Status != StatusFail {
		t.Fatalf("expected missing memory check to fail, got %#v", result.Checks)
	}
	if findCheck(result, "workflows").Status != StatusFail {
		t.Fatalf("expected missing workflows check to fail, got %#v", result.Checks)
	}
	if findCheck(result, "registry").Status != StatusFail {
		t.Fatalf("expected invalid registry check to fail, got %#v", result.Checks)
	}
}

func TestUseCaseRequiresDependencies(t *testing.T) {
	_, err := UseCase{}.Run(context.Background())
	if err == nil {
		t.Fatalf("expected missing dependency error")
	}
}

func healthyWorkspaceEntries() map[string]ports.WorkspaceEntry {
	entries := map[string]ports.WorkspaceEntry{}
	for _, path := range []string{".bqa", ".bqa/registry", ".bqa/memory", ".bqa/agents", ".bqa/skills", ".bqa/workflows"} {
		entries[path] = ports.WorkspaceEntry{Path: path, Exists: true, IsDir: true}
	}
	return entries
}

func healthyRegistry() ports.BQARegistry {
	return ports.BQARegistry{
		Agents:    []ports.BQARegistryItem{{ID: "etl-qa-agent", Path: "agents/etl-qa-agent.md", Domain: "etl"}},
		Skills:    []ports.BQARegistryItem{{ID: "etl-log-investigation", Path: "skills/etl-log-investigation.md", Domain: "etl"}},
		Workflows: []ports.BQARegistryItem{{ID: "etl-verification-workflow", Path: "workflows/etl-verification-workflow.md", Domain: "etl"}},
	}
}

func findCheck(result Result, name string) *Check {
	for i := range result.Checks {
		if result.Checks[i].Name == name {
			return &result.Checks[i]
		}
	}
	return nil
}
