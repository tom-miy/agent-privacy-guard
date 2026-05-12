package config

import (
	"errors"
	"os"

	"github.com/tom-miy/agent-privacy-guard/internal/domain"
	"gopkg.in/yaml.v3"
)

// LoadPolicy reads a YAML policy file or returns DefaultPolicy when the file is missing.
// LoadPolicy は YAML ポリシーファイルを読み込み、ファイルがない場合は DefaultPolicy を返します。
func LoadPolicy(path string) (domain.Policy, error) {
	if path == "" {
		path = "configs/policy.yaml"
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.DefaultPolicy(), nil
		}
		return domain.Policy{}, err
	}

	p := domain.DefaultPolicy()
	if err := yaml.Unmarshal(b, &p); err != nil {
		return domain.Policy{}, err
	}
	return p, nil
}

// ValidatePolicy returns human-readable problems for missing required policy fields.
// ValidatePolicy は必須項目の不足を人が読める問題一覧として返します。
func ValidatePolicy(p domain.Policy) []string {
	var problems []string
	if len(p.Targets) == 0 {
		problems = append(problems, "policy must define at least one target")
	}
	for name, target := range p.Targets {
		if target.Trust == "" {
			problems = append(problems, "target "+name+" is missing trust")
		}
		if target.Sanitize == "" {
			problems = append(problems, "target "+name+" is missing sanitize")
		}
	}
	for _, rule := range p.Entities {
		if rule.Type == "" || rule.Pattern == "" {
			problems = append(problems, "entity rules require type and pattern")
		}
	}
	return problems
}
