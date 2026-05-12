#!/usr/bin/env bash
set -euo pipefail

# Reads raw outbound prompt text from stdin.
# Writes sanitized prompt text to stdout.
# Keeps reversible placeholder mapping local.
agent-privacy-guard sanitize \
  --target claude_api \
  --policy configs/policy.yaml \
  --mapping-out .agent-privacy-guard.mapping.json
