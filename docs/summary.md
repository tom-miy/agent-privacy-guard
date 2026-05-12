# Summary

`agent-privacy-guard` demonstrates a safe AI-assisted development gateway for Claude Code, Cursor, GitHub Copilot, Codex CLI, MCP servers, and LLM APIs.

It focuses on policy-driven prompt sanitization, structured reversible placeholders, MCP-aware routing, prehook/posthook processing, and reproducible CLI demos.

The implementation avoids broad "AI magic": rules are visible in YAML, mappings are local JSON, and adapter directories document where agent-specific integration belongs.
