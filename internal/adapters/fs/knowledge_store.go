package fs

import (
	"context"
	"encoding/json"
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
	select {
	case <-ctx.Done():
		return ports.SessionIndex{}, ctx.Err()
	default:
	}
	path := filepath.Join(s.sessionBase(), "index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ports.SessionIndex{}, fmt.Errorf("session index not found: %s\nRun `bqa discover` and `bqa ingest2` first", path)
		}
		return ports.SessionIndex{}, err
	}
	var index ports.SessionIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return ports.SessionIndex{}, err
	}
	return index, nil
}

func (s KnowledgeStore) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	data, err := os.ReadFile(path)
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
