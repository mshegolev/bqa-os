package brainstore

import (
	"context"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/brain"
	"github.com/mshegolev/bqa-os/internal/ports"
)

type GitBrainStore struct{}

// Push materializes the assembled files into the connected brain cache, then
// commits and pushes via the existing brain.Sync.
func (GitBrainStore) Push(ctx context.Context, files []ports.ArchiveFile, sanitize bool) error {
	cacheDir, err := brain.CacheDir()
	if err != nil {
		return err
	}
	for _, f := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		dst := filepath.Join(cacheDir, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, f.Data, 0o600); err != nil {
			return err
		}
	}
	return brain.Sync(sanitize)
}
