package cli

import (
	"fmt"
	"strings"

	"github.com/KunjJoshi/toolguard/internal/adapters/claudecode"
	"github.com/KunjJoshi/toolguard/internal/event"
	"github.com/spf13/cobra"
)

const exitBlock = 2

func newHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Run as a Claude Code PreToolUse/PostToolUse hook",
		Long: "hook reads one Claude Code hook payload on stdin and checks it.\n" +
			"PreToolUse: a card in the arguments denies the call.\n" +
			"PostToolUse: a card in the result is reported to Claude.\n" +
			"If the payload cannot be checked, the call is blocked (fail closed).",
		Args: cobra.NoArgs,
		RunE: runHook,
	}
}

func runHook(cmd *cobra.Command, args []string) error {
	payload, err := claudecode.ParsePayload(cmd.InOrStdin())
	if err != nil {
		return failClosed(err)
	}

	ev := payload.ToEvent()
	found := fieldsToCheck(ev)
	if len(found) == 0 {
		return nil
	}

	if err := claudecode.WriteBlock(cmd.OutOrStdout(), payload.HookEventName, hookReason(ev, found)); err != nil {
		return failClosed(err)
	}

	return nil
}

func fieldsToCheck(ev event.Event) []detection {
	if ev.Direction == event.DirResponse {
		return scanFields("result", ev.ResultFields)
	}

	return scanFields("args", ev.ArgFields)
}

func failClosed(err error) error {
	return &exitError{
		code: exitBlock,
		err:  fmt.Errorf("could not check this tool call, so it was blocked: %w", err),
	}
}

func hookReason(ev event.Event, found []detection) string {
	items := make([]string, 0, len(found))
	for _, d := range found {
		items = append(items, fmt.Sprintf("%s in %s", d.match.Masked, d.location))
	}

	list := strings.Join(items, ", ")
	if ev.Direction == event.DirResponse {
		return fmt.Sprintf("toolguard: the result of %s contains payment card data (%s). "+
			"Do not repeat, store or send these values.", toolLabel(ev), list)
	}
	return fmt.Sprintf("toolguard blocked this call to %s: it contains payment card data (%s). "+
		"Do not send card numbers through tools.", toolLabel(ev), list)
}
