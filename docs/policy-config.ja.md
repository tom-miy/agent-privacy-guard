# Policy Config

`configs/policy.yaml` は `agent-privacy-guard` の中心設定です。

この file では次を定義します。

- どの target に送るか。
- target ごとの trust level と sanitization strength。
- project-specific な entity anonymization rule。
- untracked local entity file の読み込み。
- secret が見つかった場合に outbound を止めるか。
- diff-only context を推奨するか。

## Layout

```yaml
targets:
  claude_api:
    trust: public
    sanitize: strong
    allow: true
    mode: external_llm

entities:
  - type: CLIENT
    pattern: "\\b(AcmeBank|ExampleCorp|MegaRetail)\\b"
    scope: prompt

# Required by the baseline. Create the referenced file explicitly when applying
# this policy to a real repository.
# Paths are relative to this policy file.
entity_files:
  - entities.local.yaml

outbound:
  block_on_secret: true
  diff_only: true
```

## `targets`

`targets` は送信先ごとの policy です。CLI の `--target claude_api` は、この `targets.claude_api` を参照します。

| Field | Example | Meaning |
|---|---|---|
| `trust` | `public` | target の trust boundary。 |
| `sanitize` | `strong` | anonymization の強度。 |
| `allow` | `true` | target への送信を policy 上許可するか。 |
| `mode` | `external_llm` | target の種類を説明する metadata。 |

## Trust Levels

| Value | Use case |
|---|---|
| `public` | Claude API、Cursor、Copilot、外部 MCP など。 |
| `internal` | internal MCP、local service など。 |
| `confidential` | confidential data を扱う内部 target。 |
| `secret` | secret class data を扱う最も制限の強い target。 |

## Sanitization Levels

| Value | Behavior |
|---|---|
| `none` | sanitization しない。internal MCP など向け。 |
| `weak` | built-in secret detector を中心に最小限 sanitization する。 |
| `strong` | secret detector に加えて `entities` rule も適用する。 |

## `entities`

`entities` は顧客名、DB 名、project-specific な識別子を structured placeholder に変換するための rule です。

重要: 本物の顧客名、内部システム名、DB 名などは、それ自体が外部に出したくない情報です。公開 repository や通常の git 管理にそのまま入れないでください。

この repository の `entities` は demo 用の fake 値です。この baseline では `entity_files` を必ず設定しますが、参照先 file は自動生成しません。空 file があると「設定済み」と誤解しやすいため、対象 repository で明示的に作成してください。本番では次のどちらかを推奨します。

- `entity_files` で gitignore された local file から読む。
- SOPS / age / git-crypt などで暗号化し、実行時に復号した file を `entity_files` で読む。

```yaml
entities:
  - type: CLIENT
    pattern: "\\b(AcmeBank|ExampleCorp|MegaRetail)\\b"
    scope: prompt
  - type: POSTGRES_DB
    pattern: "\\b[a-z0-9-]*db[a-z0-9-]*\\b"
    scope: prompt
```

この設定では次のように置換されます。

| Input | Placeholder |
|---|---|
| `AcmeBank` | `[CLIENT#A]` |
| `prod-db-tokyo` | `[POSTGRES_DB#A]` |

`type` が placeholder prefix になります。たとえば `type: CLIENT` は `[CLIENT#A]`、`[CLIENT#B]` のように出力されます。

## `entity_files`

`entity_files` は、追加の entity rule を別 YAML から読み込みます。この baseline では、file が存在しない場合は load / validate で失敗します。

```yaml
entity_files:
  - entities.local.yaml
```

path は `configs/policy.yaml` からの相対 path として解決されます。上の例では `configs/entities.local.yaml` を読みます。

この reference repository の `configs/policy.yaml` は validate を通すために `entities.local.example.yaml` を参照しています。`install.sh` で対象 repository に適用すると `entities.local.yaml` 参照へ書き換えられます。

local file の layout:

```yaml
entities:
  - type: CLIENT
    pattern: "\\b(RealCustomerName|AnotherPrivateClient)\\b"
    scope: prompt
```

通常の application repository では、installer が `.agent-privacy-guard/entities.local.example.yaml` と `.agent-privacy-guard/.gitignore` を作成します。必要になった時点で example をコピーするか、暗号化 source から `.agent-privacy-guard/entities.local.yaml` を生成してください。sample として [../configs/entities.local.example.yaml](../configs/entities.local.example.yaml) も用意しています。

暗号化して管理したい場合は、暗号化済み file を git 管理し、復号した一時 file を `entity_files` で参照する運用にしてください。この minimal implementation は暗号化 / 復号自体は行いません。

## Adding A New Customer Name

demo 用の fake name であれば `CLIENT` rule の `pattern` に追加できます。

```yaml
entities:
  - type: CLIENT
    pattern: "\\b(AcmeBank|ExampleCorp|MegaRetail|NewCustomer)\\b"
    scope: prompt
```

その後、次の command で確認します。

```bash
agent-privacy-guard preview --input examples/prompt.txt --target claude_api
```

本物の顧客名の場合は、`configs/policy.yaml` ではなく `configs/entities.local.yaml` に追加してください。

## Built-in Secret Detectors

AWS key、AWS ARN、email、internal URL、token、SSH private key は `entities` に書かなくても built-in detector で検出します。

| Input kind | Placeholder example |
|---|---|
| AWS key | `[SECRET:AWS_KEY#A]` |
| AWS ARN | `[SECRET:AWS_ARN#A]` |
| email | `[SECRET:EMAIL#A]` |
| internal URL | `[SECRET:INTERNAL_URL#A]` |
| token / API key | `[SECRET:TOKEN#A]` |
| SSH private key | `[SECRET:SSH_KEY#A]` |

`entities` は project-specific な名前を扱うための設定、built-in detector は汎用 secret を扱うための実装です。

## `outbound`

```yaml
outbound:
  block_on_secret: true
  diff_only: true
```

| Field | Meaning |
|---|---|
| `block_on_secret` | public target に high / critical secret が含まれる場合、`outbound_ok: false` にする。 |
| `diff_only` | repository 全体ではなく diff-only context を推奨する。 |

`inspect --fail-on-block` は `outbound_ok: false` のとき non-zero exit になります。
