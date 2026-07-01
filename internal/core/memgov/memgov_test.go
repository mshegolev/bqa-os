package memgov

import (
	"context"
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

// memStore is an in-memory GovernanceStore keyed by "<dir>/<name>".
type memStore struct{ files map[string]string }

func newMemStore() *memStore { return &memStore{files: map[string]string{}} }

func (m *memStore) ReadFile(_ context.Context, dir, name string) (string, bool, error) {
	c, ok := m.files[dir+"/"+name]
	return c, ok, nil
}

func (m *memStore) WriteFile(_ context.Context, dir, name, content string) error {
	m.files[dir+"/"+name] = content
	return nil
}

func TestSaveLoadStateRoundTrip(t *testing.T) {
	store := newMemStore()
	want := GovernanceState{
		Lessons:         []MemoryItem{{ID: "lesson-1", Name: "l", Domain: "bugs", Evidence: "e", Source: "s", Status: StatusPending}},
		SkillCandidates: []MemoryItem{{ID: "skill-1", Name: "n", Domain: "etl", Evidence: "e", Source: "s", Status: StatusPending}},
		Approved:        []MemoryItem{{ID: "skill-2", Name: "a", Domain: "etl", Evidence: "e", Source: "s", Status: StatusApproved}},
		Rejected:        []MemoryItem{{ID: "skill-3", Name: "r", Domain: "etl", Evidence: "e", Source: "s", Status: StatusRejected}},
		Log:             []DecisionEntry{{ID: "skill-1", Action: "promoted", Name: "n"}},
	}
	if err := saveState(context.Background(), store, ".bqa/memory", want); err != nil {
		t.Fatalf("saveState: %v", err)
	}
	got, err := loadState(context.Background(), store, ".bqa/memory")
	if err != nil {
		t.Fatalf("loadState: %v", err)
	}
	if len(got.Lessons) != 1 || got.Lessons[0].ID != "lesson-1" {
		t.Fatalf("lessons not round-tripped: %#v", got.Lessons)
	}
	if len(got.SkillCandidates) != 1 || got.SkillCandidates[0].ID != "skill-1" {
		t.Fatalf("skill candidates not round-tripped: %#v", got.SkillCandidates)
	}
	if len(got.Approved) != 1 || got.Approved[0].Status != StatusApproved {
		t.Fatalf("approved not round-tripped: %#v", got.Approved)
	}
	if len(got.Rejected) != 1 || got.Rejected[0].Status != StatusRejected {
		t.Fatalf("rejected not round-tripped: %#v", got.Rejected)
	}
	if len(got.Log) != 1 || got.Log[0].Action != "promoted" {
		t.Fatalf("log not round-tripped: %#v", got.Log)
	}
}

func TestLoadStateEmptyWhenAbsent(t *testing.T) {
	got, err := loadState(context.Background(), newMemStore(), ".bqa/memory")
	if err != nil {
		t.Fatalf("loadState: %v", err)
	}
	if len(got.SkillCandidates) != 0 || len(got.Lessons) != 0 || len(got.Log) != 0 {
		t.Fatalf("expected empty state, got %#v", got)
	}
}
