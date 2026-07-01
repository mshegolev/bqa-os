package memgov

import (
	"crypto/sha256"
	"encoding/hex"
)

// ItemID returns a stable, content-derived id "<prefix>-<8 hex>" from the
// candidate kind, name, and source. The prefix comes from kindPrefix, which only
// maps the candidate kinds (KindLessons, KindSkillCandidates); ItemID must be
// called only for those. Stable across insertions/removals so the governance
// files diff and merge cleanly. The hash input format (kind|name|source) is a
// persisted contract — changing it invalidates ids already written to disk.
func ItemID(kind, name, source string) string {
	sum := sha256.Sum256([]byte(kind + "|" + name + "|" + source))
	return kindPrefix[kind] + "-" + hex.EncodeToString(sum[:])[:8]
}
