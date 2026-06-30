package pathsafe

import (
	"path/filepath"
	"testing"
)

func TestRelCleanAccepts(t *testing.T) {
	for _, in := range []string{"a.md", "dir/b.md", "./c.md", "x/../y.md"} {
		got, ok := RelClean(in)
		if !ok {
			t.Fatalf("RelClean(%q) rejected a safe path", in)
		}
		if got != filepath.Clean(in) {
			t.Fatalf("RelClean(%q) = %q, want %q", in, got, filepath.Clean(in))
		}
	}
}

func TestRelCleanRejects(t *testing.T) {
	for _, in := range []string{"..", "../x", "../../etc", filepath.FromSlash("/abs/path")} {
		if _, ok := RelClean(in); ok {
			t.Fatalf("RelClean(%q) accepted an unsafe path", in)
		}
	}
}
