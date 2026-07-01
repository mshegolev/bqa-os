package memgov

import (
	"context"
	"fmt"
)

// Promote moves a pending candidate to approved_patterns.
func (u UseCase) Promote(ctx context.Context, id string) (DecideResult, error) {
	return u.decide(ctx, id, StatusApproved, "promoted")
}

// Reject moves a pending candidate to rejected_patterns.
func (u UseCase) Reject(ctx context.Context, id string) (DecideResult, error) {
	return u.decide(ctx, id, StatusRejected, "rejected")
}

// decide is the shared promote/reject transition. It finds a pending candidate by
// id, sets its status, moves it to the approved/rejected list, removes it from its
// candidate list, and appends a decision_log entry. An unknown or already-decided
// id returns an error and writes nothing.
func (u UseCase) decide(ctx context.Context, id, newStatus, action string) (DecideResult, error) {
	state, err := loadState(ctx, u.Store, u.dir())
	if err != nil {
		return DecideResult{}, err
	}

	item, found := takePendingCandidate(&state, id)
	if !found {
		return DecideResult{}, fmt.Errorf("no pending candidate with id %q", id)
	}
	item.Status = newStatus
	switch newStatus {
	case StatusApproved:
		state.Approved = append(state.Approved, item)
	case StatusRejected:
		state.Rejected = append(state.Rejected, item)
	default:
		return DecideResult{}, fmt.Errorf("internal: unknown decision status %q", newStatus)
	}
	state.Log = append(state.Log, DecisionEntry{ID: item.ID, Action: action, Name: item.Name})

	if err := saveState(ctx, u.Store, u.dir(), state); err != nil {
		return DecideResult{}, err
	}
	return DecideResult{Item: item, Action: action}, nil
}

// takePendingCandidate removes and returns the pending candidate with the given
// id from whichever candidate list holds it. Reports false if no pending
// candidate matches.
func takePendingCandidate(state *GovernanceState, id string) (MemoryItem, bool) {
	if item, rest, ok := removePending(state.Lessons, id); ok {
		state.Lessons = rest
		return item, true
	}
	if item, rest, ok := removePending(state.SkillCandidates, id); ok {
		state.SkillCandidates = rest
		return item, true
	}
	return MemoryItem{}, false
}

// removePending returns the matching pending item, the list without it, and
// whether it was found.
func removePending(items []MemoryItem, id string) (MemoryItem, []MemoryItem, bool) {
	for i, it := range items {
		if it.ID == id && it.Status == StatusPending {
			rest := make([]MemoryItem, 0, len(items)-1)
			rest = append(rest, items[:i]...)
			rest = append(rest, items[i+1:]...)
			return it, rest, true
		}
	}
	return MemoryItem{}, nil, false
}
