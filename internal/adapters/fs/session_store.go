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

type SessionStore struct {
	BaseDir string
	seq     int
}

func (s *SessionStore) SaveRaw(ctx context.Context, session ports.RawSession) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	s.seq++
	ext := strings.ToLower(filepath.Ext(session.Ref.Path))
	if ext == "" {
		ext = ".txt"
	}
	path := filepath.Join(s.base(), "raw", session.Ref.Source, fmt.Sprintf("%06d-%s%s", s.seq, session.SHA256[:12], ext))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return path, os.WriteFile(path, session.Bytes, 0o600)
}

func (s *SessionStore) SaveNormalized(ctx context.Context, session ports.NormalizedSession) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	relativePath := filepath.Join("normalized", session.Ref.Source, fmt.Sprintf("%06d-%s.md", s.seq, session.SHA256[:12]))
	path := filepath.Join(s.base(), relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	return relativePath, os.WriteFile(path, []byte(session.Content), 0o600)
}

func (s *SessionStore) SaveIndex(ctx context.Context, index ports.SessionIndex) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path := filepath.Join(s.base(), "index.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func (s *SessionStore) base() string {
	if s.BaseDir == "" {
		return ".bqa/input/sessions"
	}
	return s.BaseDir
}
