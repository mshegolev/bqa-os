#!/usr/bin/env bash
set -euo pipefail

cleanup() {
  rm -rf .bqa
}
trap cleanup EXIT

rm -rf .bqa

set +e
output="$(go run ./cmd/bqa build 2>&1)"
status=$?
set -e

echo "$output"

if [[ "$status" -eq 0 ]]; then
  echo "::error title=bqa build should fail without index::Expected non-zero exit when .bqa/input/sessions/index.json is missing."
  exit 1
fi

if ! grep -Fq ".bqa/input/sessions/index.json" <<<"$output"; then
  echo "::error title=Missing index error is not actionable::Expected error to mention .bqa/input/sessions/index.json."
  exit 1
fi

if ! grep -Eq "bqa discover|bqa ingest2" <<<"$output"; then
  echo "::error title=Missing index error lacks recovery hint::Expected error to suggest running bqa discover or bqa ingest2."
  exit 1
fi

echo "OK: missing index negative test passed"
