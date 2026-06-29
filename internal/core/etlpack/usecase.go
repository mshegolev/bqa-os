package etlpack

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

const DefaultOutputDir = "output/etl-agent-pack"

type UseCase struct {
	Sessions  ports.NormalizedSessionReader
	Knowledge ports.KnowledgeArtifactReader
	Writer    ports.BQAArtifactWriter
	OutputDir string
}

type Result struct {
	ArtifactsCreated        int
	SessionsProcessed       int
	KnowledgeArtifactsFound int
	UsedSyntheticDemo       bool
	OutputDir               string
}

type statistics struct {
	SessionsAnalyzed        int
	KnowledgeArtifacts      int
	ETLPatternCount         int
	DataQualityPatternCount int
	UsedSyntheticDemo       bool
}

type artifact struct {
	Path    string
	Content string
}

func (u UseCase) Run(ctx context.Context) (Result, error) {
	if u.Writer == nil {
		return Result{}, errors.New("etl pack writer is required")
	}

	stats := u.collectStatistics(ctx)
	outputDir := cleanOutputDir(u.OutputDir)
	artifacts := renderArtifacts(outputDir, stats)

	created := 0
	for _, item := range artifacts {
		if err := u.Writer.WriteBQAArtifact(ctx, item.Path, item.Content); err != nil {
			return Result{}, err
		}
		created++
	}

	return Result{
		ArtifactsCreated:        created,
		SessionsProcessed:       stats.SessionsAnalyzed,
		KnowledgeArtifactsFound: stats.KnowledgeArtifacts,
		UsedSyntheticDemo:       stats.UsedSyntheticDemo,
		OutputDir:               outputDir,
	}, nil
}

func (u UseCase) collectStatistics(ctx context.Context) statistics {
	stats := statistics{}

	if u.Sessions != nil {
		if index, err := u.Sessions.LoadSessionIndex(ctx); err == nil {
			stats.SessionsAnalyzed = len(index.Entries)
		}
	}

	if u.Knowledge != nil {
		for _, filename := range []string{"etl_patterns.yaml", "data_quality_patterns.yaml", "project_profile.yaml"} {
			content, err := u.Knowledge.ReadKnowledgeArtifact(ctx, filename)
			if err != nil {
				continue
			}
			stats.KnowledgeArtifacts++
			switch filename {
			case "etl_patterns.yaml":
				stats.ETLPatternCount = countYAMLListItems(content)
			case "data_quality_patterns.yaml":
				stats.DataQualityPatternCount = countYAMLListItems(content)
			}
		}
	}

	stats.UsedSyntheticDemo = stats.SessionsAnalyzed == 0 && stats.KnowledgeArtifacts == 0
	return stats
}

func cleanOutputDir(outputDir string) string {
	if strings.TrimSpace(outputDir) == "" {
		return DefaultOutputDir
	}
	return filepath.ToSlash(filepath.Clean(outputDir))
}

func countYAMLListItems(content string) int {
	count := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "- name:") {
			count++
		}
	}
	return count
}

func renderArtifacts(outputDir string, stats statistics) []artifact {
	files := map[string]string{
		"statistics/summary.md":                              statisticsSummary(stats),
		"agents/codex-etl-qa-agent.md":                       codexAgent(),
		"agents/claude-code-etl-qa-agent.md":                 claudeCodeAgent(),
		"prompts/codex-etl-qa-agent-prompt.md":               codexPrompt(),
		"prompts/claude-code-etl-qa-agent-prompt.md":         claudeCodePrompt(),
		"workflows/etl-regression-workflow.md":               etlRegressionWorkflow(),
		"workflows/data-reconciliation-workflow.md":          dataReconciliationWorkflow(),
		"workflows/data-quality-validation-workflow.md":      dataQualityValidationWorkflow(),
		"specs/etl-test-spec-template.md":                    etlTestSpecTemplate(),
		"specs/source-to-target-mapping-review-checklist.md": sourceToTargetChecklist(),
		"examples/synthetic-etl-case.md":                     syntheticETLCase(),
		"README_NEXT_STEPS.md":                               readmeNextSteps(),
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	artifacts := make([]artifact, 0, len(paths))
	for _, path := range paths {
		artifacts = append(artifacts, artifact{
			Path:    filepath.ToSlash(filepath.Join(outputDir, path)),
			Content: files[path],
		})
	}
	return artifacts
}

func statisticsSummary(stats statistics) string {
	inputBasis := "processed local indexes and generated knowledge counts"
	if stats.UsedSyntheticDemo {
		inputBasis = "synthetic demo data"
	}

	return fmt.Sprintf(`# ETL Agent Pack Statistics

Input basis: %s
Sessions analyzed: %d
Knowledge artifacts available: %d
ETL patterns counted: %d
Data quality patterns counted: %d
Synthetic examples included: yes
Privacy mode: counts only; no normalized session body, knowledge evidence, private logs, customer data, or secrets are copied.
`, inputBasis, stats.SessionsAnalyzed, stats.KnowledgeArtifacts, stats.ETLPatternCount, stats.DataQualityPatternCount)
}

func codexAgent() string {
	return `# Codex ETL QA Agent

## Role

Act as a local-first ETL QA specialist for pipeline verification, regression analysis, reconciliation, and data quality checks.

## Responsibilities

- Turn a ticket, release note, or short QA request into a concrete ETL verification plan.
- Prefer repo-local scripts, generated BQA knowledge, scheduler metadata, SQL checks, and reproducible commands.
- Separate evidence, assumptions, risks, blockers, and next actions.
- Use only sanitized user-provided inputs or the synthetic examples in this pack.

## Response Contract

1. Scope and assumptions.
2. Verification plan.
3. Commands or queries to run.
4. Expected evidence.
5. Risk and blocker summary.
`
}

func claudeCodeAgent() string {
	return `# Claude Code ETL QA Agent

## Role

Act as a Claude Code ETL QA agent that plans and verifies data pipeline behavior without exposing private data.

## Responsibilities

- Inspect project-local files before proposing checks.
- Keep implementation changes small and reversible.
- Use synthetic examples unless the user provides sanitized local artifacts.
- Keep raw logs, credentials, customer identifiers, and private business context out of generated outputs.

## Output Contract

- State the ETL target, environment, source tables, target tables, and partition scope.
- List verification steps with commands or query templates.
- Report pass, fail, blocked, or needs-human-review with evidence references.
`
}

func codexPrompt() string {
	return `# Codex ETL QA Agent Prompt

Read this prompt and act as the ETL QA Agent for the current repository.

Rules:
- Work local-first. Inspect repository files and existing BQA artifacts before suggesting new tools.
- Do not paste private logs, secrets, customer data, or raw session content into your answer.
- Use synthetic examples from this pack when real sanitized evidence is unavailable.
- Keep commands copy-pasteable and separate read-only checks from destructive actions.

Task format:

ETL task:
Environment:
Pipeline or job:
Source tables:
Target tables:
Partition or time window:
Known change:

Return:
1. QA objective.
2. Data availability checks.
3. Reconciliation checks.
4. Data quality checks.
5. Regression risks.
6. Evidence to attach to the ticket.
`
}

func claudeCodePrompt() string {
	return `# Claude Code ETL QA Agent Prompt

You are an ETL QA Agent operating inside Claude Code.

Follow this process:
1. Read project-local instructions first.
2. Identify the ETL pipeline, environment, source, target, and time window.
3. Build a short verification plan before editing or running broad commands.
4. Prefer existing repo scripts and tests.
5. Use synthetic fixtures unless the user explicitly provides sanitized inputs.
6. Never include secrets, customer identifiers, private logs, or raw session transcripts in generated files.

Final response:
- Files inspected.
- Checks run.
- Results.
- Risks.
- Next steps.
`
}

func etlRegressionWorkflow() string {
	return `# ETL Regression Workflow

## Goal

Verify that a pipeline change does not break existing extraction, transformation, load, or downstream reporting behavior.

## Steps

1. Identify the change, ticket, pipeline, environment, and affected tables.
2. Review source-to-target mapping and expected transformation rules.
3. Select a synthetic or sanitized regression window.
4. Check source availability and target completeness.
5. Compare row counts, key coverage, null rates, duplicate rates, and schema compatibility.
6. Review scheduler and job logs for retries, skipped partitions, and partial loads.
7. Record evidence links, command output summaries, and unresolved risks.

## Exit Criteria

- No unexplained count deltas.
- No unexpected schema drift.
- No new critical data quality violations.
- Known limitations are documented.
`
}

func dataReconciliationWorkflow() string {
	return `# Data Reconciliation Workflow

## Goal

Compare source and target data for a defined table, key set, and processing window.

## Steps

1. Define source query, target query, primary keys, and partition filters.
2. Count source rows and target rows.
3. Compare aggregate totals for stable numeric measures.
4. Check missing keys in both directions.
5. Sample transformed fields and verify expected business rules.
6. Save query templates and summarized results.

## Query Template

~~~sql
-- Synthetic template only. Replace names with sanitized project-local tables.
select count(*) as row_count
from source_schema.source_table
where partition_date = '2026-01-01';
~~~
`
}

func dataQualityValidationWorkflow() string {
	return `# Data Quality Validation Workflow

## Goal

Detect data defects that reconciliation alone may miss.

## Checks

- Required fields are not null.
- Business keys are unique in the expected grain.
- Enumerated values stay within allowed sets.
- Date and timestamp fields stay inside the processing window.
- Numeric measures stay inside expected ranges.
- Schema changes are intentional and documented.

## Report

Use pass, fail, blocked, or needs-human-review for each check. Include sanitized query names and summarized counts only.
`
}

func etlTestSpecTemplate() string {
	return `# ETL Test Spec Template

## Ticket

- ID:
- Goal:
- Environment:
- Pipeline or job:

## Scope

- Source system:
- Source tables:
- Target tables:
- Partition or time window:
- Out of scope:

## Test Cases

| ID | Check | Input | Expected result | Evidence |
|----|-------|-------|-----------------|----------|
| ETL-001 | Source availability | Synthetic partition | Source rows exist | Count summary |
| ETL-002 | Target completeness | Synthetic partition | Target rows loaded | Count summary |
| ETL-003 | Reconciliation | Synthetic key set | No unexplained deltas | Diff summary |
| ETL-004 | Data quality | Synthetic target data | No critical DQ failures | DQ summary |

## Risks

- Data delay:
- Schema drift:
- Backfill:
- External dependency:
`
}

func sourceToTargetChecklist() string {
	return `# Source-to-Target Mapping Review Checklist

## Mapping

- Source table and target table are identified.
- Primary key or business key is documented.
- Partitioning and incremental load fields are documented.
- Transformation rules are explicit.
- Default values and null handling are explicit.
- Type conversions are safe.

## Validation

- Every required target field has a source or derivation rule.
- Dropped fields are intentional.
- Join keys are stable and unique at the expected grain.
- Aggregations define grouping keys.
- Time zone handling is documented.
- Late-arriving data behavior is documented.

## Evidence

- Use sanitized table names in shared reports.
- Store raw logs only in approved local or private locations.
- Attach summarized counts and diffs, not private data samples.
`
}

func syntheticETLCase() string {
	return `# Synthetic ETL QA Example

## Scenario

A synthetic orders pipeline loads daily order events from ` + "`source_orders`" + ` into ` + "`mart_orders_daily`" + `.

## Synthetic Inputs

- Source table: ` + "`source_orders`" + `
- Target table: ` + "`mart_orders_daily`" + `
- Partition: ` + "`2026-01-01`" + `
- Expected row count: ` + "`1000`" + `

## Example Checks

~~~sql
select count(*) from source_orders where partition_date = '2026-01-01';
select count(*) from mart_orders_daily where partition_date = '2026-01-01';
~~~

## Expected Result

Counts match after documented filters. Required fields are populated. Duplicate order IDs are zero.
`
}

func readmeNextSteps() string {
	return `# ETL QA Agent Pack Next Steps

## What Was Generated

- Codex and Claude Code ETL QA agent prompts.
- ETL regression, data reconciliation, and data quality workflows.
- ETL test spec template.
- Source-to-target mapping review checklist.
- Statistics summary based on local counts or synthetic fallback.
- Synthetic example data for copy-paste demos.

## How To Use

1. Open ` + "`prompts/codex-etl-qa-agent-prompt.md`" + ` or ` + "`prompts/claude-code-etl-qa-agent-prompt.md`" + `.
2. Paste the prompt into the matching runtime.
3. Fill in the ETL task fields with sanitized project-local context.
4. Use the workflows and spec template to plan verification.
5. Store private logs and raw evidence outside this public repo.

## Safety Rules

- Use synthetic examples by default.
- Use sanitized local artifacts only when the user explicitly provides them.
- Do not paste secrets, customer data, private logs, or raw session transcripts into generated files.
`
}
