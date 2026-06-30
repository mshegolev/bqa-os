package ports

import "context"

// RuntimeArtifact is a single entry from the BQA unified registry: an agent,
// skill, workflow, guardrail, memory index, or project profile.
type RuntimeArtifact struct {
	ID          string
	Type        string
	Title       string
	Source      string
	Destination string
	Summary     string
	Tags        []string
}

// RuntimeRegistry is the parsed BQA unified registry.
type RuntimeRegistry struct {
	Version   int
	Kind      string
	Name      string
	Artifacts []RuntimeArtifact
}

// RuntimeRegistryReader loads the BQA unified registry and the raw source
// content of each artifact it references.
type RuntimeRegistryReader interface {
	LoadRuntimeRegistry(ctx context.Context) (RuntimeRegistry, error)
	ReadArtifactSource(ctx context.Context, source string) (string, error)
}

// RuntimeArtifactWriter writes a generated runtime file (relative to a target
// repository root) for a specific AI coding runtime.
type RuntimeArtifactWriter interface {
	WriteRuntimeArtifact(ctx context.Context, relativePath string, content string) error
}
