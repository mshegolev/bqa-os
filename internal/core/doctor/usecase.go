// Package doctor inspects the health of a .bqa workspace.
package doctor

import (
	"os"
	"path/filepath"
)

// Check is the result of a single workspace health check.
type Check struct {
	Name   string
	OK     bool
	Detail string
}

// Report aggregates all workspace checks.
type Report struct {
	Checks []Check
	OK     bool
}

// requiredDirs mirrors the workspace layout created by `bqa init`.
var requiredDirs = []string{
	"input/sessions/raw",
	"input/sessions/normalized",
	"output",
	"registry",
	"memory",
	"agents",
	"skills",
	"workflows",
	"rules",
	"guardrails",
	"prompts",
}

// Run checks that every required workspace directory exists under baseDir.
func Run(baseDir string) Report {
	if baseDir == "" {
		baseDir = ".bqa"
	}
	report := Report{OK: true}
	for _, dir := range requiredDirs {
		path := filepath.Join(baseDir, filepath.FromSlash(dir))
		info, err := os.Stat(path)
		ok := err == nil && info.IsDir()
		detail := path
		if !ok {
			detail = path + " (missing)"
			report.OK = false
		}
		report.Checks = append(report.Checks, Check{Name: dir, OK: ok, Detail: detail})
	}
	return report
}
