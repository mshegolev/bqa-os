package memgov

import (
	"context"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
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

type fakeSessions struct {
	index ports.SessionIndex
	files map[string]string
}

func (f fakeSessions) LoadSessionIndex(context.Context) (ports.SessionIndex, error) {
	return f.index, nil
}

func (f fakeSessions) ReadNormalizedSession(_ context.Context, path string) (string, error) {
	return f.files[path], nil
}

func learnFixture() (UseCase, *memStore) {
	store := newMemStore()
	reader := fakeSessions{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "/s1.jsonl", NormalizedPath: "s1.md"},
		}},
		files: map[string]string{
			"s1.md": "We ran the ETL reconciliation and compared source table vs target table row count. " +
				"The job failed with a traceback during the null check.",
		},
	}
	return UseCase{Reader: reader, Store: store, MemoryDir: ".bqa/memory"}, store
}

// hasItem reports whether items contains one with the given domain and status.
func hasItem(items []MemoryItem, domain, status string) bool {
	for _, it := range items {
		if it.Domain == domain && it.Status == status {
			return true
		}
	}
	return false
}

func TestLearnExtractsPendingCandidates(t *testing.T) {
	uc, store := learnFixture()
	res, err := uc.Learn(context.Background())
	if err != nil {
		t.Fatalf("Learn: %v", err)
	}
	if res.SessionsProcessed != 1 {
		t.Fatalf("expected 1 session processed, got %d", res.SessionsProcessed)
	}
	if res.SkillsAdded == 0 || res.LessonsAdded == 0 {
		t.Fatalf("expected skills and lessons added, got %+v", res)
	}
	skills := store.files[".bqa/memory/skill_candidates.yaml"]
	if !hasItem(parseItems(skills), "etl", StatusPending) {
		t.Fatalf("skill_candidates missing pending etl item:\n%s", skills)
	}
	if !hasItem(parseItems(skills), "data_quality", StatusPending) {
		t.Fatalf("skill_candidates missing pending data_quality item:\n%s", skills)
	}
	if res.SkillsAdded != 2 {
		t.Fatalf("expected exactly 2 skill candidates (etl + data_quality), got %d", res.SkillsAdded)
	}
	// Evidence must be bounded (no raw dumps) and single line.
	for _, it := range parseItems(skills) {
		if len(it.Evidence) > evidenceWindow {
			t.Fatalf("evidence not bounded: %d chars", len(it.Evidence))
		}
		if strings.Contains(it.Evidence, "\n") {
			t.Fatalf("evidence should be single-line")
		}
	}
	lessons := store.files[".bqa/memory/lessons_learned.yaml"]
	if !hasItem(parseItems(lessons), "bugs", StatusPending) {
		t.Fatalf("lessons_learned missing bugs item:\n%s", lessons)
	}
}

func TestLearnIsIdempotent(t *testing.T) {
	uc, _ := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("first Learn: %v", err)
	}
	res, err := uc.Learn(context.Background())
	if err != nil {
		t.Fatalf("second Learn: %v", err)
	}
	if res.SkillsAdded != 0 || res.LessonsAdded != 0 {
		t.Fatalf("second learn should add nothing, got %+v", res)
	}
}
