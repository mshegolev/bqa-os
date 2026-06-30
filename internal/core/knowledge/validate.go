package knowledge

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// ValidationIssue describes a single problem found with a knowledge artifact.
type ValidationIssue struct {
	Filename string
	Detail   string
}

// ValidationReport is the result of validating build outputs.
type ValidationReport struct {
	Checked  int
	Expected int
	Issues   []ValidationIssue
}

// OK reports whether validation passed with no issues.
func (r ValidationReport) OK() bool {
	return len(r.Issues) == 0
}

// Validate inspects the knowledge artifacts produced by `bqa build` using the
// shared ExpectedArtifacts contract. It checks, for each expected file, that it
// exists, is non-empty, and contains its root key. project_profile.yaml must
// additionally contain a session count (sessions_analyzed). The total artifact
// count is compared against the contract. It never reaches external services.
func Validate(ctx context.Context, reader ports.KnowledgeArtifactReader) ValidationReport {
	expected := ExpectedArtifacts()
	report := ValidationReport{Expected: len(expected)}

	for _, spec := range expected {
		content, err := reader.ReadKnowledgeArtifact(ctx, spec.Filename)
		if err != nil {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "missing or unreadable: " + err.Error()})
			continue
		}
		report.Checked++

		if strings.TrimSpace(content) == "" {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "file is empty"})
			continue
		}
		if !strings.Contains(content, spec.RootKey+":") {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: fmt.Sprintf("missing root key %q", spec.RootKey)})
			continue
		}
		if spec.RootKey == "project_profile" && !strings.Contains(content, "sessions_analyzed:") {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "missing session count (sessions_analyzed)"})
			continue
		}
	}

	if report.Checked != report.Expected {
		report.Issues = append(report.Issues, ValidationIssue{
			Filename: "",
			Detail:   fmt.Sprintf("artifact count mismatch: found %d of %d expected", report.Checked, report.Expected),
		})
	}

	sort.SliceStable(report.Issues, func(i, j int) bool {
		return report.Issues[i].Filename < report.Issues[j].Filename
	})
	return report
}
