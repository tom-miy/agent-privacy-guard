package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tom-miy/agent-privacy-guard/internal/domain"
)

func TestLoadPolicy(t *testing.T) {
	tests := []struct {
		name    string
		content string
		missing bool
		wantErr bool
		assert  func(t *testing.T, policy domain.Policy)
	}{
		{
			name:    "falls back to default when file does not exist",
			missing: true,
			assert: func(t *testing.T, policy domain.Policy) {
				t.Helper()
				if _, ok := policy.Targets["claude_api"]; !ok {
					t.Fatalf("expected default policy target, got %#v", policy.Targets)
				}
				if len(policy.Entities) == 0 {
					t.Fatal("expected default entity rules")
				}
			},
		},
		{
			name: "overrides default fields from YAML",
			content: `
targets:
  claude_api:
    trust: public
    sanitize: weak
    allow: false
    mode: external_llm
entities:
  - type: TEAM
    pattern: "\\bCorePlatform\\b"
    scope: prompt
outbound:
  block_on_secret: false
  diff_only: false
`,
			assert: func(t *testing.T, policy domain.Policy) {
				t.Helper()
				target := policy.Targets["claude_api"]
				if target.Sanitize != domain.SanitizeWeak {
					t.Fatalf("sanitize: got %q, want %q", target.Sanitize, domain.SanitizeWeak)
				}
				if target.Allow {
					t.Fatal("expected YAML allow=false to override default")
				}
				if len(policy.Entities) != 1 || policy.Entities[0].Type != "TEAM" {
					t.Fatalf("expected YAML entities to override defaults, got %#v", policy.Entities)
				}
				if policy.Outbound.BlockOnSecret || policy.Outbound.DiffOnly {
					t.Fatalf("expected YAML outbound fields to override defaults, got %#v", policy.Outbound)
				}
			},
		},
		{
			name:    "rejects invalid YAML",
			content: "targets: [",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "policy.yaml")
			if tt.missing {
				path = filepath.Join(t.TempDir(), "missing-policy.yaml")
			} else if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}

			policy, err := LoadPolicy(path)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tt.assert != nil {
				tt.assert(t, policy)
			}
		})
	}
}

func TestValidatePolicyReportsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name         string
		policy       domain.Policy
		wantProblems []string
	}{
		{
			name: "missing target and entity fields",
			policy: domain.Policy{
				Targets: map[string]domain.TargetPolicy{
					"bad": {},
				},
				Entities: []domain.EntityRule{
					{Type: "CLIENT"},
				},
			},
			wantProblems: []string{
				"target bad is missing trust",
				"target bad is missing sanitize",
				"entity rules require type and pattern",
			},
		},
		{
			name:         "missing targets",
			policy:       domain.Policy{},
			wantProblems: []string{"policy must define at least one target"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := ValidatePolicy(tt.policy)
			for _, want := range tt.wantProblems {
				if !containsProblem(problems, want) {
					t.Fatalf("expected problem %q in %#v", want, problems)
				}
			}
		})
	}
}

func containsProblem(problems []string, want string) bool {
	for _, problem := range problems {
		if strings.Contains(problem, want) {
			return true
		}
	}
	return false
}
