package workspace

import (
	"strings"

	"github.com/mshegolev/bqa-os/internal/textutil"
)

// Parse reads a workspace file produced by Render. Envelope/header lines and
// unknown lines are ignored. "projects:" / "tasks:" switch the active section;
// each "- id:" starts a new entry in that section.
func Parse(content string) Workspace {
	var ws Workspace
	section := ""
	var curP *Project
	var curT *Task
	flushP := func() {
		if curP != nil {
			ws.Projects = append(ws.Projects, *curP)
			curP = nil
		}
	}
	flushT := func() {
		if curT != nil {
			ws.Tasks = append(ws.Tasks, *curT)
			curT = nil
		}
	}

	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		switch {
		case section == "" && strings.HasPrefix(line, "name:"):
			ws.Name = textutil.UnquoteYAML(strings.TrimPrefix(line, "name:"))
		case strings.HasPrefix(line, "projects:"):
			flushP()
			flushT()
			section = "projects"
		case strings.HasPrefix(line, "tasks:"):
			flushP()
			flushT()
			section = "tasks"
		case strings.HasPrefix(line, "- id:"):
			id := textutil.UnquoteYAML(strings.TrimPrefix(line, "- id:"))
			if section == "tasks" {
				flushT()
				curT = &Task{ID: id}
			} else {
				flushP()
				curP = &Project{ID: id}
			}
		default:
			if section == "tasks" && curT != nil {
				assignTaskField(curT, line)
			} else if section == "projects" && curP != nil {
				assignProjectField(curP, line)
			}
		}
	}
	flushP()
	flushT()
	return ws
}

func assignProjectField(p *Project, line string) {
	switch {
	case strings.HasPrefix(line, "path:"):
		p.Path = textutil.UnquoteYAML(strings.TrimPrefix(line, "path:"))
	case strings.HasPrefix(line, "repo:"):
		p.Repo = textutil.UnquoteYAML(strings.TrimPrefix(line, "repo:"))
	case strings.HasPrefix(line, "etl:"):
		p.ETL = textutil.UnquoteYAML(strings.TrimPrefix(line, "etl:"))
	case strings.HasPrefix(line, "branch_role:"):
		p.BranchRole = textutil.UnquoteYAML(strings.TrimPrefix(line, "branch_role:"))
	}
}

func assignTaskField(t *Task, line string) {
	switch {
	case strings.HasPrefix(line, "jira:"):
		t.Jira = textutil.UnquoteYAML(strings.TrimPrefix(line, "jira:"))
	case strings.HasPrefix(line, "repo:"):
		t.Repo = textutil.UnquoteYAML(strings.TrimPrefix(line, "repo:"))
	case strings.HasPrefix(line, "etl:"):
		t.ETL = textutil.UnquoteYAML(strings.TrimPrefix(line, "etl:"))
	case strings.HasPrefix(line, "path:"):
		t.Path = textutil.UnquoteYAML(strings.TrimPrefix(line, "path:"))
	case strings.HasPrefix(line, "branch:"):
		t.Branch = textutil.UnquoteYAML(strings.TrimPrefix(line, "branch:"))
	}
}
