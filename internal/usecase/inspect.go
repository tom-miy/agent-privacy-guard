package usecase

import "github.com/tom-miy/agent-privacy-guard/internal/domain"

type Inspector struct {
	Sanitizer Sanitizer
}

func (i Inspector) Inspect(input, target string) (domain.SanitizationResult, error) {
	return i.Sanitizer.Sanitize(input, target)
}
