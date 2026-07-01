package guardrails

import (
	"strings"
	"testing"
)

func TestCriticalThinkingContainsCoreRules(t *testing.T) {
	got := CriticalThinking()
	for _, want := range []string{
		"Critical thinking & memory guardrails",
		"Do not invent",
		"local project memory",
		"source",
		"ask a human",
		"assumptions",
		"facts, inference, and recommendations",
		"human-in-the-loop",
		"private",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("guardrails text missing %q", want)
		}
	}
}

func TestCriticalThinkingIsDeterministic(t *testing.T) {
	if CriticalThinking() != CriticalThinking() {
		t.Fatal("CriticalThinking must be deterministic")
	}
}
