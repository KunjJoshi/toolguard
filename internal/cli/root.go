// Package cli implements the toolguard command-line interface.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute runs the toolguard root command with the process arguments and
// returns the exit code: 0 on success, 1 if the command failed. Errors are
// printed to stderr.
func Execute() int {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "toolguard:", err)
		return 1
	}
	return 0
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "toolguard",
		Short: "Deterministic guardrails for AI agent tool calls",
		Long: "ToolGuard inspects AI agent tool calls and blocks or redacts " +
			"credentials and private data before they leak.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(newVersionCmd(), newEvalCmd())
	return root
}
