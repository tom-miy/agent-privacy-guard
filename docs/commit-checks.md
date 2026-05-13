# Commit Checks

These are repository-specific pre-commit checks for `agent-privacy-guard`.

The goal is to avoid accidentally committing local files that are especially sensitive in this gateway workflow.

## Sensitive File Check

```bash
scripts/check-sensitive-files.sh
```

The script fails if staged files include:

```text
.agent-privacy-guard/entities.local.yaml
.agent-privacy-guard/mapping.json
.agent-privacy-guard.mapping.json
*.mapping.json
```

Why:

- `entities.local.yaml` may contain real customer names or internal identifiers.
- `mapping.json` may contain reversible placeholder mappings back to raw values.

Keep these files untracked. If needed, store encrypted source files with SOPS / age / git-crypt and decrypt them only at runtime.

## Pre-commit Script

```bash
scripts/pre-commit.sh
```

It runs:

```text
scripts/check-sensitive-files.sh
go test ./...
go run ./cmd/agent-privacy-guard validate
```

## Optional Git Hook Setup

To use it as a local git hook:

```bash
ln -sf ../../scripts/pre-commit.sh .git/hooks/pre-commit
```

## Lefthook Setup

If you use Lefthook, this repository includes `lefthook.yml`.

```bash
lefthook install
```

Configuration:

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

Before commit, it blocks sensitive local files, runs Go tests, and validates policy/config files.

## Adding To An Existing Lefthook Config

If a repository already has `lefthook.yml`, do not overwrite it. Add a command under `pre-commit.commands`.

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

`install.sh` does not edit an existing `lefthook.yml` automatically, because hook policy is repository-specific. Instead, the installer prints a reminder to add this check if the target repository uses Lefthook.

This is a repository-local safety check. Generic checks such as gitleaks, shellcheck, and markdownlint are better managed by a shared hook repository or CI.
