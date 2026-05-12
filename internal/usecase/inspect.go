package usecase

import "github.com/tom-miy/agent-privacy-guard/internal/domain"

// Inspector evaluates outbound prompt risk using a configured Sanitizer.
// Inspector は設定済みの Sanitizer を使って送信プロンプトのリスクを評価します。
type Inspector struct {
	Sanitizer Sanitizer
}

// Inspect returns the same structured result as sanitization for an outbound target.
// Inspect は送信先に対するサニタイズと同じ構造化結果を返します。
func (i Inspector) Inspect(input, target string) (domain.SanitizationResult, error) {
	return i.Sanitizer.Sanitize(input, target)
}
