package workspace

import "github.com/mshegolev/bqa-os/internal/version"

// SchemaVersion is the v1 envelope version stamped on the workspace file.
const SchemaVersion = 1

// KindWorkspace is the envelope kind for the workspace registry file.
const KindWorkspace = "workspace"

// DefaultBranchRole is the branch_role assigned when --branch-role is omitted.
const DefaultBranchRole = "base"

// WorkspaceFileName is the registry filename under the base directory.
const WorkspaceFileName = "workspace.yaml"

// generatedBy is the provenance stamp: "bqa dev" in dev/test, "bqa vX.Y.Z" in a
// release build. Deterministic in tests.
func generatedBy() string { return "bqa " + version.Version }

// Project is a registered local repo/worktree root.
type Project struct {
	ID         string
	Path       string
	Repo       string
	ETL        string
	BranchRole string
}

// Task is a registered task worktree. Reserved for slice 2 (bqa task start);
// rendered/parsed now so the schema is stable.
type Task struct {
	ID     string
	Jira   string
	Repo   string
	ETL    string
	Path   string
	Branch string
}

// Workspace is the full registry: a name plus project and task lists.
type Workspace struct {
	Name     string
	Projects []Project
	Tasks    []Task
}

// InitResult reports a completed init.
type InitResult struct {
	Name string
	Path string
}

// AddResult reports an added project and an optional non-fatal warning.
type AddResult struct {
	Project Project
	Warning string
}

// ListResult holds the loaded workspace for display.
type ListResult struct {
	Workspace Workspace
}
