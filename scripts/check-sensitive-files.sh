#!/usr/bin/env bash
set -euo pipefail

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "error: this check must run inside a git work tree" >&2
  exit 2
fi

blocked_patterns=(
  '(^|/)\.agent-privacy-guard/entities\.local\.yaml$'
  '(^|/)\.agent-privacy-guard/mapping\.json$'
  '(^|/)\.agent-privacy-guard\.mapping\.json$'
  '(^|/).*\.mapping\.json$'
)

blocked=()
while IFS= read -r file; do
  [[ -z "$file" ]] && continue
  for pattern in "${blocked_patterns[@]}"; do
    if [[ "$file" =~ $pattern ]]; then
      blocked+=("$file")
      break
    fi
  done
done < <(git diff --cached --name-only --diff-filter=ACMR)

if [[ ${#blocked[@]} -gt 0 ]]; then
  echo "Blocked sensitive local files from being committed:" >&2
  for file in "${blocked[@]}"; do
    echo "  - $file" >&2
  done
  echo >&2
  echo "These files may contain real customer names, internal identifiers, or reversible placeholder mappings." >&2
  echo "Keep them untracked, or store encrypted source files and decrypt them only at runtime." >&2
  exit 1
fi

echo "Sensitive file check: OK"
