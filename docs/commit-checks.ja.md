# Commit Checks

この repository 固有の commit 前チェックです。

目的は、`agent-privacy-guard` の運用で特に危険な local file を誤って git commit しないことです。

## Sensitive File Check

```bash
scripts/check-sensitive-files.sh
```

staged file に次が含まれていたら失敗します。

```text
.agent-privacy-guard/entities.local.yaml
.agent-privacy-guard/mapping.json
.agent-privacy-guard.mapping.json
*.mapping.json
```

理由:

- `entities.local.yaml` は本物の顧客名や内部識別子を含み得る。
- `mapping.json` は placeholder から元の値へ戻す reversible mapping を含み得る。

これらは git 管理せず、必要なら SOPS / age / git-crypt などで暗号化した source を別管理してください。

## Pre-commit Script

```bash
scripts/pre-commit.sh
```

実行内容:

```text
scripts/check-sensitive-files.sh
go test ./...
go run ./cmd/agent-privacy-guard validate
```

## Optional Git Hook Setup

local git hook として使う場合:

```bash
ln -sf ../../scripts/pre-commit.sh .git/hooks/pre-commit
```

## Lefthook Setup

Lefthook を使う場合は、この repository の `lefthook.yml` を使えます。

```bash
lefthook install
```

設定内容:

```yaml
pre-commit:
  parallel: false
  commands:
    sensitive-files:
      run: scripts/check-sensitive-files.sh
    tests:
      run: go test ./...
    validate:
      run: go run ./cmd/agent-privacy-guard validate
```

commit 前に、sensitive local file の誤 commit 防止、Go test、policy validation を実行します。

## Adding To An Existing Lefthook Config

既に `lefthook.yml` がある repository では、上書きせずに `pre-commit.commands` へ追加してください。

```yaml
pre-commit:
  commands:
    agent-privacy-guard-sensitive-files:
      run: |
        blocked="$(git diff --cached --name-only --diff-filter=ACMR | grep -E '(^|/)\.agent-privacy-guard/(entities\.local\.yaml|mapping\.json)$|(^|/).*\.mapping\.json$' || true)"
        if [ -n "$blocked" ]; then
          echo "Blocked sensitive agent-privacy-guard files:"
          echo "$blocked"
          exit 1
        fi
```

`install.sh` は既存の `lefthook.yml` を自動編集しません。各 repository の hook 方針を壊さないためです。代わりに install 後の message で Lefthook への追加を案内します。

この hook は repository local な安全策です。gitleaks、shellcheck、markdownlint などの汎用 checks は別の shared hook repository や CI で管理する想定です。
