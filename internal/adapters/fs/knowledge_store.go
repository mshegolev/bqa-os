package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

var expectedKnowledgeArtifactFilenames = []string{
	"etl_patterns.yaml",
	"graphql_patterns.yaml",
	"api_patterns.yaml",
	"data_quality_patterns.yaml",
	"common_bugs.yaml",
	"successful_prompts.yaml",
	"project_profile.yaml",
}

type KnowledgeStore struct {
	SessionBaseDir string
	KnowledgeDir   string
}

func (s KnowledgeStore) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	select {
	case <-ctx.Done():
		return ports.SessionIndex{}, ctx.Err()
	default:
	}
	path := filepath.Join(s.sessionBase(), "index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ports.SessionIndex{}, fmt.Errorf("session index not found at %s; run bqa discover and bqa ingest2 first", path)
		}
		return ports.SessionIndex{}, fmt.Errorf("read session index at %s: %w", path, err)
	}
	var index ports.SessionIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return ports.SessionIndex{}, fmt.Errorf("invalid session index JSON at %s: %w", path, err)
	}
	return index, nil
}

func (s KnowledgeStore) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	safePath, err := safeExistingChildPath(s.sessionBase(), path)
	if err != nil {
		return "", fmt.Errorf("unsafe normalized session path %q: %w", path, err)
	}
	data, err := os.ReadFile(safePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s KnowledgeStore) ReadKnowledgeArtifact(ctx context.Context, filename string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	if err := validateKnowledgeArtifactFilename(filename); err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(s.knowledgeDir(), filename))
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
	if err := validateKnowledgeArtifactFilename(filename); err != nil {
		return err
	}
	path := filepath.Join(s.knowledgeDir(), filename)
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

func (s KnowledgeStore) sessionBase() string {
	if s.SessionBaseDir == "" {
		return ".bqa/input/sessions"
	}
	return s.SessionBaseDir
}

func (s KnowledgeStore) knowledgeDir() string {
	if s.KnowledgeDir == "" {
		return ".bqa/knowledge"
	}
	return s.KnowledgeDir
}

func safeExistingChildPath(base string, candidate string) (string, error) {
	if candidate == "" {
		return "", fmt.Errorf("path is empty")
	}

	target := filepath.Clean(candidate)
	if !filepath.IsAbs(target) {
		if err := ensurePathInside(base, target); err != nil {
			target = filepath.Join(base, target)
		}
	}
	target = filepath.Clean(target)

	if err := ensurePathInside(base, target); err != nil {
		return "", err
	}

	resolvedBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return "", fmt.Errorf("resolve base directory: %w", err)
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", err
	}
	if err := ensurePathInside(resolvedBase, resolvedTarget); err != nil {
		return "", err
	}
	return target, nil
}

func ensurePathInside(base string, target string) error {
	absBase, err := filepath.Abs(base)
	if err != nil {
		return err
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("must stay inside %s", absBase)
	}
	return nil
}

func validateKnowledgeArtifactFilename(filename string) error {
	if filename == "" || filename != filepath.Base(filename) || filename != filepath.Clean(filename) {
		return unsupportedKnowledgeArtifactFilenameError(filename)
	}
	for _, expected := range expectedKnowledgeArtifactFilenames {
		if filename == expected {
			return nil
		}
	}
	return unsupportedKnowledgeArtifactFilenameError(filename)
}

func unsupportedKnowledgeArtifactFilenameError(filename string) error {
	return fmt.Errorf("unsupported knowledge artifact filename %q: expected one of %s", filename, strings.Join(expectedKnowledgeArtifactFilenames, ", "))
}
