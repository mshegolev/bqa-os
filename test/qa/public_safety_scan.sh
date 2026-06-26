#!/usr/bin/env bash
set -euo pipefail

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "::error title=Safety scan requires git::Run this script from the repository root."
  exit 1
fi

files="$(git ls-files | grep -Ev '(^vendor/|^dist/|^bin/)' || true)"

if [[ -z "$files" ]]; then
  echo "OK: no tracked files to scan"
  exit 0
fi

# Keep this scan intentionally conservative. It catches obvious mistakes in public fixtures/docs
# without replacing a dedicated secret scanner.
private_data_pattern='(real session log|customer data|private repo knowledge|business-specific data|customer email|production password|production token)'
credential_assignment_pattern='(password|secret|token|api_key|access_key)[[:space:]]*[:=][[:space:]]*[^[:space:]<>]+'
private_key_pattern='BEGIN .*PRIVATE KEY'

if echo "$files" | xargs grep -I -n -E "$private_data_pattern|$credential_assignment_pattern|$private_key_pattern"; then
  cat <<'MSG'
::error title=Potential private data found::Public repo must not contain secrets, credentials, real session logs, customer data, or private BQA Brain knowledge. Replace the match with clearly synthetic fixture text.
MSG
  exit 1
fi

echo "OK: public data safety scan passed"
