package ports

import "context"

type ETLAgentPackInputReader interface {
	KnowledgeArtifactReader
	NormalizedSessionReader
}

type ETLAgentPackWriter interface {
	WriteETLAgentPackArtifact(ctx context.Context, relativePath string, content string) error
}
