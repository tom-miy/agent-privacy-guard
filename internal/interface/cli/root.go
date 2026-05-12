package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	policyPath string
	targetName string
	inputPath  string
	outputJSON bool
	auditPath  string
)

func Execute() {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "agent-privacy-guard",
		Short:         "AI Agent Gateway for policy-driven prompt sanitization",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVarP(&policyPath, "policy", "p", "configs/policy.yaml", "policy YAML path")
	cmd.PersistentFlags().StringVarP(&targetName, "target", "t", "claude_api", "target name")
	cmd.PersistentFlags().StringVar(&auditPath, "audit-log", "audit/agent-privacy-guard.jsonl", "audit JSONL path")
	cmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "output JSON")

	cmd.AddCommand(newInspectCommand())
	cmd.AddCommand(newSanitizeCommand())
	cmd.AddCommand(newPreviewCommand())
	cmd.AddCommand(newRestoreCommand())
	cmd.AddCommand(newPosthookCommand())
	cmd.AddCommand(newValidateCommand())
	return cmd
}
