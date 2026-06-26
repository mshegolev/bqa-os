package knowledge

import "time"

// Artifact is the legacy rendered artifact DTO used by the current build wiring.
// Keep it until the CLI is fully migrated to KnowledgeArtifact-based writers.
type Artifact struct {
	Filename string
	Content  string
}

// Finding is the legacy intermediate finding model used by the current YAML renderer.
type Finding struct {
	Name       string
	Domain     string
	Evidence   string
	SourcePath string
}

// Result is the legacy build result used by current app/build.go wiring.
type Result struct {
	SessionsProcessed int
	ArtifactsCreated  int
	OutputDir         string
}

// ProjectProfile is the legacy rendered profile used by the current YAML renderer.
type ProjectProfile struct {
	Sessions       int
	ETLSignals     int
	GraphQLSignals int
	APISignals     int
	DQSignals      int
	DroidSignals   int
	RuntimeSignals int
}

// Session is a normalized engineering session ready for knowledge extraction.
// It intentionally contains no filesystem-specific behavior so core extraction
// can be tested without adapters or Cobra wiring.
type Session struct {
	ID        string
	Path      string
	Title     string
	Content   string
	CreatedAt time.Time
}

// Domain describes the QA domain inferred from a normalized session.
type Domain string

const (
	DomainETL         Domain = "etl"
	DomainGraphQL     Domain = "graphql"
	DomainAPI         Domain = "api"
	DomainDataQuality Domain = "data_quality"
	DomainGeneral     Domain = "general"
)

// ArtifactKind describes the type of reusable knowledge produced by the extractor.
type ArtifactKind string

const (
	ArtifactKindPattern          ArtifactKind = "pattern"
	ArtifactKindCommonBug        ArtifactKind = "common_bug"
	ArtifactKindSuccessfulPrompt ArtifactKind = "successful_prompt"
	ArtifactKindProjectProfile   ArtifactKind = "project_profile"
)

// KnowledgeArtifact is the stable domain object produced by the heuristic extractor.
// YAML tags are kept here so future filesystem writers can serialize artifacts
// without redefining simple DTOs in adapters.
type KnowledgeArtifact struct {
	Kind       ArtifactKind `yaml:"kind"`
	Domain     Domain       `yaml:"domain,omitempty"`
	Title      string       `yaml:"title"`
	Summary    string       `yaml:"summary"`
	Evidence   []string     `yaml:"evidence,omitempty"`
	SessionIDs []string     `yaml:"session_ids,omitempty"`
	Tags       []string     `yaml:"tags,omitempty"`
}

// BuildResult is returned by the core extractor and can be printed by a thin CLI layer.
type BuildResult struct {
	SessionsProcessed int
	Artifacts         []KnowledgeArtifact
}
