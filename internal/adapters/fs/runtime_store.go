package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/pathsafe"
)

// RuntimeStore writes generated runtime files under TargetDir. When DryRun is
// true it validates paths but writes nothing.
type RuntimeStore struct {
	TargetDir string
	DryRun    bool
}

func (s RuntimeStore) WriteRuntimeArtifact(ctx context.Context, relativePath string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cleaned, ok := pathsafe.RelClean(relativePath)
	if !ok {
		return fmt.Errorf("unsafe target path: %s", relativePath)
	}
	if s.DryRun {
		return nil
	}

	path := filepath.Join(s.TargetDir, cleaned)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
