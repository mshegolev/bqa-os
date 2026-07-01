package memgov

import (
	"strings"
	"testing"
)

func TestItemIDStableAndPrefixed(t *testing.T) {
	a := ItemID(KindSkillCandidates, "etl reconciliation check", "normalized/s1.md")
	b := ItemID(KindSkillCandidates, "etl reconciliation check", "normalized/s1.md")
	if a != b {
		t.Fatalf("id not stable: %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, "skill-") {
		t.Fatalf("expected skill- prefix, got %q", a)
	}
	if ItemID(KindLessons, "x", "y") == ItemID(KindSkillCandidates, "x", "y") {
		t.Fatalf("different kinds must yield different ids")
	}
	if !strings.HasPrefix(ItemID(KindLessons, "x", "y"), "lesson-") {
		t.Fatalf("expected lesson- prefix")
	}
	// Pin the concrete id so the on-disk format cannot silently drift.
	if got := ItemID(KindSkillCandidates, "etl reconciliation check", "normalized/s1.md"); got != "skill-4a4ffa34" {
		t.Fatalf("ItemID format drifted: %q", got)
	}
}

func TestRenderParseItemsRoundTrip(t *testing.T) {
	items := []MemoryItem{
		{ID: "skill-0a1b2c3d", Name: `a "quoted" name`, Domain: "etl", Evidence: "line one\nline two", Source: "normalized/s1.md", Status: "pending"},
	}
	out := renderItems(KindSkillCandidates, items)
	if !strings.Contains(out, "schema_version: 1\n") || !strings.Contains(out, "kind: skill_candidates\n") {
		t.Fatalf("missing v1 envelope:\n%s", out)
	}
	got := parseItems(out)
	if len(got) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got))
	}
	if got[0] != items[0] {
		t.Fatalf("round-trip mismatch:\n want %#v\n got  %#v", items[0], got[0])
	}
}

func TestRenderEmptyItems(t *testing.T) {
	out := renderItems(KindApproved, nil)
	if !strings.Contains(out, "items: []\n") {
		t.Fatalf("expected empty items list, got:\n%s", out)
	}
	if len(parseItems(out)) != 0 {
		t.Fatalf("expected zero parsed items")
	}
}

func TestRenderParseLogRoundTrip(t *testing.T) {
	entries := []DecisionEntry{{ID: "skill-0a1b2c3d", Action: "promoted", Name: "etl check"}}
	out := renderLog(entries)
	if !strings.Contains(out, "kind: decision_log\n") {
		t.Fatalf("missing decision_log kind:\n%s", out)
	}
	got := parseLog(out)
	if len(got) != 1 || got[0] != entries[0] {
		t.Fatalf("log round-trip mismatch: %#v", got)
	}
}
