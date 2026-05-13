# Integration Guide

この document は、`agent-privacy-guard` を通常のアプリ開発 repository から利用する方法を説明します。

この repository は CLI / gateway 本体の source repository です。アプリ開発 repository にそのままコピーして使う雛形ではありません。

`agent-privacy-guard` は通常、自動的に適用される前提です。

開発者が毎回 `inspect` や `sanitize` を手で実行するのではなく、Claude Code、Cursor、Codex CLI、CI、または独自 wrapper の前後に組み込みます。

## How It Runs

典型的には次のように動きます。

```text
developer writes prompt / agent collects context
  -> prehook or wrapper calls agent-privacy-guard sanitize
  -> sanitized prompt is sent to external LLM
  -> posthook calls agent-privacy-guard posthook
  -> response is reviewed or applied locally
```

つまり `sanitize` と `posthook` は、人間が都度呼ぶ command ではなく、agent 側の hook から自動実行される想定です。

## What To Copy Into A Normal Development Repository

通常の開発 repository に流用する場合、この repository 全体をコピーする必要はありません。

`agent-privacy-guard` CLI は別途 install しておき、対象 repository には次のような policy / hook だけを置きます。

手動で file をコピーする代わりに、この repository の `install.sh` で配置できます。

```bash
./install.sh --target /path/to/your-app
```

既に `.agent-privacy-guard/` がある場合は上書きしません。上書きしたい場合だけ `--force` を付けます。

```bash
./install.sh --target /path/to/your-app --force
```

これにより対象 repository に次が作成されます。

```text
.agent-privacy-guard/
  policy.yaml
  mcp-trust.yaml
  entities.local.example.yaml
  .gitignore
  hooks/
    prehook.sh
    posthook.sh
```

この repository の sample から流用するなら、必要なのは次の最小セットです。

```text
templates/agent-privacy-guard/policy.yaml
hooks/claude-code-prehook.sh
hooks/agent-posthook.sh
```

必要に応じて追加します。

```text
templates/agent-privacy-guard/mcp-trust.yaml
AGENTS.md
CLAUDE.md
.cursorrules
.codex/config.toml
```

`templates/agent-privacy-guard/mcp-trust.yaml` は通常の MCP server 起動 config ではなく、MCP server ごとの trust metadata です。詳しくは [mcp-trust-config.ja.md](mcp-trust-config.ja.md) を参照してください。

## What Not To Copy

次は `agent-privacy-guard` 本体の実装や demo 用なので、通常の開発 repository には不要です。

```text
cmd/
internal/
adapters/
docs/
examples/
scripts/
go.mod
go.sum
docker/cli/Dockerfile
mise.toml
```

繰り返しになりますが、`agent-privacy-guard` は別途 install された CLI として使い、対象 repository には policy と hook だけを置くのが基本です。

## Recommended Layout In A Target Repository

対象の開発 repository では、メイン実装と gateway 設定を分けるために、次のような layout を推奨します。

```text
your-app/
  src/
  package.json
  README.md

  .agent-privacy-guard/
    policy.yaml
    mcp-trust.yaml
    entities.local.example.yaml
    .gitignore
    hooks/
      prehook.sh
      posthook.sh
```

この repository の sample を使うなら、次のように対応します。

```text
templates/agent-privacy-guard/policy.yaml                 -> .agent-privacy-guard/policy.yaml
templates/agent-privacy-guard/mcp-trust.yaml              -> .agent-privacy-guard/mcp-trust.yaml
templates/agent-privacy-guard/entities.local.example.yaml -> .agent-privacy-guard/entities.local.example.yaml
hooks/claude-code-prehook.sh                              -> .agent-privacy-guard/hooks/prehook.sh
hooks/agent-posthook.sh                                   -> .agent-privacy-guard/hooks/posthook.sh
```

hook 内の `--policy` path も合わせて変更します。

```bash
agent-privacy-guard sanitize \
  --target claude_api \
  --policy .agent-privacy-guard/policy.yaml \
  --mapping-out .agent-privacy-guard/mapping.json
```

## Why This Repository Looks Larger

この repository は portfolio / reference implementation なので、次の 2 種類を同居させています。

| Area | Purpose |
|---|---|
| `cmd/`, `internal/` | `agent-privacy-guard` CLI 本体の実装。 |
| `templates/agent-privacy-guard/`, `hooks/`, `examples/`, `docs/` | 利用例、policy sample、integration sample。 |

通常の application repository に必要なのは後者の一部だけです。

## Mental Model

```text
agent-privacy-guard repository
  = CLI product source code + examples

your development repository
  = your app source code + small .agent-privacy-guard/ policy folder
```

本体実装と対象 repository の設定を混ぜないために、実運用では `.agent-privacy-guard/` のような専用 directory に policy と hook を閉じ込めるのがおすすめです。
