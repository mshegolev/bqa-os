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
		index: ports.SessionIndex{Entries: []ports.SessionIndexEntry{
			{OriginalPath: "/tmp/session.jsonl", NormalizedPath: "s1.md"},
			{OriginalPath: "/Users/me/.factory/logs/droid-log-single.log", NormalizedPath: "droid.md"},
		}},
		files: map[string]string{
			"s1.md":    "Task: test GraphQL query, REST API endpoint, ETL reconciliation, and null check data validation",
			"droid.md": "Droid mission worker-transcripts run command failed with traceback",
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
	if result.ArtifactsCreated != 7 {
		t.Fatalf("expected 7 MVP artifacts, got %d", result.ArtifactsCreated)
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
	if len(writer.files) != len(expectedFiles) {
		t.Fatalf("expected %d written files, got %d: %#v", len(expectedFiles), len(writer.files), writer.files)
	}
	for _, filename := range expectedFiles {
		if _, ok := writer.files[filename]; !ok {
			t.Fatalf("expected artifact %s to be written", filename)
		}
	}
	if _, ok := writer.files["droid_patterns.yaml"]; ok {
		t.Fatalf("droid_patterns.yaml is outside the MVP artifact contract")
	}
	if _, ok := writer.files["runtime_patterns.yaml"]; ok {
		t.Fatalf("runtime_patterns.yaml is outside the MVP artifact contract")
	}
	if !strings.Contains(writer.files["graphql_patterns.yaml"], "graphql_functional_testing") {
		t.Fatalf("expected graphql finding, got %s", writer.files["graphql_patterns.yaml"])
	}
	if !strings.Contains(writer.files["etl_patterns.yaml"], "etl_validation") {
		t.Fatalf("expected etl finding, got %s", writer.files["etl_patterns.yaml"])
	}
	if strings.Contains(writer.files["graphql_patterns.yaml"], "github_graphql_url") {
		t.Fatalf("graphql artifact should not be driven by github_graphql_url noise")
	}
	if strings.Contains(writer.files["project_profile.yaml"], "droid") || strings.Contains(writer.files["project_profile.yaml"], "runtime") {
		t.Fatalf("project profile should only include MVP signals, got %s", writer.files["project_profile.yaml"])
	}
}
