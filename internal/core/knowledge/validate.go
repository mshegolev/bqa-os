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

// ValidationReport is the result of validating build outputs. Valid counts the
// artifacts that passed every check; Expected is the contract size.
type ValidationReport struct {
	Valid    int
	Expected int
	Issues   []ValidationIssue
}

// OK reports whether validation passed with no issues.
func (r ValidationReport) OK() bool {
	return len(r.Issues) == 0
}

// Validate inspects the knowledge artifacts produced by `bqa build` using the
// shared ExpectedArtifacts contract. For each expected file it checks that the
// file exists, is non-empty, and carries the v1 envelope: the supported
// `schema_version` and a matching `kind`. Pattern files must contain a
// `patterns:` list; project_profile.yaml must contain a `profile:` block with
// `sessions_analyzed`. The total artifact count is compared against the
// contract. It never reaches external services.
func Validate(ctx context.Context, reader ports.KnowledgeArtifactReader) ValidationReport {
	expected := ExpectedArtifacts()
	report := ValidationReport{Expected: len(expected)}

	for _, spec := range expected {
		content, err := reader.ReadKnowledgeArtifact(ctx, spec.Filename)
		if err != nil {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "missing or unreadable: " + err.Error()})
			continue
		}
		if strings.TrimSpace(content) == "" {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "file is empty"})
			continue
		}
		if !strings.Contains(content, fmt.Sprintf("schema_version: %d", SchemaVersion)) {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: fmt.Sprintf("missing or unsupported schema_version (want %d)", SchemaVersion)})
			continue
		}
		if !strings.Contains(content, "kind: "+spec.RootKey) {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: fmt.Sprintf("missing or wrong kind (want %q)", spec.RootKey)})
			continue
		}
		if spec.RootKey == "project_profile" {
			if !strings.Contains(content, "profile:") || !strings.Contains(content, "sessions_analyzed:") {
				report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "missing profile block or sessions_analyzed"})
				continue
			}
		} else if !strings.Contains(content, "patterns:") {
			report.Issues = append(report.Issues, ValidationIssue{Filename: spec.Filename, Detail: "missing patterns list"})
			continue
		}
		// Passed every check.
		report.Valid++
	}

	sort.SliceStable(report.Issues, func(i, j int) bool {
		return report.Issues[i].Filename < report.Issues[j].Filename
	})
	return report
}
