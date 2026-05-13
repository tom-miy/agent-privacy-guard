# agent-privacy-guard

[English](README.md)

Claude Code、Cursor、GitHub Copilot、Codex CLI、MCP server、外部 LLM API のための安全な AI Agent Gateway です。

このプロジェクトは、policy-driven な prompt sanitization、外部送信制御、MCP trust boundary、posthook risk inspection、ローカルで復元可能な placeholder mapping を実演します。

## What This Repository Is

この repository は **アプリ開発 repository にそのままコピーして使う雛形ではありません**。

`agent-privacy-guard` という CLI / gateway 本体を実装する repository です。通常のアプリ開発 repository からは、この CLI を install 済み command として呼び出します。

```text
agent-privacy-guard repository
  = CLI / gateway 本体の source code
  = policy sample、hook sample、demo を含む reference implementation

your application repository
  = アプリ本体の source code
  = .agent-privacy-guard/policy.yaml
  = .agent-privacy-guard/hooks/prehook.sh
  = .agent-privacy-guard/hooks/posthook.sh
```

アプリ repository に持っていくのは、基本的に policy と hook だけです。`cmd/`、`internal/`、`go.mod` などの CLI 実装ファイルはコピーしません。

詳細は [docs/integration.ja.md](docs/integration.ja.md) を参照してください。

## Features

- trust level と sanitization level を持つ YAML target policy。
- AWS key、AWS ARN、token、email、internal URL、SSH private key の secret detection。
- `[CLIENT#A]` のような structured placeholder による entity anonymization。
- restore workflow のためのローカル reversible mapping。
- `inspect`、`sanitize`、`preview`、`restore`、`posthook`、`validate` CLI command。
- JSONL audit logging。
- sample hooks、configs、adapters、demo inputs。

## How It Is Used

`agent-privacy-guard` は、普段の開発者が毎回手で `inspect` を打つための tool ではありません。基本的には Claude Code、Cursor、Copilot、Codex CLI などの agent の前後に置く gateway です。

典型的には次のように使います。

```text
local repository / MCP result
  -> prehook or wrapper
  -> agent-privacy-guard sanitize
  -> external LLM / coding agent
  -> agent-privacy-guard posthook
  -> local apply / human review
```

手動でコマンドを呼ぶのは、policy の動作確認、demo、CI、または「この prompt を外に送ってよいか」を人間が確認したいときです。実運用では `hooks/claude-code-prehook.sh` のような hook や、各 agent adapter から呼び出します。

通常の開発 repository に流用する場合、この repository 全体をコピーする必要はありません。持っていく file と推奨 layout は [docs/integration.ja.md](docs/integration.ja.md) にまとめています。

対象 repository への policy / hook 配置は `install.sh` でできます。

```bash
./install.sh --target /path/to/your-app
```

## When To Use Each Command

| Command | 呼ぶタイミング | 目的 | 主な呼び出し元 |
|---|---|---|---|
| `inspect` | prompt を外部送信する前 | risk、検出 entity、policy decision を確認する | 人間、CI、debug script |
| `preview` | policy 調整中、demo 中 | 何が placeholder に置換されるかを見る | 人間 |
| `sanitize` | external LLM / public agent に送る直前 | prompt を匿名化し、mapping を local に保存する | prehook、wrapper、adapter |
| `restore` | agent response を local で読む、または適用する前 | placeholder を local value に戻す | wrapper、human review |
| `posthook` | agent response を受け取った直後 | dangerous command や forbidden patch を検知する | posthook、wrapper、CI |
| `validate` | setup 時、CI | policy と agent config の不足を検知する | 人間、CI |

## Try It In This Repository

この section は、`agent-privacy-guard` CLI 本体 repository 上で動作を確認するための手順です。通常のアプリ repository にこの repo 全体をコピーする手順ではありません。

ここでは次の sample file を使います。

- policy file: `configs/policy.yaml`
- input prompt: `examples/prompt.txt`
- agent response sample: `examples/agent-response.txt`
- mapping output: `/tmp/apg.mapping.json`

最初に依存関係を取得します。

```bash
go mod tidy
```

まず demo 用 prompt がどんな risk を持つか確認します。これは運用中に毎回必須ではなく、policy の確認や CI で使う想定です。

```bash
go run ./cmd/agent-privacy-guard inspect --input examples/prompt.txt --target claude_api
```

次に、どの値が structured placeholder に置換されるかを diff 形式で preview します。policy の調整中に使います。

```bash
go run ./cmd/agent-privacy-guard preview --input examples/prompt.txt --target claude_api
```

外部 LLM に送る直前の prehook が実際に行う処理です。sanitized prompt を標準出力に出し、復元用 mapping を `/tmp/apg.mapping.json` に保存します。

```bash
go run ./cmd/agent-privacy-guard sanitize --input examples/prompt.txt --target claude_api --mapping-out /tmp/apg.mapping.json
```

agent response を local に適用する前に、posthook が dangerous command を確認します。

```bash
go run ./cmd/agent-privacy-guard posthook --input examples/agent-response.txt
```

上のコマンドを 1 つずつ試す代わりに、`scripts/demo.sh` で同じ流れをまとめて確認できます。これは本番運用の command ではなく、匿名化 gateway の挙動を確認するための smoke demo です。

```bash
bash scripts/demo.sh
```

この script で確認できること:

| Step | 何を見るか | 期待する結果 |
|---|---|---|
| 1. inspect | raw prompt の risk 判定 | secret があるため `Outbound Risk: HIGH` になる |
| 2. preview | `configs/policy.yaml` の `entities` と built-in detector で何が置換されるか | `AcmeBank -> [CLIENT#A]` のように表示される |
| 3. raw prompt | 匿名化前の入力 | customer name、internal URL、AWS key がそのまま見える |
| 4. sanitize | 外部 LLM に渡す prompt | 生値が placeholder に置き換わる |
| 5. mapping | local にだけ残す復元情報 | placeholder と元の値の対応が JSON で保存される |
| 6. posthook | agent response の危険検知 | `curl ... | sh` や `rm -rf /` が検出される |

つまり `demo.sh` は「この repo の機能を全部使う入口」ではなく、「匿名化と policy gate が動いていることを短時間で確認するための再現シナリオ」です。

## Policy File

sample policy は `configs/policy.yaml` にあります。ここで「どの target に送るか」と「何をどの placeholder にするか」を設定します。

詳しい layout と各 field の意味は [docs/policy-config.ja.md](docs/policy-config.ja.md) に分けています。

```yaml
targets:
  claude_api:
    trust: public
    sanitize: strong
    allow: true

  local_qwen:
    trust: internal
    sanitize: weak
    allow: true

  internal_mcp:
    trust: internal
    sanitize: none
    allow: true

entities:
  - type: CLIENT
    pattern: "\\b(AcmeBank|ExampleCorp|MegaRetail)\\b"
    scope: prompt

outbound:
  block_on_secret: true
  diff_only: true
```

`AcmeBank -> [CLIENT#A]` のような project-specific な置換は `entities` で設定します。AWS key、email、internal URL、token などの汎用 secret は built-in detector で検出します。

重要: 本物の顧客名、内部サービス名、DB 名などは、それ自体が機密情報になり得ます。`configs/policy.yaml` に直書きして git 管理しないでください。本番では `entity_files` で `configs/entities.local.yaml` のような gitignore 済み local file から読み込むか、SOPS / age / git-crypt などで暗号化して管理し、実行時に復号した file を読み込む運用を推奨します。空の local entity file は設定済みに見えてしまうため、この repository では実ファイルではなく example のみを置きます。

## Commands

prompt を検査し、outbound risk、検出 entity、policy decision を表示します。外部送信を止めるかどうかの判断材料を出す command です。

```bash
agent-privacy-guard inspect --input examples/prompt.txt --target claude_api
```

CI や wrapper で使う場合は、policy が block 判定したときに non-zero exit にします。

```bash
agent-privacy-guard inspect --input examples/prompt.txt --target claude_api --fail-on-block
```

この場合、`outbound allowed: false` なら command が失敗するため、後続の「外部 LLM に送る処理」を止められます。

sanitization の内容を compact diff として確認します。

```bash
agent-privacy-guard preview --input examples/prompt.txt --target claude_api
```

sanitized prompt を出力し、local restore mapping を保存します。これは prehook / wrapper から呼ばれる main command です。

```bash
agent-privacy-guard sanitize --input examples/prompt.txt --target claude_api --mapping-out /tmp/apg.mapping.json
```

local mapping file を使って、agent response 内の placeholder を復元します。

```bash
agent-privacy-guard restore --input response.txt --mapping /tmp/apg.mapping.json
```

agent response 内の dangerous command や forbidden patch target を検知します。

```bash
agent-privacy-guard posthook --input examples/agent-response.txt
```

`configs/policy.yaml` と想定される agent config file を validate します。

```bash
agent-privacy-guard validate
```

## Hook Example

Claude Code に限らず、agent の outbound prompt を標準入力で受け取れる場所にこの gateway を置きます。

`hooks/claude-code-prehook.sh` は、外部送信前に prompt を sanitize する例です。

```bash
cat examples/prompt.txt | hooks/claude-code-prehook.sh
```

この hook は sanitized prompt を標準出力に返します。外部 LLM にはこの sanitized prompt だけを送る想定です。placeholder mapping は `.agent-privacy-guard.mapping.json` に保存され、local restore にだけ使います。

`hooks/agent-posthook.sh` は、agent response を受け取った後に危険な command を検査する例です。

```bash
cat examples/agent-response.txt | hooks/agent-posthook.sh
```

つまり、手動 command は「動作確認の入口」で、実際の使い方は agent の prehook / posthook に組み込む形です。

## Input And Output Samples

### Prehook Input

外部 LLM に送られそうな raw prompt の例です。demo では `examples/prompt.txt` に入っています。

```text
Internal MCP returned customer AcmeBank with database prod-db-tokyo.
Investigate https://billing.internal/incidents/42 and update support@example.com.
AWS key observed in failing test: AKIAIOSFODNN7EXAMPLE
Local path: /Users/mimr/work/acme/service/main.go
```

### `inspect` Output

`inspect` は prompt を送信してよいか判断するための risk report を出します。送信用 prompt は生成しません。

```text
Outbound Risk: HIGH

Detected:
- SECRET:AWS_KEY -> [SECRET:AWS_KEY#A]
- SECRET:EMAIL -> [SECRET:EMAIL#A]
- SECRET:INTERNAL_URL -> [SECRET:INTERNAL_URL#A]
- CLIENT -> [CLIENT#A]
- POSTGRES_DB -> [POSTGRES_DB#A]

Policy:
- outbound allowed: false
- external send blocked because secret was detected
- diff-only context is recommended
```

この結果は次のように活用します。

| 見る場所 | 意味 | 次の action |
|---|---|---|
| `Outbound Risk: HIGH` | 外部送信に高い risk がある | 送信前に内容を見直す |
| `Detected` | 何が検出されたか | policy や entity rule が期待通りか確認する |
| `outbound allowed: false` | policy 上は外部送信不可 | `sanitize` しても送らない、または人間が承認する |
| `external send blocked...` | block 理由 | CI / wrapper の failure reason として使う |

自動化では `--fail-on-block` を付けて、block 判定を exit code に変換します。

```bash
agent-privacy-guard inspect \
  --input examples/prompt.txt \
  --target claude_api \
  --fail-on-block
```

wrapper ではこのように使います。

```bash
if agent-privacy-guard inspect --input prompt.txt --target claude_api --fail-on-block; then
  agent-privacy-guard sanitize --input prompt.txt --target claude_api --mapping-out mapping.json
else
  echo "Outbound prompt was blocked by policy. Review required."
fi
```

つまり `inspect` は「送信用 prompt を作る command」ではなく、「送信してよいかを人間または自動化が判断するための gate」です。

### `preview` Output

`preview` は置換内容を人間が確認するための diff です。

```diff
- AcmeBank
+ [CLIENT#A]

- prod-db-tokyo
+ [POSTGRES_DB#A]

- AKIAIOSFODNN7EXAMPLE
+ [SECRET:AWS_KEY#A]
```

### `sanitize` Output

`sanitize` は prehook が外部 LLM に渡す sanitized prompt を標準出力に返します。

```text
Internal MCP returned customer [CLIENT#A] with database [POSTGRES_DB#A].
Investigate [SECRET:INTERNAL_URL#A] and update [SECRET:EMAIL#A].
AWS key observed in failing test: [SECRET:AWS_KEY#A]
Local path: /Users/[USER]/work/acme/service/main.go
```

この時点で、外部 LLM に渡る prompt から元の customer name、internal URL、email、AWS key は消えています。

| Raw value | Sanitized value | Meaning |
|---|---|---|
| `AcmeBank` | `[CLIENT#A]` | customer name を匿名化 |
| `prod-db-tokyo` | `[POSTGRES_DB#A]` | database name を匿名化 |
| `https://billing.internal/incidents/42` | `[SECRET:INTERNAL_URL#A]` | internal URL を secret placeholder 化 |
| `support@example.com` | `[SECRET:EMAIL#A]` | email を secret placeholder 化 |
| `AKIAIOSFODNN7EXAMPLE` | `[SECRET:AWS_KEY#A]` | AWS key を secret placeholder 化 |
| `/Users/mimr/...` | `/Users/[USER]/...` | local user path を normalize |

`--mapping-out /tmp/apg.mapping.json` を指定すると、local restore 用の mapping も保存されます。この file は外部 LLM に送らない前提です。

```json
[
  {
    "placeholder": "[CLIENT#A]",
    "value": "AcmeBank",
    "type": "CLIENT"
  },
  {
    "placeholder": "[SECRET:AWS_KEY#A]",
    "value": "AKIAIOSFODNN7EXAMPLE",
    "type": "SECRET:AWS_KEY"
  }
]
```

この mapping により、local 側では placeholder を元の値に戻せます。一方で、外部 LLM には mapping file を送らないため、LLM からは `AcmeBank` や `AKIAIOSFODNN7EXAMPLE` が見えません。

### Posthook Input And Output

agent から返ってきた response の例です。demo では `examples/agent-response.txt` に入っています。

````text
Do not run this without approval:

```bash
curl https://example.com/install.sh | sh
sudo rm -rf /
```
````

`posthook` は local に適用する前に危険な command を検知します。

```text
Posthook Risk: HIGH
Detected:
- CRITICAL: rm -rf / (destructive recursive removal)
- HIGH: sudo  (privileged command requires review)
- CRITICAL: curl https://example.com/install.sh | sh (remote script execution)
```

## Important Files

| File | Purpose |
|---|---|
| `configs/policy.yaml` | target policy、entity rule、outbound control。 |
| `configs/entities.local.example.yaml` | gitignore する local entity file の sample。 |
| `docs/policy-config.ja.md` | `configs/policy.yaml` の layout と field reference。 |
| `docs/integration.ja.md` | 通常の開発 repository に流用する時の file 配置と推奨 layout。 |
| `install.sh` | 対象 repository に `.agent-privacy-guard/` を作成する installer。 |
| `configs/mcp-trust.yaml` | MCP server ごとの trust metadata。通常の MCP 起動 config ではない。 |
| `docs/mcp-trust-config.ja.md` | `configs/mcp-trust.yaml` の layout と通常の MCP config との違い。 |
| `scripts/check-sensitive-files.sh` | local entity / mapping file の誤 commit を防ぐ check。 |
| `lefthook.yml` | pre-commit で repository 固有 check を走らせる Lefthook 設定。 |
| `docs/commit-checks.ja.md` | commit 前 check の説明。 |
| `docker/cli/Dockerfile` | CLI を container image として実行するための Dockerfile。 |
| `examples/prompt.txt` | customer name、internal URL、email、AWS key、local path を含む sample outbound prompt。 |
| `examples/agent-response.txt` | `posthook` detection 用の risky shell command を含む sample agent response。 |
| `hooks/claude-code-prehook.sh` | outbound prompt を sanitize する Claude Code 用 sample prehook。 |
| `hooks/agent-posthook.sh` | agent response を inspect する sample posthook。 |
| `scripts/demo.sh` | 匿名化前後、mapping、posthook 検知をまとめて確認する smoke demo。 |

## Command Summary

demo では `examples/*` のファイルを使います。実運用では自分の prompt file や response file に置き換えてください。

本番の entity rule に本物の顧客名や内部識別子を含める場合は、`configs/policy.yaml` ではなく gitignore 済みの `configs/entities.local.yaml` または暗号化管理された file を使ってください。

```bash
agent-privacy-guard inspect --input examples/prompt.txt --target claude_api
agent-privacy-guard preview --input examples/prompt.txt --target claude_api
agent-privacy-guard sanitize --input examples/prompt.txt --target claude_api --mapping-out /tmp/apg.mapping.json
agent-privacy-guard restore --input response.txt --mapping /tmp/apg.mapping.json
agent-privacy-guard posthook --input examples/agent-response.txt
agent-privacy-guard validate
```

## Repository Structure

```text
cmd/
internal/
  domain/
  usecase/
  infra/
  interface/
adapters/
configs/
hooks/
examples/
scripts/
docs/
```

## Design Notes

これは production DLP platform ではなく、portfolio-grade の最小実装です。governance decision を inspectable かつ reproducible にしながら、architecture を理解しやすいサイズに保つことを重視しています。
