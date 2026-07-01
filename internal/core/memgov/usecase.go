package memgov

import "github.com/mshegolev/bqa-os/internal/ports"

// UseCase governs candidate QA memory: learn candidates, review pending ones, and
// promote/reject them by id. All side effects go through the two ports.
type UseCase struct {
	Reader    ports.NormalizedSessionReader
	Store     ports.GovernanceStore
	MemoryDir string
}

// dir returns the configured memory dir or the default.
func (u UseCase) dir() string {
	if u.MemoryDir == "" {
		return DefaultMemoryDir
	}
	return u.MemoryDir
}
