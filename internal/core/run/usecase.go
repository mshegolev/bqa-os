// Package run builds a BQA Master Agent task plan from a local .bqa workspace.
package run

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Plan is the set of workspace artifacts selected for a task.
type Plan struct {
	Task      string
	Agents    []string
	Skills    []string
	Workflows []string
}

// Build loads the available agents, skills, and workflows from the workspace
// under baseDir and assembles a plan for the given task.
func Build(baseDir, task string) Plan {
	if baseDir == "" {
		baseDir = ".bqa"
	}
	return Plan{
		Task:      task,
		Agents:    listArtifacts(filepath.Join(baseDir, "agents")),
		Skills:    listArtifacts(filepath.Join(baseDir, "skills")),
		Workflows: listArtifacts(filepath.Join(baseDir, "workflows")),
	}
}

// Empty reports whether the plan selected no artifacts at all.
func (p Plan) Empty() bool {
	return len(p.Agents)+len(p.Skills)+len(p.Workflows) == 0
}

// listArtifacts returns the base names (without extension) of files in dir,
// sorted. A missing directory yields an empty slice.
func listArtifacts(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		names = append(names, strings.TrimSuffix(name, filepath.Ext(name)))
	}
	sort.Strings(names)
	return names
}
