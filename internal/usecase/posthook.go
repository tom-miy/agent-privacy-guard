package usecase

import (
	"regexp"

	"github.com/tom-miy/agent-privacy-guard/internal/domain"
)

// PosthookInspector scans agent responses for risky commands or sensitive patch targets.
// PosthookInspector はエージェント応答から危険なコマンドや機微なパッチ対象を検出します。
type PosthookInspector struct{}

// Inspect returns command findings that should be reviewed before execution or application.
// Inspect は実行または適用前に確認すべきコマンド検出結果を返します。
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
