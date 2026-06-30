package etlpack

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/mshegolev/bqa-os/internal/catalog"
	"github.com/mshegolev/bqa-os/internal/ports"
	"github.com/mshegolev/bqa-os/internal/textutil"
)

type UseCase struct {
	Reader    ports.ETLAgentPackInputReader
	Writer    ports.ETLAgentPackWriter
	OutputDir string
}

type Result struct {
	ArtifactsCreated        int
	SessionsProcessed       int
	KnowledgeArtifactsFound int
	SyntheticExamplesUsed   bool
	OutputDir               string
}

type packStats struct {
	sessionsProcessed       int
	knowledgeArtifactsFound int
	etlSignals              int
	reconciliationSignals   int
	dataQualitySignals      int
	regressionSignals       int
	syntheticExamplesUsed   bool
}

func (u UseCase) Run(ctx context.Context) (Result, error) {
	stats := u.collectStats(ctx)
	stats.syntheticExamplesUsed = stats.sessionsProcessed == 0 && stats.knowledgeArtifactsFound == 0

	artifacts := map[string]string{
		"statistics/summary.md":                              summary(stats),
		"agents/codex-etl-qa-agent.md":                       catalog.RenderAgent(catalog.ETLQA(), catalog.CodexFlavor()),
		"agents/claude-code-etl-qa-agent.md":                 catalog.RenderAgent(catalog.ETLQA(), catalog.ClaudeFlavor()),
		"workflows/etl-regression-workflow.md":               catalog.Workflow("etl-regression-workflow").Content,
		"workflows/data-reconciliation-workflow.md":          catalog.Workflow("data-reconciliation-workflow").Content,
		"workflows/data-quality-validation-workflow.md":      catalog.Workflow("data-quality-validation-workflow").Content,
		"specs/etl-test-spec-template.md":                    etlTestSpecTemplate(),
		"specs/source-to-target-mapping-review-checklist.md": sourceToTargetChecklist(),
		"prompts/codex-etl-qa-agent-prompt.md":               codexPrompt(),
		"prompts/claude-code-etl-qa-agent-prompt.md":         claudeCodePrompt(),
		"examples/synthetic-etl-reconciliation-example.md":   syntheticExample(),
		"README_NEXT_STEPS.md":                               readmeNextSteps(),
	}

	created := 0
	for relativePath, content := range artifacts {
		if err := u.Writer.WriteETLAgentPackArtifact(ctx, path.Clean(relativePath), content); err != nil {
			return Result{}, err
		}
		created++
	}

	return Result{
		ArtifactsCreated:        created,
		SessionsProcessed:       stats.sessionsProcessed,
		KnowledgeArtifactsFound: stats.knowledgeArtifactsFound,
		SyntheticExamplesUsed:   stats.syntheticExamplesUsed,
		OutputDir:               u.outputDir(),
	}, nil
}

func (u UseCase) collectStats(ctx context.Context) packStats {
	var stats packStats
	if u.Reader == nil {
		return stats
	}

	for _, filename := range []string{
		"etl_patterns.yaml",
		"data_quality_patterns.yaml",
		"common_bugs.yaml",
		"project_profile.yaml",
	} {
		content, err := u.Reader.ReadKnowledgeArtifact(ctx, filename)
		if err != nil {
			continue
		}
		stats.knowledgeArtifactsFound++
		stats.etlSignals += countSignals(content, "etl")
		stats.reconciliationSignals += countSignals(content, "reconciliation", "row count", "source", "target")
		stats.dataQualitySignals += countSignals(content, "data_quality", "data quality", "null", "duplicate", "schema")
		stats.regressionSignals += countSignals(content, "regression", "failure", "failed", "bug")
	}

	index, err := u.Reader.LoadSessionIndex(ctx)
	if err != nil {
		return stats
	}
	for _, entry := range index.Entries {
		content, err := u.Reader.ReadNormalizedSession(ctx, entry.NormalizedPath)
		if err != nil {
			continue
		}
		lower := strings.ToLower(content)
		if !hasETLSignal(lower) {
			continue
		}
		stats.sessionsProcessed++
		stats.etlSignals++
		if textutil.HasAny(lower, "reconciliation", "row count", "source table", "target table", "checksum") {
			stats.reconciliationSignals++
		}
		if textutil.HasAny(lower, "data quality", "null check", "duplicate", "schema drift", "not null") {
			stats.dataQualitySignals++
		}
		if textutil.HasAny(lower, "regression", "failed", "failure", "bug", "mismatch") {
			stats.regressionSignals++
		}
	}
	return stats
}

func (u UseCase) outputDir() string {
	if u.OutputDir == "" {
		return ".bqa/output/etl-agent-pack"
	}
	return u.OutputDir
}

func countSignals(content string, needles ...string) int {
	lower := strings.ToLower(content)
	count := 0
	for _, needle := range needles {
		if strings.Contains(lower, needle) {
			count++
		}
	}
	return count
}

func hasETLSignal(text string) bool {
	return textutil.HasAny(text, "etl", "spark", "hive", "oozie", "airflow", "pipeline", "reconciliation", "row count")
}

func summary(stats packStats) string {
	sourceMode := "Processed local inputs"
	if stats.syntheticExamplesUsed {
		sourceMode = "Synthetic demo data"
	}
	return fmt.Sprintf(`# ETL Agent Pack Statistics

Source mode: %s
Sessions processed: %d
Knowledge artifacts found: %d

## Aggregate Signals

- ETL signals: %d
- Reconciliation signals: %d
- Data quality signals: %d
- Regression signals: %d

## Data Safety

- No raw normalized session content is copied into this pack.
- Examples in this pack are synthetic and safe for public demos.
`, sourceMode, stats.sessionsProcessed, stats.knowledgeArtifactsFound, stats.etlSignals, stats.reconciliationSignals, stats.dataQualitySignals, stats.regressionSignals)
}

func etlTestSpecTemplate() string {
	return `# ETL Test Spec Template

## Scope

- Ticket:
- ETL pipeline:
- Environment:
- Source tables:
- Target tables:
- Partition or time window:

## Preconditions

- Required configs:
- Required credentials location:
- Scheduler or job state:

## Test Cases

| ID | Check | Query or command | Expected result | Evidence |
|----|-------|------------------|-----------------|----------|
| ETL-001 | Source availability | synthetic query placeholder | source rows exist | |
| ETL-002 | Target availability | synthetic query placeholder | target rows exist | |
| ETL-003 | Reconciliation | synthetic query placeholder | counts match or accepted delta | |
| ETL-004 | Data quality | synthetic query placeholder | no blocking nulls or duplicates | |

## QA Result

- Status:
- Evidence:
- Risks:
- Follow-up:
`
}

func sourceToTargetChecklist() string {
	return `# Source-to-Target Mapping Review Checklist

- Source table and target table names are explicit.
- Join keys and business keys are documented.
- Type conversions are intentional and testable.
- Null handling is specified for every required field.
- Default values are documented.
- Filters and partition logic match the acceptance criteria.
- Deduplication rules are documented.
- Late-arriving data behavior is documented.
- Row count and checksum reconciliation are defined.
- Rollback or reprocessing steps are known.
`
}

func codexPrompt() string {
	return `# Codex ETL QA Agent Prompt

copy-paste into Codex:

Act as the Codex ETL QA Agent from this pack.

Task:
Validate the ETL change using local repository context, generated BQA knowledge if present, and synthetic examples only when real sanitized inputs are unavailable.

Rules:
- Do not expose secrets, private logs, customer data, or raw session content.
- Follow the repository test commands and existing helper scripts.
- Produce a concise QA report with checks, evidence, result, risks, and next actions.
`
}

func claudeCodePrompt() string {
	return `# Claude Code ETL QA Agent Prompt

copy-paste into Claude Code:

Act as the Claude Code ETL QA Agent from this pack.

Task:
Build an ETL QA validation plan, run safe checks available in the repository, and summarize aggregate evidence without copying raw private data.

Rules:
- Use project-local conventions and tools.
- Keep examples synthetic unless the user provides sanitized data.
- Ask a blocker question before destructive commands or unclear architecture choices.
- Return files changed, tests run, evidence, and risks.
`
}

func syntheticExample() string {
	return `# Synthetic ETL Reconciliation Example

This example is synthetic and safe for public demos.

## Scenario

An ETL job loads daily invoice events from synthetic_source.invoice_events into synthetic_target.invoice_daily.

## Checks

| Check | Synthetic result |
|-------|------------------|
| Source row count | 1,000 |
| Target row count | 1,000 |
| Distinct invoice IDs | 1,000 |
| Null invoice amount | 0 |
| Duplicate invoice IDs | 0 |

## QA Summary

The synthetic load passes row count, key uniqueness, and required-field checks. Use this format when no sanitized session data is available.
`
}

func readmeNextSteps() string {
	return `# ETL QA Agent Pack

## Next steps

1. Open prompts/codex-etl-qa-agent-prompt.md or prompts/claude-code-etl-qa-agent-prompt.md.
2. Copy the prompt into the selected coding runtime.
3. Attach a ticket, sanitized ETL notes, or local generated .bqa/knowledge context.
4. Use the workflows and spec template to run a human-reviewed ETL validation.
5. Keep private logs, secrets, customer data, and raw session content out of shared outputs.

## Contents

- statistics/summary.md
- agents/
- workflows/
- specs/
- prompts/
- examples/
`
}
