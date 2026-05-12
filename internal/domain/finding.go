package domain

type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

type Finding struct {
	Type        string   `json:"type"`
	Value       string   `json:"value,omitempty"`
	Placeholder string   `json:"placeholder,omitempty"`
	Severity    Severity `json:"severity"`
}

type MappingEntry struct {
	Placeholder string `json:"placeholder"`
	Value       string `json:"value"`
	Type        string `json:"type"`
}

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
