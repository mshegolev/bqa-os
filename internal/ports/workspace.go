package ports

import "context"

// WorkspaceStore reads and writes the workspace registry file under a base
// directory. It is file-level (content in/out) so all YAML parse/render logic
// lives in core/workspace and this adapter stays pure I/O.
type WorkspaceStore interface {
	// Exists reports whether a workspace file is present under baseDir.
	Exists(ctx context.Context, baseDir string) (bool, error)
	// Load returns the raw workspace file content.
	Load(ctx context.Context, baseDir string) (content string, err error)
	// Save writes the workspace file content under baseDir, creating baseDir.
	Save(ctx context.Context, baseDir string, content string) error
}

// PathInspector answers filesystem questions the workspace use case needs when
// registering a project path, kept behind a port so both branches (missing dir,
// non-git dir) are unit-testable with a fake.
type PathInspector interface {
	// IsDir reports whether path exists and is a directory.
	IsDir(path string) (bool, error)
	// IsGitRepo reports whether path is (or is inside) a git working tree.
	IsGitRepo(path string) (bool, error)
}
