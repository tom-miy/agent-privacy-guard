# Claude Code Adapter

outbound model call の前に `hooks/claude-code-prehook.sh` を使用し、response の後に `hooks/agent-posthook.sh` を使用します。

この adapter は Claude-specific wiring を shared policy engine の外側に保ちます。
