package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
