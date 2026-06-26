package ports

import "context"

type NormalizedSessionReader interface {
	LoadSessionIndex(ctx context.Context) (SessionIndex, error)
	ReadNormalizedSession(ctx context.Context, path string) (string, error)
}

type KnowledgeArtifactWriter interface {
	WriteKnowledgeArtifact(ctx context.Context, filename string, content string) error
}
