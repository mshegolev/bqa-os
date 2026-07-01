package workspace

import (
	"fmt"
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
)

// Render serializes a Workspace to deterministic v1-envelope YAML.
func Render(ws Workspace) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("schema_version: %d\nkind: %s\ngenerated_by: %s\n", SchemaVersion, KindWorkspace, generatedBy()))
	b.WriteString("name: " + textutil.QuoteYAML(ws.Name) + "\n")
	b.WriteString(renderProjects(ws.Projects))
	b.WriteString(renderTasks(ws.Tasks))
	return b.String()
}

func renderProjects(projects []Project) string {
	var b strings.Builder
	if len(projects) == 0 {
		b.WriteString("projects: []\n")
		return b.String()
	}
	b.WriteString("projects:\n")
	for _, p := range projects {
		b.WriteString("  - id: " + textutil.QuoteYAML(p.ID) + "\n")
		b.WriteString("    path: " + textutil.QuoteYAML(p.Path) + "\n")
		b.WriteString("    repo: " + textutil.QuoteYAML(p.Repo) + "\n")
		b.WriteString("    etl: " + textutil.QuoteYAML(p.ETL) + "\n")
		b.WriteString("    branch_role: " + textutil.QuoteYAML(p.BranchRole) + "\n")
	}
	return b.String()
}

func renderTasks(tasks []Task) string {
	var b strings.Builder
	if len(tasks) == 0 {
		b.WriteString("tasks: []\n")
		return b.String()
	}
	b.WriteString("tasks:\n")
	for _, t := range tasks {
		b.WriteString("  - id: " + textutil.QuoteYAML(t.ID) + "\n")
		b.WriteString("    jira: " + textutil.QuoteYAML(t.Jira) + "\n")
		b.WriteString("    repo: " + textutil.QuoteYAML(t.Repo) + "\n")
		b.WriteString("    etl: " + textutil.QuoteYAML(t.ETL) + "\n")
		b.WriteString("    path: " + textutil.QuoteYAML(t.Path) + "\n")
		b.WriteString("    branch: " + textutil.QuoteYAML(t.Branch) + "\n")
	}
	return b.String()
}
