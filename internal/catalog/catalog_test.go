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
