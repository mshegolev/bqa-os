package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// Compile-time assertion that GovernanceStore satisfies ports.GovernanceStore.
var _ ports.GovernanceStore = GovernanceStore{}

// GovernanceStore is the filesystem implementation of ports.GovernanceStore. It
// reads and writes governance memory files under a memory directory and performs
// no parsing (that lives in core/memgov).
type GovernanceStore struct{}

func (GovernanceStore) ReadFile(ctx context.Context, memoryDir string, name string) (string, bool, error) {
	select {
	case <-ctx.Done():
		return "", false, ctx.Err()
	default:
	}
	cleaned, err := cleanName(name)
	if err != nil {
		return "", false, err
	}
	path := filepath.Join(governanceDir(memoryDir), cleaned)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}

func (GovernanceStore) WriteFile(ctx context.Context, memoryDir string, name string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	cleaned, err := cleanName(name)
	if err != nil {
		return err
	}
	path := filepath.Join(governanceDir(memoryDir), cleaned)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

// cleanName validates a governance file name: it must be a simple relative name
// that does not escape the memory directory. Callers pass literal constants, but
// the guard keeps the adapter safe if that ever changes.
func cleanName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("governance store: empty file name")
	}
	cleaned := filepath.Clean(name)
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("governance store: invalid file name %q", name)
	}
	return cleaned, nil
}

func governanceDir(memoryDir string) string {
	if memoryDir == "" {
		return ".bqa/memory"
	}
	return memoryDir
}
