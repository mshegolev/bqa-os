package memgov

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// loadState reads every governance file via the store and parses it. A missing
// file yields an empty list, so a first run starts from an empty state.
func loadState(ctx context.Context, store ports.GovernanceStore, memoryDir string) (GovernanceState, error) {
	var st GovernanceState
	read := func(kind string) ([]MemoryItem, error) {
		content, exists, err := store.ReadFile(ctx, memoryDir, fileName(kind))
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, nil
		}
		return parseItems(content), nil
	}

	var err error
	if st.Lessons, err = read(KindLessons); err != nil {
		return GovernanceState{}, err
	}
	if st.SkillCandidates, err = read(KindSkillCandidates); err != nil {
		return GovernanceState{}, err
	}
	if st.Approved, err = read(KindApproved); err != nil {
		return GovernanceState{}, err
	}
	if st.Rejected, err = read(KindRejected); err != nil {
		return GovernanceState{}, err
	}

	logContent, exists, err := store.ReadFile(ctx, memoryDir, fileName(KindDecisionLog))
	if err != nil {
		return GovernanceState{}, err
	}
	if exists {
		st.Log = parseLog(logContent)
	}
	return st, nil
}

// saveState writes all governance files deterministically. Re-writing unchanged
// files is safe because render output is stable. Writes are per-file, not atomic
// across files: if the process is interrupted mid-save, re-running `learn`
// restores any dropped candidate.
func saveState(ctx context.Context, store ports.GovernanceStore, memoryDir string, st GovernanceState) error {
	byKind := map[string][]MemoryItem{
		KindLessons:         st.Lessons,
		KindSkillCandidates: st.SkillCandidates,
		KindApproved:        st.Approved,
		KindRejected:        st.Rejected,
	}
	for _, kind := range itemKinds {
		if err := store.WriteFile(ctx, memoryDir, fileName(kind), renderItems(kind, byKind[kind])); err != nil {
			return err
		}
	}
	return store.WriteFile(ctx, memoryDir, fileName(KindDecisionLog), renderLog(st.Log))
}

// idSet returns the set of every id present anywhere in the state, used to keep
// learn idempotent.
func (s GovernanceState) idSet() map[string]bool {
	set := map[string]bool{}
	for _, list := range [][]MemoryItem{s.Lessons, s.SkillCandidates, s.Approved, s.Rejected} {
		for _, it := range list {
			set[it.ID] = true
		}
	}
	return set
}
