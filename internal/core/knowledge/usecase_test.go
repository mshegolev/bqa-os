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

func TestUseCaseBuildsKnowledgeArtifacts(t *testing.T) {
	reader := fakeReader{
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{{NormalizedPath: "s1.md"}}},
		files: map[string]string{"s1.md": "Task: test GraphQL query and ETL reconciliation with null checks"},
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
	if !strings.Contains(writer.files["graphql_patterns.yaml"], "graphql_functional_testing") {
		t.Fatalf("expected graphql finding, got %s", writer.files["graphql_patterns.yaml"])
	}
	if !strings.Contains(writer.files["etl_patterns.yaml"], "etl_validation") {
		t.Fatalf("expected etl finding, got %s", writer.files["etl_patterns.yaml"])
	}
}
