package artifacts

import (
	"context"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type UseCase struct {
	Writer              ports.BQAArtifactWriter
	IncludeSalesPackage bool
}

type Result struct {
	ArtifactsCreated int
}

func (u UseCase) Run(ctx context.Context) (Result, error) {
	artifacts := map[string]string{
		"skills/etl-log-investigation.md":         etlSkill(),
		"skills/runtime-trace-review.md":          runtimeSkill(),
		"agents/etl-qa-agent.md":                  etlAgent(),
		"agents/runtime-agent.md":                 runtimeAgent(),
		"workflows/etl-verification-workflow.md":  etlWorkflow(),
		"workflows/session-knowledge-workflow.md": sessionWorkflow(),
		"registry/index.yaml":                     registryIndex(),
		"registry/skills.yaml":                    registrySkills(),
		"registry/agents.yaml":                    registryAgents(),
		"registry/workflows.yaml":                 registryWorkflows(),
	}

	if u.IncludeSalesPackage {
		artifacts["sales/pilot-offer-one-pager.md"] = pilotOfferOnePager()
		artifacts["sales/internal-demo-script.md"] = internalDemoScript()
		artifacts["sales/discovery-call-script.md"] = discoveryCallScript()
		artifacts["sales/onboarding-checklist.md"] = onboardingChecklist()
		artifacts["sales/outreach-samples.md"] = outreachSamples()
		artifacts["sales/pricing-hypothesis.md"] = pricingHypothesis()
		artifacts["sales/internal-stakeholder-faq.md"] = internalStakeholderFAQ()
		artifacts["registry/index.yaml"] = registryIndexWithSales()
		artifacts["registry/sales.yaml"] = registrySales()
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

func pilotOfferOnePager() string {
	return "# Pilot Offer One-Pager\n\n## Offer\n\n2-week QA Memory Pilot for teams that lose QA knowledge across test notes, bug reports, regression checklists, prompts, and repeated verification steps.\n\n## Buyer\n\nQA Lead, QA Automation Lead, CTO, or VP Engineering in a B2B SaaS, API, GraphQL, ETL, or data pipeline team.\n\n## Pain\n\nAI-assisted development increases delivery speed, but regression review, repeated QA checks, and onboarding knowledge remain manual and scattered.\n\n## Scope\n\nClient provides 10-30 sanitized QA artifacts. BQA-OS returns a reusable QA knowledge base and 3-5 AI-assisted QA workflows.\n\n## Deliverables\n\n- Project-specific QA knowledge artifacts.\n- 3-5 reusable workflows for regression, API, GraphQL, ETL, or bug report standardization.\n- Internal readout with before/after examples and next-scope recommendation.\n\n## Success Criteria\n\n- The pilot produces at least three reusable workflows from synthetic or sanitized inputs.\n- A QA lead can reuse one workflow during an internal validation session.\n- No private session logs, secrets, customer data, or real production details are included.\n\n## Next Step\n\nBook a 30-minute kickoff and send the sanitized artifact pack before the pilot starts.\n"
}

func internalDemoScript() string {
	return "# Internal Demo Script\n\n## Goal\n\nShow internal company validation before selling to external pilot customers.\n\n## Audience\n\nFounder, QA lead, engineering lead, and implementation owner.\n\n## Flow\n\n1. Start with a synthetic QA artifact pack: one bug report, one regression checklist, one API test note, one GraphQL note, and one ETL validation note.\n2. Run the BQA-OS build flow and show generated knowledge, skills, agents, workflows, registry entries, and sales package artifacts.\n3. Open one workflow and show how a QA lead would reuse it in a repeat regression or onboarding task.\n4. Explain the paid pilot boundary: 10-30 sanitized artifacts, 3-5 workflows, two weeks, no unlimited free work.\n5. Close with the internal company validation decision: who would pay, what measurable outcome matters, and what sample artifact pack is needed next.\n\n## Talk Track\n\nBefore BQA-OS, QA knowledge sits in tickets, notes, prompts, and memory. After the pilot, repeated checks become reusable workflows and project-specific memory that a human QA owner can inspect and reuse.\n"
}

func discoveryCallScript() string {
	return "# Discovery Call Script\n\n## Persona\n\nQA Lead, QA Automation Lead, CTO, or VP Engineering.\n\n## Objective\n\nFind out where lost QA knowledge, repeated checks, and regression bottlenecks create enough pain for a paid pilot.\n\n## Twenty-Minute Agenda\n\n1. Context: What does the team ship, and where does QA slow down delivery?\n2. Current workflow: Where do test notes, regression checklists, bug reports, prompts, and investigation steps live today?\n3. Repetition: Which checks are repeated every sprint or release but still depend on one person?\n4. Knowledge loss: What breaks when a QA owner is unavailable or a new engineer joins?\n5. AI impact: Has AI coding increased delivery speed enough that QA became the bottleneck?\n6. Pilot fit: Could the team share 10-30 sanitized artifacts for a two-week pilot?\n7. Success: What would make the pilot worth paying for?\n\n## Close\n\nIf the pain is real, propose a paid 2-week QA Memory Pilot and ask for a sample artifact pack or kickoff date.\n"
}

func onboardingChecklist() string {
	return "# Onboarding Checklist\n\n## Before Kickoff\n\n- Confirm buyer, QA owner, implementation owner, and decision maker.\n- Confirm the pilot is paid, time-boxed, and limited to the agreed scope.\n- Define success criteria and one measurable outcome.\n- Confirm Synthetic artifacts only for public demos and sanitized artifacts only for client pilots.\n\n## Client Inputs\n\n- 10-30 sanitized QA artifacts.\n- One target use case: API regression, GraphQL functional testing, ETL data quality, bug report standardization, or QA onboarding.\n- One internal validation session time slot.\n\n## BQA-OS Outputs\n\n- Reusable QA knowledge base.\n- 3-5 AI-assisted QA workflows.\n- Internal readout with risks, time saved hypothesis, and continuation recommendation.\n\n## Guardrails\n\n- No real session logs in public repos.\n- No secrets or customer data.\n- No promise of a fully autonomous QA agent.\n- Custom integration work requires a paid scope change.\n"
}

func outreachSamples() string {
	return "# Outreach Samples\n\n## Slack\n\nWe are testing a paid pilot for BQA-OS: a 2-week QA Memory Pilot that turns 10-30 sanitized QA artifacts into reusable QA workflows and project memory. Do you have repeated regression checks, API/GraphQL validation steps, or bug report patterns that still depend on one QA owner?\n\n## LinkedIn\n\nTeams using AI coding often ship faster, but QA knowledge still sits in notes, tickets, prompts, and repeated manual checks. BQA-OS is offering a paid pilot to convert 10-30 sanitized QA artifacts into reusable QA memory and 3-5 human-in-the-loop workflows. Worth comparing against your current regression bottleneck?\n\n## Email\n\nSubject: Paid pilot for reusable QA memory\n\nHi {name},\n\nI am validating BQA-OS with teams where regression knowledge is scattered across bug reports, test notes, prompts, and repeated checks.\n\nThe offer is a paid 2-week QA Memory Pilot. You send 10-30 sanitized QA artifacts; we return a project-specific QA knowledge base and 3-5 reusable workflows for API regression, GraphQL testing, ETL validation, bug report standardization, or QA onboarding.\n\nWould it be useful to review one synthetic before/after example and see whether your team has a fit?\n"
}

func pricingHypothesis() string {
	return "# Pricing Hypothesis\n\n## Starting Point\n\nPricing hypothesis: price the first internal-to-external pilot as a paid, fixed-scope service rather than a platform subscription.\n\n## Package\n\n- 2-week QA Memory Pilot.\n- 10-30 sanitized artifacts.\n- 3-5 reusable QA workflows.\n- One kickoff, one mid-pilot review, one business review.\n\n## Hypothesis\n\n- Internal validation: no charge, but require explicit success criteria and artifact owner.\n- First external pilots: USD 1,500-3,000 fixed fee.\n- High-complexity ETL or integration-heavy pilots: quote separately as paid implementation scope.\n\n## Renewal Path\n\nIf the pilot creates reusable workflows, propose a monthly package for maintaining QA memory, adding workflows, and supporting one QA owner.\n"
}

func internalStakeholderFAQ() string {
	return "# Internal Stakeholder FAQ\n\n## Is this an autonomous QA agent?\n\nNo. BQA-OS is a human-in-the-loop QA memory and workflow layer. It helps QA owners reuse knowledge and repeated checks.\n\n## What does the first pilot sell?\n\nA 2-week QA Memory Pilot, not a generic AI QA platform.\n\n## What data can be used?\n\nSynthetic data for demos and sanitized client artifacts for pilots. Do not use private logs, secrets, customer data, or business-sensitive records in public artifacts.\n\n## What is out of scope?\n\nUnlimited free pilots, custom enterprise integration, production data access, and claims that BQA-OS fully replaces QA engineers.\n\n## What proves value?\n\nA QA owner can reuse at least one generated workflow during internal validation, and stakeholders agree that 3-5 workflows reduce repeated QA work or onboarding friction.\n"
}

func registryIndex() string {
	return "registry:\n  version: 1\n  skills: registry/skills.yaml\n  agents: registry/agents.yaml\n  workflows: registry/workflows.yaml\n  knowledge: knowledge/project_profile.yaml\n"
}

func registryIndexWithSales() string {
	return "registry:\n  version: 1\n  skills: registry/skills.yaml\n  agents: registry/agents.yaml\n  workflows: registry/workflows.yaml\n  sales: registry/sales.yaml\n  knowledge: knowledge/project_profile.yaml\n"
}

func registrySkills() string {
	return "skills:\n  - id: etl-log-investigation\n    path: skills/etl-log-investigation.md\n    domain: etl\n  - id: runtime-trace-review\n    path: skills/runtime-trace-review.md\n    domain: runtime\n"
}

func registryAgents() string {
	return "agents:\n  - id: etl-qa-agent\n    path: agents/etl-qa-agent.md\n    domain: etl\n  - id: runtime-agent\n    path: agents/runtime-agent.md\n    domain: runtime\n"
}

func registryWorkflows() string {
	return "workflows:\n  - id: etl-verification-workflow\n    path: workflows/etl-verification-workflow.md\n    domain: etl\n  - id: session-knowledge-workflow\n    path: workflows/session-knowledge-workflow.md\n    domain: memory\n"
}

func registrySales() string {
	return "sales:\n  - id: pilot-offer-one-pager\n    path: sales/pilot-offer-one-pager.md\n    domain: pilot-sales\n  - id: internal-demo-script\n    path: sales/internal-demo-script.md\n    domain: pilot-sales\n  - id: discovery-call-script\n    path: sales/discovery-call-script.md\n    domain: discovery\n  - id: onboarding-checklist\n    path: sales/onboarding-checklist.md\n    domain: implementation\n  - id: outreach-samples\n    path: sales/outreach-samples.md\n    domain: outbound\n  - id: pricing-hypothesis\n    path: sales/pricing-hypothesis.md\n    domain: pricing\n  - id: internal-stakeholder-faq\n    path: sales/internal-stakeholder-faq.md\n    domain: stakeholder-alignment\n"
}
