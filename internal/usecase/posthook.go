package usecase

import (
	"regexp"

	"github.com/tom-miy/agent-privacy-guard/internal/domain"
)

type PosthookInspector struct{}

func (p PosthookInspector) Inspect(response string) []domain.CommandFinding {
	patterns := []struct {
		re     *regexp.Regexp
		reason string
		sev    domain.Severity
	}{
		{regexp.MustCompile(`(?m)\brm\s+-rf\s+(/|\$HOME|~|\.)`), "destructive recursive removal", domain.SeverityCritical},
		{regexp.MustCompile(`(?m)\bsudo\s+`), "privileged command requires review", domain.SeverityHigh},
		{regexp.MustCompile(`(?m)\b(curl|wget)\b.+\|\s*(sh|bash)`), "remote script execution", domain.SeverityCritical},
		{regexp.MustCompile(`(?m)\bgit\s+push\s+--force\b`), "force push can rewrite shared history", domain.SeverityHigh},
		{regexp.MustCompile(`(?m)\bchmod\s+-R\s+777\b`), "world-writable recursive permissions", domain.SeverityHigh},
		{regexp.MustCompile(`(?m)^\+\+\+\s+b/(?:\.env|.*private.*key|.*secret.*|.*agent-privacy-guard\.mapping\.json)`), "patch touches forbidden sensitive file", domain.SeverityCritical},
	}

	var findings []domain.CommandFinding
	for _, p := range patterns {
		for _, match := range p.re.FindAllString(response, -1) {
			findings = append(findings, domain.CommandFinding{Command: match, Reason: p.reason, Severity: p.sev})
		}
	}
	return findings
}
