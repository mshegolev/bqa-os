# Knowledge Extractor

The Knowledge Extractor is the first BQA-OS build stage that turns normalized AI coding sessions into reusable QA knowledge artifacts.

## Workflow

```bash
bqa discover
bqa ingest
bqa build
```

## Inputs

`bqa build` expects normalized sessions created by `bqa ingest`:

```text
.bqa/input/sessions/index.json
.bqa/input/sessions/normalized/
```

## Outputs

`bqa build` writes initial YAML artifacts into:

```text
.bqa/knowledge/
```

Generated files:

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

This is still an early vertical slice. Future versions will add richer parsing, deduplication, scoring, project-aware extraction, and agent/skill/workflow generation.

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
