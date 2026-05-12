#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

scripts/check-sensitive-files.sh

go test ./...
go run ./cmd/agent-privacy-guard validate

echo "agent-privacy-guard pre-commit checks: OK"
