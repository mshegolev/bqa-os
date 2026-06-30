package runtime

import (
	"context"
	"strings"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type fakeWriter struct {
	files map[string]string
	order []string
}

func newFakeWriter() *fakeWriter {
	return &fakeWriter{files: map[string]string{}}
}

func (f *fakeWriter) WriteRuntimeArtifact(ctx context.Context, relativePath string, content string) error {
	if f.files == nil {
		f.files = map[string]string{}
	}
	f.files[relativePath] = content
	f.order = append(f.order, relativePath)
	return nil
}

type fakeDetector struct {
	found map[string]string
}

func (f fakeDetector) Detect(binary string) (string, bool) {
	path, ok := f.found[binary]
	return path, ok
}

var _ ports.RuntimeArtifactWriter = (*fakeWriter)(nil)
var _ ports.RuntimeDetector = fakeDetector{}

func TestPrepareWritesMasterContext(t *testing.T) {
	writer := newFakeWriter()
	detector := fakeDetector{found: map[string]string{"claude": "/usr/local/bin/claude"}}
	uc := UseCase{Writer: writer, Detector: detector}

	res, err := uc.Prepare(context.Background(), "claude")
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}

	if res.ContextPath != masterContextPath {
		t.Errorf("ContextPath = %q, want %q", res.ContextPath, masterContextPath)
	}
	content, ok := writer.files[masterContextPath]
	if !ok {
		t.Fatalf("master context not written to %q", masterContextPath)
	}
	if !strings.Contains(content, "through the claude runtime") {
		t.Errorf("master context missing runtime marker, got:\n%s", content)
	}
	if !res.Detected {
		t.Errorf("Detected = false, want true")
	}
	if res.BinaryPath != "/usr/local/bin/claude" {
		t.Errorf("BinaryPath = %q, want %q", res.BinaryPath, "/usr/local/bin/claude")
	}
	if res.Runtime != "claude" {
		t.Errorf("Runtime = %q, want %q", res.Runtime, "claude")
	}
	if res.Command != "claude" {
		t.Errorf("Command = %q, want %q", res.Command, "claude")
	}
}

func TestPrepareRejectsUnknownRuntime(t *testing.T) {
	uc := UseCase{Writer: newFakeWriter(), Detector: fakeDetector{}}

	_, err := uc.Prepare(context.Background(), "bogus")
	if err == nil {
		t.Fatalf("Prepare(bogus) returned nil error, want error")
	}
}
