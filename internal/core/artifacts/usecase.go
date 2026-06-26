package artifacts

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type UseCase struct {
	Writer ports.BQAArtifactWriter
}

type Result struct {
	ArtifactsCreated int
}

func (u UseCase) Run(ctx context.Context) (Result, error) {
	artifacts := map[string]string{
		"skills/etl-log-investigation.md":          etlSkill(),
		"skills/runtime-trace-review.md":          runtimeSkill(),
		"agents/etl-qa-agent.md":                  etlAgent(),
		"agents/runtime-agent.md":                 runtimeAgent(),
		"workflows/etl-verification-workflow.md":  etlWorkflow(),
		"workflows/session-knowledge-workflow.md": sessionWorkflow(),
	}

	created := 0
	for path, content := range artifacts {
		if err := u.Writer.WriteBQAArtifact(ctx, path, content); err != nil {
			return Result{}, err
		}
		created++
	}
	return Result{ArtifactsCreated: created}, nil
}

func etlSkill() string {
	return "# ETL Log Investigation\n\n## Purpose\n\nInvestigate Big Data and ETL testing failures using logs, row counts, schemas, partitions, and reproducible commands.\n\n## Procedure\n\n1. Identify environment, dataset, job, partition, and failing step.\n2. Collect scheduler, Spark, Hive, and application logs.\n3. Compare source and target row counts, schema expectations, nulls, duplicates, and reprocessing windows.\n4. Capture exact evidence paths, commands, expected results, and actual results.\n"
}

func runtimeSkill() string {
	return "# Runtime Trace Review\n\n## Purpose\n\nReview AI coding runtime sessions and extract reusable engineering memory.\n\n## Procedure\n\n1. Identify runtime source and workspace context.\n2. Map user intent, tool calls, shell commands, approvals, and failures.\n3. Extract reusable prompts, recovery steps, constraints, and guardrails.\n4. Keep raw private content out of public repositories.\n"
}

func etlAgent() string {
	return "# ETL QA Agent\n\n## Role\n\nSpecialist agent for Big Data, ETL, Airflow, Spark, Hive, reconciliation, and data quality validation tasks.\n\n## Responsibilities\n\n- Build concise investigation plans.\n- Prefer commands, logs, row counts, schema evidence, and reproducible checks.\n- Report risk, evidence, and next actions clearly.\n"
}

func runtimeAgent() string {
	return "# Runtime Agent\n\n## Role\n\nSpecialist agent for AI coding session review and runtime trace analysis.\n\n## Responsibilities\n\n- Review normalized sessions and runtime traces.\n- Extract reusable prompts, failures, fixes, and workflow patterns.\n- Propose updates to skills, workflows, rules, and guardrails.\n"
}

func etlWorkflow() string {
	return "# ETL Verification Workflow\n\n1. Identify ticket, dataset, environment, job, and partition.\n2. Collect relevant logs and execution metadata.\n3. Validate source availability, target output, row counts, schema, and data quality checks.\n4. Isolate the smallest failing condition.\n5. Produce a QA report with evidence and next actions.\n"
}

func sessionWorkflow() string {
	return "# Session Knowledge Workflow\n\n1. Run `bqa discover`.\n2. Run `bqa ingest`.\n3. Run `bqa build`.\n4. Review generated knowledge, skills, agents, and workflows.\n5. Sanitize before syncing reusable knowledge to Brain.\n"
}
