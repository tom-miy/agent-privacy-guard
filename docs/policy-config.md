# Policy Config

`configs/policy.yaml` is the central configuration file for `agent-privacy-guard`.

It defines:

- Which target the prompt is going to.
- Trust level and sanitization strength per target.
- Project-specific entity anonymization rules.
- Loading untracked local entity files.
- Whether outbound send should be blocked when secrets are detected.
- Whether diff-only context is recommended.

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

# Optional. Use this for real customer names or internal identifiers.
# Paths are relative to this policy file.
entity_files:
  - entities.local.yaml

outbound:
  block_on_secret: true
  diff_only: true
```

## `targets`

`targets` configures policy per destination. CLI usage such as `--target claude_api` reads `targets.claude_api`.

| Field | Example | Meaning |
|---|---|---|
| `trust` | `public` | The target trust boundary. |
| `sanitize` | `strong` | Anonymization strength. |
| `allow` | `true` | Whether policy allows sending to this target. |
| `mode` | `external_llm` | Metadata describing the target kind. |

## Trust Levels

| Value | Use case |
|---|---|
| `public` | Claude API, Cursor, Copilot, external MCP. |
| `internal` | Internal MCP, local services. |
| `confidential` | Internal targets that can handle confidential data. |
| `secret` | Most restrictive targets for secret-class data. |

## Sanitization Levels

| Value | Behavior |
|---|---|
| `none` | Do not sanitize. Intended for internal MCP-like targets. |
| `weak` | Apply minimal sanitization, mainly built-in secret detectors. |
| `strong` | Apply built-in secret detectors and `entities` rules. |

## `entities`

`entities` turns customer names, database names, and other project-specific identifiers into structured placeholders.

Important: real customer names, internal system names, and database names may themselves be sensitive. Do not put them directly in a public repository or ordinary git history.

The committed `entities` in this repository are fake demo values. For production, prefer one of these:

- Load a gitignored local file through `entity_files`.
- Store encrypted rules with SOPS / age / git-crypt and decrypt them before runtime.

```yaml
entities:
  - type: CLIENT
    pattern: "\\b(AcmeBank|ExampleCorp|MegaRetail)\\b"
    scope: prompt
  - type: POSTGRES_DB
    pattern: "\\b[a-z0-9-]*db[a-z0-9-]*\\b"
    scope: prompt
```

This produces replacements such as:

| Input | Placeholder |
|---|---|
| `AcmeBank` | `[CLIENT#A]` |
| `prod-db-tokyo` | `[POSTGRES_DB#A]` |

`type` becomes the placeholder prefix. For example, `type: CLIENT` produces `[CLIENT#A]`, `[CLIENT#B]`, and so on.

## `entity_files`

`entity_files` loads additional entity rules from separate YAML files.

```yaml
entity_files:
  - entities.local.yaml
```

Paths are resolved relative to `configs/policy.yaml`. The example above loads `configs/entities.local.yaml`.

Local file layout:

```yaml
entities:
  - type: CLIENT
    pattern: "\\b(RealCustomerName|AnotherPrivateClient)\\b"
    scope: prompt
```

`.gitignore` ignores `configs/entities.local.yaml`. A safe sample is available at [../configs/entities.local.example.yaml](../configs/entities.local.example.yaml).

If you want encrypted storage, keep the encrypted file in git and decrypt it to a temporary local file referenced by `entity_files`. This minimal implementation does not perform encryption or decryption.

## Adding A New Customer Name

For fake demo names, add them to the `CLIENT` rule.

```yaml
entities:
  - type: CLIENT
    pattern: "\\b(AcmeBank|ExampleCorp|MegaRetail|NewCustomer)\\b"
    scope: prompt
```

Then verify with:

```bash
agent-privacy-guard preview --input examples/prompt.txt --target claude_api
```

For real customer names, add them to `configs/entities.local.yaml` instead of `configs/policy.yaml`.

## Built-in Secret Detectors

AWS keys, AWS ARNs, emails, internal URLs, tokens, and SSH private keys are detected by built-in detectors, even if they are not listed in `entities`.

| Input kind | Placeholder example |
|---|---|
| AWS key | `[SECRET:AWS_KEY#A]` |
| AWS ARN | `[SECRET:AWS_ARN#A]` |
| email | `[SECRET:EMAIL#A]` |
| internal URL | `[SECRET:INTERNAL_URL#A]` |
| token / API key | `[SECRET:TOKEN#A]` |
| SSH private key | `[SECRET:SSH_KEY#A]` |

Use `entities` for project-specific names. Built-in detectors handle generic secrets.

## `outbound`

```yaml
outbound:
  block_on_secret: true
  diff_only: true
```

| Field | Meaning |
|---|---|
| `block_on_secret` | Set `outbound_ok: false` when a high / critical secret is found for a public target. |
| `diff_only` | Recommend sending diff-only context instead of the whole repository. |

`inspect --fail-on-block` exits non-zero when `outbound_ok: false`.
