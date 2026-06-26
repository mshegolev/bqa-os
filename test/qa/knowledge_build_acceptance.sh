#!/usr/bin/env bash
set -euo pipefail

cleanup() {
  rm -rf .bqa
}
trap cleanup EXIT

rm -rf .bqa
mkdir -p .bqa/input/sessions/normalized

cat > .bqa/input/sessions/index.json <<'JSON'
{
  "sessions": [
    {
      "id": "synthetic-etl-001",
      "normalized_path": "normalized/synthetic-etl-001.md"
    },
    {
      "id": "synthetic-graphql-api-001",
      "normalized_path": "normalized/synthetic-graphql-api-001.md"
    }
  ]
}
JSON

cat > .bqa/input/sessions/normalized/synthetic-etl-001.md <<'MD'
# Synthetic ETL QA Session

## Successful prompt
Create ETL QA checks for partition freshness, row count reconciliation,
duplicate detection, null validation, schema drift, and late-arriving events.

## Bug
Failed validation: duplicate rows after retry caused row count mismatch.

## Data quality
Check freshness, completeness, accuracy, consistency, and reconciliation.
MD

cat > .bqa/input/sessions/normalized/synthetic-graphql-api-001.md <<'MD'
# Synthetic GraphQL and API QA Session

## Successful prompt
Generate GraphQL tests for query pagination, resolver authorization,
nullable fields, fragments, and schema compatibility.

## API checks
Validate status code, response payload, auth, timeout, idempotency,
and contract compatibility.
MD

set +e
output="$(go run ./cmd/bqa build 2>&1)"
status=$?
set -e

echo "$output"

if [[ "$status" -ne 0 ]]; then
  echo "::error title=bqa build failed::Expected synthetic Knowledge Extractor build to complete. Check missing index handling, normalized session reader, extractor use case, and writer wiring."
  exit 1
fi

required_summary=(
  "BQA knowledge build completed"
  "Sessions processed: 2"
  "Artifacts created: 7"
  "Output directory: .bqa/knowledge"
)

for line in "${required_summary[@]}"; do
  if ! grep -Fq "$line" <<<"$output"; then
    echo "::error title=bqa build summary mismatch::Expected summary line not found: $line"
    exit 1
  fi
done

expected_files=(
  etl_patterns.yaml
  graphql_patterns.yaml
  api_patterns.yaml
  data_quality_patterns.yaml
  common_bugs.yaml
  successful_prompts.yaml
  project_profile.yaml
)

for file in "${expected_files[@]}"; do
  path=".bqa/knowledge/$file"
  if [[ ! -s "$path" ]]; then
    echo "::error title=Missing knowledge artifact::Expected non-empty artifact file: $path"
    exit 1
  fi
done

if ! grep -Rqi "partition" .bqa/knowledge/etl_patterns.yaml; then
  echo "::error title=ETL artifact content mismatch::Expected etl_patterns.yaml to include synthetic ETL signal: partition"
  exit 1
fi

if ! grep -Rqi "pagination" .bqa/knowledge/graphql_patterns.yaml; then
  echo "::error title=GraphQL artifact content mismatch::Expected graphql_patterns.yaml to include synthetic GraphQL signal: pagination"
  exit 1
fi

if ! grep -Rqi "status code" .bqa/knowledge/api_patterns.yaml; then
  echo "::error title=API artifact content mismatch::Expected api_patterns.yaml to include synthetic API signal: status code"
  exit 1
fi

if ! grep -Rqi "duplicate" .bqa/knowledge/common_bugs.yaml; then
  echo "::error title=Common bugs artifact content mismatch::Expected common_bugs.yaml to include synthetic bug signal: duplicate"
  exit 1
fi

echo "OK: bqa build Knowledge Extractor acceptance test passed"
