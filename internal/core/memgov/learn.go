package memgov

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"
)

// evidenceWindow bounds a captured snippet: wide enough for a QA context sentence,
// far below a transcript body so no raw session leaks verbatim.
const evidenceWindow = 480

// skillRule maps signal words to a stable skill-candidate name + domain.
type skillRule struct {
	needles []string
	name    string
	domain  string
}

var skillRules = []skillRule{
	{[]string{"reconciliation", "row count", "source table", "target table"}, "etl reconciliation check", "etl"},
	{[]string{"data quality", "null check", "duplicate", "schema drift", "checksum"}, "data quality validation check", "data_quality"},
	{[]string{"rest api", "endpoint", "http status", "status code", "contract test"}, "api contract check", "api"},
	{[]string{"graphql query", "graphql mutation", "graphql schema", "graphql resolver"}, "graphql operation check", "graphql"},
}

// failureNeedles signal a lesson worth remembering. Matching is plain substring
// (so "failed" also matches within larger words); word-boundary precision is a
// deferred enhancement, intentional for this keyword-heuristic slice.
var failureNeedles = []string{"traceback", "exception", "failed", "failure", "error:", "panic", "regression", "flaky"}

// Learn reads normalized sessions, extracts skill + lesson candidates with bounded
// evidence, and merges any new ones (by id) into the candidate files as pending.
// Idempotent: an id already present anywhere in the state is skipped.
func (u UseCase) Learn(ctx context.Context) (LearnResult, error) {
	index, err := u.Reader.LoadSessionIndex(ctx)
	if err != nil {
		return LearnResult{}, err
	}
	state, err := loadState(ctx, u.Store, u.dir())
	if err != nil {
		return LearnResult{}, err
	}
	seen := state.idSet()
	res := LearnResult{}

	for _, entry := range index.Entries {
		body, err := u.Reader.ReadNormalizedSession(ctx, entry.NormalizedPath)
		if err != nil {
			return LearnResult{}, fmt.Errorf("read normalized session %q: %w", entry.NormalizedPath, err)
		}
		res.SessionsProcessed++
		text := collapse(body)
		lower := strings.ToLower(text)

		for _, item := range extractSkillCandidates(text, lower, entry.NormalizedPath) {
			if seen[item.ID] {
				continue
			}
			seen[item.ID] = true
			state.SkillCandidates = append(state.SkillCandidates, item)
			res.SkillsAdded++
		}
		for _, item := range extractLessons(text, lower, entry.NormalizedPath) {
			if seen[item.ID] {
				continue
			}
			seen[item.ID] = true
			state.Lessons = append(state.Lessons, item)
			res.LessonsAdded++
		}
	}

	if err := saveState(ctx, u.Store, u.dir(), state); err != nil {
		return LearnResult{}, err
	}
	return res, nil
}

// extractSkillCandidates returns one pending skill candidate per matched rule.
func extractSkillCandidates(text, lower, source string) []MemoryItem {
	var out []MemoryItem
	for _, rule := range skillRules {
		if needle, ok := firstMatch(lower, rule.needles); ok {
			out = append(out, MemoryItem{
				ID:       ItemID(KindSkillCandidates, rule.name, source),
				Name:     rule.name,
				Domain:   rule.domain,
				Evidence: snippet(text, lower, needle),
				Source:   source,
				Status:   StatusPending,
			})
		}
	}
	return out
}

// extractLessons returns at most one pending lesson per session (the first
// failure signal), so lessons stay high-signal.
func extractLessons(text, lower, source string) []MemoryItem {
	needle, ok := firstMatch(lower, failureNeedles)
	if !ok {
		return nil
	}
	const name = "lesson from failure signal"
	return []MemoryItem{{
		ID:       ItemID(KindLessons, name, source),
		Name:     name,
		Domain:   "bugs",
		Evidence: snippet(text, lower, needle),
		Source:   source,
		Status:   StatusPending,
	}}
}

// firstMatch returns the first needle contained in lower and whether any matched.
func firstMatch(lower string, needles []string) (string, bool) {
	for _, n := range needles {
		if strings.Contains(lower, n) {
			return n, true
		}
	}
	return "", false
}

// collapse turns all whitespace runs into single spaces, giving single-line
// evidence and stable YAML scalars.
func collapse(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// snippet returns a byte-bounded window of the (already collapsed) text centered
// on the needle. The window is evidenceWindow bytes wide; rune-boundary snapping
// may extend it by up to 3 bytes so a multi-byte character is never split. This
// caps evidence length so a raw session body is never copied whole. Needles are
// ASCII (see skillRules/failureNeedles), so the lowercase index maps cleanly onto
// the original text's byte offsets.
func snippet(text, lower, needle string) string {
	idx := strings.Index(lower, needle)
	if idx < 0 {
		return boundRunes(text, evidenceWindow)
	}
	start := idx - 120
	if start < 0 {
		start = 0
	}
	end := start + evidenceWindow
	if end > len(text) {
		end = len(text)
	}
	for start > 0 && !utf8.RuneStart(text[start]) {
		start--
	}
	for end < len(text) && !utf8.RuneStart(text[end]) {
		end++
	}
	return strings.TrimSpace(text[start:end])
}

// boundRunes caps text at n runes on a rune boundary.
func boundRunes(text string, n int) string {
	if r := []rune(text); len(r) > n {
		return strings.TrimSpace(string(r[:n]))
	}
	return text
}
