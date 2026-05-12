# MCP Trust Config

`configs/mcp-trust.yaml` は **通常の MCP server 起動設定ではありません**。

これは `agent-privacy-guard` が MCP server ごとの trust boundary を判断するための補助 metadata です。

## What This File Is

`configs/mcp-trust.yaml` は次を表します。

- どの MCP server を internal と見なすか。
- どの MCP server を external / public と見なすか。
- MCP server 由来の context にどの sanitization level を適用するか。

YAML にしているのは、この file が通常の MCP JSON config ではなく、`configs/policy.yaml` と同じ gateway policy metadata だからです。

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

これは MCP client が server を起動するための標準 config ではありません。

たとえば、一般的な MCP config にあるような次の情報は持ちません。

- command
- args
- env
- transport
- server executable path

それらは Claude Desktop、Claude Code、Cursor、Codex CLI など、それぞれの MCP client 側の config に置きます。

## Layout

| Field | Meaning |
|---|---|
| `servers` | MCP server name を key にした trust metadata。 |
| `servers.<name>.trust` | `internal` / `public` などの trust boundary。 |
| `servers.<name>.sanitize` | MCP server 由来 context に適用する sanitization level。 |

## Relationship To `configs/policy.yaml`

`configs/policy.yaml` は LLM / agent target ごとの policy です。

`configs/mcp-trust.yaml` は MCP server ごとの trust metadata です。

```text
configs/policy.yaml
  -> claude_api, cursor, codex, internal_mcp, external_mcp などの target policy

configs/mcp-trust.yaml
  -> internal-customer-db, external-docs-search などの MCP server metadata
```

現時点の minimal implementation では、`configs/mcp-trust.yaml` は integration sample / design reference として置いています。production 連携では、実際の MCP client config と突き合わせて、この metadata を gateway routing に使う想定です。
