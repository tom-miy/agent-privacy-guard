#!/usr/bin/env bash
set -euo pipefail

# Reads an agent response from stdin and reports dangerous commands or patches.
agent-privacy-guard posthook --target "${AGENT_TARGET:-claude_api}"
