package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "0000bca"
	date    = "11/17/2026"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the ToolGuard version",
		Args:  cobra.NoArgs,
		RunE:  runVersion,
	}
}

func runVersion(cmd *cobra.Command, args []string) error {
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "toolguard %s (commit %s, built %s)\n", version, commit, date)
	return err
}
