package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tom-miy/agent-privacy-guard/internal/infra/audit"
	"github.com/tom-miy/agent-privacy-guard/internal/infra/config"
	"github.com/tom-miy/agent-privacy-guard/internal/usecase"
)

func newInspectCommand() *cobra.Command {
	var failOnBlock bool
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect outbound prompt risk",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := readInput(inputPath)
			if err != nil {
				return err
			}
			policy, err := config.LoadPolicy(policyPath)
			if err != nil {
				return err
			}
			result, err := usecase.Inspector{Sanitizer: usecase.Sanitizer{Policy: policy}}.Inspect(input, targetName)
			if err != nil {
				return err
			}
			_ = audit.JSONLogger{Path: auditPath}.Write(audit.Event{Action: "inspect", Target: targetName, Risk: string(result.Risk), Allowed: result.OutboundOK, Details: result.Findings})
			if outputJSON {
				if err := printJSON(result); err != nil {
					return err
				}
				if failOnBlock && !result.OutboundOK {
					return fmt.Errorf("outbound blocked by policy")
				}
				return nil
			}
			fmt.Printf("Outbound Risk: %s\n\n", result.Risk)
			fmt.Println("Detected:")
			if len(result.Findings) == 0 {
				fmt.Println("- none")
			}
			for _, finding := range result.Findings {
				fmt.Printf("- %s -> %s\n", finding.Type, finding.Placeholder)
			}
			fmt.Println("\nSanitized Entities:")
			if len(result.Mappings) == 0 {
				fmt.Println("- none")
			}
			for _, mapping := range result.Mappings {
				fmt.Printf("- %s\n", mapping.Placeholder)
			}
			fmt.Println("\nPolicy:")
			fmt.Printf("- outbound allowed: %v\n", result.OutboundOK)
			for _, note := range result.PolicyNotes {
				fmt.Printf("- %s\n", note)
			}
			if failOnBlock && !result.OutboundOK {
				return fmt.Errorf("outbound blocked by policy")
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&inputPath, "input", "i", "-", "input prompt path or stdin")
	cmd.Flags().BoolVar(&failOnBlock, "fail-on-block", false, "exit with non-zero status when policy blocks outbound send")
	return cmd
}
