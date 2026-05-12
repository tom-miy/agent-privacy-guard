# Policy Config

`configs/policy.yaml` は `agent-privacy-guard` の中心設定です。

この file では次を定義します。

- どの target に送るか。
- target ごとの trust level と sanitization strength。
- project-specific な entity anonymization rule。
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

## Adding A New Customer Name

`NewCustomer` を匿名化したい場合は、`CLIENT` rule の `pattern` に追加します。

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
