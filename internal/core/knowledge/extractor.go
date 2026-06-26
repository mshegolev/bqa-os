package knowledge

import (
	"context"
	"errors"
	"sort"
	"strings"
)

// SessionReader is the input port required by the core extractor.
// Filesystem, GitHub, or test adapters can implement it without leaking into core.
type SessionReader interface {
	ListNormalizedSessions(ctx context.Context) ([]Session, error)
}

// ArtifactWriter is the output port required by the core extractor.
// The core package produces domain artifacts; adapters decide how to persist them.
type ArtifactWriter interface {
	WriteKnowledgeArtifacts(ctx context.Context, artifacts []KnowledgeArtifact) error
}

// Extractor builds reusable QA knowledge from normalized sessions.
type Extractor struct {
	reader SessionReader
	writer ArtifactWriter
}

func NewExtractor(reader SessionReader, writer ArtifactWriter) *Extractor {
	return &Extractor{reader: reader, writer: writer}
}

func (e *Extractor) Build(ctx context.Context) (BuildResult, error) {
	if e == nil {
		return BuildResult{}, errors.New("knowledge extractor is nil")
	}
	if e.reader == nil {
		return BuildResult{}, errors.New("knowledge session reader is nil")
	}
	if e.writer == nil {
		return BuildResult{}, errors.New("knowledge artifact writer is nil")
	}

	sessions, err := e.reader.ListNormalizedSessions(ctx)
	if err != nil {
		return BuildResult{}, err
	}

	artifacts := ExtractArtifacts(sessions)
	if err := e.writer.WriteKnowledgeArtifacts(ctx, artifacts); err != nil {
		return BuildResult{}, err
	}

	return BuildResult{SessionsProcessed: len(sessions), Artifacts: artifacts}, nil
}

// ExtractArtifacts is a deterministic heuristic extractor for the MVP.
func ExtractArtifacts(sessions []Session) []KnowledgeArtifact {
	var artifacts []KnowledgeArtifact

	for _, session := range sessions {
		content := strings.TrimSpace(session.Content)
		if content == "" {
			continue
		}

		domains := detectDomains(content)
		if len(domains) == 0 {
			domains = []Domain{DomainGeneral}
		}

		for _, domain := range domains {
			artifacts = append(artifacts, extractPatterns(session, domain)...)
			artifacts = append(artifacts, extractCommonBugs(session, domain)...)
			artifacts = append(artifacts, extractSuccessfulPrompts(session, domain)...)
		}
	}

	artifacts = append(artifacts, buildProjectProfile(sessions))
	return dedupeArtifacts(artifacts)
}

func dedupeArtifacts(input []KnowledgeArtifact) []KnowledgeArtifact {
	seen := make(map[string]KnowledgeArtifact, len(input))

	for _, artifact := range input {
		key := strings.Join([]string{
			string(artifact.Kind),
			string(artifact.Domain),
			normalizeKey(artifact.Title),
		}, "::")

		existing, ok := seen[key]
		if !ok {
			seen[key] = normalizeArtifact(artifact)
			continue
		}

		existing.SessionIDs = mergeStrings(existing.SessionIDs, artifact.SessionIDs)
		existing.Evidence = mergeStrings(existing.Evidence, artifact.Evidence)
		existing.Tags = mergeStrings(existing.Tags, artifact.Tags)
		seen[key] = normalizeArtifact(existing)
	}

	out := make([]KnowledgeArtifact, 0, len(seen))
	for _, artifact := range seen {
		out = append(out, normalizeArtifact(artifact))
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		if out[i].Domain != out[j].Domain {
			return out[i].Domain < out[j].Domain
		}
		return out[i].Title < out[j].Title
	})

	return out
}

func normalizeArtifact(artifact KnowledgeArtifact) KnowledgeArtifact {
	artifact.Title = strings.TrimSpace(artifact.Title)
	artifact.Summary = strings.TrimSpace(artifact.Summary)
	artifact.SessionIDs = mergeStrings(artifact.SessionIDs, nil)
	artifact.Evidence = mergeStrings(artifact.Evidence, nil)
	artifact.Tags = mergeStrings(artifact.Tags, nil)
	sort.Strings(artifact.SessionIDs)
	sort.Strings(artifact.Evidence)
	sort.Strings(artifact.Tags)
	return artifact
}

func mergeStrings(left, right []string) []string {
	seen := make(map[string]struct{}, len(left)+len(right))
	merged := make([]string, 0, len(left)+len(right))

	for _, value := range append(left, right...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		merged = append(merged, value)
	}

	return merged
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Join(strings.Fields(value), " ")
}
