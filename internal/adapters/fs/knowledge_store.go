package fs

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type KnowledgeStore struct {
	SessionBaseDir string
	KnowledgeDir   string
}

func (s KnowledgeStore) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	return KnowledgeSessionReader{SessionBaseDir: s.SessionBaseDir}.LoadSessionIndex(ctx)
}

func (s KnowledgeStore) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	return KnowledgeSessionReader{SessionBaseDir: s.SessionBaseDir}.ReadNormalizedSession(ctx, path)
}

func (s KnowledgeStore) ReadKnowledgeArtifact(ctx context.Context, filename string) (string, error) {
	return KnowledgeWriter{KnowledgeDir: s.KnowledgeDir}.ReadKnowledgeArtifact(ctx, filename)
}

func (s KnowledgeStore) WriteKnowledgeArtifact(ctx context.Context, filename string, content string) error {
	return KnowledgeWriter{KnowledgeDir: s.KnowledgeDir}.WriteKnowledgeArtifact(ctx, filename, content)
}

func (s KnowledgeStore) WriteBQAArtifact(ctx context.Context, relativePath string, content string) error {
	return KnowledgeWriter{KnowledgeDir: s.KnowledgeDir}.WriteBQAArtifact(ctx, relativePath, content)
}
