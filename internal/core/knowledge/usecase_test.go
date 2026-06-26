package knowledge

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeReader struct {
	index ports.SessionIndex
	files map[string]string
	err   error
}

func (f fakeReader) LoadSessionIndex(ctx context.Context) (ports.SessionIndex, error) {
	if f.err != nil {
		return ports.SessionIndex{}, f.err
	}
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

func TestUseCaseBuildsKnowledgeArtifactsFromKeywordLines(t *testing.T) {
	reader := fakeReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "/tmp/session-1.md", NormalizedPath: "s1.md"},
			{OriginalPath: "/tmp/session-2.md", NormalizedPath: "s2.md"},
		}},
		files: map[string]string{
			"s1.md": strings.Join([]string{
				"ETL check: partition row count duplicate schema drift retry late-arriving aggregation null timestamp timezone",
				"GraphQL functional check: query mutation resolver schema nullable fragment pagination authorization",
				"API regression check: status code contract payload response auth timeout idempotency",
			}, "\n"),
			"s2.md": strings.Join([]string{
				"Data Quality validation: freshness completeness accuracy consistency reconciliation duplicate null",
				"Bug signal: regression failed with timeout and error response",
				"Task: please analyze this repository and create GraphQL tests",
			}, "\n"),
		},
	}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.SessionsProcessed != 2 {
		t.Fatalf("expected 2 processed sessions, got %d", result.SessionsProcessed)
	}
	if result.ArtifactsCreated != len(DefaultArtifactFilenames) {
		t.Fatalf("expected %d artifacts, got %d", len(DefaultArtifactFilenames), result.ArtifactsCreated)
	}
	if len(writer.files) != len(DefaultArtifactFilenames) {
		t.Fatalf("expected %d written files, got %d", len(DefaultArtifactFilenames), len(writer.files))
	}

	assertContains(t, writer.files["etl_patterns.yaml"], "partition")
	assertContains(t, writer.files["graphql_patterns.yaml"], "resolver")
	assertContains(t, writer.files["api_patterns.yaml"], "status code")
	assertContains(t, writer.files["data_quality_patterns.yaml"], "freshness")
	assertContains(t, writer.files["common_bugs.yaml"], "regression")
	assertContains(t, writer.files["successful_prompts.yaml"], "successful_prompt_candidate")
	assertContains(t, writer.files["project_profile.yaml"], "sessions_analyzed: 2")
}

func TestUseCasePropagatesSessionIndexError(t *testing.T) {
	reader := fakeReader{err: errors.New("session index not found: .bqa/input/sessions/index.json")}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	_, err := uc.Run(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "session index not found") {
		t.Fatalf("expected helpful missing index error, got %v", err)
	}
}

func TestUseCaseAlwaysCreatesDefaultArtifactSet(t *testing.T) {
	reader := fakeReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{{OriginalPath: "/tmp/session.md", NormalizedPath: "empty.md"}}},
		files: map[string]string{"empty.md": "no matching qa keywords here"},
	}
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Reader: reader, Writer: writer, OutputDir: ".bqa/knowledge"}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.ArtifactsCreated != len(DefaultArtifactFilenames) {
		t.Fatalf("expected %d artifacts, got %d", len(DefaultArtifactFilenames), result.ArtifactsCreated)
	}
	for _, filename := range DefaultArtifactFilenames {
		if _, ok := writer.files[filename]; !ok {
			t.Fatalf("expected %s to be written", filename)
		}
	}
}

func assertContains(t *testing.T, got string, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("expected %q to contain %q", got, want)
	}
}
