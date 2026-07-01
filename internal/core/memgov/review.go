package memgov

import "context"

// Review returns the pending candidates (lessons first, then skills) for display.
func (u UseCase) Review(ctx context.Context) (ReviewResult, error) {
	state, err := loadState(ctx, u.Store, u.dir())
	if err != nil {
		return ReviewResult{}, err
	}
	var pending []MemoryItem
	pending = append(pending, filterPending(state.Lessons)...)
	pending = append(pending, filterPending(state.SkillCandidates)...)
	return ReviewResult{Pending: pending}, nil
}

// filterPending returns only items still in the pending state.
func filterPending(items []MemoryItem) []MemoryItem {
	var out []MemoryItem
	for _, it := range items {
		if it.Status == StatusPending {
			out = append(out, it)
		}
	}
	return out
}
