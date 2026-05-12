package domain

type TrustLevel string

const (
	TrustPublic       TrustLevel = "public"
	TrustInternal     TrustLevel = "internal"
	TrustConfidential TrustLevel = "confidential"
	TrustSecret       TrustLevel = "secret"
)

type SanitizeLevel string

const (
	SanitizeNone   SanitizeLevel = "none"
	SanitizeWeak   SanitizeLevel = "weak"
	SanitizeStrong SanitizeLevel = "strong"
)

type Policy struct {
	Targets  map[string]TargetPolicy `yaml:"targets"`
	Entities []EntityRule            `yaml:"entities"`
	Outbound OutboundPolicy          `yaml:"outbound"`
}

type TargetPolicy struct {
	Trust    TrustLevel    `yaml:"trust"`
	Sanitize SanitizeLevel `yaml:"sanitize"`
	Allow    bool          `yaml:"allow"`
	Mode     string        `yaml:"mode"`
	Notes    string        `yaml:"notes"`
}

type EntityRule struct {
	Type    string `yaml:"type"`
	Pattern string `yaml:"pattern"`
	Scope   string `yaml:"scope"`
}

type OutboundPolicy struct {
	BlockOnSecret bool `yaml:"block_on_secret"`
	DiffOnly      bool `yaml:"diff_only"`
}

func DefaultPolicy() Policy {
	return Policy{
		Targets: map[string]TargetPolicy{
			"claude_api":   {Trust: TrustPublic, Sanitize: SanitizeStrong, Allow: true, Mode: "external_llm"},
			"cursor":       {Trust: TrustPublic, Sanitize: SanitizeStrong, Allow: true, Mode: "agent"},
			"copilot":      {Trust: TrustPublic, Sanitize: SanitizeStrong, Allow: true, Mode: "agent"},
			"codex":        {Trust: TrustPublic, Sanitize: SanitizeStrong, Allow: true, Mode: "agent"},
			"local_qwen":   {Trust: TrustInternal, Sanitize: SanitizeWeak, Allow: true, Mode: "local_llm"},
			"internal_mcp": {Trust: TrustInternal, Sanitize: SanitizeNone, Allow: true, Mode: "internal_mcp"},
			"external_mcp": {Trust: TrustPublic, Sanitize: SanitizeWeak, Allow: true, Mode: "external_mcp"},
		},
		Entities: []EntityRule{
			{Type: "CLIENT", Pattern: `\b(AcmeBank|ExampleCorp|MegaRetail)\b`, Scope: "prompt"},
			{Type: "POSTGRES_DB", Pattern: `\b[a-z0-9-]*db[a-z0-9-]*\b`, Scope: "prompt"},
		},
		Outbound: OutboundPolicy{BlockOnSecret: true, DiffOnly: true},
	}
}
