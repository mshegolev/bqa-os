package artifacts

import (
	"context"
	"strings"
	"testing"
)

type fakeWriter struct {
	files map[string]string
}

func (f fakeWriter) WriteBQAArtifact(ctx context.Context, relativePath string, content string) error {
	f.files[relativePath] = content
	return nil
}

func TestUseCaseGeneratesStarterArtifacts(t *testing.T) {
	writer := fakeWriter{files: map[string]string{}}
	uc := UseCase{Writer: writer}

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.ArtifactsCreated != 10 {
		t.Fatalf("expected 10 generated artifacts, got %d", result.ArtifactsCreated)
	}
	if !strings.Contains(writer.files["skills/etl-log-investigation.md"], "ETL Log Investigation") {
		t.Fatalf("missing ETL skill artifact")
	}
	if !strings.Contains(writer.files["agents/runtime-agent.md"], "Runtime Agent") {
		t.Fatalf("missing runtime agent artifact")
	}
	if !strings.Contains(writer.files["workflows/session-knowledge-workflow.md"], "bqa build") {
		t.Fatalf("missing session workflow artifact")
	}
	if !strings.Contains(writer.files["registry/index.yaml"], "registry:") {
		t.Fatalf("missing registry index artifact")
	}
}
