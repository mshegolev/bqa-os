package demoarchive

import (
	"context"
	"errors"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type UseCase struct {
	Writer ports.DemoArchiveWriter
}

type Result struct {
	OutputPath   string
	FilesCreated int
}

func (u UseCase) Run(ctx context.Context, outputPath string) (Result, error) {
	if u.Writer == nil {
		return Result{}, errors.New("demo archive writer is required")
	}

	files := demoArchiveFiles()
	if err := u.Writer.WriteDemoArchive(ctx, outputPath, files); err != nil {
		return Result{}, err
	}

	return Result{OutputPath: outputPath, FilesCreated: len(files)}, nil
}

func demoArchiveFiles() []ports.DemoArchiveFile {
	return []ports.DemoArchiveFile{
		{Path: "manifest.json", Content: manifestJSON()},
		{Path: "normalized_sessions/session_001_api_regression.md", Content: apiRegressionSession()},
		{Path: "normalized_sessions/session_002_graphql_contract.md", Content: graphqlContractSession()},
		{Path: "normalized_sessions/session_003_data_quality.md", Content: dataQualitySession()},
		{Path: "agents/api-regression-agent.md", Content: apiRegressionAgent()},
		{Path: "agents/graphql-contract-agent.md", Content: graphqlContractAgent()},
		{Path: "workflows/api-regression-workflow.md", Content: apiRegressionWorkflow()},
		{Path: "workflows/data-quality-triage-workflow.md", Content: dataQualityWorkflow()},
		{Path: "specs/api-regression-spec.md", Content: apiRegressionSpec()},
		{Path: "specs/graphql-contract-spec.md", Content: graphqlContractSpec()},
		{Path: "knowledge/api_patterns.yaml", Content: apiPatterns()},
		{Path: "knowledge/data_quality_patterns.yaml", Content: dataQualityPatterns()},
		{Path: "knowledge/project_profile.yaml", Content: projectProfile()},
		{Path: "README_NEXT_STEPS.md", Content: readmeNextSteps()},
		{Path: "recommendations.md", Content: recommendations()},
	}
}

func manifestJSON() string {
	return `{
  "name": "bqa-os-synthetic-demo-archive",
  "schema_version": 1,
  "synthetic": true,
  "generated_at": "2026-01-01T00:00:00Z",
  "contains": [
    "normalized QA sessions",
    "generated agents",
    "workflows",
    "specs",
    "knowledge artifacts",
    "README_NEXT_STEPS.md",
    "recommendations.md"
  ]
}
`
}

func apiRegressionSession() string {
	return `# Synthetic Normalized QA Session: API Regression

Synthetic source: demo fixture.

## Context

A fictional notes app exposes REST endpoints for notes, labels, and search.

## Observations

- POST /v1/notes returns HTTP 201 for a valid request payload.
- GET /v1/notes/{id} returns HTTP 404 for a missing synthetic note id.
- Regression risk is highest around pagination and status code handling.

## Reusable QA Signals

- Verify status codes before checking response payload fields.
- Keep one negative case for missing ids and one positive case for created notes.
`
}

func graphqlContractSession() string {
	return `# Synthetic Normalized QA Session: GraphQL Contract

Synthetic source: demo fixture.

## Context

A fictional project board exposes GraphQL queries for boards, columns, and cards.

## Observations

- The board query returns stable ids, names, and column counts.
- The createCard mutation requires title and column id.
- Regression risk is highest when schema fields are renamed without updating tests.

## Reusable QA Signals

- Run an introspection check before contract assertions.
- Keep one mutation test and one query shape test.
`
}

func dataQualitySession() string {
	return `# Synthetic Normalized QA Session: Data Quality

Synthetic source: demo fixture.

## Context

A fictional analytics pipeline summarizes daily product events.

## Observations

- Row count reconciliation compares synthetic source_events and daily_summary.
- Null checks cover event_date, account_id, and event_type.
- Duplicate checks use account_id, event_type, and event_date.

## Reusable QA Signals

- Validate row count trends before field-level checks.
- Capture schema drift and duplicate checks in the same report.
`
}

func apiRegressionAgent() string {
	return `# Synthetic API Regression Agent

Synthetic source: demo fixture.

## Role

Review REST API regression notes and turn them into repeatable checks.

## Responsibilities

- Identify endpoint, method, status code, and request payload coverage.
- Separate positive, negative, and boundary cases.
- Produce concise follow-up checks for a human QA owner.
`
}

func graphqlContractAgent() string {
	return `# Synthetic GraphQL Contract Agent

Synthetic source: demo fixture.

## Role

Review GraphQL QA notes and turn query or mutation behavior into contract checks.

## Responsibilities

- Identify schema fields, operation names, and expected response shape.
- Flag renamed fields and missing required inputs.
- Recommend one query check and one mutation check when possible.
`
}

func apiRegressionWorkflow() string {
	return `# Synthetic API Regression Workflow

Synthetic source: demo fixture.

1. Identify endpoint, method, request payload, and expected status code.
2. Run the positive case and capture response shape.
3. Run the negative case and capture error format.
4. Compare pagination, sorting, and filtering behavior against the previous synthetic baseline.
5. Write a short recommendation for the next regression run.
`
}

func dataQualityWorkflow() string {
	return `# Synthetic Data Quality Triage Workflow

Synthetic source: demo fixture.

1. Confirm the synthetic source and target datasets.
2. Compare source and target row counts.
3. Check required fields for null values.
4. Check business keys for duplicates.
5. Capture schema drift and recommend the smallest next validation step.
`
}

func apiRegressionSpec() string {
	return `# Synthetic API Regression Spec

Synthetic source: demo fixture.

## Objective

Verify that fictional notes API behavior is stable across common regression paths.

## Acceptance Criteria

- Created notes return HTTP 201.
- Missing note ids return HTTP 404.
- Pagination keeps stable ordering for synthetic notes.
`
}

func graphqlContractSpec() string {
	return `# Synthetic GraphQL Contract Spec

Synthetic source: demo fixture.

## Objective

Verify that fictional project board GraphQL operations keep stable contract shape.

## Acceptance Criteria

- Board query returns id, name, and columns.
- createCard mutation returns id, title, and column id.
- Removed or renamed fields are reported before release.
`
}

func apiPatterns() string {
	return `synthetic_api_patterns:
  - name: "status_code_first"
    domain: "api"
    evidence: "Synthetic REST checks verify HTTP status before response payload fields."
  - name: "negative_missing_id"
    domain: "api"
    evidence: "Synthetic missing id check verifies stable 404 behavior."
`
}

func dataQualityPatterns() string {
	return `synthetic_data_quality_patterns:
  - name: "row_count_reconciliation"
    domain: "data_quality"
    evidence: "Synthetic source and target row counts are compared before field checks."
  - name: "duplicate_business_key"
    domain: "data_quality"
    evidence: "Synthetic duplicate checks use account id, event type, and event date."
`
}

func projectProfile() string {
	return `synthetic_project_profile:
  sessions_analyzed: 3
  signals:
    api: 1
    graphql: 1
    data_quality: 1
  maturity: "demo"
`
}

func readmeNextSteps() string {
	return `# Synthetic Demo Archive: Next Steps

Synthetic source: demo fixture.

1. Upload this archive into the static BQA-OS demo page.
2. Confirm that normalized sessions, agents, workflows, specs, and knowledge artifacts render locally.
3. Use recommendations.md to explain what a real sanitized pilot package would contain.
4. Replace this fixture only with synthetic or explicitly sanitized inputs.
`
}

func recommendations() string {
	return `# Synthetic Recommendations

Synthetic source: demo fixture.

## Recommended Pilot Shape

- Start with 10 to 30 sanitized QA artifacts.
- Include at least one regression note, one bug report, one checklist, and one prompt.
- Keep generated workflows human-in-the-loop and inspectable.

## Demo Follow-Up

- Show the API regression workflow first because it is easy to scan.
- Show the data quality workflow next because it demonstrates reusable QA memory.
- Use the synthetic project profile as the summary artifact.
`
}
