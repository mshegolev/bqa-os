package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type KnowledgeWriter struct {
	KnowledgeDir string
}

func (w KnowledgeWriter) ReadKnowledgeArtifact(ctx context.Context, filename string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	path, err := w.knowledgeArtifactPath(filename)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (w KnowledgeWriter) WriteKnowledgeArtifact(ctx context.Context, filename string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path, err := w.knowledgeArtifactPath(filename)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

func (w KnowledgeWriter) WriteBQAArtifact(ctx context.Context, relativePath string, content string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path, err := w.bqaArtifactPath(relativePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

func (w KnowledgeWriter) knowledgeArtifactPath(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("knowledge artifact filename is required")
	}
	cleaned := filepath.Clean(filename)
	if cleaned == "." || filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid knowledge artifact filename %q", filename)
	}
	return filepath.Join(w.knowledgeDir(), cleaned), nil
}

func (w KnowledgeWriter) bqaArtifactPath(relativePath string) (string, error) {
	if relativePath == "" {
		return "", fmt.Errorf("BQA artifact path is required")
	}
	cleaned := filepath.Clean(relativePath)
	if cleaned == "." || filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid BQA artifact path %q", relativePath)
	}
	return filepath.Join(filepath.Dir(w.knowledgeDir()), cleaned), nil
}

func (w KnowledgeWriter) knowledgeDir() string {
	if w.KnowledgeDir == "" {
		return ".bqa/knowledge"
	}
	return w.KnowledgeDir
}
