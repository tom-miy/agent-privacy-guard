package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPolicyLoadsEntityFilesRelativeToPolicy(t *testing.T) {
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.yaml")
	entityPath := filepath.Join(dir, "entities.local.yaml")

	if err := os.WriteFile(policyPath, []byte(`
targets:
  claude_api:
    trust: public
    sanitize: strong
    allow: true
entity_files:
  - entities.local.yaml
entities:
  - type: CLIENT
    pattern: "\\b(DemoClient)\\b"
    scope: prompt
outbound:
  block_on_secret: true
  diff_only: true
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(entityPath, []byte(`
entities:
  - type: PRIVATE_CLIENT
    pattern: "\\b(PrivateClient)\\b"
    scope: prompt
`), 0o600); err != nil {
		t.Fatal(err)
	}

	policy, err := LoadPolicy(policyPath)
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, entity := range policy.Entities {
		if entity.Type == "PRIVATE_CLIENT" && entity.Pattern == "\\b(PrivateClient)\\b" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected entity from entity_files to be loaded, got %#v", policy.Entities)
	}
}

func TestLoadPolicyDoesNotMergeDefaultEntitiesWhenPolicyExists(t *testing.T) {
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.yaml")

	if err := os.WriteFile(policyPath, []byte(`
targets:
  claude_api:
    trust: public
    sanitize: strong
    allow: true
outbound:
  block_on_secret: true
  diff_only: true
`), 0o644); err != nil {
		t.Fatal(err)
	}

	policy, err := LoadPolicy(policyPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(policy.Entities) != 0 {
		t.Fatalf("expected no implicit demo entities, got %#v", policy.Entities)
	}
}

func TestValidatePolicyWarnsWhenEntityFilesAreMissing(t *testing.T) {
	policy, err := LoadPolicy("")
	if err != nil {
		t.Fatal(err)
	}
	policy.EntityFiles = nil
	problems := ValidatePolicy(policy)
	var found bool
	for _, problem := range problems {
		if problem == "policy should define entity_files, even if the referenced local entity file is empty" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected missing entity_files warning, got %#v", problems)
	}
}
