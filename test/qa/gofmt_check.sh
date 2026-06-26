#!/usr/bin/env bash
set -euo pipefail

unformatted="$(gofmt -l .)"

if [[ -n "$unformatted" ]]; then
  echo "::error title=Go files need gofmt::Run: gofmt -w <files>"
  echo "$unformatted"
  exit 1
fi

echo "OK: all Go files are gofmt-formatted"
