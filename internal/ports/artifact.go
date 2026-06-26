package ports

import "context"

type KnowledgeArtifactReader interface {
	ReadKnowledgeArtifact(ctx context.Context, filename string) (string, error)
}

type BQAArtifactWriter interface {
	WriteBQAArtifact(ctx context.Context, relativePath string, content string) error
}
