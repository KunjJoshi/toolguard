// Package cli implements the toolguard command-line interface.
package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

func exitCodeFor(err error) int {
	var e *exitError
	if errors.As(err, &e) {
		return e.code
	}
	return 1
}

// Execute runs the toolguard root command with the process arguments and
// returns the exit code: 0 on success, the code carried by an exitError (for
// example 2 when the hook blocks a call it could not check), or 1 for any other
// failure. Errors are printed to stderr.
func Execute() int {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "toolguard:", err)
		return exitCodeFor(err)
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

	root.AddCommand(newVersionCmd(), newEvalCmd(), newHookCmd())
	return root
}
