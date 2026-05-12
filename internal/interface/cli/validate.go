package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tom-miy/agent-privacy-guard/internal/infra/config"
)

func newValidateCommand() *cobra.Command {
	var paths []string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate policy and agent configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			policy, err := config.LoadPolicy(policyPath)
			if err != nil {
				return err
			}
			problems := config.ValidatePolicy(policy)
			for _, path := range paths {
				if _, err := os.Stat(path); err != nil {
					problems = append(problems, "missing config file: "+path)
				}
			}
			if outputJSON {
				return printJSON(map[string]interface{}{"ok": len(problems) == 0, "problems": problems})
			}
			if len(problems) == 0 {
				fmt.Println("Validation: OK")
				return nil
			}
			fmt.Println("Validation: FAILED")
			for _, problem := range problems {
				fmt.Printf("- %s\n", problem)
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&paths, "agent-config", []string{"AGENTS.md", "CLAUDE.md", ".cursorrules", ".codex/config.toml", "configs/mcp-trust.yaml"}, "agent config paths to lint for existence")
	return cmd
}
