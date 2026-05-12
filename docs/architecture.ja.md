# Architecture

`agent-privacy-guard` はローカルで動作する AI Agent Gateway です。coding agent、MCP server、local repository、external LLM API の間に入り、外部送信前後の制御を行います。

## Layers

- `internal/domain`: policy、finding、placeholder mapping、command risk type。
- `internal/usecase`: sanitizer、inspect flow、restore flow、posthook inspection。
- `internal/infra`: YAML policy loading と JSONL audit logging。
- `internal/interface/cli`: Cobra command と terminal IO。
- `adapters`: Claude Code、Cursor、Copilot、Codex CLI 向けの agent-specific integration notes。

## Flow

1. agent または MCP server が context を生成する。
2. `inspect` が configured target に対する outbound risk を評価する。
3. `sanitize` が secrets と sensitive entities を structured placeholder に置換する。
4. external agent は sanitized prompt のみを受け取る。
5. `restore` が local mapping を使って placeholder を復元できる。
6. `posthook` が response 内の dangerous command を検知する。
7. audit event が JSON Lines として追記される。

## Scope Control

このプロジェクトは enterprise DLP ではなく、意図的に小さな gateway として実装しています。policy は明示的な YAML、placeholder mapping は local file、各 integration は shared usecase の上に薄く置く adapter として扱います。
