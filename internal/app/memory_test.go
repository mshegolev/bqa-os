package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// writeSessionFixture writes a normalized session + index.json under base so the
// memory command has something to learn from.
func writeSessionFixture(t *testing.T, base string) {
	t.Helper()
	normDir := filepath.Join(base, "normalized")
	if err := os.MkdirAll(normDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := "ETL reconciliation compared source table vs target table row count; job failed with a traceback."
	if err := os.WriteFile(filepath.Join(normDir, "s1.md"), []byte(body), 0o600); err != nil {
		t.Fatalf("write session: %v", err)
	}
	index := ports.SessionIndex{Entries: []ports.SessionIndexEntry{
		{OriginalPath: "/s1.jsonl", NormalizedPath: "normalized/s1.md"},
	}}
	data, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("marshal index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "index.json"), data, 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}
}

func runMemory(t *testing.T, args ...string) string {
	t.Helper()
	cmd := memoryCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("memory %v: %v\noutput:\n%s", args, err, out.String())
	}
	return out.String()
}

func TestMemoryLearnReviewPromoteFlow(t *testing.T) {
	dir := t.TempDir()
	sessions := filepath.Join(dir, "sessions")
	memoryDir := filepath.Join(dir, "memory")
	writeSessionFixture(t, sessions)

	learnOut := runMemory(t, "learn", "--sessions", sessions, "--memory-dir", memoryDir)
	if !strings.Contains(learnOut, "Skill candidates added:") {
		t.Fatalf("learn output missing counts:\n%s", learnOut)
	}

	reviewOut := runMemory(t, "review", "--memory-dir", memoryDir)
	if !strings.Contains(reviewOut, "etl reconciliation check") {
		t.Fatalf("review output missing pending item:\n%s", reviewOut)
	}

	// Grab an id from the rendered file and promote it.
	content, err := os.ReadFile(filepath.Join(memoryDir, "skill_candidates.yaml"))
	if err != nil {
		t.Fatalf("read skill_candidates: %v", err)
	}
	id := firstIDInYAML(t, string(content))
	promoteOut := runMemory(t, "promote", id, "--memory-dir", memoryDir)
	if !strings.Contains(promoteOut, "promoted") {
		t.Fatalf("promote output missing confirmation:\n%s", promoteOut)
	}

	approved, err := os.ReadFile(filepath.Join(memoryDir, "approved_patterns.yaml"))
	if err != nil {
		t.Fatalf("read approved_patterns: %v", err)
	}
	if !strings.Contains(string(approved), "id: "+id) {
		t.Fatalf("promoted id not in approved_patterns:\n%s", approved)
	}
}

func TestMemoryRejectFlow(t *testing.T) {
	dir := t.TempDir()
	sessions := filepath.Join(dir, "sessions")
	memoryDir := filepath.Join(dir, "memory")
	writeSessionFixture(t, sessions)

	runMemory(t, "learn", "--sessions", sessions, "--memory-dir", memoryDir)

	content, err := os.ReadFile(filepath.Join(memoryDir, "lessons_learned.yaml"))
	if err != nil {
		t.Fatalf("read lessons_learned: %v", err)
	}
	id := firstIDInYAML(t, string(content))

	rejectOut := runMemory(t, "reject", id, "--memory-dir", memoryDir)
	if !strings.Contains(rejectOut, "rejected") {
		t.Fatalf("reject output missing confirmation:\n%s", rejectOut)
	}
	rejected, err := os.ReadFile(filepath.Join(memoryDir, "rejected_patterns.yaml"))
	if err != nil {
		t.Fatalf("read rejected_patterns: %v", err)
	}
	if !strings.Contains(string(rejected), "id: "+id) {
		t.Fatalf("rejected id not in rejected_patterns:\n%s", rejected)
	}
	reviewOut := runMemory(t, "review", "--memory-dir", memoryDir)
	if strings.Contains(reviewOut, id) {
		t.Fatalf("rejected id should not appear in review:\n%s", reviewOut)
	}
}

// firstIDInYAML returns the first "- id: <value>" from a rendered items file.
func firstIDInYAML(t *testing.T, content string) string {
	t.Helper()
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- id:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "- id:"))
		}
	}
	t.Fatalf("no id found in:\n%s", content)
	return ""
}
