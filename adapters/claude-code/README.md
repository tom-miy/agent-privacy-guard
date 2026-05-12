# Claude Code Adapter

Use `hooks/claude-code-prehook.sh` before outbound model calls and `hooks/agent-posthook.sh` after responses.

The adapter keeps Claude-specific wiring outside the shared policy engine.
