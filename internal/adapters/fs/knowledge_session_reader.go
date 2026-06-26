package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type KnowledgeSessionReader struct {
	SessionBaseDir string
}

func (r KnowledgeSessionReader) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	select {
	case <-ctx.Done():
		return ports.SessionIndex{}, ctx.Err()
	default:
	}

	path := filepath.Join(r.sessionBase(), "index.json")
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

func (r KnowledgeSessionReader) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	resolved, err := r.safeNormalizedPath(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r KnowledgeSessionReader) safeNormalizedPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("normalized session path is empty")
	}

	base, err := filepath.Abs(r.sessionBase())
	if err != nil {
		return "", err
	}
	normalizedBase := filepath.Join(base, "normalized")

	candidate := path
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(base, candidate)
	}
	candidate, err = filepath.Abs(filepath.Clean(candidate))
	if err != nil {
		return "", err
	}

	if candidate != normalizedBase && !strings.HasPrefix(candidate, normalizedBase+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe normalized session path: %s", path)
	}
	if filepath.Ext(candidate) != ".md" {
		return "", fmt.Errorf("normalized session path must be a markdown file: %s", path)
	}
	return candidate, nil
}

func (r KnowledgeSessionReader) sessionBase() string {
	if r.SessionBaseDir == "" {
		return ".bqa/input/sessions"
	}
	return r.SessionBaseDir
}
