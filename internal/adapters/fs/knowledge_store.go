package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type KnowledgeStore struct {
	SessionBaseDir string
	KnowledgeDir   string
}

func (s KnowledgeStore) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	reader := KnowledgeSessionReader{SessionBaseDir: s.SessionBaseDir}
	return reader.LoadSessionIndex(ctx)
}

func (s KnowledgeStore) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	reader := KnowledgeSessionReader{SessionBaseDir: s.SessionBaseDir}
	return reader.ReadNormalizedSession(ctx, path)
}

func (s KnowledgeStore) ReadKnowledgeArtifact(ctx context.Context, filename string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	path, err := s.safeKnowledgeArtifactPath(filename)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s KnowledgeStore) WriteKnowledgeArtifact(ctx context.Context, filename string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path, err := s.safeKnowledgeArtifactPath(filename)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

func (s KnowledgeStore) WriteBQAArtifact(ctx context.Context, relativePath string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path := filepath.Join(filepath.Dir(s.knowledgeDir()), filepath.Clean(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

func (s KnowledgeStore) safeKnowledgeArtifactPath(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("knowledge artifact filename is empty")
	}
	if filepath.IsAbs(filename) || filename != filepath.Base(filename) {
		return "", fmt.Errorf("invalid knowledge artifact filename: %s", filename)
	}
	if filepath.Ext(filename) != ".yaml" {
		return "", fmt.Errorf("knowledge artifact must be a yaml file: %s", filename)
	}
	return filepath.Join(s.knowledgeDir(), filename), nil
}

func (s KnowledgeStore) knowledgeDir() string {
	if s.KnowledgeDir == "" {
		return ".bqa/knowledge"
	}
	return s.KnowledgeDir
}
