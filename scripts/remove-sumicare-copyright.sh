#!/usr/bin/env bash
# Remove erroneous "Sumicare" copyright lines from repository source files.
# Idempotent: safe to run multiple times.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PATTERN='Copyright \(c\) 2026 Sumicare'
changed=0

while IFS= read -r -d '' file; do
  if ! grep -q "$PATTERN" "$file"; then
    continue
  fi

  # Remove the Sumicare copyright line.
  if sed --version >/dev/null 2>&1; then
    sed -i "/^\/\/ Copyright (c) 2026 Sumicare$/d" "$file"
  else
    sed -i '' "/^\/\/ Copyright (c) 2026 Sumicare$/d" "$file"
  fi

  # Collapse a double blank comment line left before the Apache license block.
  python3 - "$file" <<'PY'
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
text = path.read_text()
updated = text.replace("//\n//\n// Licensed", "//\n// Licensed", 1)
if updated != text:
    path.write_text(updated)
PY

  echo "updated: $file"
  changed=$((changed + 1))
done < <(grep -rl "$PATTERN" . \
  --exclude-dir=.git \
  --exclude-dir=bin \
  --exclude-dir=node_modules \
  --exclude='remove-sumicare-copyright.sh' \
  -z || true)

if [[ "$changed" -eq 0 ]]; then
  echo "No Sumicare copyright lines found."
else
  echo "Removed Sumicare copyright from $changed file(s)."
fi

if grep -r "Sumicare" . \
  --exclude-dir=.git \
  --exclude-dir=bin \
  --exclude-dir=node_modules \
  --exclude='remove-sumicare-copyright.sh' >/dev/null 2>&1; then
  echo "error: Sumicare references still remain" >&2
  grep -r "Sumicare" . \
    --exclude-dir=.git \
    --exclude-dir=bin \
    --exclude-dir=node_modules \
    --exclude='remove-sumicare-copyright.sh' >&2 || true
  exit 1
fi

echo "Done."