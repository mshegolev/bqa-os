package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestDoctorCmdPassesForValidWorkspace(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	writeSyntheticBQAWorkspace(t, tmp)

	cmd := doctorCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	output := out.String()
	for _, expected := range []string{
		"PASS workspace .bqa",
		"PASS registry .bqa/registry",
		"PASS memory .bqa/memory",
		"PASS agents .bqa/agents",
		"PASS skills .bqa/skills",
		"PASS workflows .bqa/workflows",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("doctor output missing %q, got:\n%s", expected, output)
		}
	}
}

func TestDoctorCmdFailsForMissingWorkspace(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	cmd := doctorCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected doctor to fail for missing workspace")
	}

	output := out.String()
	if !strings.Contains(output, "FAIL workspace .bqa") {
		t.Fatalf("expected missing workspace failure, got:\n%s", output)
	}
	for _, marker := range placeholderMarkers() {
		if strings.Contains(output, marker) {
			t.Fatalf("doctor output must not include placeholder marker %q, got:\n%s", marker, output)
		}
	}
}
