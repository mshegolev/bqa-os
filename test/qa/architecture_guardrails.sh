#!/usr/bin/env bash
set -euo pipefail

if [[ ! -d internal/core/knowledge ]]; then
  echo "OK: internal/core/knowledge does not exist yet; skipping Knowledge Extractor core import guard"
  exit 0
fi

forbidden_imports='"(os|io/fs|path/filepath|github.com/spf13/cobra)"'

if grep -R -n -E "$forbidden_imports" internal/core/knowledge; then
  cat <<'MSG'
::error title=Knowledge core imports forbidden infrastructure::internal/core/knowledge must stay pure business logic. Do not import os, io/fs, path/filepath, Cobra, or adapter packages from core. Put filesystem and CLI wiring in ports/adapters/app layers.
MSG
  exit 1
fi

echo "OK: Knowledge Extractor core import guard passed"
