package app

import (
	"context"

	fsadapter "github.com/mshegolev/bqa-os/internal/adapters/fs"
	coreartifacts "github.com/mshegolev/bqa-os/internal/core/artifacts"
	coreknowledge "github.com/mshegolev/bqa-os/internal/core/knowledge"
)

const (
	defaultSessionBaseDir = ".bqa/input/sessions"
	defaultKnowledgeDir   = ".bqa/knowledge"
)

// BuildOptions contains CLI-level configuration for bqa build.
// It intentionally has no Cobra dependency so it can be tested directly.
type BuildOptions struct {
	SessionBaseDir string
	KnowledgeDir   string
}

// BuildSummary is the command-facing summary printed by the thin Cobra layer.
type BuildSummary struct {
	SessionsProcessed        int
	KnowledgeArtifactsCreated int
	BQAArtifactsCreated       int
	KnowledgeDir              string
	GeneratedDirs             []string
}

// RunBuild wires filesystem adapters into core use cases.
// Cobra should only parse flags, call this function, and print the result.
func RunBuild(ctx context.Context, opts BuildOptions) (BuildSummary, error) {
	sessionBaseDir := withDefault(opts.SessionBaseDir, defaultSessionBaseDir)
	knowledgeDir := withDefault(opts.KnowledgeDir, defaultKnowledgeDir)

	store := fsadapter.KnowledgeStore{SessionBaseDir: sessionBaseDir, KnowledgeDir: knowledgeDir}

	knowledgeUC := coreknowledge.UseCase{Reader: store, Writer: store, OutputDir: knowledgeDir}
	knowledgeResult, err := knowledgeUC.Run(ctx)
	if err != nil {
		return BuildSummary{}, err
	}

	artifactUC := coreartifacts.UseCase{Writer: store}
	artifactResult, err := artifactUC.Run(ctx)
	if err != nil {
		return BuildSummary{}, err
	}

	return BuildSummary{
		SessionsProcessed:        knowledgeResult.SessionsProcessed,
		KnowledgeArtifactsCreated: knowledgeResult.ArtifactsCreated,
		BQAArtifactsCreated:       artifactResult.ArtifactsCreated,
		KnowledgeDir:              knowledgeDir,
		GeneratedDirs: []string{
			".bqa/skills",
			".bqa/agents",
			".bqa/workflows",
			".bqa/registry",
		},
	}, nil
}

func withDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
