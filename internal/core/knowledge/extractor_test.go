package knowledge

import (
	"context"
	"errors"
	"testing"
)

type fakeSessionReader struct {
	sessions []Session
	err      error
}

func (f fakeSessionReader) ListNormalizedSessions(ctx context.Context) ([]Session, error) {
	return f.sessions, f.err
}

type fakeArtifactWriter struct {
	artifacts []KnowledgeArtifact
	err       error
}

func (f *fakeArtifactWriter) WriteKnowledgeArtifacts(ctx context.Context, artifacts []KnowledgeArtifact) error {
	f.artifacts = artifacts
	return f.err
}

func TestExtractArtifactsDetectsETLPatterns(t *testing.T) {
	sessions := []Session{{
		ID:      "s1",
		Title:   "ETL QA session",
		Path:    ".bqa/input/sessions/normalized/s1.md",
		Content: "We tested an ETL pipeline. Need source to target row count reconciliation. Also check incremental partition backfill. Failure: target count mismatch after backfill.",
	}}

	artifacts := ExtractArtifacts(sessions)

	assertContainsArtifact(t, artifacts, ArtifactKindPattern, DomainETL, "Validate ETL source-to-target reconciliation")
	assertContainsArtifact(t, artifacts, ArtifactKindPattern, DomainETL, "Check partition and incremental load boundaries")
	assertContainsArtifact(t, artifacts, ArtifactKindCommonBug, DomainETL, "ETL jobs may fail because of mismatched counts or incremental windows")
	assertContainsArtifact(t, artifacts, ArtifactKindProjectProfile, DomainGeneral, "Initial project QA profile")
}

func TestExtractArtifactsDetectsGraphQLPatterns(t *testing.T) {
	sessions := []Session{{
		ID:      "s1",
		Title:   "GraphQL QA session",
		Content: "GraphQL schema validation found a resolver nullability mismatch. Mutation should be verified by querying state after side effect. Prompt: Create GraphQL functional tests for schema fields, resolver behavior, and mutation side effects.",
	}}

	artifacts := ExtractArtifacts(sessions)

	assertContainsArtifact(t, artifacts, ArtifactKindPattern, DomainGraphQL, "Validate GraphQL schema and resolver behavior together")
	assertContainsArtifact(t, artifacts, ArtifactKindPattern, DomainGraphQL, "Verify GraphQL mutation side effects")
	assertContainsArtifact(t, artifacts, ArtifactKindSuccessfulPrompt, DomainGraphQL, "Reusable QA prompt")
}

func TestExtractArtifactsDedupeIsDeterministic(t *testing.T) {
	sessions := []Session{
		{ID: "s2", Title: "Second", Content: "ETL source target row count reconciliation failure mismatch."},
		{ID: "s1", Title: "First", Content: "ETL source target row count reconciliation failure mismatch."},
	}

	artifacts := ExtractArtifacts(sessions)
	var reconciliation KnowledgeArtifact
	for _, artifact := range artifacts {
		if artifact.Title == "Validate ETL source-to-target reconciliation" {
			reconciliation = artifact
			break
		}
	}
	if len(reconciliation.SessionIDs) != 2 {
		t.Fatalf("expected deduped artifact to reference 2 sessions, got %#v", reconciliation.SessionIDs)
	}
	if reconciliation.SessionIDs[0] != "s1" || reconciliation.SessionIDs[1] != "s2" {
		t.Fatalf("expected deterministic sorted session ids, got %#v", reconciliation.SessionIDs)
	}
}

func TestExtractorBuildUsesReaderAndWriter(t *testing.T) {
	writer := &fakeArtifactWriter{}
	extractor := NewExtractor(fakeSessionReader{sessions: []Session{{
		ID:      "s1",
		Title:   "API session",
		Content: "REST API contract check should validate status code headers payload.",
	}}}, writer)

	result, err := extractor.Build(context.Background())
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if result.SessionsProcessed != 1 {
		t.Fatalf("expected 1 processed session, got %d", result.SessionsProcessed)
	}
	if len(writer.artifacts) == 0 {
		t.Fatal("expected writer to receive artifacts")
	}
	assertContainsArtifact(t, writer.artifacts, ArtifactKindPattern, DomainAPI, "Validate API contract at response boundary")
}

func TestExtractorBuildReturnsHelpfulErrorForMissingReader(t *testing.T) {
	extractor := NewExtractor(nil, &fakeArtifactWriter{})
	_, err := extractor.Build(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "knowledge session reader is nil" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractorBuildPropagatesReaderError(t *testing.T) {
	wantErr := errors.New("missing normalized sessions")
	extractor := NewExtractor(fakeSessionReader{err: wantErr}, &fakeArtifactWriter{})
	_, err := extractor.Build(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected reader error %v, got %v", wantErr, err)
	}
}

func assertContainsArtifact(t *testing.T, artifacts []KnowledgeArtifact, kind ArtifactKind, domain Domain, title string) {
	t.Helper()
	for _, artifact := range artifacts {
		if artifact.Kind == kind && artifact.Domain == domain && artifact.Title == title {
			return
		}
	}
	t.Fatalf("expected artifact kind=%q domain=%q title=%q, got %#v", kind, domain, title, artifacts)
}
