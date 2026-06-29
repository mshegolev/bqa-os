package ports

import "context"

type WorkspaceInspector interface {
	InspectWorkspacePath(ctx context.Context, path string) (WorkspaceEntry, error)
}

type WorkspaceEntry struct {
	Path   string
	Exists bool
	IsDir  bool
	Size   int64
}
