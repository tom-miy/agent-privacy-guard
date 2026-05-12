package domain

// Severity represents the risk level of a finding or sanitization result.
// Severity は検出結果やサニタイズ結果のリスクレベルを表します。
type Severity string

const (
	// SeverityLow marks informational or low-risk output.
	// SeverityLow は情報目的または低リスクの出力を表します。
	SeverityLow Severity = "LOW"
	// SeverityMedium marks potentially sensitive output that should be anonymized.
	// SeverityMedium は匿名化すべき可能性がある機微な出力を表します。
	SeverityMedium Severity = "MEDIUM"
	// SeverityHigh marks secrets or risky content that can block public sends.
	// SeverityHigh は公開送信をブロックし得るシークレットや危険な内容を表します。
	SeverityHigh Severity = "HIGH"
	// SeverityCritical marks the highest-risk secrets or commands.
	// SeverityCritical は最も高リスクなシークレットやコマンドを表します。
	SeverityCritical Severity = "CRITICAL"
)

// Finding describes one detected sensitive value and its placeholder.
// Finding は検出された機微な値と対応するプレースホルダーを表します。
type Finding struct {
	Type        string   `json:"type"`
	Value       string   `json:"value,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Severity    Severity `json:"severity"`
}

// MappingEntry records the reversible mapping from placeholder to original value.
// MappingEntry はプレースホルダーから元の値へ戻すための対応関係を記録します。
type MappingEntry struct {
	Placeholder string `json:"placeholder"`
	Value       string `json:"value"`
	Type        string `json:"type"`
}

// SanitizationResult is the full result of inspecting or sanitizing prompt content.
// SanitizationResult はプロンプトの検査またはサニタイズの完全な結果です。
type SanitizationResult struct {
	Original    string         `json:"-"`
	Sanitized   string         `json:"sanitized"`
	Findings    []Finding      `json:"findings"`
	Mappings    []MappingEntry `json:"mappings"`
	Target      string         `json:"target"`
	Risk        Severity       `json:"risk"`
	OutboundOK  bool           `json:"outbound_ok"`
	PolicyNotes []string       `json:"policy_notes"`
}
