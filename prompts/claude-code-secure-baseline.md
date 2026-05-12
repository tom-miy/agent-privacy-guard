# claude-code-secure-baseline

## Goal

Claude Code を安全に使うための **Claude Code native security baseline repository** を作成する。

この repository は `agent-privacy-guard` とは別物として扱う。

- `agent-privacy-guard`
  - prompt sanitization
  - outbound gateway
  - structured placeholder mapping
  - response posthook inspection
  - multi-agent policy enforcement

- `claude-code-secure-baseline`
  - Claude Code sandbox
  - Claude Code permissions
  - deny rules
  - PreToolUse hooks
  - network / filesystem restrictions
  - Managed Settings
  - devcontainer isolation

つまり、この repository は **Claude Code 自体の実行権限を制御する baseline** であり、prompt 匿名化 gateway ではない。

---

## Classification

type: portfolio

---

## Stack

- JSON
- Shell Script
- jq
- Markdown
- devcontainer

Optional:

- mise
- GitHub Actions for lint

---

## Repository Role

この repository は、アプリ開発 repository にそのまま丸ごとコピーするものではない。

Claude Code hardening 用の reference repository として、以下を提供する。

- `.claude/settings.json` の example
- Managed Settings の example
- PreToolUse hook scripts
- devcontainer sample
- install script
- docs

通常のアプリ repository に適用する場合は、必要な設定だけをコピーまたは `install.sh` で配置する。

---

## Core Concepts

### Separate Layers

Claude Code native hardening と gateway policy を混ぜない。

```text
Claude Code hardening
  -> Claude Code が実行できる tool / command / file / network を制限する

agent-privacy-guard
  -> 外部 LLM に送る prompt を sanitize し、response を posthook で検査する
```

対象 repository では次のように分ける。

```text
your-app/
  .claude/
    settings.json
    hooks/
      validate-command.sh

  .agent-privacy-guard/
    policy.yaml
    hooks/
      prehook.sh
      posthook.sh
```

---

## Features

### Feature 1: Sandbox Baseline

`.claude/settings.json` に sandbox 設定例を用意する。

```json
{
  "sandbox": {
    "enabled": true,
    "allowUnsandboxedCommands": false
  }
}
```

macOS / Linux の差分や注意点は docs に書く。

---

### Feature 2: Dangerous Command Deny Rules

`permissions.deny` の baseline を用意する。

対象例:

- `rm -rf`
- `curl`
- `wget`
- `git push`
- `git push --force`
- `chmod 777`
- production 接続を示す command

---

### Feature 3: Secret File Access Deny Rules

`.env`、secret、credential、private key への read access を拒否する例を用意する。

対象例:

- `.env`
- `.env.*`
- `secrets/**`
- `config/credentials.json`
- `**/*.pem`
- `**/*.key`
- `~/.aws/credentials`
- `~/.ssh`

---

### Feature 4: Network Allowlist

Claude Code sandbox network allowlist の例を用意する。

許可例:

- `github.com`
- `*.githubusercontent.com`
- `*.npmjs.org`
- `registry.yarnpkg.com`
- `pypi.org`

deny-by-default の考え方を docs に書く。

---

### Feature 5: Disable Bypass Permissions

`--dangerously-skip-permissions` 相当の bypass mode を無効化する設定例を置く。

```json
{
  "permissions": {
    "disableBypassPermissionsMode": "disable"
  }
}
```

Managed Settings に置くとユーザー側で上書きしにくい、という運用メモを書く。

---

### Feature 6: PreToolUse Hook

`.claude/hooks/validate-command.sh` の sample を用意する。

要件:

- stdin の JSON から `.tool_input.command` を読む
- `jq` を使う
- dangerous command を検出したら stderr に理由を出す
- Claude Code が block と解釈できる exit code を使う
- shellcheck しやすい shell script にする

検出対象例:

- `rm -rf`
- `curl ... | sh`
- `wget ... | sh`
- `git push --force`
- `chmod 777`
- `prod`

---

### Feature 7: Managed Settings

組織向けに Managed Settings の sample を用意する。

含める設定例:

- `disableBypassPermissionsMode`
- `allowManagedPermissionRulesOnly`
- `allowManagedHooksOnly`
- `allowManagedMcpServersOnly`
- `allowManagedDomainsOnly`
- managed network allowlist

配置先の docs:

```text
macOS: /Library/Application Support/ClaudeCode/managed-settings.json
Linux: /etc/claude-code/managed-settings.json
Windows: C:\Program Files\ClaudeCode\managed-settings.json
```

---

### Feature 8: Devcontainer Isolation

Claude Code を isolated devcontainer で動かすための sample を用意する。

最低限:

```text
.devcontainer/
  devcontainer.json
  Dockerfile
```

目的:

- host machine からの隔離
- workspace boundary の明確化
- network allowlist / firewall の補助

production-grade firewall までは作らない。portfolio baseline として意図的に小さくする。

---

### Feature 9: Installer

対象 repository に Claude Code hardening files を配置する `install.sh` を作る。

例:

```bash
./install.sh --target /path/to/your-app
```

作成する layout:

```text
your-app/
  .claude/
    settings.json
    hooks/
      validate-command.sh
```

既存 file がある場合はデフォルトで上書きしない。

```bash
./install.sh --target /path/to/your-app --force
```

---

### Feature 10: Documentation

README と docs は、次の誤解を避けるように書く。

- この repository は Claude Code 専用 hardening baseline
- prompt sanitization gateway ではない
- `agent-privacy-guard` とは別レイヤ
- アプリ repository に丸ごとコピーしない
- 必要な `.claude/` 設定だけ install / copy する

---

## Non-Goals

- prompt sanitization
- LLM gateway
- outbound prompt anonymization
- reversible placeholder mapping
- MCP trust routing
- multi-agent gateway
- SaaS platform
- enterprise 完全版 EDR / DLP
- Claude Code 公式設定 schema の完全追従

---

## Repository Structure

```text
README.md
README.ja.md
install.sh

claude/
  settings.example.json
  managed-settings.example.json
  hooks/
    validate-command.sh

devcontainer/
  devcontainer.json
  Dockerfile

docs/
  architecture.ja.md
  settings.ja.md
  hooks.ja.md
  managed-settings.ja.md
  devcontainer.ja.md
  integration-with-agent-privacy-guard.ja.md

examples/
  unsafe-tool-input.json
  safe-tool-input.json

scripts/
  demo.sh
  lint.sh
```

---

## Demonstration Scenario

以下を demo で再現する。

1. unsafe command を含む Claude Code tool input JSON を用意する
2. `validate-command.sh` に渡す
3. `rm -rf` / `curl ... | sh` が block される
4. safe command は allow される
5. `.claude/settings.json` の deny rules と hook の役割を README で説明する

---

## README Requirements

README 冒頭で明確に書く。

```text
This repository is Claude Code native hardening baseline.
It is not agent-privacy-guard.
It does not sanitize prompts.
It restricts what Claude Code can do locally.
```

日本語版では次を明記する。

```text
この repository は Claude Code 自体の sandbox / permissions / hooks の baseline です。
prompt 匿名化や outbound gateway は agent-privacy-guard の責務です。
```

---

## Security Notes

- 設定 schema は Claude Code 側の変更に追従が必要。
- 実運用前に Claude Code 公式 docs と `/status` / `/permissions` で確認する。
- deny rule は allow rule より優先される前提で説明する。
- Managed Settings は組織強制用として扱う。
- devcontainer は追加防御であり、policy の代替ではない。

---

## Output

- directory structure
- minimal implementation
- sample Claude Code settings
- sample Managed Settings
- PreToolUse hook script
- install script
- README
- README.ja.md
- docs
- demo scripts

---

## Evaluation Focus

重視すること:

- Claude Code hardening と gateway policy の責務分離
- DevEx
- install しやすさ
- settings の読みやすさ
- safety baseline としての実用性
- intentional scope control

避けること:

- `agent-privacy-guard` の再実装
- prompt sanitization の混入
- Claude Code 以外の agent 設定を無理に含めること
- 過剰な enterprise security platform 化
