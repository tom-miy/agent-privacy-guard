# MCP Trust Config

`templates/agent-privacy-guard/mcp-trust.yaml` is **not a standard MCP server launch configuration**.

It is supplemental metadata used by `agent-privacy-guard` to describe the trust boundary for each MCP server.

## What This File Is

`templates/agent-privacy-guard/mcp-trust.yaml` describes:

- Which MCP servers are treated as internal.
- Which MCP servers are treated as external / public.
- Which sanitization level applies to context derived from each MCP server.

This file uses YAML because it is gateway policy metadata, not a standard MCP JSON config, and it belongs beside `templates/agent-privacy-guard/policy.yaml`.

```yaml
# agent-privacy-guard MCP trust metadata.
# This is not a standard MCP server launch config.
servers:
  internal-customer-db:
    trust: internal
    sanitize: none
  external-docs-search:
    trust: public
    sanitize: weak
```

## What This File Is Not

This is not the config used by an MCP client to launch servers.

It does not contain fields commonly found in MCP client configs, such as:

- command
- args
- env
- transport
- server executable path

Those belong in the MCP client configuration for Claude Desktop, Claude Code, Cursor, Codex CLI, or whichever client is launching the server.

## Layout

| Field | Meaning |
|---|---|
| `servers` | Trust metadata keyed by MCP server name. |
| `servers.<name>.trust` | Trust boundary such as `internal` or `public`. |
| `servers.<name>.sanitize` | Sanitization level for context derived from that MCP server. |

## Relationship To `templates/agent-privacy-guard/policy.yaml`

`templates/agent-privacy-guard/policy.yaml` is target policy for LLMs and agents.

`templates/agent-privacy-guard/mcp-trust.yaml` is trust metadata for MCP servers.

```text
templates/agent-privacy-guard/policy.yaml
  -> target policy for claude_api, cursor, codex, internal_mcp, external_mcp, etc.

templates/agent-privacy-guard/mcp-trust.yaml
  -> MCP server metadata for internal-customer-db, external-docs-search, etc.
```

In the current minimal implementation, `templates/agent-privacy-guard/mcp-trust.yaml` is an integration sample / design reference. In production integration, it is intended to be matched against the real MCP client config and used for gateway routing decisions.
