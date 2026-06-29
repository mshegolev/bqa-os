# Knowledge Extractor

The Knowledge Extractor is the first BQA-OS build stage that turns normalized AI coding sessions into reusable QA knowledge artifacts.

## Workflow

```bash
bqa discover
bqa ingest2
bqa build
```

BQA-OS is local-first. This workflow reads local session artifacts, creates local normalized inputs, and writes local knowledge files. It helps Codex reuse QA context, but it does not replace a QA owner or run as a fully autonomous QA agent.

## Inputs

`bqa build` expects normalized sessions created by `bqa ingest2`:

```text
.bqa/input/sessions/index.json
.bqa/input/sessions/normalized/
```

The index points to normalized Markdown session files. For public demos, use synthetic session text only. For private pilots, sanitize source material before it enters `.bqa/input/sessions/`.

## Outputs

`bqa build` writes initial YAML knowledge artifacts into:

```text
.bqa/knowledge/
```

Generated knowledge files:

- `etl_patterns.yaml` - ETL validation signals such as Spark, Hive, Oozie, reconciliation, row counts, and source/target checks.
- `graphql_patterns.yaml` - GraphQL testing signals such as schema, query, mutation, resolver, and introspection checks.
- `api_patterns.yaml` - REST/API contract testing signals such as endpoints, status codes, OpenAPI, and payload checks.
- `data_quality_patterns.yaml` - data validation signals such as duplicate, null, checksum, row-count, and schema-drift checks.
- `common_bugs.yaml` - recurring failure signals from errors, regressions, exceptions, stack traces, and flaky behavior.
- `successful_prompts.yaml` - prompt patterns that may be useful when asking Codex or another runtime to repeat QA work.
- `project_profile.yaml` - a compact project summary with session counts and detected domain-signal counts.

These files are intentionally small and inspectable. A typical Codex workflow is to ask the runtime to read the relevant `.bqa/knowledge/*.yaml` files before planning a QA task, then have a human review the proposed checks and evidence.

## Synthetic ETL example

Use synthetic content like this for public demos:

```markdown
# Synthetic session: DemoShop orders ETL

Task:
Validate the demo_orders_daily ETL pipeline.

Notes:
- Source table: synthetic_raw.orders_delta
- Target table: synthetic_mart.orders_daily
- QA checks: row count reconciliation, duplicate order_id check, not null customer_id check
- Failure observed: checksum mismatch after Spark transformation

Prompt:
Please summarize the ETL validation risk and propose manual checks.
```

After this content is discovered and normalized by `bqa discover` and `bqa ingest2`, `bqa build` can extract ETL, data-quality, failure, and prompt signals into the seven YAML files. The values above are synthetic placeholders, not real customer data.

## Current extraction strategy

The current implementation is intentionally heuristic and deterministic. It scans normalized session text for domain signals related to:

- Big Data and ETL testing
- GraphQL functional testing
- API and contract testing
- Data quality validation
- common bugs and failures
- successful prompt candidates

The extractor is source-aware and includes basic noise filtering. For example, it avoids treating GitHub environment variables such as `github_graphql_url` as GraphQL testing evidence.

This is still an early vertical slice. Future versions will add richer parsing, deduplication, scoring, project-aware extraction, and richer agent/skill/workflow generation.

## Privacy

Do not commit real `.bqa/` outputs, raw session logs, credentials, private URLs, customer names, production identifiers, or support data. Public examples should be synthetic. Private project value belongs in local `.bqa` workspaces or private BQA Brain repositories after sanitization.

## Manual verification

Run these checks after changing the extractor or its documentation:

```bash
go run ./cmd/bqa --help
go run ./cmd/bqa discover --help
go run ./cmd/bqa ingest2 --help
go run ./cmd/bqa build --help
go test ./...
git diff --check -- README.md docs/knowledge-extractor.md
```

After running the workflow against synthetic input, inspect the local artifacts:

```bash
ls .bqa/knowledge
sed -n '1,80p' .bqa/knowledge/etl_patterns.yaml
sed -n '1,80p' .bqa/knowledge/project_profile.yaml
```

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

The Cobra command only parses flags, creates adapters, calls the core use case, and prints a summary.
