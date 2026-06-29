package demoarchive

import (
	"context"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeArchiveWriter struct {
	outputPath string
	files      []ports.DemoArchiveFile
}

func (f *fakeArchiveWriter) WriteDemoArchive(ctx context.Context, outputPath string, files []ports.DemoArchiveFile) error {
	f.outputPath = outputPath
	f.files = append([]ports.DemoArchiveFile(nil), files...)
	return nil
}

func TestUseCaseWritesSyntheticDemoArchive(t *testing.T) {
	writer := &fakeArchiveWriter{}
	uc := UseCase{Writer: writer}

	result, err := uc.Run(context.Background(), "demo.zip")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.OutputPath != "demo.zip" {
		t.Fatalf("expected output path demo.zip, got %q", result.OutputPath)
	}
	if result.FilesCreated != len(writer.files) {
		t.Fatalf("expected result to report %d files, got %d", len(writer.files), result.FilesCreated)
	}
	if writer.outputPath != "demo.zip" {
		t.Fatalf("writer received output path %q", writer.outputPath)
	}

	expectedPaths := []string{
		"manifest.json",
		"normalized_sessions/session_001_api_regression.md",
		"normalized_sessions/session_002_graphql_contract.md",
		"normalized_sessions/session_003_data_quality.md",
		"agents/api-regression-agent.md",
		"agents/graphql-contract-agent.md",
		"workflows/api-regression-workflow.md",
		"workflows/data-quality-triage-workflow.md",
		"specs/api-regression-spec.md",
		"specs/graphql-contract-spec.md",
		"knowledge/api_patterns.yaml",
		"knowledge/data_quality_patterns.yaml",
		"knowledge/project_profile.yaml",
		"README_NEXT_STEPS.md",
		"recommendations.md",
	}

	if len(writer.files) != len(expectedPaths) {
		t.Fatalf("expected %d archive files, got %d", len(expectedPaths), len(writer.files))
	}
	for i, expectedPath := range expectedPaths {
		if writer.files[i].Path != expectedPath {
			t.Fatalf("file %d path = %q, expected %q", i, writer.files[i].Path, expectedPath)
		}
		content := strings.ToLower(writer.files[i].Content)
		if !strings.Contains(content, "synthetic") {
			t.Fatalf("file %s must be clearly synthetic, content: %q", writer.files[i].Path, writer.files[i].Content)
		}
		for _, forbidden := range []string{"password", "secret", "token", "private key"} {
			if strings.Contains(content, forbidden) {
				t.Fatalf("file %s contains forbidden sensitive marker %q", writer.files[i].Path, forbidden)
			}
		}
	}
}
