package knowledge

import (
	"context"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeReader struct {
	index ports.SessionIndex
	files map[string]string
}

func (f fakeReader) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	return f.index, nil
}

func (f fakeReader) ReadNormalizedSession(ctx context.Context, path string) (string, error) {
	return f.files[path], nil
}

type fakeWriter struct {
	files map[string]string
}

func (f fakeWriter) WriteKnowledgeArtifact(ctx context.Context, filename string, content string) error {
	f.files[filename] = content
	return nil
}

func TestExtractorFindsSyntheticETLKnowledge(t *testing.T) {
	session := NormalizedSession{
		SourcePath:     "synthetic/etl-session.md",
		NormalizedPath: "normalized/etl-session.md",
		NormalizedMarkdown: "Task: investigate ETL row count mismatch.\n" +
			"The aggregation had duplicate rows, null timestamps, and timezone drift.\n" +
			"Failure: retry exposed a schema drift regression.\n" +
			"Successful prompt: compare source and target row count, then verify data quality reconciliation.",
	}

	result := Extractor{}.Extract([]NormalizedSession{session})

	if result.Profile.Sessions != 1 {
		t.Fatalf("expected 1 profiled session, got %d", result.Profile.Sessions)
	}
	assertFinding(t, result.ETLPatterns, "etl_validation", "row count")
	assertFinding(t, result.DataQualityPatterns, "data_quality_validation", "duplicate")
	assertFinding(t, result.CommonBugs, "common_failure_signal", "schema drift")
	assertFinding(t, result.SuccessfulPrompts, "successful_prompt_candidate", "compare source")
}

func TestUseCaseWritesSevenKnowledgeArtifacts(t *testing.T) {
	reader := fakeReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "synthetic/etl-session.md", NormalizedPath: "etl.md"},
		}},
		files: map[string]string{
			"etl.md": "Task: verify ETL partition freshness. Duplicate row count failed after retry. Prompt: check data quality reconciliation.",
		},
	}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.SessionsProcessed != 1 {
		t.Fatalf("expected 1 processed session, got %d", result.SessionsProcessed)
	}
	if result.ArtifactsCreated != 7 {
		t.Fatalf("expected 7 artifacts, got %d", result.ArtifactsCreated)
	}

	expectedFiles := []string{
		"etl_patterns.yaml",
		"graphql_patterns.yaml",
		"api_patterns.yaml",
		"data_quality_patterns.yaml",
		"common_bugs.yaml",
		"successful_prompts.yaml",
		"project_profile.yaml",
	}
	for _, filename := range expectedFiles {
		if _, ok := writer.files[filename]; !ok {
			t.Fatalf("expected artifact %s to be written", filename)
		}
	}
	if _, ok := writer.files["droid_patterns.yaml"]; ok {
		t.Fatalf("did not expect droid_patterns.yaml to be written")
	}
	if _, ok := writer.files["runtime_patterns.yaml"]; ok {
		t.Fatalf("did not expect runtime_patterns.yaml to be written")
	}
}

func TestExtractorRedactsSensitiveEvidence(t *testing.T) {
	session := NormalizedSession{
		SourcePath:         "synthetic/api-session.md",
		NormalizedPath:     "normalized/api-session.md",
		NormalizedMarkdown: "Task: test REST API endpoint with password=secret123 and qa.owner@example.com",
	}

	result := Extractor{}.Extract([]NormalizedSession{session})

	assertFinding(t, result.APIPatterns, "api_contract_testing", "password=[REDACTED]")
	if strings.Contains(result.APIPatterns[0].Evidence, "secret123") {
		t.Fatalf("expected password value to be redacted, got %q", result.APIPatterns[0].Evidence)
	}
	if strings.Contains(result.APIPatterns[0].Evidence, "qa.owner@example.com") {
		t.Fatalf("expected email to be redacted, got %q", result.APIPatterns[0].Evidence)
	}
}

func assertFinding(t *testing.T, findings []Finding, name string, evidence string) {
	t.Helper()

	for _, finding := range findings {
		if finding.Name == name && strings.Contains(strings.ToLower(finding.Evidence), strings.ToLower(evidence)) {
			return
		}
	}
	t.Fatalf("expected finding %q with evidence containing %q, got %#v", name, evidence, findings)
}
