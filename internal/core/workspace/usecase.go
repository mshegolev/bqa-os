package workspace

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// UseCase manages the workspace registry: initialize it, register projects, and
// list them. All side effects go through the WorkspaceStore and PathInspector
// ports.
type UseCase struct {
	Store     ports.WorkspaceStore
	Inspector ports.PathInspector
	// BaseDir is the BQA base directory; leave empty to use ".bqa" (the default is
	// applied by baseDir(), not by reading this field directly).
	BaseDir string
}

func (u UseCase) baseDir() string {
	if u.BaseDir == "" {
		return ".bqa"
	}
	return u.BaseDir
}

// Init creates an empty workspace registry. It errors if one already exists,
// writing nothing.
func (u UseCase) Init(ctx context.Context, name string) (InitResult, error) {
	exists, err := u.Store.Exists(ctx, u.baseDir())
	if err != nil {
		return InitResult{}, err
	}
	if exists {
		return InitResult{}, fmt.Errorf("workspace already initialized in %s", u.baseDir())
	}
	if err := u.Store.Save(ctx, u.baseDir(), Render(Workspace{Name: name})); err != nil {
		return InitResult{}, err
	}
	return InitResult{Name: name, Path: filepath.Join(u.baseDir(), WorkspaceFileName)}, nil
}

// Add registers a project. It requires an initialized workspace and an existing
// directory path; a non-git path is recorded with a warning; a duplicate id
// errors and changes nothing.
func (u UseCase) Add(ctx context.Context, p Project) (AddResult, error) {
	exists, err := u.Store.Exists(ctx, u.baseDir())
	if err != nil {
		return AddResult{}, err
	}
	if !exists {
		return AddResult{}, fmt.Errorf("no workspace in %s; run 'bqa workspace init'", u.baseDir())
	}

	isDir, err := u.Inspector.IsDir(p.Path)
	if err != nil {
		return AddResult{}, err
	}
	if !isDir {
		return AddResult{}, fmt.Errorf("path %q is not an existing directory", p.Path)
	}

	content, err := u.Store.Load(ctx, u.baseDir())
	if err != nil {
		return AddResult{}, err
	}
	ws := Parse(content)
	for _, existing := range ws.Projects {
		if existing.ID == p.ID {
			return AddResult{}, fmt.Errorf("project id %q already exists", p.ID)
		}
	}

	var warning string
	isGit, err := u.Inspector.IsGitRepo(p.Path)
	if err != nil {
		return AddResult{}, err
	}
	if !isGit {
		warning = fmt.Sprintf("path %q is not a git repository; recorded anyway", p.Path)
	}

	ws.Projects = append(ws.Projects, p)
	if err := u.Store.Save(ctx, u.baseDir(), Render(ws)); err != nil {
		return AddResult{}, err
	}
	return AddResult{Project: p, Warning: warning}, nil
}

// List returns the registered projects. It requires an initialized workspace.
func (u UseCase) List(ctx context.Context) (ListResult, error) {
	exists, err := u.Store.Exists(ctx, u.baseDir())
	if err != nil {
		return ListResult{}, err
	}
	if !exists {
		return ListResult{}, fmt.Errorf("no workspace in %s; run 'bqa workspace init'", u.baseDir())
	}
	content, err := u.Store.Load(ctx, u.baseDir())
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{Workspace: Parse(content)}, nil
}
