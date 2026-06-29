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
			return ports.SessionIndex{}, fmt.Errorf("session index not found at %s; run `bqa discover` and `bqa ingest` first, or pass --sessions with a directory containing .bqa/input/sessions/index.json", path)
		}
		return ports.SessionIndex{}, err
	}

	var index ports.SessionIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return ports.SessionIndex{}, fmt.Errorf("parse session index %s: %w", path, err)
	}
	return index, nil
}

func (r KnowledgeSessionReader) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	resolved, err := r.normalizedSessionPath(path)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(resolved)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r KnowledgeSessionReader) normalizedSessionPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("normalized session path is required")
	}
	if filepath.IsAbs(path) {
		return path, nil
	}

	cleaned := filepath.Clean(path)
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid normalized session path %q: path must stay under %s", path, r.sessionBase())
	}
	return filepath.Join(r.sessionBase(), cleaned), nil
}

func (r KnowledgeSessionReader) sessionBase() string {
	if r.SessionBaseDir == "" {
		return ".bqa/input/sessions"
	}
	return r.SessionBaseDir
}
