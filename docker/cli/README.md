# CLI Docker Image

This Dockerfile builds and runs the `agent-privacy-guard` CLI as a container image.

Use cases:

- Run `agent-privacy-guard validate` in CI.
- Try the CLI without installing Go on the host.
- Mount policy / prompt files and run inspect / sanitize.

This is not a Claude Code sandbox or devcontainer.

## Build

Run from the repository root:

```bash
docker build -f docker/cli/Dockerfile -t agent-privacy-guard:local .
```

## Run

```bash
docker run --rm \
  -v "$PWD:/workspace" \
  -w /workspace \
  agent-privacy-guard:local validate
```

Example:

```bash
docker run --rm \
  -v "$PWD:/workspace" \
  -w /workspace \
  agent-privacy-guard:local inspect \
  --input examples/prompt.txt \
  --target claude_api
```
