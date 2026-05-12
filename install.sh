#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  ./install.sh [--target PATH] [--force]

Installs agent-privacy-guard policy and hook files into a target repository.

This does not copy the CLI source code. It creates:
  .agent-privacy-guard/policy.yaml
  .agent-privacy-guard/hooks/prehook.sh
  .agent-privacy-guard/hooks/posthook.sh

Options:
  --target PATH   Target repository path. Defaults to current directory.
  --force         Overwrite existing .agent-privacy-guard files.
  -h, --help      Show this help.
USAGE
}

target_dir="."
force="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --target)
      if [[ $# -lt 2 ]]; then
        echo "error: --target requires a path" >&2
        exit 2
      fi
      target_dir="$2"
      shift 2
      ;;
    --force)
      force="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "error: unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
target_dir="$(cd "$target_dir" && pwd)"

source_policy="$script_dir/configs/policy.yaml"
install_dir="$target_dir/.agent-privacy-guard"
hooks_dir="$install_dir/hooks"
policy_path="$install_dir/policy.yaml"
entity_example_path="$install_dir/entities.local.example.yaml"
gitignore_path="$install_dir/.gitignore"
prehook_path="$hooks_dir/prehook.sh"
posthook_path="$hooks_dir/posthook.sh"
mapping_path="$install_dir/mapping.json"

if [[ ! -f "$source_policy" ]]; then
  echo "error: source policy not found: $source_policy" >&2
  exit 1
fi

if [[ "$force" != "true" ]]; then
  for path in "$policy_path" "$entity_example_path" "$gitignore_path" "$prehook_path" "$posthook_path"; do
    if [[ -e "$path" ]]; then
      echo "error: $path already exists. Re-run with --force to overwrite." >&2
      exit 1
    fi
  done
fi

mkdir -p "$hooks_dir"
install -m 0644 "$source_policy" "$policy_path"

cat > "$entity_example_path" <<'ENTITIES'
# Copy this file to entities.local.yaml for real project-specific entity rules.
# Do not commit entities.local.yaml if it contains real customer names,
# internal service names, database names, or other sensitive identifiers.

entities:
  - type: CLIENT
    pattern: "\\b(ExampleCustomer|AnotherCustomer)\\b"
    scope: prompt
ENTITIES

cat > "$gitignore_path" <<'GITIGNORE'
entities.local.yaml
mapping.json
GITIGNORE

cat > "$prehook_path" <<'PREHOOK'
#!/usr/bin/env bash
set -euo pipefail

hook_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
install_dir="$(cd "$hook_dir/.." && pwd)"

# Reads raw outbound prompt text from stdin.
# Writes sanitized prompt text to stdout.
# Keeps reversible placeholder mapping local.
agent-privacy-guard sanitize \
  --target "${AGENT_TARGET:-claude_api}" \
  --policy "$install_dir/policy.yaml" \
  --mapping-out "$install_dir/mapping.json"
PREHOOK

cat > "$posthook_path" <<'POSTHOOK'
#!/usr/bin/env bash
set -euo pipefail

# Reads an agent response from stdin and reports dangerous commands or patches.
agent-privacy-guard posthook --target "${AGENT_TARGET:-claude_api}"
POSTHOOK

chmod +x "$prehook_path" "$posthook_path"

cat <<SUMMARY
Installed agent-privacy-guard integration files:
  $policy_path
  $entity_example_path
  $gitignore_path
  $prehook_path
  $posthook_path

Local restore mappings will be written to:
  $mapping_path

Next steps:
  1. Edit $policy_path for this repository's customer names and internal identifiers.
  2. For real sensitive names, copy $entity_example_path to entities.local.yaml and enable entity_files in policy.yaml.
  3. Wire $prehook_path into your agent's outbound prompt hook.
  4. Wire $posthook_path into your agent's response hook.

Smoke test:
  echo 'AcmeBank token=example-secret-value-1234' | $prehook_path
SUMMARY
