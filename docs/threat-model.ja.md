# Threat Model

## Assets

- customer name と business identifier。
- cloud credential と token。
- internal URL と MCP-derived data。
- repository path、diff、stack trace、generated patch。

## Trust Boundaries

- internal MCP と local LLM target は、より弱い sanitization で context を受け取れる。
- public LLM と external MCP target には、より強い sanitization を要求する。
- placeholder mapping は local に保持し、external model には送信しない。

## Primary Risks

- raw secret が prompt に含まれる。
- stack trace や diff から customer identifier や environment identifier が漏れる。
- internal MCP server の data が public model に越境する。
- agent response が destructive command を提案する。
- generated patch が forbidden file や security-sensitive file を変更する。

## Mitigations

- `[CLIENT#A]` や `[SECRET:AWS_KEY#A]` のような type-aware placeholder replacement。
- trust と sanitization strength に基づく target-specific policy。
- diff-only context recommendation。
- posthook command inspection。
- inspect、sanitize、posthook event の JSONL audit trail。
