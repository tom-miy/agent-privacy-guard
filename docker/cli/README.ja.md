# CLI Docker Image

この Dockerfile は `agent-privacy-guard` CLI を container image として build / run するためのものです。

用途:

- CI で `agent-privacy-guard validate` を実行する。
- host に Go toolchain を入れずに CLI を試す。
- policy / prompt file を mount して inspect / sanitize を実行する。

これは Claude Code の sandbox や devcontainer ではありません。

## Build

repository root から実行します。

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

例:

```bash
docker run --rm \
  -v "$PWD:/workspace" \
  -w /workspace \
  agent-privacy-guard:local inspect \
  --input examples/prompt.txt \
  --target claude_api
```
