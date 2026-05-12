package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tom-miy/agent-privacy-guard/internal/domain"
	"github.com/tom-miy/agent-privacy-guard/internal/infra/audit"
	"github.com/tom-miy/agent-privacy-guard/internal/infra/config"
	"github.com/tom-miy/agent-privacy-guard/internal/usecase"
)

var mappingPath string

func newSanitizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sanitize",
		Short: "Sanitize prompt before sending it to an agent or LLM",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := readInput(inputPath)
			if err != nil {
				return err
			}
			policy, err := config.LoadPolicy(policyPath)
			if err != nil {
				return err
			}
			result, err := usecase.Sanitizer{Policy: policy}.Sanitize(input, targetName)
			if err != nil {
				return err
			}
			if mappingPath != "" {
				b, err := json.MarshalIndent(result.Mappings, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(mappingPath, b, 0o600); err != nil {
					return err
				}
			}
			_ = audit.JSONLogger{Path: auditPath}.Write(audit.Event{Action: "sanitize", Target: targetName, Risk: string(result.Risk), Allowed: result.OutboundOK, Details: result.Findings})
			if outputJSON {
				return printJSON(result)
			}
			fmt.Print(result.Sanitized)
			return nil
		},
	}
	cmd.Flags().StringVarP(&inputPath, "input", "i", "-", "input prompt path or stdin")
	cmd.Flags().StringVar(&mappingPath, "mapping-out", "", "write reversible placeholder mapping JSON")
	return cmd
}

func newRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore placeholders in an agent response using a mapping file",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := readInput(inputPath)
			if err != nil {
				return err
			}
			b, err := os.ReadFile(mappingPath)
			if err != nil {
				return err
			}
			var mappings []domain.MappingEntry
			if err := json.Unmarshal(b, &mappings); err != nil {
				return err
			}
			fmt.Print(usecase.Sanitizer{}.Restore(input, mappings))
			return nil
		},
	}
	cmd.Flags().StringVarP(&inputPath, "input", "i", "-", "input response path or stdin")
	cmd.Flags().StringVar(&mappingPath, "mapping", "mapping.json", "placeholder mapping JSON")
	return cmd
}
