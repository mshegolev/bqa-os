package fs

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// Compile-time checks that the fs types satisfy the workspace ports.
var (
	_ ports.WorkspaceStore = WorkspaceStore{}
	_ ports.PathInspector  = PathInspector{}
)

// WorkspaceStore is the filesystem implementation of ports.WorkspaceStore. It
// reads/writes <baseDir>/workspace.yaml and performs no parsing.
type WorkspaceStore struct{}

func workspaceFilePath(baseDir string) string {
	if baseDir == "" {
		baseDir = ".bqa"
	}
	return filepath.Join(baseDir, "workspace.yaml")
}

func (WorkspaceStore) Exists(ctx context.Context, baseDir string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}
	_, err := os.Stat(workspaceFilePath(baseDir))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (WorkspaceStore) Load(ctx context.Context, baseDir string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	data, err := os.ReadFile(workspaceFilePath(baseDir))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (WorkspaceStore) Save(ctx context.Context, baseDir string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path := workspaceFilePath(baseDir)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

// PathInspector is the filesystem implementation of ports.PathInspector.
type PathInspector struct{}

func (PathInspector) IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// IsGitRepo walks up from path looking for a `.git` entry (a directory for a
// normal repo, or a file for a linked worktree), so any location inside a
// working tree reports true.
func (PathInspector) IsGitRepo(path string) (bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	for {
		_, err := os.Stat(filepath.Join(abs, ".git"))
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return false, nil
		}
		abs = parent
	}
}
