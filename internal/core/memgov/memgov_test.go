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

func TestReviewListsOnlyPending(t *testing.T) {
	uc, _ := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	res, err := uc.Review(context.Background())
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if len(res.Pending) == 0 {
		t.Fatalf("expected pending candidates")
	}
	for _, it := range res.Pending {
		if it.Status != StatusPending {
			t.Fatalf("review returned non-pending item: %#v", it)
		}
	}
	// Verify Lessons appear before SkillCandidates (documented ordering).
	seenSkill := false
	for _, it := range res.Pending {
		if strings.HasPrefix(it.ID, "skill-") {
			seenSkill = true
		}
		if strings.HasPrefix(it.ID, "lesson-") && seenSkill {
			t.Fatalf("lesson appeared after a skill candidate: ordering violated")
		}
	}
}

func firstPendingID(t *testing.T, uc UseCase) string {
	t.Helper()
	res, err := uc.Review(context.Background())
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	if len(res.Pending) == 0 {
		t.Fatalf("no pending candidates to decide")
	}
	return res.Pending[0].ID
}

func TestPromoteMovesItemAndLogs(t *testing.T) {
	uc, store := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	id := firstPendingID(t, uc)

	res, err := uc.Promote(context.Background(), id)
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}
	if res.Action != "promoted" || res.Item.ID != id || res.Item.Status != StatusApproved {
		t.Fatalf("unexpected promote result: %#v", res)
	}
	approved := store.files[".bqa/memory/approved_patterns.yaml"]
	if !strings.Contains(approved, "id: "+id) || !strings.Contains(approved, "status: approved") {
		t.Fatalf("approved_patterns missing promoted item:\n%s", approved)
	}
	// Removed from its candidate list.
	for _, list := range []string{"lessons_learned.yaml", "skill_candidates.yaml"} {
		if strings.Contains(store.files[".bqa/memory/"+list], "id: "+id) {
			t.Fatalf("promoted id still present in %s", list)
		}
	}
	if !strings.Contains(store.files[".bqa/memory/decision_log.yaml"], "action: promoted") {
		t.Fatalf("decision_log missing promotion entry")
	}
}

func TestRejectMovesItemToRejected(t *testing.T) {
	uc, store := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	id := firstPendingID(t, uc)

	res, err := uc.Reject(context.Background(), id)
	if err != nil {
		t.Fatalf("Reject: %v", err)
	}
	if res.Action != "rejected" || res.Item.Status != StatusRejected {
		t.Fatalf("unexpected reject result: %#v", res)
	}
	rejected := store.files[".bqa/memory/rejected_patterns.yaml"]
	if !strings.Contains(rejected, "id: "+id) || !strings.Contains(rejected, "status: rejected") {
		t.Fatalf("rejected_patterns missing item:\n%s", rejected)
	}
	for _, list := range []string{"lessons_learned.yaml", "skill_candidates.yaml"} {
		if strings.Contains(store.files[".bqa/memory/"+list], "id: "+id) {
			t.Fatalf("rejected id still present in %s", list)
		}
	}
}

func snapshotFiles(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func TestDecideTwiceErrorsAndLeavesFilesUnchanged(t *testing.T) {
	uc, store := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	id := firstPendingID(t, uc)
	if _, err := uc.Promote(context.Background(), id); err != nil {
		t.Fatalf("first Promote: %v", err)
	}
	before := snapshotFiles(store.files)

	if _, err := uc.Promote(context.Background(), id); err == nil {
		t.Fatalf("expected error promoting an already-decided id")
	}
	after := store.files
	if len(after) != len(before) {
		t.Fatalf("file set changed: before=%d after=%d", len(before), len(after))
	}
	for k, v := range before {
		if after[k] != v {
			t.Fatalf("file %q changed after failed second promote", k)
		}
	}
}

func TestPromoteUnknownIDErrors(t *testing.T) {
	uc, _ := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	if _, err := uc.Promote(context.Background(), "skill-doesnotexist"); err == nil {
		t.Fatalf("expected error for unknown id")
	}
}

func TestPromoteSkillCandidateByID(t *testing.T) {
	uc, store := learnFixture()
	if _, err := uc.Learn(context.Background()); err != nil {
		t.Fatalf("Learn: %v", err)
	}
	res, err := uc.Review(context.Background())
	if err != nil {
		t.Fatalf("Review: %v", err)
	}
	var skillID string
	for _, it := range res.Pending {
		if strings.HasPrefix(it.ID, "skill-") {
			skillID = it.ID
			break
		}
	}
	if skillID == "" {
		t.Fatalf("no skill candidate to promote")
	}
	if _, err := uc.Promote(context.Background(), skillID); err != nil {
		t.Fatalf("Promote skill: %v", err)
	}
	if strings.Contains(store.files[".bqa/memory/skill_candidates.yaml"], "id: "+skillID) {
		t.Fatalf("promoted skill id still present in skill_candidates")
	}
	if !strings.Contains(store.files[".bqa/memory/approved_patterns.yaml"], "id: "+skillID) {
		t.Fatalf("promoted skill id missing from approved_patterns")
	}
}
