package fs

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/brain"
	"github.com/mshegolev/bqa-os/internal/ports"
)

type MemoryInstaller struct{}

// InstallMemory delegates to the existing brain.Install so bundle placement
// keeps the same allow-list and non-destructive guarantees.
func (MemoryInstaller) InstallMemory(ctx context.Context, sourceDir, target string) (ports.InstalledSummary, error) {
	res, err := brain.Install(sourceDir, target)
	if err != nil {
		return ports.InstalledSummary{}, err
	}
	return ports.InstalledSummary{Target: res.BqaDir, Files: res.Files}, nil
}
