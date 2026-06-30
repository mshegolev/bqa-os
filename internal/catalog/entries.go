package catalog

// skills holds the built-in skill entries in canonical order.
var skills = []Entry{
	{
		ID:      "etl-log-investigation",
		Type:    "skill",
		Domain:  "etl",
		Title:   "ETL Log Investigation",
		Content: "# ETL Log Investigation\n\n## Purpose\n\nInvestigate Big Data and ETL testing failures using logs, row counts, schemas, partitions, and reproducible commands.\n\n## Procedure\n\n1. Identify environment, dataset, job, partition, and failing step.\n2. Collect scheduler, Spark, Hive, and application logs.\n3. Compare source and target row counts, schema expectations, nulls, duplicates, and reprocessing windows.\n4. Capture exact evidence paths, commands, expected results, and actual results.\n",
	},
	{
		ID:      "runtime-trace-review",
		Type:    "skill",
		Domain:  "runtime",
		Title:   "Runtime Trace Review",
		Content: "# Runtime Trace Review\n\n## Purpose\n\nReview AI coding runtime sessions and extract reusable engineering memory.\n\n## Procedure\n\n1. Identify runtime source and workspace context.\n2. Map user intent, tool calls, shell commands, approvals, and failures.\n3. Extract reusable prompts, recovery steps, constraints, and guardrails.\n4. Keep raw private content out of public repositories.\n",
	},
}

// workflows holds the built-in workflow entries in canonical order.
var workflows = []Entry{
	{
		ID:      "etl-verification-workflow",
		Type:    "workflow",
		Domain:  "etl",
		Title:   "ETL Verification Workflow",
		Content: "# ETL Verification Workflow\n\n1. Identify ticket, dataset, environment, job, and partition.\n2. Collect relevant logs and execution metadata.\n3. Validate source availability, target output, row counts, schema, and data quality checks.\n4. Isolate the smallest failing condition.\n5. Produce a QA report with evidence and next actions.\n",
	},
	{
		ID:      "session-knowledge-workflow",
		Type:    "workflow",
		Domain:  "memory",
		Title:   "Session Knowledge Workflow",
		Content: "# Session Knowledge Workflow\n\n1. Run `bqa discover`.\n2. Run `bqa ingest`.\n3. Run `bqa build`.\n4. Review generated knowledge, skills, agents, and workflows.\n5. Sanitize before syncing reusable knowledge to Brain.\n",
	},
	{
		ID:     "etl-regression-workflow",
		Type:   "workflow",
		Domain: "etl",
		Title:  "ETL Regression Workflow",
		Content: `# ETL Regression Workflow

1. Identify ticket, ETL name, environment, source tables, target tables, partition window, and acceptance criteria.
2. Check the repository diff, config changes, scheduler changes, and migration notes.
3. Run the smallest available automated regression checks.
4. Compare source and target row counts for the target partition window.
5. Review failed rows, schema drift, nullability, duplicate keys, and checksum deltas.
6. Record evidence, result, risk, and follow-up owner.
`,
	},
	{
		ID:     "data-reconciliation-workflow",
		Type:   "workflow",
		Domain: "etl",
		Title:  "Data Reconciliation Workflow",
		Content: `# Data Reconciliation Workflow

1. Define source query, target query, join keys, partition filters, and tolerated deltas.
2. Capture source row count, target row count, distinct key count, and duplicate key count.
3. Compare numeric aggregates and checksums for high-value fields.
4. Sample mismatched rows using synthetic-safe examples or sanitized values only.
5. Document exact commands, counts, mismatch class, and whether reprocessing is required.
`,
	},
	{
		ID:     "data-quality-validation-workflow",
		Type:   "workflow",
		Domain: "etl",
		Title:  "Data Quality Validation Workflow",
		Content: `# Data Quality Validation Workflow

1. List required not-null fields, unique keys, reference fields, and accepted value ranges.
2. Check schema compatibility between source, transformation, and target.
3. Run null, duplicate, type, range, freshness, and referential integrity checks.
4. Separate source-quality defects from ETL transformation defects.
5. Report failed rules with counts, severity, reproducible query, and owner.
`,
	},
}

// runtimeAgentContent is the runtime agent persona, kept verbatim (not de-duplicated).
const runtimeAgentContent = "# Runtime Agent\n\n## Role\n\nSpecialist agent for AI coding session review and runtime trace analysis.\n\n## Responsibilities\n\n- Review normalized sessions and runtime traces.\n- Extract reusable prompts, failures, fixes, and workflow patterns.\n- Propose updates to skills, workflows, rules, and guardrails.\n"

func find(entries []Entry, id string) Entry {
	for _, e := range entries {
		if e.ID == id {
			return e
		}
	}
	return Entry{}
}

// Skill returns the built-in skill entry with the given id.
func Skill(id string) Entry { return find(skills, id) }

// Workflow returns the built-in workflow entry with the given id.
func Workflow(id string) Entry { return find(workflows, id) }

// RuntimeAgentContent returns the verbatim runtime agent persona content.
func RuntimeAgentContent() string { return runtimeAgentContent }
