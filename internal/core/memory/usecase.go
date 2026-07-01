package memory

import "github.com/mshegolev/bqa-os/internal/ports"

// UseCase orchestrates memory export/import over ports only.
type UseCase struct {
	Source    ports.MemorySource
	Auditor   ports.MemoryAuditor
	Installer ports.MemoryInstaller
	Brain     ports.BrainStore
	Writers   map[string]ports.ArchiveWriter // keyed "zip","tar"
	Readers   map[string]ports.ArchiveReader // keyed "zip","tar"
}
