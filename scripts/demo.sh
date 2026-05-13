#!/usr/bin/env bash
set -euo pipefail

echo "== 1. Policy gate: inspect outbound risk for examples/prompt.txt =="
go run ./cmd/agent-privacy-guard inspect --policy templates/agent-privacy-guard/policy.yaml --input examples/prompt.txt --target claude_api
echo

echo "== 2. Preview: show values that will become structured placeholders =="
go run ./cmd/agent-privacy-guard preview --policy templates/agent-privacy-guard/policy.yaml --input examples/prompt.txt --target claude_api
echo

echo "== 3. Before: raw prompt before anonymization =="
cat examples/prompt.txt
echo

echo "== 4. After: sanitized prompt that would be sent to the external target =="
go run ./cmd/agent-privacy-guard sanitize --policy templates/agent-privacy-guard/policy.yaml --input examples/prompt.txt --target claude_api --mapping-out /tmp/agent-privacy-guard.mapping.json
echo

echo "== 5. Local-only restore mapping written to /tmp/agent-privacy-guard.mapping.json =="
cat /tmp/agent-privacy-guard.mapping.json
echo

echo "== 6. Posthook: inspect sample agent response for dangerous commands =="
go run ./cmd/agent-privacy-guard posthook --policy templates/agent-privacy-guard/policy.yaml --input examples/agent-response.txt --target claude_api
