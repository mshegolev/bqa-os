package runtimeemit

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeReader struct {
	registry ports.RuntimeRegistry
	sources  map[string]string
}

func (f fakeReader) LoadRuntimeRegistry(ctx context.Context) (ports.RuntimeRegistry, error) {
	if len(f.registry.Artifacts) == 0 {
		return ports.RuntimeRegistry{}, errors.New("empty registry")
	}
	return f.registry, nil
}

func (f fakeReader) ReadArtifactSource(ctx context.Context, source string) (string, error) {
	content, ok := f.sources[source]
	if !ok {
		return "", errors.New("missing source: " + source)
	}
	return content, nil
}

type fakeWriter struct {
	files map[string]string
}

func (f fakeWriter) WriteRuntimeArtifact(ctx context.Context, relativePath string, content string) error {
	f.files[relativePath] = content
	return nil
}

func sampleReader() fakeReader {
	return fakeReader{
		registry: ports.RuntimeRegistry{
			Version: 1,
			Kind:    "BQATeamUnifiedRegistry",
			Name:    "bqa-team-unified",
			Artifacts: []ports.RuntimeArtifact{
				{ID: "qa-agent", Type: "agent", Title: "QA Agent", Source: "roles/qa.md", Summary: "QA specialist"},
				{ID: "etl-skill", Type: "skill", Title: "ETL Skill", Source: "skills/etl.md", Summary: "ETL investigation"},
				{ID: "pipeline", Type: "workflow", Title: "Pipeline", Source: "workflows/pipe.md", Summary: "Team pipeline"},
				{ID: "local-first", Type: "guardrail", Title: "Local First", Source: "guardrails/lf.md", Summary: "Local-first rule"},
				{ID: "mem-1", Type: "memory_index", Title: "Memory", Source: "memory/m1.md", Summary: "Memory index"},
			},
		},
		sources: map[string]string{
			"roles/qa.md":       "# QA Agent\nBody.",
			"skills/etl.md":     "# ETL\nBody.",
			"workflows/pipe.md": "# Pipeline\nBody.",
			"guardrails/lf.md":  "# Local First\nBody.",
			"memory/m1.md":      "# Memory\nBody.",
		},
	}
}

func TestRunEmitsAllRuntimes(t *testing.T) {
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: sampleReader(), Writer: writer}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.ArtifactsRead != 5 {
		t.Fatalf("expected 5 artifacts read, got %d", result.ArtifactsRead)
	}

	want := []string{
		".claude/agents/qa-agent.md",
		".claude/skills/etl-skill/SKILL.md",
		".claude/commands/pipeline.md",
		".claude/bqa/guardrail/local-first.md",
		".claude/bqa/memory_index/mem-1.md",
		".codex/AGENTS.md",
		".codex/prompts/etl-skill.md",
		".codex/prompts/pipeline.md",
		".codex/bqa/memory_index/mem-1.md",
		".opencode/agent/qa-agent.md",
		".opencode/command/etl-skill.md",
		".opencode/command/pipeline.md",
		".opencode/bqa/guardrail/local-first.md",
		".opencode/bqa/memory_index/mem-1.md",
	}
	for _, path := range want {
		if _, ok := writer.files[path]; !ok {
			t.Fatalf("expected file %q to be emitted; got files: %v", path, keys(writer.files))
		}
	}
	if result.FilesWritten != len(writer.files) {
		t.Fatalf("FilesWritten %d != actual %d", result.FilesWritten, len(writer.files))
	}
}

func TestClaudeAgentHasFrontmatter(t *testing.T) {
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: sampleReader(), Writer: writer, Targets: []string{"claude"}}
	if _, err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	agent := writer.files[".claude/agents/qa-agent.md"]
	if !strings.HasPrefix(agent, "---\nname: qa-agent\ndescription: \"QA specialist\"\n---\n") {
		t.Fatalf("claude agent frontmatter missing or malformed:\n%s", agent)
	}
}

func TestCodexConsolidatesAgents(t *testing.T) {
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: sampleReader(), Writer: writer, Targets: []string{"codex"}}
	if _, err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	agents := writer.files[".codex/AGENTS.md"]
	if !strings.Contains(agents, "QA Agent") || !strings.Contains(agents, "Local First") {
		t.Fatalf("codex AGENTS.md should consolidate agent and guardrail:\n%s", agents)
	}
	if _, ok := writer.files[".codex/agents/qa-agent.md"]; ok {
		t.Fatalf("codex should not emit per-agent files; agents are consolidated")
	}
}

func TestUnsupportedTargetIsRejected(t *testing.T) {
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: sampleReader(), Writer: writer, Targets: []string{"vim"}}
	if _, err := uc.Run(context.Background()); err == nil {
		t.Fatalf("expected error for unsupported target")
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
