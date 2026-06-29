# Knowledge Extractor

The Knowledge Extractor is the first BQA-OS build stage that turns normalized AI coding sessions into reusable QA knowledge artifacts and starter BQA runtime artifacts.

## Workflow

```bash
bqa discover
bqa ingest
bqa build
```

For internal pilot validation, generate the Monday sales package with:

```bash
bqa build --sales-package
```

To generate a local ETL QA Agent Pack for Codex and Claude Code, run:

```bash
bqa etl-agent-pack
```

## Inputs

`bqa build` expects normalized sessions created by `bqa ingest`:

```text
.bqa/input/sessions/index.json
.bqa/input/sessions/normalized/
```

## Outputs

`bqa build` writes initial YAML knowledge artifacts into:

```text
.bqa/knowledge/
```

Generated knowledge files:

```text
etl_patterns.yaml
graphql_patterns.yaml
api_patterns.yaml
data_quality_patterns.yaml
common_bugs.yaml
successful_prompts.yaml
droid_patterns.yaml
runtime_patterns.yaml
project_profile.yaml
```

`bqa build` also writes starter BQA artifacts into:

```text
.bqa/skills/
.bqa/agents/
.bqa/workflows/
```

Generated starter artifacts:

```text
skills/etl-log-investigation.md
skills/runtime-trace-review.md
agents/etl-qa-agent.md
agents/runtime-agent.md
workflows/etl-verification-workflow.md
workflows/session-knowledge-workflow.md
```

When `--sales-package` is enabled, `bqa build` also writes internal pilot sales
materials into:

```text
.bqa/sales/
```

Generated sales package files:

```text
sales/pilot-offer-one-pager.md
sales/internal-demo-script.md
sales/discovery-call-script.md
sales/onboarding-checklist.md
sales/outreach-samples.md
sales/pricing-hypothesis.md
sales/internal-stakeholder-faq.md
registry/sales.yaml
```

These files are for the 2-week QA Memory Pilot offer: internal validation first,
then external paid pilot discovery. Use synthetic artifacts for public demos and
sanitized artifacts only for customer pilots. Do not include private repo data,
real session logs, secrets, or customer records in public outputs.

`bqa etl-agent-pack` reads `.bqa/knowledge/` and normalized sessions when they
exist, uses aggregate counts only, and writes a synthetic-safe pack into:

```text
.bqa/output/etl-agent-pack/
```

Generated ETL pack files:

```text
statistics/summary.md
agents/codex-etl-qa-agent.md
agents/claude-code-etl-qa-agent.md
workflows/etl-regression-workflow.md
workflows/data-reconciliation-workflow.md
workflows/data-quality-validation-workflow.md
specs/etl-test-spec-template.md
specs/source-to-target-mapping-review-checklist.md
prompts/codex-etl-qa-agent-prompt.md
prompts/claude-code-etl-qa-agent-prompt.md
examples/synthetic-etl-reconciliation-example.md
README_NEXT_STEPS.md
```

## Current extraction strategy

The current implementation is intentionally heuristic and deterministic. It scans normalized session text for domain signals related to:

- Big Data and ETL testing
- GraphQL functional testing
- API and contract testing
- Data quality validation
- Droid / factory.ai sessions
- runtime execution patterns
- common bugs and failures
- successful prompt candidates

The extractor is source-aware and includes basic noise filtering. For example, it avoids treating GitHub environment variables such as `github_graphql_url` as GraphQL testing evidence.

This is still an early vertical slice. Future versions will add richer parsing, deduplication, scoring, project-aware extraction, and richer agent/skill/workflow generation.

## Architecture

The implementation follows ADR-0002 hexagonal architecture:

```text
CLI
↓
core/knowledge use case
↓
ports
↓
adapters/fs
```

The build command also runs a small `core/artifacts` generation use case after knowledge extraction:

```text
CLI
↓
core/artifacts use case
↓
ports
↓
adapters/fs
```

The Cobra command only parses flags, creates adapters, calls core use cases, and prints a summary.
