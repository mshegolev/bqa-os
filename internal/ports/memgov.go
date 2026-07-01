package ports

import "context"

// GovernanceStore reads and writes the governance memory files under a memory
// directory. It is deliberately file-level (bytes in, bytes out) so that all
// YAML parse/render logic lives in core/memgov and this adapter stays pure I/O.
type GovernanceStore interface {
	// ReadFile returns the file content. exists is false (with nil error) when
	// the file is absent, so a first run reads an empty governance state.
	ReadFile(ctx context.Context, memoryDir string, name string) (content string, exists bool, err error)
	// WriteFile writes content to the named file under memoryDir, creating the
	// directory if needed.
	WriteFile(ctx context.Context, memoryDir string, name string, content string) error
}
