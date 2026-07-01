package memgov

import "github.com/mshegolev/bqa-os/internal/version"

// SchemaVersion is the v1 envelope version stamped on every governance file.
const SchemaVersion = 1

// Governance file kinds. Each kind maps 1:1 to a "<kind>.yaml" file.
const (
	KindLessons         = "lessons_learned"
	KindSkillCandidates = "skill_candidates"
	KindApproved        = "approved_patterns"
	KindRejected        = "rejected_patterns"
	KindDecisionLog     = "decision_log"
)

// Item statuses.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// DefaultMemoryDir is the private memory area (portable via `bqa brain`).
const DefaultMemoryDir = ".bqa/memory"

// kindPrefix is the short id prefix per candidate kind.
var kindPrefix = map[string]string{
	KindLessons:         "lesson",
	KindSkillCandidates: "skill",
}

// candidateFiles lists the two files that hold reviewable candidates.
// stateFiles lists every governance file (candidates + decided + log).
var (
	candidateFiles = []string{KindLessons, KindSkillCandidates}
	stateFiles     = []string{KindLessons, KindSkillCandidates, KindApproved, KindRejected, KindDecisionLog}
)

// fileName returns the on-disk filename for a kind.
func fileName(kind string) string { return kind + ".yaml" }

// generatedBy is the provenance stamp: "bqa dev" in dev/test, "bqa vX.Y.Z" in a
// release build. Deterministic in tests.
func generatedBy() string { return "bqa " + version.Version }

// MemoryItem is one governed memory record carrying the v1 envelope fields plus a
// governance status.
type MemoryItem struct {
	ID       string
	Name     string
	Domain   string
	Evidence string
	Source   string
	Status   string
}

// DecisionEntry is one appended decision_log record. No timestamp — decisions are
// deterministic for now (a future slice may add wall-clock).
type DecisionEntry struct {
	ID     string
	Action string // "promoted" | "rejected"
	Name   string
}

// GovernanceState is the full in-memory governance state.
type GovernanceState struct {
	Lessons         []MemoryItem
	SkillCandidates []MemoryItem
	Approved        []MemoryItem
	Rejected        []MemoryItem
	Log             []DecisionEntry
}

// LearnResult reports what a learn run added.
type LearnResult struct {
	SessionsProcessed int
	LessonsAdded      int
	SkillsAdded       int
}

// ReviewResult holds the pending candidates for display.
type ReviewResult struct {
	Pending []MemoryItem
}

// DecideResult reports the moved item and the action taken.
type DecideResult struct {
	Item   MemoryItem
	Action string // "promoted" | "rejected"
}
