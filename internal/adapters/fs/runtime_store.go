package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	cleaned := filepath.Clean(relativePath)
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
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
