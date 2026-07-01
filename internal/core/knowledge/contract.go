package knowledge

import "strings"

// SchemaVersion is the current knowledge artifact schema version. Every artifact
// written by bqa build carries it as `schema_version`. Additive field changes
// keep this version; removing/renaming a field or changing its meaning bumps it.
const SchemaVersion = 1

// ArtifactSpec describes a single knowledge artifact produced by `bqa build`.
// RootKey is the top-level YAML key expected inside the file (filename minus
// the ".yaml" extension).
type ArtifactSpec struct {
	Filename string
	RootKey  string
}

// expectedArtifacts is the single source of truth for the knowledge artifacts
// produced by `bqa build`. Both the writer (knowledge.UseCase) and the
// validator (knowledge.Validate) reference this list so the set never drifts.
var expectedArtifacts = []ArtifactSpec{
	{Filename: "etl_patterns.yaml"},
	{Filename: "graphql_patterns.yaml"},
	{Filename: "api_patterns.yaml"},
	{Filename: "data_quality_patterns.yaml"},
	{Filename: "common_bugs.yaml"},
	{Filename: "successful_prompts.yaml"},
	{Filename: "droid_patterns.yaml"},
	{Filename: "runtime_patterns.yaml"},
	{Filename: "project_profile.yaml"},
}

// ExpectedArtifacts returns the contract of knowledge artifacts that a
// successful `bqa build` writes. Each spec's RootKey is derived from its
// filename (filename without the ".yaml" suffix).
func ExpectedArtifacts() []ArtifactSpec {
	out := make([]ArtifactSpec, len(expectedArtifacts))
	for i, spec := range expectedArtifacts {
		spec.RootKey = strings.TrimSuffix(spec.Filename, ".yaml")
		out[i] = spec
	}
	return out
}
