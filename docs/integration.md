# Integration Guide

This document explains how to use `agent-privacy-guard` from a normal application repository.

This repository is the CLI / gateway source repository. It is not a template that should be copied wholesale into application repositories.

`agent-privacy-guard` is normally applied automatically.

Developers are not expected to run `inspect` or `sanitize` by hand for every prompt. Instead, the CLI is wired into Claude Code, Cursor, Codex CLI, CI, or a custom wrapper.

## How It Runs

The typical flow is:

```text
developer writes prompt / agent collects context
  -> prehook or wrapper calls agent-privacy-guard sanitize
  -> sanitized prompt is sent to external LLM
  -> posthook calls agent-privacy-guard posthook
  -> response is reviewed or applied locally
```

So `sanitize` and `posthook` are intended to run from agent hooks, not as manual daily commands.

## What To Copy Into A Normal Development Repository

You do not need to copy this whole repository into a normal application repository.

Install the `agent-privacy-guard` CLI separately. In the target repository, keep only policy / hook files.

Instead of copying files manually, use `install.sh` from this repository:

```bash
./install.sh --target /path/to/your-app
```

If `.agent-privacy-guard/` already exists, files are not overwritten. Use `--force` only when you want to replace them.

```bash
./install.sh --target /path/to/your-app --force
```

This creates:

```text
.agent-privacy-guard/
  policy.yaml
  mcp-trust.yaml
  entities.local.example.yaml
  .gitignore
  hooks/
    prehook.sh
    posthook.sh
```

If reusing samples from this repository, the minimum useful set is:

```text
templates/agent-privacy-guard/policy.yaml
hooks/claude-code-prehook.sh
hooks/agent-posthook.sh
```

Optionally add:

```text
templates/agent-privacy-guard/mcp-trust.yaml
AGENTS.md
CLAUDE.md
.cursorrules
.codex/config.toml
```

`templates/agent-privacy-guard/mcp-trust.yaml` is not a standard MCP server launch config. It is trust metadata per MCP server. See [mcp-trust-config.md](mcp-trust-config.md) for details.

## What Not To Copy

These are implementation or demo files and are not needed in a normal application repository:

```text
cmd/
internal/
adapters/
docs/
examples/
scripts/
go.mod
go.sum
docker/cli/Dockerfile
mise.toml
```

Again, use `agent-privacy-guard` as an installed CLI. Keep only policy and hook files in the target repository.

## Recommended Layout In A Target Repository

To keep gateway configuration separate from application code, use a dedicated directory:

```text
your-app/
  src/
  package.json
  README.md

  .agent-privacy-guard/
    policy.yaml
    mcp-trust.yaml
    entities.local.example.yaml
    .gitignore
    hooks/
      prehook.sh
      posthook.sh
```

Mapping from this reference repository:

```text
templates/agent-privacy-guard/policy.yaml                 -> .agent-privacy-guard/policy.yaml
templates/agent-privacy-guard/mcp-trust.yaml              -> .agent-privacy-guard/mcp-trust.yaml
templates/agent-privacy-guard/entities.local.example.yaml -> .agent-privacy-guard/entities.local.example.yaml
hooks/claude-code-prehook.sh                              -> .agent-privacy-guard/hooks/prehook.sh
hooks/agent-posthook.sh                                   -> .agent-privacy-guard/hooks/posthook.sh
```

Update the hook `--policy` path accordingly:

```bash
agent-privacy-guard sanitize \
  --target claude_api \
  --policy .agent-privacy-guard/policy.yaml \
  --mapping-out .agent-privacy-guard/mapping.json
```

## Why This Repository Looks Larger

This repository is a portfolio / reference implementation, so it contains both product code and examples.

| Area | Purpose |
|---|---|
| `cmd/`, `internal/` | CLI implementation. |
| `templates/agent-privacy-guard/`, `hooks/`, `examples/`, `docs/` | Usage examples, policy samples, and integration samples. |

A normal application repository only needs a small subset of the second group.

## Mental Model

```text
agent-privacy-guard repository
  = CLI product source code + examples

your development repository
  = your app source code + small .agent-privacy-guard/ policy folder
```

In real use, keep policy and hook files inside a dedicated directory such as `.agent-privacy-guard/` so they are clearly separate from the main application implementation.
