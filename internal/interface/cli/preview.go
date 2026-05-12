package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tom-miy/agent-privacy-guard/internal/infra/config"
	"github.com/tom-miy/agent-privacy-guard/internal/usecase"
)

func newPreviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview sanitization as a compact diff",
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
			if outputJSON {
				return printJSON(result.Mappings)
			}
			for _, mapping := range result.Mappings {
				fmt.Printf("- %s\n", oneLine(mapping.Value))
				fmt.Printf("+ %s\n\n", mapping.Placeholder)
			}
			if len(result.Mappings) == 0 {
				fmt.Println("No sanitization changes.")
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&inputPath, "input", "i", "-", "input prompt path or stdin")
	return cmd
}

func oneLine(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "\n", `\n`)
}
