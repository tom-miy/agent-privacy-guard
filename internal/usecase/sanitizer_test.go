package usecase

import (
	"strings"
	"testing"

	"github.com/tom-miy/agent-privacy-guard/internal/domain"
)

func TestSanitizeUsesStructuredPlaceholders(t *testing.T) {
	policy := domain.DefaultPolicy()
	result, err := Sanitizer{Policy: policy}.Sanitize("AcmeBank uses prod-db-tokyo with AKIAIOSFODNN7EXAMPLE", "claude_api")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"[CLIENT#A]", "[POSTGRES_DB#A]", "[SECRET:AWS_KEY#A]"} {
		if !strings.Contains(result.Sanitized, want) {
			t.Fatalf("sanitized output missing %s: %s", want, result.Sanitized)
		}
	}
	if result.OutboundOK {
		t.Fatal("expected public outbound to be blocked when high severity secret is present")
	}
}

func TestInternalMCPSkipsSanitization(t *testing.T) {
	policy := domain.DefaultPolicy()
	result, err := Sanitizer{Policy: policy}.Sanitize("AcmeBank AKIAIOSFODNN7EXAMPLE", "internal_mcp")
	if err != nil {
		t.Fatal(err)
	}
	if result.Sanitized != "AcmeBank AKIAIOSFODNN7EXAMPLE" {
		t.Fatalf("expected unchanged prompt, got %q", result.Sanitized)
	}
}
