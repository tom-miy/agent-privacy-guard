# Summary

`agent-privacy-guard` は Claude Code、Cursor、GitHub Copilot、Codex CLI、MCP server、LLM API のための安全な AI-assisted development gateway を実演します。

重点は、policy-driven prompt sanitization、structured reversible placeholder、MCP-aware routing、prehook/posthook processing、reproducible CLI demo です。

この実装は大きな "AI magic" を避けています。rule は YAML で見える形にし、mapping は local JSON に保持し、adapter directory は agent-specific integration の置き場所を明確にします。
