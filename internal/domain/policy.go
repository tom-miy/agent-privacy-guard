package domain

// TrustLevel classifies the trust boundary of an outbound target.
// TrustLevel は送信先がどの信頼境界に属するかを表します。
type TrustLevel string

const (
	// TrustPublic identifies destinations outside the local or internal trust boundary.
	// TrustPublic はローカルまたは内部の信頼境界の外にある送信先を表します。
	TrustPublic TrustLevel = "public"
	// TrustInternal identifies destinations controlled by the local or internal environment.
	// TrustInternal はローカルまたは内部環境で管理される送信先を表します。
	TrustInternal TrustLevel = "internal"
	// TrustConfidential identifies internal destinations allowed to receive confidential context.
	// TrustConfidential は機密コンテキストを扱える内部送信先を表します。
	TrustConfidential TrustLevel = "confidential"
	// TrustSecret identifies the most restricted destination class.
	// TrustSecret は最も制限の強い送信先区分を表します。
	TrustSecret TrustLevel = "secret"
)

// SanitizeLevel controls how much prompt content is anonymized before sending.
// SanitizeLevel は送信前にプロンプト内容をどの程度匿名化するかを制御します。
type SanitizeLevel string

const (
	// SanitizeNone leaves input unchanged for trusted targets.
	// SanitizeNone は信頼済み送信先に対して入力を変更しません。
	SanitizeNone SanitizeLevel = "none"
	// SanitizeWeak applies built-in secret detectors only.
	// SanitizeWeak は組み込みのシークレット検出のみを適用します。
	SanitizeWeak SanitizeLevel = "weak"
	// SanitizeStrong applies built-in secret detectors and configured entity rules.
	// SanitizeStrong は組み込み検出と設定済みエンティティルールを適用します。
	SanitizeStrong SanitizeLevel = "strong"
)

// Policy describes target-specific sanitization and outbound safety rules.
// Policy は送信先ごとのサニタイズ方針と送信安全ルールを表します。
type Policy struct {
	Targets  map[string]TargetPolicy `yaml:"targets"`
	Entities []EntityRule            `yaml:"entities"`
	Outbound OutboundPolicy          `yaml:"outbound"`
}

// TargetPolicy describes how a named outbound destination should be handled.
// TargetPolicy は名前付き送信先をどのように扱うかを表します。
type TargetPolicy struct {
	Trust    TrustLevel    `yaml:"trust"`
	Sanitize SanitizeLevel `yaml:"sanitize"`
	Allow    bool          `yaml:"allow"`
	Mode     string        `yaml:"mode"`
	Notes    string        `yaml:"notes"`
}

// EntityRule defines a project-specific regular expression to replace with a placeholder.
// EntityRule はプレースホルダーへ置換するプロジェクト固有の正規表現を定義します。
type EntityRule struct {
	Type    string `yaml:"type"`
	Pattern string `yaml:"pattern"`
	Scope   string `yaml:"scope"`
}

// OutboundPolicy contains global controls that can block or constrain sends.
// OutboundPolicy は送信をブロックまたは制約するグローバル制御を持ちます。
type OutboundPolicy struct {
	BlockOnSecret bool `yaml:"block_on_secret"`
	DiffOnly      bool `yaml:"diff_only"`
}

// DefaultPolicy returns the built-in sample policy used when no policy file exists.
// DefaultPolicy はポリシーファイルが存在しない場合に使う組み込みサンプルポリシーを返します。
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
