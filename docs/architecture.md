# Architecture

`agent-privacy-guard` is a local AI Agent Gateway. It sits between coding agents, MCP servers, local repositories, and external LLM APIs.

## Layers

- `internal/domain`: policy, findings, placeholder mapping, and command risk types.
- `internal/usecase`: sanitizer, inspect flow, restore flow, and posthook inspection.
- `internal/infra`: YAML policy loading and JSONL audit logging.
- `internal/interface/cli`: Cobra commands and terminal IO.
- `adapters`: agent-specific integration notes for Claude Code, Cursor, Copilot, and Codex CLI.

## Flow

1. An agent or MCP server produces context.
2. `inspect` evaluates outbound risk for the configured target.
3. `sanitize` replaces secrets and sensitive entities with structured placeholders.
4. The external agent receives only the sanitized prompt.
5. `restore` can map placeholders back locally.
6. `posthook` detects dangerous commands in the response.
7. Audit events are appended as JSON Lines.

## Scope Control

The project intentionally implements a minimal gateway, not enterprise DLP. Policy is explicit YAML, placeholder mapping is local, and integrations are thin adapters over shared use cases.
