package usecase

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/tom-miy/agent-privacy-guard/internal/domain"
)

// Sanitizer applies policy-driven placeholder replacement to outbound prompt content.
// Sanitizer は送信プロンプトに対してポリシーに基づくプレースホルダー置換を適用します。
type Sanitizer struct {
	Policy domain.Policy
}

type detector struct {
	typ      string
	pattern  *regexp.Regexp
	severity domain.Severity
}

// Sanitize detects sensitive values for the target and returns sanitized content plus policy decisions.
// Sanitize は送信先に応じて機微な値を検出し、サニタイズ済み内容とポリシー判断を返します。
func (s Sanitizer) Sanitize(input, targetName string) (domain.SanitizationResult, error) {
	target, ok := s.Policy.Targets[targetName]
	if !ok {
		return domain.SanitizationResult{}, fmt.Errorf("unknown target %q", targetName)
	}

	result := domain.SanitizationResult{
		Original:   input,
		Sanitized:  normalizePaths(input),
		Target:     targetName,
		OutboundOK: target.Allow,
		Risk:       domain.SeverityLow,
	}
	if target.Sanitize == domain.SanitizeNone {
		result.PolicyNotes = append(result.PolicyNotes, "sanitize skipped for trusted target")
		return result, nil
	}

	book := newPlaceholderBook()
	for _, d := range secretDetectors() {
		result.Sanitized = replaceMatches(result.Sanitized, d.pattern, func(value string) string {
			placeholder := book.placeholder("SECRET:"+d.typ, value)
			result.Findings = append(result.Findings, domain.Finding{Type: "SECRET:" + d.typ, Value: value, Placeholder: placeholder, Severity: d.severity})
			result.Mappings = book.entries()
			result.Risk = maxSeverity(result.Risk, d.severity)
			return placeholder
		})
	}

	if target.Sanitize == domain.SanitizeStrong {
		for _, rule := range s.Policy.Entities {
			re, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return domain.SanitizationResult{}, fmt.Errorf("invalid entity pattern for %s: %w", rule.Type, err)
			}
			result.Sanitized = replaceMatches(result.Sanitized, re, func(value string) string {
				placeholder := book.placeholder(rule.Type, value)
				result.Findings = append(result.Findings, domain.Finding{Type: rule.Type, Value: value, Placeholder: placeholder, Severity: domain.SeverityMedium})
				result.Mappings = book.entries()
				result.Risk = maxSeverity(result.Risk, domain.SeverityMedium)
				return placeholder
			})
		}
	}

	result.Mappings = book.entries()
	if s.Policy.Outbound.BlockOnSecret && hasCriticalOrHighSecret(result.Findings) && target.Trust == domain.TrustPublic {
		result.OutboundOK = false
		result.PolicyNotes = append(result.PolicyNotes, "external send blocked because secret was detected")
	}
	if s.Policy.Outbound.DiffOnly {
		result.PolicyNotes = append(result.PolicyNotes, "diff-only context is recommended")
	}
	return result, nil
}

// Restore replaces placeholders in input using the supplied reversible mappings.
// Restore は指定された可逆マッピングを使って入力内のプレースホルダーを元の値へ戻します。
func (s Sanitizer) Restore(input string, mappings []domain.MappingEntry) string {
	out := input
	sort.SliceStable(mappings, func(i, j int) bool {
		return len(mappings[i].Placeholder) > len(mappings[j].Placeholder)
	})
	for _, m := range mappings {
		out = strings.ReplaceAll(out, m.Placeholder, m.Value)
	}
	return out
}

func secretDetectors() []detector {
	return []detector{
		{typ: "AWS_KEY", pattern: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`), severity: domain.SeverityHigh},
		{typ: "AWS_ARN", pattern: regexp.MustCompile(`\barn:aws:[a-z0-9-]+:[a-z0-9-]*:\d{12}:[A-Za-z0-9/+=,.@_-]+\b`), severity: domain.SeverityHigh},
		{typ: "EMAIL", pattern: regexp.MustCompile(`\b[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}\b`), severity: domain.SeverityMedium},
		{typ: "INTERNAL_URL", pattern: regexp.MustCompile(`https?://(?:localhost|127\.0\.0\.1|[A-Za-z0-9.-]+\.internal|[A-Za-z0-9.-]+\.corp)(?:/[^\s\]\)]*)?`), severity: domain.SeverityHigh},
		{typ: "TOKEN", pattern: regexp.MustCompile(`(?i)\b(?:token|api[_-]?key|secret)\s*[:=]\s*['"]?[A-Za-z0-9._\-]{16,}`), severity: domain.SeverityHigh},
		{typ: "SSH_KEY", pattern: regexp.MustCompile(`-----BEGIN OPENSSH PRIVATE KEY-----[\s\S]*?-----END OPENSSH PRIVATE KEY-----`), severity: domain.SeverityCritical},
	}
}

func replaceMatches(input string, re *regexp.Regexp, fn func(string) string) string {
	return re.ReplaceAllStringFunc(input, fn)
}

func normalizePaths(input string) string {
	homePattern := regexp.MustCompile(`/Users/[A-Za-z0-9._-]+/`)
	return homePattern.ReplaceAllString(input, `/Users/[USER]/`)
}

func hasCriticalOrHighSecret(findings []domain.Finding) bool {
	for _, f := range findings {
		if strings.HasPrefix(f.Type, "SECRET:") && (f.Severity == domain.SeverityHigh || f.Severity == domain.SeverityCritical) {
			return true
		}
	}
	return false
}

func maxSeverity(a, b domain.Severity) domain.Severity {
	rank := map[domain.Severity]int{
		domain.SeverityLow: 1, domain.SeverityMedium: 2, domain.SeverityHigh: 3, domain.SeverityCritical: 4,
	}
	if rank[b] > rank[a] {
		return b
	}
	return a
}

type placeholderBook struct {
	byValue map[string]domain.MappingEntry
	counts  map[string]int
}

func newPlaceholderBook() *placeholderBook {
	return &placeholderBook{byValue: map[string]domain.MappingEntry{}, counts: map[string]int{}}
}

func (b *placeholderBook) placeholder(typ, value string) string {
	key := typ + "\x00" + value
	if entry, ok := b.byValue[key]; ok {
		return entry.Placeholder
	}
	b.counts[typ]++
	placeholder := fmt.Sprintf("[%s#%s]", typ, indexLabel(b.counts[typ]))
	entry := domain.MappingEntry{Placeholder: placeholder, Value: value, Type: typ}
	b.byValue[key] = entry
	return placeholder
}

func (b *placeholderBook) entries() []domain.MappingEntry {
	entries := make([]domain.MappingEntry, 0, len(b.byValue))
	for _, entry := range b.byValue {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Type == entries[j].Type {
			return entries[i].Placeholder < entries[j].Placeholder
		}
		return entries[i].Type < entries[j].Type
	})
	return entries
}

func indexLabel(n int) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if n <= len(alphabet) {
		return string(alphabet[n-1])
	}
	return fmt.Sprintf("%d", n)
}
