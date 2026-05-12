# Threat Model

## Assets

- Customer names and business identifiers.
- Cloud credentials and tokens.
- Internal URLs and MCP-derived data.
- Repository paths, diffs, stack traces, and generated patches.

## Trust Boundaries

- Internal MCP and local LLM targets can receive less sanitized context.
- Public LLM and external MCP targets require stronger sanitization.
- Placeholder mappings stay local and should not be sent to external models.

## Primary Risks

- Raw secrets included in prompts.
- Customer or environment identifiers leaked through stack traces and diffs.
- MCP data crossing from an internal server to a public model.
- Agent responses suggesting destructive commands.
- Generated patches modifying forbidden or security-sensitive files.

## Mitigations

- Type-aware placeholder replacement such as `[CLIENT#A]` and `[SECRET:AWS_KEY#A]`.
- Target-specific policy for trust and sanitization strength.
- Diff-only context recommendation.
- Posthook command inspection.
- JSONL audit trail for inspect, sanitize, and posthook events.
