package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tom-miy/agent-privacy-guard/internal/domain"
	"github.com/tom-miy/agent-privacy-guard/internal/infra/audit"
	"github.com/tom-miy/agent-privacy-guard/internal/usecase"
)

func newPosthookCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posthook",
		Short: "Inspect agent response for dangerous commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := readInput(inputPath)
			if err != nil {
				return err
			}
			findings := usecase.PosthookInspector{}.Inspect(input)
			allowed := true
			for _, f := range findings {
				if f.Severity == domain.SeverityHigh || f.Severity == domain.SeverityCritical {
					allowed = false
				}
			}
			_ = audit.JSONLogger{Path: auditPath}.Write(audit.Event{Action: "posthook", Target: targetName, Allowed: allowed, Details: findings})
			if outputJSON {
				return printJSON(map[string]interface{}{"allowed": allowed, "findings": findings})
			}
			if len(findings) == 0 {
				fmt.Println("Posthook Risk: LOW")
				fmt.Println("Detected: none")
				return nil
			}
			fmt.Println("Posthook Risk: HIGH")
			fmt.Println("Detected:")
			for _, f := range findings {
				fmt.Printf("- %s: %s (%s)\n", f.Severity, f.Command, f.Reason)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&inputPath, "input", "i", "-", "input response path or stdin")
	return cmd
}
