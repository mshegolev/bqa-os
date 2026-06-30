package catalog

import (
	"strings"
	"testing"
)

func TestRenderAgentGeneric(t *testing.T) {
	core := ETLQA()
	out := RenderAgent(core, nil)
	if !strings.HasPrefix(out, "# ETL QA Agent\n") {
		t.Fatalf("expected title prefix, got: %q", out)
	}
	if !strings.Contains(out, "## Operating Rules") {
		t.Fatalf("expected Operating Rules section, got: %q", out)
	}
	for _, rule := range core.Rules {
		if !strings.Contains(out, "- "+rule+"\n") {
			t.Fatalf("expected rule %q in output", rule)
		}
	}
}

func TestRenderAgentRuntimeFlavor(t *testing.T) {
	core := ETLQA()
	out := RenderAgent(core, CodexFlavor())
	if !strings.HasPrefix(out, "# Codex ETL QA Agent\n") {
		t.Fatalf("expected codex title prefix, got: %q", out)
	}
	if !strings.Contains(out, codexFlavor.Intro) {
		t.Fatalf("expected intro in output")
	}
	for _, rule := range codexFlavor.ExtraRules {
		if !strings.Contains(out, "- "+rule+"\n") {
			t.Fatalf("expected codex extra rule %q in output", rule)
		}
	}
	if !strings.Contains(out, "- "+core.Rules[0]+"\n") {
		t.Fatalf("expected base rule still present")
	}
}

func TestCatalogLookups(t *testing.T) {
	if Skill("etl-log-investigation").Content == "" {
		t.Fatalf("expected skill content")
	}
	if Workflow("session-knowledge-workflow").Domain != "memory" {
		t.Fatalf("expected memory domain, got %q", Workflow("session-knowledge-workflow").Domain)
	}
	if !strings.Contains(RuntimeAgentContent(), "Runtime Agent") {
		t.Fatalf("expected Runtime Agent in runtime agent content")
	}
}

func TestRegistryYAMLGolden(t *testing.T) {
	wantAgents := "agents:\n  - id: etl-qa-agent\n    path: agents/etl-qa-agent.md\n    domain: etl\n  - id: runtime-agent\n    path: agents/runtime-agent.md\n    domain: runtime\n"
	if got := RegistryAgentsYAML(); got != wantAgents {
		t.Fatalf("agents YAML mismatch:\nwant %q\ngot  %q", wantAgents, got)
	}
	wantSkills := "skills:\n  - id: etl-log-investigation\n    path: skills/etl-log-investigation.md\n    domain: etl\n  - id: runtime-trace-review\n    path: skills/runtime-trace-review.md\n    domain: runtime\n"
	if got := RegistrySkillsYAML(); got != wantSkills {
		t.Fatalf("skills YAML mismatch:\nwant %q\ngot  %q", wantSkills, got)
	}
	wantWorkflows := "workflows:\n  - id: etl-verification-workflow\n    path: workflows/etl-verification-workflow.md\n    domain: etl\n  - id: session-knowledge-workflow\n    path: workflows/session-knowledge-workflow.md\n    domain: memory\n"
	if got := RegistryWorkflowsYAML(); got != wantWorkflows {
		t.Fatalf("workflows YAML mismatch:\nwant %q\ngot  %q", wantWorkflows, got)
	}
}
