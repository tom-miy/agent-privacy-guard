package usecase

import (
	"strings"
	"testing"

	"github.com/tom-miy/agent-privacy-guard/internal/domain"
)

func TestSanitize(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		target             string
		wantSanitized      string
		wantContains       []string
		wantNotContains    []string
		wantFindings       []expectedFinding
		wantRisk           domain.Severity
		wantOutboundOK     bool
		wantMappingCount   *int
		wantPolicyNote     string
		wantRepeatedCounts map[string]int
	}{
		{
			name:             "strong sanitization uses structured placeholders and blocks public secrets",
			input:            "AcmeBank uses prod-db-tokyo with AKIAIOSFODNN7EXAMPLE",
			target:           "claude_api",
			wantContains:     []string{"[CLIENT#A]", "[POSTGRES_DB#A]", "[SECRET:AWS_KEY#A]"},
			wantFindings:     []expectedFinding{{typ: "SECRET:AWS_KEY", severity: domain.SeverityHigh}},
			wantRisk:         domain.SeverityHigh,
			wantOutboundOK:   false,
			wantMappingCount: ptr(3),
		},
		{
			name:             "internal mcp skips sanitization",
			input:            "AcmeBank AKIAIOSFODNN7EXAMPLE",
			target:           "internal_mcp",
			wantSanitized:    "AcmeBank AKIAIOSFODNN7EXAMPLE",
			wantRisk:         domain.SeverityLow,
			wantOutboundOK:   true,
			wantPolicyNote:   "sanitize skipped for trusted target",
			wantMappingCount: ptr(0),
		},
		{
			name: "built-in sensitive values are detected and removed",
			input: strings.Join([]string{
				"contact security@example.test",
				"fetch https://api.internal/v1/users",
				"token = abcdefghijklmnop123456",
			}, "\n"),
			target:          "claude_api",
			wantNotContains: []string{"security@example.test", "https://api.internal/v1/users", "abcdefghijklmnop123456"},
			wantFindings: []expectedFinding{
				{typ: "SECRET:EMAIL", severity: domain.SeverityMedium},
				{typ: "SECRET:INTERNAL_URL", severity: domain.SeverityHigh},
				{typ: "SECRET:TOKEN", severity: domain.SeverityHigh},
			},
			wantRisk:       domain.SeverityHigh,
			wantOutboundOK: false,
		},
		{
			name:             "weak sanitization leaves configured entities but masks secrets",
			input:            "AcmeBank uses prod-db-tokyo and security@example.test",
			target:           "external_mcp",
			wantContains:     []string{"AcmeBank", "prod-db-tokyo"},
			wantNotContains:  []string{"security@example.test"},
			wantFindings:     []expectedFinding{{typ: "SECRET:EMAIL", severity: domain.SeverityMedium}},
			wantRisk:         domain.SeverityMedium,
			wantOutboundOK:   true,
			wantMappingCount: ptr(1),
		},
		{
			name:             "repeated values reuse placeholders",
			input:            "AcmeBank asked AcmeBank to check prod-db and prod-db",
			target:           "claude_api",
			wantRisk:         domain.SeverityMedium,
			wantOutboundOK:   true,
			wantMappingCount: ptr(2),
			wantRepeatedCounts: map[string]int{
				"[CLIENT#A]":      2,
				"[POSTGRES_DB#A]": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Sanitizer{Policy: domain.DefaultPolicy()}.Sanitize(tt.input, tt.target)
			if err != nil {
				t.Fatal(err)
			}

			if tt.wantSanitized != "" && result.Sanitized != tt.wantSanitized {
				t.Fatalf("sanitized output: got %q, want %q", result.Sanitized, tt.wantSanitized)
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(result.Sanitized, want) {
					t.Fatalf("sanitized output missing %s: %s", want, result.Sanitized)
				}
			}
			for _, raw := range tt.wantNotContains {
				if strings.Contains(result.Sanitized, raw) {
					t.Fatalf("sanitized output still contains %q: %s", raw, result.Sanitized)
				}
			}
			for _, want := range tt.wantFindings {
				assertFinding(t, result, want.typ, want.severity)
			}
			if result.Risk != tt.wantRisk {
				t.Fatalf("risk: got %s, want %s", result.Risk, tt.wantRisk)
			}
			if result.OutboundOK != tt.wantOutboundOK {
				t.Fatalf("outbound_ok: got %v, want %v", result.OutboundOK, tt.wantOutboundOK)
			}
			if tt.wantMappingCount != nil && len(result.Mappings) != *tt.wantMappingCount {
				t.Fatalf("mapping count: got %d, want %d (%#v)", len(result.Mappings), *tt.wantMappingCount, result.Mappings)
			}
			if tt.wantPolicyNote != "" && !containsString(result.PolicyNotes, tt.wantPolicyNote) {
				t.Fatalf("policy note %q not found in %#v", tt.wantPolicyNote, result.PolicyNotes)
			}
			for placeholder, want := range tt.wantRepeatedCounts {
				if got := strings.Count(result.Sanitized, placeholder); got != want {
					t.Fatalf("placeholder %s count: got %d, want %d in %s", placeholder, got, want, result.Sanitized)
				}
			}
		})
	}
}

func TestRestoreReplacesLongerPlaceholdersFirst(t *testing.T) {
	got := Sanitizer{}.Restore("Deploy [SECRET:TOKEN#AA] after [SECRET:TOKEN#A]", []domain.MappingEntry{
		{Placeholder: "[SECRET:TOKEN#A]", Value: "short", Type: "SECRET:TOKEN"},
		{Placeholder: "[SECRET:TOKEN#AA]", Value: "long", Type: "SECRET:TOKEN"},
	})
	if got != "Deploy long after short" {
		t.Fatalf("unexpected restored text: %q", got)
	}
}

func TestInspectorDelegatesToSanitizerPolicy(t *testing.T) {
	policy := domain.DefaultPolicy()
	result, err := Inspector{Sanitizer: Sanitizer{Policy: policy}}.Inspect("AcmeBank AKIAIOSFODNN7EXAMPLE", "claude_api")
	if err != nil {
		t.Fatal(err)
	}
	if result.Target != "claude_api" {
		t.Fatalf("expected target to be preserved, got %q", result.Target)
	}
	if result.OutboundOK {
		t.Fatal("expected inspect to report blocked outbound for high severity public secret")
	}
	assertFinding(t, result, "SECRET:AWS_KEY", domain.SeverityHigh)
}

func TestSanitizeUnknownTargetReturnsError(t *testing.T) {
	_, err := Sanitizer{Policy: domain.DefaultPolicy()}.Sanitize("hello", "missing_target")
	if err == nil {
		t.Fatal("expected unknown target error")
	}
	if !strings.Contains(err.Error(), `unknown target "missing_target"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertFinding(t *testing.T, result domain.SanitizationResult, typ string, severity domain.Severity) {
	t.Helper()
	for _, finding := range result.Findings {
		if finding.Type == typ {
			if finding.Severity != severity {
				t.Fatalf("finding %s severity: got %s, want %s", typ, finding.Severity, severity)
			}
			if finding.Placeholder == "" {
				t.Fatalf("finding %s has empty placeholder", typ)
			}
			return
		}
	}
	t.Fatalf("finding %s not found in %#v", typ, result.Findings)
}

type expectedFinding struct {
	typ      string
	severity domain.Severity
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func ptr(v int) *int {
	return &v
}
